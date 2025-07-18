import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { userReactionQueries } from "@/lib/api";
import { UserReaction, UserReactionsStats } from "@/schema.types";
import {
  useEventSource,
  useEventSourceListener,
} from "@react-nano/use-event-source";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useEffect, useReducer } from "react";
import { toast } from "sonner";

type UserReactionsStats2 = UserReactionsStats & {
  last_reactions: UserReaction[];
};

// type UserReactionStatsSse = {
//   user_reaction_stats: UserReactionsStats2;
// };
export default function SayHelloPage() {
  // const [stats, setStats] = useState<UserReactionsStats2 | null>(null);
  function messageReducer(
    state: UserReactionsStats2,
    action: UserReactionsStats2
  ) {
    const prevLastReactions = state.last_reactions;
    if (action.last_created) {
      if (!prevLastReactions.some((r) => r.id === action.last_created?.id)) {
        prevLastReactions.push(action.last_created);
      }
    }
    return {
      ...action,
      last_reactions: prevLastReactions,
    };
  }
  const [stats, updateStats] = useReducer(messageReducer, {
    top_five_countries: [],
    total_reactions: 0,
    last_reactions: [],
  });

  const [eventSource] = useEventSource("api/user-reactions/sse", false);
  useEventSourceListener(
    eventSource,
    ["latest_user_reaction_stats"],
    (evt) => {
      updateStats(JSON.parse(evt.data)?.user_reaction_stats);
    },
    [updateStats]
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
    if (statsData) updateStats({ ...statsData, last_reactions: [] });
  }, [statsData]);

  const onClick = () => {
    mutation.mutate();
  };
  if (isStatsLoading) {
    return <div>Loading...</div>;
  }
  return (
    <div className="w-1/2 ml-auto mr-auto">
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
              {stats.last_reactions?.map((r) => (
                <Card key={r.id} className="bg-primary">
                  <div>{r.city}</div>
                  <div>{r.created_at}</div>
                </Card>
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
