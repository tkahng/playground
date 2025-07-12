import { Avatar } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeamNotifications } from "@/hooks/use-team-notifications";
import { Bell } from "lucide-react";
import { Link } from "react-router";

export function NotificationDropdown() {
  const { user } = useAuthProvider();
  const { notifications, notificationsLoading, notificationsIsError } =
    useTeamNotifications();
  if (!user || notificationsLoading || notificationsIsError) {
    return (
      <Button variant="ghost" className="relative h-8 w-8 rounded-full">
        <Avatar>
          <Bell />
        </Avatar>
      </Button>
    );
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="relative h-8 w-8 rounded-full shadow-sm border-2"
        ></Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56" align="end" forceMount>
        <DropdownMenuLabel className="font-normal">
          <div className="flex flex-col space-y-1">
            <p className="text-sm font-medium leading-none">Notifications</p>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          {notifications?.data?.map((link) => (
            <DropdownMenuItem key={link.id}>
              <Link to={link.id} className="w-full">
                {link.payload.notification.title}
              </Link>
            </DropdownMenuItem>
          ))}
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
