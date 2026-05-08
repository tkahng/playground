import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
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
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Popover, PopoverTrigger } from "@/components/ui/popover";
import { PopoverContentNoPortal } from "@/components/ui/popover-noportal";
import { useDialog } from "@/hooks/use-dialog";
import { useUserTeamMembers } from "@/hooks/use-user-team-members";
import { cn } from "@/lib/utils";
import { zodResolver } from "@hookform/resolvers/zod";
import { Check, ChevronsUpDown } from "lucide-react";
import { JSX, PropsWithChildren, useState } from "react";
import { useForm } from "react-hook-form";
import { Link } from "@tanstack/react-router";
import { z } from "zod";
import { CenteredSpinner } from "../centered-spinner";
const formSchema = z.object({
  teamSlug: z.string().nullable(),
});

export function OwnerTeamSelectDialog({
  children,
}: PropsWithChildren<unknown>): JSX.Element {
  const [selectedSLug, setSelectedSLug] = useState<string | null>(null);
  const {
    data: teamsData,
    error: teamsError,
    isLoading: teamsLoading,
  } = useUserTeamMembers({
    sort_by: "last_selected_at",
    sort_order: "desc",
    roles: ["owner"],
  });
  const teamDialog = useDialog();

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      teamSlug: null,
    },
  });

  const onSubmit = (values: z.infer<typeof formSchema>) => {
    console.log(values);
  };
  if (teamsLoading) {
    return <CenteredSpinner/>;
  }
  if (teamsError) {
    return <div>Error: {teamsError?.message}</div>;
  }
  if (!teamsData) {
    return <div>No teams available.</div>;
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

        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
            <div className="grid gap-4 py-4">
              <div className="w-full px-10 space-y-4">
                <FormField
                  control={form.control}
                  name="teamSlug"
                  render={({ field }) => (
                    <FormItem className="flex flex-col">
                      <FormLabel>Team</FormLabel>
                      <Popover {...teamDialog.props}>
                        <PopoverTrigger asChild>
                          <FormControl>
                            <Button
                              variant="outline"
                              role="combobox"
                              className={cn(
                                "w-[200px] justify-between",
                                !field.value && "text-muted-foreground"
                              )}
                            >
                              {field.value
                                ? teamsData.data.find((team) => {
                                    return team.id === field.value;
                                  })?.team!.name
                                : "Select team"}
                              <ChevronsUpDown className="opacity-50" />
                            </Button>
                          </FormControl>
                        </PopoverTrigger>
                        <PopoverContentNoPortal
                          aria-modal={true}
                          className={cn("z-50 w-[200px] p-0")}
                          style={{ pointerEvents: "auto" }}
                        >
                          <Command>
                            <CommandInput
                              placeholder="Search assignee..."
                              className="h-9"
                            />
                            <CommandList>
                              <CommandEmpty>No assignee found.</CommandEmpty>
                              <CommandGroup>
                                {teamsData?.data?.map((te) => (
                                  <CommandItem
                                    value={te.id}
                                    key={te.id}
                                    onSelect={() => {
                                      setSelectedSLug(te.team!.slug);
                                      form.setValue(field.name, te.id, {
                                        shouldDirty: true,
                                      });
                                      teamDialog.props.onOpenChange(false);
                                    }}
                                  >
                                    {te.team?.name}
                                    <Check
                                      className={cn(
                                        "ml-auto",
                                        te.id === field.value
                                          ? "opacity-100"
                                          : "opacity-0"
                                      )}
                                    />
                                  </CommandItem>
                                ))}
                              </CommandGroup>
                            </CommandList>
                          </Command>
                        </PopoverContentNoPortal>
                      </Popover>
                      <FormDescription>
                        This is the language that will be used in the dashboard.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                {/* <DialogFooter>
                  <Button type="submit" disabled={!form.formState.isDirty}>
                    Update Task
                  </Button>
                </DialogFooter> */}
              </div>
            </div>
          </form>
        </Form>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          {selectedSLug && (
            <Button asChild>
              <Link to={`/teams/${selectedSLug}/settings/billing`}>
                Continue
              </Link>
            </Button>
          )}
          {!selectedSLug && <Button disabled>Continue</Button>}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
