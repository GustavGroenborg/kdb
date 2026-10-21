import React, { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';
import {
  Car,
  BarChart3,
  Settings as SettingsIcon,
  LogOut,
  Users,
  ShieldCheck,
  ChevronDown,
  Menu,
  X,
  User as UserIcon,
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

  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  // Close mobile drawer on Escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setMobileMenuOpen(false);
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  // Prevent background scrolling when drawer is open
  useEffect(() => {
    if (mobileMenuOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [mobileMenuOpen]);

  const handleSelectTab = (tab: ActiveTab) => {
    onTabChange(tab);
    setMobileMenuOpen(false);
  };

  return (
    <>
      <header className="bg-white border-b border-slate-200 sticky top-0 z-40 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            {/* Logo & Brand on the Left */}
            <div
              className="flex items-center space-x-3 cursor-pointer select-none"
              onClick={() => handleSelectTab('register')}
            >
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

            {/* Desktop Navigation Tabs (Hidden on Mobile) */}
            <nav className="hidden md:flex space-x-1 sm:space-x-2">
              <button
                onClick={() => handleSelectTab('register')}
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
                onClick={() => handleSelectTab('overview')}
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
                onClick={() => handleSelectTab('settings')}
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

            {/* Desktop User & Account Controls (Hidden on Mobile) */}
            <div className="hidden md:flex items-center space-x-3">
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
              <div className="flex flex-col text-right">
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

            {/* Mobile Burger Icon Button (Visible only on Mobile) */}
            <div className="flex md:hidden items-center">
              <button
                onClick={() => setMobileMenuOpen(true)}
                className="p-2.5 rounded-xl text-slate-700 hover:text-slate-900 hover:bg-slate-100 transition-colors focus:outline-none focus:ring-2 focus:ring-brand-500"
                aria-label="Open navigation menu"
              >
                <Menu className="w-6 h-6" />
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Mobile Drawer Backdrop */}
      {mobileMenuOpen && (
        <div
          onClick={() => setMobileMenuOpen(false)}
          className="fixed inset-0 bg-slate-900/50 backdrop-blur-xs z-50 md:hidden transition-opacity duration-300"
          aria-hidden="true"
        />
      )}

      {/* Mobile Slide-in Drawer from the Right */}
      <div
        className={`fixed top-0 right-0 bottom-0 w-80 max-w-[85vw] bg-white z-50 shadow-2xl flex flex-col justify-between transform transition-transform duration-300 ease-in-out md:hidden ${
          mobileMenuOpen ? 'translate-x-0' : 'translate-x-full'
        }`}
        role="dialog"
        aria-modal="true"
      >
        {/* Drawer Header */}
        <div>
          <div className="p-5 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
            <div className="flex items-center space-x-2.5">
              <div className="w-8 h-8 rounded-lg bg-brand-600 flex items-center justify-center text-white shadow-sm">
                <Car className="w-5 h-5" />
              </div>
              <div>
                <span className="font-bold text-base text-slate-900">KDB Menu</span>
              </div>
            </div>
            <button
              onClick={() => setMobileMenuOpen(false)}
              className="p-2 text-slate-400 hover:text-slate-700 hover:bg-slate-100 rounded-lg transition-colors"
              aria-label="Close menu"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* User Profile Info Card */}
          <div className="p-4 mx-4 mt-4 bg-slate-50 border border-slate-200/70 rounded-xl flex items-center space-x-3">
            <div className="w-10 h-10 rounded-full bg-brand-100 text-brand-700 font-bold flex items-center justify-center text-sm shrink-0">
              {user?.username ? user.username.charAt(0).toUpperCase() : <UserIcon className="w-5 h-5" />}
            </div>
            <div className="min-w-0 flex-1">
              <div className="text-xs font-bold text-slate-900 truncate">{user?.username}</div>
              <div className="text-[11px] text-slate-500 truncate">{user?.email}</div>
            </div>
          </div>

          {/* Navigation Tabs List */}
          <div className="p-4 space-y-1.5">
            <div className="px-3 py-1 text-[11px] font-bold text-slate-400 uppercase tracking-wider">
              Navigation
            </div>
            <button
              onClick={() => handleSelectTab('register')}
              className={`w-full flex items-center space-x-3 px-3.5 py-3 rounded-xl text-sm font-medium transition-all ${
                activeTab === 'register'
                  ? 'bg-brand-50 text-brand-700 font-semibold shadow-sm'
                  : 'text-slate-700 hover:bg-slate-50 hover:text-slate-900'
              }`}
            >
              <Car className="w-5 h-5" />
              <span>Register Drives</span>
            </button>

            <button
              onClick={() => handleSelectTab('overview')}
              className={`w-full flex items-center space-x-3 px-3.5 py-3 rounded-xl text-sm font-medium transition-all ${
                activeTab === 'overview'
                  ? 'bg-brand-50 text-brand-700 font-semibold shadow-sm'
                  : 'text-slate-700 hover:bg-slate-50 hover:text-slate-900'
              }`}
            >
              <BarChart3 className="w-5 h-5" />
              <span>Overview & Statistics</span>
            </button>

            <button
              onClick={() => handleSelectTab('settings')}
              className={`w-full flex items-center space-x-3 px-3.5 py-3 rounded-xl text-sm font-medium transition-all ${
                activeTab === 'settings'
                  ? 'bg-brand-50 text-brand-700 font-semibold shadow-sm'
                  : 'text-slate-700 hover:bg-slate-50 hover:text-slate-900'
              }`}
            >
              <SettingsIcon className="w-5 h-5" />
              <span>Settings</span>
            </button>
          </div>

          {/* Account Switcher (if shared accounts exist) */}
          {sharedAccounts.length > 0 && (
            <div className="px-4 pt-2">
              <div className="px-3 py-1 text-[11px] font-bold text-slate-400 uppercase tracking-wider">
                Accounts
              </div>
              <div className="mt-1 space-y-1">
                <button
                  onClick={() => {
                    switchAccount(null);
                    setMobileMenuOpen(false);
                  }}
                  className={`w-full text-left px-3.5 py-2.5 rounded-xl text-xs flex items-center justify-between ${
                    !isViewingSharedAccount
                      ? 'bg-brand-50/70 text-brand-700 font-bold border border-brand-200'
                      : 'text-slate-700 hover:bg-slate-50'
                  }`}
                >
                  <span className="truncate">{user?.username} (Personal)</span>
                  {!isViewingSharedAccount && <ShieldCheck className="w-4 h-4 text-brand-600" />}
                </button>
                {sharedAccounts.map((acc) => (
                  <button
                    key={acc.owner_id}
                    onClick={() => {
                      switchAccount(acc.owner_id);
                      setMobileMenuOpen(false);
                    }}
                    className={`w-full text-left px-3.5 py-2.5 rounded-xl text-xs flex items-center justify-between ${
                      activeAccountId === acc.owner_id
                        ? 'bg-brand-50/70 text-brand-700 font-bold border border-brand-200'
                        : 'text-slate-700 hover:bg-slate-50'
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
        </div>

        {/* Drawer Footer: Logout Button */}
        <div className="p-4 border-t border-slate-100 bg-slate-50/50">
          <button
            onClick={() => {
              setMobileMenuOpen(false);
              logout();
            }}
            className="w-full flex items-center justify-center space-x-2 px-4 py-2.5 rounded-xl text-xs font-semibold text-red-600 bg-red-50 hover:bg-red-100 transition-colors"
          >
            <LogOut className="w-4 h-4" />
            <span>Log Out</span>
          </button>
        </div>
      </div>
    </>
  );
};
