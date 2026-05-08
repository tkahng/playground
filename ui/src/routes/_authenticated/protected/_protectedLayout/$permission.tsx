import ProtectedRoutePage from "@/pages/protected-routes/protected-route-page";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/protected/_protectedLayout/$permission"
)({
  component: ProtectedRoutePage,
});
