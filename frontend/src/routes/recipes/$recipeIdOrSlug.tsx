import { useSuspenseQuery } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import { Clock } from 'lucide-react';

import { recipeQueryOptions } from '@/api/recipes';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { Unit } from '@/types/recipes/units';

export const Route = createFileRoute('/recipes/$recipeIdOrSlug')({
  component: RouteComponent,
  loader: ({ context: { queryClient }, params: { recipeIdOrSlug } }) =>
    queryClient.ensureQueryData(
      recipeQueryOptions({
        recipeIdOrSlug: recipeIdOrSlug,
      }),
    ),
});

function RouteComponent() {
  const recipeIdOrSlug = Route.useParams().recipeIdOrSlug;

  const { data } = useSuspenseQuery(
    recipeQueryOptions({ recipeIdOrSlug: recipeIdOrSlug }),
  );

  const steps = data.steps?.split('\n') ?? [];

  return (
    <div className="flex flex-col">
      <div className="flex flex-col gap-4 mb-4">
        <h2 className="scroll-m-20 border-b pb-2 text-3xl font-semibold tracking-tight first:mt-0">
          {data.title}
        </h2>
        <div className="flex gap-2">
          {!!data.prepTime && (
            <div
              className={cn('flex shrink-0 align-middle pr-2', {
                'border-r-2': !!data.tags.length,
              })}
            >
              <Badge variant="secondary" className="gap-1">
                <Clock size={12} /> {data.prepTime} minutes
              </Badge>
            </div>
          )}
          {!!data.tags.length && (
            <div className="flex-1 flex flex-wrap gap-1">
              {data.tags.map((tag) => (
                <Badge key={tag}>{tag}</Badge>
              ))}
            </div>
          )}
        </div>
      </div>
      <div
        className="flex flex-col gap-4 mb-8 overflow-scroll"
        style={{ scrollSnapType: 'y mandatory' }}
      >
        <div
          className="p-4"
          style={{
            scrollSnapAlign: 'start',
          }}
        >
          <blockquote className="border-l-2 pl-6 italic text-muted-foreground text-xl">
            {data.headline}
          </blockquote>
        </div>
        <div
          style={{
            scrollSnapAlign: 'start',
          }}
        >
          <p className="leading-7 not-first:mt-6">{data.description}</p>
        </div>
        <div
          className="grid grid-cols-3 lg:grid-cols-4 gap-4"
          style={{
            scrollSnapAlign: 'start',
          }}
        >
          <div className="col-span-1">
            <h3 className="scroll-m-20 text-2xl font-semibold tracking-tight border-b">
              Ingredients
            </h3>
            <div className="flex flex-col gap-2 py-2">
              <div className="italic text-sm">(Serves {data.servings})</div>
              <ul className="list-disc list-inside [&>li]:mt-2">
                {data.ingredients.map((ingredient) => (
                  <li
                    key={ingredient.id}
                    className="text-sm font-medium leading-5"
                  >
                    {ingredient.unit === Unit.Uncountable
                      ? 'some '
                      : `${ingredient.amount} ${ingredient.unit} of `}
                    <span className="font-bold">{ingredient.name}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
          <div className="col-span-2 lg:col-span-3">
            <h3 className="scroll-m-20 text-2xl font-semibold tracking-tight border-b">
              Steps
            </h3>
            <div className="py-2 gap-2 flex flex-col">
              <ol className="list-inside my-6 ml-6 list-decimal">
                {steps.map((step, i) => (
                  <li key={i} className="">
                    {step}
                  </li>
                ))}
                {steps.map((step, i) => (
                  <li key={i} className="text-lg leading-7">
                    {step}
                  </li>
                ))}
              </ol>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
