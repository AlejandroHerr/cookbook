import { zodResolver } from '@hookform/resolvers/zod';
import { useFieldArray, useForm } from 'react-hook-form';
import { z } from 'zod';

import { Unit } from '@/types/recipes/units';

const formSchema = z.object({
  title: z
    .string({
      required_error: 'Title is required',
    })
    .min(3, 'Title must be at least 3 characters long')
    .describe('Title'),
  headline: z
    .string()
    .min(3, 'Headline must be at least 3 characters long')
    .optional(),
  description: z
    .string()
    .min(3, 'Description must be at least 3 characters long')
    .optional(),
  steps: z
    .string()
    .min(3, 'Steps must be at least 3 characters long')
    .optional(),
  prepTime: z.number({ coerce: true }).optional(),
  servings: z.number({ coerce: true }).min(1, 'Servings must be at least 1'),
  url: z.string().url('Invalid URL').optional(),
  tags: z
    .array(
      z.object({
        label: z
          .string({
            required_error: 'Tag is required',
          })
          .min(3, 'Tag must be at least 3 characters long'),
        value: z.string(),
      }),
    )
    .default([]),
  ingredients: z.array(
    z
      .object({
        name: z.string().optional(),
        quantity: z.number().optional(),
        unit: z.nativeEnum(Unit).optional(),
      })
      .optional(),
  ),
});

export type RecipeFormSchema = z.infer<typeof formSchema>;

export const useRecipeForm = () => {
  const form = useForm<RecipeFormSchema>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      servings: 1,
      ingredients: [{}, {}, {}],
    },
  });

  const ingredientsField = useFieldArray({
    control: form.control,
    name: 'ingredients',
  });

  return { form, ingredientsField };
};
