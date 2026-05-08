ALTER TABLE reservations
    ALTER COLUMN status SET DEFAULT 'PENDING';

DROP INDEX IF EXISTS uq_reservations_active_spot;

CREATE UNIQUE INDEX IF NOT EXISTS uq_reservations_pending_active_spot
    ON reservations(zone_id, spot_id)
    WHERE status IN ('PENDING', 'ACTIVE', 'CONFIRMING', 'CANCELLING', 'EXPIRING');
