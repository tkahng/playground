import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { teamQueries } from "@/lib/api";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Ellipsis, Pencil } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export function TeamNotificationActionDropdown({
  notificationId,
  read_at,
}: {
  notificationId: string;
  read_at?: string;
}) {
  const { user } = useAuthProvider();
  const queryClient = useQueryClient();
  const { teamMember } = useTeam();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const mutation = useMutation({
    mutationFn: (notificationId: string) => {
      return teamQueries.readTeamMemberNotification(
        user!.tokens?.access_token,
        teamMember!.id,
        notificationId
      );
    },
    onSuccess: async () => {
      setDropdownOpen(false);
      await queryClient.invalidateQueries({
        queryKey: ["team-member-notifications", teamMember?.id],
      });
    },
    onError: () => {
      toast.error("Failed to cancel invitation");
      setDropdownOpen(false);
    },
  });
  if (!notificationId) return null;
  return (
    <>
      <DropdownMenu open={dropdownOpen} onOpenChange={setDropdownOpen}>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon">
            <Ellipsis className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem
            disabled={!!read_at}
            onSelect={() => {
              setDropdownOpen(false);
              mutation.mutate(notificationId);
              // navigate(`${RouteMap.ADMIN_USERS}/${memberId}?tab=roles`);
            }}
          >
            <Button variant="ghost" size="sm">
              <Pencil className="h-4 w-4" />
              <span>Mark as read</span>
            </Button>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </>
  );
}
