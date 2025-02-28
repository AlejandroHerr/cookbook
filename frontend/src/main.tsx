import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createRouter, RouterProvider } from '@tanstack/react-router';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { Toaster } from 'sonner';

import { ThemeProvider } from './components/theme-provider';
import { routeTree } from './routeTree.gen';

import './index.css';

const queryClient = new QueryClient();
const context = { queryClient, foo: 'bar' };
const router = createRouter({
  routeTree,
  context: context,
  defaultPreload: 'intent',
  defaultPreloadStaleTime: 0,
  defaultErrorComponent: (options) => (
    <div>
      <h1>Oops! Something went wrong</h1>
      <p>{options.error.message}</p>
      <p>{JSON.stringify(options.info ?? {}, null, 2)}</p>
      <button onClick={options.reset}>Retry</button>
    </div>
  ),
  defaultNotFoundComponent: (options) => (
    <div>
      <h1>404 Not Found</h1>
      {!!options.data && <p>{JSON.stringify(options.data, null, 2)}</p>}
    </div>
  ),
});
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
// eslint-disable-next-line @typescript-eslint/no-non-null-assertion
createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ThemeProvider storageKey="vite-ui-theme">
      <Toaster />
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </ThemeProvider>
  </StrictMode>,
);
