import { RouteMap } from "@/components/route-map";
import { JSX } from "react";
export type LinkDto = {
  title?: string;
  icon?: JSX.Element;
  to: string;
  current?: (pathname: string) => boolean;
};

export const defaultCurrentFunc = (pathname: string, to: string) => {
  return pathname === to;
};

export const RouteLinks = {
  FEATURES: { to: RouteMap.FEATURES, title: "Features" },
  PRICING: { to: RouteMap.PRICING, title: "Pricing" },
  ABOUT: { to: RouteMap.ABOUT, title: "About" },
  CONTACT: { to: RouteMap.CONTACT, title: "Contact" },
  DASHBOARD: {
    to: RouteMap.DASHBOARD,
    title: "Dashboard",
    current: (pathname: string) => pathname.startsWith(RouteMap.DASHBOARD),
  },
  DASHBOARD_OVERVIEW: { to: RouteMap.DASHBOARD, title: "Overview" },
  TASK_PROJECTS: { to: RouteMap.TASK_PROJECTS, title: "Projects" },
  SETTINGS: {
    to: RouteMap.SETTINGS,
    title: "Settings",
    current: (pathname: string) => pathname.startsWith(RouteMap.SETTINGS),
  },
  GENERAL_SETTINGS: { to: RouteMap.SETTINGS, title: "General" },
  BILLING_SETTINGS: { to: RouteMap.BILLING_SETTINGS, title: "Billing" },
  ADMIN: {
    to: RouteMap.ADMIN,
    title: "Overview",
  },
  ADMIN_DASHBOARD_USERS: {
    to: RouteMap.ADMIN_USERS,
    title: "Users",
    current: (pathname: string) => pathname.startsWith(RouteMap.ADMIN_USERS),
  },
  ADMIN_DASHBOARD_ROLES: {
    to: RouteMap.ADMIN_ROLES,
    title: "Roles",
    current: (pathname: string) => pathname.startsWith(RouteMap.ADMIN_ROLES),
  },
  ADMIN_DASHBOARD_PERMISSIONS: {
    to: RouteMap.ADMIN_PERMISSIONS,
    title: "Permissions",
    current: (pathname: string) =>
      pathname.startsWith(RouteMap.ADMIN_PERMISSIONS),
  },
  PROTECTED: {
    to: RouteMap.PROTECTED,
    title: "Protected",
    current: (pathname: string) =>
      [
        RouteMap.PROTECTED,
        RouteMap.PROTECTED_BASIC,
        RouteMap.PROTECTED_PRO,
        RouteMap.PROTECTED_ADVANCED,
      ].includes(pathname),
  },
  PROTECTED_BASIC: { to: RouteMap.PROTECTED_BASIC, title: "Basic" },
  PROTECTED_PRO: { to: RouteMap.PROTECTED_PRO, title: "Pro" },
  PROTECTED_ADVANCED: { to: RouteMap.PROTECTED_ADVANCED, title: "Advanced" },
} as const;

export const landingLinks: LinkDto[] = [
  RouteLinks.FEATURES,
  RouteLinks.PRICING,
  RouteLinks.ABOUT,
  RouteLinks.CONTACT,
];

export const adminHeaderLinks: LinkDto[] = [
  RouteLinks.ADMIN,
  RouteLinks.ADMIN_DASHBOARD_USERS,
  RouteLinks.ADMIN_DASHBOARD_ROLES,
  RouteLinks.ADMIN_DASHBOARD_PERMISSIONS,
];

export const protectedSidebarLinks: LinkDto[] = [
  RouteLinks.PROTECTED,
  RouteLinks.PROTECTED_BASIC,
  RouteLinks.PROTECTED_PRO,
  RouteLinks.PROTECTED_ADVANCED,
];

export const authenticatedSubHeaderLinks: LinkDto[] = [
  RouteLinks.DASHBOARD_OVERVIEW,
  RouteLinks.TASK_PROJECTS,
  RouteLinks.PROTECTED,
  RouteLinks.SETTINGS,
];

export const settingsSidebarLinks: LinkDto[] = [
  RouteLinks.GENERAL_SETTINGS,
  RouteLinks.BILLING_SETTINGS,
];
