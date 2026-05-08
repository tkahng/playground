import PageSectionLayout from "@/layouts/page-section";
import JobsEdit from "@/pages/admin/jobs/jobs-edit";
import { createFileRoute } from "@tanstack/react-router";

function AdminJobEditPage() {
  return (
    <PageSectionLayout title="Jobs">
      <JobsEdit />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/jobs/$jobId"
)({
  component: AdminJobEditPage,
});
