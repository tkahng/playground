import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { useTeam } from "@/hooks/use-team";
import { BarChart, Cpu, LineChart, PieChart, Users, Zap } from "lucide-react";

export default function Dashboard() {
  // const { user } = useAuthProvider();
  // if (!user) {
  //   return <Navigate to="/signin" />;
  // }
  // if (!team || !teamMember) {
  //   return <Navigate to="/teams" />;
  // }
  // if (teamMember.user_id !== user.user.id) {
  //   return <Navigate to={`/teams`} />;
  // }
  // return <Navigate to={`/teams/${team.slug}/dashboard`} />;
  // return <Navigate to={`/account/dashboard`} />;
  return (
    <div className="container px-4 md:px-6">
      <h1 className="text-3xl font-bold mb-6">Dashboard</h1>
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <TeamCard />
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Models</CardTitle>
            <Cpu className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">12</div>
            <p className="text-xs text-muted-foreground">
              3 new models this week
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Avg. Response Time
            </CardTitle>
            <BarChart className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">237ms</div>
            <p className="text-xs text-muted-foreground">
              14ms faster than last week
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">API Usage</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">85%</div>
            <Progress value={85} className="mt-2" />
          </CardContent>
        </Card>
      </div>
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-7 mt-6">
        {TeamCard2()}
        <Card className="col-span-3">
          <CardHeader>
            <CardTitle>Top Models</CardTitle>
            <CardDescription>
              Your most used AI models this month
            </CardDescription>
          </CardHeader>
          <CardContent>
            <PieChart className="h-[200px] w-full" />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
function TeamCard2() {
  return (
    <Card className="col-span-4">
      <CardHeader>
        <CardTitle>Last Team</CardTitle>
      </CardHeader>
      <CardContent className="pl-2">
        <LineChart className="h-[200px] w-full" />
      </CardContent>
    </Card>
  );
}

function TeamCard() {
  const { team, teamMember } = useTeam();
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">Last Team</CardTitle>
        <Zap className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      {team && teamMember ? (
        <CardContent>
          <div className="text-2xl font-bold">{team?.name}</div>
          <p className="text-xs text-muted-foreground">{teamMember?.role}</p>
        </CardContent>
      ) : (
        <CardContent>
          <div className="text-2xl font-bold">No Team</div>
          <p className="text-xs text-muted-foreground">You are not in a team</p>
        </CardContent>
      )}
    </Card>
  );
}
