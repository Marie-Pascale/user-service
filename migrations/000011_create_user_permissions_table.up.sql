-- Table de jonction : permissions assignees DIRECTEMENT a un usager,
-- en plus de celles heritees de ses roles.
-- Correspond a User.PermissionIDs dans le domaine.
CREATE TABLE user_permissions (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, permission_id)
);
