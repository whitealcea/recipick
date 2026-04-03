const ACCESS_TOKEN_STORAGE_KEY = 'recipick.accessToken';
const STATIC_ACCESS_TOKEN = import.meta.env.VITE_AUTH_ACCESS_TOKEN?.trim() ?? '';

export const hasStaticAccessToken = (): boolean => STATIC_ACCESS_TOKEN.length > 0;

export const getAccessToken = (): string | null => {
  if (STATIC_ACCESS_TOKEN.length > 0) {
    return STATIC_ACCESS_TOKEN;
  }

  if (typeof window === 'undefined') return null;
  const stored = window.localStorage.getItem(ACCESS_TOKEN_STORAGE_KEY);
  if (stored && stored.trim().length > 0) {
    return stored.trim();
  }
  return null;
};

export const setAccessToken = (token: string): void => {
  if (typeof window === 'undefined') return;
  window.localStorage.setItem(ACCESS_TOKEN_STORAGE_KEY, token);
};

export const clearAccessToken = (): void => {
  if (typeof window === 'undefined') return;
  window.localStorage.removeItem(ACCESS_TOKEN_STORAGE_KEY);
};
