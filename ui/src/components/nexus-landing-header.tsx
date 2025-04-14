import { cn } from "@/lib/utils";
import AuthButton from "./auth-button";
import { LinkProps } from "./landing-links";
import { NavLink } from "./link/nav-link";
import NexusAILogo from "./nexus-logo";

export function NexusAILandingHeader({
  leftLinks,
  rightLinks,
  full = false,
}: {
  leftLinks?: LinkProps[];
  rightLinks?: LinkProps[];
  full?: boolean;
}) {
  return (
    <header className="shadow-sm">
      {/* <nav className=" flex h-14 items-center justify-between lg:px-6"> */}
      <nav
        className={cn(
          "flex h-14 items-center justify-between lg:px-6",
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
