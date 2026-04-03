import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { RecipeForm } from '../components/RecipeForm';
import { useRecipeDetailQuery } from '../hooks/useRecipeDetailQuery';
import { useUpdateRecipeMutation } from '../hooks/useUpdateRecipeMutation';
import { getErrorMessage } from '../lib/errors';
import type { Recipe, UpdateRecipeRequest } from '../types/api';
import type { RecipeFormValues } from '../types/form';

const toFormValues = (recipe: Recipe): RecipeFormValues => {
  return {
    sourceUrl: recipe.sourceUrl,
    title: recipe.title,
    note: recipe.note ?? '',
    servings: recipe.servings ?? '',
    totalMinutes: recipe.totalMinutes != null ? String(recipe.totalMinutes) : '',
    ingredientsText: recipe.ingredients.join('\n'),
    tags: recipe.tags,
    sourceRecipeJsonLd: recipe.sourceRecipeJsonLd,
  };
};

const toUpdatePayload = (values: RecipeFormValues): UpdateRecipeRequest => {
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
    ingredients,
    tags,
    sourceRecipeJsonLd: values.sourceRecipeJsonLd,
  };
};

export const RecipeEditPage = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [values, setValues] = useState<RecipeFormValues | null>(null);
  const [pageError, setPageError] = useState<string | null>(null);

  const detailQuery = useRecipeDetailQuery(id);
  const updateMutation = useUpdateRecipeMutation();

  useEffect(() => {
    if (detailQuery.data) {
      setValues(toFormValues(detailQuery.data));
    }
  }, [detailQuery.data]);

  if (detailQuery.isLoading || !values) {
    return (
      <div className="page">
        <p className="card">読み込み中...</p>
      </div>
    );
  }

  if (detailQuery.error || !id) {
    return (
      <div className="page">
        <p className="error">{getErrorMessage(detailQuery.error, 'レシピが見つかりません。')}</p>
      </div>
    );
  }

  return (
    <div className="page">
      <RecipeForm
        title="レシピ編集"
        values={values}
        submitLabel="更新"
        isSubmitting={updateMutation.isPending}
        errorMessage={
          pageError ??
          (updateMutation.error ? getErrorMessage(updateMutation.error) : null)
        }
        onChange={(next) => {
          setPageError(null);
          setValues(next);
        }}
        onSubmit={() => {
          const payload = toUpdatePayload(values);
          if (!payload.sourceUrl || !payload.title) {
            setPageError('URLとタイトルは必須です。');
            return;
          }

          updateMutation.mutate(
            { id, payload },
            {
              onSuccess: (recipe) => {
                navigate(`/recipes/${recipe.id}`);
              },
              onError: (error) => {
                setPageError(getErrorMessage(error));
              },
            },
          );
        }}
      />
    </div>
  );
};
