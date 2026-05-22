import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { listMedia, uploadMedia, type MediaItem } from "@/lib/media-queries";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ImageIcon, Upload } from "lucide-react";
import { useRef, useState } from "react";
import { toast } from "sonner";

interface MediaPickerDialogProps {
  onSelect: (item: MediaItem) => void;
  trigger?: React.ReactNode;
}

export function MediaPickerDialog({ onSelect, trigger }: MediaPickerDialogProps) {
  const { user } = useAuthProvider();
  const token = user?.tokens.access_token ?? "";
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  const [q, setQ] = useState("");
  const [page, setPage] = useState(0);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["media-picker", page, q],
    queryFn: () => listMedia(token, { page, per_page: 20, q: q || undefined }),
    enabled: open && !!token,
  });

  const uploadMutation = useMutation({
    mutationFn: (files: File[]) => uploadMedia(token, files),
    onSuccess: () => {
      toast.success("Uploaded");
      qc.invalidateQueries({ queryKey: ["media-picker"] });
      qc.invalidateQueries({ queryKey: ["admin-media"] });
    },
    onError: () => toast.error("Upload failed"),
  });

  const handleSelect = (item: MediaItem) => {
    onSelect(item);
    setOpen(false);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files ?? []);
    if (files.length > 0) uploadMutation.mutate(files);
    e.target.value = "";
  };

  const items: MediaItem[] = data?.data ?? [];
  const total = data?.meta.total ?? 0;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {trigger ?? (
          <Button variant="outline" size="sm" type="button">
            <ImageIcon className="h-4 w-4 mr-1" />
            Choose image
          </Button>
        )}
      </DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Select media</DialogTitle>
        </DialogHeader>

        <div className="flex items-center gap-2 mb-3">
          <Input
            placeholder="Search…"
            value={q}
            onChange={(e) => { setQ(e.target.value); setPage(0); }}
            className="flex-1"
          />
          <Button
            variant="outline"
            size="sm"
            onClick={() => fileInputRef.current?.click()}
            disabled={uploadMutation.isPending}
          >
            <Upload className="h-4 w-4 mr-1" />
            Upload
          </Button>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            multiple
            className="hidden"
            onChange={handleFileChange}
          />
        </div>

        {isLoading ? (
          <div className="py-8 text-center text-muted-foreground text-sm">Loading…</div>
        ) : items.length === 0 ? (
          <div className="py-8 text-center text-muted-foreground text-sm">
            No media found. Upload some images first.
          </div>
        ) : (
          <div className="grid grid-cols-3 sm:grid-cols-4 gap-3 max-h-96 overflow-y-auto pr-1">
            {items.map((item) => (
              <button
                key={item.id}
                type="button"
                className="rounded-lg overflow-hidden border-2 border-transparent hover:border-primary focus:outline-none focus:border-primary transition-colors aspect-square bg-muted"
                onClick={() => handleSelect(item)}
                aria-label={`Select ${item.original_filename}`}
              >
                {item.mime_type.startsWith("image/") ? (
                  <img
                    src={item.url}
                    alt={item.alt_text ?? item.original_filename}
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="w-full h-full flex items-center justify-center text-muted-foreground">
                    <ImageIcon className="h-6 w-6" />
                  </div>
                )}
              </button>
            ))}
          </div>
        )}

        {total > 20 && (
          <div className="flex justify-between text-sm mt-2">
            <Button variant="ghost" size="sm" disabled={page === 0} onClick={() => setPage((p) => p - 1)}>
              Previous
            </Button>
            <span className="text-muted-foreground self-center">
              {page + 1} / {Math.ceil(total / 20)}
            </span>
            <Button variant="ghost" size="sm" disabled={(page + 1) * 20 >= total} onClick={() => setPage((p) => p + 1)}>
              Next
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
