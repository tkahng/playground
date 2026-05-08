import { CenteredSpinner } from "@/components/centered-spinner";
import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { adminPlanFeaturesGet, adminPlanFeaturesUpsert } from "@/lib/api";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ChevronLeft } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { Link, useParams } from "@tanstack/react-router";
import { toast } from "sonner";
import { z } from "zod";

const schema = z.object({
  daily_ai_tokens: z.coerce.number().int().min(0),
});

type FormValues = z.infer<typeof schema>;

export default function PlanFeaturesEditPage() {
  const { user } = useAuthProvider();
  const { productId } = useParams({ strict: false });
  const queryClient = useQueryClient();

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["plan-features", productId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !productId)
        throw new Error("Missing access token");
      return adminPlanFeaturesGet(user.tokens.access_token, productId);
    },
    retry: false,
  });

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { daily_ai_tokens: 0 },
  });

  useEffect(() => {
    if (data) {
      form.reset({ daily_ai_tokens: data.daily_ai_tokens });
    }
  }, [data, form]);

  const mutation = useMutation({
    mutationFn: async (values: FormValues) => {
      if (!user?.tokens.access_token || !productId)
        throw new Error("Missing access token");
      return adminPlanFeaturesUpsert(user.tokens.access_token, productId, {
        daily_ai_tokens: values.daily_ai_tokens,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["plan-features-list"] });
      await queryClient.invalidateQueries({
        queryKey: ["plan-features", productId],
      });
      toast.success("Plan features updated");
    },
    onError: (err: Error) => {
      toast.error(`Failed to update: ${err.message}`);
    },
  });

  if (isLoading) return <CenteredSpinner />;
  if (isError)
    return (
      <div className="space-y-4">
        <Link
          to={RouteMap.ADMIN_PLAN_FEATURES}
          className="flex items-center gap-2 text-sm text-muted-foreground"
        >
          <ChevronLeft className="h-4 w-4" />
          Back to Plan Features
        </Link>
        <p className="text-destructive">{error.message}</p>
        <p className="text-sm text-muted-foreground">
          No plan features row exists for this product yet. Save to create one.
        </p>
        <PlanFeaturesForm
          form={form}
          onSubmit={(v) => mutation.mutate(v)}
          isPending={mutation.isPending}
          productId={productId ?? ""}
        />
      </div>
    );

  return (
    <div className="space-y-6">
      <Link
        to={RouteMap.ADMIN_PLAN_FEATURES}
        className="flex items-center gap-2 text-sm text-muted-foreground"
      >
        <ChevronLeft className="h-4 w-4" />
        Back to Plan Features
      </Link>
      <h1 className="text-2xl font-bold">{productId}</h1>
      <PlanFeaturesForm
        form={form}
        onSubmit={(v) => mutation.mutate(v)}
        isPending={mutation.isPending}
        productId={productId ?? ""}
      />
    </div>
  );
}

function PlanFeaturesForm({
  form,
  onSubmit,
  isPending,
  productId,
}: {
  form: ReturnType<typeof useForm<FormValues>>;
  onSubmit: (values: FormValues) => void;
  isPending: boolean;
  productId: string;
}) {
  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4 max-w-md">
        <FormField
          control={form.control}
          name="daily_ai_tokens"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Daily AI Token Limit</FormLabel>
              <FormControl>
                <Input type="number" min={0} {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <p className="text-xs text-muted-foreground">
          Product: <span className="font-mono">{productId}</span>
        </p>
        <Button type="submit" disabled={isPending}>
          {isPending ? "Saving…" : "Save"}
        </Button>
      </form>
    </Form>
  );
}
