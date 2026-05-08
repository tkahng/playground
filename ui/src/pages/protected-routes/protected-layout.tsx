import { CenteredSpinner } from "@/components/centered-spinner";
import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { permissionsList } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { Outlet } from "@tanstack/react-router";

export default function ProtectedRouteLayout() {
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["permissions-list"],
    queryFn: async () => {
      return permissionsList();
    },
  });
  if (isLoading) {
    return <CenteredSpinner />;
  }

  if (isError) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div className="flex">
      <DashboardSidebar
        links={
          data?.data?.map((p) => {
            return {
              title: p.name,
              to: `/protected/${p.name}`,
            };
          }) || []
        }
      />
      <div className="flex-1 space-y-6 p-12 w-full">
        <Outlet />
      </div>
    </div>
  );
}
