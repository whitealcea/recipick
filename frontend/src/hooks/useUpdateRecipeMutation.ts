import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updateRecipe } from '../api/recipes';
import type { UpdateRecipeRequest } from '../types/api';

type UpdateRecipeVariables = {
  id: string;
  payload: UpdateRecipeRequest;
};

export const useUpdateRecipeMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: UpdateRecipeVariables) => updateRecipe(id, payload),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['recipe', variables.id] });
      queryClient.invalidateQueries({ queryKey: ['recipes'] });
    },
  });
};
