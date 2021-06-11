CREATE TABLE IF NOT EXISTS membership.communication
(
    id SERIAL PRIMARY KEY,
    name text NOT NULL UNIQUE,
    subject text NOT NULL,
    frequency_throttle numeric NOT NULL CHECK (frequency_throttle >= 0),
    template text NOT NULL
);

CREATE TABLE IF NOT EXISTS membership.communication_log
(
    id BIGSERIAL PRIMARY KEY,
    member_id uuid NOT NULL REFERENCES membership.members(id),
    communication_id integer NOT NULL REFERENCES membership.communication(id)
);

INSERT INTO membership.communication
    (name, subject, frequency_throttle, template)
VALUES
    ('access_revoked_leadership', 'Membership Expired', 0, 'access_revoked_leadership.html.tmpl'),
    ('access_revoked', 'Membership Expired', 59, 'access_revoked.html.tmpl'),
    ('ip_changed', 'IP Address Changed', 0, 'ip_changed.html.tmpl'),
    ('pending_revokation_leadership', 'hackRVA Grace Period',  0, 'pending_revokation_leadership.html.tmpl'),
    ('pending_revokation_member', 'hackRVA Grace Period', 10, 'pending_revokation_member.html.tmpl'),
    ('welcome', 'Welcome to HackRVA', 60, 'welcome.html.tmpl');
