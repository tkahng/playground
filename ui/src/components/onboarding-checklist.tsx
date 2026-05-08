import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useOnboardingProgress } from "@/hooks/use-onboarding-progress";
import { getUserTeamMembers } from "@/lib/team-queries";
import { useQuery } from "@tanstack/react-query";
import { CheckCircle2, Circle, ArrowRight, X } from "lucide-react";
import { useEffect } from "react";
import { Link } from "react-router";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { cn } from "@/lib/utils";

type StepId = "saidHello" | "hasTeam" | "hasProject" | "visitedPricing" | "visitedRps";

const STEPS: {
  id: StepId;
  label: string;
  description: string;
  to: string;
  cta: string;
}[] = [
  {
    id: "saidHello",
    label: "Say Hello to the world",
    description: "Wave to everyone — no sign-up needed.",
    to: "/say-hello",
    cta: "Say Hello",
  },
  {
    id: "hasTeam",
    label: "Create your first team",
    description: "Invite collaborators and manage projects together.",
    to: "/account/teams",
    cta: "Create a team",
  },
  {
    id: "hasProject",
    label: "Create a project",
    description: "Start a Kanban board inside your team to track work.",
    to: "/account/teams",
    cta: "Open your team",
  },
  {
    id: "visitedPricing",
    label: "Check out plans",
    description: "Subscribe to unlock protected routes — just for fun.",
    to: "/pricing",
    cta: "See plans",
  },
  {
    id: "visitedRps",
    label: "Play Rock Paper Scissors",
    description: "Challenge someone to a game and bet points.",
    to: "/rock-paper-scissors",
    cta: "Play now",
  },
];

export function OnboardingChecklist() {
  const { user } = useAuthProvider();
  const { progress, dismiss } = useOnboardingProgress();

  const { data: teamsData } = useQuery({
    queryKey: [
      {
        key: "get-user-team-members",
        user_id: user?.user.id,
        page: 0,
        per_page: 1,
      },
    ],
    queryFn: async () => {
      if (!user) throw new Error("No user");
      return getUserTeamMembers({
        token: user.tokens.access_token,
        page: 0,
        per_page: 1,
      });
    },
    enabled: !!user,
  });

  const hasTeam = (teamsData?.meta.total ?? 0) > 0;

  const completionMap: Record<StepId, boolean> = {
    saidHello: progress.saidHello,
    hasTeam,
    hasProject: progress.hasProject,
    visitedPricing: progress.visitedPricing,
    visitedRps: progress.visitedRps,
  };

  const completed = STEPS.filter((s) => completionMap[s.id]).length;
  const total = STEPS.length;
  const allDone = completed === total;

  useEffect(() => {
    if (allDone) {
      toast.success("You've explored everything!", {
        description: "All onboarding steps complete. Nice work 🎉",
      });
    }
  }, [allDone]);

  if (progress.dismissed) return null;

  return (
    <Card className="mb-6">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base font-semibold">
            Get started —{" "}
            <span className="text-primary">
              {completed}/{total}
            </span>{" "}
            steps done
          </CardTitle>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7 text-muted-foreground hover:text-foreground"
            onClick={dismiss}
            title="Dismiss"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
        <Progress value={(completed / total) * 100} className="h-1.5 mt-2" />
      </CardHeader>
      <CardContent className="pt-0">
        <ul className="space-y-1">
          {STEPS.map((step) => {
            const done = completionMap[step.id];
            return (
              <li
                key={step.id}
                className={cn(
                  "flex items-center justify-between rounded-lg px-3 py-2.5 transition-colors",
                  done ? "opacity-50" : "hover:bg-muted/50",
                )}
              >
                <div className="flex items-center gap-3 min-w-0">
                  {done ? (
                    <CheckCircle2 className="h-5 w-5 text-primary shrink-0" />
                  ) : (
                    <Circle className="h-5 w-5 text-muted-foreground shrink-0" />
                  )}
                  <div className="min-w-0">
                    <p
                      className={cn(
                        "text-sm font-medium",
                        done && "line-through text-muted-foreground",
                      )}
                    >
                      {step.label}
                    </p>
                    {!done && (
                      <p className="text-xs text-muted-foreground truncate">
                        {step.description}
                      </p>
                    )}
                  </div>
                </div>
                {!done && (
                  <Button
                    asChild
                    variant="ghost"
                    size="sm"
                    className="gap-1 text-xs shrink-0 ml-2"
                  >
                    <Link to={step.to}>
                      {step.cta}
                      <ArrowRight className="h-3 w-3" />
                    </Link>
                  </Button>
                )}
              </li>
            );
          })}
        </ul>
        {allDone && (
          <p className="text-center text-sm text-muted-foreground pt-3 border-t mt-3">
            🎉 You've explored everything! Nice work.
          </p>
        )}
      </CardContent>
    </Card>
  );
}
