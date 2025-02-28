import { DevTool } from '@hookform/devtools';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate, useSearch } from '@tanstack/react-router';
import { Plus, X } from 'lucide-react';
import React, { useCallback, useEffect } from 'react';
import { Control, UseFormSetValue, useWatch } from 'react-hook-form';
import { toast } from 'sonner';

import { ImportFromURL } from './ImportFromURL';
import { RecipeFormSchema, useRecipeForm } from './useRecipeForm';

import { createRecipeMutationOptions } from '@/api/recipes';
import { tagsCompletionsQueryOptions } from '@/api/suggestions';
import { FormElement, FormSelectElement } from '@/components/form/FormElement';
import { Button, ButtonLink } from '@/components/ui/button';
import { Form } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  MultipleSelector,
  MultiSelectOption,
} from '@/components/ui/multiselector';
import { Textarea } from '@/components/ui/textarea';
import { CompleteRecipeResponse } from '@/types/completions';
import { CreateUpdateRecipeRequest } from '@/types/recipes';
import { Unit, DefaultUnit, UnitOptions } from '@/types/recipes/units';

interface NewType {
  control: Control<RecipeFormSchema>;
  disabled?: boolean;
  index: number;
  remove: (index: number) => void;
  update: UseFormSetValue<RecipeFormSchema>;
}

const Foo: React.FC<NewType> = ({
  control,
  disabled,
  index,
  remove,
  update,
}) => {
  const selectedUnit = useWatch({
    control,
    name: `ingredients.${index}.unit`,
  });

  useEffect(() => {
    if (selectedUnit === Unit.Uncountable) {
      update(`ingredients.${index}.quantity`, 0);
    }
  }, [index, selectedUnit, update]);

  return (
    <>
      <div className="md:col-span-2 sm:col-span-1">
        <FormElement
          control={control}
          fieldName={`ingredients.${index}.quantity`}
          render={({ field }) => (
            <Input
              {...field}
              disabled={!!disabled || selectedUnit === Unit.Uncountable}
              type="number"
            />
          )}
        />
      </div>
      <div className="md:col-span-2 sm:col-span-">
        <FormSelectElement
          control={control}
          fieldName={`ingredients.${index}.unit`}
          options={UnitOptions}
          disabled={disabled}
        />
      </div>
      <div className="sm:col-span-2 md:col-span-7">
        <FormElement
          control={control}
          fieldName={`ingredients.${index}.name`}
          render={({ field }) => <Input {...field} disabled={disabled} />}
        />
      </div>
      <div className="md:col-span-1 sm:col-span-1">
        <Button
          variant="destructive"
          size="icon"
          onClick={() => {
            remove(index);
          }}
        >
          <X />
        </Button>
      </div>
    </>
  );
};

export const CreateRecipePage: React.FC = () => {
  const { fromURL } = useSearch({
    from: '/recipes/create',
  });

  const { form, ingredientsField } = useRecipeForm();

  const queryClient = useQueryClient();
  const fetchTagsSuggestions = useCallback(
    async (search: string) => {
      const options = await queryClient.ensureQueryData(
        tagsCompletionsQueryOptions({ search }),
      );

      return options as MultiSelectOption[];
    },
    [queryClient],
  );

  const navigate = useNavigate();
  const closeImportFromURL = useCallback(() => {
    void navigate({
      search: undefined,
      replace: true,
    });
  }, [navigate]);

  const onFinish = useCallback(
    (data: CompleteRecipeResponse & { url: string }) => {
      form.reset({
        title: data.recipe.title,
        headline: data.recipe.headline,
        description: data.recipe.description,
        steps: data.recipe.steps.join('\n'),
        prepTime: data.recipe.prepTime,
        servings: data.recipe.servings,
        tags: data.recipe.tags.map((tag) => ({ label: tag, value: tag })),
        url: data.url,
        ingredients: data.recipe.ingredients.map((ingredient) => {
          return {
            name: ingredient.name,
            quantity: ingredient.quantity,
            unit:
              UnitOptions.find((unit) => ingredient.unit === unit.value)
                ?.value ?? DefaultUnit,
          };
        }),
      });

      closeImportFromURL();
    },
    [closeImportFromURL, form],
  );
  const { mutate } = useMutation({
    ...createRecipeMutationOptions,
    onError: (error) => {
      console.error(error);
      toast.error('Failed to create recipe');
    },
    onSuccess: (data) => {
      toast.success('Recipe created');
      void navigate({ to: `/recipes/${data.slug}` });
    },
  });

  const onSubmit = form.handleSubmit((values) => {
    const ingredients = values.ingredients.reduce<
      CreateUpdateRecipeRequest['ingredients']
    >((acc, ingredient) => {
      if (
        !ingredient?.name ||
        (!ingredient.quantity && ingredient.unit !== Unit.Uncountable) ||
        !ingredient.unit
      ) {
        return acc;
      }

      if (ingredient.unit === Unit.Uncountable) {
        acc.push({
          name: ingredient.name,
          quantity: 0,
          unit: ingredient.unit,
        });
      } else {
        acc.push({
          name: ingredient.name,
          // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
          quantity: ingredient.quantity!,
          unit: ingredient.unit,
        });
      }

      return acc;
    }, []);

    mutate({
      title: values.title,
      url: values.url,
      tags: values.tags.map((tag) => tag.value),
      ingredients: ingredients,
      servings: values.servings,
      description: values.description,
      headline: values.headline,
      steps: values.steps,
      prepTime: values.prepTime,
    });
  });

  return (
    <div className="relative max-w-3xl mx-auto h-full px-4">
      <Form {...form}>
        {/* eslint-disable-next-line @typescript-eslint/no-misused-promises */}
        <form onSubmit={onSubmit} className="pb-16">
          <div className="py-12 flex flex-col gap-4 border-b">
            <FormElement
              control={form.control}
              fieldName="title"
              label="Title"
              render={({ field }) => (
                <Input
                  {...field}
                  placeholder="Pizza Margherita"
                  disabled={form.formState.isSubmitting}
                />
              )}
            />
            <FormElement
              control={form.control}
              fieldName="headline"
              label="Headline (Optional)"
              render={({ field }) => (
                <Input
                  {...field}
                  placeholder="The classic Italian pizza"
                  disabled={form.formState.isSubmitting}
                />
              )}
            />
            <FormElement
              control={form.control}
              fieldName="url"
              label="URL (Optional)"
              render={({ field }) => (
                <Input
                  {...field}
                  placeholder="https://www.example.com/recipe"
                  disabled={form.formState.isSubmitting}
                />
              )}
            />
            <FormElement
              control={form.control}
              fieldName="description"
              label="Description (Optional)"
              render={({ field }) => (
                <Textarea
                  className="resize-none"
                  disabled={form.formState.isSubmitting}
                  {...field}
                  placeholder="A classic Italian pizza with tomato, mozzarella, and basil"
                />
              )}
            />
            <FormElement
              control={form.control}
              fieldName="tags"
              label="Tags (Optional)"
              render={({ field }) => (
                <MultipleSelector
                  {...field}
                  creatable
                  onChange={field.onChange}
                  placeholder="Vegetarian, Italian, Pizza,..."
                  loadingIndicator={<div>Loading...</div>}
                  onSearch={fetchTagsSuggestions}
                  disabled={form.formState.isSubmitting || undefined}
                  triggerSearchOnFocus
                />
              )}
            />
            <FormElement
              control={form.control}
              fieldName="prepTime"
              label="Preparation Time (Optional)"
              render={({ field }) => (
                <div className="flex items-center gap-2">
                  <div className="max-w-[75px]">
                    <Input
                      {...field}
                      type="number"
                      disabled={form.formState.isSubmitting}
                    />
                  </div>
                  <div>minutes</div>
                </div>
              )}
            />
            <FormElement
              control={form.control}
              fieldName="steps"
              label="Steps"
              render={({ field }) => (
                <Textarea
                  className="resize-none"
                  {...field}
                  rows={10}
                  disabled={form.formState.isSubmitting}
                  placeholder="1. Preheat the oven to 250°C..."
                />
              )}
            />
          </div>
          <div className="border-b py-12">
            <div className="flex justify-between">
              <div>
                <h2 className="text-lg font-bold">Ingredients</h2>
              </div>
              <div>
                <Button
                  onClick={() => {
                    ingredientsField.append({});
                  }}
                >
                  <Plus /> Add Another
                </Button>
              </div>
            </div>
            <div className="flex flex-col gap-2">
              <FormElement
                control={form.control}
                fieldName="servings"
                label="Number of Servings (Optional)"
                render={({ field }) => (
                  <div className="max-w-[75px]">
                    <Input
                      {...field}
                      type="number"
                      disabled={form.formState.isSubmitting}
                    />
                  </div>
                )}
              />
              <div className="grid md:grid-cols-12 sm:grid-cols-4 gap-1">
                <div className="md:col-span-2 sm:col-span-1">
                  <Label>Quantity</Label>
                </div>
                <div className="md:col-span-2 sm:col-span-1">
                  <Label>Unit</Label>
                </div>
                <div className="sm:col-span-2 md:col-span-8">
                  <Label>Name</Label>
                </div>
                {ingredientsField.fields.map((field, index) => (
                  <Foo
                    key={field.id}
                    control={form.control}
                    update={form.setValue}
                    remove={ingredientsField.remove}
                    index={index}
                    disabled={form.formState.isSubmitting}
                  />
                ))}
              </div>
            </div>
          </div>
          <div className="fixed bottom-0 left-0 right-0 bg-background py-2">
            <div className="relative max-w-3xl mx-auto h-full border-t py-2 px-4">
              <div className="flex justify-end gap-4">
                <ButtonLink variant="secondary" to="/recipes">
                  Cancel
                </ButtonLink>
                <Button type="submit" disabled={form.formState.isSubmitting}>
                  Publish
                </Button>
              </div>
            </div>
          </div>
        </form>
      </Form>
      <ImportFromURL
        isOpen={!!fromURL}
        close={closeImportFromURL}
        onFinish={onFinish}
      />
      <DevTool control={form.control} />
    </div>
  );
};
