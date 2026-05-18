import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import {
  workflowStatusCreate,
  workflowStatusDelete,
  workflowStatusReorder,
  workflowStatusUpdate,
} from "@/lib/task-queries";
import { WorkflowStatus } from "@/schema.types";
import {
  DndContext,
  DragEndEvent,
  KeyboardSensor,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import {
  SortableContext,
  arrayMove,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { GripVertical, Pencil, Settings, Trash2, X } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

const CATEGORY_OPTIONS = [
  { value: "todo", label: "Todo" },
  { value: "in_progress", label: "In Progress" },
  { value: "done", label: "Done" },
] as const;

const PRESET_COLORS = [
  "#6366f1", "#8b5cf6", "#ec4899", "#ef4444",
  "#f97316", "#eab308", "#22c55e", "#14b8a6",
  "#3b82f6", "#64748b",
];

type Category = "todo" | "in_progress" | "done";

interface EditingStatus {
  id: string;
  name: string;
  color: string;
  category: Category;
}

function SortableStatusRow({
  status,
  onEdit,
  onDelete,
  isDeleting,
}: {
  status: WorkflowStatus;
  onEdit: (s: WorkflowStatus) => void;
  onDelete: (id: string) => void;
  isDeleting: boolean;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: status.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.4 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className="flex items-center gap-2 p-2 rounded border bg-background group"
    >
      <button
        {...attributes}
        {...listeners}
        className="cursor-grab text-muted-foreground hover:text-foreground p-1"
        aria-label="Drag to reorder"
      >
        <GripVertical className="h-4 w-4" />
      </button>
      <span
        className="w-3 h-3 rounded-full flex-shrink-0"
        style={{ backgroundColor: status.color ?? "#64748b" }}
      />
      <span className="flex-1 text-sm font-medium truncate">{status.name}</span>
      <span className="text-xs text-muted-foreground capitalize hidden group-hover:inline">
        {status.category.replace("_", " ")}
      </span>
      <button
        onClick={() => onEdit(status)}
        className="p-1 text-muted-foreground hover:text-foreground"
        aria-label="Edit"
      >
        <Pencil className="h-3.5 w-3.5" />
      </button>
      <button
        onClick={() => onDelete(status.id)}
        disabled={isDeleting}
        className="p-1 text-muted-foreground hover:text-destructive disabled:opacity-50"
        aria-label="Delete"
      >
        <Trash2 className="h-3.5 w-3.5" />
      </button>
    </div>
  );
}

export function WorkflowSettingsPanel({
  teamId,
  workflowId,
  statuses,
  queryKey,
}: {
  teamId: string;
  workflowId: string;
  statuses: WorkflowStatus[];
  queryKey: object;
}) {
  const { user } = useAuthProvider();
  const queryClient = useQueryClient();
  const [localStatuses, setLocalStatuses] = useState<WorkflowStatus[]>(statuses);
  const [editing, setEditing] = useState<EditingStatus | null>(null);
  const [newName, setNewName] = useState("");
  const [newColor, setNewColor] = useState(PRESET_COLORS[0]!);
  const [newCategory, setNewCategory] = useState<Category>("todo");

  useEffect(() => {
    setLocalStatuses(statuses);
  }, [statuses]);

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: [queryKey] });

  const createMutation = useMutation({
    mutationFn: async () => {
      if (!user?.tokens.access_token || !newName.trim()) return;
      await workflowStatusCreate(user.tokens.access_token, teamId, workflowId, {
        name: newName.trim(),
        color: newColor,
        category: newCategory,
      });
    },
    onSuccess: () => {
      setNewName("");
      invalidate();
      toast.success("Status created");
    },
    onError: (e) => toast.error(`Failed to create status: ${e.message}`),
  });

  const updateMutation = useMutation({
    mutationFn: async (s: EditingStatus) => {
      if (!user?.tokens.access_token) return;
      await workflowStatusUpdate(
        user.tokens.access_token,
        teamId,
        workflowId,
        s.id,
        { name: s.name, color: s.color, category: s.category },
      );
    },
    onSuccess: () => {
      setEditing(null);
      invalidate();
      toast.success("Status updated");
    },
    onError: (e) => toast.error(`Failed to update status: ${e.message}`),
  });

  const deleteMutation = useMutation({
    mutationFn: async (statusId: string) => {
      if (!user?.tokens.access_token) return;
      await workflowStatusDelete(
        user.tokens.access_token,
        teamId,
        workflowId,
        statusId,
      );
    },
    onSuccess: () => {
      invalidate();
      toast.success("Status deleted");
    },
    onError: (e) => toast.error(`Failed to delete status: ${e.message}`),
  });

  const reorderMutation = useMutation({
    mutationFn: async (ids: string[]) => {
      if (!user?.tokens.access_token) return;
      await workflowStatusReorder(
        user.tokens.access_token,
        teamId,
        workflowId,
        ids,
      );
    },
    onSuccess: () => invalidate(),
    onError: (e) => {
      setLocalStatuses(statuses);
      toast.error(`Failed to reorder: ${e.message}`);
    },
  });

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor),
  );

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const oldIndex = localStatuses.findIndex((s) => s.id === active.id);
    const newIndex = localStatuses.findIndex((s) => s.id === over.id);
    const reordered = arrayMove(localStatuses, oldIndex, newIndex);
    setLocalStatuses(reordered);
    reorderMutation.mutate(reordered.map((s) => s.id));
  }

  return (
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <Settings className="h-4 w-4" />
          Manage columns
        </Button>
      </SheetTrigger>
      <SheetContent className="w-[min(380px,100vw-2rem)] flex flex-col gap-4 overflow-y-auto">
        <SheetHeader>
          <SheetTitle>Workflow columns</SheetTitle>
        </SheetHeader>

        {/* Sortable status list */}
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          onDragEnd={handleDragEnd}
        >
          <SortableContext
            items={localStatuses.map((s) => s.id)}
            strategy={verticalListSortingStrategy}
          >
            <div className="flex flex-col gap-2">
              {localStatuses.map((s) => (
                <SortableStatusRow
                  key={s.id}
                  status={s}
                  onEdit={(s) =>
                    setEditing({
                      id: s.id,
                      name: s.name,
                      color: s.color ?? PRESET_COLORS[0]!,
                      category: s.category as Category,
                    })
                  }
                  onDelete={(id) => deleteMutation.mutate(id)}
                  isDeleting={deleteMutation.isPending}
                />
              ))}
            </div>
          </SortableContext>
        </DndContext>

        {/* Inline edit form */}
        {editing && (
          <div className="border rounded p-3 flex flex-col gap-3 bg-muted/40">
            <div className="flex items-center justify-between">
              <span className="text-sm font-semibold">Edit status</span>
              <button onClick={() => setEditing(null)}>
                <X className="h-4 w-4" />
              </button>
            </div>
            <div className="flex flex-col gap-1">
              <Label className="text-xs">Name</Label>
              <Input
                value={editing.name}
                onChange={(e) =>
                  setEditing({ ...editing, name: e.target.value })
                }
              />
            </div>
            <div className="flex flex-col gap-1">
              <Label className="text-xs">Category</Label>
              <Select
                value={editing.category}
                onValueChange={(v) =>
                  setEditing({ ...editing, category: v as Category })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {CATEGORY_OPTIONS.map((o) => (
                    <SelectItem key={o.value} value={o.value}>
                      {o.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-1">
              <Label className="text-xs">Color</Label>
              <div className="flex flex-wrap gap-2">
                {PRESET_COLORS.map((c) => (
                  <button
                    key={c}
                    onClick={() => setEditing({ ...editing, color: c })}
                    className="w-8 h-8 sm:w-6 sm:h-6 rounded-full border-2 transition-transform hover:scale-110"
                    style={{
                      backgroundColor: c,
                      borderColor:
                        editing.color === c ? "white" : "transparent",
                      outline: editing.color === c ? `2px solid ${c}` : "none",
                    }}
                  />
                ))}
                <input
                  type="color"
                  value={editing.color}
                  onChange={(e) =>
                    setEditing({ ...editing, color: e.target.value })
                  }
                  className="w-8 h-8 sm:w-6 sm:h-6 rounded cursor-pointer border p-0"
                  title="Custom color"
                />
              </div>
            </div>
            <Button
              size="sm"
              onClick={() => updateMutation.mutate(editing)}
              disabled={updateMutation.isPending || !editing.name.trim()}
            >
              Save
            </Button>
          </div>
        )}

        {/* Add new status */}
        <div className="border rounded p-3 flex flex-col gap-3">
          <span className="text-sm font-semibold">Add status</span>
          <div className="flex flex-col gap-1">
            <Label className="text-xs">Name</Label>
            <Input
              placeholder="e.g. Review"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") createMutation.mutate();
              }}
            />
          </div>
          <div className="flex flex-col gap-1">
            <Label className="text-xs">Category</Label>
            <Select
              value={newCategory}
              onValueChange={(v) => setNewCategory(v as Category)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {CATEGORY_OPTIONS.map((o) => (
                  <SelectItem key={o.value} value={o.value}>
                    {o.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex flex-col gap-1">
            <Label className="text-xs">Color</Label>
            <div className="flex flex-wrap gap-2">
              {PRESET_COLORS.map((c) => (
                <button
                  key={c}
                  onClick={() => setNewColor(c)}
                  className="w-8 h-8 sm:w-6 sm:h-6 rounded-full border-2 transition-transform hover:scale-110"
                  style={{
                    backgroundColor: c,
                    borderColor: newColor === c ? "white" : "transparent",
                    outline: newColor === c ? `2px solid ${c}` : "none",
                  }}
                />
              ))}
              <input
                type="color"
                value={newColor}
                onChange={(e) => setNewColor(e.target.value)}
                className="w-8 h-8 sm:w-6 sm:h-6 rounded cursor-pointer border p-0"
                title="Custom color"
              />
            </div>
          </div>
          <Button
            size="sm"
            onClick={() => createMutation.mutate()}
            disabled={createMutation.isPending || !newName.trim()}
          >
            Add status
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  );
}
