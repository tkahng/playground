import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { FormControl } from "@/components/ui/form";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useDialog } from "@/hooks/use-dialog";
import { useTeam } from "@/hooks/use-team";
import { getTeamTeamMembers } from "@/lib/api";
import { cn } from "@/lib/utils";
import { TeamMember } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { Check, ChevronsUpDown } from "lucide-react";
import React from "react";

export function AssigneeComboBox({
  defaultValue,
  onValueChange,
}: {
  defaultValue: string | null;
  onValueChange(value: string): void;
}) {
  const { user } = useAuthProvider();
  const { teamMember } = useTeam();
  const { props } = useDialog();

  const [value, setValue] = React.useState<string | null>(defaultValue || null);
  const [members, setMembers] = React.useState<TeamMember[]>([]);
  const { data, isLoading, error } = useQuery({
    queryKey: ["team-members", teamMember?.team_id],
    queryFn: async () => {
      return await getTeamTeamMembers(
        user!.tokens.access_token,
        teamMember!.team_id,
        0,
        50
      );
    },
    enabled: !!teamMember?.team_id && !!user?.tokens.access_token && props.open,
  });

  React.useEffect(() => {
    if (data) {
      setMembers(data.data || []);
    }
  }, [data]);
  return (
    <Popover open={props.open} onOpenChange={props.onOpenChange}>
      <PopoverTrigger asChild>
        <FormControl>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={props.open}
            className="w-[200px] justify-between"
          >
            {value
              ? members.find((framework) => framework.id === value)?.user?.email
              : "Select framework..."}
            <ChevronsUpDown className="opacity-50" />
          </Button>
        </FormControl>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0">
        <Command>
          <CommandInput placeholder="Search framework..." className="h-9" />
          <CommandList>
            <CommandEmpty>No framework found.</CommandEmpty>
            <CommandGroup>
              {isLoading && <div>Loading...</div>}
              {error && <div>Error: {error?.message}</div>}
              {members.map((framework) => (
                <CommandItem
                  key={framework.id}
                  value={framework.id}
                  onSelect={(currentValue) => {
                    setValue(currentValue === value ? "" : currentValue);
                    onValueChange?.(currentValue);
                    props.onOpenChange?.(false);
                  }}
                >
                  {framework.user?.email}
                  <Check
                    className={cn(
                      "ml-auto",
                      value === framework.id ? "opacity-100" : "opacity-0"
                    )}
                  />
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
