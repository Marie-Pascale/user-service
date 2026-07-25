-- Nullable et sans DEFAULT, contrairement a created_at/updated_at :
-- un usager n'a jamais "zero" comme derniere connexion, il n'en a
-- simplement pas encore eu (NULL = jamais connecte).
ALTER TABLE users
    ADD COLUMN last_login_at TIMESTAMPTZ;
