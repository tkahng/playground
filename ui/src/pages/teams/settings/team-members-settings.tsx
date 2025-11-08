import { DashboardSidebar } from "@/components/dashboard-sidebar";
import { DataTable } from "@/components/data-table";
import { teamSettingLinks } from "@/components/links";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { deleteMember, getTeamTeamMembers } from "@/lib/team-queries";
import { MemberDeleteButton } from "@/pages/teams/settings/member-delete-dialog";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { useSearchParams } from "react-router";
import { toast } from "sonner";
import { InviteTeamMemberDialog } from "./invite-team-member-dialog";

export default function TeamMembersSettingPage() {
  const { user } = useAuthProvider();
  const { team } = useTeam();
  const queryClient = useQueryClient();
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
    queryKey: [
      {
        key: "team-team-members",
        team_id: team?.id,
        page: pageIndex,
        per_page: pageSize,
        active: true,
      },
    ],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (!team?.id) {
        throw new Error("Current team member team ID is required");
      }
      return getTeamTeamMembers({
        token: user.tokens.access_token,
        teamId: team.id,
        page: pageIndex,
        perPage: pageSize,
        active: true,
      });
    },
  });
  const mutation = useMutation({
    mutationFn: async (memberId: string) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      await deleteMember({
        memberId: memberId,
        token: user.tokens.access_token,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [{ key: "team-team-members" }],
      });
      toast.success("Member deleted successfully");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Failed to delete member");
    },
  });
  const onDelete = (memberId: string) => {
    mutation.mutate(memberId);
  };
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
        <div className="flex items-center justify-between">
          <p>
            Manage your team's members. Invite team members to join your team.
          </p>
          <InviteTeamMemberDialog />
        </div>
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
                    <MemberDeleteButton
                      memberId={row.original.id}
                      onDelete={onDelete}
                    />
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
