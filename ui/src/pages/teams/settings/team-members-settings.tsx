import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { DataTable } from "@/components/data-table";
import { teamSettingLinks } from "@/components/links";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeamContext } from "@/hooks/use-team-context";
import { getTeamMembers } from "@/lib/team-queries";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { useSearchParams } from "react-router";
import { TeamMemberActionDropdown } from "./team-member-action-dropdown";

export default function TeamMembersSettingPage() {
  const { user } = useAuthProvider();
  const { team } = useTeamContext();
  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);
  const onPaginationChange = (updater: Updater<PaginationState>) => {
    const newState =
      typeof updater === "function"
        ? updater({ pageIndex, pageSize })
        : updater;
    if (newState.pageIndex !== pageIndex || newState.pageSize !== pageSize) {
      setSearchParams({
        page: String(newState.pageIndex),
        per_page: String(newState.pageSize),
      });
    }
  };
  const { data, isPending, isError, error } = useQuery({
    queryKey: ["team-members"],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (!team?.id) {
        throw new Error("Current team member team ID is required");
      }
      return getTeamMembers(user.tokens.access_token, team.id);
    },
  });
  if (isPending) {
    return <div>Loading...</div>;
  }

  if (isError) {
    return <div>Error: {error.message}</div>;
  }
  if (!team) {
    return <div>Team not found</div>;
  }
  return (
    <div className="flex">
      <DashboardSidebar links={teamSettingLinks(team?.slug)} />
      <div className="flex-1 space-y-6 p-12 w-full">
        <DataTable
          data={data.data || []}
          rowCount={data.meta.total || 0}
          paginationState={{ pageIndex, pageSize }}
          onPaginationChange={onPaginationChange}
          paginationEnabled
          columns={[
            {
              header: "Name",
              accessorKey: "user.name",
            },
            {
              header: "Email",
              accessorKey: "user.email",
            },
            {
              header: "Role",
              accessorKey: "role",
            },
            {
              id: "actions",
              cell: ({ row }) => {
                return (
                  <div className="flex flex-row gap-2 justify-end">
                    <TeamMemberActionDropdown memberId={row.original.id} />
                  </div>
                );
              },
            },
          ]}
        />
      </div>
    </div>
  );
}
