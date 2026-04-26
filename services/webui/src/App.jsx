import React from 'react';
import { Routes, Route, NavLink } from 'react-router-dom';
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

const NAV = [
  { to: '/', label: 'Обзор', icon: '✦', exact: true },
  { to: '/search', label: 'Поиск рейсов', icon: '✈' },
  { to: '/notifications', label: 'Уведомления', icon: '🔔' },
  { to: '/operations', label: 'Операции', icon: '⚙' },
  { to: '/baggage', label: 'Багаж', icon: '🧳' },
  { to: '/analytics', label: 'Аналитика', icon: '📊' },
  { to: '/profile', label: 'Профиль', icon: '👤' },
];

function App() {
  return (
    <div className="layout">
      <nav className="sidebar">
        <div className="brand">
          <div className="logo-icon">✈</div>
          <h2>ClearFly</h2>
          <small className="brand-subtitle">Чистое небо</small>
        </div>
        <ul className="nav-links">
          {NAV.map((n) => (
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
      </nav>

      <main className="main-content">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/search" element={<FlightSearchPage />} />
          <Route path="/book/:flightId" element={<BookingFlowPage />} />
          <Route path="/notifications" element={<NotificationsPage />} />
          <Route path="/operations" element={<OperationsPage />} />
          <Route path="/baggage" element={<BaggagePage />} />
          <Route path="/analytics" element={<AnalyticsPage />} />
          <Route path="/profile" element={<ProfilePage />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;
