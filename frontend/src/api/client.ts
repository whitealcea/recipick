import type { ApiError } from '../types/api';
import { getAccessToken } from '../lib/auth';

const getDefaultBaseURL = (): string => {
  if (typeof window === 'undefined') {
    return 'http://localhost:8080/v1';
  }
  return `http://${window.location.hostname}:8080/v1`;
};

const baseURL = import.meta.env.VITE_API_BASE_URL ?? getDefaultBaseURL();

export class ApiRequestError extends Error {
  status: number;
  payload?: ApiError;

  constructor(message: string, status: number, payload?: ApiError) {
    super(message);
    this.name = 'ApiRequestError';
    this.status = status;
    this.payload = payload;
  }
}

const buildUrl = (path: string): string => {
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path;
  }
  return `${baseURL.replace(/\/$/, '')}/${path.replace(/^\//, '')}`;
};

const parseErrorPayload = async (response: Response): Promise<ApiError | undefined> => {
  try {
    const json = (await response.json()) as ApiError;
    if (json && typeof json === 'object' && 'error' in json) {
      return json;
    }
    return undefined;
  } catch {
    return undefined;
  }
};

export const apiClient = async <TResponse>(
  path: string,
  init: RequestInit = {},
): Promise<TResponse> => {
  const accessToken = getAccessToken();

  const response = await fetch(buildUrl(path), {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      ...(init.headers ?? {}),
    },
  });

  if (!response.ok) {
    const payload = await parseErrorPayload(response);
    throw new ApiRequestError(
      payload?.error?.message ?? `Request failed with status ${response.status}`,
      response.status,
      payload,
    );
  }

  if (response.status === 204) {
    return undefined as TResponse;
  }

  return (await response.json()) as TResponse;
};
