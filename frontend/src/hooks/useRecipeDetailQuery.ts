import { useQuery } from '@tanstack/react-query';
import { getRecipe } from '../api/recipes';

export const useRecipeDetailQuery = (recipeId?: string) => {
  return useQuery({
    queryKey: ['recipe', recipeId],
    queryFn: () => getRecipe(recipeId as string),
    enabled: Boolean(recipeId),
  });
};
