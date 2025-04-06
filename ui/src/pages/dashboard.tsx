import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useNavigate } from "react-router";

export default function Dashboard() {
  const { logout } = useAuthProvider();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate(RouteMap.SIGNIN);
  };

  return (
    <div className="flex w-full flex-col items-center justify-center">
      <h2>Dashboard - Protected Route</h2>
      <button onClick={handleLogout}>Logout</button>
    </div>
  );
}
