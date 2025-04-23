import { LinkDto } from "@/components/links";
import { cn } from "@/lib/utils";
import { Link, useLocation } from "react-router";
import BackLink from "./back-link";
import { buttonVariants } from "./ui/button";

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
    // <nav className="flex flex-grow flex-col">
    //   <div className="border flex flex-col flex-grow p-8">
    <nav className="flex flex-col w-64 px-4 md:px-6 lg:px-8 py-8 justify-start border-r">
      {/* <div className="border flex flex-col flex-grow "> */}
      {backLink && <BackLink to={backLink.to} name={backLink.title} />}
      <h1 className="text-2xl font-bold">NexusAI</h1>
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
