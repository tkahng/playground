import { CenteredSpinner } from "@/components/centered-spinner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
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
import { toast } from "sonner";

const statusVariant: Record<string, "default" | "secondary" | "outline"> = {
  published: "default",
  draft: "secondary",
  archived: "outline",
};

export default function AdminBlogListPage() {
  const { user } = useAuthProvider();
  const token = user?.tokens.access_token ?? "";
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ["admin-blog-posts"],
    queryFn: () => listBlogPosts({ per_page: 50 }, token),
    enabled: !!token,
  });

  const invalidate = () => qc.invalidateQueries({ queryKey: ["admin-blog-posts"] });

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

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">
          Manage blog posts. Admins only.
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
              <TableCell colSpan={5} className="text-center text-muted-foreground py-8">
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
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon">
                      <MoreHorizontal className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem asChild>
                      <Link to="/admin/blog/$postId/edit" params={{ postId: post.id }}>
                        Edit
                      </Link>
                    </DropdownMenuItem>
                    {post.status !== "published" && (
                      <DropdownMenuItem onClick={() => publishMutation.mutate(post.id)}>
                        Publish
                      </DropdownMenuItem>
                    )}
                    {post.status === "published" && (
                      <DropdownMenuItem onClick={() => unpublishMutation.mutate(post.id)}>
                        Unpublish
                      </DropdownMenuItem>
                    )}
                    {post.status !== "archived" && (
                      <DropdownMenuItem onClick={() => archiveMutation.mutate(post.id)}>
                        Archive
                      </DropdownMenuItem>
                    )}
                    <DropdownMenuItem
                      className="text-destructive"
                      onClick={() => {
                        if (confirm("Delete this post?")) {
                          deleteMutation.mutate(post.id);
                        }
                      }}
                    >
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
