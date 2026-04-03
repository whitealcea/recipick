import { useMemo } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { RecipeList } from '../components/RecipeList';
import { SearchForm } from '../components/SearchForm';
import { useRecipesQuery } from '../hooks/useRecipesQuery';
import { getErrorMessage } from '../lib/errors';
import type { RecipeSearchFormValues } from '../types/form';

const readInitialValues = (searchParams: URLSearchParams): RecipeSearchFormValues => {
  return {
    q: searchParams.get('q') ?? '',
    tags: searchParams.get('tags') ?? '',
    ingredients: searchParams.get('ingredients') ?? '',
    maxMinutes: searchParams.get('maxMinutes') ?? '',
  };
};

export const RecipeListPage = () => {
  const [searchParams, setSearchParams] = useSearchParams();

  const initialValues = useMemo(() => readInitialValues(searchParams), [searchParams]);
  const maxMinutes = Number(initialValues.maxMinutes.trim());

  const recipesQuery = useRecipesQuery({
    q: initialValues.q.trim() || undefined,
    tags: initialValues.tags.trim() || undefined,
    ingredients: initialValues.ingredients.trim() || undefined,
    maxMinutes:
      initialValues.maxMinutes.trim() && Number.isFinite(maxMinutes) ? maxMinutes : undefined,
    limit: 20,
    offset: 0,
  });

  return (
    <div className="page">
      <div className="header-row">
        <h1>レシピ一覧</h1>
        <Link to="/recipes/new" className="button-link">
          新規作成
        </Link>
      </div>

      <SearchForm
        initialValues={initialValues}
        onSearch={(values) => {
          const next = new URLSearchParams();
          if (values.q.trim()) next.set('q', values.q.trim());
          if (values.tags.trim()) next.set('tags', values.tags.trim());
          if (values.ingredients.trim()) next.set('ingredients', values.ingredients.trim());
          if (values.maxMinutes.trim()) next.set('maxMinutes', values.maxMinutes.trim());
          setSearchParams(next);
        }}
      />

      {recipesQuery.error ? <p className="error">{getErrorMessage(recipesQuery.error)}</p> : null}

      {recipesQuery.isLoading ? (
        <p className="card">読み込み中...</p>
      ) : (
        <RecipeList items={recipesQuery.data?.items ?? []} />
      )}
    </div>
  );
};
