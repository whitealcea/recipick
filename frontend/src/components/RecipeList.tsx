import { Link } from 'react-router-dom';
import type { Recipe } from '../types/api';

type RecipeListProps = {
  items: Recipe[];
};

export const RecipeList = ({ items }: RecipeListProps) => {
  if (items.length === 0) {
    return <p className="card">該当するレシピはありません。</p>;
  }

  return (
    <ul className="list">
      {items.map((recipe) => (
        <li key={recipe.id} className="card">
          <h3>
            <Link to={`/recipes/${recipe.id}`}>{recipe.title}</Link>
          </h3>
          <p className="muted">{recipe.servings ?? '分量未設定'}</p>
          <p className="muted">
            調理時間: {recipe.totalMinutes != null ? `${recipe.totalMinutes}分` : '未設定'}
          </p>
          {recipe.tags.length > 0 ? (
            <p className="tags">{recipe.tags.map((tag) => tag.name).join(' / ')}</p>
          ) : null}
        </li>
      ))}
    </ul>
  );
};
