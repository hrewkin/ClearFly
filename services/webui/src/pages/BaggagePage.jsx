import React, { useEffect, useMemo, useState } from 'react';
import { BAGGAGE_STAGES, airportShort, api, baggageStageIndex, formatTime } from '../api';

function formatRelative(iso) {
  if (!iso) return '';
  const ms = Date.now() - new Date(iso).getTime();
  if (ms < 60 * 1000) return 'только что';
  if (ms < 60 * 60 * 1000) return `${Math.round(ms / 60000)} мин назад`;
  if (ms < 24 * 60 * 60 * 1000) return `${Math.round(ms / 3600000)} ч назад`;
  return new Date(iso).toLocaleDateString('ru-RU');
}

function stageTone(index) {
  if (index <= 1) return 'info';
  if (index <= 3) return 'warning';
  return 'success';
}

export default function BaggagePage() {
  const [baggage, setBaggage] = useState([]);
  const [flights, setFlights] = useState([]);
  const [selectedId, setSelectedId] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [showRegister, setShowRegister] = useState(false);
  const [form, setForm] = useState({ passenger_id: '', flight_id: '' });
  const [submitting, setSubmitting] = useState(false);

  const selected = useMemo(
    () => baggage.find((b) => b.id === selectedId) || baggage[0] || null,
    [baggage, selectedId],
  );

  const load = async () => {
    try {
      const [bags, upcoming] = await Promise.all([
        api.listBaggage({ limit: 50 }),
        api.upcomingFlights(),
      ]);
      setBaggage(bags || []);
      setFlights(upcoming || []);
      setError('');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    const t = setInterval(load, 5000);
    return () => clearInterval(t);
  }, []);

  useEffect(() => {
    if (!success) return;
    const t = setTimeout(() => setSuccess(''), 2500);
    return () => clearTimeout(t);
  }, [success]);

  const flightByID = useMemo(() => {
    const m = {};
    flights.forEach((f) => { m[f.id] = f; });
    return m;
  }, [flights]);

  const onRegister = async (e) => {
    e.preventDefault();
    if (!form.passenger_id) return;
    setSubmitting(true);
    setError('');
    try {
      const bag = await api.createBaggage({
        passenger_id: form.passenger_id.trim(),
        flight_id: form.flight_id || undefined,
      });
      await load();
      setSelectedId(bag.id);
      setShowRegister(false);
      setSuccess(`Багаж ${bag.id.slice(0, 8)} зарегистрирован`);
      setForm({ passenger_id: '', flight_id: '' });
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  };

  const onScan = async (bag) => {
    if (!bag) return;
    try {
      await api.scanBaggage(bag.id);
      await load();
      setSuccess('Скан зарегистрирован');
    } catch (err) {
      setError(err.message);
    }
  };

  if (loading && baggage.length === 0) {
    return <div className="loading">Загружаем багаж…</div>;
  }

  const selectedIdx = selected ? baggageStageIndex(selected.status) : -1;

  return (
    <>
      <header className="page-header">
        <div>
          <h1>Трекинг багажа</h1>
          <p className="subtitle">Статус каждого тега — от стойки регистрации до ленты выдачи. Обновляется автоматически.</p>
        </div>
        <button className="ghost-btn" onClick={() => setShowRegister((v) => !v)}>
          {showRegister ? 'Скрыть форму' : '+ Зарегистрировать багаж'}
        </button>
      </header>

      {error && <div className="alert error">{error}</div>}
      {success && <div className="alert success">{success}</div>}

      {showRegister && (
        <section className="card glass-effect animate-in">
          <h3>Регистрация багажа</h3>
          <form onSubmit={onRegister} className="profile-grid">
            <label className="field">
              <span>ID пассажира</span>
              <input
                required
                value={form.passenger_id}
                onChange={(e) => setForm({ ...form, passenger_id: e.target.value })}
                placeholder="UUID пассажира"
              />
            </label>
            <label className="field">
              <span>Рейс (необязательно)</span>
              <select value={form.flight_id} onChange={(e) => setForm({ ...form, flight_id: e.target.value })}>
                <option value="">— без рейса —</option>
                {flights.map((f) => (
                  <option key={f.id} value={f.id}>
                    {f.flight_number} · {airportShort(f.origin)} → {airportShort(f.destination)} · {formatTime(f.departure_time)}
                  </option>
                ))}
              </select>
            </label>
            <button type="submit" className="primary-btn full" disabled={submitting}>
              {submitting ? 'Регистрируем…' : 'Выдать бирку'}
            </button>
          </form>
        </section>
      )}

      <section className="baggage-grid">
        <div className="card glass-effect baggage-list-card">
          <div className="card-head">
            <h3>Активные бирки</h3>
            <small className="muted">{baggage.length}</small>
          </div>
          {baggage.length === 0 ? (
            <div className="empty-state">
              Пока нет багажа. Зарегистрируйте первый тег, чтобы увидеть движение по стадиям.
            </div>
          ) : (
            <ul className="baggage-list">
              {baggage.map((b) => {
                const idx = baggageStageIndex(b.status);
                const stage = BAGGAGE_STAGES[idx] || { icon: '📦', label: b.status };
                const tone = stageTone(idx);
                const f = b.flight_id ? flightByID[b.flight_id] : null;
                const isActive = selected && selected.id === b.id;
                return (
                  <li
                    key={b.id}
                    className={`baggage-item ${isActive ? 'active' : ''}`}
                    onClick={() => setSelectedId(b.id)}
                  >
                    <div className="baggage-item-icon">{stage.icon}</div>
                    <div className="baggage-item-body">
                      <div className="baggage-item-head">
                        <strong>#{b.id.slice(0, 8).toUpperCase()}</strong>
                        <span className={`tag tone-${tone}`}>{stage.label}</span>
                      </div>
                      <small className="muted">
                        {f ? `${f.flight_number} · ${airportShort(f.origin)} → ${airportShort(f.destination)}` : 'Без рейса'} · {formatRelative(b.updated_at)}
                      </small>
                    </div>
                  </li>
                );
              })}
            </ul>
          )}
        </div>

        <div className="card glass-effect baggage-detail-card">
          {!selected ? (
            <div className="empty-state">Выберите тег слева, чтобы увидеть полный маршрут.</div>
          ) : (
            <>
              <div className="card-head">
                <div>
                  <h3>Тег #{selected.id.slice(0, 8).toUpperCase()}</h3>
                  <small className="muted">Обновлено {formatRelative(selected.updated_at)}</small>
                </div>
                <button
                  className="primary-btn"
                  onClick={() => onScan(selected)}
                  disabled={selectedIdx >= BAGGAGE_STAGES.length - 1}
                >
                  {selectedIdx >= BAGGAGE_STAGES.length - 1 ? 'Маршрут завершён' : '↻ Следующий скан'}
                </button>
              </div>

              <div className="baggage-meta">
                <div>
                  <small>Пассажир</small>
                  <span>{selected.passenger_id.slice(0, 8)}…</span>
                </div>
                <div>
                  <small>Рейс</small>
                  <span>
                    {selected.flight_id
                      ? (flightByID[selected.flight_id]
                          ? `${flightByID[selected.flight_id].flight_number} · ${airportShort(flightByID[selected.flight_id].origin)} → ${airportShort(flightByID[selected.flight_id].destination)}`
                          : selected.flight_id.slice(0, 8) + '…')
                      : '—'}
                  </span>
                </div>
                <div>
                  <small>Текущее положение</small>
                  <span>{selected.location}</span>
                </div>
              </div>

              <div className="baggage-timeline">
                {BAGGAGE_STAGES.map((stage, i) => {
                  const state = i < selectedIdx ? 'done' : i === selectedIdx ? 'current' : 'pending';
                  return (
                    <div key={stage.key} className={`timeline-step ${state}`}>
                      <div className="timeline-dot">
                        <span>{stage.icon}</span>
                      </div>
                      <div className="timeline-text">
                        <strong>{stage.label}</strong>
                        <small>{stage.location}</small>
                      </div>
                      {i < BAGGAGE_STAGES.length - 1 && <div className="timeline-connector" />}
                    </div>
                  );
                })}
              </div>
            </>
          )}
        </div>
      </section>
    </>
  );
}
