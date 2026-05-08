import { useTeam } from "@/hooks/use-team";
import { Navigate } from "@tanstack/react-router";

export default function TeamSettingsRedirect() {
  const { team } = useTeam();
  if (!team) {
    return <Navigate to="/teams" />;
  }
  return <Navigate to="/teams/$teamSlug/settings/billing" params={{ teamSlug: team.slug }} />;
}
