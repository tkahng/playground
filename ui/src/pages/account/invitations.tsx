import { CreateTeamDialog } from "@/components/create-team-dialog";
import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { DataTable } from "@/components/data-table";
import { accountSidebarLinks } from "@/components/links";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getUserTeamInvitations } from "@/lib/api";
import { GetError } from "@/lib/get-error";
import { Team } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { NavLink, useSearchParams } from "react-router";
import { toast } from "sonner";

export default function InvitationsPage() {
  // const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);

  const onPaginationChange = (updater: Updater<PaginationState>) => {
    const newState =
      typeof updater === "function"
        ? updater({ pageIndex, pageSize })
        : updater;
    setSearchParams({
      page: String(newState.pageIndex),
      per_page: String(newState.pageSize),
    });
  };
  const { user } = useAuthProvider();
  const { data, error, isError, isLoading } = useQuery({
    queryKey: ["user-invitations"],
    queryFn: async () => {
      if (!user) {
        throw new Error("User not found");
      }

      // const stats = await getStats(user.tokens.access_token);
      const teams = await getUserTeamInvitations(
        user.tokens.access_token,
        pageIndex,
        pageSize
      );
      return teams;
    },
  });

  const handleSelectTeam = (team: Team) => {
    toast.success(`Selected team: ${team.name}`);
  };

  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div>Error: {error?.message}</div>;
  }
  if (!data) {
    return <div>No data</div>;
  }
  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    const err = GetError(error);
    return <div>Error: {err?.detail}</div>;
  }
  return (
    <div className="flex">
      <DashboardSidebar links={accountSidebarLinks} />
      <div className="flex-1 space-y-6 p-12 w-full">
        <div className="mx-auto px-8 py-8 justify-start items-stretch flex-1 max-w-[1200px]">
          <div className="flex items-center justify-between">
            <p>Create and manage Teams for your applications.</p>
            <CreateTeamDialog />
          </div>

          <DataTable
            columns={[
              {
                accessorKey: "name",
                header: "Name",
                cell: ({ row }) => {
                  return (
                    <NavLink
                      to={{
                        pathname: `/team-invitation`,
                        search: `?token=${row.original.token}`,
                      }}
                      className="hover:underline text-blue-500"
                      onClick={() => handleSelectTeam(row.original.team!)}
                    >
                      {row.original.team?.name}
                    </NavLink>
                  );
                },
              },
              {
                accessorKey: "role",
                header: "Member Role",
                cell: ({ row }) => {
                  return (
                    <span className="text-gray-500">
                      {row.original.role || "Member"}
                    </span>
                  );
                },
              },
            ]}
            data={data?.data || []}
            rowCount={data?.meta.total || 0}
            paginationState={{ pageIndex, pageSize }}
            onPaginationChange={onPaginationChange}
            paginationEnabled
          />
        </div>
      </div>
    </div>
  );
}
