import { PlaygroundLandingHeader } from "@/components/playground-landing-header";
import { PlaygroundMinimalFooter } from "@/components/playground-minimal-footer";
import { Outlet } from "@tanstack/react-router";
import { PropsWithChildren } from "react";

export default function PublicLayout({ children }: PropsWithChildren) {
  return (
    <>
      <div className="relative flex min-h-dvh flex-col">
        <div className="px-4 md:px-6 lg:px-8 py-2 items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
          <PlaygroundLandingHeader />
        </div>
        <main className="flex-1">
          {children ?? <Outlet />}
        </main>
        <PlaygroundMinimalFooter />
      </div>
    </>
  );
}
