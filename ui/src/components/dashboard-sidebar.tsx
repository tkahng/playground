import { LinkDto } from "@/components/links";
import { cn } from "@/lib/utils";
import { Link, useLocation } from "react-router";
import BackLink from "./back-link";

export const DashboardSidebar = ({
  links,
  backLink,
}: {
  links: LinkDto[];
  backLink?: LinkDto;
}) => {
  const { pathname } = useLocation();
  if (links.length === 0) {
    return null;
  }
  return (
    <nav className="flex flex-col w-64 py-8 space-y-2 justify-start border-r grow-0">
      {backLink && <BackLink to={backLink.to} name={backLink.title} />}
      {links.map((item) => (
        <Link
          key={item.to}
          to={item.to}
          className={cn(
            // buttonVariants({ variant: "ghost" }),
            pathname === item.to ? "underline" : "text-muted-foreground",
            "text-sm font-normal hover:text-primary transition-colors hover:bg-muted rounded-md p-2"
          )}
          // className={cn(
          //   // buttonVariants({ variant: "ghost" }),
          //   pathname === link.to && "text-primary underline",
          //   "text-muted-foreground text-sm font-medium transition-colors hover:text-primary"
          // )}
        >
          <span>{item.title}</span>
        </Link>
      ))}
    </nav>
  );
};
