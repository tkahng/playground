import { Card } from "@/components/ui/card";
import { Github, ExternalLink } from "lucide-react";

export default function LandingAboutPage() {
  return (
    <div className="container mx-auto max-w-4xl px-6 py-16 md:py-24">
      {/* Header */}
      <div className="mb-16">
        <h1 className="text-4xl md:text-5xl font-bold text-foreground mb-4 text-balance">
          About Playground
        </h1>
        <p className="text-lg text-muted-foreground leading-relaxed">
          A personal space for exploration and learning
        </p>
      </div>

      {/* What is Playground Section */}
      <section className="mb-20">
        <h2 className="text-2xl font-semibold text-foreground mb-4">
          What is this place?
        </h2>
        <p className="text-base leading-relaxed text-muted-foreground">
          Playground is my personal experimentation space—a digital workshop
          where I implement features that spark my curiosity. This is where I
          learn about the web and computing in general, testing ideas and
          building things that interest me. Each project here represents a
          learning journey, exploring new technologies, patterns, and approaches
          to software development.
        </p>
      </section>
      {/* Links Section */}
      <section className="mb-20">
        <h2 className="text-2xl font-semibold text-foreground mb-6">
          Source Code
        </h2>
        <div className="flex flex-col sm:flex-row gap-4">
          <a
            href="https://github.com/tkahng/playground"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 px-6 py-3 rounded-lg border border-border bg-card hover:bg-accent transition-colors group"
          >
            <Github className="w-5 h-5" />
            <span className="font-medium">GitHub Repository</span>
            <ExternalLink className="w-4 h-4 ml-auto opacity-50 group-hover:opacity-100 transition-opacity" />
          </a>
          <a
            href="https://forgejo.k2dv.io/"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 px-6 py-3 rounded-lg border border-border bg-card hover:bg-accent transition-colors group"
          >
            <svg
              className="w-5 h-5"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <circle cx="12" cy="12" r="10" />
              <circle cx="12" cy="12" r="3" />
            </svg>
            <span className="font-medium">Forgejo (Self-hosted)</span>
            <ExternalLink className="w-4 h-4 ml-auto opacity-50 group-hover:opacity-100 transition-opacity" />
          </a>
        </div>
      </section>
      {/* Tech Stack Section */}
      <section className="mb-20">
        <h2 className="text-2xl font-semibold text-foreground mb-6">
          Tech Stack
        </h2>
        <div className="grid gap-6 md:grid-cols-2">
          <Card className="p-6 border-border bg-card">
            <h3 className="text-sm font-medium text-muted-foreground mb-3">
              Backend
            </h3>
            <ul className="space-y-2 text-foreground">
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-primary" />
                Language: Golang
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-primary" />
                Protocols: REST API, SSE
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-primary" />
                Database: PostgreSQL, PostGIS
              </li>
            </ul>
          </Card>

          <Card className="p-6 border-border bg-card">
            <h3 className="text-sm font-medium text-muted-foreground mb-3">
              Frontend
            </h3>
            <ul className="space-y-2 text-foreground">
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-primary" />
                Framework: React TypeScript
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-primary" />
                UI: Tailwind, Shadcn, Motion
              </li>
            </ul>
          </Card>

          <Card className="p-6 border-border bg-card">
            <h3 className="text-sm font-medium text-muted-foreground mb-3">
              Infrastructure
            </h3>
            <ul className="space-y-2 text-foreground">
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-primary" />
                Deployment: Docker in VPS
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-primary" />
                VPN: Tailscale
              </li>
            </ul>
          </Card>

          <Card className="p-6 border-border bg-card">
            <h3 className="text-sm font-medium text-muted-foreground mb-3">
              DevOps
            </h3>
            <ul className="space-y-2 text-foreground">
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-primary" />
                Source: GitHub, Forgejo (self-hosted)
              </li>
              <li className="flex items-center gap-2">
                <span className="w-1.5 h-1.5 rounded-full bg-primary" />
                CI/CD: GitHub Actions
              </li>
            </ul>
          </Card>
        </div>
      </section>

      {/* Credits Section */}
      <section>
        <h2 className="text-2xl font-semibold text-foreground mb-4">Credits</h2>
        <p className="text-sm text-muted-foreground leading-relaxed">
          Icons and graphics used in Playground are sourced and modified from{" "}
          <a
            href="https://thenounproject.com/browse/icons/term/shapes/"
            target="_blank"
            rel="noopener noreferrer"
            className="text-foreground underline underline-offset-4 hover:text-primary transition-colors"
          >
            The Noun Project
          </a>
          .
        </p>
      </section>
    </div>
  );
}
