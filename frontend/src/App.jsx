import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import { Navbar } from './components/Navbar';
import { Watermark } from './components/Security';

// Pages
import Login from './pages/Login';
import Directory from './pages/Directory';
import Level2Dashboard from './pages/level2/Dashboard';
import CreateReferral from './pages/level2/CreateReferral';
import Notifications from './pages/level2/Notifications';
import ChuDashboard from './pages/chu/Dashboard';
import HistoryTab from './pages/shared/History';
import AdminDashboard from './pages/admin/Dashboard';
import AuditLogs from './pages/admin/AuditLogs';
import AnalystDashboard from './pages/analyst/Stats';

// Shared layout wrapper to inject standard nav and security overlays
const MainLayout = ({ children }) => {
  return (
    <div className="min-h-screen bg-gray-50 pb-20 md:pb-0 flex flex-col">
      <Watermark />
      <Navbar />
      <main className="flex-1 w-full max-w-7xl mx-auto md:p-6 p-4">
        {children}
      </main>
    </div>
  );
};

// Route protection component
const ProtectedRoute = ({ children, allowedRoles }) => {
  const { user, isAuthenticated } = useAuth();

  if (!isAuthenticated) return <Navigate to="/login" replace />;
  if (allowedRoles && !allowedRoles.includes(user.role)) return <Navigate to="/directory" replace />;

  return children;
};

// Dynamic Dashboard router based on authenticated user's role
const RoleBasedDashboard = () => {
  const { user } = useAuth();
  if (user?.role === 'LEVEL_2_DOC') return <Level2Dashboard />;
  if (user?.role === 'CHU_DOC') return <ChuDashboard />;
  if (user?.role === 'SUPER_ADMIN') return <AdminDashboard />;
  if (user?.role === 'ANALYST') return <AnalystDashboard />;
  return <Navigate to="/directory" replace />; // Fallback
};

function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />

      {/* ── Protected Routes ── */}
      <Route path="/" element={<ProtectedRoute><MainLayout><RoleBasedDashboard /></MainLayout></ProtectedRoute>} />
      <Route path="/dashboard" element={<ProtectedRoute><MainLayout><RoleBasedDashboard /></MainLayout></ProtectedRoute>} />

      {/* Accessible to all logged-in users */}
      <Route path="/directory" element={<ProtectedRoute><MainLayout><Directory /></MainLayout></ProtectedRoute>} />

      {/* ── Level 2 Specific ── */}
      <Route path="/referrals/new" element={
        <ProtectedRoute allowedRoles={['LEVEL_2_DOC']}>
          <MainLayout><CreateReferral /></MainLayout>
        </ProtectedRoute>
      } />
      <Route path="/notifications" element={
        <ProtectedRoute allowedRoles={['LEVEL_2_DOC']}>
          <MainLayout><Notifications /></MainLayout>
        </ProtectedRoute>
      } />

      {/* ── Shared or Admin ── */}
      <Route path="/history" element={
        <ProtectedRoute allowedRoles={['LEVEL_2_DOC', 'CHU_DOC']}>
          <MainLayout><HistoryTab /></MainLayout>
        </ProtectedRoute>
      } />
      <Route path="/analyst" element={
        <ProtectedRoute allowedRoles={['ANALYST', 'SUPER_ADMIN']}>
          <MainLayout><AnalystDashboard /></MainLayout>
        </ProtectedRoute>
      } />
      <Route path="/admin" element={
        <ProtectedRoute allowedRoles={['SUPER_ADMIN']}>
          <MainLayout><AdminDashboard /></MainLayout>
        </ProtectedRoute>
      } />
      <Route path="/admin/audit-logs" element={
        <ProtectedRoute allowedRoles={['SUPER_ADMIN']}>
          <MainLayout><AuditLogs /></MainLayout>
        </ProtectedRoute>
      } />

      {/* Catch-all redirect */}
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <AppRoutes />
      </BrowserRouter>
    </AuthProvider>
  );
}
