import { KanbanBoard } from "@/components/board/kanban-board";
import { Input } from "@/components/ui/input";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useOnboardingProgress } from "@/hooks/use-onboarding-progress";
import { useProject } from "@/hooks/use-project";
import { useTeam } from "@/hooks/use-team";
import { taskList } from "@/lib/task-queries";
import { TaskStatus } from "@/schema.types";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { ChevronLeft } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "@tanstack/react-router";
import { ProjectEditDialog } from "./edit-project-dialog";
import { CenteredSpinner } from "@/components/centered-spinner";

export default function ProjectEdit() {
  const queryClient = useQueryClient();
  const { user } = useAuthProvider();
  const { markStep } = useOnboardingProgress();
  const { team } = useTeam();

  useEffect(() => {
    markStep("hasProject");
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  const { data: project, isLoading: isProjectLoading, error } = useProject();
  const [input, setInput] = useState("");
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
        columnId: task.status as "todo" | "done" | "in_progress",
        content: task.description,
        id: task.id,
      }));
      return {
        meta: data.meta,
        data: res,
      };
    },
    queryKey: [{ key: "project-tasks", project_id: project?.id, input }],
    queryFn: async () => {
      return await taskList(user!.tokens.access_token, project!.id, {
        sort_by: "rank",
        sort_order: "asc",
        per_page: 50,
        q: input,
      });
    },
    enabled: !!user?.tokens.access_token && !!project?.id,
  });

  useEffect(() => {
    queryClient.invalidateQueries({
      queryKey: [{ key: "project-tasks", project_id: project?.id }],
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [input]);

  if (isProjectLoading) return <CenteredSpinner />;
  if (error) return <p>Error: {error.message}</p>;
  if (!project) return <p>Project not found</p>;
  if (isTasksLoading) return <CenteredSpinner />;
  if (tasksError) return <p>Error: {tasksError.message}</p>;
  return (
    <div className="flex-1 space-y-6 w-full px-8">
      <Link
        to={`/teams/${team?.slug}/dashboard`}
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
            status: project.status as TaskStatus,
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
      <p>Manage your tasks.</p>
      <KanbanBoard cards={tasks?.data || []} projectId={project.id!} />

      {/* <KanbanBoardProvider>
        <MyKanbanBoard projectId={project.id} tasks={tasks?.data || []} />
      </KanbanBoardProvider> */}
    </div>
  );
}
