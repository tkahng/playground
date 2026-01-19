import { CenteredSpinner } from "@/components/centered-spinner";
import { DataTableFooter } from "@/components/data-table";
import { ErrorCard } from "@/components/error-card";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
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
import { UserPlus } from "lucide-react";
import { Link, useSearchParams } from "react-router";
import { toast } from "sonner";
import { CreateProjectAiDialog } from "./projects/create-project-ai-dialog";
import { CreateProjectDialog } from "./projects/create-project-dialog";
import { ProjectCard } from "./projects/project-card";

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
            Manage your team's AI usage and collaboration
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button>
            <UserPlus className="mr-2 h-4 w-4" />
            Invite Member
          </Button>
        </div>
      </div>
      <Card>
        <CardContent className="m-8">
          <p>This is your Team's dashboard!</p>
          <br />
          <p>
            While we work on polishing this page, here are some things to try:
          </p>
          <ul className="list-disc mx-6">
            <li>
              Create a project with AI. Go to the{" "}
              <Link
                to={`/teams/${team.slug}/projects`}
                className="text-primary hover:text-accent-foreground underline hover:no-underline"
              >
                Team Projects page
              </Link>
              , click on the Create Project with AI button, and describe your
              project. It will generate a list of tasks for it!
            </li>
            <li>
              Invite a team member! Send invitations to your team via
              email.{" "}
            </li>
            <li>Assign project tasks to your team mates, or your self.</li>
          </ul>
        </CardContent>
      </Card>
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
    </div>
  );
}
