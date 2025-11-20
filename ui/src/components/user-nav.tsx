import { NavbarLink } from "@/components/link/nav-link";
import { RouteLinks, userDropdownLinks } from "@/components/links";
import { ModeToggle } from "@/components/mode-toggle";
import { RouteMap } from "@/components/route-map";
import { themes } from "@/components/themes";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
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
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTheme } from "next-themes";
import { Link, useNavigate } from "react-router";

export function UserNav() {
  const { theme, setTheme } = useTheme();
  const { user: auth, logout } = useAuthProvider();
  const user = auth?.user;
  const isAdmin = auth?.roles?.includes("superuser");
  const links2 = [...userDropdownLinks, ...(isAdmin ? [RouteLinks.ADMIN] : [])];
  const navigate = useNavigate();

  const handleLogout = async (event: React.FormEvent) => {
    event.preventDefault();
    await logout();
    navigate(RouteMap.HOME);
  };
  if (!auth) {
    return (
      <>
        <NavbarLink title="Sign In" to={RouteMap.SIGNIN} />
        <NavbarLink title="Sign Up" to={RouteMap.SIGNUP} />
        <ModeToggle />
      </>
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
          {links2.map((link) => (
            <>
              <DropdownMenuItem key={link.to}>
                <Link to={link.to} className="w-full">
                  <div className="flex flex-row gap-2 items-center">
                    {link.icon && link.icon}
                    {link.title}
                  </div>
                </Link>
              </DropdownMenuItem>
            </>
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
