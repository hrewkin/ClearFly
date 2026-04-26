import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { airportShort, api, formatTime } from '../api';

const PRESET_REASONS = {
  FLIGHT_DELAYED: 'Метеоусловия в аэропорту прибытия',
  FLIGHT_CANCELLED: 'Технические причины',
  GATE_CHANGED: '',
};

export default function OperationsPage() {
  const [flights, setFlights] = useState([]);
  const [selected, setSelected] = useState('');
  const [type, setType] = useState('FLIGHT_DELAYED');
  const [reason, setReason] = useState(PRESET_REASONS.FLIGHT_DELAYED);
  const [newGate, setNewGate] = useState('A12');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  useEffect(() => {
    api.upcomingFlights()
      .then((data) => {
        setFlights(data || []);
        if ((data || []).length > 0) setSelected(data[0].flight_number);
      })
      .catch((err) => setError(err.message));
  }, []);

  useEffect(() => {
    setReason(PRESET_REASONS[type] ?? '');
  }, [type]);

  const submit = async (event) => {
    event.preventDefault();
    setLoading(true);
    setError('');
    setSuccess('');
    try {
      const payload = {
        type,
        flight_id: selected,
        reason: type === 'GATE_CHANGED' ? newGate : reason,
      };
      if (type === 'GATE_CHANGED') {
        payload.new_gate = newGate;
      }
      await api.createIncident(payload);
      setSuccess('Событие отправлено в шину уведомлений. Откройте «Уведомления», чтобы увидеть его в ленте.');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <header className="page-header">
        <div>
          <h1>Центр операций</h1>
          <p className="subtitle">Имитация действий диспетчера: задержка, отмена и смена выхода. Уведомления приходят пассажирам через шину RabbitMQ.</p>
        </div>
        <Link to="/notifications" className="ghost-btn">К уведомлениям →</Link>
      </header>

      <section className="ops-grid">
        <div className="card glass-effect">
          <h3>Создать инцидент</h3>
          <form onSubmit={submit}>
            <label className="field">
              <span>Рейс</span>
              <select value={selected} onChange={(e) => setSelected(e.target.value)}>
                {flights.map((f) => (
                  <option key={f.id} value={f.flight_number}>
                    {f.flight_number} — {airportShort(f.origin)} → {airportShort(f.destination)} · {formatTime(f.departure_time)}
                  </option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>Тип события</span>
              <select value={type} onChange={(e) => setType(e.target.value)}>
                <option value="FLIGHT_DELAYED">Задержка рейса</option>
                <option value="FLIGHT_CANCELLED">Отмена рейса</option>
                <option value="GATE_CHANGED">Смена выхода</option>
              </select>
            </label>
            {type === 'GATE_CHANGED' ? (
              <label className="field">
                <span>Новый выход</span>
                <input value={newGate} onChange={(e) => setNewGate(e.target.value)} required />
              </label>
            ) : (
              <label className="field">
                <span>Причина</span>
                <input value={reason} onChange={(e) => setReason(e.target.value)} required />
              </label>
            )}
            <button className="primary-btn full" type="submit" disabled={loading || !selected}>
              {loading ? 'Отправляем…' : 'Опубликовать событие'}
            </button>
            {error && <div className="alert error">{error}</div>}
            {success && <div className="alert success">{success}</div>}
          </form>
        </div>

        <div className="card glass-effect">
          <h3>Ближайшие рейсы</h3>
          <ul className="ops-flights">
            {flights.slice(0, 6).map((f) => (
              <li key={f.id}>
                <span className="flight-number">{f.flight_number}</span>
                <span>{airportShort(f.origin)} → {airportShort(f.destination)}</span>
                <span className="muted">{formatTime(f.departure_time)}</span>
                <span className={`tag tone-${(f.status || 'SCHEDULED').toLowerCase()}`}>{f.status || 'SCHEDULED'}</span>
              </li>
            ))}
            {flights.length === 0 && <li className="empty-state">Нет ближайших рейсов.</li>}
          </ul>
        </div>
      </section>
    </>
  );
}
