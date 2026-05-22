-- migrate:up
-- Align DB default with the editor default (markdown, not tiptap).
-- The tiptap format requires a rich-text editor; markdown is the plain-text default
-- until @tiptap/react is integrated.
alter table blog.posts alter column content_format set default 'markdown';

-- migrate:down
alter table blog.posts alter column content_format set default 'tiptap';
