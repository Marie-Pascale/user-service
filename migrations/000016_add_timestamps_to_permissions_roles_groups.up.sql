-- Ajoute created_at/updated_at a permissions, roles et groups -- User,
-- AuthMethod et Session les avaient deja depuis leur creation.
-- DEFAULT now() s'applique aussi aux lignes deja seedees par les
-- migrations 000013-000015 (self-service, role admin, leurs permissions).
ALTER TABLE permissions
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE roles
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE groups
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
