-- role_permissions et permissions sont supprimees en cascade via les
-- contraintes ON DELETE CASCADE (voir migrations 000004 et 000006).
DELETE FROM roles WHERE id = '00000000-0000-0000-0000-000000000002';
DELETE FROM services WHERE id = '00000000-0000-0000-0000-000000000001';
