import { KanbanBoard } from "@/components/board/kanban-board";
import { CenteredSpinner } from "@/components/centered-spinner";
import { Input } from "@/components/ui/input";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useOnboardingProgress } from "@/hooks/use-onboarding-progress";
import { useProject } from "@/hooks/use-project";
import { useTeam } from "@/hooks/use-team";
import { taskList, workflowList } from "@/lib/task-queries";
import { WorkflowStatus } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { ChevronLeft } from "lucide-react";
import { useEffect, useState } from "react";
import { ProjectEditDialog } from "./edit-project-dialog";

export default function ProjectEdit() {
  const { user } = useAuthProvider();
  const { markStep } = useOnboardingProgress();
  const { team } = useTeam();

  useEffect(() => {
    markStep("hasProject");
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const { data: project, isLoading: isProjectLoading, error } = useProject();
  const [input, setInput] = useState("");

  const { data: workflowStatuses, isLoading: isWorkflowLoading } = useQuery({
    queryKey: [{ key: "workflow-statuses", workflow_id: project?.workflow_id }],
    queryFn: async (): Promise<WorkflowStatus[]> => {
      if (!project?.workflow_id) return [];
      const workflows = await workflowList(
        user!.tokens.access_token,
        project.team_id,
        { ids: [project.workflow_id], applies_to: ["task"] },
      );
      return workflows?.[0]?.statuses ?? [];
    },
    enabled: !!user?.tokens.access_token && !!project?.workflow_id,
  });

  const {
    data: tasks,
    isLoading: isTasksLoading,
    error: tasksError,
  } = useQuery({
    select: (data) => {
      const res = data.data?.map((task) => ({
        task: task,
        name: task.name,
        rank: task.rank,
        columnId: task.workflow_status_id ?? "",
        content: task.description,
        id: task.id,
      }));
      return { meta: data.meta, data: res };
    },
    queryKey: [{ key: "project-tasks", project_id: project?.id, input }],
    queryFn: async () => {
      return await taskList(user!.tokens.access_token, project!.id, {
        sort_by: "rank",
        sort_order: "asc",
        per_page: 100,
        q: input,
      });
    },
    enabled: !!user?.tokens.access_token && !!project?.id,
  });

  if (isProjectLoading || isWorkflowLoading) return <CenteredSpinner />;
  if (error) return <p>Error: {error.message}</p>;
  if (!project) return <p>Project not found</p>;
  if (isTasksLoading) return <CenteredSpinner />;
  if (tasksError) return <p>Error: {tasksError.message}</p>;

  return (
    <div className="flex-1 space-y-6 w-full px-8">
      <Link
        to="/teams/$teamSlug/dashboard"
        params={{ teamSlug: team?.slug ?? "" }}
        className="flex items-center gap-2 text-sm text-muted-foreground"
      >
        <ChevronLeft className="h-4 w-4" />
        Back to Projects
      </Link>
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{project.name}</h1>
        <ProjectEditDialog
          project={{
            description: project.description || "",
            id: project.id,
            name: project.name,
            rank: project.rank,
            status: project.status,
          }}
        />
      </div>
      <div>{project.description}</div>
      <div>
        <Input
          id="search"
          placeholder="Filter tasks..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          className="h-8 w-[150px] lg:w-[250px]"
        />
      </div>
      {tasks?.meta && tasks.meta.total > 100 && (
        <p className="text-sm text-muted-foreground">
          Showing 100 of {tasks.meta.total} tasks. Use the filter to narrow
          results.
        </p>
      )}
      <KanbanBoard
        cards={tasks?.data ?? []}
        projectId={project.id!}
        workflowStatuses={workflowStatuses ?? []}
      />
    </div>
  );
}
