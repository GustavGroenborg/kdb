import React from 'react';
import { useAuth } from '../context/AuthContext';
import {
  Car,
  BarChart3,
  Settings as SettingsIcon,
  LogOut,
  Users,
  ShieldCheck,
  ChevronDown,
} from 'lucide-react';

export type ActiveTab = 'register' | 'overview' | 'settings';

interface NavbarProps {
  activeTab: ActiveTab;
  onTabChange: (tab: ActiveTab) => void;
}

export const Navbar: React.FC<NavbarProps> = ({ activeTab, onTabChange }) => {
  const {
    user,
    logout,
    sharedAccounts,
    activeAccountId,
    activeAccountName,
    isViewingSharedAccount,
    switchAccount,
  } = useAuth();

  return (
    <header className="bg-white border-b border-slate-200 sticky top-0 z-40 shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo & Brand */}
          <div className="flex items-center space-x-3 cursor-pointer" onClick={() => onTabChange('register')}>
            <div className="w-10 h-10 rounded-xl bg-brand-600 flex items-center justify-center text-white shadow-md shadow-brand-500/20">
              <Car className="w-6 h-6" />
            </div>
            <div>
              <span className="text-xl font-bold tracking-tight text-slate-900">KDB</span>
              <span className="hidden sm:inline-block ml-2 text-xs font-semibold px-2 py-0.5 rounded-full bg-brand-50 text-brand-700 border border-brand-200">
                Befordringsfradrag
              </span>
            </div>
          </div>

          {/* Navigation Tabs */}
          <nav className="flex space-x-1 sm:space-x-2">
            <button
              onClick={() => onTabChange('register')}
              className={`flex items-center space-x-2 px-3 sm:px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                activeTab === 'register'
                  ? 'bg-brand-50 text-brand-700 shadow-sm font-semibold'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
              }`}
            >
              <Car className="w-4 h-4" />
              <span>Register</span>
            </button>

            <button
              onClick={() => onTabChange('overview')}
              className={`flex items-center space-x-2 px-3 sm:px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                activeTab === 'overview'
                  ? 'bg-brand-50 text-brand-700 shadow-sm font-semibold'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
              }`}
            >
              <BarChart3 className="w-4 h-4" />
              <span>Overview</span>
            </button>

            <button
              onClick={() => onTabChange('settings')}
              className={`flex items-center space-x-2 px-3 sm:px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                activeTab === 'settings'
                  ? 'bg-brand-50 text-brand-700 shadow-sm font-semibold'
                  : 'text-slate-600 hover:text-slate-900 hover:bg-slate-100'
              }`}
            >
              <SettingsIcon className="w-4 h-4" />
              <span>Settings</span>
            </button>
          </nav>

          {/* User & Account Switcher */}
          <div className="flex items-center space-x-3">
            {/* Account Switcher if shared accounts exist */}
            {sharedAccounts.length > 0 && (
              <div className="relative group">
                <button className="flex items-center space-x-2 text-xs font-medium bg-slate-100 hover:bg-slate-200 text-slate-700 px-3 py-1.5 rounded-lg border border-slate-200">
                  <Users className="w-3.5 h-3.5 text-slate-500" />
                  <span className="max-w-[120px] truncate">{activeAccountName}</span>
                  <ChevronDown className="w-3 h-3 text-slate-400" />
                </button>
                <div className="hidden group-hover:block absolute right-0 mt-1 w-60 bg-white rounded-xl shadow-lg border border-slate-200 py-1.5 z-50">
                  <div className="px-3 py-1 text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                    Switch Account
                  </div>
                  <button
                    onClick={() => switchAccount(null)}
                    className={`w-full text-left px-3 py-2 text-xs flex items-center justify-between hover:bg-slate-50 ${
                      !isViewingSharedAccount ? 'font-bold text-brand-600 bg-brand-50/50' : 'text-slate-700'
                    }`}
                  >
                    <span>{user?.username} (Personal)</span>
                    {!isViewingSharedAccount && <ShieldCheck className="w-3.5 h-3.5" />}
                  </button>
                  {sharedAccounts.map((acc) => (
                    <button
                      key={acc.owner_id}
                      onClick={() => switchAccount(acc.owner_id)}
                      className={`w-full text-left px-3 py-2 text-xs flex items-center justify-between hover:bg-slate-50 ${
                        activeAccountId === acc.owner_id
                          ? 'font-bold text-brand-600 bg-brand-50/50'
                          : 'text-slate-700'
                      }`}
                    >
                      <span className="truncate">{acc.username}</span>
                      <span className="text-[10px] px-1.5 py-0.5 rounded bg-slate-100 text-slate-600">
                        {acc.permission === 'write' ? 'Write' : 'Read'}
                      </span>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* User Greeting & Logout */}
            <div className="hidden md:flex flex-col text-right">
              <span className="text-xs font-semibold text-slate-800">{user?.username}</span>
              <span className="text-[10px] text-slate-500 truncate max-w-[140px]">{user?.email}</span>
            </div>

            <button
              onClick={() => logout()}
              title="Log out"
              className="p-2 text-slate-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
            >
              <LogOut className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </header>
  );
};
