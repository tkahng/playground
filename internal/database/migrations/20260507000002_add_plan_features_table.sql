-- migrate:up
CREATE TABLE billing.plan_features (
  id                uuid        PRIMARY KEY DEFAULT uuidv7(),
  stripe_product_id text        NOT NULL UNIQUE REFERENCES billing.stripe_products(id) ON DELETE CASCADE ON UPDATE CASCADE,
  daily_ai_tokens   bigint      NOT NULL DEFAULT 10000,
  created_at        timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at        timestamptz NOT NULL DEFAULT clock_timestamp()
);

-- migrate:down
DROP TABLE IF EXISTS billing.plan_features;
