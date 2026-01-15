import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { useTeam } from "@/hooks/use-team";
import { UserPlus } from "lucide-react";
import { Link } from "react-router";

export default function TeamDashboard() {
  const { team } = useTeam();

  if (!team) {
    return <div>No team selected</div>;
  }
  return (
    <div className="mx-auto px-8 py-8 justify-start items-stretch flex-1 max-w-[1200px]">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            {team.name} Dashboard
          </h1>
          <p className="text-muted-foreground">
            Manage your team's AI usage and collaboration
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button>
            <UserPlus className="mr-2 h-4 w-4" />
            Invite Member
          </Button>
        </div>
      </div>
      <Card>
        <CardContent className="m-8">
          <p>This is your Team's dashboard!</p>
          <br />
          <p>
            While we work on polishing this page, here are some things to try:
          </p>
          <ul className="list-disc mx-6">
            <li>
              Create a project with AI. Go to the{" "}
              <Link
                to={`/teams/${team.slug}/projects`}
                className="text-primary hover:text-accent-foreground underline hover:no-underline"
              >
                Team Projects page
              </Link>
              , click on the Create Project with AI button, and describe your
              project. It will generate a list of tasks for it!
            </li>
            <li>
              Invite a team member! Send invitations to your team via email.{" "}
            </li>
            <li>Assign project tasks to your team mates, or your self.</li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}
