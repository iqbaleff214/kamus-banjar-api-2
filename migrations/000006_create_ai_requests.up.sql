CREATE TABLE ai_requests (
    id                      UUID PRIMARY KEY,
    type                    VARCHAR(50)  NOT NULL,
    target_word_id          UUID         REFERENCES words(id) ON DELETE SET NULL,
    target_contribution_id  UUID         REFERENCES contributions(id) ON DELETE SET NULL,
    requested_by            UUID         NOT NULL REFERENCES users(id),
    model                   TEXT         NOT NULL,
    prompt                  TEXT         NOT NULL,
    response                JSONB,
    parsed_output           JSONB,
    status                  VARCHAR(20)  NOT NULL DEFAULT 'pending',
    review_status           VARCHAR(20)  NOT NULL DEFAULT 'unreviewed',
    reviewed_by             UUID         REFERENCES users(id),
    reviewed_at             TIMESTAMPTZ,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_requests_target_word_id     ON ai_requests (target_word_id);
CREATE INDEX idx_ai_requests_status             ON ai_requests (status);
CREATE INDEX idx_ai_requests_review_status      ON ai_requests (review_status);
CREATE INDEX idx_ai_requests_created_at         ON ai_requests (created_at);
