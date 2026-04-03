import type { FormEvent } from 'react';
import { IngredientInput } from './IngredientInput';
import { TagInput } from './TagInput';
import type { RecipeFormValues } from '../types/form';

type RecipeFormProps = {
  title: string;
  values: RecipeFormValues;
  submitLabel: string;
  isSubmitting?: boolean;
  isImporting?: boolean;
  errorMessage?: string | null;
  onChange: (next: RecipeFormValues) => void;
  onSubmit: () => void;
  onImportPreview?: () => void;
};

const onTextChange = (
  values: RecipeFormValues,
  onChange: (next: RecipeFormValues) => void,
  key: keyof RecipeFormValues,
  value: string,
) => {
  onChange({ ...values, [key]: value });
};

export const RecipeForm = ({
  title,
  values,
  submitLabel,
  isSubmitting = false,
  isImporting = false,
  errorMessage,
  onChange,
  onSubmit,
  onImportPreview,
}: RecipeFormProps) => {
  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    onSubmit();
  };

  return (
    <form className="card form" onSubmit={handleSubmit}>
      <h2>{title}</h2>

      {errorMessage ? <p className="error">{errorMessage}</p> : null}

      <div className="field">
        <label htmlFor="sourceUrl">レシピURL</label>
        <input
          id="sourceUrl"
          value={values.sourceUrl}
          onChange={(event) => onTextChange(values, onChange, 'sourceUrl', event.target.value)}
          placeholder="https://example.com/recipe"
        />
      </div>

      {onImportPreview ? (
        <div className="actions">
          <button
            type="button"
            onClick={onImportPreview}
            disabled={isImporting || values.sourceUrl.trim().length === 0}
          >
            {isImporting ? 'Import中...' : 'Import preview'}
          </button>
        </div>
      ) : null}

      <div className="field">
        <label htmlFor="title">タイトル</label>
        <input
          id="title"
          value={values.title}
          onChange={(event) => onTextChange(values, onChange, 'title', event.target.value)}
          placeholder="肉じゃが"
        />
      </div>

      <div className="field">
        <label htmlFor="note">メモ</label>
        <textarea
          id="note"
          rows={4}
          value={values.note}
          onChange={(event) => onTextChange(values, onChange, 'note', event.target.value)}
        />
      </div>

      <div className="field-row">
        <div className="field">
          <label htmlFor="servings">分量</label>
          <input
            id="servings"
            value={values.servings}
            onChange={(event) => onTextChange(values, onChange, 'servings', event.target.value)}
            placeholder="2人分"
          />
        </div>

        <div className="field">
          <label htmlFor="totalMinutes">調理時間（分）</label>
          <input
            id="totalMinutes"
            inputMode="numeric"
            value={values.totalMinutes}
            onChange={(event) => onTextChange(values, onChange, 'totalMinutes', event.target.value)}
            placeholder="35"
          />
        </div>
      </div>

      <IngredientInput
        value={values.ingredientsText}
        onChange={(next) => onChange({ ...values, ingredientsText: next })}
      />

      <TagInput value={values.tags} onChange={(next) => onChange({ ...values, tags: next })} />

      <div className="actions">
        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? '保存中...' : submitLabel}
        </button>
      </div>
    </form>
  );
};
