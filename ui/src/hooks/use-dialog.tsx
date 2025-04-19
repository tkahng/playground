import { Dialog, DialogContent } from "@/components/ui/dialog";
import { PropsWithChildren, useCallback, useState } from "react";

export function useDialog() {
  const [open, onOpenChange] = useState(false);

  const trigger = useCallback(() => {
    onOpenChange(true);
  }, [onOpenChange]);

  return { props: { open, onOpenChange }, trigger };
}

export type DialogProps = {
  open: boolean;
  onOpenChange: React.Dispatch<React.SetStateAction<boolean>>;
};

export function ConfirmDialog({
  dialogProps,
  // onClick,
  children,
}: PropsWithChildren<{
  dialogProps: DialogProps;
  // onClick: React.MouseEventHandler<HTMLButtonElement>;
}>) {
  return (
    <Dialog {...dialogProps}>
      {" "}
      {/* This will contain the open and onOpenChange props */}
      <DialogContent>
        {children}
        {/* <DialogHeader>
          <DialogTitle>Are you absolutely sure?</DialogTitle>
        </DialogHeader> */}
        {/* Dialog Content */}
        {/* <DialogDescription>This action cannot be undone.</DialogDescription> */}
        {/* <DialogFooter>
          <Button
            variant="outline"
            onClick={() => dialogProps.onOpenChange(false)}
          >
            Cancel
          </Button>
          <Button variant="destructive" onClick={onClick}>
            Delete
          </Button>
        </DialogFooter> */}
      </DialogContent>
    </Dialog>
  );
}
