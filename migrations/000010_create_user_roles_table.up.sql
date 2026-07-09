-- Table de jonction : roles assignes DIRECTEMENT a un usager.
-- Correspond a User.RoleIDs dans le domaine.
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);
