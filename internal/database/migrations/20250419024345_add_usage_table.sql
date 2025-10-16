-- migrate:up
create table public.ai_usages (
  id uuid primary key DEFAULT uuidv7(),
  user_id uuid NOT NULL REFERENCES public.users(id) on delete cascade on update cascade,
  prompt_tokens bigint not null,
  completion_tokens bigint not null,
  total_tokens bigint not null,
  created_at timestamptz not null default clock_timestamp(),
  updated_at timestamptz not null default clock_timestamp()
);
-- migrate:down
drop table public.ai_usages;