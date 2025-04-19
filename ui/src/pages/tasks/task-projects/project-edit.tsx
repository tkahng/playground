import { KanbanBoard } from "@/components/board/KanbanBoard";
import { RouteMap } from "@/components/route-map";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useTabs } from "@/hooks/use-tabs";
import { taskList, taskProjectGet } from "@/lib/queries";
import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import { ChevronLeft } from "lucide-react";
import { useForm } from "react-hook-form";
import { Link, useNavigate, useParams } from "react-router";
import { z } from "zod";
import { CreateProjectTaskDialog } from "./create-project-task-dialog";

const formSchema = z.object({
  name: z.string().min(2, {
    message: "name must be at least 2 characters.",
  }),
  description: z
    .string()
    .min(2, { message: "description must be at least 2 characters." })
    .optional(),
});
export default function ProjectEdit() {
  const navigate = useNavigate();
  const { tab, onClick } = useTabs("general");
  // const queryClient = useQueryClient();
  const { user } = useAuthProvider();
  const { projectId } = useParams<{ projectId: string }>();
  const {
    data: project,
    isLoading: loading,
    error,
  } = useQuery({
    queryKey: ["project", projectId],
    queryFn: async () => {
      if (!user?.tokens.access_token || !projectId) {
        throw new Error("Missing access token or project ID");
      }
      const project = await taskProjectGet(user.tokens.access_token, projectId);
      const tasks = await taskList(user.tokens.access_token, {
        project_id: projectId,
        sort_by: "order",
        sort_order: "asc",
        per_page: 50,
      });
      return {
        ...project,
        tasks: tasks.data,
      };
    },
  });
  // const mutation = useMutation({
  //   mutationFn: (values: z.infer<typeof formSchema>) =>
  //     updateRole(user!.tokens.access_token, projectId!, values),
  //   onSuccess: async () => {
  //     await queryClient.invalidateQueries({
  //       queryKey: ["project", projectId],
  //     });
  //     const updatedRole = await queryClient.fetchQuery({
  //       queryKey: ["project", projectId],
  //       queryFn: () => taskProjectGet(user!.tokens.access_token, projectId!),
  //     });
  //     form.reset(updatedRole);
  //     toast.success("Project updated!");
  //   },
  //   onError: (err: any) => {
  //     toast.error(`Failed to update project: ${err.message}`);
  //   },
  // });
  // const deletePermissionMutation = useMutation({
  //   mutationFn: (permissionId: string) =>
  //     deleteRolePermission(user!.tokens.access_token, projectId!, permissionId),
  //   onSuccess: () => {
  //     queryClient.invalidateQueries({
  //       queryKey: ["project", projectId],
  //     });
  //     toast.success("Permission deleted!");
  //   },
  //   onError: (err: any) => {
  //     toast.error(`Failed to delete permission: ${err.message}`);
  //   },
  // });
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: project?.name || "",
      description: project?.description || "",
    },
  });
  function onSubmit(_: z.infer<typeof formSchema>) {
    // mutation.mutate(values);
  }
  // const onDelete = (permissionId: string) => {
  //   // deletePermissionMutation.mutate(permissionId);
  // };
  // useEffect(() => {
  //   if (project) {
  //     form.reset(project);
  //   }
  // }, [project, form.reset]);
  if (!user) {
    navigate(RouteMap.SIGNIN);
  }
  if (loading) return <p>Loading...</p>;
  if (error) return <p>Error: {error.message}</p>;
  if (!project) return <p>Project not found</p>;

  return (
    // <div className="h-full px-4 py-6 lg:px-8 space-y-6">
    <div>
      <Link
        to={RouteMap.TASK_PROJECTS}
        className="flex items-center gap-2 text-sm text-muted-foreground"
      >
        <ChevronLeft className="h-4 w-4" />
        Back to projects
      </Link>
      <h1 className="text-2xl font-bold">{project.name}</h1>
      <Tabs value={tab} onValueChange={onClick} className="h-full space-y-6">
        <TabsList>
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="tasks">Tasks</TabsTrigger>
        </TabsList>
        <TabsContent value="general">
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormDescription>
                      This is your public display name.
                    </FormDescription>
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
                      <Input placeholder="shadcn" {...field} />
                    </FormControl>
                    <FormDescription>
                      This is your public display name.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button type="submit" disabled={!form.formState.isDirty}>
                Submit
              </Button>
            </form>
          </Form>
        </TabsContent>
        <TabsContent value="tasks" className="flex flex-col gap-4 mx-4 ">
          <div className="space-y-4 flex flex-row space-x-16">
            <p className="flex-1">
              Add Tasks to this Project. Users who have this Project will
              receive all Tasks below that match the API of their login request.
            </p>
            <CreateProjectTaskDialog projectId={projectId!} />
          </div>
          <div className="flex flex-col grow">
            <KanbanBoard
              cars={
                project.tasks?.map((task) => ({
                  columnId: task.status as "todo" | "done" | "in_progress",
                  content: task.name,
                  id: task.id,
                })) || []
              }
              projectId={projectId!}
            />
          </div>
          {/* <DataTable
            columns={[
              {
                header: "Name",
                accessorKey: "name",
              },
              {
                header: "Description",
                accessorKey: "description",
              },
              {
                header: "Status",
                accessorKey: "status",
              },
              {
                header: "Order",
                accessorKey: "order",
              },
            ]}
            data={project.tasks || []}
          /> */}
        </TabsContent>
      </Tabs>
    </div>
  );
}

// function DeleteButton({
//   permissionId,
//   onDelete,
// }: {
//   permissionId: string;
//   onDelete: (permissionId: string) => void;
// }) {
//   const editDialog = useDialog();
//   return (
//     <>
//       <Button variant="outline" size="icon" onClick={editDialog.trigger}>
//         <Trash className="h-4 w-4" />
//       </Button>
//       <ConfirmDialog dialogProps={editDialog.props}>
//         <>
//           <DialogHeader>
//             <DialogTitle>Are you absolutely sure?</DialogTitle>
//           </DialogHeader>
//           {/* Dialog Content */}
//           <DialogDescription>This action cannot be undone.</DialogDescription>
//           <DialogFooter>
//             <DialogClose asChild>
//               <Button
//                 variant="outline"
//                 onClick={() => {
//                   console.log("cancel");
//                   // editDialog.props.onOpenChange(false);
//                 }}
//               >
//                 Cancel
//               </Button>
//             </DialogClose>
//             <DialogClose asChild>
//               <Button
//                 variant="destructive"
//                 onClick={() => {
//                   console.log("delete");
//                   // editDialog.props.onOpenChange(false);
//                   onDelete(permissionId);
//                 }}
//               >
//                 Delete
//               </Button>
//             </DialogClose>
//           </DialogFooter>
//         </>
//       </ConfirmDialog>
//     </>
//   );
// }
