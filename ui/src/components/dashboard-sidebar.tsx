import { cn } from "@/lib/utils";
import { Link, useLocation } from "react-router";
import { LinkProps } from "./landing-links";

export const DashboardSidebar = ({ links }: { links: LinkProps[] }) => {
  const { pathname } = useLocation();
  return (
    <nav className="flex flex-grow flex-col">
      <div className="border flex flex-col flex-grow p-8">
        {links.map((item) => (
          <Link
            key={item.title}
            to={item.to}
            // className="flex items-center gap-2 rounded-md p-2 hover:bg-muted"
            className={cn(
              "flex items-center gap-2 rounded-md p-2 hover:bg-muted",
              pathname === item.to ? "shadow-xl" : "text-foreground"
            )}
          >
            {item.icon}
            <span>{item.title}</span>
          </Link>
        ))}
      </div>
    </nav>
  );
};
