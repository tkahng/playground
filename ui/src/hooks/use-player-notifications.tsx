import { useAuthProvider } from "@/hooks/use-auth-provider";
import { usePlayer } from "@/hooks/use-current-player";
import { friendsQueries } from "@/lib/friends-queries";
import { useQueryClient } from "@tanstack/react-query";
import { useEventSource } from "@react-nano/use-event-source";
import { useCallback, useEffect, useRef, useState } from "react";

export function usePlayerNotifications() {
  const { user } = useAuthProvider();
  const { player } = usePlayer();
  const queryClient = useQueryClient();

  // Track the token+playerId combination we've already issued a ticket for
  // so that object-reference churn on `player` doesn't re-issue tickets.
  const issuedForRef = useRef<string | null>(null);
  const [sseUrl, setSseUrl] = useState<string | null>(null);

  const token = user?.tokens.access_token;
  const playerId = player?.id;

  useEffect(() => {
    if (!token || !playerId) return;

    const key = `${token}:${playerId}`;
    if (issuedForRef.current === key) return; // already connected for this identity

    let cancelled = false;
    issuedForRef.current = key;

    friendsQueries
      .issuePlayerSseTicket({ token })
      .then(({ ticket }) => {
        if (!cancelled) {
          setSseUrl(`/api/players/${playerId}/sse?ticket=${ticket}`);
        }
      })
      .catch(() => {
        // Reset so the next mount can retry
        issuedForRef.current = null;
      });

    return () => {
      cancelled = true;
    };
  }, [token, playerId]);

  const onFriendRequest = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: [{ key: "friend-requests" }] });
  }, [queryClient]);

  useEventSource(sseUrl, "friend_request", onFriendRequest);
}
