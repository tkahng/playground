import { Button } from "@/components/ui/button";
import { useSSE } from "@/hooks/use-sse";
import { userReactionQueries } from "@/lib/api";
import {
  UserReaction,
  UserReactionsSseMessage,
  UserReactionsStats,
} from "@/schema.types";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { toast } from "sonner";

export default function SayHelloPage() {
  const [stats, setStats] = useState<UserReactionsStats | null>(null);
  const [latestReactions, setLatestReactions] = useState<UserReaction[]>([]);
  const { data: sseData, error: sseError } = useSSE<UserReactionsSseMessage>(
    `/api/user-reactions/sse`
  );
  const { data: statsData, isLoading: isStatsLoading } = useQuery({
    queryKey: ["user-reactions-stats"],
    queryFn: async () => {
      return userReactionQueries.getStats();
    },
  });
  const mutation = useMutation({
    mutationFn: async () => {
      return userReactionQueries.createReaction();
    },
    onSuccess: async () => {
      toast.success("Success");
    },
    onError: async () => {
      toast.error("Error");
    },
  });
  useEffect(() => {
    if (statsData) {
      setStats(statsData);
    }
  }, [statsData]);
  useEffect(() => {
    if (sseError) {
      toast.error("Error", {
        description: sseError,
        action: {
          label: "Close",
          onClick: () => console.log("Close"),
        },
      });
    } else if (sseData) {
      if (sseData.event === "latest_user_reaction_stats") {
        console.log(sseData);
        toast.success("Success", {
          description: sseData.event,
          action: {
            label: "Close",
            onClick: () => console.log("Close"),
          },
        });
        setStats(sseData.data.user_reaction_stats);
        if (sseData.data.user_reaction_stats.last_created) {
          const latestReaction = sseData.data.user_reaction_stats.last_created;
          setLatestReactions((prev) => [...[latestReaction], ...prev]);
        }
      }
    }

    return () => {};
  }, [sseData, sseError]);
  const onClick = () => {
    mutation.mutate();
  };
  return (
    <div>
      <div>SayHelloPage</div>
      <div>
        <Button onClick={onClick}>Say Hello</Button>
      </div>
      <div>
        {isStatsLoading ? (
          <div>Loading...</div>
        ) : (
          <>
            <div>Total Reactions: {stats?.total_reactions}</div>
            <div>Latest Reactions: From {stats?.last_created?.city}</div>
            <div>
              {latestReactions.map((r) => (
                <div key={r.id}>{r.city}</div>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
