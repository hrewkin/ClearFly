import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api, airportLabel, airportShort, durationLabel, flightStatusLabel, formatDate, formatPrice, formatTime } from '../api';
import { isAdmin, useAuth } from '../auth';

const POPULAR_ROUTES = [
  { from: 'SVO', to: 'LED' },
  { from: 'SVO', to: 'AER' },
  { from: 'LED', to: 'KJA' },
  { from: 'SVO', to: 'KZN' },
];

function todayInputValue() {
  const d = new Date();
  const yyyy = d.getFullYear();
  const mm = `${d.getMonth() + 1}`.padStart(2, '0');
  const dd = `${d.getDate()}`.padStart(2, '0');
  return `${yyyy}-${mm}-${dd}`;
}

export default function FlightSearchPage() {
  const [origin, setOrigin] = useState('SVO');
  const [destination, setDestination] = useState('LED');
  const [date, setDate] = useState(todayInputValue());
  const [flights, setFlights] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [searched, setSearched] = useState(false);
  const navigate = useNavigate();
  const { user } = useAuth();
  const admin = isAdmin(user);

  useEffect(() => {
    let mounted = true;
    api.upcomingFlights()
      .then((data) => {
        if (mounted) setFlights(data || []);
      })
      .catch((err) => mounted && setError(err.message));
    return () => { mounted = false; };
  }, []);

  const submit = async (event) => {
    event.preventDefault();
    setLoading(true);
    setError('');
    setSearched(true);
    try {
      const data = await api.searchFlights({ origin, destination, date });
      setFlights(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const applyPopular = (route) => {
    setOrigin(route.from);
    setDestination(route.to);
  };

  return (
    <>
      <header className="page-header">
        <div>
          <h1>Поиск рейсов</h1>
          <p className="subtitle">Выберите маршрут, дату и место — забронируем за 30 секунд.</p>
        </div>
      </header>

      <section className="card glass-effect search-card animate-in">
        <form className="search-grid" onSubmit={submit}>
          <label className="field">
            <span>Откуда</span>
            <input value={origin} onChange={(e) => setOrigin(e.target.value.toUpperCase())} maxLength={3} required />
            <small>{airportLabel(origin)}</small>
          </label>
          <div className="field-arrow" aria-hidden>→</div>
          <label className="field">
            <span>Куда</span>
            <input value={destination} onChange={(e) => setDestination(e.target.value.toUpperCase())} maxLength={3} required />
            <small>{airportLabel(destination)}</small>
          </label>
          <label className="field">
            <span>Дата</span>
            <input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
          </label>
          <button className="primary-btn" type="submit" disabled={loading}>
            {loading ? 'Ищем…' : 'Найти рейсы'}
          </button>
        </form>
        <div className="popular-routes">
          <span>Популярные направления:</span>
          {POPULAR_ROUTES.map((r) => (
            <button type="button" key={`${r.from}-${r.to}`} className="chip" onClick={() => applyPopular(r)}>
              {airportShort(r.from)} → {airportShort(r.to)}
            </button>
          ))}
        </div>
      </section>

      {error && (
        <div className="alert error animate-in">
          <strong>Ошибка:</strong> {error}
        </div>
      )}

      <section className="flight-list">
        {flights.length === 0 && !loading && (
          <div className="empty-state">
            {searched
              ? 'По заданным критериям рейсов нет. Попробуйте другие даты или маршрут.'
              : 'Ищем ближайшие рейсы…'}
          </div>
        )}
        {flights.map((flight, idx) => (
          <article key={flight.id} className="flight-card animate-in" style={{ animationDelay: `${idx * 0.05}s` }}>
            <div className="flight-meta">
              <span className="flight-number">{flight.flight_number}</span>
              <span className={`flight-status status-${(flight.status || 'SCHEDULED').toLowerCase()}`}>
                {flightStatusLabel(flight.status)}
              </span>
              <span className="flight-aircraft">{flight.aircraft_type}</span>
            </div>
            <div className="flight-route">
              <div className="route-side">
                <div className="time">{formatTime(flight.departure_time)}</div>
                <div className="airport">{airportShort(flight.origin)}</div>
                <div className="muted">{formatDate(flight.departure_time)}</div>
              </div>
              <div className="route-line">
                <div className="line" />
                <div className="duration">{durationLabel(flight.departure_time, flight.arrival_time)}</div>
                <div className="line" />
              </div>
              <div className="route-side">
                <div className="time">{formatTime(flight.arrival_time)}</div>
                <div className="airport">{airportShort(flight.destination)}</div>
                <div className="muted">{formatDate(flight.arrival_time)}</div>
              </div>
            </div>
            <div className="flight-actions">
              <div className="seats-info">
                <strong>{flight.available_seats}</strong>
                <span>свободно из {flight.total_seats}</span>
              </div>
              {admin ? (
                <span className="muted small">Админ не бронирует</span>
              ) : (
                <button className="primary-btn" onClick={() => navigate(`/book/${flight.id}`)}>
                  Выбрать место
                </button>
              )}
            </div>
          </article>
        ))}
      </section>
    </>
  );
}

export { formatPrice };
