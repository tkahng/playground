import ReactMarkdown from "react-markdown";
import { cn } from "@/lib/utils";

// Transforms @[Display Name](uuid) mention syntax into highlighted spans
// before passing content to the markdown renderer.
function renderMentions(content: string): string {
  return content.replace(/@\[([^\]]+)\]\([0-9a-f-]{36}\)/g, "**@$1**");
}

interface MentionTextProps {
  content: string;
  className?: string;
}

export function MentionText({ content, className }: MentionTextProps) {
  const rendered = renderMentions(content);
  return (
    <div className={cn("prose prose-sm dark:prose-invert max-w-none", className)}>
      <ReactMarkdown>{rendered}</ReactMarkdown>
    </div>
  );
}
