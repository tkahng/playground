import { Button } from "@/components/ui/button";
import { Feature, features } from "@/pages/landing/features-list";
import { Link } from "@tanstack/react-router";
import { ArrowRight } from "lucide-react";

export default function Features() {
  return (
    <div className="min-h-screen">
      <div className="max-w-(--breakpoint-lg) w-full py-10 px-6 mx-auto">

        {/* Header */}
        <section className="sm:text-center mb-10">
          <h2 className="text-4xl md:text-[2.75rem] md:leading-[1.2] font-semibold tracking-[-0.03em] sm:max-w-xl text-pretty sm:mx-auto">
            <span className="italic font-light">Everything</span> in this playground
          </h2>
          <p className="mt-3 text-muted-foreground text-lg sm:text-xl max-w-xl sm:mx-auto">
            Ordered by how to discover them. Start with Say Hello — no account needed.
          </p>
        </section>

        {/* Quick-jump nav */}
        <nav className="flex flex-wrap gap-2 sm:justify-center mb-16">
          {features.map((f, i) => (
            <a
              key={f.fragment}
              href={`#${f.fragment}`}
              onClick={(e) => {
                e.preventDefault();
                document
                  .getElementById(f.fragment)
                  ?.scrollIntoView({ behavior: "smooth", block: "center" });
              }}
              className="inline-flex items-center gap-1.5 rounded-full border border-border px-3 py-1 text-sm text-muted-foreground hover:border-primary hover:text-foreground transition-colors"
            >
              <span className="text-xs font-mono text-muted-foreground/60">
                {String(i + 1).padStart(2, "0")}
              </span>
              {f.title}
            </a>
          ))}
        </nav>

        {/* Feature list */}
        <div className="w-full mx-auto space-y-24">
          {features.map((feature, index) => (
            <FeatureCard
              key={feature.fragment}
              feature={feature}
              step={index + 1}
              total={features.length}
              nextFeature={features[index + 1]}
            />
          ))}
        </div>
      </div>
    </div>
  );
}

function FeatureCard({
  feature,
  step,
  total,
  nextFeature,
}: {
  feature: Feature;
  step: number;
  total: number;
  nextFeature?: Feature;
}) {
  return (
    <div
      id={feature.fragment}
      className="flex flex-col md:flex-row items-center gap-8 md:gap-16 md:even:flex-row-reverse scroll-mt-8"
    >
      {/* Image */}
      <div className="w-full md:basis-1/2 md:shrink-0 min-w-0">
        <div className="rounded-2xl border border-border/50 bg-muted overflow-hidden p-6 flex items-center justify-center">
          {feature.featureImageComponent ? (
            <div className="w-full flex items-center justify-center">
              {feature.featureImageComponent}
            </div>
          ) : (
            <div className="aspect-[4/3] w-full flex items-center justify-center">
              {feature.icon}
            </div>
          )}
        </div>
      </div>

      {/* Text */}
      <div className="w-full md:basis-1/2 md:shrink-0 min-w-0 flex flex-col gap-4">
        {/* Step + badge row */}
        <div className="flex flex-wrap items-center gap-2">
          <span className="font-mono text-xs text-muted-foreground/60 bg-muted px-2 py-0.5 rounded">
            {String(step).padStart(2, "0")} / {String(total).padStart(2, "0")}
          </span>
          {feature.badge && (
            <span className="rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-medium text-primary">
              {feature.badge}
            </span>
          )}
        </div>

        <div>
          <h4 className="text-2xl font-semibold tracking-tight mb-2">
            {feature.title}
          </h4>
          <p className="text-muted-foreground leading-relaxed">
            {feature.mainContent.map((item) => (
              <span key={item}>{item}</span>
            ))}
          </p>
        </div>

        {/* CTAs */}
        <div className="flex flex-wrap items-center gap-3 mt-2">
          {feature.detailLink && (
            <Button asChild size="lg" className="rounded-full gap-2">
              <Link to={feature.detailLink}>
                {feature.detailLinkText || "Try it"}
                <ArrowRight className="h-4 w-4" />
              </Link>
            </Button>
          )}
          {nextFeature && (
            <a
              href={`#${nextFeature.fragment}`}
              onClick={(e) => {
                e.preventDefault();
                document
                  .getElementById(nextFeature.fragment)
                  ?.scrollIntoView({ behavior: "smooth", block: "center" });
              }}
              className="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1 transition-colors"
            >
              Next: {nextFeature.title}
              <ArrowRight className="h-3.5 w-3.5" />
            </a>
          )}
        </div>
      </div>
    </div>
  );
}
