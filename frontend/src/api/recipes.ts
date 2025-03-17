import {
  DefaultError,
  MutationOptions,
  queryOptions,
} from '@tanstack/react-query';
import axios from 'axios';

import {
  CreateRecipeResponse,
  CreateUpdateRecipeRequest,
  GetRecipesResponse,
  Recipe,
} from '@/types/recipes';

export const recipesListQueryOptions = () =>
  queryOptions({
    queryKey: ['recipes'],
    queryFn: () =>
      axios
        .get<GetRecipesResponse>('http://localhost:8080/recipes')
        .then((res) => res.data),
  });

export const recipeQueryOptions = ({
  recipeIdOrSlug,
}: {
  recipeIdOrSlug: string;
}) =>
  queryOptions({
    queryKey: ['recipe', recipeIdOrSlug],
    queryFn: () =>
      axios
        .get<Recipe>(`http://localhost:8080/recipes/${recipeIdOrSlug}`)
        .then((res) => res.data),
  });

export const createRecipeMutationOptions: MutationOptions<
  CreateRecipeResponse,
  DefaultError,
  CreateUpdateRecipeRequest
> = {
  mutationFn: async (data: CreateUpdateRecipeRequest) =>
    axios
      .post<CreateRecipeResponse>('http://localhost:8080/recipes', data)
      .then((res) => res.data),
};
