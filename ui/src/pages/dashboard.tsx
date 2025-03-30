import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useNavigate } from "react-router";

export default function Dashboard() {
  const auth = useAuthProvider();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await auth.logout();
    navigate("/login");
  };

  return (
    <div>
      <h2>Dashboard - Protected Route</h2>
      <button onClick={handleLogout}>Logout</button>
    </div>
  );
}
