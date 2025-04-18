import { LinkDto } from "@/components/landing-links";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Link, useLocation } from "react-router";

export const DashboardSidebar = ({ links }: { links: LinkDto[] }) => {
  const { pathname } = useLocation();
  return (
    // <nav className="flex flex-grow flex-col">
    //   <div className="border flex flex-col flex-grow p-8">
    <nav className="flex flex-col border p-12 items-center">
      {/* <div className="border flex flex-col flex-grow "> */}
      {links.map((item) => (
        <Link
          key={item.title}
          to={item.to}
          // className="flex items-center gap-2 rounded-md p-2 hover:bg-muted"
          className={cn(
            buttonVariants({ variant: "ghost" }),
            pathname === item.to
              ? "bg-muted hover:bg-muted underline"
              : "hover:bg-transparent hover:underline",
            "justify-start text-lg"
          )}
        >
          {item.icon}
          <span>{item.title}</span>
        </Link>
      ))}
      {/* </div> */}
    </nav>
  );
};
