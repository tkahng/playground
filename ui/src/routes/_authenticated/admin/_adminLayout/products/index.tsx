import PageSectionLayout from "@/layouts/page-section";
import ProductsListPage from "@/pages/admin/products/products-list";
import { createFileRoute } from "@tanstack/react-router";

function AdminProductsPage() {
  return (
    <PageSectionLayout title="Products">
      <ProductsListPage />
    </PageSectionLayout>
  );
}

export const Route = createFileRoute(
  "/_authenticated/admin/_adminLayout/products/"
)({
  component: AdminProductsPage,
});
