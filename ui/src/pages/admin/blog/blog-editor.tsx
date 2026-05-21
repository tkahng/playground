import { CenteredSpinner } from "@/components/centered-spinner";
import { Badge } from "@/components/ui/badge";
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
import { Textarea } from "@/components/ui/textarea";
import { useAuthProvider } from "@/hooks/use-auth-provider";
import {
  BlogPost,
  BlogTag,
  CreateBlogPostInput,
  UpdateBlogPostInput,
  archiveBlogPost,
  createBlogPost,
  createBlogTag,
  getBlogPost,
  listBlogTags,
  publishBlogPost,
  unpublishBlogPost,
  updateBlogPost,
} from "@/lib/blog-queries";
import { useNavigate } from "@tanstack/react-router";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { X } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

interface BlogEditorProps {
  postId?: string; // undefined = new post
}

export default function BlogEditor({ postId }: BlogEditorProps) {
  const { user } = useAuthProvider();
  const token = user?.tokens.access_token ?? "";
  const navigate = useNavigate();
  const qc = useQueryClient();
  const isEdit = !!postId;

  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [contentFormat, setContentFormat] = useState<"tiptap" | "markdown">("markdown");
  const [seoTitle, setSeoTitle] = useState("");
  const [seoDescription, setSeoDescription] = useState("");
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([]);
  const [newTagName, setNewTagName] = useState("");

  const { data: existingPost, isLoading: loadingPost } = useQuery({
    queryKey: ["admin-blog-post", postId],
    queryFn: () => getBlogPost(postId!, token),
    enabled: isEdit && !!token,
  });

  const { data: allTags = [] } = useQuery<BlogTag[]>({
    queryKey: ["blog-tags"],
    queryFn: listBlogTags,
  });

  useEffect(() => {
    if (existingPost) {
      setTitle(existingPost.title);
      setContent(existingPost.content);
      setContentFormat(existingPost.content_format);
      setSeoTitle(existingPost.seo_title ?? "");
      setSeoDescription(existingPost.seo_description ?? "");
      setSelectedTagIds((existingPost.tags ?? []).map((t) => t.id));
    }
  }, [existingPost]);

  const createMutation = useMutation({
    mutationFn: (input: CreateBlogPostInput) => createBlogPost(token, input),
    onSuccess: (post: BlogPost) => {
      toast.success("Post created");
      qc.invalidateQueries({ queryKey: ["admin-blog-posts"] });
      navigate({ to: "/admin/blog/$postId/edit", params: { postId: post.id } });
    },
    onError: () => toast.error("Failed to create post"),
  });

  const updateMutation = useMutation({
    mutationFn: (input: UpdateBlogPostInput) => updateBlogPost(token, postId!, input),
    onSuccess: () => {
      toast.success("Saved");
      qc.invalidateQueries({ queryKey: ["admin-blog-posts"] });
      qc.invalidateQueries({ queryKey: ["admin-blog-post", postId] });
    },
    onError: () => toast.error("Failed to save"),
  });

  const publishMutation = useMutation({
    mutationFn: () => publishBlogPost(token, postId!),
    onSuccess: () => {
      toast.success("Published");
      qc.invalidateQueries({ queryKey: ["admin-blog-post", postId] });
    },
    onError: () => toast.error("Failed to publish"),
  });

  const unpublishMutation = useMutation({
    mutationFn: () => unpublishBlogPost(token, postId!),
    onSuccess: () => {
      toast.success("Unpublished");
      qc.invalidateQueries({ queryKey: ["admin-blog-post", postId] });
      qc.invalidateQueries({ queryKey: ["admin-blog-posts"] });
    },
    onError: () => toast.error("Failed to unpublish"),
  });

  const archiveMutation = useMutation({
    mutationFn: () => archiveBlogPost(token, postId!),
    onSuccess: () => {
      toast.success("Archived");
      qc.invalidateQueries({ queryKey: ["admin-blog-post", postId] });
      qc.invalidateQueries({ queryKey: ["admin-blog-posts"] });
    },
    onError: () => toast.error("Failed to archive"),
  });

  const createTagMutation = useMutation({
    mutationFn: (name: string) => createBlogTag(token, { name }),
    onSuccess: (tag: BlogTag) => {
      toast.success(`Tag "${tag.name}" created`);
      setNewTagName("");
      setSelectedTagIds((ids) => [...ids, tag.id]);
      qc.invalidateQueries({ queryKey: ["blog-tags"] });
    },
    onError: () => toast.error("Failed to create tag"),
  });

  const handleSave = () => {
    const input = {
      title,
      content,
      content_format: contentFormat,
      seo_title: seoTitle || undefined,
      seo_description: seoDescription || undefined,
      tag_ids: selectedTagIds,
    };
    if (isEdit) {
      updateMutation.mutate(input);
    } else {
      createMutation.mutate(input);
    }
  };

  const toggleTag = (tagId: string) => {
    setSelectedTagIds((ids) =>
      ids.includes(tagId) ? ids.filter((id) => id !== tagId) : [...ids, tagId],
    );
  };

  if (isEdit && loadingPost) return <CenteredSpinner />;

  const isSaving = createMutation.isPending || updateMutation.isPending;
  const status = existingPost?.status;
  const isPublished = status === "published";
  const isArchived = status === "archived";

  return (
    <div className="space-y-6 max-w-3xl">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">
          {isEdit ? "Edit Post" : "New Post"}
        </h2>
        <div className="flex gap-2">
          {isEdit && !isPublished && !isArchived && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => publishMutation.mutate()}
              disabled={publishMutation.isPending}
            >
              Publish
            </Button>
          )}
          {isEdit && isPublished && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => unpublishMutation.mutate()}
              disabled={unpublishMutation.isPending}
            >
              Unpublish
            </Button>
          )}
          {isEdit && !isArchived && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => archiveMutation.mutate()}
              disabled={archiveMutation.isPending}
            >
              Archive
            </Button>
          )}
          <Button size="sm" onClick={handleSave} disabled={isSaving}>
            {isSaving ? "Saving…" : "Save"}
          </Button>
        </div>
      </div>

      {isEdit && existingPost && (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <span>Status:</span>
          <Badge
            variant={
              existingPost.status === "published"
                ? "default"
                : existingPost.status === "archived"
                  ? "outline"
                  : "secondary"
            }
          >
            {existingPost.status}
          </Badge>
          {existingPost.slug && (
            <>
              <span>·</span>
              <span className="font-mono text-xs">{existingPost.slug}</span>
            </>
          )}
        </div>
      )}

      <div className="space-y-2">
        <Label htmlFor="title">Title</Label>
        <Input
          id="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Post title"
        />
      </div>

      {isEdit && existingPost?.slug && (
        <div className="space-y-1">
          <Label htmlFor="slug" className="text-muted-foreground text-xs">
            Slug (read-only — set on creation)
          </Label>
          <Input
            id="slug"
            value={existingPost.slug}
            readOnly
            className="font-mono text-sm bg-muted text-muted-foreground cursor-default"
          />
        </div>
      )}

      <div className="space-y-2">
        <div className="flex items-center justify-between">
          <Label htmlFor="content">Content</Label>
          <Select
            value={contentFormat}
            onValueChange={(v) => setContentFormat(v as "tiptap" | "markdown")}
          >
            <SelectTrigger className="w-36 h-7 text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="markdown">Markdown</SelectItem>
              <SelectItem value="tiptap">Tiptap JSON</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <Textarea
          id="content"
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder={
            contentFormat === "markdown"
              ? "Write your post in Markdown…"
              : "Paste Tiptap JSON here (or integrate @tiptap/react editor)…"
          }
          className="min-h-[400px] font-mono text-sm"
        />
        <p className="text-xs text-muted-foreground">
          {contentFormat === "markdown"
            ? "Plain Markdown. Install react-markdown for rendered preview."
            : "Tiptap JSON. Install @tiptap/react for a rich editor experience."}
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="seo-title">SEO Title</Label>
          <Input
            id="seo-title"
            value={seoTitle}
            onChange={(e) => setSeoTitle(e.target.value)}
            placeholder="Optional — defaults to post title"
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="seo-desc">SEO Description</Label>
          <Input
            id="seo-desc"
            value={seoDescription}
            onChange={(e) => setSeoDescription(e.target.value)}
            placeholder="Brief description for search engines"
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label>Tags</Label>
        <div className="flex flex-wrap gap-2 mb-2">
          {allTags.map((tag) => (
            <button
              key={tag.id}
              type="button"
              onClick={() => toggleTag(tag.id)}
              className="focus:outline-none"
            >
              <Badge
                variant={selectedTagIds.includes(tag.id) ? "default" : "outline"}
                className="cursor-pointer"
              >
                {tag.name}
                {selectedTagIds.includes(tag.id) && (
                  <X className="ml-1 h-3 w-3" />
                )}
              </Badge>
            </button>
          ))}
        </div>
        <div className="flex gap-2">
          <Input
            placeholder="New tag name"
            value={newTagName}
            onChange={(e) => setNewTagName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && newTagName.trim()) {
                e.preventDefault();
                createTagMutation.mutate(newTagName.trim());
              }
            }}
            className="max-w-xs"
          />
          <Button
            variant="outline"
            size="sm"
            onClick={() => newTagName.trim() && createTagMutation.mutate(newTagName.trim())}
            disabled={createTagMutation.isPending || !newTagName.trim()}
          >
            Add Tag
          </Button>
        </div>
      </div>
    </div>
  );
}
