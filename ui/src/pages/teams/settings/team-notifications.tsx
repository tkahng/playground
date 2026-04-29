import { CenteredSpinner } from "@/components/centered-spinner";
import { DataTable } from "@/components/data-table";
import { Button } from "@/components/ui/button";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import {
  getTeamMemberNotifications,
  markAllTeamMemberNotificationsRead,
} from "@/lib/team-queries";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { CheckCheck, CheckCircle, Circle } from "lucide-react";
import { useSearchParams } from "react-router";
import { toast } from "sonner";
import { TeamNotificationActionDropdown } from "./team-notifications-action";

export default function TeamNotifications() {
  const { user } = useAuthProvider();
  const { teamMember, team } = useTeam();
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
      "team-member-notifications",
      teamMember?.id,
      pageIndex,
      pageSize,
    ],
    queryFn: async () => {
      if (!user?.tokens.access_token) throw new Error("Missing access token");
      if (!teamMember?.id) throw new Error("Current team member ID is required");
      const notifications = await getTeamMemberNotifications(
        user.tokens.access_token,
        teamMember.id,
        pageIndex,
        pageSize
      );
      const data = notifications.data?.map((n) => {
        const payload = JSON.parse(n.payload) as {
          notification: { title: string; body: string };
          data: Record<string, unknown>;
        };
        return { ...n, payload };
      });
      return { data, meta: notifications.meta };
    },
  });

  const markAllReadMutation = useMutation({
    mutationFn: () =>
      markAllTeamMemberNotificationsRead(
        user!.tokens.access_token,
        teamMember!.id
      ),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["team-member-notifications", teamMember?.id],
      });
      await queryClient.invalidateQueries({
        queryKey: ["team-member-notifications-unread-count", teamMember?.id],
      });
      toast.success("All notifications marked as read");
    },
    onError: () => toast.error("Failed to mark all as read"),
  });

  if (isPending) return <CenteredSpinner />;
  if (isError) return <div>Error: {error.message}</div>;
  if (!teamMember || !team) return <div>Team not found</div>;

  return (
    <div className="flex">
      <div className="flex-1 space-y-6 p-12 w-full">
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Manage your notifications.
          </p>
          <Button
            variant="outline"
            size="sm"
            onClick={() => markAllReadMutation.mutate()}
            disabled={markAllReadMutation.isPending}
          >
            <CheckCheck className="mr-2 h-4 w-4" />
            Mark all as read
          </Button>
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
              cell: ({ row }) => (
                <div className="flex items-center justify-center">
                  {row.original.read_at ? (
                    <CheckCircle className="h-4 w-4 text-green-500" />
                  ) : (
                    <Circle className="h-4 w-4 text-muted-foreground" />
                  )}
                </div>
              ),
            },
            {
              header: "Title",
              accessorKey: "payload.notification.title",
              cell: ({ row }) => (
                <span className={row.original.read_at ? "" : "font-semibold"}>
                  {row.original.payload.notification.title}
                </span>
              ),
            },
            {
              header: "Body",
              accessorKey: "payload.notification.body",
            },
            {
              id: "actions",
              cell: ({ row }) => (
                <div className="flex flex-row gap-2 justify-end">
                  <TeamNotificationActionDropdown
                    notificationId={row.original.id}
                    read_at={row.original.read_at}
                  />
                </div>
              ),
            },
          ]}
        />
      </div>
    </div>
  );
}
