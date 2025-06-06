WITH latest_subscriptions AS (
        SELECT DISTINCT ON (stripe_customer_id) *
        FROM stripe_subscriptions
        WHERE status IN ('active', 'trialing')
        ORDER BY stripe_customer_id,
                created DESC
)
SELECT t.id AS team_id,
        sc.id AS stripe_customer_id,
        sc.email AS stripe_customer_email,
        sc.customer_type,
        ss.id AS stripe_subscription_id,
        ss.status AS stripe_subscription_status,
        ss.created AS subscription_created_at,
        ss.current_period_end
FROM teams t
        LEFT JOIN stripe_customers sc ON sc.team_id = t.id
        AND sc.customer_type = 'team'
        LEFT JOIN latest_subscriptions ss ON ss.stripe_customer_id = sc.id -- WHERE t.id = ANY($1::uuid []);
WHERE t.id IN (
                '01972f1f-6244-7dcb-ab0b-bd0f39674659'
        );