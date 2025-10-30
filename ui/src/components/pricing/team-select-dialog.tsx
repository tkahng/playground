import { TeamSelect } from "@/components/pricing/team-select-combo";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useUserTeams } from "@/hooks/use-user-teams";
import { Team } from "@/schema.types";
import { PropsWithChildren, useState } from "react";
import { Link } from "react-router";

export function TeamSelectDialog({ children }: PropsWithChildren<unknown>) {
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
  const { data, error: teamsError, isLoading: teamsLoading } = useUserTeams();
  if (teamsLoading) {
    return <div>Loading...</div>;
  }
  if (teamsError) {
    return <div>Error: {teamsError?.message}</div>;
  }
  if (!data) {
    return <div>No teams available.</div>;
  }
  function handleSelectTeam(team: Team) {
    setSelectedTeam(team);
  }
  return (
    <Dialog>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Upgrade Team</DialogTitle>
          <DialogDescription>
            Choose a team to upgrade to. You can only choose teams where you are
            a owner.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4">
          <TeamSelect onTeamSelect={handleSelectTeam} />
        </div>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button asChild disabled={!selectedTeam}>
            <Link to={`/teams/${selectedTeam?.slug}/settings/billing`}>
              Continue
            </Link>
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
