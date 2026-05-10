import { useAuthProvider } from "@/hooks/use-auth-provider";
import { friendsQueries, Friendship, Player } from "@/lib/friends-queries";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { CenteredSpinner } from "@/components/centered-spinner";
import { Badge } from "@/components/ui/badge";
import { useState } from "react";
import { CreateGameDialog } from "@/pages/account/rock-paper-scissors/create-game-dialog";

function getOtherPlayer(friendship: Friendship, currentPlayerId: string): Player | null | undefined {
  if (friendship.requesting_player_id === currentPlayerId) return friendship.invited_player;
  return friendship.requesting_player;
}

function PlayerRow({ player }: { player: Player | null | undefined }) {
  if (!player) return <span className="text-muted-foreground">Unknown player</span>;
  return (
    <div>
      <p className="font-medium text-sm">{player.display_name || player.email}</p>
      {player.display_name && (
        <p className="text-xs text-muted-foreground">{player.email}</p>
      )}
    </div>
  );
}

function BlockConfirmButton({
  onConfirm,
  disabled,
}: {
  onConfirm: () => void;
  disabled?: boolean;
}) {
  const [confirming, setConfirming] = useState(false);
  if (confirming) {
    return (
      <div className="flex gap-1">
        <Button size="sm" variant="destructive" onClick={() => { setConfirming(false); onConfirm(); }} disabled={disabled}>
          Confirm
        </Button>
        <Button size="sm" variant="ghost" onClick={() => setConfirming(false)}>×</Button>
      </div>
    );
  }
  return (
    <Button size="sm" variant="outline" onClick={() => setConfirming(true)} disabled={disabled}>
      Block
    </Button>
  );
}

function FriendsList({ token, currentPlayerId }: { token: string; currentPlayerId: string }) {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: [{ key: "friends" }],
    queryFn: () => friendsQueries.listFriends({ token }),
  });

  const removeMutation = useMutation({
    mutationFn: (friendshipId: string) => friendsQueries.removeFriend({ token, friendshipId }),
    onSuccess: () => {
      toast.success("Friend removed");
      queryClient.invalidateQueries({ queryKey: [{ key: "friends" }] });
    },
    onError: (e) => toast.error(e.message),
  });

  const blockMutation = useMutation({
    mutationFn: (playerId: string) => friendsQueries.blockPlayer({ token, playerId }),
    onSuccess: () => {
      toast.success("Player blocked");
      queryClient.invalidateQueries({ queryKey: [{ key: "friends" }] });
    },
    onError: (e) => toast.error(e.message),
  });

  if (isLoading) return <CenteredSpinner />;

  const friends = data?.data ?? [];

  if (friends.length === 0) {
    return (
      <p className="text-muted-foreground text-sm py-4">
        No friends yet. Search for players by email to add them.
      </p>
    );
  }

  return (
    <ul className="divide-y">
      {friends.map((f) => {
        const other = getOtherPlayer(f, currentPlayerId);
        return (
          <li key={f.id} className="flex items-center justify-between py-3 gap-3">
            <PlayerRow player={other} />
            <div className="flex gap-2 items-center">
              {other && (
                <CreateGameDialog
                  initialPlayer={other}
                  trigger={<Button size="sm" variant="secondary">Challenge</Button>}
                />
              )}
              <BlockConfirmButton
                onConfirm={() => blockMutation.mutate(other?.id ?? "")}
                disabled={blockMutation.isPending || !other}
              />
              <Button
                size="sm"
                variant="destructive"
                onClick={() => removeMutation.mutate(f.id)}
                disabled={removeMutation.isPending}
              >
                Remove
              </Button>
            </div>
          </li>
        );
      })}
    </ul>
  );
}

function PendingRequestsList({ token, currentPlayerId }: { token: string; currentPlayerId: string }) {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: [{ key: "friend-requests" }],
    queryFn: () => friendsQueries.listFriendRequests({ token }),
  });

  const acceptMutation = useMutation({
    mutationFn: (requestId: string) => friendsQueries.acceptFriendRequest({ token, requestId }),
    onSuccess: () => {
      toast.success("Friend request accepted");
      queryClient.invalidateQueries({ queryKey: [{ key: "friend-requests" }] });
      queryClient.invalidateQueries({ queryKey: [{ key: "friends" }] });
    },
    onError: (e) => toast.error(e.message),
  });

  const declineMutation = useMutation({
    mutationFn: (requestId: string) => friendsQueries.declineFriendRequest({ token, requestId }),
    onSuccess: () => {
      toast.success("Request declined");
      queryClient.invalidateQueries({ queryKey: [{ key: "friend-requests" }] });
    },
    onError: (e) => toast.error(e.message),
  });

  const cancelMutation = useMutation({
    mutationFn: (friendshipId: string) => friendsQueries.removeFriend({ token, friendshipId }),
    onSuccess: () => {
      toast.success("Request cancelled");
      queryClient.invalidateQueries({ queryKey: [{ key: "friend-requests" }] });
    },
    onError: (e) => toast.error(e.message),
  });

  if (isLoading) return <CenteredSpinner />;

  const requests = data?.data ?? [];
  const incoming = requests.filter((f) => f.invited_player_id === currentPlayerId);
  const outgoing = requests.filter((f) => f.requesting_player_id === currentPlayerId);

  if (requests.length === 0) {
    return <p className="text-muted-foreground text-sm py-4">No pending requests.</p>;
  }

  return (
    <div className="space-y-4">
      {incoming.length > 0 && (
        <div>
          <h3 className="text-sm font-semibold mb-2 text-muted-foreground uppercase tracking-wide">
            Incoming
          </h3>
          <ul className="divide-y">
            {incoming.map((f) => (
              <li key={f.id} className="flex items-center justify-between py-3 gap-3">
                <PlayerRow player={f.requesting_player} />
                <div className="flex gap-2">
                  <Button size="sm" onClick={() => acceptMutation.mutate(f.id)} disabled={acceptMutation.isPending}>
                    Accept
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => declineMutation.mutate(f.id)} disabled={declineMutation.isPending}>
                    Decline
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}
      {outgoing.length > 0 && (
        <div>
          <h3 className="text-sm font-semibold mb-2 text-muted-foreground uppercase tracking-wide">
            Outgoing
          </h3>
          <ul className="divide-y">
            {outgoing.map((f) => (
              <li key={f.id} className="flex items-center justify-between py-3 gap-3">
                <PlayerRow player={f.invited_player} />
                <div className="flex items-center gap-2">
                  <Badge variant="secondary">Pending</Badge>
                  <Button size="sm" variant="ghost" onClick={() => cancelMutation.mutate(f.id)} disabled={cancelMutation.isPending}>
                    Cancel
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

export default function FriendsPage({ currentPlayerId }: { currentPlayerId: string }) {
  const { user } = useAuthProvider();
  const token = user?.tokens.access_token;
  const { data: requestsData } = useQuery({
    queryKey: [{ key: "friend-requests" }],
    enabled: !!token,
    queryFn: () => friendsQueries.listFriendRequests({ token: token! }),
  });

  const incomingCount = (requestsData?.data ?? []).filter(
    (f) => f.invited_player_id === currentPlayerId,
  ).length;

  if (!token) return null;

  return (
    <div>
      <Tabs defaultValue="friends">
        <TabsList>
          <TabsTrigger value="friends">Friends</TabsTrigger>
          <TabsTrigger value="requests">
            Requests
            {incomingCount > 0 && (
              <Badge className="ml-1.5 h-4 min-w-4 px-1 text-[10px]">{incomingCount}</Badge>
            )}
          </TabsTrigger>
        </TabsList>
        <TabsContent value="friends" className="mt-4">
          <FriendsList token={token} currentPlayerId={currentPlayerId} />
        </TabsContent>
        <TabsContent value="requests" className="mt-4">
          <PendingRequestsList token={token} currentPlayerId={currentPlayerId} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
