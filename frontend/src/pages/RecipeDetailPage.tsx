import { Link, useNavigate, useParams } from 'react-router-dom';
import { useDeleteRecipeMutation } from '../hooks/useDeleteRecipeMutation';
import { useRecipeDetailQuery } from '../hooks/useRecipeDetailQuery';
import { getErrorMessage } from '../lib/errors';

export const RecipeDetailPage = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const recipeQuery = useRecipeDetailQuery(id);
  const deleteMutation = useDeleteRecipeMutation();

  if (recipeQuery.isLoading) {
    return (
      <div className="page">
        <p className="card">読み込み中...</p>
      </div>
    );
  }

  if (recipeQuery.error || !recipeQuery.data) {
    return (
      <div className="page">
        <p className="error">{getErrorMessage(recipeQuery.error, 'レシピが見つかりませんでした。')}</p>
      </div>
    );
  }

  const recipe = recipeQuery.data;

  return (
    <div className="page">
      <div className="header-row">
        <h1>{recipe.title}</h1>
        <div className="actions-inline">
          <Link to={`/recipes/${recipe.id}/edit`} className="button-link">
            編集
          </Link>
          <button
            type="button"
            className="danger"
            onClick={() => {
              const ok = window.confirm('このレシピを削除しますか？');
              if (!ok) return;

              deleteMutation.mutate(recipe.id, {
                onSuccess: () => {
                  navigate('/recipes');
                },
              });
            }}
            disabled={deleteMutation.isPending}
          >
            {deleteMutation.isPending ? '削除中...' : '削除'}
          </button>
        </div>
      </div>

      {deleteMutation.error ? <p className="error">{getErrorMessage(deleteMutation.error)}</p> : null}

      <div className="card">
        <p>
          <strong>URL:</strong>{' '}
          <a href={recipe.sourceUrl} target="_blank" rel="noreferrer">
            {recipe.sourceUrl}
          </a>
        </p>
        <p>
          <strong>分量:</strong> {recipe.servings ?? '未設定'}
        </p>
        <p>
          <strong>調理時間:</strong>{' '}
          {recipe.totalMinutes != null ? `${recipe.totalMinutes}分` : '未設定'}
        </p>
        <p>
          <strong>メモ:</strong> {recipe.note?.trim() ? recipe.note : 'なし'}
        </p>

        <section>
          <h3>材料</h3>
          {recipe.ingredients.length > 0 ? (
            <ul>
              {recipe.ingredients.map((ingredient) => (
                <li key={ingredient}>{ingredient}</li>
              ))}
            </ul>
          ) : (
            <p>なし</p>
          )}
        </section>

        <section>
          <h3>タグ</h3>
          {recipe.tags.length > 0 ? (
            <ul>
              {recipe.tags.map((tag) => (
                <li key={`${tag.tagType}-${tag.name}`}>
                  {tag.name} ({tag.tagType})
                </li>
              ))}
            </ul>
          ) : (
            <p>なし</p>
          )}
        </section>
      </div>
    </div>
  );
};
