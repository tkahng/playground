import { LinkDto } from "@/components/links";
import PlaygroundLogo from "@/components/playground-logo";
import TeamSwitcher from "@/components/team-switcher";
import { UserNav } from "@/components/user-nav";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { cn } from "@/lib/utils";
import { Link, useLocation } from "react-router";

export function PlaygroundLandingHeader({
  leftLinks,
  rightLinks,
}: {
  leftLinks?: LinkDto[];
  rightLinks?: LinkDto[];
}) {
  const { user } = useAuthProvider();
  const { pathname } = useLocation();
  return (
    <header>
      <nav className={cn("flex h-14 items-center box-border")}>
        <div className="flex flex-grow items-center space-x-4">
          <PlaygroundLogo />
          {leftLinks?.length
            ? leftLinks?.map(({ to, title, current }) => (
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
        {user && <TeamSwitcher />}
        <div className="flex shrink items-center space-x-4">
          {rightLinks?.length && rightLinks.length > 0
            ? rightLinks.map(({ to, title, current }) => (
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
          <UserNav />
        </div>
      </nav>
    </header>
  );
}
