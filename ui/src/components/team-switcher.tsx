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
import { Check, ChevronsUpDown } from "lucide-react";
import { useState } from "react";
import { useNavigate } from "react-router";
import { CreateTeamDialog } from "./create-team-dialog";

export default function TeamSwitcher() {
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const { data, error: teamsError, isLoading: teamsLoading } = useUserTeams();
  const { team, setTeam } = useTeam();

  if (teamsLoading) {
    return <div>Loading...</div>;
  }
  if (teamsError) {
    return <div>Error: {teamsError?.message}</div>;
  }
  if (!data || data.data.length === 0) {
    return <div>No teams available.</div>;
  }
  function handleSelectTeam(team: Team) {
    setOpen(false);
    setTeam(team);
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
                  {team?.name.slice(0, 2).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <span className="truncate">{team?.name}</span>
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
                {data.data.map((te) => (
                  <CommandItem
                    key={te.id}
                    onSelect={() => {
                      handleSelectTeam(te);
                    }}
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
                        te?.id === team?.id ? "opacity-100" : "opacity-0"
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
                  {/* <Plus className="mr-2 h-4 w-4" />
                  Create Team */}
                </CommandItem>
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
    </div>
  );
}
