import { DataTable } from "@/components/data-table";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { taskProjectList } from "@/lib/queries";
import { useQuery } from "@tanstack/react-query";
import { PaginationState, Updater } from "@tanstack/react-table";
import { NavLink, useSearchParams } from "react-router";
import { CreateProjectAiDialog } from "./create-project-ai-dialog";
import { CreateProjectDialog } from "./create-project-dialog";

export default function ProjectListPage() {
  const { user } = useAuthProvider();
  const [searchParams, setSearchParams] = useSearchParams();
  const { team, error: teamError } = useTeam();
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

  const teamId = team?.id;
  const { data, error, isError, isLoading } = useQuery({
    queryKey: ["projects-list", pageIndex, pageSize],
    queryFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or role ID");
      }
      if (!teamId) {
        throw new Error("Current team member team ID is required");
      }
      const data = await taskProjectList(user.tokens.access_token, teamId, {
        page: pageIndex,
        per_page: pageSize,
        sort_by: "updated_at",
        sort_order: "desc",
      });

      return data;
    },
    enabled: !!user?.tokens.access_token && !!teamId,
  });
  if (teamError) {
    return <div>Error: {teamError.message}</div>;
  }
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
                  to={`/teams/${team?.slug}/projects/${row.original.id}`}
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
