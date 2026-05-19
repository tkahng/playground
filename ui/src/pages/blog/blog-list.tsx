import { CenteredSpinner } from "@/components/centered-spinner";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { useSearchParams } from "@/hooks/use-search-params";
import { listBlogPosts, BlogPost } from "@/lib/blog-queries";
import { Link } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { Calendar, Clock } from "lucide-react";
import { useState } from "react";

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

function PostCard({ post }: { post: BlogPost }) {
  return (
    <article className="group border rounded-xl p-6 hover:shadow-md transition-shadow bg-card">
      <div className="flex items-center gap-2 text-sm text-muted-foreground mb-3">
        <Calendar className="h-4 w-4" />
        <span>{post.published_at ? formatDate(post.published_at) : "—"}</span>
        {post.reading_time_minutes && (
          <>
            <span>·</span>
            <Clock className="h-4 w-4" />
            <span>{post.reading_time_minutes} min read</span>
          </>
        )}
      </div>
      <Link to="/blog/$slug" params={{ slug: post.slug }}>
        <h2 className="text-xl font-semibold mb-2 group-hover:text-primary transition-colors">
          {post.title}
        </h2>
      </Link>
      {post.seo_description && (
        <p className="text-muted-foreground text-sm line-clamp-3 mb-4">
          {post.seo_description}
        </p>
      )}
      {post.tags && post.tags.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {post.tags.map((tag) => (
            <Badge key={tag.id} variant="secondary">
              {tag.name}
            </Badge>
          ))}
        </div>
      )}
    </article>
  );
}

export default function BlogListPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const pageIndex = parseInt(searchParams.get("page") || "0", 10);
  const perPage = parseInt(searchParams.get("per_page") || "10", 10);
  const [q, setQ] = useState(searchParams.get("q") || "");

  const { data, isLoading, isError } = useQuery({
    queryKey: ["blog-posts", pageIndex, perPage, q],
    queryFn: () =>
      listBlogPosts({
        page: pageIndex,
        per_page: perPage,
        sort_by: "published_at",
        sort_order: "desc",
        q: q || undefined,
      }),
  });

  const handleSearch = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setSearchParams({ q, page: "0" });
  };

  if (isLoading) return <CenteredSpinner />;
  if (isError) return <div className="text-destructive">Failed to load posts.</div>;

  const posts = data?.data ?? [];
  const meta = data?.meta;

  return (
    <div className="max-w-3xl mx-auto py-12 px-4 space-y-8">
      <div>
        <h1 className="text-4xl font-bold tracking-tight">Blog</h1>
        <p className="text-muted-foreground mt-2">
          Articles, tutorials, and updates from the team.
        </p>
      </div>

      <form onSubmit={handleSearch} className="flex gap-2">
        <Input
          placeholder="Search posts…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          className="max-w-sm"
        />
      </form>

      {posts.length === 0 ? (
        <p className="text-muted-foreground">No posts found.</p>
      ) : (
        <div className="space-y-6">
          {posts.map((post) => (
            <PostCard key={post.id} post={post} />
          ))}
        </div>
      )}

      {meta && (meta.prev_page !== null || meta.has_more) && (
        <div className="flex justify-between">
          {meta.prev_page !== null ? (
            <button
              className="text-sm underline"
              onClick={() =>
                setSearchParams({ page: String(meta.prev_page) })
              }
            >
              ← Previous
            </button>
          ) : (
            <span />
          )}
          {meta.has_more && (
            <button
              className="text-sm underline"
              onClick={() =>
                setSearchParams({ page: String(meta.next_page) })
              }
            >
              Next →
            </button>
          )}
        </div>
      )}
    </div>
  );
}
