// import { LinkProps } from "react-router";
import { JSX } from "react";
import { RouteMap } from "./route-map";
export type LinkProps = {
  title?: string;
  icon?: JSX.Element;
  to: string;
};
export const landingLinks: LinkProps[] = [
  { to: RouteMap.FEATURES, title: "Features" },
  { to: RouteMap.PRICING, title: "Pricing" },
  { to: RouteMap.ABOUT, title: "About" },
  { to: RouteMap.CONTACT, title: "Contact" },
];
