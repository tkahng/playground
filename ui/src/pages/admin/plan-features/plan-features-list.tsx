import { CenteredSpinner } from "@/components/centered-spinner";
import { DataTable } from "@/components/data-table";
import { Button } from "@/components/ui/button";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { adminPlanFeaturesList } from "@/lib/api";
import { useQuery } from "@tanstack/react-query";
import { Pencil } from "lucide-react";
import { Link, useNavigate } from "@tanstack/react-router";

export default function PlanFeaturesListPage() {
  const { user } = useAuthProvider();

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["plan-features-list"],
    queryFn: async () => {
      if (!user?.tokens.access_token) throw new Error("Missing access token");
      return adminPlanFeaturesList(user.tokens.access_token);
    },
  });

  if (isLoading) return <CenteredSpinner />;
  if (isError) return <div>Error: {error.message}</div>;

  return (
    <div className="space-y-6">
      <p>
        Configure daily AI token limits per Stripe product. These limits control
        how many tokens a team on a given plan can consume per day.
      </p>
      <DataTable
        columns={[
          {
            accessorKey: "stripe_product_id",
            header: "Product ID",
            cell: ({ row }) => (
              <Link
                to='/admin/plan-features/$productId' params={{ productId: row.original.stripe_product_id }}
                className="hover:underline text-blue-500"
              >
                {row.original.stripe_product_id}
              </Link>
            ),
          },
          {
            accessorKey: "daily_ai_tokens",
            header: "Daily AI Tokens",
            cell: ({ row }) =>
              row.original.daily_ai_tokens.toLocaleString(),
          },
          {
            accessorKey: "updated_at",
            header: "Updated At",
            cell: ({ row }) =>
              new Date(row.original.updated_at).toLocaleDateString(),
          },
          {
            id: "actions",
            cell: ({ row }) => (
              <div className="flex flex-row gap-2 justify-end">
                <EditButton productId={row.original.stripe_product_id} />
              </div>
            ),
          },
        ]}
        data={data ?? []}
      />
    </div>
  );
}

function EditButton({ productId }: { productId: string }) {
  const navigate = useNavigate();
  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={() =>
        navigate({ to: '/admin/plan-features/$productId', params: { productId } })
      }
    >
      <Pencil className="h-4 w-4" />
    </Button>
  );
}
