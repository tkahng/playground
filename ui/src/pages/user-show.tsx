import { useAuthProvider } from "@/hooks/use-auth-provider";

export default function Dashboard() {
  const { user: auth } = useAuthProvider();
  // const [user, setUser] = useState<UserInfo(auth);
  // const navigate = useNavigate();

  return (
    <div className="flex w-full flex-col items-center justify-center">
      <h2>Dashboard - Protected Route</h2>
      {/* <button onClick={handleLogout}>Logout</button> */}
    </div>
  );
}
