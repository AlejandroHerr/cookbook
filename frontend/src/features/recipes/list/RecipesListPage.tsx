import { useSuspenseQuery } from '@tanstack/react-query';
import { Clock, Globe } from 'lucide-react';
import { PropsWithChildren } from 'react';

import { recipesListQueryOptions } from '@/api/recipes';
import { Badge } from '@/components/ui/badge';
import { ButtonLink } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
  LinkedCard,
} from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { RecipeWithoutIngredients } from '@/types/recipes';

const SkeletonCard = () => (
  <Card>
    <CardHeader>
      <Skeleton className="h-4 w-full" />
    </CardHeader>
    <CardContent className="flex flex-col gap-2">
      <Skeleton className="h-4" />
      <Skeleton className="h-4" />
      <Skeleton className="h-4" />
    </CardContent>
    <CardFooter className="flex gap-1 items-start">
      <Skeleton className="h-4 w-16" />
      <Skeleton className="h-4 w-16" />
      <Skeleton className="h-4 w-16" />
      <Skeleton className="h-4 w-16" />
      <Skeleton className="h-4 w-16" />
    </CardFooter>
  </Card>
);

interface RecipeCardProps {
  recipe: RecipeWithoutIngredients;
}

const RecipeCard: React.FC<RecipeCardProps> = ({ recipe }) => {
  return (
    <LinkedCard
      className="cursor-pointer justify-between"
      style={{
        scrollSnapAlign: 'start',
      }}
      to={'/recipes/$recipeIdOrSlug'}
      params={{ recipeIdOrSlug: recipe.slug }}
    >
      <CardHeader>
        <CardTitle>{recipe.title}</CardTitle>
        {!!recipe.prepTime && (
          <CardDescription className="flex gap-1 items-center">
            <Clock size={12} /> {recipe.prepTime} minutes
          </CardDescription>
        )}
      </CardHeader>
      <CardContent>{recipe.headline}</CardContent>
      <CardFooter className="flex flex-col gap-1 items-start">
        <div className="flex flex-wrap gap-1">
          {recipe.tags.map((tag) => (
            <Badge key={tag}>{tag}</Badge>
          ))}
        </div>
      </CardFooter>
    </LinkedCard>
  );
};

export const ListLayout: React.FC<PropsWithChildren> = ({ children }) => {
  return (
    <div className="gap-6 flex flex-col">
      <div className="flex justify-between border-b-2">
        <div>
          <h2 className="scroll-m-20 pb-2 text-3xl font-semibold tracking-tight first:mt-0">
            Recipes
          </h2>
        </div>
        <div className="flex justify-between gap-2">
          <ButtonLink
            to="/recipes/create"
            search={{ fromURL: true }}
            variant="secondary"
            size="sm"
          >
            <Globe /> Import from URL
          </ButtonLink>
          <ButtonLink
            to="/recipes/create"
            search={{ fromURL: undefined }}
            size="sm"
          >
            Create Recipe
          </ButtonLink>
        </div>
      </div>
      <div
        className="p-6 overflow-scroll grid grid-cols-1 sm:grid-cols-2  md:grid-cols-3 lg:grid-cols-4 gap-4"
        style={{ scrollSnapType: 'y mandatory' }}
      >
        {children}
      </div>
    </div>
  );
};

export const PendingRecipeListPage = () => (
  <ListLayout>
    <SkeletonCard />
    <SkeletonCard />
    <SkeletonCard />
    <SkeletonCard />
  </ListLayout>
);

export const RecipesListPage = () => {
  const { data } = useSuspenseQuery(recipesListQueryOptions());

  return (
    <ListLayout>
      {data.recipes.map((recipe) => (
        <RecipeCard key={recipe.id} recipe={recipe} />
      ))}
    </ListLayout>
  );
};
