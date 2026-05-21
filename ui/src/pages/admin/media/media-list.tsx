import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { CenteredSpinner } from "@/components/centered-spinner";
import { Input } from "@/components/ui/input";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useSearchParams } from "@/hooks/use-search-params";
import {
  deleteMedia,
  listMedia,
  updateMediaAltText,
  uploadMedia,
  type MediaItem,
} from "@/lib/media-queries";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ImageIcon, Trash2, Upload } from "lucide-react";
import { useRef, useState } from "react";
import { toast } from "sonner";

function formatBytes(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function MediaCard({
  item,
  token,
  onDeleted,
}: {
  item: MediaItem;
  token: string;
  onDeleted: () => void;
}) {
  const qc = useQueryClient();
  const [altText, setAltText] = useState(item.alt_text ?? "");
  const [editing, setEditing] = useState(false);

  const altTextMutation = useMutation({
    mutationFn: (text: string) => updateMediaAltText(token, item.id, text || null),
    onSuccess: () => {
      toast.success("Alt text saved");
      setEditing(false);
      qc.invalidateQueries({ queryKey: ["admin-media"] });
    },
    onError: () => toast.error("Failed to save alt text"),
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteMedia(token, item.id),
    onSuccess: () => {
      toast.success("Deleted");
      onDeleted();
    },
    onError: () => toast.error("Failed to delete"),
  });

  const isImage = item.mime_type.startsWith("image/");

  return (
    <div className="border rounded-lg overflow-hidden bg-card">
      <div className="aspect-video bg-muted flex items-center justify-center overflow-hidden">
        {isImage ? (
          <img
            src={item.url}
            alt={item.alt_text ?? item.original_filename}
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="flex flex-col items-center gap-1 text-muted-foreground">
            <ImageIcon className="h-8 w-8" />
            <span className="text-xs">{item.mime_type}</span>
          </div>
        )}
      </div>
      <div className="p-3 space-y-2">
        <p className="text-sm font-medium truncate" title={item.original_filename}>
          {item.original_filename}
        </p>
        <p className="text-xs text-muted-foreground">
          {formatBytes(item.size)}
          {item.width && item.height && ` · ${item.width}×${item.height}`}
        </p>

        {editing ? (
          <div className="flex gap-1">
            <Input
              value={altText}
              onChange={(e) => setAltText(e.target.value)}
              placeholder="Alt text…"
              className="h-7 text-xs"
              autoFocus
              onKeyDown={(e) => {
                if (e.key === "Enter") altTextMutation.mutate(altText);
                if (e.key === "Escape") setEditing(false);
              }}
            />
            <Button
              size="sm"
              className="h-7 px-2 text-xs"
              onClick={() => altTextMutation.mutate(altText)}
              disabled={altTextMutation.isPending}
            >
              Save
            </Button>
          </div>
        ) : (
          <button
            type="button"
            className="text-xs text-muted-foreground hover:text-foreground w-full text-left truncate"
            onClick={() => setEditing(true)}
            aria-label="Edit alt text"
          >
            {item.alt_text ? item.alt_text : <em>Add alt text…</em>}
          </button>
        )}

        <div className="flex justify-end">
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-7 w-7 text-destructive hover:text-destructive"
                aria-label="Delete media"
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Delete "{item.original_filename}"?</AlertDialogTitle>
                <AlertDialogDescription>
                  This removes the media record. Any posts referencing this image
                  will lose the link.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction
                  className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                  onClick={() => deleteMutation.mutate()}
                >
                  Delete
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </div>
      </div>
    </div>
  );
}

export default function AdminMediaListPage() {
  const { user } = useAuthProvider();
  const token = user?.tokens.access_token ?? "";
  const qc = useQueryClient();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [searchParams, setSearchParams] = useSearchParams();

  const page = parseInt(searchParams.get("page") || "0", 10);
  const perPage = 20;

  const { data, isLoading } = useQuery({
    queryKey: ["admin-media", page, perPage],
    queryFn: () => listMedia(token, { page, per_page: perPage }),
    enabled: !!token,
  });

  const uploadMutation = useMutation({
    mutationFn: (files: File[]) => uploadMedia(token, files),
    onSuccess: () => {
      toast.success("Uploaded");
      qc.invalidateQueries({ queryKey: ["admin-media"] });
    },
    onError: () => toast.error("Upload failed"),
  });

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files ?? []);
    if (files.length > 0) uploadMutation.mutate(files);
    e.target.value = "";
  };

  const invalidate = () => qc.invalidateQueries({ queryKey: ["admin-media"] });

  if (isLoading) return <CenteredSpinner />;

  const items: MediaItem[] = data?.data ?? [];
  const total = data?.meta.total ?? 0;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">
          {total} file{total !== 1 ? "s" : ""}
        </p>
        <Button
          size="sm"
          onClick={() => fileInputRef.current?.click()}
          disabled={uploadMutation.isPending}
        >
          <Upload className="h-4 w-4 mr-1" />
          {uploadMutation.isPending ? "Uploading…" : "Upload"}
        </Button>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*,video/*,application/pdf"
          multiple
          className="hidden"
          onChange={handleFileChange}
        />
      </div>

      {items.length === 0 ? (
        <div className="border rounded-xl py-16 text-center text-muted-foreground">
          <ImageIcon className="h-10 w-10 mx-auto mb-3 opacity-40" />
          <p>No media yet. Upload your first file.</p>
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
          {items.map((item) => (
            <MediaCard
              key={item.id}
              item={item}
              token={token}
              onDeleted={invalidate}
            />
          ))}
        </div>
      )}

      {total > perPage && (
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">
            Page {page + 1} of {Math.ceil(total / perPage)}
          </span>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={page === 0}
              onClick={() => setSearchParams({ page: String(page - 1) })}
            >
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={(page + 1) * perPage >= total}
              onClick={() => setSearchParams({ page: String(page + 1) })}
            >
              Next
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
