import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { taskProjectList } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { NavLink, useSearchParams } from "react-router";
import { CreateProjectDialog } from "./create-project-dialog";

export default function ProjectListPage() {
  const { user } = useAuthProvider();
  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);
  // const queryClient = useQueryClient();
  const onPaginationChange = (updater: Updater<PaginationState>) => {
    const newState =
      typeof updater === "function"
        ? updater({ pageIndex, pageSize })
        : updater;
    if (newState.pageIndex !== pageIndex || newState.pageSize !== pageSize) {
      setSearchParams({
        page: String(newState.pageIndex),
        per_page: String(newState.pageSize),
      });
    }
  };

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["projects-list", pageIndex, pageSize],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      const data = await taskProjectList(user.tokens.access_token, {
        page: pageIndex + 1,
        per_page: pageSize,
      });

      return data;
    },
  });
  // const mutation = useMutation({
  //   mutationFn: async (roleId: string) => {
  //     if (!user?.tokens.access_token) {
  //       throw new Error("Missing access token or role ID");
  //     }
  //     await deleteRole(user.tokens.access_token, roleId);
  //   },
  //   onSuccess: () => {
  //     queryClient.invalidateQueries({ queryKey: ["roles-list"] });
  //     toast.success("Role deleted successfully");
  //   },
  //   onError: (error) => {
  //     console.error(error);
  //     toast.error("Failed to delete role");
  //   },
  // });
  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div>Error: {error?.message}</div>;
  }
  const projects = data?.data || [];
  const rowCount = data?.meta.total || 0;

  return (
    // <div className="flex w-full flex-col items-center justify-center">
    <div>
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Projects</h1>
        <CreateProjectDialog />
      </div>
      <p>
        Create and manage Projects for your applications. Projects contain
        collections of Tasks and can be assigned to Users.
      </p>
      <DataTable
        columns={[
          {
            accessorKey: "name",
            header: "Name",
            cell: ({ row }) => {
              return (
                <NavLink
                  to={`${RouteMap.TASK_PROJECTS}/${row.original.id}`}
                  className="hover:underline text-blue-500"
                >
                  {row.original.name}
                </NavLink>
              );
            },
          },
          {
            accessorKey: "description",
            header: "Description",
          },
          // {
          //   id: "actions",
          //   cell: ({ row }) => {
          //     return (
          //       <div className="flex flex-row gap-2 justify-end">
          //         <RoleEllipsisDropdown
          //           roleId={row.original.id}
          //           onDelete={(roleId) => {
          //             // mutation.mutate(roleId);
          //           }}
          //         />
          //       </div>
          //     );
          //   },
          // },
        ]}
        data={projects}
        rowCount={rowCount}
        paginationState={{ pageIndex, pageSize }}
        onPaginationChange={onPaginationChange}
        paginationEnabled
      />
    </div>
  );
}
