import { QueryClient } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import {
  Link,
  Outlet,
  createLink,
  createRootRouteWithContext,
  useLocation,
} from '@tanstack/react-router';
import { TanStackRouterDevtools } from '@tanstack/router-devtools';
import * as React from 'react';

import { MainLayout } from '@/components/MainLayout';
import { useTheme } from '@/components/theme-provider';
import { Switch } from '@/components/ui/switch';
import { cn } from '@/lib/utils';

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()(
  {
    component: RootComponent,
  },
);

const LinkedDiv = createLink('div');

const Header: React.FC = () => {
  const context = useTheme();
  const location = useLocation();

  return (
    <header className="flex justify-between items-center px-4">
      <LinkedDiv className="py-2" to="/recipes">
        <h1 className="title-shadow scroll-m-20 text-4xl font-extrabold tracking-tight lg:text-5xl cursor-pointer">
          the Cookbook
        </h1>
      </LinkedDiv>
      <div className="flex gap-4 items-center">
        <div
          className={cn('flex p-2', {
            'border-b-2 border-primary':
              location.pathname.startsWith('/recipes'),
          })}
        >
          <Link to="/recipes" className="text-sm font-medium">
            Recipes
          </Link>
        </div>
        <div>
          <Switch
            checked={context.theme === 'dark'}
            onClick={() => {
              context.setTheme(context.theme === 'dark' ? 'light' : 'dark');
            }}
          />
        </div>
      </div>
    </header>
  );
};

function RootComponent() {
  return (
    <React.Fragment>
      <Header />
      <MainLayout>
        <Outlet />
      </MainLayout>
      {import.meta.env.DEV && (
        <ReactQueryDevtools
          initialIsOpen={false}
          buttonPosition="bottom-left"
        />
      )}
      {import.meta.env.DEV && (
        <TanStackRouterDevtools initialIsOpen={false} position="bottom-right" />
      )}
    </React.Fragment>
  );
}
