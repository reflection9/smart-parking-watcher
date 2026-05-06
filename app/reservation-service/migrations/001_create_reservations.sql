CREATE TABLE IF NOT EXISTS reservations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    zone_id BIGINT NOT NULL,
    spot_id BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    expires_at TIMESTAMP NOT NULL,
    confirmed_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reservations_user_id ON reservations(user_id);
CREATE INDEX IF NOT EXISTS idx_reservations_zone_id ON reservations(zone_id);
CREATE INDEX IF NOT EXISTS idx_reservations_spot_id ON reservations(spot_id);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status);
CREATE INDEX IF NOT EXISTS idx_reservations_expires_at ON reservations(expires_at);

CREATE UNIQUE INDEX IF NOT EXISTS uq_reservations_active_spot
    ON reservations(zone_id, spot_id)
    WHERE status = 'ACTIVE';

