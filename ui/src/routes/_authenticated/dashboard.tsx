import DashboardLayout from "@/layouts/dashboard-layout";
import Dashboard from "@/pages/dasboard";
import { createFileRoute } from "@tanstack/react-router";

function DashboardPage() {
  return (
    <DashboardLayout>
      <Dashboard />
    </DashboardLayout>
  );
}

export const Route = createFileRoute("/_authenticated/dashboard")({
  component: DashboardPage,
});
