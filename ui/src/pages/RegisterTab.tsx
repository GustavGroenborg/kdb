import React, { useState, useEffect } from 'react';
import { Route, Drive, TaxRate } from '../types';
import { api } from '../services/api';
import { useAuth } from '../context/AuthContext';
import {
  Car,
  Calendar,
  CheckCircle2,
  Clock,
  Sparkles,
  AlertTriangle,
  PlusCircle,
  Trash2,
  MapPin,
  Eye,
  EyeOff,
} from 'lucide-react';

interface RegisterTabProps {
  onGoToSettings: () => void;
}

export const RegisterTab: React.FC<RegisterTabProps> = ({ onGoToSettings }) => {
  const { activePermission, activeAccountName, isViewingSharedAccount, activeAccountId } = useAuth();
  const [routes, setRoutes] = useState<Route[]>([]);
  const [selectedRouteId, setSelectedRouteId] = useState<string>('');
  const [date, setDate] = useState<string>(new Date().toISOString().split('T')[0]);
  const [notes, setNotes] = useState<string>('');
  const [distanceKmOverride, setDistanceKmOverride] = useState<string>('');
  const [rates, setRates] = useState<TaxRate | null>(null);
  const [todayDrives, setTodayDrives] = useState<Drive[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  const canWrite = activePermission === 'write';

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const [fetchedRoutes, fetchedRates, fetchedDrives] = await Promise.all([
        api.getRoutes(),
        api.getRates(),
        api.getDrives({ date, limit: 10 }),
      ]);

      setRoutes(fetchedRoutes);
      setRates(fetchedRates);
      setTodayDrives(fetchedDrives.filter((d) => d.date === date));

      // Auto-select default route or first visible route
      const visible = fetchedRoutes.filter((r) => !r.is_hidden);
      if (visible.length > 0) {
        const defaultRoute = visible.find((r) => r.is_default) || visible[0];
        setSelectedRouteId(defaultRoute.id);
      } else {
        setSelectedRouteId('');
      }
    } catch (err: any) {
      setErrorMsg(err.message || 'Failed to load routes');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [date, activeAccountId, activeAccountName]);

  const visibleRoutes = routes.filter((r) => !r.is_hidden);
  const selectedRoute = visibleRoutes.find((r) => r.id === selectedRouteId) || routes.find((r) => r.id === selectedRouteId);

  // Compute live distance
  const currentDistance = distanceKmOverride
    ? parseFloat(distanceKmOverride) || 0
    : selectedRoute?.distance_round_trip_km || 0;

  // Calculate live deduction preview
  const calculateLiveFradrag = (km: number) => {
    if (!rates || km <= rates.threshold_km) return 0;
    if (km <= rates.tier2_threshold_km) {
      return (km - rates.threshold_km) * rates.base_rate;
    }
    const tier1 = (rates.tier2_threshold_km - rates.threshold_km) * rates.base_rate;
    const tier2 = (km - rates.tier2_threshold_km) * rates.tier2_rate;
    return Math.round((tier1 + tier2) * 100) / 100;
  };

  const previewFradrag = calculateLiveFradrag(currentDistance);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canWrite) return;
    if (!selectedRouteId) {
      setErrorMsg('Please select a route');
      return;
    }

    setIsSubmitting(true);
    setErrorMsg(null);
    setSuccessMsg(null);

    try {
      await api.createDrive({
        route_id: selectedRouteId,
        date,
        distance_km: distanceKmOverride ? parseFloat(distanceKmOverride) : undefined,
        notes: notes.trim(),
      });

      setSuccessMsg(`Drive registered! Calculated fradrag: ${previewFradrag.toFixed(2)} DKK`);
      setNotes('');
      setDistanceKmOverride('');

      // Refresh today's list
      const updatedDrives = await api.getDrives({ date, limit: 10 });
      setTodayDrives(updatedDrives.filter((d) => d.date === date));
    } catch (err: any) {
      setErrorMsg(err.message || 'Failed to register drive');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteDrive = async (id: string) => {
    if (!canWrite) return;
    try {
      await api.deleteDrive(id);
      setTodayDrives((prev) => prev.filter((d) => d.id !== id));
    } catch (err: any) {
      setErrorMsg(err.message || 'Failed to delete drive');
    }
  };

  if (isLoading) {
    return (
      <div className="max-w-4xl mx-auto py-12 px-4 text-center text-slate-500">
        <div className="animate-spin w-8 h-8 border-4 border-brand-500 border-t-transparent rounded-full mx-auto mb-3"></div>
        Loading route configuration...
      </div>
    );
  }

  // 1. If NO routes exist: Zero Routes Guard Onboarding
  if (routes.length === 0) {
    return (
      <div className="max-w-2xl mx-auto py-12 px-4">
        <div className="bg-white rounded-2xl p-8 border border-slate-200 shadow-sm text-center">
          <div className="w-16 h-16 bg-brand-50 text-brand-600 rounded-2xl flex items-center justify-center mx-auto mb-4">
            <PlusCircle className="w-8 h-8" />
          </div>
          <h2 className="text-2xl font-bold text-slate-900 mb-2">No Routes Configured Yet</h2>
          <p className="text-sm text-slate-600 mb-6 max-w-md mx-auto">
            To register drives and claim befordringsfradrag, you need to configure your Home address,
            add your Workplace or University, and set up your first route.
          </p>
          <button
            onClick={onGoToSettings}
            className="inline-flex items-center space-x-2 px-5 py-2.5 bg-brand-600 hover:bg-brand-700 text-white font-medium text-sm rounded-xl shadow transition-all"
          >
            <PlusCircle className="w-4 h-4" />
            <span>Go to Settings to Create Route</span>
          </button>
        </div>
      </div>
    );
  }

  // 1b. If routes exist but ALL are hidden:
  if (visibleRoutes.length === 0) {
    return (
      <div className="max-w-2xl mx-auto py-12 px-4">
        <div className="bg-white rounded-2xl p-8 border border-slate-200 shadow-sm text-center">
          <div className="w-16 h-16 bg-amber-50 text-amber-600 rounded-2xl flex items-center justify-center mx-auto mb-4">
            <EyeOff className="w-8 h-8" />
          </div>
          <h2 className="text-2xl font-bold text-slate-900 mb-2">All Routes Are Hidden</h2>
          <p className="text-sm text-slate-600 mb-6 max-w-md mx-auto">
            You have configured routes, but all of them are currently hidden from the main page.
            You can make them visible again anytime in Settings.
          </p>
          <button
            onClick={onGoToSettings}
            className="inline-flex items-center space-x-2 px-5 py-2.5 bg-brand-600 hover:bg-brand-700 text-white font-medium text-sm rounded-xl shadow transition-all"
          >
            <Eye className="w-4 h-4" />
            <span>Go to Settings to Show Routes</span>
          </button>
        </div>
      </div>
    );
  }

  // 2. Default Page: Routes exist -> Prompt user to register a drive on that route
  return (
    <div className="max-w-4xl mx-auto py-8 px-4 sm:px-6">
      {/* Shared Account Banner */}
      {isViewingSharedAccount && (
        <div className="mb-6 p-3 bg-amber-50 border border-amber-200 rounded-xl flex items-center justify-between text-xs text-amber-800">
          <div className="flex items-center space-x-2">
            <AlertTriangle className="w-4 h-4 shrink-0 text-amber-600" />
            <span>
              You are currently viewing <strong>{activeAccountName}</strong>.
              {!canWrite && ' (Read-only mode: drive logging is disabled)'}
            </span>
          </div>
        </div>
      )}

      {/* Main Drive Registration Card */}
      <div className="bg-white rounded-2xl border border-slate-200 shadow-sm overflow-hidden mb-8">
        <div className="bg-gradient-to-r from-brand-600 to-brand-700 p-6 text-white">
          <div className="flex items-center space-x-3">
            <div className="w-10 h-10 rounded-xl bg-white/10 flex items-center justify-center backdrop-blur-sm">
              <Car className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-xl font-bold">Register Drive (Back & Forth)</h2>
              <p className="text-xs text-brand-100">
                A single registered drive counts for both directions of your commute.
              </p>
            </div>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="p-4 sm:p-6 space-y-6">
          {successMsg && (
            <div className="p-3.5 bg-emerald-50 border border-emerald-200 text-emerald-800 rounded-xl text-xs flex items-center space-x-2">
              <CheckCircle2 className="w-4 h-4 shrink-0 text-emerald-600" />
              <span>{successMsg}</span>
            </div>
          )}

          {errorMsg && (
            <div className="p-3.5 bg-red-50 border border-red-200 text-red-800 rounded-xl text-xs flex items-center space-x-2">
              <AlertTriangle className="w-4 h-4 shrink-0 text-red-600" />
              <span>{errorMsg}</span>
            </div>
          )}

          {/* Route Selection */}
          <div>
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-2">
              Select Route
            </label>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {visibleRoutes.map((route) => {
                const isSelected = route.id === selectedRouteId;
                return (
                  <div
                    key={route.id}
                    onClick={() => setSelectedRouteId(route.id)}
                    className={`p-3.5 rounded-xl border cursor-pointer transition-all flex flex-col justify-between ${
                      isSelected
                        ? 'border-brand-600 bg-brand-50/60 ring-2 ring-brand-500/20'
                        : 'border-slate-200 hover:border-slate-300 hover:bg-slate-50'
                    }`}
                  >
                    <div className="flex items-start justify-between">
                      <div className="font-semibold text-sm text-slate-900">{route.name}</div>
                      {route.is_default && (
                        <span className="text-[10px] uppercase font-bold px-1.5 py-0.5 rounded bg-brand-100 text-brand-700">
                          Default
                        </span>
                      )}
                    </div>
                    <div className="mt-2 flex items-center text-xs text-slate-500 space-x-2">
                      <MapPin className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                      <span className="truncate">
                        {route.origin_location_name} ↔ {route.destination_location_name}
                      </span>
                    </div>
                    <div className="mt-2 text-xs font-medium text-slate-700 flex justify-between">
                      <span>Round trip:</span>
                      <span className="font-bold text-slate-900">{route.distance_round_trip_km} km</span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {/* Date Selection */}
            <div className="min-w-0">
              <label className="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-2">
                Commute Date
              </label>
              <div className="relative min-w-0">
                <Calendar className="w-4 h-4 text-slate-400 absolute left-3 top-3 pointer-events-none z-10" />
                <input
                  type="date"
                  value={date}
                  onChange={(e) => setDate(e.target.value)}
                  className="w-full min-w-0 max-w-full block pl-9 pr-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-brand-500 bg-white"
                />
              </div>
              <div className="mt-2 flex space-x-2 text-xs">
                <button
                  type="button"
                  onClick={() => setDate(new Date().toISOString().split('T')[0])}
                  className="px-2 py-1 bg-slate-100 hover:bg-slate-200 rounded-lg text-slate-700"
                >
                  Today
                </button>
                <button
                  type="button"
                  onClick={() => {
                    const y = new Date();
                    y.setDate(y.getDate() - 1);
                    setDate(y.toISOString().split('T')[0]);
                  }}
                  className="px-2 py-1 bg-slate-100 hover:bg-slate-200 rounded-lg text-slate-700"
                >
                  Yesterday
                </button>
              </div>
            </div>

            {/* Custom Distance Override */}
            <div className="min-w-0">
              <label className="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-2">
                Distance Override (Optional)
              </label>
              <input
                type="number"
                step="0.1"
                min="0"
                value={distanceKmOverride}
                onChange={(e) => setDistanceKmOverride(e.target.value)}
                placeholder={`Default: ${selectedRoute?.distance_round_trip_km || 0} km`}
                className="w-full min-w-0 max-w-full block px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-brand-500"
              />
              <span className="text-[11px] text-slate-500 mt-1 block">
                Leave blank to use default round-trip distance.
              </span>
            </div>
          </div>

          {/* Notes */}
          <div>
            <label className="block text-xs font-bold uppercase tracking-wider text-slate-500 mb-2">
              Notes (Optional)
            </label>
            <input
              type="text"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="e.g. Office work, University exam, Lab day"
              className="w-full px-3 py-2 text-sm border border-slate-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-brand-500"
            />
          </div>

          {/* Live Fradrag Calculation Card */}
          <div className="p-4 bg-slate-50 border border-slate-200/80 rounded-xl flex items-center justify-between">
            <div>
              <div className="flex items-center space-x-1.5 text-xs font-semibold text-slate-600">
                <Sparkles className="w-4 h-4 text-amber-500" />
                <span>Calculated Befordringsfradrag:</span>
              </div>
              <div className="text-2xl font-black text-brand-700 tracking-tight mt-0.5">
                {previewFradrag.toFixed(2)}{' '}
                <span className="text-sm font-semibold text-slate-500">DKK</span>
              </div>
              <div className="text-[11px] text-slate-500">
                Round-trip distance: <strong>{currentDistance} km</strong>
                {currentDistance <= (rates?.threshold_km || 24)
                  ? ' (Below 24 km threshold: 0 DKK)'
                  : ` (${currentDistance - (rates?.threshold_km || 24)} km eligible @ ${rates?.base_rate || 2.23} DKK/km)`}
              </div>
            </div>

            <button
              type="submit"
              disabled={isSubmitting || !canWrite}
              className="py-3 px-6 bg-brand-600 hover:bg-brand-700 disabled:opacity-50 text-white font-bold text-sm rounded-xl shadow-md transition-all flex items-center space-x-2"
            >
              <Car className="w-4 h-4" />
              <span>{isSubmitting ? 'Registering...' : 'Register Drive'}</span>
            </button>
          </div>
        </form>
      </div>

      {/* Drives Registered for this Date */}
      <div className="bg-white rounded-2xl border border-slate-200 shadow-sm p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-2">
            <Clock className="w-4 h-4 text-slate-500" />
            <h3 className="font-bold text-sm text-slate-800">
              Drives Logged for {date === new Date().toISOString().split('T')[0] ? 'Today' : date}
            </h3>
          </div>
          <span className="text-xs font-semibold text-slate-500">
            {todayDrives.length} {todayDrives.length === 1 ? 'drive' : 'drives'}
          </span>
        </div>

        {todayDrives.length === 0 ? (
          <div className="text-center py-6 text-xs text-slate-400">
            No drives registered for this date yet.
          </div>
        ) : (
          <div className="divide-y divide-slate-100">
            {todayDrives.map((drive) => (
              <div key={drive.id} className="py-3 flex items-center justify-between">
                <div>
                  <div className="text-xs font-bold text-slate-900">{drive.route_name}</div>
                  <div className="text-[11px] text-slate-500">
                    {drive.distance_km} km round trip • {drive.calculated_fradrag.toFixed(2)} DKK
                    {drive.notes && ` • "${drive.notes}"`}
                  </div>
                </div>
                {canWrite && (
                  <button
                    onClick={() => handleDeleteDrive(drive.id)}
                    className="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                    title="Delete drive"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
