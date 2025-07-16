import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import MultipleSelector from "@/components/ui/multiple-selector";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import {
  adminStripeProductRolesCreate as adminStripeProductPermissionsCreate,
  permissionsPaginate,
} from "@/lib/api";
import { ProductWithPrices } from "@/schema.types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

const formSchema = z.object({
  // userId: z.string().uuid(),
  // roleIds: z.string().uuid().array().min(1),
  permissions: z
    .object({
      value: z.string().uuid(),
      label: z.string(),
    })
    .array()
    .min(1),
});

export function ProductPermissionsDialog({
  product,
}: {
  product: ProductWithPrices;
}) {
  const { user } = useAuthProvider();
  const [isDialogOpen, setDialogOpen] = useState(false);
  const queryClient = useQueryClient();
  // const [value, setValue] = useState<Option[]>([]);
  const productId = product?.id;
  const { data, isLoading, error } = useQuery({
    queryKey: ["product-permissions-reverse", productId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !productId) {
        throw new Error("Missing access token or role ID");
      }
      const { data } = await permissionsPaginate(user.tokens.access_token, {
        product_id: productId,
        product_reverse: true,
        page: 0,
        per_page: 50,
      });
      return data;
    },
  });
  const mutation = useMutation({
    mutationFn: async (values: z.infer<typeof formSchema>) => {
      if (!user?.tokens.access_token || !productId) {
        throw new Error("Missing access token or role ID");
      }
      await adminStripeProductPermissionsCreate(
        user.tokens.access_token,
        productId,
        {
          permission_ids: values.permissions.map(
            (permission) => permission.value
          ),
        }
      );
      setDialogOpen(false);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["product", productId],
      });
      await queryClient.invalidateQueries({
        queryKey: ["product-permissions-reverse", productId],
      });
    },
  });
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      permissions: [],
    },
  });

  const onSubmit = (values: z.infer<typeof formSchema>) => {
    mutation.mutate(values);
  };
  useEffect(() => {
    if (data) {
      form.reset({ permissions: [] });
    }
  }, [data, form]);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error.message}</div>;
  }

  if (!data?.length) {
    return <div>Permissions not found</div>;
  }
  return (
    <Dialog open={isDialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        <Button variant="outline">Assign Permissions</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Assign Permissions</DialogTitle>
          <DialogDescription>
            Select the permissions you want to assign to this product.
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <div className="grid">
              <div className="space-y-4">
                <FormField
                  control={form.control}
                  name="permissions"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Permissions</FormLabel>
                      <FormControl>
                        <MultipleSelector
                          {...field}
                          defaultOptions={data.map((permission) => ({
                            label: permission.name,
                            value: permission.id,
                          }))}
                          placeholder="Select Permissions you like..."
                          emptyIndicator={
                            <p className="text-center text-lg leading-10 text-gray-600 dark:text-gray-400">
                              no results found.
                            </p>
                          }
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <DialogFooter>
                  <Button type="submit">Assign Permissions</Button>
                </DialogFooter>
              </div>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
