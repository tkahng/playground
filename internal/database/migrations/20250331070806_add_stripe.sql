-- migrate:up
-- enum for stripe_customer_types. 'user' or 'team'
create type public.stripe_customer_type as enum ('user', 'team');
create table public.stripe_customers (
    -- UUID from auth.users
    -- id uuid references public.users on delete cascade on update cascade not null primary key,
    id text primary key,
    -- -- The user's customer ID in Stripe. User must not be able to update this.
    -- stripe_id text not null unique,
    email text not null,
    name text,
    -- The type of customer, either 'user' or 'team'.
    customer_type public.stripe_customer_type not null,
    user_id uuid unique references public.users on delete cascade on update cascade,
    team_id uuid unique references public.teams on delete cascade on update cascade,
    -- The customer's billing address, stored in JSON format.
    billing_address jsonb,
    -- Stores your customer's payment instruments.
    payment_method jsonb,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    CONSTRAINT stripe_customers_only_one_reference_check CHECK (
        (
            user_id IS NOT NULL
            AND team_id IS NULL
        )
        OR (
            user_id IS NULL
            AND team_id IS NOT NULL
        )
    )
);
CREATE TRIGGER handle_stripe_customers_updated_at before
update on public.stripe_customers for each row execute procedure set_current_timestamp_updated_at();
--------------- CUSTOMERS TABLE END -----------------------------------------------------------------------
create table public.stripe_products (
    -- Product ID from Stripe, e.g. prod_1234.
    id text primary key,
    -- Whether the product is currently available for purchase.
    active boolean not null default false,
    -- The product's name, meant to be displayable to the customer. Whenever this product is sold via a subscription, name will show up on associated invoice line item descriptions.
    name text not null,
    -- The product's description, meant to be displayable to the customer. Use this field to optionally store a long form explanation of the product being sold for your own rendering purposes.
    description text,
    -- A URL of the product image in Stripe, meant to be displayable to the customer.
    image text,
    -- Set of key-value pairs, used to store additional information about the object in a structured format.
    metadata jsonb not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
CREATE TRIGGER handle_stripe_products_updated_at before
update on public.stripe_products for each row execute procedure set_current_timestamp_updated_at();
create type public.stripe_pricing_type as enum ('one_time', 'recurring');
create type public.stripe_pricing_plan_interval as enum ('day', 'week', 'month', 'year');
--------------- PRICES TYPE ENUM END -----------------------------------------------------------------------
--
--------------- PRICES TABLE START -----------------------------------------------------------------------
create table public.stripe_prices (
    -- Price ID from Stripe, e.g. price_1234.
    id text primary key,
    -- The ID of the prduct that this price belongs to.
    product_id text not null references public.stripe_products,
    -- lookup_key
    lookup_key text,
    -- Whether the price can be used for new purchases.
    active boolean not null default false,
    -- The unit amount as a positive integer in the smallest currency unit (e.g., 100 cents for US$1.00 or 100 for ¥100, a zero-decimal currency).
    unit_amount bigint,
    -- Three-letter ISO currency code, in lowercase.
    currency text not null check (char_length(currency) = 3),
    -- One of `one_time` or `recurring` depending on whether the price is for a one-time purchase or a recurring (subscription) purchase.
    type public.stripe_pricing_type not null,
    -- The frequency at which a subscription is billed. One of `day`, `week`, `month` or `year`.
    interval public.stripe_pricing_plan_interval,
    -- The number of intervals (specified in the `interval` attribute) between subscription billings. For example, `interval=month` and `interval_count=3` bills every 3 months.
    interval_count bigint,
    -- Default number of trial days when subscribing a customer to this price using [`trial_from_plan=true`](https://stripe.com/docs/api#create_subscription-trial_from_plan).
    trial_period_days bigint,
    -- Set of key-value pairs, used to store additional information about the object in a structured format.
    metadata jsonb not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
CREATE TRIGGER handle_stripe_prices_updated_at before
update on public.stripe_prices for each row execute procedure set_current_timestamp_updated_at();
/**
 * STRIPE_SUBSCRIPTIONS
 * Note: stripe_subscriptions are created and managed in Stripe and synced to our DB via Stripe webhooks.
 */
--------------- STRIPE_SUBSCRIPTIONS TYPE ENUM START -----------------------------------------------------------------------
create type public.stripe_subscription_status as enum (
    'trialing',
    'active',
    'canceled',
    'incomplete',
    'incomplete_expired',
    'past_due',
    'unpaid',
    'paused'
);
--------------- STRIPE_SUBSCRIPTIONS TYPE ENUM END -----------------------------------------------------------------------
--
--------------- STRIPE_SUBSCRIPTIONS TABLE START -----------------------------------------------------------------------
create table public.stripe_subscriptions (
    -- Subscription ID from Stripe, e.g. sub_1234.
    id text primary key,
    stripe_customer_id text not null references public.stripe_customers,
    -- user_id uuid references public.users on delete cascade on update cascade not null,
    -- team_id uuid references public.teams on delete cascade on update cascade not null,
    -- The status of the subscription object, one of stripe_subscription_status type above.
    status public.stripe_subscription_status not null,
    -- Set of key-value pairs, used to store additional information about the object in a structured format.
    metadata jsonb not null,
    item_id text not null,
    price_id text not null references public.stripe_prices,
    -- item_id text not null,
    -- Quantity multiplied by the unit amount of the price creates the amount of the subscription. Can be used to charge multiple seats.
    quantity bigint not null,
    -- If true the subscription has been canceled by the user and will be deleted at the end of the billing period.
    cancel_at_period_end boolean not null default false,
    -- Time at which the subscription was created.
    created timestamptz not null default now(),
    -- Start of the current period that the subscription has been invoiced for.
    current_period_start timestamptz not null default now(),
    -- End of the current period that the subscription has been invoiced for. At the end of this period, a new invoice will be created.
    current_period_end timestamptz not null default now(),
    -- If the subscription has ended, the timestamptz of the date the subscription ended.
    ended_at timestamptz null,
    -- A date in the future at which the subscription will automatically get canceled.
    cancel_at timestamptz null,
    -- If the subscription has been canceled, the date of that cancellation. If the subscription was canceled with `cancel_at_period_end`, `canceled_at` will still reflect the date of the initial cancellation request, not the end of the subscription period when the subscription is automatically moved to a canceled state.
    canceled_at timestamptz null,
    -- If the subscription has a trial, the beginning of that trial.
    trial_start timestamptz null,
    -- If the subscription has a trial, the end of that trial.
    trial_end timestamptz null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
CREATE TRIGGER handle_stripe_subscriptions_updated_at before
update on public.stripe_subscriptions for each row execute procedure set_current_timestamp_updated_at();
--------------- STRIPE_SUBSCRIPTIONS TABLE END -----------------------------------------------------------------------
--------------- STRIPE WEBHOOK EVENTS TABLE START -----------------------------------------------------------------------
create table public.stripe_webhook_events (
    -- event.id from Stripe
    id text primary key,
    -- event.type from Stripe
    type text not null,
    -- event.object from Stripe
    object_type text not null,
    -- objects id from Stripe
    object_stripe_id text not null,
    -- event creation date
    event_creation_date timestamptz not null,
    -- stripe.event.request.id
    request_id text null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
CREATE TRIGGER handle_stripe_webhook_events_updated_at before
update on public.stripe_webhook_events for each row execute procedure set_current_timestamp_updated_at();
--------------- STRIPE WEBHOOK EVENTS TABLE END -----------------------------------------------------------------------
-- migrate:down
-- Drop the stripe_webhook_events table
DROP TRIGGER IF EXISTS handle_stripe_webhook_events_updated_at on public.stripe_webhook_events;
DROP TABLE IF EXISTS public.stripe_webhook_events;
-- Drop the stripe_subscriptions table
DROP TRIGGER IF EXISTS handle_stripe_subscriptions_updated_at on public.stripe_subscriptions;
DROP TABLE IF EXISTS public.stripe_subscriptions;
-- Drop the stripe_subscription_status type
DROP TYPE IF EXISTS public.stripe_subscription_status;
-- Drop the stripe_prices table
DROP TRIGGER IF EXISTS handle_stripe_prices_updated_at on public.stripe_prices;
DROP TABLE IF EXISTS public.stripe_prices;
-- Drop the stripe_subscription_status type
DROP TYPE IF EXISTS public.stripe_pricing_plan_interval;
DROP TYPE IF EXISTS public.stripe_pricing_type;
-- Drop the stripe_products table
DROP TRIGGER IF EXISTS handle_stripe_products_updated_at on public.stripe_products;
DROP TABLE IF EXISTS public.stripe_products;
-- Drop the stripe_customers table
DROP TRIGGER IF EXISTS handle_stripe_customers_updated_at on public.stripe_customers;
DROP TABLE IF EXISTS public.stripe_customers;
-- Drop the stripe_customer_type type
DROP TYPE IF EXISTS public.stripe_customer_type;