import Features from "@/pages/landing/features";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_landing/features")({
  component: Features,
});
