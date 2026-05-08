import PricingPage from "@/pages/landing/pricing";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_landing/pricing")({
  component: PricingPage,
});
