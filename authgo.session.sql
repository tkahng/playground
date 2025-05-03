SELECT t.*,
        (
                SELECT to_json(u.*)
                FROM users u
                WHERE u.id = t.user_id
                LIMIT 1
        ) as "user",
        (
                SELECT json_agg(to_json(tc.*))
                FROM tasks tc
                WHERE tc.parent_id = t.id
        ) as "children"
FROM tasks t