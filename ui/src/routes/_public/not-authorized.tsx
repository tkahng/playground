import NotAuthorizedPage from "@/pages/not-authorized";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_public/not-authorized")({
  component: NotAuthorizedPage,
});
