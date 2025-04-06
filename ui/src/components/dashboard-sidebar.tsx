import { cn } from "@/lib/utils";
import { Home, Key, User } from "lucide-react";
import { Link, useLocation } from "react-router";
import { LinkProps } from "./landing-links";
const items: LinkProps[] = [
  {
    title: "Home",
    to: "/dashboard",
    icon: <Home />,
  },
  {
    title: "Users",
    to: "/dashboard/users",
    icon: <User />,
  },
  {
    title: "Roles",
    to: "/dashboard/roles",
    icon: <Key />,
  },
];

export const DashboardSidebar = () => {
  const { pathname } = useLocation();
  return (
    <nav className="flex flex-grow flex-col">
      <div className="border flex flex-col flex-grow p-8">
        {items.map((item) => (
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
