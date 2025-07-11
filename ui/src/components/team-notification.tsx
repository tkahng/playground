import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useSSE } from "@/hooks/use-sse";
import { useTeam } from "@/hooks/use-team";
import { GetError } from "@/lib/get-error";
import { TeamMemberNotificationData } from "@/schema.types";

function TeamNotification() {
  const { user } = useAuthProvider();
  const { teamMember } = useTeam();
  const { data, error } = useSSE<TeamMemberNotificationData>(
    "/team-members/" + teamMember?.id + "/notifications/sse"
  );
  if (!user) {
    return <div> </div>;
  }
  if (!user) return <p>User not found</p>;
  if (!teamMember) return <p>User not found</p>;
  if (error) {
    const err = GetError(error);
    return <p>Error: {err?.detail}</p>;
  }
  if (!data) return <p>User not found</p>;
  return <div> {data.notification.title}</div>;
}

export default TeamNotification;
