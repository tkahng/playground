import PageSectionLayout from "@/layouts/page-section";
import TaskLayout from "@/pages/teams/projects/task-layout";
import ProjectEdit from "@/pages/teams/projects/project-edit";
import { createFileRoute } from "@tanstack/react-router";

function ProjectPage() {
  return (
    <PageSectionLayout title="Projects">
      <TaskLayout>
        <ProjectEdit />
      </TaskLayout>
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/teams/$teamSlug/_teamLayout/projects/$projectId/"
)({
  component: ProjectPage,
});
