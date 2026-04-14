import React, { useState } from 'react';
import './App.css';
import './index.css';

function App() {
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
      // Proxy is running on gateway :8080
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
    <div className="layout">
      <nav className="sidebar">
        <div className="brand">
          <div className="logo-icon">✈</div>
          <h2>Clear Sky</h2>
        </div>
        <ul className="nav-links">
          <li className="active">Dashboard</li>
          <li>My Flights</li>
          <li>Settings</li>
        </ul>
      </nav>

      <main className="main-content">
        <header>
          <h1>Passenger Dashboard</h1>
          <p className="subtitle">Welcome back. Ready for your next adventure?</p>
        </header>

        <section className="dashboard-cards">
          <div className="card glass-effect animate-in">
            <h3>Find Your Booking</h3>
            <p>Enter your Booking ID (UUID) to see details.</p>
            <form onSubmit={fetchBooking} className="search-form">
              <input 
                type="text" 
                placeholder="e.g. 123e4567-e89b..." 
                value={bookingId}
                onChange={(e) => setBookingId(e.target.value)}
                required
              />
              <button type="submit" disabled={loading}>
                {loading ? 'Searching...' : 'Search'}
              </button>
            </form>
          </div>

          <div className="card glass-effect highlight animate-in" style={{ animationDelay: '0.1s' }}>
            <h3>Fast Track</h3>
            <p>Need a quick test booking? Generate an instant test reservation.</p>
            <button className="primary-btn" onClick={createFakeBooking} disabled={loading}>
              Create Demo Booking
            </button>
          </div>
        </section>

        {error && (
          <div className="alert error animate-in">
            <strong>Error:</strong> {error}
          </div>
        )}

        {bookingResult && (
          <section className="results-section animate-in">
            <h2>{bookingResult.isNew ? 'Booking Created!' : 'Booking Details'}</h2>
            <div className="ticket">
              <div className="ticket-header">
                <span className="status">{bookingResult.status}</span>
                <span className="booking-id">ID: {bookingResult.id}</span>
              </div>
              <div className="ticket-body">
                <div className="info-group">
                  <label>Flight ID</label>
                  <span>{bookingResult.flight_id}</span>
                </div>
                <div className="info-group">
                  <label>Passenger ID</label>
                  <span>{bookingResult.passenger_id}</span>
                </div>
              </div>
            </div>
          </section>
        )}
      </main>
    </div>
  );
}

export default App;
