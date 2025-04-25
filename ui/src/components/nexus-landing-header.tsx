import AuthButton from "@/components/auth-button";
import { NavbarLink } from "@/components/link/nav-link";
import { LinkDto } from "@/components/links";
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
      <nav className={cn("flex h-14 items-center box-border")}>
        <div className="flex flex-grow items-center space-x-4">
          <NexusAILogo />
          {leftLinks?.length &&
            leftLinks.length > 0 &&
            leftLinks.map(({ to, title }) => (
              <NavbarLink key={title} title={title} to={to} />
            ))}
        </div>
        <div className="flex shrink items-center space-x-4">
          {rightLinks?.length &&
            rightLinks.length > 0 &&
            rightLinks.map(({ to, title }) => (
              <NavbarLink key={title} title={title} to={to} />
            ))}
          <AuthButton />
        </div>
      </nav>
    </header>
  );
}
