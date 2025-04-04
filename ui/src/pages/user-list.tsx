import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useSortParams } from "@/hooks/use-sort-params";
import { userPaginate } from "@/lib/api";
import { UserInfo } from "@/schema.types";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router";

export default function UserListPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const { user } = useAuthProvider();
  const [users, setUsers] = useState<UserInfo[]>([]);
  const { searchParams } = useSortParams();
  const { page } = searchParams;

  useEffect(() => {
    const fetchUsers = async () => {
      setLoading(true);
      if (!user) {
        navigate(RouteMap.SIGNIN);
        setLoading(false);

        return;
      }
      try {
        const { data } = await userPaginate(user.tokens.access_token, {
          page: Number(page || 1),
        });
        if (!data) {
          setLoading(false);
          return;
        }
        console.log(data);
        setUsers(data);
        setLoading(false);
      } catch (error) {
        console.error("Error fetching users:", error);
        setLoading(false);
      }
    };
    fetchUsers();
  }, [page]);
  if (loading) {
    return <div>Loading...</div>;
  }
  return (
    <div className="flex w-full flex-col items-center justify-center">
      <h1>Users</h1>
      <ul>
        {users.length &&
          user &&
          !loading &&
          users.map((user) => (
            <li key={user.id}>
              Id: {user.id} Email: {user.email}
            </li>
          ))}
      </ul>
    </div>
  );
}
