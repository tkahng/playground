// import { Button } from "@/components/ui/button";
// import {
//   Command,
//   CommandEmpty,
//   CommandGroup,
//   CommandInput,
//   CommandItem,
//   CommandList,
// } from "@/components/ui/command";
// import {
//   DialogDescription,
//   DialogFooter,
//   DialogHeader,
//   DialogTitle,
// } from "@/components/ui/dialog";
// import {
//   Form,
//   FormControl,
//   FormDescription,
//   FormField,
//   FormItem,
//   FormLabel,
//   FormMessage,
// } from "@/components/ui/form";
// import {
//   Popover,
//   PopoverContent,
//   PopoverTrigger,
// } from "@/components/ui/popover";
// import { useAuthProvider } from "@/hooks/use-auth-provider";
// import { useDialog } from "@/hooks/use-dialog";
// import { useUserTeams } from "@/hooks/use-user-teams";
// import { cn } from "@/lib/utils";
// import { Team } from "@/schema.types";
// import { zodResolver } from "@hookform/resolvers/zod";
// import { useQueryClient } from "@tanstack/react-query";
// import { Check, ChevronsUpDown } from "lucide-react";
// import { PropsWithChildren, useState } from "react";
// import { useForm } from "react-hook-form";
// import { z } from "zod";

// const formSchema = z.object({
//   assignee_id: z.string().nullable(),
// });

// export function TeamSelectDialog({
//   props,
//   children,
// }: PropsWithChildren<{
//   props?: {
//     open: boolean;
//     onOpenChange: React.Dispatch<React.SetStateAction<boolean>>;
//   };
// }>) {
//   const { user } = useAuthProvider();
//   const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
//   const { data, error: teamsError, isLoading: teamsLoading } = useUserTeams();
//   const assigneeDialog = useDialog();

//   const queryClient = useQueryClient();

//   const form = useForm<z.infer<typeof formSchema>>({
//     resolver: zodResolver(formSchema),
//     defaultValues: {
//       assignee_id: task.assignee_id,
//     },
//   });

//   const onSubmit = (values: z.infer<typeof formSchema>) => {
//     console.log(values);
//   };
//   return (
//     <>
//       <DialogHeader>
//         <DialogTitle>Upgrade Team</DialogTitle>
//         <DialogDescription>
//           Choose a team to upgrade to. You can only choose teams where you are a
//           owner.
//         </DialogDescription>
//       </DialogHeader>

//       <Form {...form}>
//         <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
//           <div className="grid gap-4 py-4">
//             <div className="w-full px-10 space-y-4">
//               <FormField
//                 control={form.control}
//                 name="assignee_id"
//                 render={({ field }) => (
//                   <FormItem className="flex flex-col">
//                     <FormLabel>Assignee</FormLabel>
//                     <Popover {...assigneeDialog.props}>
//                       <PopoverTrigger asChild>
//                         <FormControl>
//                           <Button
//                             variant="outline"
//                             role="combobox"
//                             className={cn(
//                               "w-[200px] justify-between",
//                               !field.value && "text-muted-foreground"
//                             )}
//                           >
//                             {field.value
//                               ? members?.data?.find((member) => {
//                                   return member.id === field.value;
//                                 })?.user?.email
//                               : "Select assignee"}
//                             <ChevronsUpDown className="opacity-50" />
//                           </Button>
//                         </FormControl>
//                       </PopoverTrigger>
//                       <PopoverContent
//                         aria-modal={true}
//                         className={cn("z-50 w-[200px] p-0")}
//                         style={{ pointerEvents: "auto" }}
//                         portal={false}
//                       >
//                         <Command>
//                           <CommandInput
//                             placeholder="Search assignee..."
//                             className="h-9"
//                           />
//                           <CommandList>
//                             <CommandEmpty>No assignee found.</CommandEmpty>
//                             <CommandGroup>
//                               <CommandItem
//                                 key={te.id}
//                                 onSelect={() => {
//                                   handleSelectTeam(te);
//                                 }}
//                                 disabled={te.member?.role !== "owner"}
//                                 onSelect={() => {
//                                   form.setValue(field.name, null, {
//                                     shouldDirty: true,
//                                   });
//                                   assigneeDialog.props.onOpenChange(false);
//                                 }}
//                               >
//                                 None
//                                 <Check
//                                   className={cn(
//                                     "ml-auto",
//                                     !field.value ? "opacity-100" : "opacity-0"
//                                   )}
//                                 />
//                               </CommandItem>
//                               {data?.data?.map((te) => (
//                                 <CommandItem
//                                   value={te.user?.email}
//                                   key={te.id}
//                                   onSelect={() => {
//                                     form.setValue(field.name, te.id, {
//                                       shouldDirty: true,
//                                     });
//                                     assigneeDialog.props.onOpenChange(false);
//                                   }}
//                                 >
//                                   {te.user?.email}
//                                   <Check
//                                     className={cn(
//                                       "ml-auto",
//                                       te.id === field.value
//                                         ? "opacity-100"
//                                         : "opacity-0"
//                                     )}
//                                   />
//                                 </CommandItem>
//                               ))}
//                             </CommandGroup>
//                           </CommandList>
//                         </Command>
//                       </PopoverContent>
//                     </Popover>
//                     <FormDescription>
//                       This is the language that will be used in the dashboard.
//                     </FormDescription>
//                     <FormMessage />
//                   </FormItem>
//                 )}
//               />
//               <DialogFooter>
//                 <Button type="submit" disabled={!form.formState.isDirty}>
//                   Update Task
//                 </Button>
//               </DialogFooter>
//             </div>
//           </div>
//         </form>
//       </Form>
//     </>
//   );
// }
