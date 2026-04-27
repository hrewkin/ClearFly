import React from 'react';
import { Navigate, Route, Routes, NavLink, useLocation } from 'react-router-dom';
import './App.css';
import './index.css';
import HomePage from './pages/HomePage';
import FlightSearchPage from './pages/FlightSearchPage';
import BookingFlowPage from './pages/BookingFlowPage';
import NotificationsPage from './pages/NotificationsPage';
import OperationsPage from './pages/OperationsPage';
import BaggagePage from './pages/BaggagePage';
import AnalyticsPage from './pages/AnalyticsPage';
import ProfilePage from './pages/ProfilePage';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import MyBookingsPage from './pages/MyBookingsPage';
import { AuthProvider, isAdmin, useAuth } from './auth';

const NAV = [
  { to: '/', label: 'Обзор', icon: '◈', exact: true, roles: ['admin', 'passenger'] },
  { to: '/search', label: 'Поиск рейсов', icon: '▸', roles: ['admin', 'passenger'] },
  { to: '/my-bookings', label: 'Мои брони', icon: '⊞', roles: ['passenger'] },
  { to: '/notifications', label: 'Уведомления', icon: '◉', roles: ['admin', 'passenger'] },
  { to: '/operations', label: 'Операции', icon: '⎔', roles: ['admin'] },
  { to: '/baggage', label: 'Багаж', icon: '⊡', roles: ['admin', 'passenger'] },
  { to: '/analytics', label: 'Аналитика', icon: '⊿', roles: ['admin'] },
  { to: '/profile', label: 'Профиль', icon: '◎', roles: ['admin', 'passenger'] },
];

function Shell({ children }) {
  const { user, logout } = useAuth();
  const allowed = NAV.filter((n) => !user || n.roles.includes(user.role));
  return (
    <div className="layout">
      <nav className="sidebar">
        <div className="brand">
          <div className="logo-icon">✈</div>
          <h2>ClearFly</h2>
          <small className="brand-subtitle">Чистое небо</small>
        </div>
        <ul className="nav-links">
          {allowed.map((n) => (
            <li key={n.to}>
              <NavLink
                to={n.to}
                end={n.exact}
                className={({ isActive }) => isActive ? 'nav-item active' : 'nav-item'}
              >
                <span className="nav-icon" aria-hidden>{n.icon}</span>
                <span>{n.label}</span>
              </NavLink>
            </li>
          ))}
        </ul>
        <div className="sidebar-footer">
          <div className="ops-status">
            <span className="dot ok" />
            <span>Все системы работают</span>
          </div>
        </div>
        {user && (
          <div className="sidebar-user">
            <div>
              <div className="sidebar-user-name">{user.full_name}</div>
              <div className="sidebar-user-role">{isAdmin(user) ? 'Администратор' : 'Пассажир'}</div>
            </div>
            <button className="logout-btn" onClick={logout}>Выйти</button>
          </div>
        )}
      </nav>
      <main className="main-content">{children}</main>
    </div>
  );
}

function RequireAuth({ children, roles }) {
  const { user, loading } = useAuth();
  const location = useLocation();
  if (loading) return <div className="loading">Загрузка…</div>;
  if (!user) return <Navigate to="/login" state={{ from: location.pathname }} replace />;
  if (roles && !roles.includes(user.role)) return <Navigate to="/" replace />;
  return children;
}

function GuestOnly({ children }) {
  const { user, loading } = useAuth();
  if (loading) return <div className="loading">Загрузка…</div>;
  if (user) return <Navigate to="/" replace />;
  return children;
}

function AppRoutes() {
  const passengerOrAdmin = ['admin', 'passenger'];
  return (
    <Routes>
      <Route path="/login" element={<GuestOnly><LoginPage /></GuestOnly>} />
      <Route path="/register" element={<GuestOnly><RegisterPage /></GuestOnly>} />
      <Route
        path="/"
        element={<RequireAuth roles={passengerOrAdmin}><Shell><HomePage /></Shell></RequireAuth>}
      />
      <Route
        path="/search"
        element={<RequireAuth roles={passengerOrAdmin}><Shell><FlightSearchPage /></Shell></RequireAuth>}
      />
      <Route
        path="/book/:flightId"
        element={<RequireAuth roles={['passenger']}><Shell><BookingFlowPage /></Shell></RequireAuth>}
      />
      <Route
        path="/my-bookings"
        element={<RequireAuth roles={['passenger']}><Shell><MyBookingsPage /></Shell></RequireAuth>}
      />
      <Route
        path="/notifications"
        element={<RequireAuth roles={passengerOrAdmin}><Shell><NotificationsPage /></Shell></RequireAuth>}
      />
      <Route
        path="/operations"
        element={<RequireAuth roles={['admin']}><Shell><OperationsPage /></Shell></RequireAuth>}
      />
      <Route
        path="/baggage"
        element={<RequireAuth roles={passengerOrAdmin}><Shell><BaggagePage /></Shell></RequireAuth>}
      />
      <Route
        path="/analytics"
        element={<RequireAuth roles={['admin']}><Shell><AnalyticsPage /></Shell></RequireAuth>}
      />
      <Route
        path="/profile"
        element={<RequireAuth roles={passengerOrAdmin}><Shell><ProfilePage /></Shell></RequireAuth>}
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

function App() {
  return (
    <AuthProvider>
      <AppRoutes />
    </AuthProvider>
  );
}

export default App;
