import ProtectedRouteIndex from "@/pages/protected-routes/route-index";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/_authenticated/protected/_protectedLayout/"
)({
  component: ProtectedRouteIndex,
});
