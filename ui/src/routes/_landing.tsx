import RootLayout from "@/layouts/root";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_landing")({
  component: RootLayout,
});
