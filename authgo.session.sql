SELECT ss.id AS "subscription.id",
        ss.user_id AS "subscription.user_id",
        ss.status AS "subscription.status",
        ss.metadata AS "subscription.metadata",
        ss.price_id AS "subscription.price_id",
        ss.quantity AS "subscription.quantity",
        ss.cancel_at_period_end AS "subscription.cancel_at_period_end",
        ss.created AS "subscription.created",
        ss.current_period_start AS "subscription.current_period_start",
        ss.current_period_end AS "subscription.current_period_end",
        ss.ended_at AS "subscription.ended_at",
        ss.cancel_at AS "subscription.cancel_at",
        ss.canceled_at AS "subscription.canceled_at",
        ss.trial_start AS "subscription.trial_start",
        ss.trial_end AS "subscription.trial_end",
        ss.created_at AS "subscription.created_at",
        ss.updated_at AS "subscription.updated_at",
        sp.id AS "price.id",
        sp.product_id AS "price.product_id",
        sp.lookup_key AS "price.lookup_key",
        sp.active AS "price.active",
        sp.unit_amount AS "price.unit_amount",
        sp.currency AS "price.currency",
        sp.type AS "price.type",
        sp.interval AS "price.interval",
        sp.interval_count AS "price.interval_count",
        sp.trial_period_days AS "price.trial_period_days",
        sp.metadata AS "price.metadata",
        sp.created_at AS "price.created_at",
        sp.updated_at AS "price.updated_at",
        p.id AS "product.id",
        p.name AS "product.name",
        p.description AS "product.description",
        p.active AS "product.active",
        p.image AS "product.image",
        p.metadata AS "product.metadata",
        p.created_at AS "product.created_at",
        p.updated_at AS "product.updated_at"
FROM public.stripe_subscriptions ss
        JOIN public.stripe_prices sp ON ss.price_id = sp.id
        JOIN public.stripe_products p ON sp.product_id = p.id
WHERE ss.user_id = $1
        AND ss.status IN ('active', 'trialing')
ORDER BY ss.updated_at DESC;
-- SELECT tp.*,
--         u.id AS "user.id",
--         u.email AS "user.email",
--         u.name AS "user.name",
--         u.image AS "user.image",
--         u.email_verified_at AS "user.email_verified_at",
--         u.created_at AS "user.created_at",
--         u.updated_at AS "user.updated_at",
--         json_agg(to_json(t.*)) AS "tasks"
-- FROM public.task_projects tp
--         LEFT JOIN public.users u ON tp.user_id = u.id
--         LEFT JOIN public.tasks t ON tp.id = t.project_id -- WHERE tp.id = ANY ($1::uuid [])
-- GROUP BY tp.id,
--         u.id -- tp.id AS "task_project_id",
--         -- to_json(tp.*) AS "task_project",
--         -- u.id AS "user",
--         -- to_json(u.*) AS "task_project_user",
--         -- json_agg(to_json(t.*)) AS "tasks"
--         -- FROM public.task_projects tp
--         --         LEFT JOIN public.users u ON tp.user_id = u.id
--         --         LEFT JOIN public.tasks t ON tp.id = t.project_id
--         -- WHERE tp.id IN (
--         --                 'a939ca64-6b78-4102-9321-afaacf8e4cf0',
--         --                 'b1558c55-41b0-415b-a4be-5a4ee6e255d2'
--         --         )
--         -- GROUP BY tp.id,
--         --         u.id