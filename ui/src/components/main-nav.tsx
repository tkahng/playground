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
            // buttonVariants({ variant: "ghost" }),
            pathname === link.to ? "underline" : "text-muted-foreground",
            "text-sm font-normal hover:text-primary transition-colors hover:bg-muted rounded-md p-2"
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
