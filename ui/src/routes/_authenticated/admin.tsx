import AdminLayoutBase from "@/layouts/admin-layout-base";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/admin")({
  component: AdminLayoutBase,
});
