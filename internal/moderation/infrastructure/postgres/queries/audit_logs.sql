-- name: CreateAuditLog :exec
INSERT INTO audit_logs (id, actor_id, action, target_type, target_id, metadata, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetModerationStats :one
SELECT
    (SELECT COUNT(*) FROM contributions WHERE status = 'pending')::int                                                          AS pending_contributions,
    (SELECT COUNT(*) FROM comments WHERE is_flagged = true)::int                                                                AS flagged_comments,
    (SELECT COUNT(*) FROM contributions WHERE status = 'approved' AND reviewed_at >= NOW() - INTERVAL '7 days')::int           AS approved_this_week,
    (SELECT COUNT(*) FROM contributions WHERE status = 'rejected' AND reviewed_at >= NOW() - INTERVAL '7 days')::int           AS rejected_this_week;
