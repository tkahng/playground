import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useState } from "react";
import { CarCard, Task } from "./CarCard";

export function DialogWrapper({ task }: { task: Task }) {
  const [open, setOpen] = useState(false);

  const handleDoubleClick = () => {
    setOpen(true);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <div onDoubleClick={handleDoubleClick}>
          <CarCard task={task} />
        </div>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Dialog Title</DialogTitle>
          <DialogDescription>This is the dialog content.</DialogDescription>
        </DialogHeader>
      </DialogContent>
    </Dialog>
  );
}
