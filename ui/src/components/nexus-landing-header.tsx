import AuthButton from "@/components/auth-button";
import { LinkDto } from "@/components/landing-links";
import { NavLink } from "@/components/link/nav-link";
import NexusAILogo from "@/components/nexus-logo";
import { cn } from "@/lib/utils";

export function NexusAILandingHeader({
  leftLinks,
  rightLinks,
}: {
  leftLinks?: LinkDto[];
  rightLinks?: LinkDto[];
}) {
  return (
    <header>
      <nav className={cn("flex h-14 items-center")}>
        <div className="flex flex-grow items-center space-x-4">
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
