import { createFileRoute } from '@tanstack/react-router';

import { CreateRecipePage } from '@/features/recipes/create/CreateRecipesPage';

export const Route = createFileRoute('/recipes/create')({
  component: CreateRecipePage,
  validateSearch: (search) => ({
    fromURL: search.fromURL === true ? true : undefined,
  }),
});
