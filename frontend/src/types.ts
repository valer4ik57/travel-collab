export interface User {
  id: string
  email?: string | null
  display_name: string
  avatar_url?: string | null
  created_at: string
}

export interface Trip {
  id: string
  name: string
  description?: string | null
  owner_id: string
  invite_code: string
  start_date?: string | null
  end_date?: string | null
  created_at: string
  members_count?: number
  locations_count?: number
}

export interface TripMember {
  trip_id: string
  user_id: string
  role: 'owner' | 'editor' | 'viewer'
  status?: 'active' | 'left' | 'removed'
  joined_at: string
  display_name: string
  email?: string | null
  avatar_url?: string | null
}

export interface LocationPoint {
  id: string
  trip_id: string
  name: string
  description?: string | null
  lat: number
  lng: number
  category: string
  created_by: string
  created_at: string
}

export interface Expense {
  id: string
  trip_id: string
  description: string
  amount: number
  currency: string
  paid_by: string
  paid_by_name?: string
  split_with: string[]
  created_at: string
}

export interface Balance {
  user_id: string
  display_name: string
  amount: number
  currency: string
}

export interface ExpenseParticipantSummary {
  user_id: string
  display_name: string
  paid_total: number
  share_total: number
  net_balance: number
  currency: string
}

export interface Settlement {
  from_user_id: string
  from_name: string
  to_user_id: string
  to_name: string
  amount: number
  currency: string
}

export interface ExpensesSummary {
  expenses: Expense[]
  balances: Balance[]
  participants: ExpenseParticipantSummary[]
  settlements: Settlement[]
  total_amount: number
  currency: string
}

export interface Message {
  id: string
  trip_id: string
  user_id: string
  display_name: string
  avatar_url?: string | null
  text: string
  sent_at: string
}

export interface TripDetails {
  trip: Trip
  members: TripMember[]
  locations: LocationPoint[]
}

export interface WSEvent<T = unknown> {
  type: string
  payload: T
}
