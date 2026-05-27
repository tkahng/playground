import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { TeamMemberNotification } from "@/schema.types";
import {
  useEventSource,
  useEventSourceListener,
} from "@react-nano/use-event-source";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

const SSE_EVENTS = [
  "new_team_member",
  "assigned_to_task",
  "task_due_today",
  "task_completed",
  "task_overdue",
  "task_status_changed",
  "project_status_changed",
  "task_comment_created",
  "task_comment_mention",
];

// Ticket TTL is 60 s; refresh every 55 s to ensure the URL stays valid on reconnect.
const TICKET_REFRESH_MS = 55_000;

// Inner component — only mounted when a valid ticket URL is ready.
function SSEListener({ url }: { url: string }) {
  const [eventSource] = useEventSource(url, false);
  useEventSourceListener(
    eventSource,
    SSE_EVENTS,
    (evt) => {
      const n = JSON.parse(evt.data) as TeamMemberNotification;
      toast.info(n.notification.title, { description: n.notification.body });
    },
    [],
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
    const cleanup = fetchTicket();
    const interval = setInterval(fetchTicket, TICKET_REFRESH_MS);
    return () => {
      cleanup?.();
      clearInterval(interval);
    };
  }, [fetchTicket]);

  if (!ticket || !teamMember?.id) return null;

  return (
    <SSEListener
      url={`/api/team-members/${teamMember.id}/sse?ticket=${ticket}`}
    />
  );
}

export default TeamNotification;
