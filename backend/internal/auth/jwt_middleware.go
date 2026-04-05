package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const userIDCtxKey contextKey = "userID"

type Config struct {
	Issuer        string
	Audience      string
	JWKSURL       string
	HTTPTimeout   time.Duration
	RefreshWindow time.Duration
	Leeway        time.Duration
}

type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KID string `json:"kid"`
	KTY string `json:"kty"`
	ALG string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	X   string `json:"x"`
	Y   string `json:"y"`
	CRV string `json:"crv"`
}

type Middleware struct {
	cfg Config

	client *http.Client

	mu          sync.RWMutex
	keysByKID   map[string]any
	lastRefresh time.Time
}

func NewMiddleware(cfg Config) (*Middleware, error) {
	if strings.TrimSpace(cfg.Issuer) == "" {
		return nil, errors.New("issuer is required")
	}
	if strings.TrimSpace(cfg.JWKSURL) == "" {
		return nil, errors.New("jwks url is required")
	}
	if cfg.HTTPTimeout <= 0 {
		cfg.HTTPTimeout = 5 * time.Second
	}
	if cfg.RefreshWindow <= 0 {
		cfg.RefreshWindow = 5 * time.Minute
	}
	if cfg.Leeway <= 0 {
		cfg.Leeway = 60 * time.Second
	}

	return &Middleware{
		cfg:       cfg,
		client:    &http.Client{Timeout: cfg.HTTPTimeout},
		keysByKID: make(map[string]any),
	}, nil
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := strings.TrimSpace(r.Header.Get("Authorization"))
		if authz == "" {
			writeAuthError(w, http.StatusUnauthorized, "missing_token", "Authorization header is required")
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authz, prefix) {
			writeAuthError(w, http.StatusUnauthorized, "invalid_token", "Authorization header must be Bearer token")
			return
		}
		tokenString := strings.TrimSpace(strings.TrimPrefix(authz, prefix))
		if tokenString == "" {
			writeAuthError(w, http.StatusUnauthorized, "invalid_token", "Bearer token is empty")
			return
		}

		claims := &jwt.RegisteredClaims{}
		parserOptions := []jwt.ParserOption{
			jwt.WithValidMethods([]string{"RS256", "ES256"}),
			jwt.WithIssuer(m.cfg.Issuer),
			jwt.WithLeeway(m.cfg.Leeway),
		}
		if m.cfg.Audience != "" {
			parserOptions = append(parserOptions, jwt.WithAudience(m.cfg.Audience))
		}

		parsedToken, err := jwt.ParseWithClaims(tokenString, claims, m.keyfunc, parserOptions...)
		if err != nil || !parsedToken.Valid {
			writeAuthError(w, http.StatusUnauthorized, "invalid_token", "token validation failed")
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "invalid_token", "token subject must be UUID")
			return
		}

		ctx := ContextWithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) keyfunc(token *jwt.Token) (any, error) {
	kidRaw, ok := token.Header["kid"]
	if !ok {
		return nil, errors.New("kid not found in token header")
	}
	kid, ok := kidRaw.(string)
	if !ok || strings.TrimSpace(kid) == "" {
		return nil, errors.New("kid is invalid")
	}

	if key := m.getKey(kid); key != nil {
		return key, nil
	}

	if err := m.refreshKeys(); err != nil {
		return nil, err
	}
	if key := m.getKey(kid); key != nil {
		return key, nil
	}
	return nil, fmt.Errorf("no jwk found for kid=%s", kid)
}

func (m *Middleware) getKey(kid string) any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.lastRefresh.IsZero() && time.Since(m.lastRefresh) > m.cfg.RefreshWindow {
		return nil
	}
	return m.keysByKID[kid]
}

func (m *Middleware) refreshKeys() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.lastRefresh.IsZero() && time.Since(m.lastRefresh) <= m.cfg.RefreshWindow {
		return nil
	}

	req, err := http.NewRequest(http.MethodGet, m.cfg.JWKSURL, nil)
	if err != nil {
		return err
	}
	res, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks endpoint returned status=%d", res.StatusCode)
	}

	var doc jwksDocument
	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return err
	}

	next := make(map[string]any, len(doc.Keys))
	for _, key := range doc.Keys {
		if strings.TrimSpace(key.KID) == "" {
			continue
		}
		switch strings.ToUpper(strings.TrimSpace(key.KTY)) {
		case "RSA":
			pub, err := parseRSAPublicKey(key.N, key.E)
			if err != nil {
				continue
			}
			next[key.KID] = pub
		case "EC":
			pub, err := parseECPublicKey(key.CRV, key.X, key.Y)
			if err != nil {
				continue
			}
			next[key.KID] = pub
		}
	}

	if len(next) == 0 {
		return errors.New("no usable jwk keys in jwks response")
	}

	m.keysByKID = next
	m.lastRefresh = time.Now()
	return nil
}

func parseRSAPublicKey(nBase64URL, eBase64URL string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nBase64URL)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eBase64URL)
	if err != nil {
		return nil, err
	}
	if len(nBytes) == 0 || len(eBytes) == 0 {
		return nil, errors.New("rsa key has empty n/e")
	}

	modulus := new(big.Int).SetBytes(nBytes)
	exponent := new(big.Int).SetBytes(eBytes)
	if !exponent.IsInt64() {
		return nil, errors.New("rsa exponent too large")
	}

	return &rsa.PublicKey{N: modulus, E: int(exponent.Int64())}, nil
}

func parseECPublicKey(crv, xBase64URL, yBase64URL string) (*ecdsa.PublicKey, error) {
	var curve elliptic.Curve
	switch strings.ToUpper(strings.TrimSpace(crv)) {
	case "P-256":
		curve = elliptic.P256()
	default:
		return nil, fmt.Errorf("unsupported ec curve: %s", crv)
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(xBase64URL)
	if err != nil {
		return nil, err
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(yBase64URL)
	if err != nil {
		return nil, err
	}
	if len(xBytes) == 0 || len(yBytes) == 0 {
		return nil, errors.New("ec key has empty x/y")
	}

	x := new(big.Int).SetBytes(xBytes)
	y := new(big.Int).SetBytes(yBytes)
	if !curve.IsOnCurve(x, y) {
		return nil, errors.New("ec key point is not on curve")
	}

	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

type authErrorEnvelope struct {
	Error authErrorBody `json:"error"`
}

type authErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeAuthError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(authErrorEnvelope{
		Error: authErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDCtxKey, userID)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(userIDCtxKey).(uuid.UUID)
	return v, ok
}
