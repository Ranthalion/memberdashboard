CREATE TABLE IF NOT EXISTS membership.communication
(
    id SERIAL PRIMARY KEY,
    title text NOT NULL,
    frequency_throttle numeric NOT NULL CHECK (frequency_throttle >= 0)
);

CREATE TABLE IF NOT EXISTS membership.communication_log
(
    id BIGSERIAL PRIMARY KEY,
    member_id uuid NOT NULL REFERENCES membership.members(id),
    communication_id integer NOT NULL REFERENCES membership.communication(id)
);

INSERT INTO membership.communication
    (title, frequency_throttle)
VALUES
    ('access_revoked_leadership', 0),
    ('access_revoked', 59),
    ('ip_changed', 0),
    ('pending_revokation_leadership', 0),
    ('pending_revokation_member', 10),
    ('welcome', 60);
