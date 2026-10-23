import {
  User,
  Location,
  Route,
  Drive,
  OverviewStats,
  AccountShare,
  TaxRate,
  MeResponse,
  SharePermission,
  LocationType,
  AdminMetrics,
} from '../types';

let currentTargetAccountId: string | null = localStorage.getItem('kdb_target_account_id');
let currentToken: string | null = localStorage.getItem('kdb_token');

export const setTargetAccountId = (accountId: string | null) => {
  currentTargetAccountId = accountId;
  if (accountId) {
    localStorage.setItem('kdb_target_account_id', accountId);
  } else {
    localStorage.removeItem('kdb_target_account_id');
  }
};

export const setAuthToken = (token: string | null) => {
  currentToken = token;
  if (token) {
    localStorage.setItem('kdb_token', token);
  } else {
    localStorage.removeItem('kdb_token');
  }
};

export const getAuthToken = () => currentToken;

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers || {});
  headers.set('Content-Type', 'application/json');

  if (currentToken) {
    headers.set('Authorization', `Bearer ${currentToken}`);
  }

  if (currentTargetAccountId) {
    headers.set('X-Account-ID', currentTargetAccountId);
  }

  const res = await fetch(endpoint, {
    ...options,
    headers,
    credentials: 'include',
  });

  if (!res.ok) {
    let errorMsg = 'An error occurred';
    try {
      const errorData = await res.json();
      errorMsg = errorData.error || errorMsg;
    } catch {
      errorMsg = res.statusText || errorMsg;
    }
    throw new Error(errorMsg);
  }

  // Handle empty or 204 response
  const contentType = res.headers.get('content-type');
  if (contentType && contentType.includes('application/json')) {
    return res.json();
  }
  return {} as T;
}

export const api = {
  // Auth
  register: (payload: { username: string; email: string; password: string }) =>
    request<{ token: string; user: User }>('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  login: (payload: { username_or_email: string; password: string }) =>
    request<{ token: string; user: User }>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  logout: () =>
    request<{ message: string }>('/api/auth/logout', {
      method: 'POST',
    }),

  me: () => request<MeResponse>('/api/auth/me'),

  updateProfile: (payload: { username: string; email: string }) =>
    request<{ token: string; user: User }>('/api/auth/profile', {
      method: 'PUT',
      body: JSON.stringify(payload),
    }),

  changePassword: (payload: { current_password: string; new_password: string }) =>
    request<{ message: string }>('/api/auth/password', {
      method: 'PUT',
      body: JSON.stringify(payload),
    }),

  // Locations
  getLocations: () => request<Location[]>('/api/locations'),

  createLocation: (payload: { name: string; type: LocationType; address: string; recalculate_drives?: boolean }) =>
    request<Location>('/api/locations', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  updateLocation: (id: string, payload: { name: string; address: string; recalculate_drives?: boolean }) =>
    request<Location>(`/api/locations/${id}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    }),

  deleteLocation: (id: string) =>
    request<{ message: string }>(`/api/locations/${id}`, {
      method: 'DELETE',
    }),

  autocompleteAddress: (query: string) =>
    request<Array<{ id: string; text: string }>>(`/api/addresses/autocomplete?q=${encodeURIComponent(query)}`),

  recalculateDrives: () =>
    request<{ message: string; recalculated_count: number; is_yderkommune: boolean; kommune_name: string }>(
      '/api/drives/recalculate',
      { method: 'POST' }
    ),

  // Routes
  getRoutes: () => request<Route[]>('/api/routes'),

  createRoute: (payload: {
    name: string;
    origin_location_id: string;
    destination_location_id: string;
    distance_one_way_km?: number;
    distance_round_trip_km?: number;
    is_default?: boolean;
    is_hidden?: boolean;
  }) =>
    request<Route>('/api/routes', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  updateRoute: (
    id: string,
    payload: {
      name: string;
      origin_location_id: string;
      destination_location_id: string;
      distance_one_way_km?: number;
      distance_round_trip_km?: number;
      is_default?: boolean;
      is_hidden?: boolean;
    }
  ) =>
    request<Route>(`/api/routes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    }),

  deleteRoute: (id: string) =>
    request<{ message: string }>(`/api/routes/${id}`, {
      method: 'DELETE',
    }),

  // Drives
  getDrives: (params?: { year?: string; month?: string; date?: string; route_id?: string; limit?: number }) => {
    const searchParams = new URLSearchParams();
    if (params?.year) searchParams.set('year', params.year);
    if (params?.month) searchParams.set('month', params.month);
    if (params?.date) searchParams.set('date', params.date);
    if (params?.route_id) searchParams.set('route_id', params.route_id);
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    const q = searchParams.toString();
    return request<Drive[]>(`/api/drives${q ? `?${q}` : ''}`);
  },

  createDrive: (payload: {
    route_id: string;
    date?: string;
    distance_km?: number;
    notes?: string;
  }) =>
    request<Drive>('/api/drives', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  deleteDrive: (id: string) =>
    request<{ message: string }>(`/api/drives/${id}`, {
      method: 'DELETE',
    }),

  // Statistics
  getStats: (year?: string) => {
    const q = year ? `?year=${year}` : '';
    return request<OverviewStats>(`/api/stats${q}`);
  },

  getExportCSVUrl: (year?: string) => {
    const base = '/api/export';
    return year ? `${base}?year=${year}` : base;
  },

  // Account Sharing
  getShares: () =>
    request<{ my_shares: AccountShare[]; shared_with_me: AccountShare[] }>('/api/shares'),

  createShare: (payload: { username_or_email: string; permission: SharePermission }) =>
    request<AccountShare>('/api/shares', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  deleteShare: (id: string) =>
    request<{ message: string }>(`/api/shares/${id}`, {
      method: 'DELETE',
    }),

  // Tax Rates
  getRates: (year?: number) => {
    const path = year ? `/api/rates/${year}` : '/api/rates';
    return request<TaxRate>(path);
  },

  updateRates: (payload: TaxRate) =>
    request<TaxRate>('/api/rates', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),

  // Admin Telemetry (GDPR compliant)
  getAdminMetrics: () => request<AdminMetrics>('/api/admin/metrics'),
};
