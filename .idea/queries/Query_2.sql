SELECT id, (
    SELECT array_agg(u.username ORDER BY t.ord)
    FROM unnest(c."userIds") WITH ORDINALITY AS t(uid, ord)
             JOIN users u ON u.id = t.uid
) AS usernames FROM "Citys" as c;