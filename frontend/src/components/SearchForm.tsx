import { useEffect, useState } from 'react';
import type { RecipeSearchFormValues } from '../types/form';

type SearchFormProps = {
  initialValues: RecipeSearchFormValues;
  onSearch: (values: RecipeSearchFormValues) => void;
};

export const SearchForm = ({ initialValues, onSearch }: SearchFormProps) => {
  const [values, setValues] = useState<RecipeSearchFormValues>(initialValues);

  useEffect(() => {
    setValues(initialValues);
  }, [initialValues]);

  return (
    <form
      className="card form"
      onSubmit={(event) => {
        event.preventDefault();
        onSearch(values);
      }}
    >
      <h2>検索</h2>

      <div className="field-row">
        <div className="field">
          <label htmlFor="q">キーワード</label>
          <input
            id="q"
            value={values.q}
            onChange={(event) => setValues((prev) => ({ ...prev, q: event.target.value }))}
            placeholder="タイトル・メモ"
          />
        </div>

        <div className="field">
          <label htmlFor="maxMinutes">最大調理時間（分）</label>
          <input
            id="maxMinutes"
            inputMode="numeric"
            value={values.maxMinutes}
            onChange={(event) => setValues((prev) => ({ ...prev, maxMinutes: event.target.value }))}
            placeholder="30"
          />
        </div>
      </div>

      <div className="field-row">
        <div className="field">
          <label htmlFor="tags">タグ</label>
          <input
            id="tags"
            value={values.tags}
            onChange={(event) => setValues((prev) => ({ ...prev, tags: event.target.value }))}
            placeholder="和食,平日夜"
          />
        </div>

        <div className="field">
          <label htmlFor="ingredients">材料</label>
          <input
            id="ingredients"
            value={values.ingredients}
            onChange={(event) =>
              setValues((prev) => ({ ...prev, ingredients: event.target.value }))
            }
            placeholder="玉ねぎ,鶏もも肉"
          />
        </div>
      </div>

      <div className="actions">
        <button type="submit">検索</button>
      </div>
    </form>
  );
};
