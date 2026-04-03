package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"

	"recipick/backend/internal/model"
)

type JSONLDImporter struct {
	client *http.Client
}

func NewJSONLDImporter() *JSONLDImporter {
	return &JSONLDImporter{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *JSONLDImporter) ImportPreview(ctx context.Context, sourceURL string) model.ImportPreview {
	preview := model.ImportPreview{
		SourceURL:      sourceURL,
		Ingredients:    []string{},
		SystemTags:     []string{},
		ImportWarnings: []string{},
	}

	if _, err := url.ParseRequestURI(sourceURL); err != nil {
		preview.ImportWarnings = append(preview.ImportWarnings, "invalid sourceUrl")
		return preview
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		preview.ImportWarnings = append(preview.ImportWarnings, "failed to build fetch request")
		return preview
	}
	req.Header.Set("User-Agent", "recipick-mvp/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		preview.ImportWarnings = append(preview.ImportWarnings, fmt.Sprintf("failed to fetch HTML: %v", err))
		return preview
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		preview.ImportWarnings = append(preview.ImportWarnings, fmt.Sprintf("source responded with status %d", resp.StatusCode))
		return preview
	}

	htmlBytes, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		preview.ImportWarnings = append(preview.ImportWarnings, "failed to read HTML body")
		return preview
	}

	scripts := extractJSONLDScripts(string(htmlBytes))
	if len(scripts) == 0 {
		preview.ImportWarnings = append(preview.ImportWarnings, "json-ld script not found")
		return preview
	}

	recipeObj, rawRecipeObj, err := firstRecipeObject(scripts)
	if err != nil {
		preview.ImportWarnings = append(preview.ImportWarnings, err.Error())
		return preview
	}

	preview.JSONLDFound = true
	preview.RawJSONLD = rawRecipeObj

	title := asString(recipeObj["name"])
	if title != "" {
		preview.Title = &title
	}
	servings := asString(recipeObj["recipeYield"])
	if servings != "" {
		preview.Servings = &servings
	}

	totalMinutes, ok := parseDurationMinutes(asString(recipeObj["totalTime"]))
	if ok {
		preview.TotalMinutes = &totalMinutes
	}

	preview.Ingredients = toStringList(recipeObj["recipeIngredient"])
	preview.SystemTags = extractSystemTags(recipeObj)

	return preview
}

func extractJSONLDScripts(body string) []string {
	t := html.NewTokenizer(strings.NewReader(body))
	result := make([]string, 0)
	for {
		tt := t.Next()
		switch tt {
		case html.ErrorToken:
			if errors.Is(t.Err(), io.EOF) {
				return result
			}
			return result
		case html.StartTagToken:
			tok := t.Token()
			if tok.Data != "script" {
				continue
			}
			if !isJSONLD(tok.Attr) {
				continue
			}
			if t.Next() != html.TextToken {
				continue
			}
			content := strings.TrimSpace(t.Token().Data)
			if content != "" {
				result = append(result, content)
			}
		}
	}
}

func isJSONLD(attrs []html.Attribute) bool {
	for _, attr := range attrs {
		if strings.EqualFold(attr.Key, "type") && strings.EqualFold(attr.Val, "application/ld+json") {
			return true
		}
	}
	return false
}

func firstRecipeObject(scripts []string) (map[string]any, json.RawMessage, error) {
	for _, script := range scripts {
		var payload any
		if err := json.Unmarshal([]byte(script), &payload); err != nil {
			continue
		}
		for _, candidate := range flattenJSONLD(payload) {
			if !isRecipeType(candidate["@type"]) {
				continue
			}
			raw, err := json.Marshal(candidate)
			if err != nil {
				return nil, nil, err
			}
			return candidate, raw, nil
		}
	}
	return nil, nil, errors.New("json-ld found but Recipe type not found")
}

func flattenJSONLD(payload any) []map[string]any {
	result := make([]map[string]any, 0)
	switch v := payload.(type) {
	case map[string]any:
		if graph, ok := v["@graph"].([]any); ok {
			for _, node := range graph {
				if m, ok := node.(map[string]any); ok {
					result = append(result, m)
				}
			}
		}
		result = append(result, v)
	case []any:
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				result = append(result, m)
			}
		}
	}
	return result
}

func isRecipeType(v any) bool {
	switch t := v.(type) {
	case string:
		return strings.EqualFold(t, "Recipe")
	case []any:
		for _, item := range t {
			if s, ok := item.(string); ok && strings.EqualFold(s, "Recipe") {
				return true
			}
		}
	}
	return false
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case []any:
		if len(t) == 0 {
			return ""
		}
		if s, ok := t[0].(string); ok {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func toStringList(v any) []string {
	result := make([]string, 0)
	switch t := v.(type) {
	case []any:
		for _, item := range t {
			if s, ok := item.(string); ok {
				s = strings.TrimSpace(s)
				if s != "" {
					result = append(result, s)
				}
			}
		}
	case string:
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return dedupeStrings(result)
}

func extractSystemTags(recipe map[string]any) []string {
	result := make([]string, 0)
	result = append(result, toStringList(recipe["recipeCategory"])...)
	result = append(result, toStringList(recipe["recipeCuisine"])...)
	result = append(result, parseKeywords(recipe["keywords"])...)
	return dedupeStrings(result)
}

func parseKeywords(v any) []string {
	if arr, ok := v.([]any); ok {
		return toStringList(arr)
	}
	s := asString(v)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

var durationRegex = regexp.MustCompile(`^P(?:\d+D)?(?:T(?:(\d+)H)?(?:(\d+)M)?)?$`)

func parseDurationMinutes(v string) (int, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	matches := durationRegex.FindStringSubmatch(v)
	if len(matches) == 0 {
		return 0, false
	}
	hours := parseInt(matches[1])
	minutes := parseInt(matches[2])
	total := hours*60 + minutes
	if total <= 0 {
		return 0, false
	}
	return total, true
}

func parseInt(v string) int {
	if v == "" {
		return 0
	}
	var n int
	_, _ = fmt.Sscanf(v, "%d", &n)
	return n
}

func dedupeStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, v := range values {
		t := strings.TrimSpace(v)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		result = append(result, t)
	}
	return result
}
