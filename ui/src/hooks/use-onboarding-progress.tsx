import { useLocalStorage } from "./use-local-storage";
import { useAuthProvider } from "./use-auth-provider";

export type OnboardingProgress = {
  saidHello: boolean;
  hasProject: boolean;
  visitedPricing: boolean;
  visitedRps: boolean;
  dismissed: boolean;
};

const DEFAULT_PROGRESS: OnboardingProgress = {
  saidHello: false,
  hasProject: false,
  visitedPricing: false,
  visitedRps: false,
  dismissed: false,
};

export function useOnboardingProgress() {
  const { user } = useAuthProvider();
  // Key scoped to user so progress is per-account on the same browser
  const key = `onboarding_${user?.user.id ?? "anon"}`;
  const [progress, setProgress] = useLocalStorage<OnboardingProgress>(
    key,
    DEFAULT_PROGRESS,
  );

  const markStep = (step: keyof Omit<OnboardingProgress, "dismissed">) => {
    if (progress[step]) return;
    setProgress({ ...progress, [step]: true });
  };

  const dismiss = () => {
    setProgress({ ...progress, dismissed: true });
  };

  return { progress, markStep, dismiss };
}
