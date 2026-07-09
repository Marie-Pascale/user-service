-- Table de jonction : roles assignes a un groupe (herites par ses membres).
-- Correspond a Group.RoleIDs dans le domaine.
CREATE TABLE group_roles (
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, role_id)
);
