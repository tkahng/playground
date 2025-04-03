import { useAuthProvider } from "@/hooks/use-auth-provider";
import { cn } from "@/lib/utils";
import { useState } from "react";
import { NavLink, useNavigate } from "react-router";
import { LinkProps } from "./landing-links";
import { ModeToggle } from "./mode-toggle";
import NexusAILogo from "./nexus-logo";
import { RouteMap } from "./route-map";
import { Button } from "./ui/button";

export function NexusAILandingHeader({
  leftLinks,
  rightLinks,
}: {
  leftLinks?: LinkProps[];
  rightLinks?: LinkProps[];
}) {
  const [loading, setLoading] = useState(false);
  const { user, logout } = useAuthProvider();
  const navigate = useNavigate();
  const handleLogout = async (event: React.FormEvent) => {
    event.preventDefault();
    setLoading(true);
    await logout();
    navigate(RouteMap.HOME);
  };
  return (
    <header>
      <nav className="flex h-14 items-center justify-between shadow-sm lg:px-6">
        <div className="flex flex-grow items-center">
          <NexusAILogo />
          {leftLinks?.length &&
            leftLinks.map(({ to: href, title }) => (
              <NavLink
                key={title}
                to={href}
                className={({ isActive }) =>
                  cn(
                    "text-lg font-medium px-4 underline-offset-4 hover:underline active:bg-secondary active:text-secondary-foreground",
                    isActive ? "underline font-bold" : ""
                  )
                }
              >
                {" "}
                {title}{" "}
              </NavLink>
            ))}
        </div>
        <div className="flex shrink items-center space-x-4">
          {rightLinks?.length &&
            rightLinks.map(({ to: href, title }) => (
              <NavLink key={title} title={title} to={href} />
            ))}
          <ModeToggle />
          {user == null ? (
            <>
              <Button asChild>
                <NavLink to={RouteMap.SIGNIN}>
                  <span>Sign in</span>
                </NavLink>
              </Button>
              <Button asChild>
                <NavLink to={RouteMap.SIGNUP}>
                  <span>Sign up</span>
                </NavLink>
              </Button>
            </>
          ) : (
            <form onSubmit={handleLogout}>
              <Button type="submit" disabled={loading} variant="default">
                {/* <NavLink onClick={handleLogout} to={RouteMap.HOME}> */}
                <span>Sign out</span>
                {/* </NavLink> */}
              </Button>
            </form>
          )}
          {/* <AuthButton /> */}
        </div>
      </nav>
    </header>
  );
}
