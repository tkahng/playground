import { KanbanBoard } from "@/components/board/kanban-board";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { taskList, taskProjectGet } from "@/lib/queries";
import { TaskStatus } from "@/schema.types";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router";
import { ProjectEditDialog } from "./edit-project-dialog";

export default function ProjectEdit() {
  const { user } = useAuthProvider();
  const { projectId } = useParams<{ projectId: string }>();
  const {
    data: project,
    isLoading: loading,
    error,
  } = useQuery({
    select: (data) => {
      return {
        ...data,
        tasks: data.tasks?.map((task) => ({
          name: task.name,
          order: task.order,
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
      try {
        const project = await taskProjectGet(
          user.tokens.access_token,
          projectId
        );
        const tasks = await taskList(user.tokens.access_token, {
          project_id: projectId,
          sort_by: "order",
          sort_order: "asc",
          per_page: 50,
        });
        return {
          ...project,
          tasks: tasks.data,
        };
      } catch (error) {
        throw error;
      }
    },
  });

  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;
  if (!project) return <p>Project not found</p>;

  return (
    <div className="space-y-6 px-12 py-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{project.name}</h1>
        <ProjectEditDialog
          project={{
            description: project.description || "",
            id: project.id,
            name: project.name,
            order: project.order,
            status: project.status as TaskStatus,
          }}
        />
      </div>
      <div>{project.description}</div>
      <p>
        Create and manage Roles for your applications. Roles contain collections
        of Permissions and can be assigned to Users.
      </p>
      <div className="flex flex-row gap-4">
        <div className="gap-4 flex-1">
          <div className="flex flex-col">
            <KanbanBoard cars={project.tasks || []} projectId={projectId!} />
          </div>
        </div>
      </div>
    </div>
  );
}
