import { cn } from "@/lib/utils";
import { Link, useLocation } from "react-router";
import { RouteMap } from "./route-map";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "./ui/accordion";
import { buttonVariants } from "./ui/button";

const protectedRoutes = [
  {
    title: "Basic",
    to: RouteMap.PROTECTED_BASIC,
  },
  {
    title: "Pro",
    to: RouteMap.PROTECTED_PRO,
  },
  {
    title: "Advanced",
    to: RouteMap.PROTECTED_ADVANCED,
  },
];

const tasksRoutes = [
  {
    title: "Projects",
    to: RouteMap.TASK_PROJECTS,
  },
  {
    title: "Kanban",
    to: RouteMap.DASHBOARD_KANBAN,
  },
];

export const AccordionSidebar = () => {
  const { pathname } = useLocation();
  return (
    // <nav className="flex flex-grow flex-col">
    //   <div className="border flex flex-col flex-grow p-8">
    <nav className="flex flex-col border p-12 justify-start">
      <Accordion type="single" collapsible className="w-full">
        <AccordionItem value="item-1">
          <AccordionTrigger>Protected Routes</AccordionTrigger>
          <AccordionContent className="flex flex-col gap-2">
            {protectedRoutes.map((item) => (
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
                {/* {item.icon} */}
                <span>{item.title}</span>
              </Link>
            ))}
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-2">
          <AccordionTrigger>Tasks</AccordionTrigger>
          <AccordionContent className="flex flex-col gap-2">
            {tasksRoutes.map((item) => (
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
                {/* {item.icon} */}
                <span>{item.title}</span>
              </Link>
            ))}
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </nav>
  );
};
