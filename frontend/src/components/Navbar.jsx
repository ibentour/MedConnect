import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { LogOut, Home, FilePlus, Bell, Activity, Users, Search, Clock, Settings, BarChart3 } from 'lucide-react';

export const Navbar = () => {
  const { user, logout } = useAuth();
  const location = useLocation();

  if (!user) return null;

  const isLevel2 = user.role === 'LEVEL_2_DOC';
  const isCHU = user.role === 'CHU_DOC';
  const isSuperAdmin = user.role === 'SUPER_ADMIN';
  const isAnalyst = user.role === 'ANALYST';

  // Helper to highlight active tab
  const isActive = (path) => location.pathname === path;
  
  // Base classes for nav links
  const defaultBaseClass = "flex flex-col md:flex-row items-center justify-center p-2 text-sm font-medium transition-colors";
  
  const getLinkClasses = (path) => {
    return `${defaultBaseClass} ${
      isActive(path) 
        ? "text-brand-600 bg-brand-50 rounded-lg" 
        : "text-gray-500 hover:text-gray-900 hover:bg-gray-100 rounded-lg"
    }`;
  };

  return (
    <nav className="bg-white border-b border-gray-200 sticky top-0 z-[100] safe-area-top shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          
          {/* Brand/Logo + Facility Name */}
          <div className="flex items-center">
            <Link to="/dashboard" className="flex items-center gap-2">
              <Activity className="h-8 w-8 text-brand-500" />
              <div className="flex flex-col leading-tight">
                <span className="font-bold text-gray-900 text-lg tracking-tight">
                  MedConnect <span className="text-brand-600 font-normal">Oriental</span>
                </span>
                <span className="text-xs text-gray-500 truncate max-w-[150px] md:max-w-xs" title={user.facility_name}>
                  {user.facility_name}
                </span>
              </div>
            </Link>
          </div>

          {/* Desktop Navigation Links */}
          <div className="hidden md:flex flex-1 justify-center items-center px-8 border-x border-gray-100 mx-6">
            <div className="flex space-x-2">
              {!isSuperAdmin && (
                <Link to="/dashboard" className={getLinkClasses('/dashboard')}>
                  {isAnalyst ? <BarChart3 className="w-5 h-5 mr-1 text-purple-600" /> : <Home className="w-5 h-5 mr-1" />}
                  <span>{isAnalyst ? 'Analytique' : 'Dashboard'}</span>
                </Link>
              )}
              
              <Link to="/directory" className={getLinkClasses('/directory')}>
                <Search className="w-5 h-5 mr-1" />
                <span>Directory</span>
              </Link>

              {(isSuperAdmin) && (
                <Link to="/analyst" className={getLinkClasses('/analyst')}>
                  <BarChart3 className="w-5 h-5 mr-1 text-purple-600" />
                  <span className="font-bold">Analytique</span>
                </Link>
              )}

              {(isLevel2 || isCHU) && (
                <Link to="/history" className={getLinkClasses('/history')}>
                  <Clock className="w-5 h-5 mr-1" />
                  <span>Historique</span>
                </Link>
              )}
              
              {isLevel2 && (
                <>
                  <Link to="/referrals/new" className={getLinkClasses('/referrals/new')}>
                    <FilePlus className="w-5 h-5 mr-1" />
                    <span>Nouveau Dossier</span>
                  </Link>
                  <Link to="/notifications" className={getLinkClasses('/notifications')}>
                    <Bell className="w-5 h-5 mr-1" />
                    <span>Alertes</span>
                  </Link>
                </>
              )}

              {isSuperAdmin && (
                <Link to="/admin" className={getLinkClasses('/admin')}>
                  <Settings className="w-5 h-5 mr-1" />
                  <span className="font-bold">Super Admin</span>
                </Link>
              )}
            </div>
          </div>

          {/* User Menu / Logout */}
          <div className="flex items-center gap-4">
            <div className="hidden md:flex flex-col items-end leading-tight">
              <span className="text-sm font-medium text-gray-900">Dr. {user.username}</span>
              <span className="text-xs text-brand-600 bg-brand-50 px-2 py-0.5 rounded-full mt-1 font-semibold tracking-wide border border-brand-100">
                {user.role}
              </span>
            </div>
            
            <button 
              onClick={logout}
              className="p-2 text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors flex items-center"
              aria-label="Logout"
            >
              <LogOut className="w-5 h-5 md:mr-1" />
              <span className="hidden md:block text-sm font-medium">Exit</span>
            </button>
          </div>
        </div>
      </div>

      {/* Mobile Bottom Navigation (Visible only on small screens) */}
      <div className="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 z-[100] pb-safe">
        <div className="flex justify-around items-center h-16 px-2">
          {!isSuperAdmin && (
            <Link to="/dashboard" className={`flex flex-col items-center p-2 flex-1 ${isActive('/dashboard') ? 'text-brand-600' : 'text-gray-500'}`}>
              {isAnalyst ? <BarChart3 className="w-6 h-6 mb-1" /> : <Home className="w-6 h-6 mb-1" />}
              <span className="text-[10px] uppercase font-semibold">{isAnalyst ? 'Stats' : 'Home'}</span>
            </Link>
          )}
          
          <Link to="/directory" className={`flex flex-col items-center p-2 flex-1 ${isActive('/directory') ? 'text-brand-600' : 'text-gray-500'}`}>
            <Search className="w-6 h-6 mb-1" />
            <span className="text-[10px] uppercase font-semibold">Directory</span>
          </Link>

          {isLevel2 && (
            <>
              <Link to="/referrals/new" className="flex flex-col items-center justify-center p-2 flex-1 -mt-6">
                <div className={`rounded-full p-3 shadow-md border-[3px] border-white ${isActive('/referrals/new') ? 'bg-brand-600 text-white' : 'bg-brand-500 text-white hover:bg-brand-600'}`}>
                  <FilePlus className="w-6 h-6" />
                </div>
                <span className={`text-[10px] uppercase font-semibold mt-1 ${isActive('/referrals/new') ? 'text-brand-600' : 'text-gray-500'}`}>New</span>
              </Link>
              
              <Link to="/notifications" className={`flex flex-col items-center p-2 flex-1 ${isActive('/notifications') ? 'text-brand-600' : 'text-gray-500'}`}>
                <div className="relative">
                  <Bell className="w-6 h-6 mb-1" />
                  {/* Optional unread badge logic here */}
                  <span className="absolute -top-1 -right-1 flex h-3 w-3">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
                    <span className="relative inline-flex rounded-full h-3 w-3 bg-red-500"></span>
                  </span>
                </div>
                <span className="text-[10px] uppercase font-semibold">Alerts</span>
              </Link>
            </>
          )}

          {isCHU && (
             <Link to="/history" className={`flex flex-col items-center p-2 flex-1 ${isActive('/history') ? 'text-brand-600' : 'text-gray-500'}`}>
               <Clock className="w-6 h-6 mb-1" />
               <span className="text-[10px] uppercase font-semibold">Historique</span>
             </Link>
          )}
          
          {isSuperAdmin && (
             <Link to="/admin" className={`flex flex-col items-center p-2 flex-1 ${isActive('/admin') ? 'text-brand-600' : 'text-gray-500'}`}>
               <Settings className="w-6 h-6 mb-1" />
               <span className="text-[10px] uppercase font-semibold">Admin</span>
             </Link>
          )}

          {(isSuperAdmin) && (
             <Link to="/analyst" className={`flex flex-col items-center p-2 flex-1 ${isActive('/analyst') ? 'text-brand-600' : 'text-gray-500'}`}>
               <BarChart3 className="w-6 h-6 mb-1" />
               <span className="text-[10px] uppercase font-semibold">Stats</span>
             </Link>
          )}
        </div>
      </div>
    </nav>
  );
};
