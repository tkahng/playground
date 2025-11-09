"use client";

import { MoreHorizontalIcon } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { GetError } from "@/lib/get-error";
import { deleteMember, updateTeamMember } from "@/lib/team-queries";
import { TeamMember, TeamMemberRole } from "@/schema.types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";
import z from "zod";

const roles = [
  { value: "owner", label: "Owner" },
  { value: "member", label: "Member" },
  { value: "guest", label: "Guest" },
];

const updateFormSchema = z.object({
  role: z.enum(["owner", "member", "guest"]),
});

export function MemberRowDropdownMenuDialog({
  member,
}: {
  member: TeamMember;
}) {
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  const { user } = useAuthProvider();
  const queryClient = useQueryClient();

  const updateMutation = useMutation({
    mutationFn: async ({ role }: { role: TeamMemberRole }) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      await updateTeamMember({
        memberId: member.id,
        token: user.tokens.access_token,
        role,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [{ key: "team-team-members" }],
      });
      toast.success("Member updated successfully");
    },
    onError: (error) => {
      console.error(error);
      toast.error("Failed to update member");
    },
  });
  const updateForm = useForm<z.infer<typeof updateFormSchema>>({
    resolver: zodResolver(updateFormSchema),
    defaultValues: {
      role: member.role,
    },
  });
  function onUpdateSubmit(data: z.infer<typeof updateFormSchema>) {
    updateMutation.mutate(data);
  }
  const deleteMutation = useMutation({
    mutationFn: async () => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      await deleteMember({
        memberId: member.id,
        token: user.tokens.access_token,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [{ key: "team-team-members" }],
      });
      toast.success("Member deleted successfully");
    },
    onError: (error) => {
      const err = GetError(error);
      toast.error("Failed to delete member", {
        description: err?.detail,
      });
    },
  });
  const onDeleteSubmit = () => {
    deleteMutation.mutate();
  };
  return (
    <>
      {/* dropdown */}
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            aria-label="Open menu"
            className="data-[state=open]:bg-muted"
            size="icon-sm"
          >
            <MoreHorizontalIcon />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent className="w-[160px]" align="end">
          <DropdownMenuItem onSelect={() => setShowEditDialog(true)}>
            Edit
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            onSelect={() => setShowDeleteDialog(true)}
            variant="destructive"
          >
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      {/* edit dialog */}
      <Dialog open={showEditDialog} onOpenChange={setShowEditDialog}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Edit Member</DialogTitle>
            <DialogDescription>
              Make changes to your member here. Click save when you&apos;re
              done.
            </DialogDescription>
          </DialogHeader>
          <form
            id="form-update-team-member"
            onSubmit={updateForm.handleSubmit(onUpdateSubmit)}
          >
            <Controller
              name="role"
              control={updateForm.control}
              render={({ field, fieldState }) => {
                return (
                  <FieldGroup className="pb-3">
                    <Field>
                      <FieldContent>
                        <FieldLabel htmlFor="form-rhf-select-language">
                          Role
                        </FieldLabel>
                        <FieldDescription>
                          Select a role for this member
                        </FieldDescription>
                        {fieldState.invalid && (
                          <FieldError errors={[fieldState.error]} />
                        )}
                      </FieldContent>
                      <Select
                        name={field.name}
                        value={field.value}
                        onValueChange={field.onChange}
                      >
                        <SelectTrigger
                          id="form-rhf-select-language"
                          aria-invalid={fieldState.invalid}
                          className="min-w-[120px]"
                        >
                          <SelectValue placeholder="Select" />
                        </SelectTrigger>
                        <SelectContent position="item-aligned">
                          {roles.map((language) => (
                            <SelectItem
                              key={language.value}
                              value={language.value}
                            >
                              {language.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </Field>
                  </FieldGroup>
                );
              }}
            />
          </form>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline">Cancel</Button>
            </DialogClose>
            <Button type="submit" form="form-update-team-member">
              Save
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      {/* delete dialog */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>Delete Member</DialogTitle>
            <DialogDescription>
              Are you sure you want to do this?
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline">Cancel</Button>
            </DialogClose>
            <Button variant="destructive" onClick={onDeleteSubmit}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
