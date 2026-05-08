import { CreateTeamDialog } from "@/components/create-team-dialog";
import { NavbarLink } from "@/components/link/nav-link";
import { RouteLinks, userDropdownLinks } from "@/components/links";
import { ModeToggle } from "@/components/mode-toggle";
import { RouteMap } from "@/components/route-map";
import { themes } from "@/components/themes";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuPortal,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { useUserTeamMembers } from "@/hooks/use-user-team-members";
import { ApiError } from "@/lib/error";
import { useUpdateMemberLastSelectedAt } from "@/lib/mutation";
import { Team } from "@/schema.types";
import { Check, CheckCircle, CircleX } from "lucide-react";
import { useTheme } from "next-themes";
import { Link, useNavigate } from "@tanstack/react-router";
import { Spinner } from "./ui/spinner";

export function UserNav() {
  const { theme, setTheme } = useTheme();
  const { user: auth, logout, checkAuth } = useAuthProvider();
  const navigate = useNavigate();
  const {
    data,
    error: teamsError,
    isError: teamsIsError,
    isLoading: teamsLoading,
  } = useUserTeamMembers({ sort_by: "last_selected_at", sort_order: "desc" });
  const mutation = useUpdateMemberLastSelectedAt();
  const { team } = useTeam();
  const user = auth?.user;
  const isAdmin = auth?.roles?.includes("superuser");
  const links2 = [...userDropdownLinks, ...(isAdmin ? [RouteLinks.ADMIN] : [])];
  function handleSelectTeam(team: Team) {
    navigate(`/teams/${team.slug}/dashboard`, { flushSync: true });
  }
  const handleLogout = async (event: React.FormEvent) => {
    event.preventDefault();
    await logout();
    navigate(RouteMap.HOME);
  };
  if (!auth) {
    return (
      <div className="flex items-center gap-4">
        <NavbarLink title="Sign In" to={RouteMap.SIGNIN} />
        <NavbarLink title="Sign Up" to={RouteMap.SIGNUP} />
        <ModeToggle />
      </div>
    );
  }
  if (teamsLoading) {
    return <Spinner />;
  }
  if (teamsIsError) {
    if (ApiError.isApiError(teamsError)) {
      if (teamsError.status === 401) {
        checkAuth();
      }
    }
    <div className="flex items-center gap-4">
      <NavbarLink title="Sign In" to={RouteMap.SIGNIN} />
      <NavbarLink title="Sign Up" to={RouteMap.SIGNUP} />
      <ModeToggle />
    </div>;
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="relative h-8 w-8 rounded-full shadow-sm border-2"
        >
          <Avatar>
            <AvatarFallback>
              {auth.user.email.charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56" align="end" forceMount>
        <DropdownMenuLabel className="font-normal flex justify-between">
          <div className="flex flex-col space-y-1">
            <p className="text-sm font-medium leading-none">{user?.name}</p>
            <p className="text-xs leading-none text-muted-foreground">
              {user?.email}
            </p>
          </div>
          <Tooltip>
            <TooltipTrigger className="h-8 w-8">
              {auth.user.email_verified_at ? (
                <CheckCircle className="text-green-600 dark:text-green-300" />
              ) : (
                <CircleX className="h-8 w-8 text-destructive" />
              )}
            </TooltipTrigger>
            <TooltipContent>
              {auth.user.email_verified_at
                ? "Email verified"
                : "Email not verified"}
            </TooltipContent>
          </Tooltip>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuSub>
            <DropdownMenuSubTrigger>Teams</DropdownMenuSubTrigger>
            <DropdownMenuPortal>
              <DropdownMenuSubContent>
                <Command>
                  <CommandList>
                    <CommandInput placeholder="Search team..." />
                    <CommandEmpty>No team found.</CommandEmpty>
                    <CommandGroup key={"teams"} heading="Teams">
                      {data?.data.map((te) => (
                        <CommandItem
                          key={te.id}
                          onSelect={() => {
                            mutation.mutate({ teamId: te.team_id });
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
                              te?.team_id === team?.id
                                ? "opacity-100"
                                : "opacity-0"
                            }`}
                          />
                        </CommandItem>
                      ))}
                    </CommandGroup>
                    <CommandGroup key={"create-team"}>
                      <CommandItem className="items-center justify-center">
                        {user?.email_verified_at ? (
                          <CreateTeamDialog />
                        ) : (
                          <Tooltip>
                            <TooltipTrigger>
                              <Button disabled variant="outline">
                                Create Team
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent>
                              <p>
                                You must verify your email to create a team.
                              </p>
                            </TooltipContent>
                          </Tooltip>
                        )}
                      </CommandItem>
                    </CommandGroup>
                  </CommandList>
                </Command>
              </DropdownMenuSubContent>
            </DropdownMenuPortal>
          </DropdownMenuSub>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          {links2.map((link, index) => (
            <div key={index}>
              <DropdownMenuItem key={link.to}>
                <Link to={link.to} className="w-full">
                  <div className="flex flex-row gap-2 items-center">
                    {link.icon && link.icon}
                    {link.title}
                  </div>
                </Link>
              </DropdownMenuItem>
            </div>
          ))}
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuSub>
            <DropdownMenuSubTrigger>Themes</DropdownMenuSubTrigger>
            <DropdownMenuPortal>
              <DropdownMenuSubContent>
                <DropdownMenuRadioGroup
                  value={theme}
                  onValueChange={(value) => {
                    setTheme(value);
                  }}
                >
                  {themes.map((theme) => (
                    <DropdownMenuRadioItem
                      key={theme.value}
                      value={theme.value}
                      onSelect={(e) => e.preventDefault()}
                    >
                      {theme.name}
                    </DropdownMenuRadioItem>
                  ))}
                </DropdownMenuRadioGroup>
              </DropdownMenuSubContent>
            </DropdownMenuPortal>
          </DropdownMenuSub>
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuItem>
          <Button onClick={handleLogout} variant={"ghost"}>
            Sign out
          </Button>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
