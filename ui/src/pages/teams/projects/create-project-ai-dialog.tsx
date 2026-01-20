import * as React from "react";
import {
  Sparkles,
  Loader2,
  CheckCircle2,
  ArrowRight,
  Lightbulb,
  AlertCircle,
  RotateCcw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import { useNavigate } from "react-router";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { taskProjectCreateWithAi } from "@/lib/task-queries";
import { Project } from "@/schema.types";
import { useMutation } from "@tanstack/react-query";
import { ApiError } from "@/lib/error";

type DialogStep = "input" | "generating" | "success" | "error";

const examplePrompts = [
  "Launch a new mobile app for fitness tracking",
  "Migrate our database to a new cloud provider",
  "Plan a company-wide hackathon event",
];

export function CreateProjectAiDialog() {
  const [open, setOpen] = React.useState(false);
  const [step, setStep] = React.useState<DialogStep>("input");
  const [prompt, setPrompt] = React.useState("");
  const [generatedProject, setGeneratedProject] =
    React.useState<Project | null>(null);
  const [generatingPhase, setGeneratingPhase] = React.useState(0);
  const [errorMessage, setErrorMessage] = React.useState("");
  const { user } = useAuthProvider();

  const { team: currentTeam } = useTeam();
  const navigate = useNavigate();

  const generatingPhases = [
    "Analyzing your objective...",
    "Generating project structure...",
    "Creating tasks and milestones...",
    "Finalizing your project...",
  ];

  const resetDialog = () => {
    setStep("input");
    setPrompt("");
    setGeneratedProject(null);
    setGeneratingPhase(0);
    setErrorMessage("");
  };

  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen);
    if (!newOpen) {
      // Reset after close animation
      setTimeout(resetDialog, 200);
    }
  };
  const mutation = useMutation({
    mutationFn: async ({ input }: { input: string }) => {
      if (!user?.tokens.access_token)
        throw new ApiError("Missing access token or role ID");

      if (!currentTeam) throw new Error("Missing team");
      const res = await taskProjectCreateWithAi(
        user.tokens.access_token,
        currentTeam.id,
        {
          input,
        },
      );

      return res;
    },
    onMutate: () => {
      setStep("generating");
      setGeneratingPhase(0); // Reset when not loading
      setErrorMessage("");
    },
    onSuccess: async (project) => {
      setGeneratedProject(project);
      setStep("success");
      setTimeout(() => {
        navigate(`/teams/${currentTeam?.slug}/projects/${project?.id}`);
        handleOpenChange(false);
      }, 1800);
    },
    onError: (error: ApiError) => {
      const msg = `${error.message} - ${error.detail || "The AI service is temporarily unavailable. Please try again."}`;
      setErrorMessage(msg);
      setStep("error");
    },
  });
  React.useEffect(() => {
    if (!mutation.isPending) {
      setGeneratingPhase(0); // Reset when not loading
      return;
    }

    const interval = setInterval(() => {
      setGeneratingPhase((prev) => {
        // Stop at the last phase, don't loop back
        if (prev < generatingPhases.length - 1) {
          return prev + 1;
        }
        return prev;
      });
    }, 800); // Change phase every 2 seconds

    return () => clearInterval(interval);
  }, [mutation.isPending, generatingPhases.length]);

  const handleSubmit = () => {
    mutation.mutate({ input: prompt });
  };

  const handleRetry = () => {
    handleSubmit();
  };

  const handleBackToInput = () => {
    setStep("input");
    setErrorMessage("");
    setGeneratingPhase(0);
  };
  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button className="gap-2">
          <Sparkles className="size-4" />
          Create Project with AI
        </Button>
      </DialogTrigger>
      <DialogContent
        className={cn(
          "overflow-hidden transition-all duration-300",
          step === "generating" && "sm:max-w-md",
          step === "success" && "sm:max-w-md",
          step === "error" && "sm:max-w-md",
        )}
        showCloseButton={step === "input" || step === "error"}
      >
        {/* Input Step */}
        <div
          className={cn(
            "grid transition-all duration-300",
            step === "input"
              ? "grid-rows-[1fr] opacity-100"
              : "grid-rows-[0fr] opacity-0 absolute",
          )}
        >
          <div className="overflow-hidden">
            <DialogHeader>
              <DialogTitle className="flex items-center gap-2">
                <span className="flex size-8 items-center justify-center rounded-lg bg-primary/10">
                  <Sparkles className="size-4 text-primary" />
                </span>
                Create Project with AI
              </DialogTitle>
              <DialogDescription>
                Describe what you want to accomplish, and AI will generate a
                project with tasks for you.
              </DialogDescription>
            </DialogHeader>

            <div className="mt-6 space-y-4">
              <Textarea
                placeholder="e.g., Build a customer onboarding flow that reduces churn and improves user activation..."
                className="min-h-30 resize-none"
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
              />

              <div className="space-y-2">
                <p className="flex items-center gap-1.5 text-xs text-muted-foreground">
                  <Lightbulb className="size-3" />
                  Try one of these examples:
                </p>
                <div className="flex flex-wrap gap-2">
                  {examplePrompts.map((example) => (
                    <button
                      key={example}
                      type="button"
                      onClick={() => setPrompt(example)}
                      className="rounded-full border bg-secondary/50 px-3 py-1 text-xs text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                    >
                      {example}
                    </button>
                  ))}
                </div>
              </div>

              <Button
                onClick={handleSubmit}
                disabled={!prompt.trim()}
                className="w-full gap-2"
              >
                Generate Project
                <ArrowRight className="size-4" />
              </Button>
            </div>
          </div>
        </div>

        {/* Generating Step */}
        <div
          className={cn(
            "grid transition-all duration-300",
            step === "generating"
              ? "grid-rows-[1fr] opacity-100"
              : "grid-rows-[0fr] opacity-0 absolute",
          )}
        >
          <div className="overflow-hidden">
            <div className="flex flex-col items-center py-8 text-center">
              <div className="relative mb-6">
                <div className="absolute inset-0 animate-ping rounded-full bg-primary/20" />
                <div className="relative flex size-16 items-center justify-center rounded-full bg-primary/10">
                  <Loader2 className="size-8 animate-spin text-primary" />
                </div>
              </div>

              <h3 className="mb-2 text-lg font-semibold">
                Creating Your Project
              </h3>

              <div className="mb-6 h-5">
                <p
                  key={generatingPhase}
                  className="animate-in fade-in slide-in-from-bottom-2 text-sm text-muted-foreground duration-300"
                >
                  {generatingPhases[generatingPhase]}
                </p>
              </div>

              {/* Progress indicators */}
              <div className="flex gap-1.5">
                {generatingPhases.map((_, index) => (
                  <div
                    key={index}
                    className={cn(
                      "h-1.5 w-8 rounded-full transition-all duration-300",
                      index <= generatingPhase ? "bg-primary" : "bg-muted",
                    )}
                  />
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Success Step */}
        <div
          className={cn(
            "grid transition-all duration-300",
            step === "success"
              ? "grid-rows-[1fr] opacity-100"
              : "grid-rows-[0fr] opacity-0 absolute",
          )}
        >
          <div className="overflow-hidden">
            <div className="flex flex-col items-center py-8 text-center">
              <div className="relative mb-6">
                <div className="absolute inset-0 animate-ping rounded-full bg-green-500/20 animation-duration-[1s]" />
                <div className="relative flex size-16 items-center justify-center rounded-full bg-green-500/10">
                  <CheckCircle2 className="size-8 text-green-500" />
                </div>
              </div>

              <h3 className="mb-2 text-lg font-semibold">Project Created!</h3>
              <p className="mb-4 text-sm text-muted-foreground">
                {generatedProject?.name}
              </p>

              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <Loader2 className="size-3 animate-spin" />
                Redirecting to your project...
              </div>
            </div>
          </div>
        </div>

        {/* Error Step */}
        <div
          className={cn(
            "grid transition-all duration-300",
            step === "error"
              ? "grid-rows-[1fr] opacity-100"
              : "grid-rows-[0fr] opacity-0 absolute",
          )}
        >
          <div className="overflow-hidden">
            <div className="flex flex-col items-center py-8 text-center">
              <div className="relative mb-6">
                <div className="relative flex size-16 items-center justify-center rounded-full bg-destructive/10">
                  <AlertCircle className="size-8 text-destructive" />
                </div>
              </div>

              <h3 className="mb-2 text-lg font-semibold">
                Something went wrong
              </h3>
              <p className="mb-6 max-w-70 text-sm text-muted-foreground">
                {errorMessage}
              </p>

              <div className="flex w-full flex-col gap-2 sm:flex-row sm:justify-center">
                <Button onClick={handleRetry} className="gap-2">
                  <RotateCcw className="size-4" />
                  Try Again
                </Button>
                <Button
                  variant="outline"
                  onClick={handleBackToInput}
                  className="bg-transparent"
                >
                  Edit Prompt
                </Button>
              </div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
