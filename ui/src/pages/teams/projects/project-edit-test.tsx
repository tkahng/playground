import { KanbanBoardProvider } from "@/components/ui/kanban";
import { useProject } from "@/hooks/use-project";
import { TaskStatus } from "@/schema.types";
import { MyKanbanBoard } from "./board";
import { ProjectEditDialog } from "./edit-project-dialog";

export default function ProjectEditTest() {
  const { data: project, isLoading: isProjectLoading, error } = useProject();

  if (isProjectLoading) return <p>Loading...</p>;
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
      <KanbanBoardProvider>
        <MyKanbanBoard projectId={project.id} />
      </KanbanBoardProvider>
    </div>
  );
}
