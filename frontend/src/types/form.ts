import type { TagType } from './api';

export type FormTag = {
  name: string;
  tagType: TagType;
};

export type RecipeFormValues = {
  sourceUrl: string;
  title: string;
  note: string;
  servings: string;
  totalMinutes: string;
  ingredientsText: string;
  tags: FormTag[];
  sourceRecipeJsonLd?: unknown;
};

export type RecipeSearchFormValues = {
  q: string;
  tags: string;
  ingredients: string;
  maxMinutes: string;
};
