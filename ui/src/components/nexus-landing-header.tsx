import AuthButton from "./auth-button";
import { LinkProps } from "./landing-links";
import { NavLink } from "./link/nav-link";
import { ModeToggle } from "./mode-toggle";
import NexusAILogo from "./nexus-logo";

export function NexusAILandingHeader({
  leftLinks,
  rightLinks,
}: {
  leftLinks?: LinkProps[];
  rightLinks?: LinkProps[];
}) {
  return (
    <header className="shadow-sm">
      <nav className="container mx-auto flex h-14 items-center justify-between lg:px-6">
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
          <ModeToggle />
          <AuthButton />
        </div>
      </nav>
    </header>
  );
}
