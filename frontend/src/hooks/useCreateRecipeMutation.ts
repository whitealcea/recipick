import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createRecipe } from '../api/recipes';

export const useCreateRecipeMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createRecipe,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['recipes'] });
    },
  });
};
