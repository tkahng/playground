import { LinkDto } from "@/components/landing-links";
import { cn } from "@/lib/utils";
import { Link } from "react-router";

export function MainNav({
  className,
  links,
  ...props
}: React.HTMLAttributes<HTMLElement> & { links: LinkDto[] }) {
  return (
    <nav
      className={cn("flex items-center space-x-4 lg:space-x-6", className)}
      {...props}
    >
      {links.map((link) => (
        <Link
          to={link.to}
          className="text-sm font-medium transition-colors hover:text-primary"
        >
          {link.title}
        </Link>
      ))}
    </nav>
  );
}
