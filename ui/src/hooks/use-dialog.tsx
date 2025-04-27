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
  children,
}: PropsWithChildren<{
  dialogProps: DialogProps;
}>) {
  return (
    <Dialog {...dialogProps}>
      {/* This will contain the open and onOpenChange props */}
      <DialogContent>{children}</DialogContent>
    </Dialog>
  );
}
