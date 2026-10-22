import React, { useState, useEffect } from 'react';
import { OverviewStats, Drive, Route } from '../types';
import { api } from '../services/api';
import { useAuth } from '../context/AuthContext';
import {
  BarChart3,
  Car,
  Coins,
  Navigation,
  Download,
  Calendar,
  Filter,
  Trash2,
  Building2,
  GraduationCap,
  ChevronDown,
  ChevronUp,
} from 'lucide-react';

const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
];

const getMonthDisplayName = (monthStr: string): string => {
  const parts = monthStr.split('-');
  const monthIdx = parseInt(parts[parts.length - 1], 10) - 1;
  if (monthIdx >= 0 && monthIdx < 12) {
    return MONTH_NAMES[monthIdx];
  }
  return monthStr;
};

const ROUTE_PALETTES = [
  { bg: 'bg-brand-600', text: 'text-brand-600' },
  { bg: 'bg-emerald-500', text: 'text-emerald-600' },
  { bg: 'bg-amber-500', text: 'text-amber-600' },
  { bg: 'bg-purple-500', text: 'text-purple-600' },
  { bg: 'bg-rose-500', text: 'text-rose-600' },
  { bg: 'bg-cyan-500', text: 'text-cyan-600' },
  { bg: 'bg-indigo-500', text: 'text-indigo-600' },
  { bg: 'bg-teal-500', text: 'text-teal-600' },
];

const getRoutePalette = (routeId: string, fallbackIndex: number) => {
  if (routeId) {
    let hash = 0;
    for (let i = 0; i < routeId.length; i++) {
      hash = (hash << 5) - hash + routeId.charCodeAt(i);
      hash |= 0;
    }
    const idx = Math.abs(hash) % ROUTE_PALETTES.length;
    return ROUTE_PALETTES[idx];
  }
  return ROUTE_PALETTES[fallbackIndex % ROUTE_PALETTES.length];
};

export const OverviewTab: React.FC = () => {
  const { activePermission, activeAccountName } = useAuth();
  const [selectedYear, setSelectedYear] = useState<string>(new Date().getFullYear().toString());
  const [stats, setStats] = useState<OverviewStats | null>(null);
  const [drives, setDrives] = useState<Drive[]>([]);
  const [routes, setRoutes] = useState<Route[]>([]);
  const [selectedRouteFilter, setSelectedRouteFilter] = useState<string>('');
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [activeExpandedMonth, setActiveExpandedMonth] = useState<string | null>(null);
  const [hoveredRoute, setHoveredRoute] = useState<{
    month: string;
    routeId: string;
    routeName: string;
    trips: number;
    pct: number;
    fradrag: number;
    km: number;
  } | null>(null);

  const canWrite = activePermission === 'write';

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const [fetchedStats, fetchedDrives, fetchedRoutes] = await Promise.all([
        api.getStats(selectedYear),
        api.getDrives({ year: selectedYear, route_id: selectedRouteFilter || undefined, limit: 100 }),
        api.getRoutes(),
      ]);
      setStats(fetchedStats);
      setDrives(fetchedDrives);
      setRoutes(fetchedRoutes);
    } catch (err) {
      console.error('Failed to load overview statistics:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [selectedYear, selectedRouteFilter, activeAccountName]);

  const handleDeleteDrive = async (id: string) => {
    if (!canWrite) return;
    if (!confirm('Are you sure you want to delete this drive record?')) return;
    try {
      await api.deleteDrive(id);
      fetchData();
    } catch (err: any) {
      alert(err.message || 'Failed to delete drive');
    }
  };

  const handleExportCSV = () => {
    const url = api.getExportCSVUrl(selectedYear);
    window.open(url, '_blank');
  };

  return (
    <div className="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
      {/* Header and Controls */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between pb-6 border-b border-slate-200 mb-8 gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Overview & Statistics</h1>
          <p className="text-xs text-slate-500 mt-1">
            Tax deduction breakdown and drive commute history.
          </p>
        </div>

        <div className="flex items-center space-x-3">
          {/* Year Selector */}
          <div className="flex items-center space-x-2 bg-white border border-slate-200 rounded-xl px-3 py-1.5 shadow-sm">
            <Calendar className="w-4 h-4 text-slate-400" />
            <select
              value={selectedYear}
              onChange={(e) => setSelectedYear(e.target.value)}
              className="text-xs font-semibold text-slate-700 bg-transparent focus:outline-none"
            >
              <option value="2026">Tax Year 2026</option>
              <option value="2025">Tax Year 2025</option>
              <option value="2024">Tax Year 2024</option>
            </select>
          </div>

          {/* Export to CSV */}
          <button
            onClick={handleExportCSV}
            className="flex items-center space-x-2 px-3.5 py-2 bg-slate-900 hover:bg-slate-800 text-white rounded-xl text-xs font-semibold shadow-sm transition-all"
          >
            <Download className="w-3.5 h-3.5" />
            <span>Export CSV for SKAT</span>
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="py-16 text-center text-slate-400">
          <div className="animate-spin w-8 h-8 border-4 border-brand-500 border-t-transparent rounded-full mx-auto mb-3"></div>
          Calculating statistics...
        </div>
      ) : (
        <>
          {/* Top KPI Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            <div className="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-sm">
              <div className="flex items-center justify-between text-slate-500 mb-2">
                <span className="text-xs font-bold uppercase tracking-wider">Total Deduction</span>
                <div className="w-8 h-8 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center">
                  <Coins className="w-4 h-4" />
                </div>
              </div>
              <div className="text-2xl font-black text-slate-900">
                {(stats?.total_fradrag || 0).toLocaleString('da-DK', {
                  minimumFractionDigits: 2,
                  maximumFractionDigits: 2,
                })}{' '}
                <span className="text-xs font-semibold text-slate-500">DKK</span>
              </div>
              <div className="text-[11px] text-emerald-600 mt-1 font-medium">
                Befordringsfradrag for {selectedYear}
              </div>
            </div>

            <div className="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-sm">
              <div className="flex items-center justify-between text-slate-500 mb-2">
                <span className="text-xs font-bold uppercase tracking-wider">Total Drives</span>
                <div className="w-8 h-8 rounded-lg bg-blue-50 text-blue-600 flex items-center justify-center">
                  <Car className="w-4 h-4" />
                </div>
              </div>
              <div className="text-2xl font-black text-slate-900">
                {stats?.total_drives || 0}
              </div>
              <div className="text-[11px] text-slate-500 mt-1">
                Round-trip commutes recorded
              </div>
            </div>

            <div className="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-sm">
              <div className="flex items-center justify-between text-slate-500 mb-2">
                <span className="text-xs font-bold uppercase tracking-wider">Total Distance</span>
                <div className="w-8 h-8 rounded-lg bg-violet-50 text-violet-600 flex items-center justify-center">
                  <Navigation className="w-4 h-4" />
                </div>
              </div>
              <div className="text-2xl font-black text-slate-900">
                {(stats?.total_distance_km || 0).toLocaleString('da-DK', {
                  maximumFractionDigits: 1,
                })}{' '}
                <span className="text-xs font-semibold text-slate-500">km</span>
              </div>
              <div className="text-[11px] text-slate-500 mt-1">
                Total kilometers traveled
              </div>
            </div>

            <div className="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-sm">
              <div className="flex items-center justify-between text-slate-500 mb-2">
                <span className="text-xs font-bold uppercase tracking-wider">Average Commute</span>
                <div className="w-8 h-8 rounded-lg bg-amber-50 text-amber-600 flex items-center justify-center">
                  <BarChart3 className="w-4 h-4" />
                </div>
              </div>
              <div className="text-2xl font-black text-slate-900">
                {(stats?.average_km || 0).toFixed(1)}{' '}
                <span className="text-xs font-semibold text-slate-500">km</span>
              </div>
              <div className="text-[11px] text-slate-500 mt-1">
                Average round-trip per drive
              </div>
            </div>
          </div>

          {/* Breakdown Section: Monthly Trends and Route Split */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
            {/* Monthly Breakdown */}
            <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-sm">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-sm font-bold text-slate-900 flex items-center space-x-2">
                  <Calendar className="w-4 h-4 text-brand-600" />
                  <span>Monthly Deduction Breakdown</span>
                </h3>
                <span className="text-[11px] text-slate-400 hidden sm:inline">
                  Hover or tap bar for route trips
                </span>
              </div>

              {!stats?.monthly_breakdown || stats.monthly_breakdown.length === 0 ? (
                <div className="text-xs text-slate-400 py-8 text-center">
                  No monthly activity recorded for {selectedYear}.
                </div>
              ) : (
                <div className="space-y-4">
                  {stats.monthly_breakdown.map((m) => {
                    const maxFradrag = Math.max(
                      ...stats.monthly_breakdown.map((item) => item.total_fradrag),
                      1
                    );
                    const pct = Math.min(100, Math.round((m.total_fradrag / maxFradrag) * 100));
                    const isExpanded = activeExpandedMonth === m.month_name;
                    const monthLabel = getMonthDisplayName(m.month_name);

                    return (
                      <div
                        key={m.month_name}
                        className="text-xs group"
                      >
                        {/* Month header row */}
                        <div
                          onClick={() => setActiveExpandedMonth(isExpanded ? null : m.month_name)}
                          className="flex justify-between items-center font-semibold text-slate-700 mb-1.5 cursor-pointer select-none"
                        >
                          <div className="flex items-center space-x-1.5">
                            <span className="font-bold text-slate-800 text-sm">{monthLabel}</span>
                            <span className="text-slate-400 font-normal text-xs">
                              ({m.drives_count} {m.drives_count === 1 ? 'trip' : 'trips'})
                            </span>
                            <span className="text-[10px] text-slate-400 sm:hidden">
                              {isExpanded ? (
                                <ChevronUp className="w-3 h-3 inline" />
                              ) : (
                                <ChevronDown className="w-3 h-3 inline" />
                              )}
                            </span>
                          </div>
                          <div className="text-right">
                            <span className="font-bold text-slate-900">
                              {m.total_fradrag.toFixed(2)} DKK
                            </span>
                          </div>
                        </div>

                        {/* Relative Tooltip Container for PC Hover */}
                        <div className="relative">
                          {hoveredRoute && hoveredRoute.month === m.month_name && (
                            <div className="hidden sm:flex absolute -top-9 left-1/2 -translate-x-1/2 z-20 bg-slate-900 text-white text-[11px] px-2.5 py-1 rounded-lg shadow-lg pointer-events-none items-center space-x-2 whitespace-nowrap">
                              <span className="font-semibold">{hoveredRoute.routeName}:</span>
                              <span>
                                {hoveredRoute.trips} {hoveredRoute.trips === 1 ? 'trip' : 'trips'} ({hoveredRoute.pct}%)
                              </span>
                              <span className="text-slate-400">•</span>
                              <span className="text-brand-300 font-medium">{hoveredRoute.fradrag.toFixed(2)} DKK</span>
                            </div>
                          )}

                          {/* Multi-part Segmented Bar */}
                          <div
                            onClick={() => setActiveExpandedMonth(isExpanded ? null : m.month_name)}
                            className="w-full bg-slate-100 rounded-full h-3.5 overflow-hidden cursor-pointer p-0.5 border border-slate-200/50 hover:border-slate-300 transition-colors"
                          >
                            <div
                              className="h-full rounded-full overflow-hidden flex transition-all duration-500 shadow-sm"
                              style={{ width: `${Math.max(pct, 6)}%` }}
                            >
                              {m.routes && m.routes.length > 0 ? (
                                m.routes.map((route, rIdx) => {
                                  const routeSharePct =
                                    m.drives_count > 0 ? (route.drives_count / m.drives_count) * 100 : 0;
                                  const palette = getRoutePalette(route.route_id, rIdx);
                                  const isSegmentHovered =
                                    hoveredRoute?.month === m.month_name &&
                                    hoveredRoute?.routeId === route.route_id;

                                  return (
                                    <div
                                      key={route.route_id || rIdx}
                                      style={{ width: `${routeSharePct}%` }}
                                      className={`h-full ${palette.bg} border-r border-white/20 last:border-r-0 transition-all duration-200 relative ${
                                        isSegmentHovered ? 'brightness-110 ring-1 ring-white/50' : ''
                                      }`}
                                      onMouseEnter={() =>
                                        setHoveredRoute({
                                          month: m.month_name,
                                          routeId: route.route_id,
                                          routeName: route.route_name,
                                          trips: route.drives_count,
                                          pct: Math.round(routeSharePct),
                                          fradrag: route.total_fradrag,
                                          km: route.distance_km,
                                        })
                                      }
                                      onMouseLeave={() => setHoveredRoute(null)}
                                      title={`${route.route_name}: ${route.drives_count} trips (${Math.round(routeSharePct)}%) • ${route.total_fradrag.toFixed(2)} DKK`}
                                    />
                                  );
                                })
                              ) : (
                                <div className="h-full w-full bg-brand-600" />
                              )}
                            </div>
                          </div>
                        </div>

                        {/* Distance traveled summary & mobile tap trigger */}
                        <div className="flex justify-between items-center text-[10px] text-slate-400 mt-1">
                          <span>{m.distance_km.toFixed(1)} km traveled</span>
                          <button
                            type="button"
                            onClick={() => setActiveExpandedMonth(isExpanded ? null : m.month_name)}
                            className="text-brand-600 hover:text-brand-700 font-medium sm:hidden flex items-center space-x-0.5"
                          >
                            <span>{isExpanded ? 'Hide routes' : 'View routes'}</span>
                            {isExpanded ? (
                              <ChevronUp className="w-3 h-3" />
                            ) : (
                              <ChevronDown className="w-3 h-3" />
                            )}
                          </button>
                        </div>

                        {/* Mobile & Tap-to-inspect Route Breakdown Drawer */}
                        {isExpanded && m.routes && m.routes.length > 0 && (
                          <div className="mt-2.5 p-3 bg-slate-50 border border-slate-200/80 rounded-xl space-y-2">
                            <div className="text-[10px] font-bold uppercase tracking-wider text-slate-400 flex justify-between pb-1 border-b border-slate-200/60">
                              <span>Routes in {monthLabel}</span>
                              <span>Trips / Deduction</span>
                            </div>
                            {m.routes.map((route, rIdx) => {
                              const palette = getRoutePalette(route.route_id, rIdx);
                              const routeSharePct =
                                m.drives_count > 0 ? Math.round((route.drives_count / m.drives_count) * 100) : 0;
                              return (
                                <div
                                  key={route.route_id || rIdx}
                                  className="flex items-center justify-between text-xs py-0.5"
                                >
                                  <div className="flex items-center space-x-2 truncate pr-2">
                                    <span className={`w-2.5 h-2.5 rounded-full ${palette.bg} shrink-0`} />
                                    <span className="font-semibold text-slate-800 truncate">
                                      {route.route_name}
                                    </span>
                                  </div>
                                  <div className="flex items-center space-x-2 shrink-0">
                                    <span className="text-slate-500 text-[11px]">
                                      {route.total_fradrag.toFixed(2)} DKK
                                    </span>
                                    <span className="font-bold text-slate-800 bg-white px-2 py-0.5 rounded-md border border-slate-200 text-[11px]">
                                      {route.drives_count} {route.drives_count === 1 ? 'trip' : 'trips'} ({routeSharePct}%)
                                    </span>
                                  </div>
                                </div>
                              );
                            })}
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              )}
            </div>

            {/* Route & Destination Breakdown */}
            <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-sm">
              <h3 className="text-sm font-bold text-slate-900 mb-4 flex items-center space-x-2">
                <Navigation className="w-4 h-4 text-brand-600" />
                <span>Commute by Destination (Work vs University)</span>
              </h3>

              {!stats?.route_breakdown || stats.route_breakdown.length === 0 ? (
                <div className="text-xs text-slate-400 py-8 text-center">
                  No route activity recorded for {selectedYear}.
                </div>
              ) : (
                <div className="space-y-4">
                  {stats.route_breakdown.map((r) => (
                    <div
                      key={r.route_id}
                      className="p-3 bg-slate-50 border border-slate-200/60 rounded-xl flex items-center justify-between"
                    >
                      <div className="flex items-center space-x-3">
                        <div
                          className={`w-8 h-8 rounded-lg flex items-center justify-center ${
                            r.destination_type === 'university'
                              ? 'bg-purple-100 text-purple-700'
                              : 'bg-blue-100 text-blue-700'
                          }`}
                        >
                          {r.destination_type === 'university' ? (
                            <GraduationCap className="w-4 h-4" />
                          ) : (
                            <Building2 className="w-4 h-4" />
                          )}
                        </div>
                        <div>
                          <div className="text-xs font-bold text-slate-900">{r.route_name}</div>
                          <div className="text-[10px] text-slate-500">
                            {r.destination_type === 'university' ? 'University' : 'Workplace'} • {r.drives_count} drives • {r.distance_km.toFixed(1)} km
                          </div>
                        </div>
                      </div>

                      <div className="text-right">
                        <div className="text-xs font-bold text-brand-700">
                          {r.total_fradrag.toFixed(2)} DKK
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Drives History Table */}
          <div className="bg-white rounded-2xl border border-slate-200/80 shadow-sm overflow-hidden">
            <div className="p-5 border-b border-slate-200 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
              <h3 className="font-bold text-sm text-slate-900">Registered Drives History</h3>

              <div className="flex items-center space-x-2">
                <Filter className="w-3.5 h-3.5 text-slate-400" />
                <select
                  value={selectedRouteFilter}
                  onChange={(e) => setSelectedRouteFilter(e.target.value)}
                  className="text-xs border border-slate-200 rounded-lg px-2.5 py-1.5 focus:outline-none focus:ring-1 focus:ring-brand-500 text-slate-700"
                >
                  <option value="">All Routes</option>
                  {routes.map((r) => (
                    <option key={r.id} value={r.id}>
                      {r.name}
                    </option>
                  ))}
                </select>
              </div>
            </div>

            {drives.length === 0 ? (
              <div className="py-12 text-center text-xs text-slate-400">
                No drives match the selected filters.
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead className="bg-slate-50 text-slate-500 uppercase tracking-wider text-[10px] border-b border-slate-200">
                    <tr>
                      <th className="py-3 px-4">Date</th>
                      <th className="py-3 px-4">Route</th>
                      <th className="py-3 px-4">Distance (Round Trip)</th>
                      <th className="py-3 px-4">Calculated Fradrag</th>
                      <th className="py-3 px-4">Notes</th>
                      <th className="py-3 px-4">Logged By</th>
                      {canWrite && <th className="py-3 px-4 text-right">Action</th>}
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100 text-slate-700">
                    {drives.map((drive) => (
                      <tr key={drive.id} className="hover:bg-slate-50/80 transition-colors">
                        <td className="py-3 px-4 font-semibold text-slate-900 whitespace-nowrap">
                          {drive.date}
                        </td>
                        <td className="py-3 px-4 whitespace-nowrap">
                          <span className="font-medium text-slate-800">{drive.route_name}</span>
                        </td>
                        <td className="py-3 px-4 whitespace-nowrap">
                          {drive.distance_km} km
                        </td>
                        <td className="py-3 px-4 whitespace-nowrap font-bold text-brand-700">
                          {drive.calculated_fradrag.toFixed(2)} DKK
                        </td>
                        <td className="py-3 px-4 max-w-[200px] truncate text-slate-500">
                          {drive.notes || '—'}
                        </td>
                        <td className="py-3 px-4 whitespace-nowrap text-slate-500">
                          {drive.registered_by_name}
                        </td>
                        {canWrite && (
                          <td className="py-3 px-4 text-right whitespace-nowrap">
                            <button
                              onClick={() => handleDeleteDrive(drive.id)}
                              className="text-slate-400 hover:text-red-600 transition-colors p-1"
                              title="Delete drive"
                            >
                              <Trash2 className="w-4 h-4" />
                            </button>
                          </td>
                        )}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
};
