import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { TeamMemberNotificationData } from "@/schema.types";
import {
  useEventSource,
  useEventSourceListener,
} from "@react-nano/use-event-source";
import { useReducer } from "react";

function TeamNotification() {
  const { user } = useAuthProvider();
  const { teamMember } = useTeam();
  function messageReducer(
    state: TeamMemberNotificationData,
    action: TeamMemberNotificationData
  ) {
    return {
      ...state,
      ...action,
    };
  }
  const [stats, updateStats] = useReducer(messageReducer, {
    notification: {
      title: "",
      body: "",
    },
    data: {
      email: "",
      team_id: "",
      team_member_id: "",
    },
  });

  const [eventSource] = useEventSource(
    "/api/team-members/" +
      teamMember?.id +
      "/notifications/sse?access_token=" +
      user?.tokens.access_token,
    false
  );
  useEventSourceListener(
    eventSource,
    ["new_team_member"],
    (evt) => {
      updateStats(JSON.parse(evt.data)?.user_reaction_stats);
    },
    [updateStats]
  );

  if (!user) return <p>User not found</p>;
  if (!teamMember) return <p>teamMember not found</p>;

  return <div> {stats?.notification.title}</div>;
}

export default TeamNotification;
