import type { JSX, PropsWithChildren } from "react";
import { Link, type LinkProps } from "react-router";
import { cn } from "~/lib/utils";


type NavLinkProps = { className?: string } & LinkProps;

export function NavLink({
  title,
  to: href,
  // href,
  // icon,
  
  className,
}: PropsWithChildren<NavLinkProps>): JSX.Element {
  return (
    <Link
      className={cn(
        "text-sm font-medium underline-offset-4 hover:underline active:bg-secondary active:text-secondary-foreground",
        className,
      )}
      to={href}
    >
      {/* {icon && icon} */}
      {title}
    </Link>
  );
}
