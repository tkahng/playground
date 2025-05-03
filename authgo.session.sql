SELECT tp.*,
        u.id AS "user.id",
        u.email AS "user.email",
        u.name AS "user.name",
        u.image AS "user.image",
        u.email_verified_at AS "user.email_verified_at",
        u.created_at AS "user.created_at",
        u.updated_at AS "user.updated_at",
        json_agg(to_json(t.*)) AS "tasks"
FROM public.task_projects tp
        LEFT JOIN public.users u ON tp.user_id = u.id
        LEFT JOIN public.tasks t ON tp.id = t.project_id -- WHERE tp.id = ANY ($1::uuid [])
GROUP BY tp.id,
        u.id -- tp.id AS "task_project_id",
        -- to_json(tp.*) AS "task_project",
        -- u.id AS "user",
        -- to_json(u.*) AS "task_project_user",
        -- json_agg(to_json(t.*)) AS "tasks"
        -- FROM public.task_projects tp
        --         LEFT JOIN public.users u ON tp.user_id = u.id
        --         LEFT JOIN public.tasks t ON tp.id = t.project_id
        -- WHERE tp.id IN (
        --                 'a939ca64-6b78-4102-9321-afaacf8e4cf0',
        --                 'b1558c55-41b0-415b-a4be-5a4ee6e255d2'
        --         )
        -- GROUP BY tp.id,
        --         u.id