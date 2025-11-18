import { Button } from "@/components/ui/button";
import { Feature, features } from "@/pages/landing/features-list";
import { ArrowRight } from "lucide-react";
import { Link } from "react-router";

export default function Features() {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="max-w-(--breakpoint-lg) w-full py-10 px-6">
        <section>
          <h2 className="text-4xl md:text-[2.75rem] md:leading-[1.2] font-semibold tracking-[-0.03em] sm:max-w-xl text-pretty sm:mx-auto sm:text-center">
            <span className="italic font-light">Features</span> of this website
          </h2>
          <p className="mt-2 text-muted-foreground text-lg sm:text-xl sm:text-center">
            This website is not a SaaS, or some kind of service, but a canvas
            for me to freely implement whatever I find interesting or useful.
          </p>
        </section>
        <p className="mt-2 text-muted-foreground text-lg sm:text-xl sm:text-center"></p>
        <div className="mt-8 md:mt-16 w-full mx-auto space-y-20">
          {features.map((feature) => FeatureCard(feature))}
        </div>
      </div>
    </div>
  );
}

function FeatureCard(feature: Feature) {
  return (
    <div
      id={feature.fragment}
      key={feature.fragment}
      className="flex flex-col md:flex-row items-center gap-x-12 gap-y-6 md:even:flex-row-reverse"
    >
      {feature.featureImageComponent ? (
        <div className="w-fit bg-muted rounded-xl border">
          {feature.featureImageComponent}
        </div>
      ) : (
        <div className="w-full aspect-[4/3] bg-muted rounded-xl border border-border/50 basis-1/2">
          <div className="flex items-center justify-center w-full h-full">
            <div className="">{feature.icon}</div>
          </div>
        </div>
      )}
      <div className="basis-1/2 shrink-0">
        <span className="uppercase font-medium text-sm text-muted-foreground">
          {feature.title}
        </span>
        <h4 className="my-3 text-2xl font-semibold tracking-tight">
          {feature.title}
        </h4>
        <p className="text-muted-foreground">
          {feature.mainContent.map((item) => (
            <span>{item}</span>
          ))}
        </p>
        {feature.detailLink && (
          <Button asChild size="lg" className="mt-6 rounded-full gap-3">
            <Link to={feature.detailLink}>
              Learn More <ArrowRight />
            </Link>
          </Button>
        )}
      </div>
    </div>
  );
}
