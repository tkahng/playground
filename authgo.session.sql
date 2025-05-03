SELECT tp.*,
        (
                SELECT to_json(u.*)
                FROM users u
                WHERE u.id = tp.user_id
                LIMIT 1
        ) as "user",
        (
                SELECT json_agg(to_json(t.*))
                FROM tasks t
                WHERE t.task_project_id = tp.id
        ) as "tasks"
FROM task_projects tp