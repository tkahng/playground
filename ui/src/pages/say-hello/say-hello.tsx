import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardTitle } from "@/components/ui/card";
import { userReactionQueries } from "@/lib/api";
import { getCountryName } from "@/lib/get-country-name";
import { UserReactionsStatsWithReactions } from "@/schema.types";
import {
  useEventSource,
  useEventSourceListener,
} from "@react-nano/use-event-source";
import { useMutation, useQuery } from "@tanstack/react-query";
import { AnimatePresence, motion } from "motion/react";
import { useEffect, useReducer } from "react";
import TimeAgo from "react-timeago";
import { toast } from "sonner";
import { TopCountryCard } from "./top-country";

const maxItems = 3;
function messageReducer(
  state: UserReactionsStatsWithReactions,
  action: UserReactionsStatsWithReactions
) {
  return {
    ...state,
    ...action,
    last_reactions: [
      ...(!!action.last_created &&
      !state.last_reactions.some((r) => r.id === action.last_created?.id)
        ? [action.last_created]
        : []),
      ...state.last_reactions,
    ].slice(0, maxItems),
  };
}
export default function SayHelloPage() {
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
    <div className="flex flex-col mx-auto max-w-4xl">
      <Card className="items-center">
        <CardTitle>Say Hello</CardTitle>
        <Button onClick={onClick} asChild>
          <motion.div
            initial={{ opacity: 0 }}
            whileHover={{ scale: 1.2, opacity: 0.8 }}
            whileTap={{ scale: 0.8, rotate: 60 }}
            whileInView={{ opacity: 1 }}
          >
            Say Hello
          </motion.div>
        </Button>
        <CardContent>
          <p>Total Reactions: {stats.total_reactions}</p>
        </CardContent>
        <CardFooter>
          <p>Card Footer</p>
        </CardFooter>
      </Card>
      <div>
        {isStatsLoading ? (
          <div>Loading...</div>
        ) : (
          <>
            <div className="flex grow">
              {stats.top_five_countries?.map((c, idx) => {
                return (
                  <TopCountryCard
                    key={c.country}
                    className="grow m-2"
                    country={{
                      ...c,
                      countryName: getCountryName(c.country) || "",
                      number: idx,
                    }}
                  />
                );
              })}
            </div>
            <div className="flex items-center justify-center space-x-2 py-4">
              <div className="relative w-full max-w-sm overflow-hidden">
                <motion.div layout className="space-y-2 relative">
                  <AnimatePresence initial={false}>
                    {stats.last_reactions?.map((record) => {
                      const countryName = getCountryName(record.country);
                      return (
                        <motion.div
                          key={record.id}
                          layout
                          variants={{
                            hidden: { opacity: 0, y: -200 },
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
                              {countryName} <TimeAgo date={record.created_at} />
                            </CardContent>
                          </Card>
                        </motion.div>
                      );
                    })}
                  </AnimatePresence>
                </motion.div>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
