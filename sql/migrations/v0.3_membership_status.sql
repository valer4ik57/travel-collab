ALTER TABLE trip_members ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';
UPDATE trip_members SET status = 'active' WHERE status IS NULL;
CREATE INDEX IF NOT EXISTS idx_trip_members_active_user ON trip_members(user_id) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_trip_members_trip_status ON trip_members(trip_id, status);
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'trip_members_status_check'
    ) THEN
        ALTER TABLE trip_members ADD CONSTRAINT trip_members_status_check CHECK (status IN ('active', 'left', 'removed'));
    END IF;
END $$;
