import { LinkDto } from "@/components/links";
import { cn } from "@/lib/utils";
import { Link, useLocation } from "react-router";

export function MainNav({
  className,
  links,
  ...props
}: React.HTMLAttributes<HTMLElement> & { links: LinkDto[] }) {
  const { pathname } = useLocation();

  return (
    <nav className={cn("flex items-center h-12", className)} {...props}>
      {links.map((link) => (
        <Link
          key={link.to}
          to={link.to}
          className={cn(
            link.current
              ? link.current(pathname)
                ? "underline"
                : "text-muted-foreground"
              : link.to === pathname
              ? "underline"
              : "text-muted-foreground",
            "text-sm font-normal hover:text-primary transition-colors hover:bg-muted rounded-md p-2"
          )}
        >
          {link.title}
        </Link>
      ))}
    </nav>
  );
}
