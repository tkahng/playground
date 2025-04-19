import { RouteMap } from "@/components/route-map";
import { Home, Key, User } from "lucide-react";
import { JSX } from "react";
export type LinkDto = {
  title?: string;
  icon?: JSX.Element;
  to: string;
};
export const landingLinks: LinkDto[] = [
  { to: RouteMap.FEATURES, title: "Features" },
  { to: RouteMap.PRICING, title: "Pricing" },
  { to: RouteMap.ABOUT, title: "About" },
  { to: RouteMap.CONTACT, title: "Contact" },
];

export const adminHeaderLinks: LinkDto[] = [
  { to: RouteMap.DASHBOARD_HOME, title: "Dashboard" },
];

export const dashboardSidebarLinks: LinkDto[] = [
  {
    title: "Home",
    to: RouteMap.DASHBOARD_HOME,
  },
  {
    title: "Projects",
    to: RouteMap.TASKS_HOME,
  },
  {
    title: "Kanban",
    to: RouteMap.DASHBOARD_KANBAN,
  },
  {
    title: "Protected",
    to: RouteMap.PROTECTED_HOME,
  },
];

export const tasksSidebarLinks: LinkDto[] = [
  {
    title: "Home",
    to: RouteMap.TASKS_HOME,
  },
  {
    title: "Projects",
    to: RouteMap.TASK_PROJECTS,
  },
];

export const adminSidebarLinks: LinkDto[] = [
  {
    title: "Home",
    to: RouteMap.ADMIN_DASHBOARD_HOME,
    icon: <Home />,
  },
  {
    title: "Users",
    to: RouteMap.ADMIN_DASHBOARD_USERS,
    icon: <User />,
  },
  {
    title: "Roles",
    to: RouteMap.ADMIN_DASHBOARD_ROLES,
    icon: <Key />,
  },
  {
    title: "Permissions",
    to: RouteMap.ADMIN_DASHBOARD_PERMISSIONS,
    icon: <Key />,
  },
];

export const protectedSidebarLinks: LinkDto[] = [
  {
    title: "Basic",
    to: RouteMap.PROTECTED_BASIC,
  },
  {
    title: "Pro ",
    to: RouteMap.PROTECTED_PRO,
  },
  {
    title: "Advanced ",
    to: RouteMap.PROTECTED_ADVANCED,
  },
];
