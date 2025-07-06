-- migrate:up
create table if not exists public.product_roles (
    product_id text references public.stripe_products on delete cascade on update cascade not null,
    role_id uuid references public.roles on delete cascade on update cascade not null,
    primary key (product_id, role_id)
);
create table if not exists public.product_permissions (
    product_id text references public.stripe_products on delete cascade on update cascade not null,
    permission_id uuid references public.permissions on delete cascade on update cascade not null,
    primary key (product_id, permission_id)
);
-- migrate:down
drop table if exists public.product_permissions;
drop table if exists public.product_roles;