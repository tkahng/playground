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
  DASHBOARD_HOME: { to: RouteMap.DASHBOARD_HOME, title: "Dashboard" },
  TASK_PROJECTS: { to: RouteMap.TASK_PROJECTS, title: "Projects" },
  PROTECTED_HOME: { to: RouteMap.PROTECTED_HOME, title: "Protected" },
  SETTINGS_HOME: { to: RouteMap.SETTINGS_HOME, title: "Settings" },
  SETTINGS_ACCOUNT: { to: RouteMap.ACCOUNT_SETTINGS, title: "Account" },
  SETTINGS_BILLING: { to: RouteMap.BILLING_SETTINGS, title: "Billing" },
  ADMIN_DASHBOARD_HOME: {
    to: RouteMap.ADMIN_DASHBOARD_HOME,
    title: "Dashboard",
  },
  ADMIN_DASHBOARD_USERS: { to: RouteMap.ADMIN_DASHBOARD_USERS, title: "Users" },
  ADMIN_DASHBOARD_ROLES: { to: RouteMap.ADMIN_DASHBOARD_ROLES, title: "Roles" },
  ADMIN_DASHBOARD_PERMISSIONS: {
    to: RouteMap.ADMIN_DASHBOARD_PERMISSIONS,
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

export const adminHeaderLinks: LinkDto[] = [RouteLinks.ADMIN_DASHBOARD_HOME];

export const dashboardSidebarLinks: LinkDto[] = [
  RouteLinks.DASHBOARD_HOME,
  RouteLinks.TASK_PROJECTS,
  RouteLinks.PROTECTED_HOME,
  RouteLinks.SETTINGS_HOME,
];

export const authenticatedSubHeaderLinks: LinkDto[] = [
  RouteLinks.DASHBOARD_HOME,
  RouteLinks.TASK_PROJECTS,
  RouteLinks.PROTECTED_HOME,
  RouteLinks.SETTINGS_HOME,
];

export const tasksSidebarLinks: LinkDto[] = [RouteLinks.TASK_PROJECTS];

export const adminSidebarLinks: LinkDto[] = [
  RouteLinks.ADMIN_DASHBOARD_HOME,
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
  RouteLinks.SETTINGS_ACCOUNT,
  RouteLinks.SETTINGS_BILLING,
];
