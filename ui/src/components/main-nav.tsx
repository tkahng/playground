import { LinkDto } from "@/components/landing-links";
import { cn } from "@/lib/utils";
import { Link, useLocation } from "react-router";

export function MainNav({
  className,
  links,
  ...props
}: React.HTMLAttributes<HTMLElement> & { links: LinkDto[] }) {
  const { pathname } = useLocation();

  return (
    <nav className={cn("flex items-center px-2 md:px-4", className)} {...props}>
      {links.map((link) => (
        <Link
          key={link.to}
          to={link.to}
          className={cn(
            // buttonVariants({ variant: "ghost" }),
            pathname === link.to ? "underline" : "text-muted-foreground",
            "text-sm font-normal hover:text-primary transition-colors py-4 px-3 hover:bg-muted rounded-md"
          )}
          // className={cn(
          //   // buttonVariants({ variant: "ghost" }),
          //   pathname === link.to && "text-primary underline",
          //   "text-muted-foreground text-sm font-medium transition-colors hover:text-primary"
          // )}
        >
          {link.title}
        </Link>
      ))}
    </nav>
  );
}
