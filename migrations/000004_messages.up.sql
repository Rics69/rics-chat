CREATE TABLE chat.messages (
    id BIGSERIAL PRIMARY KEY,
    sender_id BIGINT NOT NULL REFERENCES chat.users(id),
    recipient_id BIGINT NOT NULL REFERENCES chat.users(id),
    text VARCHAR(4096) NOT NULL CHECK (char_length(text) BETWEEN 1 AND 4096),
    created_at TIMESTAMPTZ NOT NULL,

    peer_low BIGINT GENERATED ALWAYS AS (LEAST(sender_id, recipient_id)) STORED,
    peer_high BIGINT GENERATED ALWAYS AS (GREATEST(sender_id, recipient_id)) STORED,

    CHECK (sender_id <> recipient_id)
);

CREATE INDEX messages_dialog_idx ON chat.messages (peer_low, peer_high, id DESC);
CREATE INDEX messages_sender_idx ON chat.messages (sender_id);
CREATE INDEX messages_recipient_idx ON chat.messages (recipient_id);
