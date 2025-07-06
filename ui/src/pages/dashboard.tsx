import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { Navigate } from "react-router";

export default function Dashboard() {
  const { user } = useAuthProvider();
  const { team, teamMember } = useTeam();
  if (!user) {
    return <Navigate to="/signin" />;
  }
  if (!team || !teamMember) {
    return <Navigate to="/teams" />;
  }
  if (teamMember.user_id !== user.user.id) {
    return <Navigate to={`/teams`} />;
  }
  return <Navigate to={`/teams/${team.slug}/dashboard`} />;
}
