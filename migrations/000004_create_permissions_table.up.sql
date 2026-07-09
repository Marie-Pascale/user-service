CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    description TEXT,
    UNIQUE (service_id, key)
);

CREATE INDEX idx_permissions_service_id ON permissions(service_id);
