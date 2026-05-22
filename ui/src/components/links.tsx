import { RouteMap } from "@/components/route-map";
import { Home, LayoutDashboard, Settings, ShieldUser } from "lucide-react";
import { JSX } from "react";
export type LinkDto = {
  title?: string;
  icon?: JSX.Element;
  to: string;
  current?: (pathname: string) => boolean;
  badge?: string;
};

export const RouteLinks = {
  HOME: {
    to: RouteMap.HOME,
    title: "Home",
    current: (pathname: string) => pathname === RouteMap.HOME,
    icon: <Home />,
  },
  TEAM_LIST: {
    to: RouteMap.TEAM_LIST,
    title: "Teams",
    current: (pathname: string) => pathname.startsWith(RouteMap.TEAM_LIST),
  },
  FEATURES: { to: RouteMap.FEATURES, title: "Features" },
  PRICING: { to: RouteMap.PRICING, title: "Pricing" },
  ABOUT: { to: RouteMap.ABOUT, title: "About" },
  SAY_HELLO: { to: RouteMap.SAY_HELLO, title: "Say Hello", badge: "Try it!" },
  ACCOUNT_DASHBOARD: {
    to: RouteMap.ACCOUNT_DASHBOARD,
    title: "Dashboard",
    current: (pathname: string) => pathname === RouteMap.ACCOUNT_DASHBOARD,
    icon: <LayoutDashboard />,
  },
  ACCOUNT_ROCK_PAPER_SCISSORS: {
    to: RouteMap.ACCOUNT_ROCK_PAPER_SCISSORS,
    title: "Rock Paper Scissors",
    current: (pathname: string) =>
      pathname === RouteMap.ACCOUNT_ROCK_PAPER_SCISSORS,
  },
  ACCOUNT_FRIENDS: {
    to: RouteMap.ACCOUNT_FRIENDS,
    title: "Friends",
    current: (pathname: string) => pathname === RouteMap.ACCOUNT_FRIENDS,
  },
  ROCK_PAPER_SCISSORS: {
    to: RouteMap.ROCK_PAPER_SCISSORS,
    title: "Rock Paper Scissors",
    current: (pathname: string) =>
      pathname === RouteMap.ACCOUNT_ROCK_PAPER_SCISSORS,
  },
  ACCOUNT_TEAMS: {
    to: RouteMap.ACCOUNT_OVERVIEW_TEAMS,
    title: "Teams",
    current: (pathname: string) => pathname === RouteMap.ACCOUNT_OVERVIEW_TEAMS,
  },
  ACCOUNT_OVERVIEW_TEAMS: {
    to: RouteMap.ACCOUNT_OVERVIEW_TEAMS,
    title: "Teams",
    current: (pathname: string) =>
      pathname.startsWith(RouteMap.ACCOUNT_OVERVIEW_TEAMS),
  },
  ACCOUNT_OVERVIEW_TEAM_INVITATIONS: {
    to: RouteMap.ACCOUNT_OVERVIEW_TEAMS_INVITATION,
    title: "Invitations",
    current: (pathname: string) =>
      pathname.startsWith(RouteMap.ACCOUNT_OVERVIEW_TEAMS_INVITATION),
  },
  DASHBOARD_OVERVIEW: {
    to: RouteMap.ACCOUNT_OVERVIEW_TEAMS,
    title: "Overview",
  },
  TASK_PROJECTS: {
    to: RouteMap.TASK_PROJECTS,
    title: "Projects",
    current: (pathname: string) => pathname.startsWith(RouteMap.TASK_PROJECTS),
  },
  SETTINGS: {
    to: RouteMap.SETTINGS,
    title: "Settings",
    icon: <Settings />,
    current: (pathname: string) => pathname.startsWith(RouteMap.SETTINGS),
  },
  GENERAL_SETTINGS: { to: RouteMap.SETTINGS, title: "General" },
  BILLING_SETTINGS: { to: RouteMap.BILLING_SETTINGS, title: "Billing" },
  POINTS_SETTINGS: { to: RouteMap.POINTS_SETTINGS, title: "Points" },
  ADMIN: {
    to: RouteMap.ADMIN,
    title: "Admin",
    current: (pathname: string) => pathname.startsWith(RouteMap.ADMIN),
    icon: <ShieldUser />,
  },
  ADMIN_OVERVIEW: {
    to: RouteMap.ADMIN,
    title: "Overview",
  },
  ADMIN_DASHBOARD_USERS: {
    to: RouteMap.ADMIN_USERS,
    title: "Users",
    current: (pathname: string) => pathname.startsWith(RouteMap.ADMIN_USERS),
  },
  ADMIN_DASHBOARD_JOBS: {
    to: RouteMap.ADMIN_JOBS,
    title: "Jobs",
    current: (pathname: string) => pathname.startsWith(RouteMap.ADMIN_JOBS),
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
  ADMIN_DASHBOARD_PRODUCTS: {
    to: RouteMap.ADMIN_PRODUCTS,
    title: "Products",
    current: (pathname: string) => pathname.startsWith(RouteMap.ADMIN_PRODUCTS),
  },
  ADMIN_DASHBOARD_SUBSCRIPTIONS: {
    to: RouteMap.ADMIN_SUBSCRIPTIONS,
    title: "Subscriptions",
    current: (pathname: string) =>
      pathname.startsWith(RouteMap.ADMIN_SUBSCRIPTIONS),
  },
  ADMIN_DASHBOARD_PLAN_FEATURES: {
    to: RouteMap.ADMIN_PLAN_FEATURES,
    title: "Plan Features",
    current: (pathname: string) =>
      pathname.startsWith(RouteMap.ADMIN_PLAN_FEATURES),
  },
  ADMIN_DASHBOARD_AI_USAGE: {
    to: RouteMap.ADMIN_AI_USAGE,
    title: "AI Usage",
    current: (pathname: string) =>
      pathname.startsWith(RouteMap.ADMIN_AI_USAGE),
  },
  ADMIN_DASHBOARD_BLOG: {
    to: RouteMap.ADMIN_BLOG,
    title: "Blog",
    current: (pathname: string) => pathname.startsWith(RouteMap.ADMIN_BLOG),
  },
  ADMIN_DASHBOARD_MEDIA: {
    to: RouteMap.ADMIN_MEDIA,
    title: "Media",
    current: (pathname: string) => pathname.startsWith(RouteMap.ADMIN_MEDIA),
  },
  BLOG: {
    to: RouteMap.BLOG,
    title: "Blog",
    current: (pathname: string) => pathname.startsWith(RouteMap.BLOG),
  },
  PROTECTED: {
    to: RouteMap.PROTECTED,
    title: "Protected",
    current: (pathname: string) => pathname.startsWith(RouteMap.PROTECTED),
  },
  PROTECTED_BASIC: { to: RouteMap.PROTECTED_BASIC, title: "Basic" },
  PROTECTED_PRO: { to: RouteMap.PROTECTED_PRO, title: "Pro" },
  PROTECTED_ADVANCED: { to: RouteMap.PROTECTED_ADVANCED, title: "Advanced" },
} as const;

export const landingLinks: LinkDto[] = [
  RouteLinks.SAY_HELLO,
  RouteLinks.ROCK_PAPER_SCISSORS,
  RouteLinks.FEATURES,
  RouteLinks.PRICING,
  RouteLinks.ABOUT,
  RouteLinks.BLOG,
];

export const userDropdownLinks: LinkDto[] = [
  RouteLinks.ACCOUNT_DASHBOARD,
  RouteLinks.SETTINGS,
  RouteLinks.HOME,
];
export const adminHeaderLinks: LinkDto[] = [
  RouteLinks.ADMIN_OVERVIEW,
  RouteLinks.ADMIN_DASHBOARD_USERS,
  RouteLinks.ADMIN_DASHBOARD_ROLES,
  RouteLinks.ADMIN_DASHBOARD_PERMISSIONS,
  RouteLinks.ADMIN_DASHBOARD_PRODUCTS,
  RouteLinks.ADMIN_DASHBOARD_SUBSCRIPTIONS,
  RouteLinks.ADMIN_DASHBOARD_PLAN_FEATURES,
  RouteLinks.ADMIN_DASHBOARD_AI_USAGE,
  RouteLinks.ADMIN_DASHBOARD_JOBS,
  RouteLinks.ADMIN_DASHBOARD_BLOG,
  RouteLinks.ADMIN_DASHBOARD_MEDIA,
];

export const authenticatedSubHeaderLinks: LinkDto[] = [
  RouteLinks.DASHBOARD_OVERVIEW,
  RouteLinks.TASK_PROJECTS,
  RouteLinks.PROTECTED,
  RouteLinks.SETTINGS,
];
export const accountSidebarLinks: LinkDto[] = [
  RouteLinks.ACCOUNT_OVERVIEW_TEAMS,
  RouteLinks.ACCOUNT_OVERVIEW_TEAM_INVITATIONS,
];
export const settingsSidebarLinks: LinkDto[] = [
  RouteLinks.GENERAL_SETTINGS,
  RouteLinks.BILLING_SETTINGS,
  RouteLinks.POINTS_SETTINGS,
];

export const userDashboardLinks: LinkDto[] = [
  RouteLinks.ACCOUNT_DASHBOARD,
  RouteLinks.ACCOUNT_TEAMS,
  RouteLinks.ACCOUNT_ROCK_PAPER_SCISSORS,
  RouteLinks.ACCOUNT_FRIENDS,
  RouteLinks.SETTINGS,
];

export const teamLinks = (slug: string): LinkDto[] => [
  createTeamDashboardLink(slug),
  createTeamSettingsLink(slug),
  createTeamNotificationsLink(slug),
];

export const teamSettingLinks = (slug: string): LinkDto[] => [
  createTeamSettingsLink(slug),
  createTeamBillingSettingsLink(slug),
  createTeamMembersSettingsLink(slug),
];
export const teamNotifications = (slug: string): LinkDto[] => [
  createTeamNotificationsLink(slug),
];
export const createTeamDashboardLink = (slug: string) => {
  const path = `/teams/${slug}/dashboard`;
  return {
    to: path,
    title: "Team Dashboard",
    current: (pathname: string) => pathname.startsWith(path),
  };
};
export const createTeamProjectsLink = (slug: string) => {
  const path = `/teams/${slug}/projects`;
  return {
    to: path,
    title: "Team Projects",
    current: (pathname: string) => pathname.startsWith(path),
  };
};

export const createTeamSettingsLink = (slug: string) => {
  const path = `/teams/${slug}/settings`;
  return {
    to: path,
    title: "Team Settings",
    current: (pathname: string) => pathname.startsWith(path),
  };
};

export const createTeamNotificationsLink = (slug: string) => {
  const path = `/teams/${slug}/notifications`;
  return {
    to: path,
    title: "Notifications",
    current: (pathname: string) => pathname.startsWith(path),
  };
};

export const createTeamBillingSettingsLink = (slug: string) => {
  const path = `/teams/${slug}/settings/billing`;
  return {
    to: path,
    title: "Team Billing Settings",
    current: (pathname: string) => pathname.startsWith(path),
  };
};

export const createTeamMembersSettingsLink = (slug: string) => {
  const path = `/teams/${slug}/settings/members`;
  return {
    to: path,
    title: "Team Members Settings",
    current: (pathname: string) => pathname.startsWith(path),
  };
};
