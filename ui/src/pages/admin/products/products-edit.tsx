import { DataTable } from "@/components/data-table";
import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import {
  DialogClose,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { ConfirmDialog, useDialog } from "@/hooks/use-dialog";
import {
  adminStripeProduct,
  adminStripeProductPermissionsDelete,
} from "@/lib/api";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ChevronLeft, Trash } from "lucide-react";
import { Link, useParams } from "react-router";
import { toast } from "sonner";
import { ProductPermissionsDialog } from "./product-permissions-dialog";

export default function ProductEditPage() {
  const { user } = useAuthProvider();

  const queryClient = useQueryClient();
  const { productId } = useParams<{ productId: string }>();
  const { data, isError, isLoading, error } = useQuery({
    queryKey: ["product", productId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !productId) {
        throw new Error("Missing access token");
      }
      return adminStripeProduct(user.tokens.access_token, productId);
    },
  });
  const deleteProductPermissionsMutation = useMutation({
    mutationFn: async (permissionId: string) => {
      if (!user?.tokens.access_token || !productId) {
        throw new Error("Missing access token");
      }
      // Call the API to delete the role
      return adminStripeProductPermissionsDelete(
        user.tokens.access_token,
        productId,
        permissionId
      );
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["product", productId] });
      toast.success("Role deleted successfully");
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete role: ${error.message}`);
    },
  });
  if (isLoading) {
    return <div>Loading...</div>;
  }
  if (isError) {
    return (
      <div>
        <h1>Error</h1>
        <p>{error.message}</p>
      </div>
    );
  }
  if (!data) {
    return <div>No product found</div>;
  }
  return (
    <div className="space-y-6">
      <Link
        to={RouteMap.ADMIN_PRODUCTS}
        className="flex items-center gap-2 text-sm text-muted-foreground"
      >
        <ChevronLeft className="h-4 w-4" />
        Back to Products
      </Link>
      <h1 className="text-2xl font-bold">{data?.name}</h1>
      <div className="space-y-4 flex flex-row space-x-16">
        <p className="flex-1">Add Permissions to this product.</p>
        <ProductPermissionsDialog product={data} />
      </div>
      <DataTable
        columns={[
          {
            header: "Role",
            accessorKey: "name",
          },
          {
            header: "Description",
            accessorKey: "description",
          },
          {
            id: "actions",
            cell: ({ row }) => {
              return (
                <div className="flex flex-row gap-2 justify-end">
                  <DeleteButton
                    onDelete={() => {
                      deleteProductPermissionsMutation.mutate(row.original.id);
                    }}
                    permissionId={row.original.id}
                    // disabled={!row.original.is_directly_assigned}
                  />
                </div>
              );
            },
          },
        ]}
        data={data?.permissions || []}
      />
    </div>
  );
}

function DeleteButton({
  permissionId,
  onDelete,
}: {
  permissionId: string;
  onDelete: (permissionId: string) => void;
}) {
  const editDialog = useDialog();
  return (
    <>
      <Button variant="outline" size="icon" onClick={editDialog.trigger}>
        <Trash className="h-4 w-4" />
      </Button>
      <ConfirmDialog dialogProps={editDialog.props}>
        <>
          <DialogHeader>
            <DialogTitle>Are you absolutely sure?</DialogTitle>
          </DialogHeader>
          {/* Dialog Content */}
          <DialogDescription>This action cannot be undone.</DialogDescription>
          <DialogFooter>
            <DialogClose asChild>
              <Button
                variant="outline"
                onClick={() => {
                  console.log("cancel");
                  // editDialog.props.onOpenChange(false);
                }}
              >
                Cancel
              </Button>
            </DialogClose>
            <DialogClose asChild>
              <Button
                variant="destructive"
                onClick={() => {
                  console.log("delete");
                  // editDialog.props.onOpenChange(false);
                  onDelete(permissionId);
                }}
              >
                Delete
              </Button>
            </DialogClose>
          </DialogFooter>
        </>
      </ConfirmDialog>
    </>
  );
}
