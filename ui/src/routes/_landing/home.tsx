import Landing from "@/pages/landing/landing";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_landing/home")({
  component: Landing,
});
