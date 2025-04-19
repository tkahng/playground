-- migrate:up
create table public.ai_usages (
  id uuid NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES public.users(id) on delete cascade on update cascade,
    prompt_tokens bigint not null,
    completion_tokens bigint not null,
    total_tokens bigint not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- migrate:down
drop table public.ai_usages;
