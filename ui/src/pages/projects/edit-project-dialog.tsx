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
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { taskProjectUpdate } from "@/lib/queries";
import { TaskStatus } from "@/schema.types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

const formSchema = z.object({
  name: z.string().min(1),
  description: z.string().min(0).optional(),
  status: z.enum(["todo", "in_progress", "done"]),
  rank: z.number(),
});

export function ProjectEditDialog({
  project,
}: {
  project: {
    id: string;
    name: string;
    description: string;
    status: TaskStatus;
    rank: number;
  };
}) {
  const { user } = useAuthProvider();
  const [isDialogOpen, setDialogOpen] = useState(false);
  const queryClient = useQueryClient();

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: project.name,
      description: project.description,
      status: project.status,
      rank: project.rank,
    },
  });

  const mutation = useMutation({
    mutationFn: async (values: z.infer<typeof formSchema>) => {
      if (!user?.tokens.access_token) {
        throw new Error("Missing access token");
      }
      await taskProjectUpdate(user.tokens.access_token, project.id, values);
    },
    onSuccess: async () => {
      setDialogOpen(false);
      await queryClient.invalidateQueries({
        queryKey: ["project-with-tasks", project.id],
      });
      await queryClient.invalidateQueries({
        queryKey: ["recent-projects"],
      });
      toast.success("Project updated successfully");
    },
    onError: (error) => {
      toast.error(`Failed to update project: ${error.message}`);
    },
  });
  const onSubmit = (values: z.infer<typeof formSchema>) => {
    mutation.mutate(values);
  };
  return (
    <Dialog open={isDialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        <Button variant="outline">Edit Project</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Edit Project</DialogTitle>
          <DialogDescription>Edit the project details</DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
            <div className="grid gap-4 py-4">
              <div className="w-full px-10">
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Name</FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          placeholder="Task Name"
                          defaultValue={project.name}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="description"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Description</FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          placeholder="Task Description"
                          defaultValue={project.description}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="status"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Status</FormLabel>
                      <Select
                        onValueChange={field.onChange}
                        defaultValue={project.status}
                      >
                        <FormControl {...field}>
                          <SelectTrigger>
                            <SelectValue placeholder="Select Task Status" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          <SelectItem value="todo">Todo</SelectItem>
                          <SelectItem value="in_progress">
                            In Progress
                          </SelectItem>
                          <SelectItem value="done">Done</SelectItem>
                        </SelectContent>
                      </Select>
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="rank"
                  render={({ field }) => (
                    <FormItem hidden>
                      <FormLabel>Rank</FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          placeholder="Project Rank"
                          defaultValue={project.rank}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <DialogFooter>
                  <Button type="submit">Update Project</Button>
                </DialogFooter>
              </div>
            </div>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
