import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { features } from "@/pages/landing/features-list";
import { ArrowRight } from "lucide-react";
import { JSX } from "react";
import { HashLink } from "react-router-hash-link";

export default function Landing() {
  return (
    <>
      <section className="flex w-full flex-col items-center py-12 md:py-16 lg:py-24 xl:py-32">
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

      <section className="flex w-full flex-col items-center">
        <div className="container px-4 md:px-6">
          <h2 className="mb-12 text-center text-3xl font-bold tracking-tighter sm:text-2xl md:text-3xl lg:text-4xl/none">
            What you can do here:
          </h2>
          <div className="@container flex flex-wrap justify-center gap-10">
            {features.map(
              (
                {
                  icon,
                  title,
                  landingLink,
                  landingLinkText,
                  shortContent: content,
                },
                index
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
              )
            )}
          </div>
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
      className={cn(" w-[30cqw] justify-center grow py-6", className)}
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
