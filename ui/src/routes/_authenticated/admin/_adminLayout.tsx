import AdminLayout from "@/layouts/admin-layout";
import { adminHeaderLinks } from "@/components/links";
import { createFileRoute } from "@tanstack/react-router";

function AdminLayoutWrapper() {
  return <AdminLayout headerLinks={adminHeaderLinks} />;
}

export const Route = createFileRoute("/_authenticated/admin/_adminLayout")({
  component: AdminLayoutWrapper,
});
