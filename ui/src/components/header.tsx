import { LinkDto } from "@/components/landing-links";
import { NavLink } from "@/components/link/nav-link";
import { ModeToggle } from "@/components/mode-toggle";
import NexusAILogo from "@/components/nexus-logo";
import NonAuthenticatedButton from "@/components/non-authenticated-button";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";

export function NexusAILandingHeader({
  leftLinks,
  rightLinks,
}: {
  leftLinks?: LinkDto[];
  rightLinks?: LinkDto[];
}) {
  // const [loading, setLoading] = useState(false);
  const { user } = useAuthProvider();
  // const navigate = useNavigate();
  // const handleLogout = async (event: React.FormEvent) => {
  //   event.preventDefault();
  //   setLoading(true);
  //   await logout();
  //   navigate(RouteMap.HOME);
  // };
  return (
    <header className="shadow-sm">
      {/* <nav className="container mx-auto flex h-14 items-center justify-between  lg:px-6"> */}
      <nav className=" flex h-14 items-center justify-between  lg:px-6">
        <div className="flex flex-grow items-center content-center">
          <NexusAILogo />
          {leftLinks?.length &&
            leftLinks.map(({ to: href, title }) => (
              <NavLink key={title} title={title} to={href} />
            ))}
        </div>
        <div className="flex shrink items-center space-x-4">
          {rightLinks?.length &&
            rightLinks.map(({ to: href, title }) => (
              <NavLink key={title} title={title} to={href} />
            ))}
          <ModeToggle />
          {
            user ? (
              <>
                <NavLink title="Dashboard" to={RouteMap.DASHBOARD_HOME} />
                {/* <AccountDropdown /> */}
              </>
            ) : (
              <NonAuthenticatedButton />
            )
            // (
            //   <form onSubmit={handleLogout}>
            //     <Button type="submit" disabled={loading} variant="default">
            //       {/* <NavLink onClick={handleLogout} to={RouteMap.HOME}> */}
            //       <span>Sign out</span>
            //       {/* </NavLink> */}
            //     </Button>
            //   </form>
            // )
          }
          {/* <AuthButton /> */}
        </div>
      </nav>
    </header>
  );
}
