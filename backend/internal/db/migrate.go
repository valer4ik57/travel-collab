package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func EnsureMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	statements := []string{
		`ALTER TABLE trip_members ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active'`,
		`UPDATE trip_members SET status = 'active' WHERE status IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_trip_members_active_user ON trip_members(user_id) WHERE status = 'active'`,
		`CREATE INDEX IF NOT EXISTS idx_trip_members_trip_status ON trip_members(trip_id, status)`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'trip_members_status_check'
			) THEN
				ALTER TABLE trip_members ADD CONSTRAINT trip_members_status_check CHECK (status IN ('active', 'left', 'removed'));
			END IF;
		END $$;`,

		`CREATE TABLE IF NOT EXISTS trip_routes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			route_date DATE,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_trip_routes_trip_id ON trip_routes(trip_id)`,
		`CREATE INDEX IF NOT EXISTS idx_trip_routes_trip_order ON trip_routes(trip_id, route_date, sort_order)`,

		`ALTER TABLE locations ADD COLUMN IF NOT EXISTS visit_at TIMESTAMPTZ`,
		`ALTER TABLE locations ADD COLUMN IF NOT EXISTS route_id UUID REFERENCES trip_routes(id) ON DELETE SET NULL`,
		`ALTER TABLE locations ADD COLUMN IF NOT EXISTS route_order INTEGER`,
		`CREATE INDEX IF NOT EXISTS idx_locations_trip_visit_at ON locations(trip_id, visit_at)`,
		`CREATE INDEX IF NOT EXISTS idx_locations_route_id ON locations(route_id)`,
		`CREATE INDEX IF NOT EXISTS idx_locations_route_order ON locations(route_id, route_order)`,

		`ALTER TABLE expenses ADD COLUMN IF NOT EXISTS location_id UUID REFERENCES locations(id) ON DELETE SET NULL`,
		`ALTER TABLE expenses ADD COLUMN IF NOT EXISTS route_id UUID REFERENCES trip_routes(id) ON DELETE SET NULL`,
		`ALTER TABLE expenses ADD COLUMN IF NOT EXISTS expense_at TIMESTAMPTZ`,
		`ALTER TABLE expenses ADD COLUMN IF NOT EXISTS payments JSONB NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE expenses ADD COLUMN IF NOT EXISTS shares JSONB NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE expenses ADD COLUMN IF NOT EXISTS split_mode VARCHAR(20) NOT NULL DEFAULT 'equal'`,
		`CREATE INDEX IF NOT EXISTS idx_expenses_location_id ON expenses(location_id)`,
		`CREATE INDEX IF NOT EXISTS idx_expenses_route_id ON expenses(route_id)`,
		`CREATE INDEX IF NOT EXISTS idx_expenses_trip_route ON expenses(trip_id, route_id)`,
		`CREATE INDEX IF NOT EXISTS idx_expenses_expense_at ON expenses(trip_id, expense_at)`,
	}

	for _, statement := range statements {
		if _, err := pool.Exec(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
