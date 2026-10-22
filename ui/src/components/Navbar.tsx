import React, { useState, useEffect, useRef } from 'react';
import { useAuth } from '../context/AuthContext';
import {
  Car,
  BarChart3,
  Settings as SettingsIcon,
  LogOut,
  ArrowRightLeft,
  Check,
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
    activePermission,
    isViewingSharedAccount,
    switchAccount,
  } = useAuth();

  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [accountMenuOpen, setAccountMenuOpen] = useState(false);
  const accountMenuRef = useRef<HTMLDivElement>(null);

  // Close account menu on click outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent | TouchEvent) => {
      if (accountMenuRef.current && !accountMenuRef.current.contains(event.target as Node)) {
        setAccountMenuOpen(false);
      }
    };
    if (accountMenuOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      document.addEventListener('touchstart', handleClickOutside);
    }
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('touchstart', handleClickOutside);
    };
  }, [accountMenuOpen]);

  // Close mobile drawer and account menu on Escape key
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setMobileMenuOpen(false);
        setAccountMenuOpen(false);
      }
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
    setAccountMenuOpen(false);
  };

  const activeSharedAccount = isViewingSharedAccount
    ? sharedAccounts.find((a) => a.owner_id === activeAccountId)
    : null;
  const displayedUsername = activeSharedAccount?.username || user?.username || 'User';

  return (
    <>
      <header className="bg-white border-b border-slate-200 sticky top-0 z-40 shadow-sm">
        <div className="max-w-7xl mx-auto px-3 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            {/* Logo & Brand on the Left */}
            <div
              className="flex items-center space-x-2.5 sm:space-x-3 cursor-pointer select-none shrink-0"
              onClick={() => handleSelectTab('register')}
            >
              <div className="w-9 h-9 sm:w-10 sm:h-10 rounded-xl bg-brand-600 flex items-center justify-center text-white shadow-md shadow-brand-500/20">
                <Car className="w-5 h-5 sm:w-6 sm:h-6" />
              </div>
              <div>
                <span className="text-lg sm:text-xl font-bold tracking-tight text-slate-900">KDB</span>
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

            {/* Top Menu Right Controls: User Card with Account Switcher & Drawer / Logout */}
            <div className="flex items-center space-x-2 sm:space-x-3">
              {/* User Card with Integrated ArrowRightLeft Account Switcher */}
              <div className="relative" ref={accountMenuRef}>
                <button
                  type="button"
                  onClick={() => setAccountMenuOpen((prev) => !prev)}
                  className={`flex items-center space-x-2 sm:space-x-2.5 p-1 sm:p-1.5 pl-2 sm:pl-2.5 pr-1.5 sm:pr-2 rounded-xl border transition-all text-left focus:outline-none focus:ring-2 focus:ring-brand-500/30 ${
                    isViewingSharedAccount
                      ? 'bg-amber-50/90 border-amber-300 hover:bg-amber-100/90 text-amber-900 shadow-xs'
                      : 'bg-slate-50 hover:bg-slate-100 border-slate-200 text-slate-800'
                  }`}
                  title={`Active Account: ${displayedUsername}. Click to switch view.`}
                  aria-expanded={accountMenuOpen}
                  aria-haspopup="true"
                >
                  {/* User Initial Avatar */}
                  <div
                    className={`w-7 h-7 sm:w-8 sm:h-8 rounded-lg flex items-center justify-center font-bold text-xs shadow-xs shrink-0 ${
                      isViewingSharedAccount ? 'bg-amber-500 text-white' : 'bg-brand-600 text-white'
                    }`}
                  >
                    {displayedUsername.charAt(0).toUpperCase()}
                  </div>

                  {/* Username & Subtitle */}
                  <div className="flex flex-col min-w-0 pr-0.5 sm:pr-1">
                    <div className="flex items-center space-x-1 sm:space-x-1.5">
                      <span className="text-xs font-bold truncate max-w-[80px] sm:max-w-[120px]">
                        {displayedUsername}
                      </span>
                      {isViewingSharedAccount ? (
                        <span className="text-[9px] sm:text-[10px] px-1 sm:px-1.5 py-0.2 rounded font-semibold bg-amber-200/80 text-amber-900 border border-amber-300/60 shrink-0">
                          {activePermission === 'write' ? 'Write' : 'Read'}
                        </span>
                      ) : (
                        <span className="hidden sm:inline-block text-[9px] px-1.5 py-0.2 rounded font-medium bg-slate-200/70 text-slate-600 shrink-0">
                          Personal
                        </span>
                      )}
                    </div>
                    <span className="hidden sm:block text-[10px] text-slate-500 truncate max-w-[120px]">
                      {isViewingSharedAccount ? 'Collaborator view' : (user?.email || 'Personal')}
                    </span>
                  </div>

                  {/* ArrowRightLeft Switcher Icon Button */}
                  <div
                    className={`p-1 sm:p-1.5 rounded-lg transition-colors shrink-0 ${
                      isViewingSharedAccount
                        ? 'bg-amber-200/80 text-amber-900 hover:bg-amber-300'
                        : 'bg-white text-slate-600 border border-slate-200/90 hover:text-slate-900 shadow-xs'
                    }`}
                    title="Switch Account"
                  >
                    <ArrowRightLeft className="w-3.5 h-3.5" />
                  </div>
                </button>

                {/* Popover Dropdown Menu */}
                {accountMenuOpen && (
                  <div
                    className="absolute right-0 top-full mt-2 w-80 max-w-[calc(100vw-1.5rem)] bg-white rounded-2xl shadow-xl border border-slate-200 py-2.5 z-50 animate-in fade-in zoom-in-95 duration-100"
                    role="menu"
                    aria-label="Account switcher options"
                  >
                    {/* Signed-in user header */}
                    <div className="px-4 py-2 border-b border-slate-100">
                      <div className="text-[10px] font-bold text-slate-400 uppercase tracking-wider">
                        Signed in as
                      </div>
                      <div className="text-xs font-bold text-slate-900 truncate">
                        {user?.username}{' '}
                        <span className="text-[11px] font-normal text-slate-500">({user?.email})</span>
                      </div>
                      {isViewingSharedAccount && (
                        <div className="mt-1.5 text-[10px] font-medium text-amber-800 bg-amber-50 px-2 py-1 rounded-md border border-amber-200/60">
                          Viewing delegated account of <strong>{displayedUsername}</strong>
                        </div>
                      )}
                    </div>

                    {/* Section: Switch View */}
                    <div className="px-3 pt-2.5 pb-1">
                      <div className="flex items-center justify-between px-2 pb-1.5">
                        <div className="flex items-center space-x-1.5 text-[11px] font-bold text-slate-500 uppercase tracking-wider">
                          <ArrowRightLeft className="w-3.5 h-3.5 text-brand-600" />
                          <span>Switch Account</span>
                        </div>
                        <span className="text-[10px] text-slate-400 font-medium">
                          {sharedAccounts.length + 1} account{sharedAccounts.length > 0 ? 's' : ''}
                        </span>
                      </div>

                      <div className="space-y-1">
                        {/* Personal Account Option */}
                        <button
                          onClick={() => {
                            switchAccount(null);
                            setAccountMenuOpen(false);
                          }}
                          className={`w-full text-left px-3 py-2.5 rounded-xl text-xs flex items-center justify-between transition-colors ${
                            !isViewingSharedAccount
                              ? 'bg-brand-50 text-brand-800 font-semibold border border-brand-200/80 shadow-xs'
                              : 'text-slate-700 hover:bg-slate-50 border border-transparent'
                          }`}
                        >
                          <div className="flex items-center space-x-2.5 min-w-0">
                            <div
                              className={`w-7 h-7 rounded-lg flex items-center justify-center text-xs font-bold shrink-0 ${
                                !isViewingSharedAccount
                                  ? 'bg-brand-600 text-white'
                                  : 'bg-slate-100 text-slate-600'
                              }`}
                            >
                              {user?.username?.charAt(0).toUpperCase() || 'U'}
                            </div>
                            <div className="min-w-0">
                              <div className="truncate font-semibold">{user?.username} (Personal)</div>
                              <div className="text-[10px] text-slate-400">
                                Your personal drives & routes
                              </div>
                            </div>
                          </div>
                          {!isViewingSharedAccount && (
                            <div className="w-5 h-5 rounded-full bg-brand-600 text-white flex items-center justify-center shrink-0 ml-2">
                              <Check className="w-3 h-3 stroke-[3]" />
                            </div>
                          )}
                        </button>

                        {/* Shared Accounts List */}
                        {sharedAccounts.length > 0 && (
                          <div className="pt-1.5 space-y-1">
                            <div className="px-2 pt-1 pb-0.5 text-[10px] font-semibold text-slate-400 uppercase tracking-wider">
                              Shared Collaborators
                            </div>
                            {sharedAccounts.map((acc) => {
                              const isSelected = activeAccountId === acc.owner_id;
                              return (
                                <button
                                  key={acc.owner_id}
                                  onClick={() => {
                                    switchAccount(acc.owner_id);
                                    setAccountMenuOpen(false);
                                  }}
                                  className={`w-full text-left px-3 py-2 rounded-xl text-xs flex items-center justify-between transition-colors ${
                                    isSelected
                                      ? 'bg-amber-50 text-amber-900 font-semibold border border-amber-200/80 shadow-xs'
                                      : 'text-slate-700 hover:bg-slate-50 border border-transparent'
                                  }`}
                                >
                                  <div className="flex items-center space-x-2.5 min-w-0">
                                    <div
                                      className={`w-7 h-7 rounded-lg flex items-center justify-center text-xs font-bold shrink-0 ${
                                        isSelected
                                          ? 'bg-amber-500 text-white'
                                          : 'bg-slate-100 text-slate-600'
                                      }`}
                                    >
                                      {acc.username.charAt(0).toUpperCase()}
                                    </div>
                                    <div className="min-w-0">
                                      <div className="truncate font-semibold flex items-center space-x-1.5">
                                        <span className="truncate">{acc.username}</span>
                                        <span
                                          className={`text-[9px] px-1.5 py-0.2 rounded-full font-medium ${
                                            acc.permission === 'write'
                                              ? 'bg-emerald-100 text-emerald-800 border border-emerald-200'
                                              : 'bg-slate-100 text-slate-600 border border-slate-200'
                                          }`}
                                        >
                                          {acc.permission === 'write' ? 'Write' : 'Read-only'}
                                        </span>
                                      </div>
                                      <div className="text-[10px] text-slate-400 truncate">
                                        Shared by collaborator
                                      </div>
                                    </div>
                                  </div>
                                  {isSelected && (
                                    <div className="w-5 h-5 rounded-full bg-amber-500 text-white flex items-center justify-center shrink-0 ml-2">
                                      <Check className="w-3 h-3 stroke-[3]" />
                                    </div>
                                  )}
                                </button>
                              );
                            })}
                          </div>
                        )}

                        {sharedAccounts.length === 0 && (
                          <div className="px-3 py-3 rounded-xl bg-slate-50 border border-dashed border-slate-200 text-center">
                            <p className="text-[11px] text-slate-500">
                              No collaborator accounts shared with you yet.
                            </p>
                          </div>
                        )}
                      </div>
                    </div>

                    {/* Dropdown Actions */}
                    <div className="mt-2 pt-2 border-t border-slate-100 px-3 space-y-1">
                      <button
                        onClick={() => {
                          handleSelectTab('settings');
                          setAccountMenuOpen(false);
                        }}
                        className="w-full text-left px-3 py-2 rounded-lg text-xs font-medium text-slate-600 hover:text-slate-900 hover:bg-slate-100 flex items-center justify-between transition-colors"
                      >
                        <div className="flex items-center space-x-2">
                          <SettingsIcon className="w-3.5 h-3.5 text-slate-500" />
                          <span>Manage Collaborators</span>
                        </div>
                        <span className="text-[10px] text-brand-600 font-semibold">Settings &rarr;</span>
                      </button>

                      <button
                        onClick={() => {
                          setAccountMenuOpen(false);
                          logout();
                        }}
                        className="w-full text-left px-3 py-2 rounded-lg text-xs font-medium text-red-600 hover:bg-red-50 flex items-center space-x-2 transition-colors"
                      >
                        <LogOut className="w-3.5 h-3.5" />
                        <span>Log Out</span>
                      </button>
                    </div>
                  </div>
                )}
              </div>

              {/* Desktop Direct Logout Button */}
              <button
                onClick={() => logout()}
                title="Log out"
                className="hidden md:flex p-2 text-slate-500 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
              >
                <LogOut className="w-4 h-4" />
              </button>

              {/* Mobile Burger Menu Icon Button */}
              <div className="flex md:hidden items-center">
                <button
                  onClick={() => {
                    setAccountMenuOpen(false);
                    setMobileMenuOpen(true);
                  }}
                  className="p-2 rounded-xl text-slate-700 hover:text-slate-900 hover:bg-slate-100 transition-colors focus:outline-none focus:ring-2 focus:ring-brand-500"
                  aria-label="Open navigation menu"
                >
                  <Menu className="w-6 h-6" />
                </button>
              </div>
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

          {/* User Profile Info Card in Drawer */}
          <div className="p-4 mx-4 mt-4 bg-slate-50 border border-slate-200/70 rounded-xl flex items-center space-x-3">
            <div
              className={`w-10 h-10 rounded-full font-bold flex items-center justify-center text-sm shrink-0 ${
                isViewingSharedAccount
                  ? 'bg-amber-100 text-amber-800'
                  : 'bg-brand-100 text-brand-700'
              }`}
            >
              {displayedUsername.charAt(0).toUpperCase() || <UserIcon className="w-5 h-5" />}
            </div>
            <div className="min-w-0 flex-1">
              <div className="text-xs font-bold text-slate-900 truncate">{displayedUsername}</div>
              <div className="text-[11px] text-slate-500 truncate">
                {isViewingSharedAccount
                  ? `Delegated (${activePermission === 'write' ? 'Write' : 'Read-only'})`
                  : (user?.email || 'Personal')}
              </div>
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

          {/* Account Switcher Section in Drawer */}
          <div className="px-4 pt-1">
            <div className="px-3 py-1 flex items-center justify-between text-[11px] font-bold text-slate-400 uppercase tracking-wider">
              <div className="flex items-center space-x-1.5">
                <ArrowRightLeft className="w-3.5 h-3.5 text-brand-600" />
                <span>Switch Account</span>
              </div>
            </div>
            <div className="mt-1 space-y-1">
              <button
                onClick={() => {
                  switchAccount(null);
                  setMobileMenuOpen(false);
                }}
                className={`w-full text-left px-3.5 py-2.5 rounded-xl text-xs flex items-center justify-between ${
                  !isViewingSharedAccount
                    ? 'bg-brand-50/80 text-brand-800 font-bold border border-brand-200'
                    : 'text-slate-700 hover:bg-slate-50'
                }`}
              >
                <span className="truncate">{user?.username} (Personal)</span>
                {!isViewingSharedAccount && <Check className="w-4 h-4 text-brand-600 stroke-[3]" />}
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
                      ? 'bg-amber-50/80 text-amber-900 font-bold border border-amber-200'
                      : 'text-slate-700 hover:bg-slate-50'
                  }`}
                >
                  <span className="truncate">{acc.username}</span>
                  <div className="flex items-center space-x-2">
                    <span
                      className={`text-[10px] px-1.5 py-0.5 rounded ${
                        acc.permission === 'write'
                          ? 'bg-emerald-100 text-emerald-800'
                          : 'bg-slate-100 text-slate-600'
                      }`}
                    >
                      {acc.permission === 'write' ? 'Write' : 'Read'}
                    </span>
                    {activeAccountId === acc.owner_id && (
                      <Check className="w-4 h-4 text-amber-600 stroke-[3]" />
                    )}
                  </div>
                </button>
              ))}
            </div>
          </div>
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
