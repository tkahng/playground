import LandingContactPage from "@/pages/landing/contact";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_landing/contact")({
  component: LandingContactPage,
});
