import { PlaygroundLandingHeader } from "@/components/playground-landing-header";
import { PlaygroundMinimalFooter } from "@/components/playground-minimal-footer";
import { Outlet } from "react-router";

export default function PublicLayout() {
  return (
    <>
      <div className="relative flex min-h-dvh flex-col">
        <div className="px-4 md:px-6 lg:px-8 py-2 items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
          <PlaygroundLandingHeader />
        </div>
        <main className="flex-1">
          <Outlet />
        </main>
        <PlaygroundMinimalFooter />
      </div>
    </>
  );
}
