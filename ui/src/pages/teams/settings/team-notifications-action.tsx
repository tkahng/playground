import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import {
  deleteTeamMemberNotification,
  readTeamMemberNotification,
} from "@/lib/team-queries";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Ellipsis, CheckCheck, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export function TeamNotificationActionDropdown({
  notificationId,
  read_at,
}: {
  notificationId: string;
  read_at?: string | null;
}) {
  const { user } = useAuthProvider();
  const queryClient = useQueryClient();
  const { teamMember } = useTeam();
  const [dropdownOpen, setDropdownOpen] = useState(false);

  const invalidate = () =>
    queryClient.invalidateQueries({
      queryKey: ["team-member-notifications", teamMember?.id],
    });

  const markReadMutation = useMutation({
    mutationFn: (id: string) =>
      readTeamMemberNotification(user!.tokens?.access_token, teamMember!.id, id),
    onSuccess: async () => {
      setDropdownOpen(false);
      await invalidate();
      await queryClient.invalidateQueries({
        queryKey: ["team-member-notifications-unread-count", teamMember?.id],
      });
    },
    onError: () => {
      toast.error("Failed to mark as read");
      setDropdownOpen(false);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) =>
      deleteTeamMemberNotification(
        user!.tokens?.access_token,
        teamMember!.id,
        id
      ),
    onSuccess: async () => {
      setDropdownOpen(false);
      await invalidate();
      await queryClient.invalidateQueries({
        queryKey: ["team-member-notifications-unread-count", teamMember?.id],
      });
    },
    onError: () => {
      toast.error("Failed to delete notification");
      setDropdownOpen(false);
    },
  });

  if (!notificationId) return null;

  return (
    <DropdownMenu open={dropdownOpen} onOpenChange={setDropdownOpen}>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon">
          <Ellipsis className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem
          disabled={!!read_at || markReadMutation.isPending}
          onSelect={() => markReadMutation.mutate(notificationId)}
        >
          <CheckCheck className="mr-2 h-4 w-4" />
          Mark as read
        </DropdownMenuItem>
        <DropdownMenuItem
          className="text-destructive focus:text-destructive"
          disabled={deleteMutation.isPending}
          onSelect={() => deleteMutation.mutate(notificationId)}
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
