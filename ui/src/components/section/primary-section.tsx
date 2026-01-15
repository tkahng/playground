import type { PropsWithChildren } from "react";
import React from "react";

export interface LandingTopSectionProps {
  children?: React.ReactNode;
}

const PrimarySection: React.FC<PropsWithChildren> = ({ children }) => {
  return (
    <section className="w-full py-12 md:py-24 lg:py-32 flex flex-col items-center">
      <div className="container px-4 md:px-6">
        <div className="flex flex-col items-center justify-center space-y-4 text-center">
          {children}
        </div>
      </div>
    </section>
  );
};
export default PrimarySection;
