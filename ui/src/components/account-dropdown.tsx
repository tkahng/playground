import { useAuthProvider } from "@/hooks/use-auth-provider";
import { Avatar, AvatarFallback } from "@radix-ui/react-avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  Separator,
} from "@radix-ui/react-dropdown-menu";
import { Settings } from "lucide-react";
import { Link, useNavigate } from "react-router";
import { RouteMap } from "./route-map";
import { Button } from "./ui/button";

export default function AccountDropdown() {
  const { user, logout } = useAuthProvider();
  // const role = AuthHelper.getSubscriptionRoles(user);
  const role = "ahah";
  // const { pathname } = useLocation();
  // const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const handleLogout = async (event: React.FormEvent) => {
    event.preventDefault();
    // setLoading(true);
    await logout();
    navigate(RouteMap.HOME);
  };
  return (
    <DropdownMenu defaultOpen={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="relative h-8 w-8 rounded-full border-solid"
        >
          <Avatar>
            {/* <AvatarImage
              width={32}
              height={32}
              src="/placeholder.svg?height=32&width=32"
              alt="User avatar"
            /> */}
            <AvatarFallback>CN</AvatarFallback>
          </Avatar>
        </Button>
      </DropdownMenuTrigger>
      {/* <DropdownMenuContent className="container rounded p-1" align="end"> */}
      <DropdownMenuContent align="end" className="container">
        <DropdownMenuLabel>
          <div className="flex flex-col space-y-1">
            <p className="text-sm font-medium leading-none">
              {user?.user?.name}
            </p>
            <p className="text-xs leading-none text-muted-foreground">
              {user?.user?.email}
            </p>
          </div>
          <Separator className="h-3 bg-transparent" />
          <div className="flex flex-col space-y-1">
            <p className="text-sm font-medium leading-none">Current Plant</p>
            <p className="text-xs leading-none text-muted-foreground">
              {role ?? "Free"}
            </p>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem>
          <Link to={RouteMap.LANDING_HOME} className="flex items-center">
            <Settings className="mr-2 h-4 w-4" />
            <span>Home</span>
          </Link>
        </DropdownMenuItem>
        <DropdownMenuItem>
          <Link to={RouteMap.DASHBOARD_HOME} className="flex items-center">
            <Settings className="mr-2 h-4 w-4" />
            <span>Dashboard</span>
          </Link>
        </DropdownMenuItem>
        <DropdownMenuItem>
          <Link to={RouteMap.ACCOUNT_SETTINGS} className="flex items-center">
            <Settings className="mr-2 h-4 w-4" />
            <span>Settings</span>
          </Link>
        </DropdownMenuItem>
        <DropdownMenuItem>
          {/* <form onSubmit={handleLogout}> */}
          <Button onClick={handleLogout}>
            {/* <NavLink onClick={handleLogout} to={RouteMap.HOME}> */}
            <span>Sign out</span>
            {/* </NavLink> */}
          </Button>
          {/* </form> */}
          {/* <Form action={RouteMap.SIGNOUT} method="post">
            <input type="hidden" name="pathname" value={pathname} />
            <button type="submit" className="relative flex">
              <LogOut className="mr-2 h-4 w-4" />
              <span>Log out</span>
            </button>
          </Form> */}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
