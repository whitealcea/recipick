import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { RecipeForm } from '../components/RecipeForm';
import { useCreateRecipeMutation } from '../hooks/useCreateRecipeMutation';
import { useImportPreviewMutation } from '../hooks/useImportPreviewMutation';
import { getErrorMessage } from '../lib/errors';
import type { CreateRecipeRequest, ImportPreview } from '../types/api';
import type { RecipeFormValues } from '../types/form';

const initialFormValues: RecipeFormValues = {
  sourceUrl: '',
  title: '',
  note: '',
  servings: '',
  totalMinutes: '',
  ingredientsText: '',
  tags: [],
};

const mergePreview = (current: RecipeFormValues, preview: ImportPreview): RecipeFormValues => {
  const systemTags = preview.systemTags.map((name) => ({ name, tagType: 'system' as const }));
  return {
    ...current,
    sourceUrl: preview.sourceUrl ?? current.sourceUrl,
    title: preview.title ?? current.title,
    servings: preview.servings ?? current.servings,
    totalMinutes:
      preview.totalMinutes != null ? String(preview.totalMinutes) : current.totalMinutes,
    ingredientsText:
      preview.ingredients.length > 0
        ? preview.ingredients.join('\n')
        : current.ingredientsText,
    tags: systemTags.length > 0 ? systemTags : current.tags,
    sourceRecipeJsonLd: preview.rawJsonLd,
  };
};

const toCreatePayload = (values: RecipeFormValues): CreateRecipeRequest => {
  const ingredients = values.ingredientsText
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean);

  const tags = values.tags
    .map((tag) => ({ name: tag.name.trim(), tagType: tag.tagType }))
    .filter((tag) => tag.name.length > 0);

  const totalMinutesText = values.totalMinutes.trim();
  const totalMinutesParsed = Number(totalMinutesText);

  return {
    sourceUrl: values.sourceUrl.trim(),
    title: values.title.trim(),
    note: values.note.trim() || null,
    servings: values.servings.trim() || null,
    totalMinutes:
      totalMinutesText.length > 0 && Number.isFinite(totalMinutesParsed)
        ? totalMinutesParsed
        : null,
    ingredients: ingredients.length > 0 ? ingredients : undefined,
    tags: tags.length > 0 ? tags : undefined,
    sourceRecipeJsonLd: values.sourceRecipeJsonLd,
  };
};

export const RecipeNewPage = () => {
  const navigate = useNavigate();
  const [values, setValues] = useState<RecipeFormValues>(initialFormValues);
  const [pageError, setPageError] = useState<string | null>(null);

  const importPreviewMutation = useImportPreviewMutation();
  const createMutation = useCreateRecipeMutation();

  return (
    <div className="page">
      <RecipeForm
        title="レシピ新規作成"
        values={values}
        submitLabel="保存"
        isSubmitting={createMutation.isPending}
        isImporting={importPreviewMutation.isPending}
        errorMessage={
          pageError ??
          (createMutation.error ? getErrorMessage(createMutation.error) : null) ??
          (importPreviewMutation.error ? getErrorMessage(importPreviewMutation.error) : null)
        }
        onChange={(next) => {
          setPageError(null);
          setValues(next);
        }}
        onImportPreview={() => {
          const sourceUrl = values.sourceUrl.trim();
          if (!sourceUrl) {
            setPageError('URLを入力してください。');
            return;
          }

          importPreviewMutation.mutate(
            { sourceUrl },
            {
              onSuccess: (preview) => {
                setPageError(null);
                setValues((current) => mergePreview(current, preview));
              },
              onError: (error) => {
                setPageError(getErrorMessage(error));
              },
            },
          );
        }}
        onSubmit={() => {
          const payload = toCreatePayload(values);
          if (!payload.sourceUrl || !payload.title) {
            setPageError('URLとタイトルは必須です。');
            return;
          }

          createMutation.mutate(payload, {
            onSuccess: (recipe) => {
              navigate(`/recipes/${recipe.id}`);
            },
            onError: (error) => {
              setPageError(getErrorMessage(error));
            },
          });
        }}
      />
    </div>
  );
};
