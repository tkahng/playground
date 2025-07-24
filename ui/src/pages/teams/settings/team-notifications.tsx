import { DataTable } from "@/components/data-table";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { teamQueries } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { CheckCircle, Circle } from "lucide-react";
import { useSearchParams } from "react-router";
import { TeamNotificationActionDropdown } from "./team-notifications-action";
export default function TeamNotifications() {
  const { user } = useAuthProvider();
  const { teamMember, team } = useTeam();
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
    queryKey: ["team-member-notifications", teamMember?.id],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      if (!teamMember?.id) {
        throw new Error("Current team member team ID is required");
      }
      const notifications = await teamQueries.getTeamMemberNotifications(
        user.tokens.access_token,
        teamMember.id,
        pageIndex,
        pageSize
      );
      const data = notifications.data?.map((n) => {
        const payload = JSON.parse(n.payload) as {
          notification: {
            title: string;
            body: string;
          };
          data: Record<string, unknown>;
        };
        return {
          ...n,
          payload,
        };
      });
      return {
        data,
        meta: notifications.meta,
      };
    },
  });
  if (isPending) {
    return <div>Loading...</div>;
  }

  if (isError) {
    return <div>Error: {error.message}</div>;
  }
  if (!teamMember || !team) {
    return <div>Team not found</div>;
  }
  return (
    <div className="flex">
      {/* <DashboardSidebar links={teamNotifications(team?.slug)} /> */}
      <div className="flex-1 space-y-6 p-12 w-full">
        <div className="flex items-center justify-between">
          <p>Manage your notifications..</p>
        </div>
        <DataTable
          data={data.data || []}
          rowCount={data.meta.total || 0}
          paginationState={{ pageIndex, pageSize }}
          onPaginationChange={onPaginationChange}
          paginationEnabled
          columns={[
            {
              header: "Read",
              accessorKey: "read_at",
              cell: ({ row }) => {
                return (
                  <div className="flex items-center justify-end">
                    {row.original.read_at ? (
                      <CheckCircle className="text-green-500" />
                    ) : (
                      <Circle className="text-gray-500" />
                    )}
                  </div>
                );
              },
            },
            {
              header: "Title",
              accessorKey: "payload.notification.title",
            },
            {
              header: "Body",
              accessorKey: "payload.notification.body",
            },
            {
              id: "actions",
              cell: ({ row }) => {
                return (
                  <div className="flex flex-row gap-2 justify-end">
                    <TeamNotificationActionDropdown
                      notificationId={row.original.id}
                      read_at={row.original.read_at}
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
