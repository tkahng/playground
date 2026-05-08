import AdminDashboardPage from "@/pages/admin/admin-dashboard";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authenticated/admin/_adminLayout/")({
  component: AdminDashboardPage,
});
