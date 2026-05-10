import { useAuthProvider } from "@/hooks/use-auth-provider";
import { usePlayer } from "@/hooks/use-current-player";
import { client } from "@/lib/client";
import { ApiError } from "@/lib/error";
import { useQueryClient } from "@tanstack/react-query";
import { useEventSource } from "@react-nano/use-event-source";
import { useCallback, useEffect, useState } from "react";

async function issuePlayerSseTicket(token: string): Promise<string> {
  const { data, error } = await client.POST("/api/players/sse/ticket", {
    headers: { Authorization: `Bearer ${token}` },
  } as Parameters<typeof client.POST>[1]);
  if (error) throw ApiError.fromErrorModel(error);
  return (data as unknown as { ticket: string }).ticket;
}

export function usePlayerNotifications() {
  const { user } = useAuthProvider();
  const { player } = usePlayer();
  const queryClient = useQueryClient();
  const [sseUrl, setSseUrl] = useState<string | null>(null);

  useEffect(() => {
    if (!user?.tokens.access_token || !player) return;
    let cancelled = false;
    issuePlayerSseTicket(user.tokens.access_token)
      .then((t) => {
        if (!cancelled) {
          setSseUrl(`/api/players/${player.id}/sse?ticket=${t}`);
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
