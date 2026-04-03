import { apiClient } from './client';
import type {
  CreateRecipeRequest,
  ImportPreview,
  ImportPreviewRequest,
  Recipe,
  SearchRecipesParams,
  SearchRecipesResponse,
  UpdateRecipeRequest,
} from '../types/api';

const toSearchParams = (params: SearchRecipesParams): string => {
  const search = new URLSearchParams();

  if (params.q) search.set('q', params.q);
  if (params.tags) search.set('tags', params.tags);
  if (params.ingredients) search.set('ingredients', params.ingredients);
  if (typeof params.maxMinutes === 'number') search.set('maxMinutes', String(params.maxMinutes));
  if (typeof params.limit === 'number') search.set('limit', String(params.limit));
  if (typeof params.offset === 'number') search.set('offset', String(params.offset));

  const query = search.toString();
  return query ? `?${query}` : '';
};

export const searchRecipes = async (
  params: SearchRecipesParams,
): Promise<SearchRecipesResponse> => {
  return apiClient<SearchRecipesResponse>(`/recipes${toSearchParams(params)}`);
};

export const getRecipe = async (id: string): Promise<Recipe> => {
  return apiClient<Recipe>(`/recipes/${id}`);
};

export const importRecipePreview = async (
  payload: ImportPreviewRequest,
): Promise<ImportPreview> => {
  return apiClient<ImportPreview>('/recipes/import', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
};

export const createRecipe = async (payload: CreateRecipeRequest): Promise<Recipe> => {
  return apiClient<Recipe>('/recipes', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
};

export const updateRecipe = async (
  id: string,
  payload: UpdateRecipeRequest,
): Promise<Recipe> => {
  return apiClient<Recipe>(`/recipes/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  });
};

export const deleteRecipe = async (id: string): Promise<void> => {
  return apiClient<void>(`/recipes/${id}`, { method: 'DELETE' });
};
