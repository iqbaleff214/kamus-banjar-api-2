CREATE TABLE audit_logs (
    id          UUID        PRIMARY KEY,
    actor_id    UUID        NOT NULL REFERENCES users(id),
    action      VARCHAR(50) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id   UUID        NOT NULL,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_actor_id  ON audit_logs (actor_id);
CREATE INDEX idx_audit_logs_action    ON audit_logs (action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at);
