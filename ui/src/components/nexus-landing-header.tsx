import AuthButton from "@/components/auth-button";
import { LinkDto } from "@/components/landing-links";
import { NavLink } from "@/components/link/nav-link";
import NexusAILogo from "@/components/nexus-logo";
import { cn } from "@/lib/utils";

export function NexusAILandingHeader({
  leftLinks,
  rightLinks,
  full = false,
}: {
  leftLinks?: LinkDto[];
  rightLinks?: LinkDto[];
  full?: boolean;
}) {
  return (
    <header className="shadow-sm">
      {/* <nav className=" flex h-14 items-center justify-between lg:px-6"> */}
      <nav
        className={cn(
          "flex h-14 items-center justify-between lg:px-8",
          !full ? "mx-auto container" : undefined
        )}
      >
        <div className="flex flex-grow items-center content-center space-x-4">
          <NexusAILogo />
          {leftLinks?.length &&
            leftLinks.map(({ to, title }) => (
              <NavLink key={title} title={title} to={to} />
            ))}
        </div>
        <div className="flex shrink items-center space-x-4">
          {rightLinks?.length &&
            rightLinks.map(({ to, title }) => (
              <NavLink key={title} title={title} to={to} />
            ))}
          <AuthButton />
        </div>
      </nav>
    </header>
  );
}
