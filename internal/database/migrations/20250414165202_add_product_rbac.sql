-- migrate:up
create table if not exists billing.product_roles (
    product_id text references billing.stripe_products on delete cascade on update cascade not null,
    role_id uuid references auth.roles on delete cascade on update cascade not null,
    primary key (product_id, role_id)
);
create table if not exists billing.product_permissions (
    product_id text references billing.stripe_products on delete cascade on update cascade not null,
    permission_id uuid references auth.permissions on delete cascade on update cascade not null,
    primary key (product_id, permission_id)
);
-- migrate:down
drop table if exists billing.product_permissions;
drop table if exists billing.product_roles;