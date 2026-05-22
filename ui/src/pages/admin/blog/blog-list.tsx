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
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CenteredSpinner } from "@/components/centered-spinner";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import { useSearchParams } from "@/hooks/use-search-params";
import {
  archiveBlogPost,
  deleteBlogPost,
  BlogPost,
  listBlogPosts,
  publishBlogPost,
  unpublishBlogPost,
} from "@/lib/blog-queries";
import { Link } from "@tanstack/react-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { MoreHorizontal, Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

const statusVariant: Record<string, "default" | "secondary" | "outline"> = {
  published: "default",
  draft: "secondary",
  archived: "outline",
};

function DeletePostDialog({ post, onConfirm }: { post: BlogPost; onConfirm: () => void }) {
  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <DropdownMenuItem
          className="text-destructive focus:text-destructive"
          onSelect={(e) => e.preventDefault()}
        >
          Delete
        </DropdownMenuItem>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete "{post.title}"?</AlertDialogTitle>
          <AlertDialogDescription>
            This permanently removes the post and cannot be undone. Consider
            archiving instead to hide it from readers.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            onClick={onConfirm}
          >
            Delete
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

export default function AdminBlogListPage() {
  const { user } = useAuthProvider();
  const token = user?.tokens.access_token ?? "";
  const qc = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();

  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const pageSize = parseInt(searchParams.get("per_page") || "10", 10);

  const [openMenuPostId, setOpenMenuPostId] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["admin-blog-posts", pageIndex, pageSize],
    queryFn: () => listBlogPosts({ page: pageIndex, per_page: pageSize }, token),
    enabled: !!token,
  });

  const invalidate = () =>
    qc.invalidateQueries({ queryKey: ["admin-blog-posts"] });

  const publishMutation = useMutation({
    mutationFn: (id: string) => publishBlogPost(token, id),
    onSuccess: () => { toast.success("Published"); invalidate(); },
    onError: () => toast.error("Failed to publish"),
  });

  const unpublishMutation = useMutation({
    mutationFn: (id: string) => unpublishBlogPost(token, id),
    onSuccess: () => { toast.success("Unpublished"); invalidate(); },
    onError: () => toast.error("Failed to unpublish"),
  });

  const archiveMutation = useMutation({
    mutationFn: (id: string) => archiveBlogPost(token, id),
    onSuccess: () => { toast.success("Archived"); invalidate(); },
    onError: () => toast.error("Failed to archive"),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteBlogPost(token, id),
    onSuccess: () => { toast.success("Deleted"); invalidate(); },
    onError: () => toast.error("Failed to delete"),
  });

  if (isLoading) return <CenteredSpinner />;

  const posts: BlogPost[] = data?.data ?? [];
  const total = data?.meta.total ?? 0;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">
          {total} post{total !== 1 ? "s" : ""} total
        </p>
        <Button asChild size="sm">
          <Link to="/admin/blog/new">
            <Plus className="h-4 w-4 mr-1" />
            New Post
          </Link>
        </Button>
      </div>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Title</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Published</TableHead>
            <TableHead>Views</TableHead>
            <TableHead className="w-12" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {posts.length === 0 && (
            <TableRow>
              <TableCell
                colSpan={5}
                className="text-center text-muted-foreground py-8"
              >
                No posts yet.
              </TableCell>
            </TableRow>
          )}
          {posts.map((post) => (
            <TableRow key={post.id}>
              <TableCell>
                <Link
                  to="/admin/blog/$postId/edit"
                  params={{ postId: post.id }}
                  className="font-medium hover:underline"
                >
                  {post.title}
                </Link>
              </TableCell>
              <TableCell>
                <Badge variant={statusVariant[post.status] ?? "secondary"}>
                  {post.status}
                </Badge>
              </TableCell>
              <TableCell className="text-muted-foreground text-sm">
                {post.published_at
                  ? new Date(post.published_at).toLocaleDateString()
                  : "—"}
              </TableCell>
              <TableCell className="text-muted-foreground text-sm">
                {post.view_count}
              </TableCell>
              <TableCell>
                <DropdownMenu
                  open={openMenuPostId === post.id}
                  onOpenChange={(open) =>
                    setOpenMenuPostId(open ? post.id : null)
                  }
                >
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem asChild>
                      <Link
                        to="/admin/blog/$postId/edit"
                        params={{ postId: post.id }}
                      >
                        Edit
                      </Link>
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    {post.status !== "published" && (
                      <DropdownMenuItem
                        onClick={() => publishMutation.mutate(post.id)}
                      >
                        Publish
                      </DropdownMenuItem>
                    )}
                    {post.status === "published" && (
                      <DropdownMenuItem
                        onClick={() => unpublishMutation.mutate(post.id)}
                      >
                        Unpublish
                      </DropdownMenuItem>
                    )}
                    {post.status !== "archived" && (
                      <DropdownMenuItem
                        onClick={() => archiveMutation.mutate(post.id)}
                      >
                        Archive
                      </DropdownMenuItem>
                    )}
                    <DropdownMenuSeparator />
                    <DeletePostDialog
                      post={post}
                      onConfirm={() => deleteMutation.mutate(post.id)}
                    />
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      {total > pageSize && (
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">
            Page {pageIndex + 1} of {Math.ceil(total / pageSize)}
          </span>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={pageIndex === 0}
              onClick={() =>
                setSearchParams({ page: String(pageIndex - 1), per_page: String(pageSize) })
              }
            >
              Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={(pageIndex + 1) * pageSize >= total}
              onClick={() =>
                setSearchParams({ page: String(pageIndex + 1), per_page: String(pageSize) })
              }
            >
              Next
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
