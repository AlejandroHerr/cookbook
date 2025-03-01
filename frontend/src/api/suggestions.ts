import axios from 'axios';

import { GetSuggestionsReponse } from '@/types/suggestions';

export const tagsCompletionsQueryOptions = ({
  search,
}: {
  search: string;
}) => ({
  queryKey: ['tags', search],
  queryFn: async () => {
    const res = await axios.get<GetSuggestionsReponse>(
      `http://localhost:8080/suggestions/tags?search=${search}`,
    );

    return res.data.options;
  },
});
