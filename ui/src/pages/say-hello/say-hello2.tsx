import { CenteredSpinner } from "@/components/centered-spinner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useOnboardingProgress } from "@/hooks/use-onboarding-progress";
import { ApiError } from "@/lib/error";
import { getCountryName } from "@/lib/get-country-name";
import { userReactionQueries } from "@/lib/user-reaction-queries";
import { UserReactionsStatsWithReactions } from "@/schema.types";
import {
  useEventSource,
  useEventSourceListener,
} from "@react-nano/use-event-source";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useGeolocation } from "@uidotdev/usehooks";
import { Clock, Globe } from "lucide-react";
import { AnimatePresence, motion } from "motion/react";
import { useEffect, useReducer } from "react";
import TimeAgo from "react-timeago";
import { toast } from "sonner";
import { useNavigate } from "@tanstack/react-router";

const maxItems = 3;
function messageReducer(
  state: UserReactionsStatsWithReactions,
  action: UserReactionsStatsWithReactions,
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
  const state = useGeolocation();
  const { markStep, progress } = useOnboardingProgress();
  const { user } = useAuthProvider();
  const navigate = useNavigate();
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
    [updateStats],
  );
  const { data: statsData, isLoading: isStatsLoading } = useQuery({
    queryKey: ["user-reactions-stats"],
    queryFn: async () => {
      return userReactionQueries.getStats();
    },
  });
  const mutation = useMutation({
    mutationFn: async () => {
      return userReactionQueries.createReaction({
        coords:
          state.latitude && state.longitude
            ? { latitude: state.latitude, longitude: state.longitude }
            : undefined,
      });
    },
    onSuccess: async () => {
      toast.success("Hello sent!");
      if (!progress.saidHello) {
        const nextPath = user ? "/account/teams" : "/signup";
        const nextLabel = user ? "Create a team →" : "Sign up →";
        toast("Ready to go further?", {
          description: "Build a team and start collaborating on projects.",
          action: { label: nextLabel, onClick: () => navigate(nextPath) },
        });
      }
      markStep("saidHello");
    },
    onError: async (err: ApiError) => {
      switch (err.status) {
        case 400:
          toast.error("We are having trouble locating you", {
            description: "Please make sure to allow location permissions",
          });
          break;
        case 429:
          toast.error("Too many hellos!", {
            description: "Please try again after a breather or two.",
          });
          break;
        default:
          toast.error("Opps, something went wrong", {
            description: "Please try again later.",
          });
          break;
      }
    },
  });
  useEffect(() => {
    if (statsData) updateStats({ ...statsData, last_reactions: [] });
  }, [statsData]);

  const handleSayHello = () => {
    mutation.mutate();
  };
  if (state.loading) {
    return <CenteredSpinner />;
  }

  if (isStatsLoading) {
    return <CenteredSpinner />;
  }

  return (
    <div className="min-h-screen bg-secondary p-4">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="text-center mb-8 pt-8">
          <h1 className="text-4xl md:text-6xl font-bold text-primary mb-4">
            Say Hello
          </h1>
          <p className="text-lg text-muted-foreground">
            Join people around the world in spreading positivity
          </p>
        </div>

        {/* Global Counter */}
        <Card className="mb-8 backdrop-blur-sm border-0 shadow-lg">
          <CardContent className="space-x-4">
            <div className="text-center">
              <div className="flex items-center justify-center gap-2 mb-2">
                <Globe className="h-6 w-6" />
                <span className="text-lg font-medium text-primary">
                  Global Hellos
                </span>
              </div>
              <div className="text-5xl md:text-7xl font-bold text-accent-foreground mb-2">
                {stats.total_reactions.toLocaleString()}
              </div>
              <p className="text-muted-foreground">hellos shared worldwide</p>
            </div>
            {/* Say Hello Button */}
            <div className="flex justify-center">
              <Button
                onClick={handleSayHello}
                size="lg"
                className="text-2xl px-12 py-8 h-auto bg-gradient-to-r from-yellow-400 to-orange-500 hover:from-yellow-500 hover:to-orange-600 transform hover:scale-105 transition-all duration-200 shadow-xl"
                asChild
              >
                <motion.div
                  initial={{ opacity: 0 }}
                  whileHover={{ backgroundColor: "rgba(220, 220, 220, 1)" }}
                  whileTap={{ backgroundColor: "rgba(255, 255, 255, 1)" }}
                  whileInView={{ opacity: 1 }}
                >
                  {/* 👋 Say Hello */}
                  Say Hello
                </motion.div>
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Latest Hellos */}
        <Card className="mb-8  backdrop-blur-sm border-0 shadow-lg">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Clock className="h-5 w-5" />
              Latest Hellos
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {stats.last_reactions?.length && (
                <motion.div layout className="space-y-2 relative">
                  <AnimatePresence initial={false}>
                    {stats.last_reactions?.map((hello) => {
                      const countryName = getCountryName(hello.country);
                      return (
                        <motion.div
                          key={hello.id}
                          layout
                          variants={{
                            hidden: { opacity: 0, y: -200 },
                            visible: { opacity: 1, y: 0 },
                            exit: {
                              opacity: 0,
                              y: 30,
                              position: "absolute",
                            },
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
                          <div
                            key={hello.id}
                            className="flex items-center justify-between p-3 bg-secondary rounded-lg hover:bg-secondary transition-colors"
                          >
                            <div className="flex items-center gap-3">
                              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                              <span className="font-medium text-secondary-foreground">
                                Someone from {countryName}
                              </span>
                            </div>
                            <span className="text-sm text-muted-foreground">
                              <TimeAgo date={hello.created_at} />
                            </span>
                          </div>
                        </motion.div>
                      );
                    })}
                  </AnimatePresence>
                </motion.div>
              )}
            </div>
            {stats.last_reactions.length === 0 && (
              <div className="text-center py-8 text-muted-foreground">
                No hellos yet. Be the first to say hello!
              </div>
            )}
          </CardContent>
        </Card>
        {/* Top 5 Countries */}
        <Card className="mb-8  backdrop-blur-sm border-0 shadow-lg">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <span className="text-xl">🏆</span>
              Top 5 Countries
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {stats.top_five_countries?.map(
                ({ country, total_reactions: count }, index) => (
                  <div
                    key={country}
                    className="flex items-center justify-between p-3 rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-gradient-to-r from-yellow-400 to-orange-500 text-white font-bold text-sm">
                        {index + 1}
                      </div>
                      <span className="font-medium text-secondary-foreground">
                        {country}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-lg font-bold">{count}</span>
                      <span className="text-sm text-muted-foreground">
                        hellos
                      </span>
                    </div>
                  </div>
                ),
              )}
            </div>
          </CardContent>
        </Card>

        {/* Footer */}
        <div className="text-center mt-8 pb-8">
          <p className="text-muted-foreground">
            Spread kindness, one hello at a time 💙
          </p>
        </div>
      </div>
    </div>
  );
}
