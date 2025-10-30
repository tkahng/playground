import { CreateTeamDialog } from "@/components/create-team-dialog";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { useDialog } from "@/hooks/use-dialog";
import { useUserTeams } from "@/hooks/use-user-teams";
import { Team } from "@/schema.types";
import { Check, ChevronsUpDown } from "lucide-react";
import { useState } from "react";

export type TeamSelectProps = {
  onTeamSelect: (team: Team) => void;
};
export function TeamSelect({ onTeamSelect }: TeamSelectProps) {
  // const navigate = useNavigate();
  const { props } = useDialog();
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
    onTeamSelect(team);
    props.onOpenChange(false);
  }
  return (
    <div className="ml-6">
      <Popover open={props.open} onOpenChange={props.onOpenChange}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={props.open}
            aria-label="Select a team"
            className="w-[200px] justify-between"
          >
            {selectedTeam ? selectedTeam.name : "Select a team"}
            <ChevronsUpDown className="ml-auto h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[200px] p-0">
          <Command>
            <CommandList>
              <CommandInput placeholder="Select team..." />
              <CommandEmpty>No team found.</CommandEmpty>
              <CommandGroup heading="Teams">
                {data.data.map((te) => (
                  <CommandItem
                    key={te.id}
                    onSelect={() => {
                      handleSelectTeam(te);
                    }}
                    disabled={te.member?.role !== "owner"}
                    className="text-sm"
                  >
                    <Avatar className="mr-2 h-5 w-5">
                      <AvatarFallback>
                        {te.name.slice(0, 2).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1">
                      <div className="font-medium">{te.name}</div>
                      {/* <div className="text-xs text-muted-foreground">
                        {team.plan} • {team.role}
                      </div> */}
                    </div>
                    <Check
                      className={`ml-auto h-4 w-4 ${
                        te?.id === selectedTeam?.id
                          ? "opacity-100"
                          : "opacity-0"
                      }`}
                    />
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
            <CommandSeparator />
            <CommandList>
              <CommandGroup>
                <CommandItem>
                  <CreateTeamDialog />
                </CommandItem>
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
    </div>
  );
}
