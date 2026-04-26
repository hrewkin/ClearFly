import React, { useEffect, useMemo, useState } from 'react';
import { airportShort, api, formatPrice, formatTime } from '../api';

function pricingTier(factor) {
  if (factor > 80) return { label: '×1.5 — высокий спрос', tone: 'danger', multiplier: 1.5 };
  if (factor > 50) return { label: '×1.2 — средний спрос', tone: 'warning', multiplier: 1.2 };
  return { label: '×1.0 — базовая цена', tone: 'success', multiplier: 1.0 };
}

function loadColor(factor) {
  if (factor > 80) return 'var(--danger)';
  if (factor > 50) return 'var(--warning)';
  return 'var(--success)';
}

export default function AnalyticsPage() {
  const [flights, setFlights] = useState([]);
  const [loads, setLoads] = useState({}); // flight_id -> analytics
  const [selectedId, setSelectedId] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const selected = useMemo(
    () => flights.find((f) => f.id === selectedId) || flights[0] || null,
    [flights, selectedId],
  );
  const selectedLoad = selected ? loads[selected.id] : null;

  const load = async () => {
    try {
      const upcoming = await api.upcomingFlights();
      const list = upcoming || [];
      setFlights(list);

      const analytics = await Promise.all(
        list.slice(0, 10).map((f) =>
          api.flightLoadFactor(f.id)
            .then((data) => ({ id: f.id, data }))
            .catch(() => ({ id: f.id, data: null })),
        ),
      );
      const map = {};
      analytics.forEach(({ id, data }) => {
        if (data) map[id] = data;
      });
      setLoads(map);
      setError('');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    const t = setInterval(load, 7000);
    return () => clearInterval(t);
  }, []);

  if (loading && flights.length === 0) {
    return <div className="loading">Загружаем данные…</div>;
  }

  const fleetLoads = flights.slice(0, 10).map((f) => loads[f.id]?.analytics?.load_factor ?? 0);
  const avgLoad = fleetLoads.length ? Math.round(fleetLoads.reduce((a, b) => a + b, 0) / fleetLoads.length) : 0;
  const maxLoad = fleetLoads.length ? Math.max(...fleetLoads) : 0;
  const totalBookings = flights.reduce(
    (acc, f) => acc + (loads[f.id]?.analytics?.total_bookings ?? 0),
    0,
  );

  const factor = selectedLoad?.analytics?.load_factor ?? 0;
  const tier = pricingTier(factor);
  const suggested = selectedLoad?.suggested_price ?? 0;
  const gaugeDeg = Math.min(factor, 100) * 3.6;

  return (
    <>
      <header className="page-header">
        <div>
          <h1>Аналитика рейсов</h1>
          <p className="subtitle">Загрузка флота в реальном времени и рекомендации динамического ценообразования.</p>
        </div>
      </header>

      {error && <div className="alert error">{error}</div>}

      <section className="kpi-row">
        <div className="kpi-card">
          <div className="kpi-icon">📊</div>
          <div className="kpi-info">
            <strong>{avgLoad}%</strong>
            <span>Средняя загрузка</span>
            <small>по {fleetLoads.length} ближайшим рейсам</small>
          </div>
        </div>
        <div className="kpi-card kpi-warning">
          <div className="kpi-icon">🔥</div>
          <div className="kpi-info">
            <strong>{maxLoad}%</strong>
            <span>Пиковая загрузка</span>
            <small>автоматический коэффициент цены</small>
          </div>
        </div>
        <div className="kpi-card">
          <div className="kpi-icon">🎟️</div>
          <div className="kpi-info">
            <strong>{totalBookings}</strong>
            <span>Активных бронирований</span>
            <small>на ближайшие рейсы</small>
          </div>
        </div>
      </section>

      <section className="analytics-grid-main">
        <div className="card glass-effect">
          <div className="card-head">
            <h3>Рейс для анализа</h3>
            <small className="muted">{flights.length} в горизонте 24 ч</small>
          </div>
          <ul className="analytics-flight-list">
            {flights.slice(0, 10).map((f) => {
              const l = loads[f.id]?.analytics?.load_factor ?? 0;
              const isActive = selected && selected.id === f.id;
              return (
                <li
                  key={f.id}
                  className={`analytics-flight ${isActive ? 'active' : ''}`}
                  onClick={() => setSelectedId(f.id)}
                >
                  <div className="analytics-flight-head">
                    <strong>{f.flight_number}</strong>
                    <span className="muted">{airportShort(f.origin)} → {airportShort(f.destination)} · {formatTime(f.departure_time)}</span>
                  </div>
                  <div className="load-bar">
                    <div
                      className="load-bar-fill"
                      style={{ width: `${Math.min(l, 100)}%`, background: loadColor(l) }}
                    />
                  </div>
                  <small className="muted">{l}% · {loads[f.id]?.analytics?.total_bookings ?? 0} мест</small>
                </li>
              );
            })}
            {flights.length === 0 && <li className="empty-state">Нет рейсов в ближайшие сутки.</li>}
          </ul>
        </div>

        {selected && (
          <div className="card glass-effect">
            <div className="card-head">
              <div>
                <h3>{selected.flight_number}</h3>
                <small className="muted">{airportShort(selected.origin)} → {airportShort(selected.destination)} · {formatTime(selected.departure_time)}</small>
              </div>
              <span className={`tag tone-${tier.tone}`}>{tier.label}</span>
            </div>

            <div className="gauge-wrap">
              <div
                className="gauge"
                style={{ background: `conic-gradient(${loadColor(factor)} ${gaugeDeg}deg, rgba(255,255,255,0.08) ${gaugeDeg}deg)` }}
              >
                <div className="gauge-inner">
                  <strong>{factor}%</strong>
                  <small>Load factor</small>
                </div>
              </div>
              <div className="gauge-legend">
                <div><small>Забронировано</small><strong>{selectedLoad?.analytics?.total_bookings ?? 0}</strong></div>
                <div><small>Вместимость</small><strong>{selected.aircraft ? '—' : 150}</strong></div>
                <div><small>Рекомендуемая цена</small><strong>{formatPrice(suggested)}</strong></div>
                <div><small>Множитель</small><strong>×{tier.multiplier}</strong></div>
              </div>
            </div>

            <div className="pricing-rules">
              <h4>Правила динамического ценообразования</h4>
              <ul>
                <li className={factor <= 50 ? 'rule-active' : ''}><span>≤ 50%</span> базовый тариф ×1.0</li>
                <li className={factor > 50 && factor <= 80 ? 'rule-active' : ''}><span>50–80%</span> ×1.2 — повышенный спрос</li>
                <li className={factor > 80 ? 'rule-active' : ''}><span>&gt; 80%</span> ×1.5 — пиковая загрузка</li>
              </ul>
            </div>
          </div>
        )}
      </section>
    </>
  );
}
