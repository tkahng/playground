import { cn } from "@/lib/utils";
import { JSX, PropsWithChildren } from "react";
import { Link } from "react-router";
import { LinkProps } from "../landing-links";

type NavLinkProps = { className?: string } & LinkProps;

export function NavLink({
  title,
  to,
  icon,
  className,
}: PropsWithChildren<NavLinkProps>): JSX.Element {
  return (
    <Link
      className={cn(
        "text-sm font-medium underline-offset-4 hover:underline active:bg-secondary active:text-secondary-foreground",
        className
      )}
      to={to}
    >
      {icon && icon}
      {title}
    </Link>
  );
}
