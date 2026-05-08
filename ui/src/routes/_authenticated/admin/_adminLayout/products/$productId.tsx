import PageSectionLayout from "@/layouts/page-section";
import ProductEditPage from "@/pages/admin/products/products-edit";
import { createFileRoute } from "@tanstack/react-router";

function AdminProductEditPage() {
  return (
    <PageSectionLayout title="Products">
      <ProductEditPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/products/$productId"
)({
  component: AdminProductEditPage,
});
