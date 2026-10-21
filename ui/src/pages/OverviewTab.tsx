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
} from 'lucide-react';

export const OverviewTab: React.FC = () => {
  const { activePermission, activeAccountName } = useAuth();
  const [selectedYear, setSelectedYear] = useState<string>(new Date().getFullYear().toString());
  const [stats, setStats] = useState<OverviewStats | null>(null);
  const [drives, setDrives] = useState<Drive[]>([]);
  const [routes, setRoutes] = useState<Route[]>([]);
  const [selectedRouteFilter, setSelectedRouteFilter] = useState<string>('');
  const [isLoading, setIsLoading] = useState<boolean>(true);

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
              <h3 className="text-sm font-bold text-slate-900 mb-4 flex items-center space-x-2">
                <Calendar className="w-4 h-4 text-brand-600" />
                <span>Monthly Deduction Breakdown</span>
              </h3>

              {!stats?.monthly_breakdown || stats.monthly_breakdown.length === 0 ? (
                <div className="text-xs text-slate-400 py-8 text-center">
                  No monthly activity recorded for {selectedYear}.
                </div>
              ) : (
                <div className="space-y-3">
                  {stats.monthly_breakdown.map((m) => {
                    const maxFradrag = Math.max(
                      ...stats.monthly_breakdown.map((item) => item.total_fradrag),
                      1
                    );
                    const pct = Math.min(100, Math.round((m.total_fradrag / maxFradrag) * 100));

                    return (
                      <div key={m.month_name} className="text-xs">
                        <div className="flex justify-between font-semibold text-slate-700 mb-1">
                          <span>{m.month_name} ({m.drives_count} drives)</span>
                          <span className="font-bold text-slate-900">
                            {m.total_fradrag.toFixed(2)} DKK
                          </span>
                        </div>
                        <div className="w-full bg-slate-100 rounded-full h-2.5 overflow-hidden">
                          <div
                            className="bg-brand-600 h-2.5 rounded-full transition-all duration-500"
                            style={{ width: `${pct}%` }}
                          ></div>
                        </div>
                        <div className="text-[10px] text-slate-400 mt-0.5">
                          {m.distance_km.toFixed(1)} km traveled
                        </div>
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
