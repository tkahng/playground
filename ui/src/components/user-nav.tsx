import { LinkDto } from "@/components/links";
import { RouteMap } from "@/components/route-map";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
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
import { Link, useNavigate } from "react-router";
import ThemeSwitcher from "./theme-switcher";

type UserNavProps = {
  links: LinkDto[];
};

export function UserNav({ links }: UserNavProps) {
  const { user: auth, logout } = useAuthProvider();
  // const { setTheme, theme } = useTheme();
  const user = auth?.user;
  //   const { pathname } = useLocation();
  const isAdmin = auth?.roles?.includes("superuser");
  const navigate = useNavigate();
  const handleLogout = async (event: React.FormEvent) => {
    event.preventDefault();
    await logout();
    navigate(RouteMap.HOME);
  };
  if (!auth) {
    return (
      <Button variant="ghost" className="relative h-8 w-8 rounded-full">
        <Avatar>
          <AvatarImage src="https://avatars.githubusercontent.com/u/124599?v=4" />
          <AvatarFallback>SC</AvatarFallback>
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
        >
          <Avatar>
            <AvatarImage src="https://avatars.githubusercontent.com/u/124599?v=4" />
            <AvatarFallback>SC</AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56" align="end" forceMount>
        <DropdownMenuLabel className="font-normal">
          <div className="flex flex-col space-y-1">
            <p className="text-sm font-medium leading-none">{user?.name}</p>
            <p className="text-xs leading-none text-muted-foreground">
              {user?.email}
            </p>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuGroup>
          {links.map((link) => (
            <DropdownMenuItem key={link.to}>
              <Link to={link.to} className="w-full">
                {link.title}
              </Link>
            </DropdownMenuItem>
          ))}
        </DropdownMenuGroup>
        <>
          <DropdownMenuSeparator />
          <DropdownMenuItem>
            <Link to={RouteMap.LANDING_HOME} className="w-full">
              Home
            </Link>
          </DropdownMenuItem>
        </>
        {isAdmin && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem>
              <Link to={RouteMap.ADMIN} className="w-full">
                Admin
              </Link>
            </DropdownMenuItem>
          </>
        )}
        {/* <DropdownMenuGroup>
          <DropdownMenuItem> */}
        {/* </DropdownMenuItem> */}
        {/* <DropdownMenuItem onClick={() => setTheme("light")}>
            <span>Light</span>
            <Dot className={theme === "light" ? "ml-2" : "hidden"} />
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => setTheme("dark")}>
            <span>Dark</span>
            <Dot className={theme === "dark" ? "ml-2" : "hidden"} />
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => setTheme("system")}>
            <span>System</span>
            <Dot className={theme === "system" ? "ml-2" : "hidden"} />
          </DropdownMenuItem> */}
        {/* </DropdownMenuGroup> */}
        <DropdownMenuSeparator />
        <div className="flex">
          <DropdownMenuItem>
            <Button onClick={handleLogout}>
              {/* <NavLink onClick={handleLogout} to={RouteMap.HOME}> */}
              <span>Sign out</span>
              {/* </NavLink> */}
            </Button>
            {/* <DropdownMenuShortcut>⇧⌘Q</DropdownMenuShortcut> */}
          </DropdownMenuItem>
          <ThemeSwitcher />
        </div>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
