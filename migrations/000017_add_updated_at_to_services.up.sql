-- Service avait deja created_at ; il ne lui manquait que updated_at,
-- contrairement a Permission/Role/Group qui n'avaient ni l'un ni
-- l'autre avant la migration 000016.
ALTER TABLE services
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();
