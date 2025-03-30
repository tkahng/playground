import { useNavigate } from "react-router";

export default function Dashboard() {
  const navigate = useNavigate();

  const handleLogout = () => {
    localStorage.removeItem("auth");
    navigate("/login");
  };

  return (
    <div>
      <h2>Dashboard - Protected Route</h2>
      <button onClick={handleLogout}>Logout</button>
    </div>
  );
}
