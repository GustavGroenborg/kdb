import React, { useState, useEffect } from 'react';
import { Location, Route, AccountShare, LocationType, SharePermission } from '../types';
import { api } from '../services/api';
import { useAuth } from '../context/AuthContext';
import {
  Home,
  Building2,
  GraduationCap,
  Plus,
  Trash2,
  Users,
  Shield,
  CheckCircle2,
  AlertCircle,
  Car,
  Pencil,
  X,
  Check,
  UserCog,
  KeyRound,
  Lock,
} from 'lucide-react';

export const SettingsTab: React.FC = () => {
  const { user, activePermission, activeAccountName, isViewingSharedAccount, refreshMe, updateUserSession } = useAuth();
  const canWrite = activePermission === 'write';

  const [locations, setLocations] = useState<Location[]>([]);
  const [routes, setRoutes] = useState<Route[]>([]);
  const [shares, setShares] = useState<{ my_shares: AccountShare[]; shared_with_me: AccountShare[] }>({
    my_shares: [],
    shared_with_me: [],
  });

  const [isLoading, setIsLoading] = useState(true);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  // Forms state
  // 1. Home Form
  const [homeName, setHomeName] = useState('My Residence');
  const [homeAddress, setHomeAddress] = useState('');

  // 2. Add Location Form
  const [locName, setLocName] = useState('');
  const [locType, setLocType] = useState<LocationType>('workplace');
  const [locAddress, setLocAddress] = useState('');

  // 3. Add / Edit Route Form
  const [editingRouteId, setEditingRouteId] = useState<string | null>(null);
  const [routeName, setRouteName] = useState('');
  const [routeDestId, setRouteDestId] = useState('');
  const [routeOneWayKm, setRouteOneWayKm] = useState('');
  const [routeRoundTripKm, setRouteRoundTripKm] = useState('');
  const [routeIsDefault, setRouteIsDefault] = useState(false);

  // 4. Share Form
  const [shareUserOrEmail, setShareUserOrEmail] = useState('');
  const [sharePermission, setSharePermission] = useState<SharePermission>('read');

  // 5. Account Settings (Profile & Password)
  const [profileUsername, setProfileUsername] = useState(user?.username || '');
  const [profileEmail, setProfileEmail] = useState(user?.email || '');
  const [isUpdatingProfile, setIsUpdatingProfile] = useState(false);
  const [profileSuccess, setProfileSuccess] = useState<string | null>(null);
  const [profileError, setProfileError] = useState<string | null>(null);

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isChangingPassword, setIsChangingPassword] = useState(false);
  const [passwordSuccess, setPasswordSuccess] = useState<string | null>(null);
  const [passwordError, setPasswordError] = useState<string | null>(null);

  useEffect(() => {
    if (user) {
      setProfileUsername(user.username);
      setProfileEmail(user.email);
    }
  }, [user]);

  const fetchData = async () => {
    setIsLoading(true);
    try {
      const [fetchedLocs, fetchedRoutes, fetchedShares] = await Promise.all([
        api.getLocations(),
        api.getRoutes(),
        api.getShares(),
      ]);

      setLocations(fetchedLocs);
      setRoutes(fetchedRoutes);
      setShares(fetchedShares);

      const home = fetchedLocs.find((l) => l.type === 'home');
      if (home) {
        setHomeName(home.name);
        setHomeAddress(home.address);
      }
    } catch (err: any) {
      setErrorMsg(err.message || 'Failed to load settings data');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [activeAccountName]);

  const showSuccess = (msg: string) => {
    setSuccessMsg(msg);
    setTimeout(() => setSuccessMsg(null), 4000);
  };

  // 1. Save Home
  const handleSaveHome = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canWrite) return;
    try {
      await api.createLocation({
        name: homeName.trim() || 'My Residence',
        type: 'home',
        address: homeAddress.trim(),
      });
      showSuccess('Home address updated successfully!');
      await fetchData();
      await refreshMe();
    } catch (err: any) {
      setErrorMsg(err.message || 'Failed to update home');
    }
  };

  // 2. Add Workplace or University
  const handleAddLocation = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canWrite) return;
    if (!locName || !locAddress) {
      setErrorMsg('Please provide location name and address');
      return;
    }
    try {
      await api.createLocation({
        name: locName.trim(),
        type: locType,
        address: locAddress.trim(),
      });
      setLocName('');
      setLocAddress('');
      showSuccess(`${locType === 'workplace' ? 'Workplace' : 'University'} added!`);
      await fetchData();
    } catch (err: any) {
      setErrorMsg(err.message || 'Failed to add location');
    }
  };

  // Delete Location
  const handleDeleteLocation = async (id: string) => {
    if (!canWrite) return;
    if (!confirm('Are you sure you want to delete this location?')) return;
    try {
      await api.deleteLocation(id);
      showSuccess('Location removed');
      await fetchData();
    } catch (err: any) {
      alert(err.message || 'Failed to delete location');
    }
  };

  // 3. Add / Edit Route
  const handleStartEditRoute = (route: Route) => {
    setEditingRouteId(route.id);
    setRouteName(route.name);
    setRouteDestId(route.destination_location_id);
    setRouteRoundTripKm(route.distance_round_trip_km.toString());
    setRouteOneWayKm(route.distance_one_way_km.toString());
    setRouteIsDefault(route.is_default);
    window.scrollTo({ top: 380, behavior: 'smooth' });
  };

  const handleCancelEditRoute = () => {
    setEditingRouteId(null);
    setRouteName('');
    setRouteDestId('');
    setRouteOneWayKm('');
    setRouteRoundTripKm('');
    setRouteIsDefault(false);
  };

  const handleSaveRoute = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canWrite) return;
    const homeLoc = locations.find((l) => l.type === 'home');
    if (!homeLoc) {
      setErrorMsg('Please configure your Home address before creating or editing a route');
      return;
    }
    if (!routeDestId) {
      setErrorMsg('Please select a destination workplace or university');
      return;
    }

    const oneWay = parseFloat(routeOneWayKm) || 0;
    const roundTrip = parseFloat(routeRoundTripKm) || (oneWay * 2);

    if (roundTrip <= 0) {
      setErrorMsg('Distance must be greater than 0 km');
      return;
    }

    try {
      if (editingRouteId) {
        await api.updateRoute(editingRouteId, {
          name: routeName.trim() || `Home ↔ Destination`,
          origin_location_id: homeLoc.id,
          destination_location_id: routeDestId,
          distance_one_way_km: oneWay || roundTrip / 2,
          distance_round_trip_km: roundTrip,
          is_default: routeIsDefault,
        });
        showSuccess('Route updated successfully!');
      } else {
        await api.createRoute({
          name: routeName.trim() || `Home ↔ Destination`,
          origin_location_id: homeLoc.id,
          destination_location_id: routeDestId,
          distance_one_way_km: oneWay || roundTrip / 2,
          distance_round_trip_km: roundTrip,
          is_default: routeIsDefault,
        });
        showSuccess('Route created successfully!');
      }

      handleCancelEditRoute();
      await fetchData();
      await refreshMe();
    } catch (err: any) {
      setErrorMsg(err.message || 'Failed to save route');
    }
  };

  // Delete Route
  const handleDeleteRoute = async (id: string) => {
    if (!canWrite) return;
    if (!confirm('Are you sure you want to delete this route?')) return;
    try {
      await api.deleteRoute(id);
      showSuccess('Route deleted');
      await fetchData();
      await refreshMe();
    } catch (err: any) {
      alert(err.message || 'Failed to delete route');
    }
  };

  // 4. Share Access
  const handleCreateShare = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!shareUserOrEmail.trim()) {
      setErrorMsg('Please enter a username or email to grant access');
      return;
    }
    try {
      await api.createShare({
        username_or_email: shareUserOrEmail.trim(),
        permission: sharePermission,
      });
      setShareUserOrEmail('');
      showSuccess(`Access granted to ${shareUserOrEmail}!`);
      await fetchData();
    } catch (err: any) {
      setErrorMsg(err.message || 'Failed to grant access');
    }
  };

  // Revoke Share
  const handleRevokeShare = async (id: string) => {
    if (!confirm('Revoke access for this user?')) return;
    try {
      await api.deleteShare(id);
      showSuccess('Access revoked');
      await fetchData();
    } catch (err: any) {
      alert(err.message || 'Failed to revoke access');
    }
  };

  // 5. Update Profile (Username & Email)
  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    setProfileError(null);
    setProfileSuccess(null);

    const trimmedUsername = profileUsername.trim();
    const trimmedEmail = profileEmail.trim();

    if (trimmedUsername.length < 3) {
      setProfileError('Username must be at least 3 characters');
      return;
    }
    if (!trimmedEmail.includes('@')) {
      setProfileError('Please enter a valid email address');
      return;
    }

    setIsUpdatingProfile(true);
    try {
      const res = await api.updateProfile({
        username: trimmedUsername,
        email: trimmedEmail,
      });
      updateUserSession(res.user, res.token);
      setProfileSuccess('Profile updated successfully!');
      setTimeout(() => setProfileSuccess(null), 4000);
    } catch (err: any) {
      setProfileError(err.message || 'Failed to update profile');
    } finally {
      setIsUpdatingProfile(false);
    }
  };

  // 6. Change Password
  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setPasswordError(null);
    setPasswordSuccess(null);

    if (!currentPassword) {
      setPasswordError('Please enter your current password');
      return;
    }
    if (newPassword.length < 6) {
      setPasswordError('New password must be at least 6 characters');
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordError('New password and confirmation do not match');
      return;
    }

    setIsChangingPassword(true);
    try {
      await api.changePassword({
        current_password: currentPassword,
        new_password: newPassword,
      });
      setPasswordSuccess('Password changed successfully!');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      setTimeout(() => setPasswordSuccess(null), 4000);
    } catch (err: any) {
      setPasswordError(err.message || 'Failed to change password');
    } finally {
      setIsChangingPassword(false);
    }
  };

  const home = locations.find((l) => l.type === 'home');
  const destinations = locations.filter((l) => l.type !== 'home');

  if (isLoading) {
    return (
      <div className="max-w-6xl mx-auto py-12 px-4 text-center text-slate-500">
        <div className="animate-spin w-8 h-8 border-4 border-brand-500 border-t-transparent rounded-full mx-auto mb-3"></div>
        Loading settings...
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
      <div className="pb-6 border-b border-slate-200 mb-8">
        <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Settings & Configuration</h1>
        <p className="text-xs text-slate-500 mt-1">
          Manage your home, workplaces, university campuses, routes, and account access delegation.
        </p>
      </div>

      {successMsg && (
        <div className="mb-6 p-3.5 bg-emerald-50 border border-emerald-200 text-emerald-800 rounded-xl text-xs flex items-center space-x-2">
          <CheckCircle2 className="w-4 h-4 shrink-0 text-emerald-600" />
          <span>{successMsg}</span>
        </div>
      )}

      {errorMsg && (
        <div className="mb-6 p-3.5 bg-red-50 border border-red-200 text-red-800 rounded-xl text-xs flex items-center space-x-2">
          <AlertCircle className="w-4 h-4 shrink-0 text-red-600" />
          <span>{errorMsg}</span>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* SECTION 1: HOME LOCATION */}
        <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-sm flex flex-col justify-between">
          <div>
            <div className="flex items-center space-x-2.5 mb-4 text-slate-900">
              <div className="w-8 h-8 rounded-lg bg-blue-50 text-brand-600 flex items-center justify-center">
                <Home className="w-4 h-4" />
              </div>
              <h2 className="text-base font-bold">Home Location (Single Home)</h2>
            </div>
            <p className="text-xs text-slate-500 mb-4">
              Your registered Danish tax residence. All routes originate from this home address.
            </p>

            <form onSubmit={handleSaveHome} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">
                  Home Label
                </label>
                <input
                  type="text"
                  value={homeName}
                  onChange={(e) => setHomeName(e.target.value)}
                  placeholder="e.g. My Flat"
                  disabled={!canWrite}
                  className="w-full px-3 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-brand-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">
                  Full Street Address
                </label>
                <input
                  type="text"
                  required
                  value={homeAddress}
                  onChange={(e) => setHomeAddress(e.target.value)}
                  placeholder="e.g. Nørrebrogade 10, 2200 København N"
                  disabled={!canWrite}
                  className="w-full px-3 py-2 text-xs border border-slate-200 rounded-xl focus:outline-none focus:ring-1 focus:ring-brand-500"
                />
              </div>

              {canWrite && (
                <button
                  type="submit"
                  className="px-4 py-2 bg-brand-600 hover:bg-brand-700 text-white rounded-xl text-xs font-semibold shadow-sm transition-all"
                >
                  {home ? 'Update Home Address' : 'Save Home Address'}
                </button>
              )}
            </form>
          </div>
        </div>

        {/* SECTION 2: WORKPLACES & UNIVERSITIES */}
        <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-sm">
          <div className="flex items-center space-x-2.5 mb-4 text-slate-900">
            <div className="w-8 h-8 rounded-lg bg-purple-50 text-purple-600 flex items-center justify-center">
              <Building2 className="w-4 h-4" />
            </div>
            <h2 className="text-base font-bold">Workplaces & Universities</h2>
          </div>
          <p className="text-xs text-slate-500 mb-4">
            Register your employers or educational institutions for commuting.
          </p>

          {/* Add Destination Form */}
          {canWrite && (
            <form onSubmit={handleAddLocation} className="p-3.5 bg-slate-50 rounded-xl mb-4 space-y-3">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label className="block text-[11px] font-semibold text-slate-600 mb-1">Name</label>
                  <input
                    type="text"
                    required
                    value={locName}
                    onChange={(e) => setLocName(e.target.value)}
                    placeholder="e.g. DTU Lyngby / Novo Nordisk"
                    className="w-full px-2.5 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none"
                  />
                </div>
                <div>
                  <label className="block text-[11px] font-semibold text-slate-600 mb-1">Type</label>
                  <select
                    value={locType}
                    onChange={(e) => setLocType(e.target.value as LocationType)}
                    className="w-full px-2.5 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none"
                  >
                    <option value="workplace">Workplace (Arbejdsplads)</option>
                    <option value="university">University (Uddannelsessted)</option>
                  </select>
                </div>
              </div>
              <div>
                <label className="block text-[11px] font-semibold text-slate-600 mb-1">Address</label>
                <input
                  type="text"
                  required
                  value={locAddress}
                  onChange={(e) => setLocAddress(e.target.value)}
                  placeholder="e.g. Anker Engelunds Vej 1, 2800 Kgs. Lyngby"
                  className="w-full px-2.5 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none"
                />
              </div>
              <button
                type="submit"
                className="w-full py-2 bg-slate-900 hover:bg-slate-800 text-white rounded-lg text-xs font-semibold transition-all flex items-center justify-center space-x-1.5"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>Add Destination</span>
              </button>
            </form>
          )}

          {/* Destinations List */}
          <div className="space-y-2 max-h-56 overflow-y-auto">
            {destinations.length === 0 ? (
              <div className="text-xs text-slate-400 text-center py-4">No workplaces or universities added yet.</div>
            ) : (
              destinations.map((loc) => (
                <div
                  key={loc.id}
                  className="p-2.5 border border-slate-200/80 rounded-xl flex items-center justify-between hover:bg-slate-50 transition-colors"
                >
                  <div className="flex items-center space-x-2.5">
                    <div
                      className={`w-7 h-7 rounded-lg flex items-center justify-center text-xs ${
                        loc.type === 'university'
                          ? 'bg-purple-100 text-purple-700'
                          : 'bg-blue-100 text-blue-700'
                      }`}
                    >
                      {loc.type === 'university' ? <GraduationCap className="w-3.5 h-3.5" /> : <Building2 className="w-3.5 h-3.5" />}
                    </div>
                    <div>
                      <div className="text-xs font-bold text-slate-900">{loc.name}</div>
                      <div className="text-[10px] text-slate-500 truncate max-w-[200px]">{loc.address}</div>
                    </div>
                  </div>
                  {canWrite && (
                    <button
                      onClick={() => handleDeleteLocation(loc.id)}
                      className="text-slate-400 hover:text-red-600 p-1"
                      title="Delete location"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  )}
                </div>
              ))
            )}
          </div>
        </div>

        {/* SECTION 3: ROUTES */}
        <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-sm lg:col-span-2">
          <div className="flex items-center space-x-2.5 mb-4 text-slate-900">
            <div className="w-8 h-8 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center">
              <Car className="w-4 h-4" />
            </div>
            <h2 className="text-base font-bold">Routes Configuration</h2>
          </div>
          <p className="text-xs text-slate-500 mb-4">
            Configure routes between your Home and destination. A drive on this route counts for both back and forth.
          </p>

          {/* Add / Edit Route Form */}
          {canWrite && (
            <form
              onSubmit={handleSaveRoute}
              className={`p-4 rounded-xl mb-6 space-y-4 transition-all border ${
                editingRouteId
                  ? 'bg-amber-50/70 border-amber-300 ring-2 ring-amber-400/20'
                  : 'bg-slate-50 border-slate-200/80'
              }`}
            >
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-slate-800 flex items-center space-x-1.5">
                  {editingRouteId ? (
                    <>
                      <Pencil className="w-3.5 h-3.5 text-amber-600" />
                      <span className="text-amber-900 font-bold">Edit Route: {routeName || 'Selected Route'}</span>
                    </>
                  ) : (
                    <>
                      <Plus className="w-3.5 h-3.5 text-brand-600" />
                      <span>Add New Route</span>
                    </>
                  )}
                </span>
                {editingRouteId && (
                  <button
                    type="button"
                    onClick={handleCancelEditRoute}
                    className="text-[11px] text-slate-600 hover:text-slate-900 flex items-center space-x-1 bg-white px-2 py-0.5 rounded-md border border-slate-200 transition-colors"
                  >
                    <X className="w-3 h-3" />
                    <span>Cancel Editing</span>
                  </button>
                )}
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                <div>
                  <label className="block text-[11px] font-semibold text-slate-600 mb-1">Route Name</label>
                  <input
                    type="text"
                    required
                    value={routeName}
                    onChange={(e) => setRouteName(e.target.value)}
                    placeholder="e.g. Home ↔ DTU"
                    className="w-full px-2.5 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-semibold text-slate-600 mb-1">Destination</label>
                  <select
                    required
                    value={routeDestId}
                    onChange={(e) => setRouteDestId(e.target.value)}
                    className="w-full px-2.5 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none"
                  >
                    <option value="">Select workplace or university</option>
                    {destinations.map((d) => (
                      <option key={d.id} value={d.id}>
                        {d.name} ({d.type})
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-[11px] font-semibold text-slate-600 mb-1">
                    Round-Trip Distance (km)
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    min="1"
                    required
                    value={routeRoundTripKm}
                    onChange={(e) => {
                      setRouteRoundTripKm(e.target.value);
                      if (e.target.value) {
                        setRouteOneWayKm((parseFloat(e.target.value) / 2).toFixed(1));
                      }
                    }}
                    placeholder="e.g. 52.0"
                    className="w-full px-2.5 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none"
                  />
                </div>
              </div>

              <div className="flex items-center justify-between pt-1">
                <label className="flex items-center space-x-2 text-xs text-slate-700 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={routeIsDefault}
                    onChange={(e) => setRouteIsDefault(e.target.checked)}
                    className="rounded text-brand-600"
                  />
                  <span>Make this my default route for new drives</span>
                </label>

                <div className="flex items-center space-x-2">
                  {editingRouteId && (
                    <button
                      type="button"
                      onClick={handleCancelEditRoute}
                      className="px-3.5 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-xs font-semibold transition-all"
                    >
                      Cancel
                    </button>
                  )}
                  <button
                    type="submit"
                    className={`px-4 py-2 text-white rounded-lg text-xs font-semibold shadow-sm transition-all flex items-center space-x-1.5 ${
                      editingRouteId
                        ? 'bg-amber-600 hover:bg-amber-700'
                        : 'bg-brand-600 hover:bg-brand-700'
                    }`}
                  >
                    {editingRouteId ? (
                      <>
                        <Check className="w-3.5 h-3.5" />
                        <span>Save Changes</span>
                      </>
                    ) : (
                      <>
                        <Plus className="w-3.5 h-3.5" />
                        <span>Create Route</span>
                      </>
                    )}
                  </button>
                </div>
              </div>
            </form>
          )}

          {/* Routes Table */}
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-slate-50 text-slate-500 uppercase tracking-wider text-[10px] border-b border-slate-200">
                <tr>
                  <th className="py-2.5 px-3">Route</th>
                  <th className="py-2.5 px-3">Origin (Home)</th>
                  <th className="py-2.5 px-3">Destination</th>
                  <th className="py-2.5 px-3">Round Trip</th>
                  <th className="py-2.5 px-3">Status</th>
                  {canWrite && <th className="py-2.5 px-3 text-right">Action</th>}
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100 text-slate-700">
                {routes.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="py-6 text-center text-slate-400">
                      No routes created yet. Add a destination above to create your first route!
                    </td>
                  </tr>
                ) : (
                  routes.map((r) => (
                    <tr
                      key={r.id}
                      className={`transition-colors ${
                        editingRouteId === r.id ? 'bg-amber-50/70 font-medium' : 'hover:bg-slate-50/80'
                      }`}
                    >
                      <td className="py-2.5 px-3 font-semibold text-slate-900">{r.name}</td>
                      <td className="py-2.5 px-3 text-slate-500">{r.origin_location_name}</td>
                      <td className="py-2.5 px-3 text-slate-500">{r.destination_location_name}</td>
                      <td className="py-2.5 px-3 font-bold text-slate-800">{r.distance_round_trip_km} km</td>
                      <td className="py-2.5 px-3">
                        {r.is_default && (
                          <span className="text-[10px] px-1.5 py-0.5 rounded bg-brand-100 text-brand-700 font-bold uppercase">
                            Default
                          </span>
                        )}
                      </td>
                      {canWrite && (
                        <td className="py-2.5 px-3 text-right">
                          <div className="flex items-center justify-end space-x-1">
                            <button
                              onClick={() => handleStartEditRoute(r)}
                              className={`p-1 rounded transition-colors ${
                                editingRouteId === r.id
                                  ? 'text-amber-700 bg-amber-100 font-bold'
                                  : 'text-slate-400 hover:text-amber-600 hover:bg-amber-50'
                              }`}
                              title="Edit route"
                            >
                              <Pencil className="w-3.5 h-3.5" />
                            </button>
                            <button
                              onClick={() => handleDeleteRoute(r.id)}
                              className="text-slate-400 hover:text-red-600 hover:bg-red-50 transition-colors p-1 rounded"
                              title="Delete route"
                            >
                              <Trash2 className="w-3.5 h-3.5" />
                            </button>
                          </div>
                        </td>
                      )}
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* SECTION: ACCOUNT SETTINGS */}
        <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-sm lg:col-span-2">
          <div className="flex items-center space-x-2.5 mb-2 text-slate-900">
            <div className="w-8 h-8 rounded-lg bg-blue-50 text-brand-600 flex items-center justify-center">
              <UserCog className="w-4 h-4" />
            </div>
            <h2 className="text-base font-bold">Account Settings</h2>
          </div>
          <p className="text-xs text-slate-500 mb-6">
            Manage your personal login credentials, change your email address, and update your password.
          </p>

          {isViewingSharedAccount ? (
            <div className="p-4 bg-amber-50 border border-amber-200 rounded-xl text-xs text-amber-800 flex items-center space-x-2.5">
              <AlertCircle className="w-4 h-4 text-amber-600 shrink-0" />
              <span>
                You are currently viewing a shared account (<strong>{activeAccountName}</strong>).
                Account credentials can only be edited when viewing your personal account.
              </span>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
              {/* Profile Details Form */}
              <div className="p-5 bg-slate-50 rounded-xl border border-slate-200/80">
                <div className="flex items-center space-x-2 mb-4">
                  <UserCog className="w-4 h-4 text-slate-600" />
                  <h3 className="text-xs font-bold uppercase tracking-wider text-slate-700">
                    Profile Information
                  </h3>
                </div>

                {profileSuccess && (
                  <div className="mb-4 p-3 bg-emerald-50 border border-emerald-200 text-emerald-800 rounded-lg text-xs flex items-center space-x-2">
                    <CheckCircle2 className="w-4 h-4 shrink-0 text-emerald-600" />
                    <span>{profileSuccess}</span>
                  </div>
                )}

                {profileError && (
                  <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-800 rounded-lg text-xs flex items-center space-x-2">
                    <AlertCircle className="w-4 h-4 shrink-0 text-red-600" />
                    <span>{profileError}</span>
                  </div>
                )}

                <form onSubmit={handleUpdateProfile} className="space-y-4">
                  <div>
                    <label className="block text-[11px] font-semibold text-slate-600 mb-1">
                      Username
                    </label>
                    <input
                      type="text"
                      required
                      value={profileUsername}
                      onChange={(e) => setProfileUsername(e.target.value)}
                      placeholder="Username (min 3 chars)"
                      className="w-full px-2.5 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none focus:ring-1 focus:ring-brand-500"
                    />
                  </div>

                  <div>
                    <label className="block text-[11px] font-semibold text-slate-600 mb-1">
                      Email Address
                    </label>
                    <input
                      type="email"
                      required
                      value={profileEmail}
                      onChange={(e) => setProfileEmail(e.target.value)}
                      placeholder="you@domain.com"
                      className="w-full px-2.5 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none focus:ring-1 focus:ring-brand-500"
                    />
                  </div>

                  <button
                    type="submit"
                    disabled={isUpdatingProfile}
                    className="w-full py-2 bg-brand-600 hover:bg-brand-700 disabled:opacity-50 text-white rounded-lg text-xs font-semibold shadow-sm transition-all"
                  >
                    {isUpdatingProfile ? 'Saving Profile...' : 'Save Profile Changes'}
                  </button>
                </form>
              </div>

              {/* Password Change Form */}
              <div className="p-5 bg-slate-50 rounded-xl border border-slate-200/80">
                <div className="flex items-center space-x-2 mb-4">
                  <KeyRound className="w-4 h-4 text-slate-600" />
                  <h3 className="text-xs font-bold uppercase tracking-wider text-slate-700">
                    Change Password
                  </h3>
                </div>

                {passwordSuccess && (
                  <div className="mb-4 p-3 bg-emerald-50 border border-emerald-200 text-emerald-800 rounded-lg text-xs flex items-center space-x-2">
                    <CheckCircle2 className="w-4 h-4 shrink-0 text-emerald-600" />
                    <span>{passwordSuccess}</span>
                  </div>
                )}

                {passwordError && (
                  <div className="mb-4 p-3 bg-red-50 border border-red-200 text-red-800 rounded-lg text-xs flex items-center space-x-2">
                    <AlertCircle className="w-4 h-4 shrink-0 text-red-600" />
                    <span>{passwordError}</span>
                  </div>
                )}

                <form onSubmit={handleChangePassword} className="space-y-4">
                  <div>
                    <label className="block text-[11px] font-semibold text-slate-600 mb-1">
                      Current Password
                    </label>
                    <div className="relative">
                      <Lock className="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-2" />
                      <input
                        type="password"
                        required
                        value={currentPassword}
                        onChange={(e) => setCurrentPassword(e.target.value)}
                        placeholder="••••••••"
                        className="w-full pl-8 pr-2.5 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none focus:ring-1 focus:ring-brand-500"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-[11px] font-semibold text-slate-600 mb-1">
                      New Password
                    </label>
                    <div className="relative">
                      <Lock className="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-2" />
                      <input
                        type="password"
                        required
                        value={newPassword}
                        onChange={(e) => setNewPassword(e.target.value)}
                        placeholder="At least 6 characters"
                        className="w-full pl-8 pr-2.5 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none focus:ring-1 focus:ring-brand-500"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-[11px] font-semibold text-slate-600 mb-1">
                      Confirm New Password
                    </label>
                    <div className="relative">
                      <Lock className="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-2" />
                      <input
                        type="password"
                        required
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                        placeholder="Repeat new password"
                        className="w-full pl-8 pr-2.5 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none focus:ring-1 focus:ring-brand-500"
                      />
                    </div>
                  </div>

                  <button
                    type="submit"
                    disabled={isChangingPassword}
                    className="w-full py-2 bg-slate-900 hover:bg-slate-800 disabled:opacity-50 text-white rounded-lg text-xs font-semibold shadow-sm transition-all"
                  >
                    {isChangingPassword ? 'Updating Password...' : 'Update Password'}
                  </button>
                </form>
              </div>
            </div>
          )}
        </div>

        {/* SECTION 4: ACCOUNT ACCESS SHARING */}
        <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-sm lg:col-span-2">
          <div className="flex items-center space-x-2.5 mb-4 text-slate-900">
            <div className="w-8 h-8 rounded-lg bg-indigo-50 text-indigo-600 flex items-center justify-center">
              <Users className="w-4 h-4" />
            </div>
            <h2 className="text-base font-bold">Account Access Sharing</h2>
          </div>
          <p className="text-xs text-slate-500 mb-6">
            Allow others to have <strong>read</strong> or <strong>write</strong> access to your KDB account.
            Write access allows them to log drives on your behalf (e.g. spouse, accountant, or peer).
          </p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Grant Access Form */}
            <div>
              <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-3">
                Grant Access to a User
              </h3>
              <form onSubmit={handleCreateShare} className="space-y-3 p-4 bg-slate-50 rounded-xl">
                <div>
                  <label className="block text-[11px] font-semibold text-slate-700 mb-1">
                    Username or Email
                  </label>
                  <input
                    type="text"
                    required
                    value={shareUserOrEmail}
                    onChange={(e) => setShareUserOrEmail(e.target.value)}
                    placeholder="Enter user's username or email"
                    className="w-full px-2.5 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none"
                  />
                </div>

                <div>
                  <label className="block text-[11px] font-semibold text-slate-700 mb-1">
                    Permission Level
                  </label>
                  <select
                    value={sharePermission}
                    onChange={(e) => setSharePermission(e.target.value as SharePermission)}
                    className="w-full px-2.5 py-1.5 text-xs bg-white border border-slate-200 rounded-lg focus:outline-none"
                  >
                    <option value="read">Read Only (View drives & overview statistics)</option>
                    <option value="write">Read & Write (Log drives & manage routes)</option>
                  </select>
                </div>

                <button
                  type="submit"
                  className="w-full py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-xs font-semibold shadow-sm transition-all"
                >
                  Grant Access
                </button>
              </form>
            </div>

            {/* Existing Shared Access List */}
            <div>
              <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-3">
                Users Who Have Access to Your Account
              </h3>
              <div className="space-y-2">
                {shares.my_shares.length === 0 ? (
                  <div className="p-4 border border-dashed border-slate-200 rounded-xl text-center text-xs text-slate-400">
                    You have not shared your account with anyone yet.
                  </div>
                ) : (
                  shares.my_shares.map((s) => (
                    <div
                      key={s.id}
                      className="p-3 border border-slate-200 rounded-xl flex items-center justify-between"
                    >
                      <div>
                        <div className="text-xs font-bold text-slate-900">{s.shared_with_username}</div>
                        <div className="text-[10px] text-slate-500">
                          Role: <strong className="uppercase">{s.permission}</strong>
                        </div>
                      </div>
                      <button
                        onClick={() => handleRevokeShare(s.id)}
                        className="text-xs text-red-600 hover:text-red-700 font-medium px-2 py-1 rounded hover:bg-red-50"
                      >
                        Revoke
                      </button>
                    </div>
                  ))
                )}
              </div>

              {/* Accounts Shared With You */}
              {shares.shared_with_me.length > 0 && (
                <div className="mt-6">
                  <h3 className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-2">
                    Accounts Shared With You
                  </h3>
                  <div className="space-y-2">
                    {shares.shared_with_me.map((s) => (
                      <div
                        key={s.id}
                        className="p-3 bg-brand-50/50 border border-brand-200 rounded-xl flex items-center justify-between text-xs"
                      >
                        <div>
                          <div className="font-bold text-slate-900">{s.owner_username}'s Account</div>
                          <div className="text-[10px] text-slate-500">
                            You have <strong className="uppercase">{s.permission}</strong> permission
                          </div>
                        </div>
                        <span className="text-[10px] px-2 py-0.5 rounded-full bg-brand-100 text-brand-700 font-semibold">
                          Active
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* SECTION 5: DANISH SKAT TAX DEDUCTION RATES */}
        <div className="bg-white p-6 rounded-2xl border border-slate-200/80 shadow-sm lg:col-span-2">
          <div className="flex items-center space-x-2.5 mb-2 text-slate-900">
            <div className="w-8 h-8 rounded-lg bg-amber-50 text-amber-600 flex items-center justify-center">
              <Shield className="w-4 h-4" />
            </div>
            <h2 className="text-base font-bold">Danish Befordringsfradrag Rules (SKAT)</h2>
          </div>
          <p className="text-xs text-slate-500 mb-4">
            Official statutory rates applied automatically to your registered commutes:
          </p>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
            <div className="p-3.5 bg-slate-50 border border-slate-200 rounded-xl">
              <div className="font-semibold text-slate-700">0 – 24 km/day</div>
              <div className="text-xl font-bold text-slate-900 mt-1">0.00 DKK</div>
              <div className="text-[10px] text-slate-500 mt-0.5">Threshold (Bundgrænse)</div>
            </div>

            <div className="p-3.5 bg-slate-50 border border-slate-200 rounded-xl">
              <div className="font-semibold text-slate-700">25 – 120 km/day</div>
              <div className="text-xl font-bold text-brand-700 mt-1">2.23 DKK/km</div>
              <div className="text-[10px] text-slate-500 mt-0.5">Standard Rate</div>
            </div>

            <div className="p-3.5 bg-slate-50 border border-slate-200 rounded-xl">
              <div className="font-semibold text-slate-700">Above 120 km/day</div>
              <div className="text-xl font-bold text-indigo-700 mt-1">1.12 DKK/km</div>
              <div className="text-[10px] text-slate-500 mt-0.5">High Distance Rate</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
