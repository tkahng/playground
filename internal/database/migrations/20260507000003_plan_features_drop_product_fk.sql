-- migrate:up
ALTER TABLE billing.plan_features
  DROP CONSTRAINT IF EXISTS plan_features_stripe_product_id_fkey;

-- migrate:down
ALTER TABLE billing.plan_features
  ADD CONSTRAINT plan_features_stripe_product_id_fkey
    FOREIGN KEY (stripe_product_id)
    REFERENCES billing.stripe_products(id)
    ON DELETE CASCADE ON UPDATE CASCADE;
