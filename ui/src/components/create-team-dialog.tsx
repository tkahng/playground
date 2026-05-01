import React from "react";
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
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTeam } from "@/hooks/use-team";
import { GetError } from "@/lib/error";
import { createTeam } from "@/lib/team-queries";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { z } from "zod";

const formSchema = z.object({
  name: z.string().min(3, "Team name is required"),
  slug: z
    .string()
    .regex(/^[A-Za-z0-9-]+$/, {
      message: "Only alphanumeric characters and dashes are allowed",
    })
    .or(z.literal(""))
    .optional(),
});

export function CreateTeamDialog({
  trigger,
}: {
  trigger?: React.ReactNode;
}) {
  const { user } = useAuthProvider();
  const isUserVerified = !!user?.user?.email_verified_at;
  const { setTeam } = useTeam();
  const navigate = useNavigate();
  const [isDialogOpen, setDialogOpen] = useState(false);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (values: z.infer<typeof formSchema>) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token or user ID");
      }
      return await createTeam(user.tokens.access_token, {
        name: values.name,
        slug: values.slug == "" ? undefined : values.slug,
      });
    },
    onSuccess: async (data) => {
      await queryClient.invalidateQueries({
        queryKey: ["user-teams-list"],
      });
      setDialogOpen(false);
      setTeam(data);
      toast.success("Team created successfully");
      navigate(`/teams/${data.slug}/dashboard`);
    },
    onError: (error) => {
      const err = GetError(error);
      toast.error(err?.detail);
    },
  });
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: "",
      slug: undefined,
    },
  });

  const onSubmit = (values: z.infer<typeof formSchema>) => {
    mutation.mutate(values);
  };

  return (
    <Dialog open={isDialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        {trigger ?? (
          <Button variant="outline" disabled={!isUserVerified}>
            Create Team
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Create Team</DialogTitle>
          <DialogDescription>Create a new team.</DialogDescription>
        </DialogHeader>
        <form
          id="form-create-team-dialog"
          onSubmit={form.handleSubmit(onSubmit)}
        >
          <FieldGroup>
            <div className="grid">
              <div className="space-y-4">
                <Controller
                  control={form.control}
                  name="name"
                  render={({ field, fieldState }) => (
                    <Field>
                      <FieldLabel htmlFor="form-create-team-dialog-name">
                        Name
                      </FieldLabel>
                      <Input
                        {...field}
                        id="form-create-team-dialog-name"
                        aria-invalid={fieldState.invalid}
                        placeholder="Name"
                      />
                      {fieldState.invalid && (
                        <FieldError errors={[fieldState.error]} />
                      )}
                    </Field>
                  )}
                />
                <Controller
                  control={form.control}
                  name="slug"
                  render={({ field, fieldState }) => (
                    <Field>
                      <FieldLabel htmlFor="form-create-team-dialog-slug">
                        Slug(optional)
                      </FieldLabel>
                      <FieldDescription>
                        Must be alphanumeric without any special characters. If
                        none is provided, a random one will be generated
                      </FieldDescription>
                      <Input
                        {...field}
                        id="form-create-team-dialog-slug"
                        aria-invalid={fieldState.invalid}
                        placeholder="Slug"
                      />
                      {fieldState.invalid && (
                        <FieldError errors={[fieldState.error]} />
                      )}
                    </Field>
                  )}
                />

                <DialogFooter>
                  <Button type="submit" form="form-create-team-dialog">
                    Create Team
                  </Button>
                </DialogFooter>
              </div>
            </div>
          </FieldGroup>
        </form>
      </DialogContent>
    </Dialog>
  );
}
