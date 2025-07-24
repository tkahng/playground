import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { TeamMemberNotification } from "@/schema.types";
import {
  useEventSource,
  useEventSourceListener,
} from "@react-nano/use-event-source";
import { useReducer } from "react";
import { toast } from "sonner";

function TeamNotification() {
  const { user } = useAuthProvider();
  const { teamMember } = useTeam();
  function messageReducer(
    state: TeamMemberNotification,
    action: TeamMemberNotification
  ) {
    toast.info(action.notification.title, {
      description: action.notification.body,
    });
    return {
      ...state,
      ...action,
    };
  }
  const [notification, updateLatestNotification] = useReducer(messageReducer, {
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
      "/sse?access_token=" +
      user?.tokens.access_token,
    false
  );
  useEventSourceListener(
    eventSource,
    ["new_team_member", "assigned_to_task", "task_due_today"],
    (evt) => {
      const noti: TeamMemberNotification = JSON.parse(evt.data);
      updateLatestNotification(noti);
    },
    [updateLatestNotification]
  );
  // useEffect(() => {
  //   if (notification) {
  //     toast.info(notification.notification.title, {
  //       description: notification.notification.body,
  //     });
  //   }
  // }, [notification]);

  return <div> {notification?.notification.title}</div>;
}

export default TeamNotification;
