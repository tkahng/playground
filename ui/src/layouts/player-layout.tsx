import { CenteredSpinner } from "@/components/centered-spinner";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { usePlayer } from "@/hooks/use-current-player";
import { isErrorModel } from "@/lib/error";
import { rpsGameQueries } from "@/lib/rps-game-queries";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useEffect } from "react";
import { Controller, useForm } from "react-hook-form";
import { Outlet } from "react-router";
import { toast } from "sonner";
import z from "zod";

const formSchema = z.object({
  displayName: z.string().min(2, {
    message: "name must be at least 2 characters.",
  }),
});
export default function PlayerLayout() {
  const { user } = useAuthProvider();
  const { player, setPlayer } = usePlayer();
  const { data, isLoading, error, isError } = useQuery({
    queryKey: ["current-player"],
    queryFn: () => {
      if (!user?.tokens.access_token) {
        throw new Error("No access token");
      }
      return rpsGameQueries.getUserPlayer({
        token: user.tokens.access_token,
      });
    },
    enabled: !!user?.tokens.access_token,
  });

  const mutation = useMutation({
    mutationFn: (displayName: string) => {
      if (!user?.tokens.access_token) {
        throw new Error("No access token");
      }
      return rpsGameQueries.PutUserPlayer({
        token: user.tokens.access_token,
        displayName,
      });
    },
    onSuccess: (data) => {
      setPlayer(data.data);
      toast.success("Player updated successfully");
    },
  });

  useEffect(() => {
    if (data) {
      setPlayer(data.data);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data]);

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      displayName: "",
    },
  });

  function onSubmit(values: z.infer<typeof formSchema>) {
    mutation.mutate(values.displayName);
  }

  if (isError) {
    if (isErrorModel(error)) {
      toast.error(error.message);
    }
  }
  if (isLoading) {
    return <CenteredSpinner />;
  }
  if (!player) {
    return (
      <div>
        <h1>Create your player. enter a display name</h1>
        <form id="put-player" onSubmit={form.handleSubmit(onSubmit)}>
          <Controller
            name="displayName"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor={field.name}>Display Name</FieldLabel>
                <Input
                  {...field}
                  id={field.name}
                  aria-invalid={fieldState.invalid}
                  placeholder="Enter your display name"
                  autoComplete="off"
                />
                <FieldDescription>Enter your display name</FieldDescription>
                {fieldState.invalid && (
                  <FieldError errors={[fieldState.error]} />
                )}
              </Field>
            )}
          />
        </form>

        <Button type="submit" form="put-player">
          Submit
        </Button>
      </div>
    );
  }

  return <Outlet />;
}
