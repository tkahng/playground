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
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { useUserTeamMembers } from "@/hooks/use-user-team-members";
import { useUpdateMemberLastSelectedAt } from "@/lib/mutation";
import { Team } from "@/schema.types";
import { Check, ChevronsUpDown } from "lucide-react";
import { useState } from "react";
import { useNavigate, useParams } from "@tanstack/react-router";
import { CreateTeamDialog } from "./create-team-dialog";
import { CenteredSpinner } from "./centered-spinner";

export default function TeamSwitcher() {
  const { teamSlug } = useParams({ strict: false });
  const navigate = useNavigate();
  const { user } = useAuthProvider();
  const [open, setOpen] = useState(false);
  const {
    data,
    error: teamsError,
    isError: teamsIsError,
    isLoading: teamsLoading,
  } = useUserTeamMembers({ sort_by: "last_selected_at", sort_order: "desc" });
  const mutation = useUpdateMemberLastSelectedAt();
  const { team } = useTeam();
  if (!user) {
    return <></>;
  }
  if (teamsLoading) {
    return <CenteredSpinner />;
  }
  if (teamsIsError) {
    return <div>Error: {teamsError?.message}</div>;
  }

  function handleSelectTeam(team: Team) {
    mutation.mutate({ teamId: team.id });
    setOpen(false);
    navigate({ to: '/teams/$teamSlug/dashboard', params: { teamSlug: team.slug } });
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
            {teamSlug ? (
              <div className="flex items-center">
                <Avatar className="mr-2 h-5 w-5">
                  <AvatarFallback>
                    {team?.name.slice(0, 2).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <span className="truncate">{team?.name}</span>
              </div>
            ) : (
              <span className="truncate">Select a team</span>
            )}
            <ChevronsUpDown className="ml-auto h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[200px] p-0">
          <Command>
            <CommandList>
              <CommandInput placeholder="Search team..." />
              <CommandEmpty>No team found.</CommandEmpty>
              <CommandGroup heading="Teams">
                {data?.data.map((te) => (
                  <CommandItem
                    key={te.id}
                    onSelect={() => {
                      handleSelectTeam(te.team!);
                    }}
                    className="text-sm"
                  >
                    <Avatar className="mr-2 h-5 w-5">
                      <AvatarFallback>
                        {te.team?.name.slice(0, 2).toUpperCase()}
                      </AvatarFallback>
                    </Avatar>
                    <div className="flex-1">
                      <div className="font-medium">{te.team?.name}</div>
                      {/* <div className="text-xs text-muted-foreground">
                        {team.plan} • {team.role}
                      </div> */}
                    </div>
                    <Check
                      className={`ml-auto h-4 w-4 ${
                        te?.team_id === team?.id ? "opacity-100" : "opacity-0"
                      }`}
                    />
                  </CommandItem>
                ))}
              </CommandGroup>
            </CommandList>
            <CommandSeparator />
            <CommandList>
              <CommandGroup>
                <CommandItem className="items-center justify-center">
                  {user?.user.email_verified_at ? (
                    <CreateTeamDialog />
                  ) : (
                    <Tooltip>
                      <TooltipTrigger>
                        <Button disabled variant="outline">
                          Create Team
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>
                        <p>You must verify your email to create a team.</p>
                      </TooltipContent>
                    </Tooltip>
                  )}
                </CommandItem>
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
    </div>
  );
}
