import DashboardLayout from "@/layouts/dashboard-layout";
import { userDashboardLinks } from "@/components/links";
import { createFileRoute, Outlet } from "@tanstack/react-router";

function AccountLayout() {
  return (
    <DashboardLayout headerLinks={userDashboardLinks}>
      <Outlet />
    </DashboardLayout>
  );
}

export const Route = createFileRoute("/_authenticated/account/_accountLayout")({
  component: AccountLayout,
});
