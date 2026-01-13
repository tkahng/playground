import { Button } from "@/components/ui/button";
import {
  DialogClose,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ConfirmDialog, useDialog } from "@/hooks/use-dialog";
import { Trash } from "lucide-react";

export function RoleDeleteButton({
  permissionId,
  onDelete,
}: {
  permissionId: string;
  onDelete: (permissionId: string) => void;
}) {
  const editDialog = useDialog();
  return (
    <>
      <Button variant="outline" size="icon" onClick={editDialog.trigger}>
        <Trash className="h-4 w-4" />
      </Button>
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
                  onDelete(permissionId);
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
