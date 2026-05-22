import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MarkdownRenderer, TiptapRenderer, TiptapNodeRenderer, type TiptapNode } from "../content-renderer";

// ── MarkdownRenderer ──────────────────────────────────────────────────────────

describe("MarkdownRenderer", () => {
  it("renders a heading", () => {
    render(<MarkdownRenderer content="# Hello World" />);
    expect(screen.getByRole("heading", { level: 1, name: "Hello World" })).toBeInTheDocument();
  });

  it("renders bold text", () => {
    const { container } = render(<MarkdownRenderer content="**bold**" />);
    expect(container.querySelector("strong")).toHaveTextContent("bold");
  });

  it("renders italic text", () => {
    const { container } = render(<MarkdownRenderer content="_italic_" />);
    expect(container.querySelector("em")).toHaveTextContent("italic");
  });

  it("renders a link", () => {
    const { container } = render(<MarkdownRenderer content="[click](https://example.com)" />);
    const a = container.querySelector("a");
    expect(a).toHaveAttribute("href", "https://example.com");
    expect(a).toHaveTextContent("click");
  });

  it("renders a code block", () => {
    const { container } = render(<MarkdownRenderer content={"```\nconst x = 1;\n```"} />);
    expect(container.querySelector("code")).toBeInTheDocument();
  });

  it("renders a paragraph", () => {
    const { container } = render(<MarkdownRenderer content="Plain paragraph." />);
    expect(container.querySelector("p")).toHaveTextContent("Plain paragraph.");
  });

  it("renders a bullet list", () => {
    render(<MarkdownRenderer content={"- item one\n- item two"} />);
    expect(screen.getByText("item one")).toBeInTheDocument();
    expect(screen.getByText("item two")).toBeInTheDocument();
  });
});

// ── TiptapRenderer ────────────────────────────────────────────────────────────

function tiptapDoc(nodes: TiptapNode[]) {
  return JSON.stringify({ type: "doc", content: nodes });
}

describe("TiptapRenderer", () => {
  it("renders a paragraph", () => {
    const { container } = render(
      <TiptapRenderer content={tiptapDoc([{ type: "paragraph", content: [{ type: "text", text: "Hello" }] }])} />,
    );
    expect(container.querySelector("p")).toHaveTextContent("Hello");
  });

  it("renders heading levels 1-3", () => {
    for (const level of [1, 2, 3] as const) {
      const { container } = render(
        <TiptapRenderer content={tiptapDoc([{ type: "heading", attrs: { level }, content: [{ type: "text", text: `H${level}` }] }])} />,
      );
      expect(container.querySelector(`h${level}`)).toHaveTextContent(`H${level}`);
    }
  });

  it("renders bullet list", () => {
    const { container } = render(
      <TiptapRenderer
        content={tiptapDoc([{
          type: "bulletList",
          content: [
            { type: "listItem", content: [{ type: "paragraph", content: [{ type: "text", text: "A" }] }] },
            { type: "listItem", content: [{ type: "paragraph", content: [{ type: "text", text: "B" }] }] },
          ],
        }])}
      />,
    );
    expect(container.querySelector("ul")).toBeInTheDocument();
    expect(screen.getByText("A")).toBeInTheDocument();
    expect(screen.getByText("B")).toBeInTheDocument();
  });

  it("renders horizontal rule", () => {
    const { container } = render(
      <TiptapRenderer content={tiptapDoc([{ type: "horizontalRule" }])} />,
    );
    expect(container.querySelector("hr")).toBeInTheDocument();
  });

  it("renders image", () => {
    const { container } = render(
      <TiptapRenderer content={tiptapDoc([{ type: "image", attrs: { src: "https://example.com/img.jpg", alt: "test" } }])} />,
    );
    const img = container.querySelector("img");
    expect(img).toHaveAttribute("src", "https://example.com/img.jpg");
    expect(img).toHaveAttribute("alt", "test");
  });

  it("falls back to plain text for invalid JSON", () => {
    const { container } = render(<TiptapRenderer content="not json" />);
    expect(container.querySelector("p")).toHaveTextContent("not json");
  });

  it("falls back when doc type is wrong", () => {
    const { container } = render(<TiptapRenderer content={JSON.stringify({ type: "unknown" })} />);
    expect(container.querySelector("p")).toBeInTheDocument();
  });
});

// ── TiptapNodeRenderer marks ──────────────────────────────────────────────────

describe("TiptapNodeRenderer marks", () => {
  it("renders bold mark", () => {
    const node: TiptapNode = { type: "text", text: "bold", marks: [{ type: "bold" }] };
    const { container } = render(<TiptapNodeRenderer node={node} />);
    expect(container.querySelector("strong")).toHaveTextContent("bold");
  });

  it("renders italic mark", () => {
    const node: TiptapNode = { type: "text", text: "italic", marks: [{ type: "italic" }] };
    const { container } = render(<TiptapNodeRenderer node={node} />);
    expect(container.querySelector("em")).toHaveTextContent("italic");
  });

  it("renders link mark with href", () => {
    const node: TiptapNode = {
      type: "text",
      text: "link",
      marks: [{ type: "link", attrs: { href: "https://example.com" } }],
    };
    const { container } = render(<TiptapNodeRenderer node={node} />);
    const a = container.querySelector("a");
    expect(a).toHaveAttribute("href", "https://example.com");
    expect(a).toHaveTextContent("link");
  });

  it("renders strike mark", () => {
    const node: TiptapNode = { type: "text", text: "strike", marks: [{ type: "strike" }] };
    const { container } = render(<TiptapNodeRenderer node={node} />);
    expect(container.querySelector("s")).toHaveTextContent("strike");
  });

  it("renders code mark", () => {
    const node: TiptapNode = { type: "text", text: "code", marks: [{ type: "code" }] };
    const { container } = render(<TiptapNodeRenderer node={node} />);
    expect(container.querySelector("code")).toHaveTextContent("code");
  });
});
