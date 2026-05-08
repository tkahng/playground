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
import { useTeam } from "@/hooks/use-team";
import { useTeamNotifications } from "@/hooks/use-team-notifications";
import { useUnreadNotificationCount } from "@/hooks/use-unread-notification-count";
import { Bell } from "lucide-react";
import { Link } from "@tanstack/react-router";

export function NotificationDropdown() {
  const { team } = useTeam();
  const { notifications, notificationsLoading, notificationsIsError } =
    useTeamNotifications();
  const { unreadCount } = useUnreadNotificationCount();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="relative">
          <Bell className="h-5 w-5" />
          {unreadCount > 0 && (
            <span className="absolute -top-1 -right-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-500 text-[10px] font-bold text-white">
              {unreadCount > 99 ? "99+" : unreadCount}
            </span>
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-72" align="end" forceMount>
        <DropdownMenuLabel className="font-normal">
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium">Notifications</p>
            {unreadCount > 0 && (
              <span className="text-xs text-muted-foreground">
                {unreadCount} unread
              </span>
            )}
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          {notificationsLoading && (
            <DropdownMenuItem disabled>Loading…</DropdownMenuItem>
          )}
          {notificationsIsError && (
            <DropdownMenuItem disabled>Failed to load</DropdownMenuItem>
          )}
          {!notificationsLoading &&
            !notificationsIsError &&
            (notifications?.data?.length === 0 ? (
              <DropdownMenuItem disabled>No notifications</DropdownMenuItem>
            ) : (
              notifications?.data?.map((n) => (
                <DropdownMenuItem key={n.id} className="flex flex-col items-start gap-0.5">
                  <span className={`text-sm ${!n.read_at ? "font-semibold" : "font-normal"}`}>
                    {n.payload.notification.title}
                  </span>
                  {n.payload.notification.body && (
                    <span className="text-xs text-muted-foreground line-clamp-1">
                      {n.payload.notification.body}
                    </span>
                  )}
                </DropdownMenuItem>
              ))
            ))}
        </DropdownMenuGroup>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          <DropdownMenuItem asChild>
            <Link to='/teams/$teamSlug/notifications' params={{ teamSlug: team?.slug ?? '' }} className="w-full cursor-pointer">
              View all notifications
            </Link>
          </DropdownMenuItem>
        </DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
