import { useAuthProvider } from "@/hooks/use-auth-provider";
import { usePlayer } from "@/hooks/use-current-player";
import { friendsQueries } from "@/lib/friends-queries";
import { useQueryClient } from "@tanstack/react-query";
import { useEventSource } from "@react-nano/use-event-source";
import { useCallback, useEffect, useState } from "react";

export function usePlayerNotifications() {
  const { user } = useAuthProvider();
  const { player } = usePlayer();
  const queryClient = useQueryClient();
  const [sseUrl, setSseUrl] = useState<string | null>(null);

  useEffect(() => {
    if (!user?.tokens.access_token || !player) return;
    let cancelled = false;
    friendsQueries
      .issuePlayerSseTicket({ token: user.tokens.access_token })
      .then(({ ticket }) => {
        if (!cancelled) {
          setSseUrl(`/api/players/${player.id}/sse?ticket=${ticket}`);
        }
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, [user?.tokens.access_token, player]);

  const onFriendRequest = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: [{ key: "friend-requests" }] });
  }, [queryClient]);

  useEventSource(sseUrl, "friend_request", onFriendRequest);
}
