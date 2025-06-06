SELECT "id",
        "stripe_customer_id",
        "status",
        "metadata",
        "item_id",
        "price_id",
        "quantity",
        "cancel_at_period_end",
        "created",
        "current_period_start",
        "current_period_end",
        "ended_at",
        "cancel_at",
        "canceled_at",
        "trial_start",
        "trial_end",
        "created_at",
        "updated_at"
FROM "stripe_subscriptions"
WHERE (
                (
                        "stripe_customer_id" IN (
                                SELECT "id"
                                FROM "stripe_customers"
                                WHERE "team_id" IN ($1)
                        )
                        AND "status" = $2
                )
                OR (
                        "stripe_customer_id" IN (
                                SELECT "id"
                                FROM "stripe_customers"
                                WHERE "team_id" IN ($3)
                        )
                        AND "status" = $4
                        AND "trial_end" > $5
                )
        ) args [019746c3-4a44-7ed8-8ea0-73faf20715ab active 019746c3-4a44-7ed8-8ea0-73faf20715ab trialing 2025-06-06T12:41:33.134683-07:00]
SELECT stripe_subscriptions.id AS "id",
        stripe_subscriptions.stripe_customer_id AS "stripe_customer_id",
        stripe_subscriptions.status AS "status",
        stripe_subscriptions.metadata AS "metadata",
        stripe_subscriptions.item_id AS "item_id",
        stripe_subscriptions.price_id AS "price_id",
        stripe_subscriptions.quantity AS "quantity",
        stripe_subscriptions.cancel_at_period_end AS "cancel_at_period_end",
        stripe_subscriptions.created AS "created",
        stripe_subscriptions.current_period_start AS "current_period_start",
        stripe_subscriptions.current_period_end AS "current_period_end",
        stripe_subscriptions.ended_at AS "ended_at",
        stripe_subscriptions.cancel_at AS "cancel_at",
        stripe_subscriptions.canceled_at AS "canceled_at",
        stripe_subscriptions.trial_start AS "trial_start",
        stripe_subscriptions.trial_end AS "trial_end",
        stripe_subscriptions.created_at AS "created_at",
        stripe_subscriptions.updated_at AS "updated_at",
        stripe_customers.id AS "stripe_customer.id",
        stripe_customers.email AS "stripe_customer.email",
        stripe_custom ers.name AS "stripe_customer.name",
        stripe_customers.user_id AS "stripe_customer.user_id",
        stripe_customers.team_id AS "stripe_customer.team_id",
        stripe_customers.customer_type AS "stripe_customer.customer_type",
        stripe_customers.billing_address AS "stripe_customer.billing_address",
        stripe_customers.payment_method AS "stripe_customer.payment_method",
        stripe_customers.created_at AS "stripe_customer.created_at",
        stripe_customers.updated_at AS "stripe_customer.updated_at"
FROM stripe_subscriptions
        JOIN stripe_customers ON stripe_subscriptions.stripe_customer_id = stripe_customers.id
WHERE (
                (
                        stripe_customers.team_id IN ($1)
                        AND stripe_subscriptions.status = $2
                )
                OR (
                        stripe_customers.team_id IN ($3)
                        AND stripe_subscriptions.status = $4
                        AND stripe_subscriptions.trial_end > $5
                )
        ) args [019746c4-7436-790c-8cf1-8861b9be7d98 active 019746c4-7436-790c-8cf1-8861b9be7d98 trialing 2025-06-06T12:42:49.407587-07:00]