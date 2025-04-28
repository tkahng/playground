import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { taskProjectList } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { NavLink, useSearchParams } from "react-router";
import { CreateProjectAiDialog } from "./create-project-ai-dialog";
import { CreateProjectDialog } from "./create-project-dialog";

export default function ProjectListPage() {
  const { user, checkAuth } = useAuthProvider();
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

  const { data, error, isError, isLoading } = useQuery({
    queryKey: ["projects-list", pageIndex, pageSize],
    queryFn: async () => {
      await checkAuth(); // Ensure user is authenticated
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      const data = await taskProjectList(user.tokens.access_token, {
        page: pageIndex,
        per_page: pageSize,
        sort_by: "updated_at",
        sort_order: "desc",
      });

      return data;
    },
  });
  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return <div>Error: {error.message}</div>;
  }
  const projects = data?.data || [];
  const rowCount = data?.meta.total || 0;

  return (
    // <div className="flex w-full flex-col items-center justify-center">
    <div className="py-12 px-4 @lg:px-6 @xl:px-12 @2xl:px-20 @3xl:px-24">
      <div className="flex items-center justify-between">
        <p>
          Create and manage Projects for your applications. Projects contain
          collections of Tasks and can be assigned to Users.
        </p>
        <CreateProjectDialog />
        <CreateProjectAiDialog />
      </div>

      <DataTable
        columns={[
          {
            accessorKey: "id",
            header: "ID",
            cell: ({ row }) => {
              return (
                <NavLink
                  to={`${RouteMap.TASK_PROJECTS}/${row.original.id}`}
                  className="hover:underline text-blue-500"
                >
                  {row.original.id}
                </NavLink>
              );
            },
          },
          {
            accessorKey: "name",
            header: "Name",
          },
          {
            accessorKey: "description",
            header: "Description",
          },
          {
            accessorKey: "updated_at",
            header: "Updated At",
            cell: ({ row }) => {
              return new Date(row.original.updated_at).toLocaleDateString();
            },
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
