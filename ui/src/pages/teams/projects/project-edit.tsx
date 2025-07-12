import { KanbanBoard } from "@/components/board/kanban-board";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { getTeamBySlug, taskList, taskProjectGet } from "@/lib/queries";
import { TaskStatus } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router";
import { ProjectEditDialog } from "./edit-project-dialog";

export default function ProjectEdit() {
  const { user } = useAuthProvider();
  const { projectId, teamSlug } = useParams<{
    projectId: string;
    teamSlug: string;
  }>();
  const {
    data: project,
    isLoading: loading,
    error,
  } = useQuery({
    select: (data) => {
      return {
        ...data,
        tasks: data.tasks?.map((task) => ({
          task: task,
          name: task.name,
          rank: task.rank,
          columnId: task.status as "todo" | "done" | "in_progress",
          content: task.description,
          id: task.id,
        })),
      };
    },
    queryKey: ["project-with-tasks", projectId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !projectId) {
        throw new Error("Missing access token or project ID");
      }
      if (!teamSlug) {
        throw new Error("Current team member team ID is required");
      }
      const team = await getTeamBySlug(user.tokens.access_token, teamSlug);
      if (!team) {
        throw new Error("Team not found");
      }
      const project = await taskProjectGet(user.tokens.access_token, projectId);
      const tasks = await taskList(user.tokens.access_token, projectId, {
        sort_by: "order",
        sort_order: "asc",
        per_page: 50,
      });
      return {
        ...project,
        tasks: tasks.data,
      };
    },
  });

  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;
  if (!project) return <p>Project not found</p>;

  return (
    <div className="flex-1 space-y-6 w-full px-8">
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
      <p>
        Create and manage Roles for your applications. Roles contain collections
        of Permissions and can be assigned to Users.
      </p>
      <KanbanBoard cars={project.tasks || []} projectId={projectId!} />
    </div>
  );
}
