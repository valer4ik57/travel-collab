CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255),
    display_name  VARCHAR(100) NOT NULL,
    avatar_url    TEXT,
    github_id     VARCHAR(100) UNIQUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_auth_check CHECK (email IS NOT NULL OR github_id IS NOT NULL)
);

CREATE TABLE IF NOT EXISTS trips (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id    UUID NOT NULL REFERENCES users(id),
    invite_code VARCHAR(20) UNIQUE NOT NULL,
    start_date  DATE,
    end_date    DATE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT trips_dates_check CHECK (start_date IS NULL OR end_date IS NULL OR start_date <= end_date)
);

CREATE TABLE IF NOT EXISTS trip_members (
    trip_id   UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role      VARCHAR(20) NOT NULL DEFAULT 'editor',
    status    VARCHAR(20) NOT NULL DEFAULT 'active',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (trip_id, user_id),
    CONSTRAINT trip_members_role_check CHECK (role IN ('owner', 'editor', 'viewer')),
    CONSTRAINT trip_members_status_check CHECK (status IN ('active', 'left', 'removed'))
);

CREATE INDEX IF NOT EXISTS idx_trip_members_user_id ON trip_members(user_id);
CREATE INDEX IF NOT EXISTS idx_trip_members_active_user ON trip_members(user_id) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_trip_members_trip_status ON trip_members(trip_id, status);


CREATE TABLE IF NOT EXISTS trip_routes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id    UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    title      VARCHAR(255) NOT NULL,
    route_date DATE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_trip_routes_trip_id ON trip_routes(trip_id);
CREATE INDEX IF NOT EXISTS idx_trip_routes_trip_order ON trip_routes(trip_id, route_date, sort_order);

CREATE TABLE IF NOT EXISTS locations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id     UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    coordinates GEOMETRY(Point, 4326) NOT NULL,
    category    VARCHAR(50) DEFAULT 'other',
    visit_at    TIMESTAMPTZ,
    route_id    UUID REFERENCES trip_routes(id) ON DELETE SET NULL,
    route_order INTEGER,
    created_by  UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_locations_trip_id ON locations(trip_id);
CREATE INDEX IF NOT EXISTS idx_locations_coords ON locations USING GIST(coordinates);
CREATE INDEX IF NOT EXISTS idx_locations_trip_visit_at ON locations(trip_id, visit_at);
CREATE INDEX IF NOT EXISTS idx_locations_route_id ON locations(route_id);
CREATE INDEX IF NOT EXISTS idx_locations_route_order ON locations(route_id, route_order);

CREATE TABLE IF NOT EXISTS expenses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id     UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    description VARCHAR(255) NOT NULL,
    amount      NUMERIC(12, 2) NOT NULL,
    currency    CHAR(3) NOT NULL DEFAULT 'RUB',
    paid_by     UUID NOT NULL REFERENCES users(id),
    split_with  JSONB NOT NULL DEFAULT '[]',
    location_id UUID REFERENCES locations(id) ON DELETE SET NULL,
    route_id    UUID REFERENCES trip_routes(id) ON DELETE SET NULL,
    expense_at  TIMESTAMPTZ,
    payments    JSONB NOT NULL DEFAULT '[]',
    shares      JSONB NOT NULL DEFAULT '[]',
    split_mode  VARCHAR(20) NOT NULL DEFAULT 'equal',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT expenses_amount_positive CHECK (amount > 0)
);

CREATE INDEX IF NOT EXISTS idx_expenses_trip_id ON expenses(trip_id);
CREATE INDEX IF NOT EXISTS idx_expenses_location_id ON expenses(location_id);
CREATE INDEX IF NOT EXISTS idx_expenses_route_id ON expenses(route_id);
CREATE INDEX IF NOT EXISTS idx_expenses_trip_route ON expenses(trip_id, route_id);
CREATE INDEX IF NOT EXISTS idx_expenses_expense_at ON expenses(trip_id, expense_at);

CREATE TABLE IF NOT EXISTS messages (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    text    TEXT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_messages_trip_id ON messages(trip_id);
CREATE INDEX IF NOT EXISTS idx_messages_trip_sent_at ON messages(trip_id, sent_at DESC);
