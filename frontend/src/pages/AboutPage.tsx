import React from 'react';

export const AboutPage: React.FC = () => {
  return (
    <div className="flex flex-col items-center justify-center h-full bg-background text-foreground">
      <h1 className="text-5xl font-bold">About catalyst</h1>
      <p className="mt-4 text-lg text-center max-w-2xl">
        catalyst is a platform architecture intended to provide a foundation upon which applications that ingest csv or other data can be built upon.
      </p>
    </div>
  );
};
