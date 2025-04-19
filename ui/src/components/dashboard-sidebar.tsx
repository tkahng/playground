import { LinkDto } from "@/components/landing-links";
import { cn } from "@/lib/utils";
import { JSX } from "react";
import { Link, useLocation } from "react-router";
import { buttonVariants } from "./ui/button";

export const DashboardSidebar = ({
  links,
  backLink,
}: {
  links: LinkDto[];
  backLink?: JSX.Element;
}) => {
  const { pathname } = useLocation();
  return (
    // <nav className="flex flex-grow flex-col">
    //   <div className="border flex flex-col flex-grow p-8">
    <nav className="flex flex-col border w-64 px-8 py-12 gap-4 justify-start">
      {/* <div className="border flex flex-col flex-grow "> */}
      {backLink}
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
            "justify-start text-md"
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
