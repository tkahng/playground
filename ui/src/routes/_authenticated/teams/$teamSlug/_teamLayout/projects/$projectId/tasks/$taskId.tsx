import PageSectionLayout from "@/layouts/page-section";
import TaskLayout from "@/pages/teams/projects/task-layout";
import TaskEdit from "@/pages/teams/projects/tasks/task-edit";
import { createFileRoute } from "@tanstack/react-router";

function TaskPage() {
  return (
    <PageSectionLayout title="Projects">
      <TaskLayout>
        <TaskEdit />
      </TaskLayout>
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/teams/$teamSlug/_teamLayout/projects/$projectId/tasks/$taskId"
)({
  component: TaskPage,
});
