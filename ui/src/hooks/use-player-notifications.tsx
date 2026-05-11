import { useAuthProvider } from "@/hooks/use-auth-provider";
import { usePlayer } from "@/hooks/use-current-player";
import { friendsQueries } from "@/lib/friends-queries";
import { useQueryClient } from "@tanstack/react-query";
import {
  useEventSource,
  useEventSourceListener,
} from "@react-nano/use-event-source";
import { useCallback, useEffect, useRef, useState } from "react";

export function usePlayerNotifications() {
  const { user } = useAuthProvider();
  const { player } = usePlayer();
  const queryClient = useQueryClient();

  const issuedForRef = useRef<string | null>(null);
  const [sseUrl, setSseUrl] = useState<string | null>(null);

  const token = user?.tokens.access_token;
  const playerId = player?.id;

  useEffect(() => {
    if (!token || !playerId) return;

    const key = `${token}:${playerId}`;
    if (issuedForRef.current === key) return;

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
        issuedForRef.current = null;
      });

    return () => {
      cancelled = true;
    };
  }, [token, playerId]);

  // useEventSource requires a string; pass empty string when not yet connected
  // so the hook is always called unconditionally (React rules).
  const [eventSource] = useEventSource(sseUrl ?? "", false);

  // Only attach listeners when we have a real URL (and thus a real connection).
  const activeSource = sseUrl ? eventSource : null;

  const onFriendRequest = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: [{ key: "friend-requests" }] });
  }, [queryClient]);

  const onRpsGameEvent = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: [{ key: "rps-games" }] });
    queryClient.invalidateQueries({ queryKey: [{ key: "find-rps-game" }] });
  }, [queryClient]);

  const onRpsGameCompleted = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: [{ key: "rps-games" }] });
    queryClient.invalidateQueries({ queryKey: [{ key: "find-rps-game" }] });
    queryClient.invalidateQueries({ queryKey: [{ key: "ledger-balance" }] });
  }, [queryClient]);

  useEventSourceListener(
    activeSource,
    ["friend_request"],
    onFriendRequest,
    [onFriendRequest]
  );

  useEventSourceListener(
    activeSource,
    [
      "rps_game_challenged",
      "rps_rematch_requested",
      "rps_rematch_accepted",
      "rps_rematch_declined",
      "rps_rematch_expired",
    ],
    onRpsGameEvent,
    [onRpsGameEvent]
  );

  useEventSourceListener(
    activeSource,
    ["rps_game_completed"],
    onRpsGameCompleted,
    [onRpsGameCompleted]
  );
}
