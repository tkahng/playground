import { useSearchParams } from "@/hooks/use-search-params";
import { CenteredSpinner } from "@/components/centered-spinner";
import { DataTablePagination } from "@/components/data-table-pagination";
import { ErrorCard } from "@/components/error-card";
import { Button } from "@/components/ui/button";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { ApiError } from "@/lib/error";
import { taskProjectList, deleteTaskProject } from "@/lib/task-queries";
import { useQueryClient, useQuery, useMutation } from "@tanstack/react-query";
import {
  PaginationState,
  Updater,
  useReactTable,
  getCoreRowModel,
  getPaginationRowModel,
} from "@tanstack/react-table";
import { FolderOpen, UserPlus } from "lucide-react";
import {  } from "@tanstack/react-router";
import { toast } from "sonner";
import { CreateProjectAiDialog } from "./projects/create-project-ai-dialog";
import { CreateProjectDialog } from "./projects/create-project-dialog";
import { ProjectCard } from "./projects/project-card";
import { Separator } from "@/components/ui/separator";

export default function TeamDashboard() {
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
    <div className="mx-auto px-8 py-8 justify-start items-stretch flex-1 max-w-[1200px]">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight">
            Welcome to <span className="font-extrabold">{team.name}</span>
          </h1>
          <p className="text-muted-foreground">
            Manage your team's projects, tasks, and members.
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button>
            <UserPlus className="mr-2 h-4 w-4" />
            Invite Member
          </Button>
        </div>
      </div>
      <Separator />
      <div className="mt-8 mx-auto max-w-6xl">
        <div className="mb-8 flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold tracking-tight">Your Projects</h2>
          </div>
          <div className="flex items-center gap-3">
            <CreateProjectDialog />
            <CreateProjectAiDialog />
          </div>
        </div>
        {projects.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-24 text-center">
            <div className="flex items-center justify-center w-16 h-16 rounded-full bg-muted mb-4">
              <FolderOpen className="h-8 w-8 text-muted-foreground" />
            </div>
            <h3 className="text-lg font-semibold mb-2">No projects yet</h3>
            <p className="text-muted-foreground text-sm max-w-xs mb-6">
              Create a project to start managing tasks with a Kanban board.
              Assign work to team members and track progress.
            </p>
            <div className="flex items-center gap-3">
              <CreateProjectDialog />
              <CreateProjectAiDialog />
            </div>
          </div>
        ) : (
          <>
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
            <DataTablePagination table={table} />
          </>
        )}
      </div>
    </div>
  );
}
