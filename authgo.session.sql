SELECT p.*,
        NULL::uuid AS role_id,
        sprice.product_id AS product_id -- Null indicates not directly assigned
FROM public.stripe_subscriptions ss
        JOIN public.stripe_prices sprice ON ss.price_id = sprice.id
        JOIN public.stripe_products sproduct ON sprice.product_id = sproduct.id
        JOIN public.product_permissions pr ON sproduct.id = pr.product_id
        JOIN public.permissions p ON pr.permission_id = p.id
WHERE ss.user_id = 'bee744f5-a2cf-46c2-b02c-501a562cfc5d'
        AND ss.status IN ('active', 'trialing');
-- WITH -- Get permissions assigned through roles
-- role_based_permissions AS (
--         SELECT p.*,
--                 rp.role_id,
--                 NULL::uuid AS direct_assignment -- Null indicates not directly assigned
--         FROM public.user_roles ur
--                 JOIN public.role_permissions rp ON ur.role_id = rp.role_id
--                 JOIN public.permissions p ON rp.permission_id = p.id
--         WHERE ur.user_id = 'bee744f5-a2cf-46c2-b02c-501a562cfc5d'
-- ),
-- -- Get permissions assigned through subscription product roles
-- role_based_permissions AS (
--         SELECT p.*,
--                 rp.role_id,
--                 NULL::uuid AS direct_assignment -- Null indicates not directly assigned
--         FROM public.user_roles ur
--                 JOIN public.role_permissions rp ON ur.role_id = rp.role_id
--                 JOIN public.permissions p ON rp.permission_id = p.id
--         WHERE ur.user_id = 'bee744f5-a2cf-46c2-b02c-501a562cfc5d'
-- ),
-- -- Get permissions assigned directly to user
-- direct_permissions AS (
--         SELECT p.*,
--                 NULL::uuid AS role_id,
--                 -- Null indicates not from a role
--                 up.user_id AS direct_assignment
--         FROM public.user_permissions up
--                 JOIN public.permissions p ON up.permission_id = p.id
--         WHERE up.user_id = 'bee744f5-a2cf-46c2-b02c-501a562cfc5d'
-- ),
-- -- Combine both sources
-- combined_permissions AS (
--         SELECT *
--         FROM role_based_permissions
--         UNION ALL
--         SELECT *
--         FROM direct_permissions
-- ) -- Final result with aggregated role information
-- SELECT p.id,
--         p.name,
--         p.description,
--         p.created_at,
--         p.updated_at,
--         -- Array of role IDs that grant this permission (empty if direct)
--         array_remove(array_agg(DISTINCT rp.role_id), NULL) AS role_ids,
--         -- Boolean indicating if permission is directly assigned
--         bool_or(rp.direct_assignment IS NOT NULL) AS is_directly_assigned
-- FROM (
--                 SELECT DISTINCT id,
--                         name,
--                         description,
--                         created_at,
--                         updated_at
--                 FROM combined_permissions
--         ) p
--         LEFT JOIN combined_permissions rp ON p.id = rp.id
-- GROUP BY p.id,
--         p.name,
--         p.description,
--         p.created_at,
--         p.updated_at
-- ORDER BY p.name,
--         p.id;