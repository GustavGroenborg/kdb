import React, { useState } from 'react';
import { AuthProvider, useAuth } from './context/AuthContext';
import { Navbar, ActiveTab } from './components/Navbar';
import { AuthPage } from './pages/AuthPage';
import { RegisterTab } from './pages/RegisterTab';
import { OverviewTab } from './pages/OverviewTab';
import { SettingsTab } from './pages/SettingsTab';

const MainApp: React.FC = () => {
  const { user, isLoading } = useAuth();
  const [activeTab, setActiveTab] = useState<ActiveTab>('register');

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-50 text-slate-500">
        <div className="text-center">
          <div className="animate-spin w-9 h-9 border-4 border-brand-500 border-t-transparent rounded-full mx-auto mb-3"></div>
          <div className="text-sm font-medium">Connecting to KDB...</div>
        </div>
      </div>
    );
  }

  if (!user) {
    return <AuthPage />;
  }

  return (
    <div className="min-h-screen bg-slate-50 flex flex-col">
      <Navbar activeTab={activeTab} onTabChange={setActiveTab} />
      <main className="flex-1 pb-16">
        {activeTab === 'register' && <RegisterTab onGoToSettings={() => setActiveTab('settings')} />}
        {activeTab === 'overview' && <OverviewTab />}
        {activeTab === 'settings' && <SettingsTab />}
      </main>
    </div>
  );
};

export const App: React.FC = () => {
  return (
    <AuthProvider>
      <MainApp />
    </AuthProvider>
  );
};

export default App;
