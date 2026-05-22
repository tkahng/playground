import ReactMarkdown from "react-markdown";

// ── Markdown ──────────────────────────────────────────────────────────────────

export function MarkdownRenderer({ content }: { content: string }) {
  return <ReactMarkdown>{content}</ReactMarkdown>;
}

// ── Tiptap ────────────────────────────────────────────────────────────────────

export interface TiptapNode {
  type: string;
  attrs?: Record<string, unknown>;
  content?: TiptapNode[];
  text?: string;
  marks?: TiptapMark[];
}

export interface TiptapMark {
  type: string;
  attrs?: Record<string, unknown>;
}

export interface TiptapDoc {
  type: "doc";
  content: TiptapNode[];
}

export function TiptapRenderer({ content }: { content: string }) {
  let parsed: unknown;
  try {
    parsed = JSON.parse(content);
  } catch {
    return <p>{content}</p>;
  }

  const doc = parsed as TiptapDoc;
  if (!doc || doc.type !== "doc" || !Array.isArray(doc.content)) {
    return <p>{content}</p>;
  }

  return (
    <>
      {doc.content.map((node, i) => (
        <TiptapNodeRenderer key={i} node={node} />
      ))}
    </>
  );
}

function applyMarks(text: string, marks: TiptapMark[] = []): React.ReactNode {
  let node: React.ReactNode = text;
  for (const mark of marks) {
    switch (mark.type) {
      case "bold":
        node = <strong>{node}</strong>;
        break;
      case "italic":
        node = <em>{node}</em>;
        break;
      case "code":
        node = <code>{node}</code>;
        break;
      case "strike":
        node = <s>{node}</s>;
        break;
      case "underline":
        node = <u>{node}</u>;
        break;
      case "link": {
        const href = (mark.attrs?.href as string) ?? "#";
        node = (
          <a href={href} target="_blank" rel="noopener noreferrer">
            {node}
          </a>
        );
        break;
      }
    }
  }
  return node;
}

function renderChildren(nodes: TiptapNode[] = []): React.ReactNode {
  return nodes.map((child, i) => <TiptapNodeRenderer key={i} node={child} />);
}

export function TiptapNodeRenderer({ node }: { node: TiptapNode }) {
  switch (node.type) {
    case "paragraph":
      return <p>{renderChildren(node.content)}</p>;

    case "heading": {
      const level = (node.attrs?.level as number) ?? 2;
      const children = renderChildren(node.content);
      switch (level) {
        case 1: return <h1>{children}</h1>;
        case 2: return <h2>{children}</h2>;
        case 3: return <h3>{children}</h3>;
        case 4: return <h4>{children}</h4>;
        case 5: return <h5>{children}</h5>;
        default: return <h6>{children}</h6>;
      }
    }

    case "bulletList":
      return <ul>{renderChildren(node.content)}</ul>;

    case "orderedList":
      return <ol>{renderChildren(node.content)}</ol>;

    case "listItem":
      return <li>{renderChildren(node.content)}</li>;

    case "blockquote":
      return <blockquote>{renderChildren(node.content)}</blockquote>;

    case "codeBlock": {
      const lang = node.attrs?.language as string | undefined;
      return (
        <pre data-language={lang}>
          <code>{renderChildren(node.content)}</code>
        </pre>
      );
    }

    case "horizontalRule":
      return <hr />;

    case "hardBreak":
      return <br />;

    case "image": {
      const src = node.attrs?.src as string | undefined;
      if (!src) return null;
      const alt = node.attrs?.alt as string | undefined;
      const title = node.attrs?.title as string | undefined;
      return <img src={src} alt={alt ?? ""} title={title} />;
    }

    case "text":
      return <>{applyMarks(node.text ?? "", node.marks)}</>;

    default:
      return <>{renderChildren(node.content)}</>;
  }
}
