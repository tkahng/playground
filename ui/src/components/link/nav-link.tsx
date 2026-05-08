import { LinkDto } from "@/components/links";
import { cn } from "@/lib/utils";
import { Link } from "@tanstack/react-router";
import { JSX, PropsWithChildren } from "react";

type NavbarLinkProps = { className?: string } & LinkDto;

export function NavbarLink({
  title,
  to,
  icon,
  className,
}: PropsWithChildren<NavbarLinkProps>): JSX.Element {
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
