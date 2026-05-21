import { CenteredSpinner } from "@/components/centered-spinner";
import { MarkdownRenderer, TiptapRenderer } from "@/components/blog/content-renderer";
import { Badge } from "@/components/ui/badge";
import { getBlogPost } from "@/lib/blog-queries";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { ArrowLeft, Calendar, Clock } from "lucide-react";
import { Helmet } from "react-helmet-async";

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

export default function BlogPostPage({ slug }: { slug: string }) {
  const { data: post, isLoading, isError } = useQuery({
    queryKey: ["blog-post", slug],
    queryFn: () => getBlogPost(slug),
    retry: false,
  });

  if (isLoading) return <CenteredSpinner />;
  if (isError || !post) {
    return (
      <div className="max-w-3xl mx-auto py-12 px-4 text-center">
        <h1 className="text-2xl font-bold mb-4">Post not found</h1>
        <Link to="/blog" className="text-primary underline">
          ← Back to blog
        </Link>
      </div>
    );
  }

  const metaTitle = post.seo_title ?? post.title;
  const metaDescription = post.seo_description ?? undefined;

  return (
    <article className="max-w-3xl mx-auto py-12 px-4">
      <Helmet>
        <title>{metaTitle}</title>
        {metaDescription && <meta name="description" content={metaDescription} />}
        <meta property="og:title" content={metaTitle} />
        {metaDescription && <meta property="og:description" content={metaDescription} />}
        <meta property="og:type" content="article" />
        {post.published_at && (
          <meta property="article:published_time" content={post.published_at} />
        )}
      </Helmet>
      <Link
        to="/blog"
        className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-8 w-fit"
      >
        <ArrowLeft className="h-4 w-4" />
        Back to blog
      </Link>

      <header className="mb-8">
        {post.tags && post.tags.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-4">
            {post.tags.map((tag) => (
              <Badge key={tag.id} variant="secondary">
                {tag.name}
              </Badge>
            ))}
          </div>
        )}
        <h1 className="text-4xl font-bold tracking-tight mb-4">{post.title}</h1>
        <div className="flex items-center gap-3 text-sm text-muted-foreground">
          {post.published_at && (
            <>
              <Calendar className="h-4 w-4" />
              <span>{formatDate(post.published_at)}</span>
            </>
          )}
          {post.reading_time_minutes && (
            <>
              <span>·</span>
              <Clock className="h-4 w-4" />
              <span>{post.reading_time_minutes} min read</span>
            </>
          )}
        </div>
      </header>

      <div className="prose prose-neutral dark:prose-invert max-w-none">
        {post.content_format === "markdown" ? (
          <MarkdownRenderer content={post.content} />
        ) : (
          <TiptapRenderer content={post.content} />
        )}
      </div>
    </article>
  );
}

