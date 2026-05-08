import CallbackComponent from "@/pages/auth/callback";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/auth/callback")({
  component: CallbackComponent,
});
