import { LinkDto } from "@/components/links";
import { MainNav } from "@/components/main-nav";
import { PlaygroundLandingHeader } from "@/components/playground-landing-header";
import { PlaygroundMinimalFooter } from "@/components/playground-minimal-footer";
import { Outlet } from "@tanstack/react-router";
import { PropsWithChildren } from "react";

export default function DashboardLayout({
  headerLinks,
  children,
}: PropsWithChildren<{
  leftLinks?: LinkDto[];
  rightLinks?: LinkDto[];
  headerLinks?: LinkDto[];
}>) {
  return (
    <div className="min-h-screen flex flex-col">
      <div className="px-4 md:px-6 lg:px-8 py-2 items-center sticky top-0 z-50 w-full bg-background shadow-sm border-b">
        <PlaygroundLandingHeader />
        {headerLinks && headerLinks.length > 0 && (
          <MainNav links={headerLinks} />
        )}
      </div>
      <main className="flex-1">
        {children ?? <Outlet />}
      </main>
      <PlaygroundMinimalFooter />
    </div>
  );
}
