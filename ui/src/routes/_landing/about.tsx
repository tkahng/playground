import LandingAboutPage from "@/pages/landing/about";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_landing/about")({
  component: LandingAboutPage,
});
