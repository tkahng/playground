import { Button } from "@/components/ui/button";
import {
  DialogClose,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ConfirmDialog, useDialog } from "@/hooks/use-dialog";
import { MemberEditDialog } from "@/pages/teams/settings/member-edit-dialog";
import { Ellipsis, Trash } from "lucide-react";
import { useState } from "react";
export function TeamMemberActionDropdown({
  memberId,
  role,
  onDelete,
}: {
  memberId: string;
  role: "owner" | "member" | "guest";
  onDelete: (memberId: string) => void;
}) {
  const editDialog = useDialog();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  return (
    <>
      <DropdownMenu open={dropdownOpen} onOpenChange={setDropdownOpen}>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon">
            <Ellipsis className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem
            onSelect={() => {
              setDropdownOpen(false);
            }}
          >
            <MemberEditDialog member={{ id: memberId, role: role }} />
          </DropdownMenuItem>
          <DropdownMenuItem
            onSelect={() => {
              setDropdownOpen(false);
              editDialog.trigger();
            }}
          >
            <Button variant="ghost" size="sm">
              <Trash className="h-4 w-4" />
              <span>Remove</span>
            </Button>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
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
                  onDelete(memberId);
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
// export function TeamMemberActionDropdown({ memberId }: { memberId: string }) {
//   const [dropdownOpen, setDropdownOpen] = useState(false);
//   if (!memberId) return null;
//   return (
//     <>
//       <DropdownMenu open={dropdownOpen} onOpenChange={setDropdownOpen}>
//         <DropdownMenuTrigger asChild>
//           <Button variant="ghost" size="icon">
//             <Ellipsis className="h-4 w-4" />
//           </Button>
//         </DropdownMenuTrigger>
//         <DropdownMenuContent>
//           <DropdownMenuItem
//             onSelect={() => {
//               setDropdownOpen(false);
//               // navigate(`${RouteMap.ADMIN_USERS}/${memberId}?tab=roles`);
//             }}
//           >
//             <Button variant="ghost" size="sm">
//               <Pencil className="h-4 w-4" />
//               <span>Cancel Invitation</span>
//             </Button>
//           </DropdownMenuItem>
//         </DropdownMenuContent>
//       </DropdownMenu>
//     </>
//   );
// }
