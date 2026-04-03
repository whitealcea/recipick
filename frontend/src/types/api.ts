export type TagType = 'system' | 'user';

export type RecipeTag = {
  name: string;
  tagType: TagType;
};

export type Recipe = {
  id: string;
  sourceUrl: string;
  title: string;
  note: string | null;
  servings: string | null;
  totalMinutes: number | null;
  ingredients: string[];
  tags: RecipeTag[];
  sourceRecipeJsonLd?: unknown;
  createdAt: string;
  updatedAt: string;
};

export type ImportPreview = {
  sourceUrl: string;
  title: string | null;
  servings: string | null;
  totalMinutes: number | null;
  ingredients: string[];
  systemTags: string[];
  rawJsonLd?: unknown;
  jsonLdFound: boolean;
  warnings: string[];
};

export type SearchRecipesResponse = {
  items: Recipe[];
};

export type ApiError = {
  error: {
    code: string;
    message: string;
    details?: unknown;
  };
};

export type ImportPreviewRequest = {
  sourceUrl: string;
};

export type CreateRecipeRequest = {
  sourceUrl: string;
  title: string;
  note?: string | null;
  servings?: string | null;
  totalMinutes?: number | null;
  ingredients?: string[];
  tags?: RecipeTag[];
  sourceRecipeJsonLd?: unknown;
};

export type UpdateRecipeRequest = Partial<CreateRecipeRequest>;

export type SearchRecipesParams = {
  q?: string;
  tags?: string;
  ingredients?: string;
  maxMinutes?: number;
  limit?: number;
  offset?: number;
};
