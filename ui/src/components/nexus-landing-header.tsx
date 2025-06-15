import AuthButton from "@/components/auth-button";
import { LinkDto } from "@/components/links";
import NexusAILogo from "@/components/nexus-logo";
import { cn } from "@/lib/utils";
import { Link, useLocation } from "react-router";
import TeamSwitcher from "./team-switcher-2";

export function NexusAILandingHeader({
  leftLinks,
  rightLinks,
}: {
  leftLinks?: LinkDto[];
  rightLinks?: LinkDto[];
}) {
  const { pathname } = useLocation();
  return (
    <header>
      <nav className={cn("flex h-14 items-center box-border")}>
        <div className="flex flex-grow items-center space-x-4">
          <NexusAILogo />
          <TeamSwitcher />
          {leftLinks?.length
            ? leftLinks.map(({ to, title, current }) => (
                <Link
                  key={to}
                  className={cn(
                    current
                      ? current(pathname)
                        ? "underline"
                        : "text-muted-foreground"
                      : to === pathname
                      ? "underline"
                      : "text-muted-foreground",
                    "text-sm font-medium underline-offset-4 hover:underline"
                  )}
                  to={to}
                >
                  {title}
                </Link>
              ))
            : null}
        </div>
        <div className="flex shrink items-center space-x-4">
          {rightLinks?.length && rightLinks.length > 0
            ? rightLinks.map(({ to, title, current }) => (
                <Link
                  className={cn(
                    current
                      ? current(pathname)
                        ? "underline"
                        : "text-muted-foreground"
                      : to === pathname
                      ? "underline"
                      : "text-muted-foreground",
                    "text-sm font-medium underline-offset-4 hover:underline"
                  )}
                  to={to}
                >
                  {title}
                </Link>
              ))
            : null}
          <AuthButton />
        </div>
      </nav>
    </header>
  );
}
