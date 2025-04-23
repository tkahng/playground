import { RouteMap } from "@/components/route-map";
import { JSX } from "react";
export type LinkDto = {
  title?: string;
  icon?: JSX.Element;
  to: string;
};

export const RouteLinks = {
  FEATURES: { to: RouteMap.FEATURES, title: "Features" },
  PRICING: { to: RouteMap.PRICING, title: "Pricing" },
  ABOUT: { to: RouteMap.ABOUT, title: "About" },
  CONTACT: { to: RouteMap.CONTACT, title: "Contact" },
  DASHBOARD: { to: RouteMap.DASHBOARD, title: "Dashboard" },
  DASHBOARD_OVERVIEW: { to: RouteMap.DASHBOARD, title: "Overview" },
  TASK_PROJECTS: { to: RouteMap.TASK_PROJECTS, title: "Projects" },
  PROTECTED: { to: RouteMap.PROTECTED, title: "Protected" },
  SETTINGS: { to: RouteMap.SETTINGS, title: "Settings" },
  ACCOUNT_SETTINGS: { to: RouteMap.ACCOUNT_SETTINGS, title: "Account" },
  BILLING_SETTINGS: { to: RouteMap.BILLING_SETTINGS, title: "Billing" },
  ADMIN: {
    to: RouteMap.ADMIN,
    title: "Dashboard",
  },
  ADMIN_DASHBOARD_USERS: { to: RouteMap.ADMIN_USERS, title: "Users" },
  ADMIN_DASHBOARD_ROLES: { to: RouteMap.ADMIN_ROLES, title: "Roles" },
  ADMIN_DASHBOARD_PERMISSIONS: {
    to: RouteMap.ADMIN_PERMISSIONS,
    title: "Permissions",
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

export const dashboardSidebarLinks: LinkDto[] = [
  RouteLinks.DASHBOARD,
  RouteLinks.TASK_PROJECTS,
  RouteLinks.PROTECTED,
  RouteLinks.SETTINGS,
];

export const authenticatedSubHeaderLinks: LinkDto[] = [
  RouteLinks.DASHBOARD_OVERVIEW,
  RouteLinks.TASK_PROJECTS,
  RouteLinks.PROTECTED,
  RouteLinks.SETTINGS,
];

export const tasksSidebarLinks: LinkDto[] = [RouteLinks.TASK_PROJECTS];

export const adminSidebarLinks: LinkDto[] = [
  RouteLinks.ADMIN,
  RouteLinks.ADMIN_DASHBOARD_USERS,
  RouteLinks.ADMIN_DASHBOARD_ROLES,
  RouteLinks.ADMIN_DASHBOARD_PERMISSIONS,
];

export const protectedSidebarLinks: LinkDto[] = [
  RouteLinks.PROTECTED_BASIC,
  RouteLinks.PROTECTED_PRO,
  RouteLinks.PROTECTED_ADVANCED,
];

export const settingsSidebarLinks: LinkDto[] = [
  RouteLinks.ACCOUNT_SETTINGS,
  RouteLinks.BILLING_SETTINGS,
];
