// import { formSchema } from "@/components/pricing/pricing";
// import { useAuthProvider } from "@/hooks/use-auth-provider";
// import { updateTask } from "@/lib/api";
// import { useTaskQuery } from "@/lib/queries";
// import { zodResolver } from "@hookform/resolvers/zod";
// import { useMutation, useQueryClient } from "@tanstack/react-query";
// import { useForm } from "react-hook-form";
// import { useParams } from "react-router";
// import { toast } from "sonner";
// import { z } from "zod";
// const formSchema = z.object({
//   name: z.string().min(1),
//   // name: string;
//   description: z.string().min(0).optional(),
//   // description?: string;
//   status: z.enum(["todo", "in_progress", "done"]),
//   // status: "todo" | "in_progress" | "done";
//   assignee_id: z.string().nullable(),
//   //  assignee_id: string | null;
//   // created_by_member_id: z.string().nullable(),
//   // created_by_member_id: string | null;
//   end_at: z.string().nullable(),
//   // end_at: string | null;
//   parent_id: z.string().nullable(),
//   // parent_id: string | null;
//   position: z.number().optional(),
//   // position?: number;
//   // project_id: z.string(),
//   // project_id: string;
//   rank: z.number().optional(),
//   // rank?: number;
//   reporter_id: z.string().nullable(),
//   // reporter_id: string | null;
//   start_at: z.string().nullable(),
//   // start_at: string | null;
//   // team_id: z.string(),
// });
// export default function TaskEdit() {
//   const { taskId } = useParams<{
//     projectId: string;
//     taskId: string;
//     teamId: string;
//   }>();
//   const {
//     data: task,
//     // isLoading: isTaskLoading,
//     // error: taskError,
//   } = useTaskQuery(taskId);
//   const { user } = useAuthProvider();
//   //   const { teamMember } = useTeam();

//   const queryClient = useQueryClient();

//   // const [search, setSearch] = useState<string>("");
//   // const navigate = useNavigate();
//   //   const { data, isFetched } = useQuery({
//   //     queryKey: ["team-members", teamMember?.team_id],
//   //     queryFn: async () => {
//   //       return await getTeamTeamMembers(
//   //         user!.tokens.access_token,
//   //         teamMember!.team_id,
//   //         0,
//   //         50
//   //       );
//   //     },
//   //     enabled: !!teamMember?.team_id && !!user?.tokens.access_token,
//   //   });

//   const form = useForm<z.infer<typeof formSchema>>({
//     resolver: zodResolver(formSchema),
//     defaultValues: {
//       name: task?.name,
//       description: task.description || "",
//       status: task.status,
//       assignee_id: task.assignee_id,
//       // created_by_member_id: task.created_by_member_id,
//       end_at: task.end_at,
//       parent_id: task.parent_id,
//       // project_id: task.project_id,
//       rank: task.rank,
//       reporter_id: task.reporter_id,
//       start_at: task.start_at,
//       // team_id: task.team_id,
//     },
//   });

//   const mutation = useMutation({
//     mutationFn: async (values: z.infer<typeof formSchema>) => {
//       if (!user?.tokens.access_token) {
//         throw new Error("Missing access token");
//       }
//       await updateTask(user.tokens.access_token, task.id, {
//         name: values.name,
//         status: values.status,
//         description: values.description || null,
//         assignee_id: values.assignee_id || null,
//         end_at: values.end_at || null,
//         parent_id: values.parent_id || null,
//         reporter_id: values.reporter_id || null,
//         start_at: values.start_at || null,
//       });
//     },
//     onSuccess: async () => {
//       await queryClient.invalidateQueries({
//         queryKey: ["project-with-tasks", task.project_id],
//       });
//       toast.success("Task created successfully");
//     },
//     onError: (error) => {
//       toast.error(`Failed to create task: ${error.message}`);
//     },
//   });
//   const onSubmit = (values: z.infer<typeof formSchema>) => {
//     mutation.mutate(values);
//   };
//   return <div>TaskEdit</div>;
// }
