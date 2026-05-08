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
    <section
      id={feature.fragment}
      className="scroll-mt-20 flex flex-col sm:flex-row gap-8 items-start"
    >
      <div className="sm:sticky sm:top-24 sm:w-48 shrink-0">
        <div className="text-xs font-mono text-muted-foreground/60 mb-1">
          {String(step).padStart(2, "0")} / {String(total).padStart(2, "0")}
        </div>
        <div className="text-sm font-semibold">{feature.title}</div>
        {nextFeature && (
          <a
            href={`#${nextFeature.fragment}`}
            onClick={(e) => {
              e.preventDefault();
              document
                .getElementById(nextFeature.fragment)
                ?.scrollIntoView({ behavior: "smooth", block: "center" });
            }}
            className="mt-3 flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors"
          >
            Next <ArrowRight className="h-3 w-3" />
          </a>
        )}
      </div>

      <div className="flex-1 space-y-3">
        <p className="text-muted-foreground">{feature.description}</p>
        {feature.to && (
          <Link
            to={feature.to}
            className="inline-flex items-center gap-1 text-sm font-medium hover:underline"
          >
            Try it <ArrowRight className="h-4 w-4" />
          </Link>
        )}
      </div>
    </section>
  );
}
