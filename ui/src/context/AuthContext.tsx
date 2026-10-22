import React, { createContext, useContext, useState, useEffect } from 'react';
import { User, SharedAccountInfo, SharePermission } from '../types';
import { api, setAuthToken, setTargetAccountId } from '../services/api';

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  hasRoutes: boolean;
  hasHome: boolean;
  sharedAccounts: SharedAccountInfo[];
  activeAccountId: string | null;
  activeAccountName: string;
  activePermission: SharePermission;
  isViewingSharedAccount: boolean;
  login: (payload: { username_or_email: string; password: string }) => Promise<void>;
  register: (payload: { username: string; email: string; password: string }) => Promise<void>;
  logout: () => Promise<void>;
  switchAccount: (accountId: string | null) => void;
  refreshMe: () => Promise<void>;
  updateUserSession: (updatedUser: User, token: string) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [hasRoutes, setHasRoutes] = useState<boolean>(false);
  const [hasHome, setHasHome] = useState<boolean>(false);
  const [sharedAccounts, setSharedAccounts] = useState<SharedAccountInfo[]>([]);
  const [activeAccountId, setActiveAccountIdState] = useState<string | null>(() => {
    return localStorage.getItem('kdb_target_account_id');
  });

  const refreshMe = async () => {
    try {
      const data = await api.me();
      setUser(data.user);
      setHasRoutes(data.has_routes);
      setHasHome(data.has_home);
      const shares = data.shared_accounts || [];
      setSharedAccounts(shares);

      // Verify currently active account still exists in shares
      const currentTarget = localStorage.getItem('kdb_target_account_id');
      if (currentTarget && !shares.some((s) => s.owner_id === currentTarget)) {
        setActiveAccountIdState(null);
        setTargetAccountId(null);
      }
    } catch {
      setUser(null);
      setAuthToken(null);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    refreshMe();
  }, []);

  const login = async (payload: { username_or_email: string; password: string }) => {
    const res = await api.login(payload);
    setAuthToken(res.token);
    setUser(res.user);
    setActiveAccountIdState(null);
    setTargetAccountId(null);
    await refreshMe();
  };

  const register = async (payload: { username: string; email: string; password: string }) => {
    const res = await api.register(payload);
    setAuthToken(res.token);
    setUser(res.user);
    setActiveAccountIdState(null);
    setTargetAccountId(null);
    await refreshMe();
  };

  const logout = async () => {
    try {
      await api.logout();
    } catch {
      // ignore
    }
    setAuthToken(null);
    setUser(null);
    setActiveAccountIdState(null);
    setTargetAccountId(null);
    setSharedAccounts([]);
  };

  const switchAccount = (accountId: string | null) => {
    if (accountId === user?.id) {
      accountId = null;
    }
    setActiveAccountIdState(accountId);
    setTargetAccountId(accountId);
  };

  const isViewingSharedAccount = !!activeAccountId && activeAccountId !== user?.id;

  let activeAccountName = user ? `${user.username} (Personal)` : '';
  let activePermission: SharePermission = 'write';

  if (isViewingSharedAccount) {
    const shared = sharedAccounts.find((a) => a.owner_id === activeAccountId);
    if (shared) {
      activeAccountName = `${shared.username} (Shared - ${shared.permission === 'write' ? 'Write' : 'Read Only'})`;
      activePermission = shared.permission;
    }
  }

  const updateUserSession = (updatedUser: User, token: string) => {
    setAuthToken(token);
    setUser(updatedUser);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        hasRoutes,
        hasHome,
        sharedAccounts,
        activeAccountId,
        activeAccountName,
        activePermission,
        isViewingSharedAccount,
        login,
        register,
        logout,
        switchAccount,
        refreshMe,
        updateUserSession,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
