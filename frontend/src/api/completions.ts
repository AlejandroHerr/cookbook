import { DefaultError, MutationOptions } from '@tanstack/react-query';
import axios from 'axios';

import { CompleteRecipeResponse } from '@/types/completions';

export const CompleteRecipeMutationOption: MutationOptions<
  CompleteRecipeResponse,
  DefaultError,
  { url: string }
> = {
  mutationFn: async (data: { url: string }) =>
    axios
      .post<CompleteRecipeResponse>(
        'http://localhost:8080/completions/recipe',
        data,
      )
      .then((res) => res.data),
};
