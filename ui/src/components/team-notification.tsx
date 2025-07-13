import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useSSE } from "@/hooks/use-sse";
import { useTeam } from "@/hooks/use-team";
import { GetError } from "@/lib/get-error";
import { TeamMemberNotificationData } from "@/schema.types";

function TeamNotification() {
  const { user } = useAuthProvider();
  const { teamMember } = useTeam();
  const { data, error } = useSSE<TeamMemberNotificationData>(
    "/api/team-members/" +
      teamMember?.id +
      "/notifications/sse?access_token=" +
      user?.tokens.access_token
  );

  if (!user) return <p>User not found</p>;
  if (!teamMember) return <p>teamMember not found</p>;
  if (error) {
    const err = GetError(error);
    return <p>Error: {err?.detail}</p>;
  }
  return <div> {data?.notification.title}</div>;
}

export default TeamNotification;
