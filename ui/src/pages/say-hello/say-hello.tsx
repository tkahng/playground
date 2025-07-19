import CountAnimation from "@/components/count-animation";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { userReactionQueries } from "@/lib/api";
import { UserReaction, UserReactionsStats } from "@/schema.types";
import {
  useEventSource,
  useEventSourceListener,
} from "@react-nano/use-event-source";
import { useMutation, useQuery } from "@tanstack/react-query";
import { AnimatePresence, motion } from "motion/react";
import { useEffect, useReducer } from "react";
import ReactCountryFlag from "react-country-flag";
import { toast } from "sonner";
type UserReactionsStatsWithReactions = UserReactionsStats & {
  last_reactions: UserReaction[];
};

export default function SayHelloPage() {
  function messageReducer(
    state: UserReactionsStatsWithReactions,
    action: UserReactionsStatsWithReactions
  ) {
    return {
      ...action,
      last_reactions: [
        ...(action.last_created ? [action.last_created] : []),
        ...state.last_reactions,
      ].slice(0, 5),
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
    <div className="flex flex-col">
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
            <CountAnimation number={stats.total_reactions} />
            <div className="flex grow">
              {stats.top_five_countries?.map((c, idx) => (
                <Card key={c.country} className="grow m-2">
                  <CardHeader>
                    <CardTitle>
                      #{idx + 1} {c.country}{" "}
                    </CardTitle>
                    <CardAction>
                      <ReactCountryFlag
                        countryCode={c.country}
                        svg
                        style={{
                          width: "2rem",
                          height: "2rem",
                        }}
                      />
                    </CardAction>
                  </CardHeader>
                  <CardContent>
                    Total Reactions: {c.total_reactions}
                  </CardContent>
                </Card>
              ))}
            </div>
            <div className="relative w-full max-w-sm overflow-hidden">
              <motion.div layout className="space-y-2 relative">
                <AnimatePresence initial={false}>
                  {stats.last_reactions?.map((record) => (
                    <motion.div
                      key={record.id}
                      layout
                      variants={{
                        hidden: { opacity: 0, x: -200 },
                        visible: { opacity: 1, y: 0 },
                        exit: { opacity: 0, y: 30, position: "absolute" },
                      }}
                      initial="hidden"
                      animate="visible"
                      exit="exit"
                      transition={{
                        type: "spring",
                        stiffness: 300,
                        damping: 25,
                      }}
                    >
                      <Card className="shadow-md border border-gray-200 bg-white dark:bg-neutral-900">
                        <CardContent className="p-4 text-gray-800 dark:text-gray-100">
                          {record.country}
                          {new Date(record.created_at).toUTCString()}
                        </CardContent>
                      </Card>
                    </motion.div>
                  ))}
                </AnimatePresence>
              </motion.div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
