import { useQuery } from '@tanstack/react-query';
import { searchRecipes } from '../api/recipes';
import type { SearchRecipesParams } from '../types/api';

export const useRecipesQuery = (params: SearchRecipesParams) => {
  return useQuery({
    queryKey: ['recipes', params],
    queryFn: () => searchRecipes(params),
  });
};
