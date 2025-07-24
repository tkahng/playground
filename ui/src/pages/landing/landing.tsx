import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Brain, CheckCircle, Zap } from "lucide-react";

export default function Landing() {
  return (
    <>
      <section className="flex w-full flex-col items-center py-12 md:py-24 lg:py-32 xl:py-48">
        <div className="container px-4 md:px-6">
          <div className="flex flex-col items-center space-y-4 text-center">
            <div className="space-y-2">
              <h1 className="text-3xl font-bold tracking-tighter sm:text-4xl md:text-5xl lg:text-6xl/none">
                Welcome to the Playground
              </h1>
              <p className="mx-auto max-w-[700px]">
                This is a place of learning and experimentation.
              </p>
            </div>
            <div className="space-x-4">
              <Button>Get Started</Button>
              <Button variant="outline">Learn More</Button>
            </div>
          </div>
        </div>
      </section>
      <section className="flex w-full flex-col items-center py-12 md:py-24 lg:py-32">
        <div className="container px-4 md:px-6">
          <h2 className="mb-12 text-center text-3xl font-bold tracking-tighter sm:text-5xl">
            Key Features
          </h2>
          <div className="grid gap-10 sm:grid-cols-2 md:grid-cols-3">
            <div className="flex flex-col items-center space-y-3 text-center">
              <Zap className="h-10 w-10 text-primary" />
              <h3 className="text-xl font-bold">AI Task Generation</h3>
              <p className="text-gray-500 dark:text-gray-400">
                Describe your project in plain language, and our AI will break
                it down into actionable tasks with deadlines.
              </p>
            </div>
            <div className="flex flex-col items-center space-y-3 text-center">
              <Brain className="h-10 w-10 text-primary" />
              <h3 className="text-xl font-bold">Smart Suggestions</h3>
              <p className="text-gray-500 dark:text-gray-400">
                Get intelligent recommendations for task prioritization,
                resource allocation, and timeline adjustments.
              </p>
            </div>
            <div className="flex flex-col items-center space-y-3 text-center">
              <CheckCircle className="h-10 w-10 text-primary" />
              <h3 className="text-xl font-bold">Automated Workflows</h3>
              <p className="text-gray-500 dark:text-gray-400">
                Set up custom workflows that trigger automatically as tasks
                progress through different stages.
              </p>
            </div>
          </div>
        </div>
      </section>
      <section className="flex w-full flex-col items-center py-12 md:py-24 lg:py-32">
        <div className="container px-4 md:px-6">
          <div className="grid gap-10 px-10 md:gap-16 lg:grid-cols-2">
            <div className="space-y-4">
              <h2 className="text-3xl font-bold tracking-tighter sm:text-5xl">
                Join the AI Revolution
              </h2>
              <p className="text-gray-500 dark:text-gray-400">
                Don't get left behind. Embrace the future of work with NexusAI
                and stay ahead of the competition.
              </p>
            </div>
            <div className="flex flex-col items-start space-y-4">
              <form className="flex w-full flex-col items-start gap-4 sm:flex-row sm:gap-2">
                <Input
                  className="max-w-lg flex-1"
                  placeholder="Enter your email"
                  type="email"
                />
                <Button type="submit">Sign Up</Button>
              </form>
              <p className="text-xs text-gray-500 dark:text-gray-400">
                By signing up, you agree to our Terms of Service and Privacy
                Policy.
              </p>
            </div>
          </div>
        </div>
      </section>
    </>
  );
}
