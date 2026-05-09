import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { TeamMemberNotification } from "@/schema.types";
import {
  useEventSource,
  useEventSourceListener,
} from "@react-nano/use-event-source";
import { useCallback, useEffect, useReducer, useState } from "react";
import { toast } from "sonner";

const SSE_EVENTS = [
  "new_team_member",
  "assigned_to_task",
  "task_due_today",
  "task_completed",
  "task_overdue",
  "task_status_changed",
  "project_status_changed",
];

const INITIAL_NOTIFICATION: TeamMemberNotification = {
  notification: { title: "", body: "" },
  data: { email: "", team_id: "", team_member_id: "" },
};

function messageReducer(
  state: TeamMemberNotification,
  action: TeamMemberNotification,
): TeamMemberNotification {
  toast.info(action.notification.title, {
    description: action.notification.body,
  });
  return { ...state, ...action };
}

// Inner component — only mounted when a valid ticket URL is ready.
function SSEListener({ url }: { url: string }) {
  const [, dispatch] = useReducer(messageReducer, INITIAL_NOTIFICATION);
  const [eventSource] = useEventSource(url, false);
  useEventSourceListener(
    eventSource,
    SSE_EVENTS,
    (evt) => {
      dispatch(JSON.parse(evt.data) as TeamMemberNotification);
    },
    [dispatch],
  );
  return null;
}

function TeamNotification() {
  const { user } = useAuthProvider();
  const { teamMember } = useTeam();
  const [ticket, setTicket] = useState<string | null>(null);

  const fetchTicket = useCallback(() => {
    if (!user?.tokens.access_token || !teamMember?.id) return;
    let active = true;

    fetch(`/api/team-members/${teamMember.id}/sse/ticket`, {
      method: "POST",
      headers: { Authorization: `Bearer ${user.tokens.access_token}` },
    })
      .then((r) => r.json())
      .then((d: { ticket?: string }) => {
        if (active && d?.ticket) setTicket(d.ticket);
      })
      .catch(() => {});

    return () => {
      active = false;
    };
  }, [user?.tokens.access_token, teamMember?.id]);

  useEffect(() => {
    return fetchTicket();
  }, [fetchTicket]);

  if (!ticket || !teamMember?.id) return null;

  return (
    <SSEListener
      url={`/api/team-members/${teamMember.id}/sse?ticket=${ticket}`}
    />
  );
}

export default TeamNotification;
