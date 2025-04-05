import { columns, DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { userPaginate } from "@/lib/api";
import { UserInfo } from "@/schema.types";
import { PaginationState, Updater } from "@tanstack/react-table";
import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";

export default function RolesList() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const { user } = useAuthProvider();
  const [users, setUsers] = useState<UserInfo[]>([]);
  const [rowCount, setRowCount] = useState(0);

  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);
  const onPaginationChange = (updater: Updater<PaginationState>) => {
    const newState =
      typeof updater === "function"
        ? updater({ pageIndex, pageSize })
        : updater;
    setSearchParams({
      page: String(newState.pageIndex),
      per_page: String(newState.pageSize),
    });
  };

  useEffect(() => {
    const fetchUsers = async () => {
      setLoading(true);
      if (!user) {
        navigate(RouteMap.SIGNIN);
        setLoading(false);

        return;
      }
      try {
        console.log(pageIndex);
        const { data, meta } = await userPaginate(user.tokens.access_token, {
          page: pageIndex + 1,
          per_page: pageSize,
        });
        if (!data) {
          setLoading(false);
          return;
        }
        console.log(data);
        setLoading(false);
        setUsers(data);
        setRowCount(meta.total);
      } catch (error) {
        console.error("Error fetching users:", error);
        setLoading(false);
      }
    };
    fetchUsers();
  }, [pageIndex, pageSize]);
  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="flex w-full flex-col items-center justify-center">
      <h1>Users</h1>
      {users.length && user && !loading && (
        <DataTable
          columns={columns}
          data={users}
          rowCount={rowCount}
          pagination={{ pageIndex, pageSize }}
          onPaginationChange={onPaginationChange}
        />
      )}
    </div>
  );
}
