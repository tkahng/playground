import { cn } from "@/lib/utils";
import { NavLink } from "react-router";
import { LinkProps } from "./landing-links";
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
          {/* <AuthButton /> */}
        </div>
      </nav>
    </header>
  );
}
