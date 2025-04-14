-- VALUES ('{ "message": "test" }', 'default');
WITH -- Get permissions assigned through roles
user_role_permissions AS (
    SELECT ur.user_id AS user_id,
        p.name AS permission,
        r.name AS role
    FROM public.user_roles ur
        JOIN public.roles r ON ur.role_id = r.id
        JOIN public.role_permissions rp ON ur.role_id = rp.role_id
        JOIN public.permissions p ON rp.permission_id = p.id -- WHERE ur.user_id = '575a9f91-159e-4680-ba8b-3fc4db40d194'
),
user_direct_permissions AS (
    SELECT up.user_id AS user_id,
        p.name AS permission,
        NULL::text AS role
    FROM public.user_permissions up
        JOIN public.permissions p ON up.permission_id = p.id
),
user_sub_role_permissions AS (
    SELECT u.id AS user_id,
        p.name AS permission,
        r.name AS role
    FROM public.stripe_subscriptions s
        JOIN public.users u ON s.user_id = u.id
        JOIN public.stripe_prices price ON s.price_id = price.id
        JOIN public.stripe_products product ON price.product_id = product.id
        JOIN public.product_roles pr ON product.id = pr.product_id
        JOIN public.roles r ON pr.role_id = r.id
        JOIN public.role_permissions rp ON r.id = rp.role_id
        JOIN public.permissions p ON rp.permission_id = p.id -- WHERE u.id = '575a9f91-159e-4680-ba8b-3fc4db40d194'
),
combined_permissions AS (
    SELECT *
    FROM user_role_permissions
    UNION ALL
    SELECT *
    FROM user_direct_permissions
    UNION ALL
    SELECT *
    FROM user_sub_role_permissions
)
SELECT u.id AS user_id,
    u.email AS email,
    array_remove(ARRAY_AGG(DISTINCT p.role), NULL)::text [] AS roles,
    array_remove(ARRAY_AGG(DISTINCT p.permission), NULL)::text [] AS permissions,
    array_remove(ARRAY_AGG(DISTINCT ua.provider), NULL)::public.providers [] AS providers
FROM public.users u
    LEFT JOIN combined_permissions p ON u.id = p.user_id
    LEFT JOIN public.user_accounts ua ON u.id = ua.user_id
WHERE u.email = 'tkahng+01@gmail.com'
GROUP BY u.id
LIMIT 1;
-- -- Get permissions assigned directly to user
-- direct_permissions AS (
--     SELECT p.*,
--         NULL::uuid AS role_id,
--         -- Null indicates not from a role
--         up.user_id AS direct_assignment
--     FROM public.user_permissions up
--         JOIN public.permissions p ON up.permission_id = p.id
--     WHERE up.user_id = 'bb59e199-8748-43fc-a3e0-407e658234e2'
-- ),
-- -- Combine both sources
-- combined_permissions AS (
--     SELECT *
--     FROM role_based_permissions
--     UNION ALL
--     SELECT *
--     FROM direct_permissions
-- ) -- Final result with aggregated role information
-- -- SELECT p.id,
--     p.name,
--     p.description,
--     p.created_at,
--     p.updated_at,
--     -- Array of role IDs that grant this permission (empty if direct)
--     array []::uuid [] AS role_ids,
--     -- Boolean indicating if permission is directly assigned
--     false AS is_directly_assigned
-- SELECT COUNT(DISTINCT p.id)
-- FROM public.permissions p
--     LEFT JOIN combined_permissions cp ON p.id = cp.id
-- WHERE cp.id IS NULL;
-- GROUP BY p.id
-- ORDER BY p.name,
-- p.id
-- LIMIT 10 OFFSET 0;
-- SELECT p.*
-- FROM public.permissions p
-- LEFT JOIN combined_permissions rp ON p.id = rp.id
-- WHERE rp.id IS NULL;
-- SELECT p.*
-- FROM public.permissions p
--     LEFT JOIN public.user_permissions up ON p.id = up.permission_id
--     AND up.user_id = '4481343d-a744-4685-8586-80df2f6ddf85'
--     LEFT JOIN public.user_roles ur ON up.user_id = ur.user_id
--     LEFT JOIN public.roles r ON ur.role_id = r.id
--     LEFT JOIN public.role_permissions rp ON r.id = rp.role_id
--     AND rp.permission_id = p.id
-- WHERE up.permission_id IS NULL
--     AND rp.permission_id IS NULL
-- GROUP BY p.id
-- ORDER BY p.name
-- LIMIT 10 OFFSET 0;
-- GROUP BY p.id,
--     p.name,
--     p.description,
--     p.created_at,
--     p.updated_at,
--     rp.id
-- ORDER BY p.name,
--     p.id;
-- WITH -- Get permissions assigned through roles
-- role_based_permissions AS (
--     SELECT p.*,
--         rp.role_id,
--         NULL::uuid AS direct_assignment -- Null indicates not directly assigned
--     FROM public.user_roles ur
--         JOIN public.role_permissions rp ON ur.role_id = rp.role_id
--         JOIN public.permissions p ON rp.permission_id = p.id
--     WHERE ur.user_id = '4481343d-a744-4685-8586-80df2f6ddf85'
-- ),
-- -- Get permissions assigned directly to user
-- direct_permissions AS (
--     SELECT p.*,
--         NULL::uuid AS role_id,
--         -- Null indicates not from a role
--         up.user_id AS direct_assignment
--     FROM public.user_permissions up
--         JOIN public.permissions p ON up.permission_id = p.id
--     WHERE up.user_id = '4481343d-a744-4685-8586-80df2f6ddf85'
-- ),
-- -- Combine both sources
-- combined_permissions AS (
--     SELECT *
--     FROM role_based_permissions
--     UNION ALL
--     SELECT *
--     FROM direct_permissions
-- ) -- Final result with aggregated role information
-- SELECT p.id,
--     p.name,
--     p.description,
--     p.created_at,
--     p.updated_at,
--     -- Array of role IDs that grant this permission (empty if direct)
--     array_remove(array_agg(DISTINCT rp.role_id), NULL) AS role_ids,
--     -- Boolean indicating if permission is directly assigned
--     bool_or(rp.direct_assignment IS NOT NULL) AS is_directly_assigned
-- FROM (
--         SELECT DISTINCT id,
--             name,
--             description,
--             created_at,
--             updated_at
--         FROM combined_permissions
--     ) p
--     LEFT JOIN combined_permissions rp ON p.id = rp.id
-- GROUP BY p.id,
--     p.name,
--     p.description,
--     p.created_at,
--     p.updated_at
-- ORDER BY p.name,
--     p.id;
-- SELECT p.*
-- FROM public.permissions p
--     LEFT JOIN public.role_permissions rp ON p.id = rp.permission_id
--     AND rp.role_id = 'eb2ad8b3-eac7-4e88-8361-82845cc57624'
-- WHERE rp.permission_id IS NULL
-- ORDER BY p.name
-- LIMIT 10 OFFSET 0;
-- SELECT COUNT(p.*)
-- FROM public.permissions p
--     LEFT JOIN public.role_permissions rp ON p.id = rp.permission_id
--     AND rp.role_id = 'eb2ad8b3-eac7-4e88-8361-82845cc57624'
-- WHERE rp.permission_id IS NULL;
-- WITH RolePermissions AS (
--     SELECT ur.user_id as user_id,
--         rp.role_id::uuid as role,
--         rp.permission_id as permission
--     FROM user_roles ur
--         LEFT JOIN roles r ON ur.role_id = r.id
--         LEFT JOIN role_permissions rp ON r.id = rp.role_id
-- ),
-- UserPermissions AS (
--     SELECT up.user_id as user_id,
--         NULL::uuid as role,
--         up.permission_id as permission
--     FROM user_permissions up
-- ),
-- AllPermissions AS(
--     SELECT user_id,
--         role,
--         permission
--     FROM RolePermissions rp
--     UNION
--     SELECT user_id,
--         role,
--         permission
--     FROM UserPermissions up
-- )
-- SELECT user_id,
--     role,
--     permission
-- FROM AllPermissions
-- WHERE user_id = '4481343d-a744-4685-8586-80df2f6ddf85';
-- WITH RolePermissions AS (
--     SELECT u.id as user_id,
--         u.email as email,
--         p.name as permission,
--         ar.name as role
--     FROM public.permissions p
--         LEFT JOIN public.role_permissions rp ON p.id = rp.permission_id
--         LEFT JOIN public.roles ar ON rp.role_id = ar.id
--         LEFT JOIN public.user_roles ur ON ar.id = ur.role_id
--         LEFT JOIN public.users u ON ur.user_id = u.id
-- ),
-- UserPermissions AS (
--     SELECT u.id as user_id,
--         u.email as email,
--         p.name as permission,
--         NULL as role
--     FROM public.permissions p
--         LEFT JOIN public.user_permissions up ON p.id = up.permission_id
--         LEFT JOIN public.users u ON up.user_id = u.id
-- ),
-- AllPermissions AS(
--     SELECT user_id,
--         email,
--         role,
--         permission
--     FROM RolePermissions rp
--     UNION
--     SELECT user_id,
--         email,
--         role,
--         permission
--     FROM UserPermissions up
-- )
-- SELECT user_id,
--     email,
--     role,
--     permission
-- FROM AllPermissions
-- WHERE email = 'tkahng+01@gmail.com';
-- SELECT u.id AS user_id,
--     u.email AS email,
--     ar.name AS role,
--     p.name AS permission,
--     p2.name AS permission2
-- FROM public.users u
--     LEFT JOIN public.user_roles ur ON u.id = ur.user_id
--     LEFT JOIN public.roles ar ON ur.role_id = ar.id
--     LEFT JOIN public.role_permissions rp ON ar.id = rp.role_id
--     LEFT JOIN public.permissions p ON rp.permission_id = p.id
--     LEFT JOIN public.user_permissions up ON u.id = up.user_id
--     LEFT JOIN public.permissions p2 ON up.permission_id = p2.id
-- WHERE u.email = 'tkahng+01@gmail.com';
-- )
-- SELECT fa.user_id AS id,
--     to_json(fa.*) AS info
-- FROM FilteredAccounts fa
-- WHERE fa.user_id IN (
--         '7d2574db-bd61-4b68-be42-c5b6d96ff564',
--         '43dddcb1-4ac3-4ce0-bcbd-faa662b25cfc',
--         '35f39cd6-558d-4bd4-ab10-441ac6d90e6a'
--     );
-- SELECT u.id AS user_id,
--     u.email AS email,
--     to_json(u.*) AS user,
--     ARRAY_AGG(DISTINCT ar.name)::text [] AS roles,
--     ARRAY_AGG(DISTINCT p.name)::text [] AS permissions,
--     ARRAY_AGG(DISTINCT ua.provider)::public.providers [] AS providers
-- FROM public.users u
--     LEFT JOIN public.user_roles ur ON u.id = ur.user_id
--     LEFT JOIN public.roles ar ON ur.role_id = ar.id
--     LEFT JOIN public.role_permissions rp ON ar.id = rp.role_id
--     LEFT JOIN public.permissions p ON rp.permission_id = p.id
--     LEFT JOIN public.user_accounts ua ON u.id = ua.user_id
-- GROUP BY u.id OFFSET 20
-- LIMIT 10;
-- INSERT INTO "roles" AS "roles" (
--         "id",
--         "name",
--         "description",
--         "created_at",
--         "updated_at"
--     )
-- VALUES (DEFAULT, 'hello', DEFAULT, DEFAULT, DEFAULT) ON CONFLICT (name) DO
-- UPDATE
-- SET "created_at" = now()
-- RETURNING *;
-- SELECT to_json(obj) AS user
-- FROM (
--         SELECT u.*,
--             ARRAY_AGG(DISTINCT ar.name)::text [] AS roles,
--             ARRAY_AGG(DISTINCT p.name)::text [] AS permissions
--         FROM public.users u
--             LEFT JOIN public.user_roles ur ON u.id = ur.user_id
--             LEFT JOIN public.roles ar ON ur.role_id = ar.id
--             LEFT JOIN public.role_permissions rp ON ar.id = rp.role_id
--             LEFT JOIN public.permissions p ON rp.permission_id = p.id
--         WHERE u.email = 'tkahng@gmail.com'
--         GROUP BY u.id
--         LIMIT 1
--     ) AS obj
-- LIMIT 1;
-- SELECT u.*,
--     ARRAY_AGG(DISTINCT ar.name)::text [] AS roles,
--     ARRAY_AGG(DISTINCT p.name)::text [] AS permissions
-- FROM public.users u
--     LEFT JOIN public.user_roles ur ON u.id = ur.user_id
--     LEFT JOIN public.roles ar ON ur.role_id = ar.id
--     LEFT JOIN public.role_permissions rp ON ar.id = rp.role_id
--     LEFT JOIN public.permissions p ON rp.permission_id = p.id
-- GROUP BY u.id;
-- WITH FilteredAccounts AS (
--     SELECT *
--     FROM public.user_accounts
--     WHERE provider = 'github'
-- )
-- SELECT u.*,
--     a.*
-- FROM public.users u
--     LEFT JOIN FilteredAccounts a ON u.id = a."user_id"
-- WHERE u.email = 'tkahng@gmail.com'
-- LIMIT 1;
-- 
-- INSERT INTO public.user_accounts (
--         "user_id",
--         type,
--         provider,
--         provider_account_id
--     )
-- VALUES (
--         '6331c5d3-4f7f-4301-b627-07dd8b496535',
--         'credentials',
--         'credentials',
--         'tkahng'
--     )
-- RETURNING *;
-- INSERT INTO public.users (email, name)
-- VALUES ('tkahng@gmail.com', 'tkahng')
-- RETURNING *;
-- SELECT p.*
-- from roles r
--     LEFT JOIN role_permissions rp ON r.id = rp.role_id
--     LEFT JOIN permissions p ON rp.permission_id = p.id
-- WHERE r.name = 'pro';
-- SELECT r.name,
--     ARRAY_AGG(p.name)
-- from roles r
--     LEFT JOIN role_permissions rp ON r.id = rp.role_id
--     LEFT JOIN permissions p ON rp.permission_id = p.id
-- WHERE r.name = 'pro'
--     OR r.name = 'admin'
-- GROUP BY r.name;
-- FROM users u
--     LEFT JOIN user_roles ur ON u.id = ur.user_id
--     LEFT JOIN roles ar ON ur.role_id = ar.id
--     LEFT JOIN role_permissions rp ON ar.id = rp.role_id
--     LEFT JOIN permissions p ON rp.permission_id = p.id