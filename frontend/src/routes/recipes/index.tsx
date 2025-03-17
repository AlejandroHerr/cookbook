import { createFileRoute } from '@tanstack/react-router';

import { recipesListQueryOptions } from '@/api/recipes';
import {
  PendingRecipeListPage,
  RecipesListPage,
} from '@/features/recipes/list/RecipesListPage';

export const Route = createFileRoute('/recipes/')({
  component: RecipesListPage,
  loader: ({ context: { queryClient } }) =>
    queryClient.ensureQueryData(recipesListQueryOptions()),
  pendingComponent: PendingRecipeListPage,
});
