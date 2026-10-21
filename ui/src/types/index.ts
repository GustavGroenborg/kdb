export type LocationType = 'home' | 'workplace' | 'university';

export interface User {
  id: string;
  username: string;
  email: string;
  created_at: string;
}

export interface Location {
  id: string;
  user_id: string;
  name: string;
  type: LocationType;
  address: string;
  is_primary_home: boolean;
  created_at: string;
}

export interface Route {
  id: string;
  user_id: string;
  name: string;
  origin_location_id: string;
  destination_location_id: string;
  origin_location_name?: string;
  destination_location_name?: string;
  distance_one_way_km: number;
  distance_round_trip_km: number;
  is_default: boolean;
  created_at: string;
}

export interface Drive {
  id: string;
  user_id: string;
  registered_by_user_id: string;
  registered_by_name?: string;
  route_id: string;
  route_name?: string;
  date: string;
  distance_km: number;
  calculated_fradrag: number;
  notes: string;
  created_at: string;
}

export type SharePermission = 'read' | 'write';

export interface AccountShare {
  id: string;
  owner_user_id: string;
  owner_username?: string;
  shared_with_user_id: string;
  shared_with_username?: string;
  permission: SharePermission;
  created_at: string;
}

export interface SharedAccountInfo {
  owner_id: string;
  username: string;
  permission: SharePermission;
}

export interface MeResponse {
  user: User;
  has_routes: boolean;
  has_home: boolean;
  shared_accounts: SharedAccountInfo[];
}

export interface MonthlyStat {
  month_name: string;
  drives_count: number;
  distance_km: number;
  total_fradrag: number;
}

export interface RouteStat {
  route_id: string;
  route_name: string;
  destination_type: string;
  drives_count: number;
  distance_km: number;
  total_fradrag: number;
}

export interface OverviewStats {
  total_drives: number;
  total_distance_km: number;
  total_fradrag: number;
  average_km: number;
  monthly_breakdown: MonthlyStat[];
  route_breakdown: RouteStat[];
}

export interface TaxRate {
  year: number;
  threshold_km: number;
  base_rate: number;
  tier2_threshold_km: number;
  tier2_rate: number;
}
