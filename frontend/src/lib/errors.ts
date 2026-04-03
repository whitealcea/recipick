import { ApiRequestError } from '../api/client';

export const getErrorMessage = (error: unknown, fallback = 'エラーが発生しました'): string => {
  if (error instanceof ApiRequestError) {
    return error.payload?.error?.message ?? error.message;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return fallback;
};
