import { useMutation, useQueryClient } from '@tanstack/react-query';
import { deleteRecipe } from '../api/recipes';

export const useDeleteRecipeMutation = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deleteRecipe(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['recipes'] });
    },
  });
};
