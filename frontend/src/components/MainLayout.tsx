import React from 'react';

export const MainLayout: React.FC<React.PropsWithChildren> = ({ children }) => (
  <div className="p-4 max-w-7xl mx-auto h-full flex overflow-auto">
    {children}
  </div>
);
