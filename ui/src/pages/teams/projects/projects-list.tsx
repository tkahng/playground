import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { deleteTaskProject, taskProjectList } from "@/lib/task-queries";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { CreateProjectDialog } from "./create-project-dialog";
import { CenteredSpinner } from "@/components/centered-spinner";
import { CreateProjectAiDialog } from "./create-project-ai-dialog.tsx";
import { ProjectCard } from "./project-card.tsx";
import { ErrorCard } from "@/components/error-card.tsx";
import { ApiError } from "@/lib/error.ts";
import { toast } from "sonner";
import {
  getCoreRowModel,
  getPaginationRowModel,
  PaginationState,
  Updater,
  useReactTable,
  Table as ReactTable,
} from "@tanstack/react-table";
import { useSearchParams } from "react-router";
import { Button } from "@/components/ui/button.tsx";
import { ChevronLeft, ChevronRight } from "lucide-react";
import {
  Pagination,
  PaginationContent,
  PaginationItem,
} from "@/components/ui/pagination.tsx";

export default function ProjectListPage() {
  const { user } = useAuthProvider();
  const { team } = useTeam();
  const teamId = team?.id;
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "6", 10);
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
    queryKey: [{ key: "projects-list", page: pageIndex, per_page: pageSize }],
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
  const projects = data?.data || [];

  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({
    onPaginationChange,
    data: projects,
    columns: [],
    rowCount: data?.meta.total,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    manualPagination: true,
    state: {
      pagination: {
        pageIndex,
        pageSize,
      },
    },
  });
  const mutation = useMutation({
    mutationFn: async ({ projectId }: { projectId: string }) => {
      if (!user?.tokens.access_token) {
        throw new ApiError("Missing access token");
      }
      await deleteTaskProject({
        token: user.tokens.access_token,
        projectId: projectId,
      });
    },
    onSuccess: async () => {
      queryClient.invalidateQueries({
        queryKey: [{ key: "projects-list" }],
      });
      toast.success("Project deleted");
    },
    onError: (error: ApiError) => {
      queryClient.invalidateQueries({
        queryKey: [{ key: "projects-list" }],
      });
      toast.error(error.message, {
        description: error.detail,
      });
    },
  });
  if (!team) {
    return <ErrorCard title="Team not found" />;
  }
  if (isLoading) {
    return <CenteredSpinner />;
  }
  if (isError) {
    return <div>Error: {error.message}</div>;
  }
  return (
    <div className="mx-auto max-w-6xl px-4 py-12">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Your Projects</h1>
          <p className="mt-1 text-muted-foreground">
            Manage your projects and track progress across teams.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <CreateProjectDialog />
          <CreateProjectAiDialog />
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {table.getRowModel().rows.map(({ original: project }) => (
          <ProjectCard
            team={team}
            key={project.id}
            project={project}
            onDelete={(projectId) => mutation.mutate({ projectId })}
          />
        ))}
      </div>
      <DataTableFooter table={table} />
    </div>
  );
}

export function DataTableFooter<TData>({
  table,
}: {
  table: ReactTable<TData>;
}) {
  return (
    <>
      {table.getPageCount() > 1 && (
        <div className="mt-10">
          <Pagination>
            <PaginationContent>
              <PaginationItem>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => table.previousPage()}
                  disabled={!table.getCanPreviousPage()}
                  className="gap-1 bg-transparent"
                >
                  <ChevronLeft className="size-4" />
                  <span className="hidden sm:inline">Previous</span>
                </Button>
              </PaginationItem>

              <PaginationItem>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => table.nextPage()}
                  disabled={!table.getCanNextPage()}
                  className="gap-1 bg-transparent"
                >
                  <span className="hidden sm:inline">Next</span>
                  <ChevronRight className="size-4" />
                </Button>
              </PaginationItem>
            </PaginationContent>
          </Pagination>
        </div>
      )}
    </>
  );
}
