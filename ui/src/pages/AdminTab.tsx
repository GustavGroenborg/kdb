import React, { useState, useEffect } from 'react';
import { api } from '../services/api';
import { AdminMetrics } from '../types';
import {
  Users,
  ShieldCheck,
  Activity,
  Calendar,
  Car,
  TrendingUp,
  MapPin,
  RefreshCw,
  Lock,
  CheckCircle2,
  AlertCircle,
  Clock,
} from 'lucide-react';

export const AdminTab: React.FC = () => {
  const [metrics, setMetrics] = useState<AdminMetrics | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [cohortPage, setCohortPage] = useState<number>(1);
  const cohortsPerPage = 10;

  const fetchMetrics = async () => {
    setIsLoading(true);
    setErrorMsg(null);
    try {
      const data = await api.getAdminMetrics();
      setMetrics(data);
    } catch (err: any) {
      setErrorMsg(err.message || 'Failed to load administrative telemetry.');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchMetrics();
  }, []);

  if (isLoading && !metrics) {
    return (
      <div className="max-w-6xl mx-auto py-16 px-4 text-center text-slate-500">
        <div className="animate-spin w-8 h-8 border-4 border-purple-600 border-t-transparent rounded-full mx-auto mb-3"></div>
        <p className="text-sm font-medium">Loading aggregated telemetry...</p>
      </div>
    );
  }

  const userCohorts = metrics?.user_cohorts || [];
  const totalPages = Math.ceil(userCohorts.length / cohortsPerPage) || 1;
  const paginatedCohorts = userCohorts.slice((cohortPage - 1) * cohortsPerPage, cohortPage * cohortsPerPage);

  return (
    <div className="max-w-6xl mx-auto py-8 px-4 sm:px-6 lg:px-8 space-y-8 animate-fade-in">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between pb-6 border-b border-slate-200 gap-4">
        <div>
          <div className="flex items-center space-x-2.5">
            <div className="w-9 h-9 rounded-xl bg-purple-100 text-purple-700 flex items-center justify-center font-bold">
              <ShieldCheck className="w-5 h-5" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-slate-900 tracking-tight flex items-center gap-2.5">
                <span>Platform Telemetry & Usage</span>
                <span className="px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-100 text-emerald-800 border border-emerald-200 flex items-center gap-1">
                  <CheckCircle2 className="w-3.5 h-3.5" />
                  GDPR Art. 5(1)(c) Compliant
                </span>
              </h1>
              <p className="text-xs text-slate-500 mt-1">
                Real-time capacity, active user metrics, and commute adoption statistics in aggregated format (Zero PII).
              </p>
            </div>
          </div>
        </div>

        <button
          onClick={fetchMetrics}
          disabled={isLoading}
          className="inline-flex items-center justify-center space-x-2 px-3.5 py-2 text-xs font-semibold text-purple-700 bg-purple-50 hover:bg-purple-100 border border-purple-200 rounded-xl transition-all self-start sm:self-auto disabled:opacity-50"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin' : ''}`} />
          <span>Refresh Telemetry</span>
        </button>
      </div>

      {errorMsg && (
        <div className="p-4 bg-red-50 border border-red-200 text-red-800 rounded-2xl text-xs flex items-center space-x-2.5">
          <AlertCircle className="w-4 h-4 text-red-600 shrink-0" />
          <span>{errorMsg}</span>
        </div>
      )}

      {/* 1. Active Users & Growth KPIs */}
      <div>
        <h2 className="text-xs font-bold uppercase tracking-wider text-slate-400 mb-3 flex items-center space-x-1.5">
          <Users className="w-3.5 h-3.5" />
          <span>User Adoption & Active User Telemetry</span>
        </h2>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          {/* Total Registered Users */}
          <div className="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-xs">
            <div className="flex items-center justify-between text-slate-500 text-xs font-medium mb-1">
              <span>Total Accounts</span>
              <Users className="w-4 h-4 text-slate-400" />
            </div>
            <div className="text-2xl sm:text-3xl font-extrabold text-slate-900 tracking-tight">
              {metrics?.total_users ?? 0}
            </div>
            <div className="text-[11px] text-slate-400 mt-1">Total registered users</div>
          </div>

          {/* 30-Day Active Users (MAU) */}
          <div className="bg-white p-5 rounded-2xl border border-purple-200/80 shadow-xs bg-gradient-to-br from-white to-purple-50/30">
            <div className="flex items-center justify-between text-purple-700 text-xs font-medium mb-1">
              <span>Monthly Active (MAU)</span>
              <Activity className="w-4 h-4 text-purple-500" />
            </div>
            <div className="text-2xl sm:text-3xl font-extrabold text-purple-950 tracking-tight">
              {metrics?.active_users_30d ?? 0}
            </div>
            <div className="text-[11px] text-purple-700/80 mt-1">Active in past 30 days</div>
          </div>

          {/* 7-Day Active Users (WAU) */}
          <div className="bg-white p-5 rounded-2xl border border-blue-200/80 shadow-xs bg-gradient-to-br from-white to-blue-50/30">
            <div className="flex items-center justify-between text-blue-700 text-xs font-medium mb-1">
              <span>Weekly Active (WAU)</span>
              <Calendar className="w-4 h-4 text-blue-500" />
            </div>
            <div className="text-2xl sm:text-3xl font-extrabold text-blue-950 tracking-tight">
              {metrics?.active_users_7d ?? 0}
            </div>
            <div className="text-[11px] text-blue-700/80 mt-1">Active in past 7 days</div>
          </div>

          {/* Daily Active Users (DAU) */}
          <div className="bg-white p-5 rounded-2xl border border-emerald-200/80 shadow-xs bg-gradient-to-br from-white to-emerald-50/30">
            <div className="flex items-center justify-between text-emerald-700 text-xs font-medium mb-1">
              <span>Daily Active (DAU)</span>
              <TrendingUp className="w-4 h-4 text-emerald-500" />
            </div>
            <div className="text-2xl sm:text-3xl font-extrabold text-emerald-950 tracking-tight">
              {metrics?.active_users_today ?? 0}
            </div>
            <div className="text-[11px] text-emerald-700/80 mt-1">Active commuters today</div>
          </div>
        </div>
      </div>

      {/* 2. Platform Commuting Totals */}
      <div>
        <h2 className="text-xs font-bold uppercase tracking-wider text-slate-400 mb-3 flex items-center space-x-1.5">
          <Car className="w-3.5 h-3.5" />
          <span>Platform-Wide Commuting Activity</span>
        </h2>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <div className="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-xs">
            <div className="text-xs font-medium text-slate-500 mb-1">Total Drives Logged</div>
            <div className="text-xl sm:text-2xl font-bold text-slate-900">
              {metrics?.total_drives.toLocaleString() ?? 0}
            </div>
            <div className="text-[11px] text-slate-400 mt-1">Across all users</div>
          </div>

          <div className="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-xs">
            <div className="text-xs font-medium text-slate-500 mb-1">Total Commute Distance</div>
            <div className="text-xl sm:text-2xl font-bold text-brand-700">
              {metrics ? Math.round(metrics.total_distance_km).toLocaleString() : 0} <span className="text-sm font-semibold">km</span>
            </div>
            <div className="text-[11px] text-slate-400 mt-1">Total registered kilometers</div>
          </div>

          <div className="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-xs">
            <div className="text-xs font-medium text-slate-500 mb-1">Total Befordringsfradrag</div>
            <div className="text-xl sm:text-2xl font-bold text-emerald-700">
              {metrics ? Math.round(metrics.total_fradrag).toLocaleString() : 0} <span className="text-sm font-semibold">kr.</span>
            </div>
            <div className="text-[11px] text-slate-400 mt-1">Danish tax deduction sum</div>
          </div>

          <div className="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-xs">
            <div className="text-xs font-medium text-slate-500 mb-1">Active Commute Routes</div>
            <div className="text-xl sm:text-2xl font-bold text-slate-900">
              {metrics?.total_routes ?? 0}
            </div>
            <div className="text-[11px] text-slate-400 mt-1">Workplace & university routes</div>
          </div>
        </div>
      </div>

      {/* 3. Monthly Activity & New Signups Breakdown */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Monthly Commute Activity */}
        <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-bold text-slate-900 flex items-center space-x-2">
              <Activity className="w-4 h-4 text-purple-600" />
              <span>Monthly Activity (Last 12 Months)</span>
            </h3>
            <span className="text-[11px] text-slate-400">Aggregated by month</span>
          </div>

          {metrics?.monthly_activity && metrics.monthly_activity.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-slate-100 text-slate-400 font-semibold text-[11px]">
                    <th className="pb-2">Month</th>
                    <th className="pb-2 text-right">Active Users</th>
                    <th className="pb-2 text-right">Drives</th>
                    <th className="pb-2 text-right">Distance</th>
                    <th className="pb-2 text-right">Fradrag</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-50">
                  {metrics.monthly_activity.map((m) => (
                    <tr key={m.month_name} className="hover:bg-slate-50/60 transition-colors">
                      <td className="py-2.5 font-bold text-slate-900">{m.month_name}</td>
                      <td className="py-2.5 text-right font-medium text-purple-700">{m.active_users}</td>
                      <td className="py-2.5 text-right text-slate-600">{m.drives_count}</td>
                      <td className="py-2.5 text-right text-slate-600">{Math.round(m.distance_km)} km</td>
                      <td className="py-2.5 text-right font-semibold text-emerald-700">
                        {Math.round(m.total_fradrag).toLocaleString()} kr.
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="p-8 text-center text-xs text-slate-400 border border-dashed border-slate-200 rounded-xl">
              No drive activity recorded yet.
            </div>
          )}
        </div>

        {/* Municipal Geography Breakdown */}
        <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-bold text-slate-900 flex items-center space-x-2">
              <MapPin className="w-4 h-4 text-brand-600" />
              <span>Registered Municipalities (Kommuner)</span>
            </h3>
            <span className="text-[11px] text-slate-400">DAR Resolved</span>
          </div>

          {metrics?.kommune_distribution && metrics.kommune_distribution.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-slate-100 text-slate-400 font-semibold text-[11px]">
                    <th className="pb-2">Kommune</th>
                    <th className="pb-2">Status</th>
                    <th className="pb-2 text-right">Residences</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-50">
                  {metrics.kommune_distribution.map((k) => (
                    <tr key={k.kommune_name} className="hover:bg-slate-50/60 transition-colors">
                      <td className="py-2.5 font-bold text-slate-900">{k.kommune_name}</td>
                      <td className="py-2.5">
                        {k.is_yderkommune ? (
                          <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-100 text-emerald-800 border border-emerald-200">
                            🌟 Yderkommune
                          </span>
                        ) : (
                          <span className="px-2 py-0.5 rounded-full text-[10px] font-medium bg-slate-100 text-slate-600">
                            Standard
                          </span>
                        )}
                      </td>
                      <td className="py-2.5 text-right font-semibold text-slate-800">{k.user_count}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="p-8 text-center text-xs text-slate-400 border border-dashed border-slate-200 rounded-xl">
              No registered residences with municipality resolution yet.
            </div>
          )}
        </div>
      </div>

      {/* 4. Anonymized User Cohorts (Zero PII, Cryptographic Pseudonyms) */}
      <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-xs">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-4 gap-2">
          <div>
            <h3 className="text-sm font-bold text-slate-900 flex items-center space-x-2">
              <Lock className="w-4 h-4 text-emerald-600" />
              <span>Anonymized User Cohort Telemetry</span>
            </h3>
            <p className="text-[11px] text-slate-500 mt-0.5">
              Individual adoption levels pseudonomized using SHA-256 tokens. No names, emails, or locations are accessible.
            </p>
          </div>
          <span className="text-[11px] font-semibold text-slate-500 bg-slate-100 px-2.5 py-1 rounded-lg shrink-0">
            {userCohorts.length} Accounts
          </span>
        </div>

        {paginatedCohorts.length > 0 ? (
          <>
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-slate-100 text-slate-400 font-semibold text-[11px]">
                    <th className="pb-2.5">Pseudonym</th>
                    <th className="pb-2.5">Joined</th>
                    <th className="pb-2.5">Last Active</th>
                    <th className="pb-2.5 text-right">Commute Drives</th>
                    <th className="pb-2.5 text-right">Distance (km)</th>
                    <th className="pb-2.5 text-center">DAR Residence</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-50">
                  {paginatedCohorts.map((cohort) => (
                    <tr key={cohort.pseudo_id} className="hover:bg-slate-50/60 transition-colors">
                      <td className="py-3 font-mono font-bold text-purple-900 text-[11px]">
                        {cohort.pseudo_id}
                      </td>
                      <td className="py-3 text-slate-600">{cohort.signup_month}</td>
                      <td className="py-3 text-slate-600">
                        {cohort.last_active_days_ago === -1 ? (
                          <span className="text-slate-400 italic">No drives logged</span>
                        ) : cohort.last_active_days_ago === 0 ? (
                          <span className="font-semibold text-emerald-700">Today</span>
                        ) : (
                          <span>{cohort.last_active_days_ago} days ago</span>
                        )}
                      </td>
                      <td className="py-3 text-right font-semibold text-slate-900">
                        {cohort.total_drives}
                      </td>
                      <td className="py-3 text-right text-slate-600">
                        {Math.round(cohort.total_distance_km).toLocaleString()}
                      </td>
                      <td className="py-3 text-center">
                        {cohort.has_dar_residence ? (
                          <span className="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-blue-50 text-blue-700 border border-blue-200">
                            {cohort.is_yderkommune ? '✓ DAR (Yder)' : '✓ DAR Approved'}
                          </span>
                        ) : (
                          <span className="text-[10px] text-slate-400">Unverified</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination Controls */}
            {totalPages > 1 && (
              <div className="flex items-center justify-between pt-4 border-t border-slate-100 text-xs text-slate-500">
                <div>
                  Page {cohortPage} of {totalPages}
                </div>
                <div className="flex space-x-1.5">
                  <button
                    onClick={() => setCohortPage((p) => Math.max(p - 1, 1))}
                    disabled={cohortPage === 1}
                    className="px-2.5 py-1 border border-slate-200 rounded-lg text-slate-700 disabled:opacity-40 hover:bg-slate-50"
                  >
                    Previous
                  </button>
                  <button
                    onClick={() => setCohortPage((p) => Math.min(p + 1, totalPages))}
                    disabled={cohortPage === totalPages}
                    className="px-2.5 py-1 border border-slate-200 rounded-lg text-slate-700 disabled:opacity-40 hover:bg-slate-50"
                  >
                    Next
                  </button>
                </div>
              </div>
            )}
          </>
        ) : (
          <div className="p-8 text-center text-xs text-slate-400 border border-dashed border-slate-200 rounded-xl">
            No user cohorts available.
          </div>
        )}
      </div>

      {/* 5. GDPR Compliance & Privacy Architecture Notice */}
      <div className="bg-slate-900 text-slate-300 p-6 rounded-2xl shadow-sm text-xs space-y-3">
        <div className="flex items-center space-x-2 text-white font-bold text-sm">
          <ShieldCheck className="w-4 h-4 text-emerald-400" />
          <span>GDPR Compliance & Privacy Architecture (Art. 5(1)(c) & Recital 26)</span>
        </div>
        <p className="leading-relaxed text-slate-400">
          This administration console is engineered with <strong className="text-white">Privacy-by-Design</strong> and operates strictly on aggregated metrics in adherence with the European Union General Data Protection Regulation (GDPR).
        </p>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2 text-[11px]">
          <div className="p-3 bg-slate-800/80 rounded-xl border border-slate-700/60">
            <div className="font-semibold text-white mb-1">Data Minimization</div>
            <p className="text-slate-400 leading-relaxed">
              Zero PII is transmitted in administrative telemetry. No names, emails, addresses, or trip notes are included in payloads.
            </p>
          </div>
          <div className="p-3 bg-slate-800/80 rounded-xl border border-slate-700/60">
            <div className="font-semibold text-white mb-1">Irreversible Pseudonyms</div>
            <p className="text-slate-400 leading-relaxed">
              Individual cohorts use cryptographic truncations derived from internal keys to prevent identity correlation.
            </p>
          </div>
          <div className="p-3 bg-slate-800/80 rounded-xl border border-slate-700/60">
            <div className="font-semibold text-white mb-1">Self-Hosted Isolation</div>
            <p className="text-slate-400 leading-relaxed">
              All computations execute locally inside the KDB SQLite instance. No external analytics SDKs or third-party tracking cookies are utilized.
            </p>
          </div>
        </div>
        {metrics?.gdpr_audit && (
          <div className="pt-2 text-[10px] text-slate-500 flex items-center space-x-1.5">
            <Clock className="w-3 h-3 text-slate-400" />
            <span>
              Telemetry Audit Standard: <strong>{metrics.gdpr_audit.compliance_standard}</strong> (Audited at {metrics.gdpr_audit.audit_timestamp})
            </span>
          </div>
        )}
      </div>
    </div>
  );
};
