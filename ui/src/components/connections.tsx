import { Button } from "@/components/ui/button";
import { Icons } from "@/components/ui/icons";
import { getAuthUrl } from "@/lib/queries";
import { z } from "zod";
export const GITHUB_PROVIDER_NAME = "github";
export const GOOGLE_PROVIDER_NAME = "google";
// to add another provider, set their name here and add it to the providerNames below

export const providerNames = [
  GITHUB_PROVIDER_NAME,
  GOOGLE_PROVIDER_NAME,
] as const;
export const ProviderNameSchema = z.enum(providerNames);
export type ProviderName = z.infer<typeof ProviderNameSchema>;

export const providerLabels: Record<ProviderName, string> = {
  [GITHUB_PROVIDER_NAME]: "GitHub",
  [GOOGLE_PROVIDER_NAME]: "Google",
} as const;

export const providerIcons: Record<ProviderName, React.ReactNode> = {
  [GITHUB_PROVIDER_NAME]: (
    <Icons.gitHub title="github-logo" className="h-4 w-4" />
  ),
  [GOOGLE_PROVIDER_NAME]: (
    <Icons.google title="google-logo" className="h-4 w-4" />
  ),
} as const;

export function ProviderConnectionForm({
  redirectTo,
  providerName,
}: {
  redirectTo?: string | null;
  type: "Connect" | "Login" | "Signup";
  providerName: ProviderName;
}) {
  const callbackUrl = new URL(
    window.location.origin + "/auth/callback"
  ).toString();
  const onSubmit = async () => {
    const url = await getAuthUrl({
      provider: providerName,
      redirect: redirectTo || callbackUrl,
    });
    window.location.href = url;
  };
  return (
    // Using a button to trigger the authentication flow
    <Button
      type="submit"
      variant="outline"
      size="icon"
      onClick={() => onSubmit()}
    >
      {providerIcons[providerName]}
    </Button>
  );
}
