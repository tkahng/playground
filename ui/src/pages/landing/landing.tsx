import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { userReactionQueries } from "@/lib/user-reaction-queries";
import { UserReactionsStatsWithReactions } from "@/schema.types";
import {
  useEventSource,
  useEventSourceListener,
} from "@react-nano/use-event-source";
import { useQuery } from "@tanstack/react-query";
import {
  ArrowRight,
  CreditCard,
  Globe,
  Hand,
  Scissors,
  Users,
} from "lucide-react";
import { useEffect, useReducer } from "react";
import { JSX } from "react";
import { Link } from "react-router";
import {
  SiDocker,
  SiGithub,
  SiGithubactions,
  SiGo,
  SiPostgresql,
  SiReact,
  SiTailscale,
  SiTailwindcss,
  SiTypescript,
} from "@icons-pack/react-simple-icons";

export const techIconList: { icon: JSX.Element; name: string }[] = [
  { icon: <SiGo />, name: "Go" },
  { icon: <SiPostgresql />, name: "Postgres" },
  { icon: <SiReact />, name: "React" },
  { icon: <SiTypescript />, name: "Typescript" },
  { icon: <SiTailwindcss />, name: "Tailwind" },
  { icon: <SiDocker />, name: "Docker" },
  { icon: <SiGithub />, name: "Github" },
  { icon: <SiGithubactions />, name: "Github Actions" },
  { icon: <SiTailscale />, name: "Tailscale" },
];

const journeySteps = [
  {
    step: 1,
    icon: <Hand className="h-6 w-6" />,
    title: "Say Hello",
    description: "Wave to the world — no sign-up needed. One click, instant.",
    cta: "Try it now",
    to: "/say-hello",
    noAuth: true,
  },
  {
    step: 2,
    icon: <Users className="h-6 w-6" />,
    title: "Build a Team",
    description:
      "Create a team, invite members, spin up a Kanban project board.",
    cta: "Get started",
    to: "/signup",
    noAuth: false,
  },
  {
    step: 3,
    icon: <CreditCard className="h-6 w-6" />,
    title: "Pick a Plan",
    description:
      "Subscribe to unlock protected routes. No real money required — just for fun.",
    cta: "See plans",
    to: "/pricing",
    noAuth: false,
  },
  {
    step: 4,
    icon: <Scissors className="h-6 w-6" />,
    title: "Play a Game",
    description:
      "Challenge anyone to Rock Paper Scissors. Share a link, place bets with points.",
    cta: "Play now",
    to: "/rock-paper-scissors",
    noAuth: true,
  },
];

function messageReducer(
  state: UserReactionsStatsWithReactions,
  action: UserReactionsStatsWithReactions,
) {
  return { ...state, ...action };
}

function LiveHelloCount() {
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

  const { data: statsData } = useQuery({
    queryKey: ["user-reactions-stats"],
    queryFn: () => userReactionQueries.getStats(),
  });

  useEffect(() => {
    if (statsData) updateStats({ ...statsData, last_reactions: [] });
  }, [statsData]);

  return (
    <div className="flex items-center gap-2 text-muted-foreground text-sm">
      <Globe className="h-4 w-4 text-primary" />
      <span>
        <span className="font-bold text-foreground tabular-nums">
          {stats.total_reactions.toLocaleString()}
        </span>{" "}
        hellos shared worldwide — live
      </span>
    </div>
  );
}

export default function Landing() {
  return (
    <>
      {/* Hero */}
      <section className="flex w-full flex-col items-center pt-16 md:pt-24 lg:pt-32 pb-12">
        <div className="container px-4 md:px-6 max-w-3xl text-center">
          <div className="inline-flex items-center gap-2 rounded-full bg-primary/10 px-4 py-1.5 text-sm font-medium text-primary mb-6">
            <Hand className="h-4 w-4" />
            Start here — no sign-up needed
          </div>
          <h1 className="text-4xl font-bold tracking-tighter sm:text-5xl md:text-6xl lg:text-7xl mb-4">
            Wave to the world.
            <br />
            <span className="text-muted-foreground font-light">
              Instantly.
            </span>
          </h1>
          <p className="text-lg text-muted-foreground mb-6 max-w-xl mx-auto">
            A playground to explore, experiment, and have fun. Start with a
            single click — then discover what's built inside.
          </p>
          <div className="flex flex-col sm:flex-row gap-3 justify-center mb-6">
            <Button asChild size="lg" className="gap-2 text-base px-8">
              <Link to="/say-hello">
                Say Hello Now
                <ArrowRight className="h-4 w-4" />
              </Link>
            </Button>
            <Button asChild size="lg" variant="outline" className="text-base px-8">
              <Link to="/signup">Create account</Link>
            </Button>
          </div>
          <LiveHelloCount />
        </div>
      </section>

      {/* Discovery Journey */}
      <section className="py-16 md:py-24 bg-muted/30">
        <div className="container mx-auto max-w-5xl px-4 md:px-6">
          <div className="text-center mb-12">
            <h2 className="text-2xl md:text-3xl font-bold tracking-tight mb-3">
              Discover what's here
            </h2>
            <p className="text-muted-foreground text-lg">
              Four things to try, in order. Each one unlocks the next.
            </p>
          </div>

          <div className="grid md:grid-cols-4 gap-4 relative">
            {journeySteps.map((step, i) => (
              <div key={step.step} className="relative flex flex-col">
                <Card className="flex-1 hover:shadow-md hover:border-primary/50 transition-all duration-300">
                  <CardContent className="pt-6 flex flex-col h-full gap-4">
                    <div className="flex items-center gap-3">
                      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-primary text-primary-foreground font-bold text-sm shrink-0">
                        {step.step}
                      </div>
                      <div className="text-muted-foreground">{step.icon}</div>
                    </div>
                    <div className="flex-1">
                      <h3 className="font-semibold text-base mb-1">
                        {step.title}
                      </h3>
                      <p className="text-sm text-muted-foreground leading-relaxed">
                        {step.description}
                      </p>
                    </div>
                    <Button
                      asChild
                      variant={step.step === 1 ? "default" : "outline"}
                      size="sm"
                      className="w-full gap-1.5"
                    >
                      <Link to={step.to}>
                        {step.cta}
                        <ArrowRight className="h-3.5 w-3.5" />
                      </Link>
                    </Button>
                  </CardContent>
                </Card>
                {i < journeySteps.length - 1 && (
                  <div className="hidden md:flex absolute -right-2.5 top-1/2 -translate-y-1/2 z-10 w-5 h-5 items-center justify-center bg-background border border-border rounded-full text-muted-foreground">
                    <ArrowRight className="h-3 w-3" />
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Tech Stack */}
      <section className="py-20 md:py-24">
        <div className="mx-auto max-w-6xl px-6">
          <div className="text-center mb-12">
            <h2 className="text-2xl md:text-3xl font-bold text-foreground mb-3">
              Built With
            </h2>
            <p className="text-muted-foreground">
              Modern full-stack technologies
            </p>
            <div className="flex flex-wrap items-center justify-center gap-8 mt-8">
              {techIconList.map((tech) => (
                <div
                  key={tech.name}
                  className="flex items-center justify-center gap-2 text-muted-foreground hover:text-foreground transition-colors"
                  title={tech.name}
                >
                  {tech.icon}
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* Footer CTA */}
      <section className="border-t border-border py-20 md:py-24">
        <div className="container mx-auto max-w-4xl px-6 text-center">
          <h2 className="text-2xl md:text-3xl font-bold text-foreground mb-4">
            Ready to go deeper?
          </h2>
          <p className="text-lg text-muted-foreground mb-8">
            Create an account to unlock teams, projects, plans, and more.
          </p>
          <div className="flex gap-3 justify-center">
            <Button asChild size="lg">
              <Link to="/signup">Sign up free</Link>
            </Button>
            <Button asChild size="lg" variant="outline">
              <Link to="/features">See all features</Link>
            </Button>
          </div>
        </div>
      </section>
    </>
  );
}

