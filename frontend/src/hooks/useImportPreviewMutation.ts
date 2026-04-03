import { useMutation } from '@tanstack/react-query';
import { importRecipePreview } from '../api/recipes';

export const useImportPreviewMutation = () => {
  return useMutation({
    mutationFn: importRecipePreview,
  });
};
