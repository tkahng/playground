import PageSectionLayout from "@/layouts/page-section";
import JobsListPage from "@/pages/admin/jobs/jobs-list";
import { createFileRoute } from "@tanstack/react-router";

function AdminJobsPage() {
  return (
    <PageSectionLayout title="Jobs">
      <JobsListPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/jobs/"
)({
  component: AdminJobsPage,
});
