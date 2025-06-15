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
import { useTeam } from "@/hooks/use-team";
import { useUserTeams } from "@/hooks/use-user-teams";
import { Team } from "@/schema.types";
import { Check, ChevronsUpDown, Plus } from "lucide-react";
import { useState } from "react";
import { useNavigate } from "react-router";

export default function TeamSwitcher() {
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const {
    data: teams,
    error: teamsError,
    isLoading: teamsLoading,
  } = useUserTeams();
  const { team, isLoading: teamLoading, error: teamError } = useTeam();
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(team);

  if (teamsLoading || teamLoading) {
    return <div>Loading...</div>;
  }
  if (teamsError || teamError) {
    return <div>Error: {teamsError?.message || teamError?.message}</div>;
  }
  if (!teams || teams.data.length === 0 || !selectedTeam) {
    return <div>No teams available.</div>;
  }
  function handleSelectTeam(team: Team) {
    setSelectedTeam(team);
    setOpen(false);
    // Optionally, you can add logic to navigate to the selected team's dashboard
    // For example, using a router:
    navigate(`/teams/${team.slug}/dashboard`);
  }
  return (
    <div className="ml-6">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            aria-label="Select a team"
            className="w-[200px] justify-between"
          >
            <div className="flex items-center">
              <Avatar className="mr-2 h-5 w-5">
                <AvatarFallback>
                  {selectedTeam?.name.slice(0, 2).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <span className="truncate">{selectedTeam?.name}</span>
            </div>
            <ChevronsUpDown className="ml-auto h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[200px] p-0">
          <Command>
            <CommandList>
              <CommandInput placeholder="Search team..." />
              <CommandEmpty>No team found.</CommandEmpty>
              <CommandGroup heading="Teams">
                {teams.data.map((team) => (
                  <CommandItem
                    key={team.id}
                    onSelect={() => {
                      handleSelectTeam(team);
                    }}
                    className="text-sm"
                  >
                    <Avatar className="mr-2 h-5 w-5">
                      {/* <AvatarImage
                        src={team.avatar || "/placeholder.svg"}
                        alt={team.name}
                      /> */}
                      <AvatarFallback>
                        {team.name.slice(0, 2).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1">
                      <div className="font-medium">{team.name}</div>
                      {/* <div className="text-xs text-muted-foreground">
                        {team.plan} • {team.role}
                      </div> */}
                    </div>
                    <Check
                      className={`ml-auto h-4 w-4 ${
                        selectedTeam?.id === team.id
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
                  <Plus className="mr-2 h-4 w-4" />
                  Create Team
                </CommandItem>
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
    </div>
  );
}
