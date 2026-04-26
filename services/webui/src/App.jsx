import React, { useState } from 'react';
import { Routes, Route, NavLink } from 'react-router-dom';
import './App.css';
import './index.css';
import BaggagePage from './pages/BaggagePage';
import AnalyticsPage from './pages/AnalyticsPage';
import ProfilePage from './pages/ProfilePage';

function DashboardPage() {
  const [bookingId, setBookingId] = useState('');
  const [bookingResult, setBookingResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchBooking = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setBookingResult(null);

    try {
      const res = await fetch(`http://localhost:8080/api/v1/bookings/${bookingId}`);
      if (!res.ok) {
        throw new Error('Booking not found');
      }
      const data = await res.json();
      setBookingResult(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const createFakeBooking = async () => {
    setLoading(true);
    setError('');
    setBookingResult(null);

    try {
      const flightId = "123e4567-e89b-12d3-a456-426614174000";
      const passengerId = "123e4567-e89b-12d3-a456-426614174001";
      const res = await fetch(`http://localhost:8080/api/v1/bookings`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ flight_id: flightId, passenger_id: passengerId })
      });
      if (!res.ok) {
        throw new Error('Failed to create booking');
      }
      const data = await res.json();
      setBookingResult({ ...data, isNew: true });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <header>
        <h1>Панель управления</h1>
        <p className="subtitle">Добро пожаловать. Готовы к следующему приключению?</p>
      </header>

      <section className="dashboard-cards">
        <div className="card glass-effect animate-in">
          <h3>Найти бронирование</h3>
          <p>Введите ID бронирования (UUID) для просмотра деталей.</p>
          <form onSubmit={fetchBooking} className="search-form">
            <input 
              type="text" 
              placeholder="e.g. 123e4567-e89b..." 
              value={bookingId}
              onChange={(e) => setBookingId(e.target.value)}
              required
            />
            <button type="submit" disabled={loading}>
              {loading ? 'Поиск...' : 'Найти'}
            </button>
          </form>
        </div>

        <div className="card glass-effect highlight animate-in" style={{ animationDelay: '0.1s' }}>
          <h3>Быстрое бронирование</h3>
          <p>Нужно тестовое бронирование? Создайте мгновенную демо-резервацию.</p>
          <button className="primary-btn" onClick={createFakeBooking} disabled={loading}>
            Создать демо-бронирование
          </button>
        </div>
      </section>

      {error && (
        <div className="alert error animate-in">
          <strong>Ошибка:</strong> {error}
        </div>
      )}

      {bookingResult && (
        <section className="results-section animate-in">
          <h2>{bookingResult.isNew ? 'Бронирование создано!' : 'Детали бронирования'}</h2>
          <div className="ticket">
            <div className="ticket-header">
              <span className="status">{bookingResult.status}</span>
              <span className="booking-id">ID: {bookingResult.id}</span>
            </div>
            <div className="ticket-body">
              <div className="info-group">
                <label>ID Рейса</label>
                <span>{bookingResult.flight_id}</span>
              </div>
              <div className="info-group">
                <label>ID Пассажира</label>
                <span>{bookingResult.passenger_id}</span>
              </div>
            </div>
          </div>
        </section>
      )}
    </>
  );
}

function App() {
  return (
    <div className="layout">
      <nav className="sidebar">
        <div className="brand">
          <div className="logo-icon">✈</div>
          <h2>ClearFly</h2>
        </div>
        <ul className="nav-links">
          <li><NavLink to="/" className={({ isActive }) => isActive ? 'nav-item active' : 'nav-item'}>Бронирования</NavLink></li>
          <li><NavLink to="/baggage" className={({ isActive }) => isActive ? 'nav-item active' : 'nav-item'}>Багаж</NavLink></li>
          <li><NavLink to="/analytics" className={({ isActive }) => isActive ? 'nav-item active' : 'nav-item'}>Аналитика</NavLink></li>
          <li><NavLink to="/profile" className={({ isActive }) => isActive ? 'nav-item active' : 'nav-item'}>Профиль</NavLink></li>
        </ul>
      </nav>

      <main className="main-content">
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/baggage" element={<BaggagePage />} />
          <Route path="/analytics" element={<AnalyticsPage />} />
          <Route path="/profile" element={<ProfilePage />} />
        </Routes>
      </main>
    </div>
  );
}

export default App;
