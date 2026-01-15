import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { features } from "@/pages/landing/features-list";
import { ArrowRight } from "lucide-react";
import { JSX } from "react";
import { Link } from "react-router";
import { HashLink } from "react-router-hash-link";
import {
  SiGo,
  SiPostgresql,
  SiReact,
  SiTypescript,
  SiTailwindcss,
  SiDocker,
  SiGithub,
  SiGithubactions,
  SiTailscale,
} from "@icons-pack/react-simple-icons";

export const techIconList: { icon: JSX.Element; name: string }[] = [
  {
    icon: <SiGo />,
    name: "Go",
  },
  {
    icon: <SiPostgresql />,
    name: "Postgres",
  },
  {
    icon: <SiReact />,
    name: "React",
  },
  {
    icon: <SiTypescript />,
    name: "Typescript",
  },
  {
    icon: <SiTailwindcss />,
    name: "Tailwind",
  },
  {
    icon: <SiDocker />,
    name: "Docker",
  },
  {
    icon: <SiGithub />,
    name: "Github",
  },
  {
    icon: <SiGithubactions />,
    name: "Github Actions",
  },
  {
    icon: <SiTailscale />,
    name: "Tailscale",
  },
];
export default function Landing() {
  return (
    <>
      <section className="flex w-full flex-col items-center pt-12 md:pt-16 lg:pt-24 xl:pt-32">
        <div className="container px-4 md:px-6">
          <div className="flex flex-col items-center space-y-4 text-center">
            <div className="space-y-2">
              <h1 className="text-3xl font-bold tracking-tighter sm:text-2xl md:text-3xl lg:text-4xl/none">
                Welcome to my Playground
              </h1>
              <h2 className="mx-auto max-w-[700px] text-2xl font-bold text-shadow-muted-foreground">
                Lets build cool toys
              </h2>
              <p className="mx-auto max-w-[700px] text-lg text-muted-foreground">
                This is where I experiment, learn, but most importantly have fun
                implementing cool features.
              </p>
            </div>
          </div>
        </div>
      </section>

      <section className="mx-auto max-w-4xl px-6 pt-12 md:pt-16 lg:pt-24 xl:pt-32">
        <h2 className="mb-12 text-center text-3xl font-bold tracking-tighter sm:text-2xl md:text-3xl lg:text-4xl/none">
          What you can do here:
        </h2>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {features.map(
            (
              {
                icon,
                title,
                landingLink,
                landingLinkText,
                shortContent: content,
              },
              index,
            ) => (
              <LandingFeatureCard
                featureLink={landingLink}
                landingLinkText={landingLinkText}
                index={index}
                key={index}
                icon={icon}
                title={title}
                content={content}
              />
            ),
          )}
        </div>
      </section>
      {/* Tech Stack Section */}
      <section className="py-20 md:py-24">
        <div className="mx-auto max-w-6xl px-6">
          <div className="text-center mb-12">
            <h2 className="text-3xl md:text-4xl font-bold text-foreground mb-4">
              Built With
            </h2>
            <p className="text-lg text-muted-foreground mb-4">
              Modern technologies for a robust full-stack experience
            </p>
            <div className="flex flex-wrap items-center justify-center gap-8">
              {techIconList.map((tech) => (
                <div
                  key={tech.name}
                  className="flex items-center justify-center gap-2 m-4"
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
            Ready to explore?
          </h2>
          <p className="text-lg text-muted-foreground mb-8">
            Dive into the features and see what's possible
          </p>
          <Button asChild size="lg">
            <Link to="/features">Get Started</Link>
          </Button>
        </div>
      </section>
    </>
  );
}

function LandingFeatureCard({
  index,
  icon,
  title,
  content,
  className,
  featureLink,
  landingLinkText,
}: {
  index: number;
  icon: JSX.Element;
  title: string;
  content: string[];
  className?: string;
  featureLink: string;
  landingLinkText?: string;
}) {
  return (
    <Card
      key={index}
      className={cn(
        "group hover:shadow-md transition-all duration-300 hover:border-primary/50",
        className,
      )}
    >
      <CardHeader className="flex flex-col items-center">
        {icon}
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col gap-10">
        <div>
          {content.map((item) => {
            return <p key={item}>{item}</p>;
          })}
        </div>
        {featureLink && (
          <Button asChild variant={"ghost"}>
            <HashLink
              to={featureLink}
              scroll={(el) =>
                el.scrollIntoView({ behavior: "smooth", block: "center" })
              }
              className="gap-2 text-muted-foreground hover:bg-muted"
            >
              {landingLinkText || "Learn More"}
              <ArrowRight className="mr-2 h-4 w-4" />
            </HashLink>
          </Button>
        )}
      </CardContent>
    </Card>
  );
}
