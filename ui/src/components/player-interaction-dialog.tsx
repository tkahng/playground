import { useAuthProvider } from "@/hooks/use-auth-provider";
import { friendsQueries, Friendship } from "@/lib/friends-queries";
import { Player } from "@/schema.types";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Spinner } from "@/components/ui/spinner";
import { useState } from "react";
import { CreateGameDialog } from "@/pages/account/rock-paper-scissors/create-game-dialog";

interface PlayerInteractionDialogProps {
  player: Player;
  currentPlayerId?: string;
  children: React.ReactNode;
}

function BlockButton({
  onConfirm,
  disabled,
}: {
  onConfirm: () => void;
  disabled?: boolean;
}) {
  const [confirming, setConfirming] = useState(false);
  if (confirming) {
    return (
      <div className="flex gap-2">
        <Button size="sm" variant="destructive" onClick={() => { setConfirming(false); onConfirm(); }} disabled={disabled}>
          Confirm Block
        </Button>
        <Button size="sm" variant="ghost" onClick={() => setConfirming(false)}>
          Cancel
        </Button>
      </div>
    );
  }
  return (
    <Button size="sm" variant="outline" onClick={() => setConfirming(true)} disabled={disabled}>
      Block Player
    </Button>
  );
}

function FriendshipActions({
  friendship,
  player,
  currentPlayerId,
  token,
  onDone,
}: {
  friendship: Friendship | null | undefined;
  player: Player;
  currentPlayerId: string;
  token: string;
  onDone: () => void;
}) {
  const queryClient = useQueryClient();

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: [{ key: "friendship", playerId: player.id }] });
    queryClient.invalidateQueries({ queryKey: [{ key: "friends" }] });
    queryClient.invalidateQueries({ queryKey: [{ key: "friend-requests" }] });
  };

  const sendRequestMutation = useMutation({
    mutationFn: () => friendsQueries.sendFriendRequest({ token, invitedPlayerId: player.id }),
    onSuccess: () => { toast.success("Friend request sent"); invalidate(); onDone(); },
    onError: (e) => toast.error(e.message),
  });

  const acceptMutation = useMutation({
    mutationFn: () => friendsQueries.acceptFriendRequest({ token, requestId: friendship!.id }),
    onSuccess: () => { toast.success("Friend request accepted"); invalidate(); onDone(); },
    onError: (e) => toast.error(e.message),
  });

  const declineMutation = useMutation({
    mutationFn: () => friendsQueries.declineFriendRequest({ token, requestId: friendship!.id }),
    onSuccess: () => { toast.success("Friend request declined"); invalidate(); onDone(); },
    onError: (e) => toast.error(e.message),
  });

  const removeMutation = useMutation({
    mutationFn: () => friendsQueries.removeFriend({ token, friendshipId: friendship!.id }),
    onSuccess: () => { toast.success("Friend removed"); invalidate(); onDone(); },
    onError: (e) => toast.error(e.message),
  });

  const blockMutation = useMutation({
    mutationFn: () => friendsQueries.blockPlayer({ token, playerId: player.id }),
    onSuccess: () => { toast.success("Player blocked"); invalidate(); onDone(); },
    onError: (e) => toast.error(e.message),
  });

  const unblockMutation = useMutation({
    mutationFn: () => friendsQueries.unblockPlayer({ token, playerId: player.id }),
    onSuccess: () => { toast.success("Player unblocked"); invalidate(); onDone(); },
    onError: (e) => toast.error(e.message),
  });

  const isBusy =
    sendRequestMutation.isPending ||
    acceptMutation.isPending ||
    declineMutation.isPending ||
    removeMutation.isPending ||
    blockMutation.isPending ||
    unblockMutation.isPending;

  if (!friendship) {
    return (
      <div className="flex flex-col gap-2">
        <Button size="sm" onClick={() => sendRequestMutation.mutate()} disabled={isBusy}>
          Add Friend
        </Button>
        <BlockButton onConfirm={() => blockMutation.mutate()} disabled={isBusy} />
      </div>
    );
  }

  if (friendship.status === "blocked") {
    if (friendship.requesting_player_id === currentPlayerId) {
      return (
        <Button size="sm" variant="outline" onClick={() => unblockMutation.mutate()} disabled={isBusy}>
          Unblock
        </Button>
      );
    }
    return <p className="text-sm text-muted-foreground">This player is unavailable.</p>;
  }

  if (friendship.status === "accepted") {
    return (
      <div className="flex flex-col gap-2">
        <CreateGameDialog
          initialPlayer={player}
          trigger={<Button size="sm" onClick={onDone}>Challenge</Button>}
        />
        <Button size="sm" variant="outline" onClick={() => removeMutation.mutate()} disabled={isBusy}>
          Remove Friend
        </Button>
        <BlockButton onConfirm={() => blockMutation.mutate()} disabled={isBusy} />
      </div>
    );
  }

  if (friendship.status === "pending") {
    if (friendship.invited_player_id === currentPlayerId) {
      return (
        <div className="flex flex-col gap-2">
          <Button size="sm" onClick={() => acceptMutation.mutate()} disabled={isBusy}>
            Accept Request
          </Button>
          <Button size="sm" variant="ghost" onClick={() => declineMutation.mutate()} disabled={isBusy}>
            Decline
          </Button>
        </div>
      );
    }
    return (
      <div className="flex flex-col gap-2">
        <p className="text-sm text-muted-foreground">Request pending</p>
        <Button size="sm" variant="ghost" onClick={() => removeMutation.mutate()} disabled={isBusy}>
          Cancel Request
        </Button>
      </div>
    );
  }

  // Declined — allow re-sending
  return (
    <div className="flex flex-col gap-2">
      <Button size="sm" onClick={() => sendRequestMutation.mutate()} disabled={isBusy}>
        Add Friend
      </Button>
      <BlockButton onConfirm={() => blockMutation.mutate()} disabled={isBusy} />
    </div>
  );
}

export function PlayerInteractionDialog({
  player,
  currentPlayerId,
  children,
}: PlayerInteractionDialogProps) {
  const { user } = useAuthProvider();
  const [open, setOpen] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: [{ key: "friendship", playerId: player.id }],
    enabled: open && !!user?.tokens.access_token && !!currentPlayerId,
    staleTime: 30_000,
    queryFn: () =>
      friendsQueries.getFriendship({
        token: user!.tokens.access_token,
        playerId: player.id,
      }),
  });

  const displayName = player.display_name || player.email;

  if (!currentPlayerId || !user || player.id === currentPlayerId) {
    return <>{children}</>;
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <span className="cursor-pointer hover:underline">{children}</span>
      </PopoverTrigger>
      <PopoverContent className="w-60 p-3" side="top" align="start">
        <div className="space-y-3">
          <div>
            <p className="font-medium text-sm truncate">{displayName}</p>
            {player.display_name && (
              <p className="text-xs text-muted-foreground truncate">{player.email}</p>
            )}
          </div>
          <hr />
          {isLoading ? (
            <Spinner className="h-4 w-4" />
          ) : (
            <FriendshipActions
              friendship={data?.data ?? null}
              player={player}
              currentPlayerId={currentPlayerId}
              token={user.tokens.access_token}
              onDone={() => setOpen(false)}
            />
          )}
        </div>
      </PopoverContent>
    </Popover>
  );
}
