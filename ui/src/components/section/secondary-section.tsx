import { cn } from "@/lib/utils";
import type { PropsWithChildren } from "react";

export interface Props {
  cols?: 2 | 3;
}

export default function SecondarySection({
  children,
  cols = 3,
}: PropsWithChildren<Props>) {
  return (
    <section className="w-full py-12 md:py-24 lg:py-32 bg-secondary text-secondary flex flex-col items-center">
      <div className="container px-4 md:px-6">
        <div
          className={cn(
            "grid gap-10",
            cols === 3
              ? "sm:grid-cols-2 lg:grid-cols-3"
              : "px-10 md:gap-16 lg:grid-cols-2"
          )}
        >
          {children}
        </div>
      </div>
    </section>
  );
}
