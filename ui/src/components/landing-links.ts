// import { LinkProps } from "react-router";
import { RouteMap } from "@/components/route-map";
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

export const adminLinks: LinkDto[] = [
  { to: RouteMap.DASHBOARD_HOME, title: "Dashboard" },
];
