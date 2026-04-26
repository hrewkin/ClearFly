import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { airportShort, api, durationLabel, formatPrice, formatTime } from '../api';

export default function HomePage() {
  const [flights, setFlights] = useState([]);
  const [notifications, setNotifications] = useState([]);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    let mounted = true;
    Promise.all([
      api.upcomingFlights().catch(() => []),
      api.listNotifications().catch(() => []),
    ]).then(([f, n]) => {
      if (!mounted) return;
      setFlights(f || []);
      setNotifications((n || []).slice(0, 4));
    }).catch((err) => mounted && setError(err.message));
    return () => { mounted = false; };
  }, []);

  const totalSeats = flights.reduce((s, f) => s + (f.total_seats || 0), 0);
  const availableSeats = flights.reduce((s, f) => s + (f.available_seats || 0), 0);
  const loadFactor = totalSeats === 0 ? 0 : Math.round(((totalSeats - availableSeats) / totalSeats) * 100);

  return (
    <>
      <header className="page-header">
        <div>
          <h1>Чистое небо для каждого пассажира</h1>
          <p className="subtitle">Микросервисная платформа для бронирования рейсов с динамическими тарифами и реальной шиной событий.</p>
        </div>
        <button className="primary-btn" onClick={() => navigate('/search')}>Найти рейс →</button>
      </header>

      <section className="kpi-row">
        <KpiCard icon="✈" label="Активных рейсов" value={flights.length} hint="на ближайшие сутки" />
        <KpiCard icon="🪑" label="Свободных мест" value={availableSeats} hint={`из ${totalSeats}`} />
        <KpiCard icon="📊" label="Загрузка флота" value={`${loadFactor}%`} hint="в реальном времени" tone={loadFactor > 70 ? 'warning' : 'ok'} />
        <KpiCard icon="🔔" label="События за сутки" value={notifications.length} hint="нотификации" />
      </section>

      <section className="home-grid">
        <div className="card glass-effect">
          <div className="card-head">
            <h3>Ближайшие рейсы</h3>
            <Link to="/search" className="link">Все рейсы →</Link>
          </div>
          <ul className="home-flights">
            {flights.slice(0, 5).map((f) => (
              <li key={f.id} className="home-flight">
                <div className="time">
                  <strong>{formatTime(f.departure_time)}</strong>
                  <small>{f.flight_number}</small>
                </div>
                <div className="route">
                  <strong>{airportShort(f.origin)} → {airportShort(f.destination)}</strong>
                  <small>{durationLabel(f.departure_time, f.arrival_time)} · выход {f.gate || '—'}</small>
                </div>
                <div className="meta">
                  <span className={`tag tone-${(f.status || 'SCHEDULED').toLowerCase()}`}>{f.status || 'SCHEDULED'}</span>
                  <button className="ghost-btn" onClick={() => navigate(`/book/${f.id}`)}>Забронировать</button>
                </div>
              </li>
            ))}
            {flights.length === 0 && <li className="empty-state">Загружаем расписание…</li>}
          </ul>
          {error && <div className="alert error">{error}</div>}
        </div>

        <div className="card glass-effect">
          <div className="card-head">
            <h3>Лента событий</h3>
            <Link to="/notifications" className="link">Все уведомления →</Link>
          </div>
          <ul className="home-events">
            {notifications.length === 0 && <li className="empty-state">Пока тихо. Создайте инцидент на странице «Операции».</li>}
            {notifications.map((n) => (
              <li key={n.id} className="home-event">
                <span className="dot" />
                <div>
                  <strong>{n.title}</strong>
                  <small>{n.content}</small>
                </div>
              </li>
            ))}
          </ul>
        </div>
      </section>

      <section className="features">
        <FeatureCard
          title="Атомарное распределение мест"
          desc="SELECT … FOR UPDATE предотвращает двойное бронирование даже при пиковой нагрузке."
        />
        <FeatureCard
          title="Шина событий RabbitMQ"
          desc="Бронирование, изменения рейсов и регистрация публикуются в очередь и доходят до пассажира."
        />
        <FeatureCard
          title="Динамические тарифы"
          desc="Цена пересчитывается по фактору загрузки: до +50% для последних мест."
        />
      </section>

      <section className="cta-row">
        <button className="primary-btn" onClick={() => navigate('/search')}>Поиск рейса</button>
        <button className="ghost-btn" onClick={() => navigate('/operations')}>Открыть центр операций</button>
      </section>
    </>
  );
}

function KpiCard({ icon, label, value, hint, tone = 'default' }) {
  return (
    <div className={`kpi-card kpi-${tone}`}>
      <div className="kpi-icon">{icon}</div>
      <div className="kpi-info">
        <strong>{value}</strong>
        <span>{label}</span>
        <small>{hint}</small>
      </div>
    </div>
  );
}

function FeatureCard({ title, desc }) {
  return (
    <article className="feature-card">
      <h4>{title}</h4>
      <p>{desc}</p>
    </article>
  );
}

// Re-export so tooling knows it's used.
export { formatPrice };
