import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { airportShort, api, bookingStatusLabel, durationLabel, formatDate, formatPrice, formatTime } from '../api';
import { useAuth } from '../auth';

export default function MyBookingsPage() {
  const { user } = useAuth();
  const [items, setItems] = useState([]);
  const [flightsById, setFlightsById] = useState({});
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [cancellingId, setCancellingId] = useState('');
  const [info, setInfo] = useState('');

  useEffect(() => {
    let mounted = true;
    if (!user?.passenger_id) {
      setLoading(false);
      return;
    }
    (async () => {
      try {
        const bookings = await api.listBookingsByPassenger(user.passenger_id);
        if (!mounted) return;
        setItems(bookings || []);
        const flightIds = Array.from(new Set((bookings || []).map((b) => b.flight_id)));
        const fetched = await Promise.all(
          flightIds.map((id) => api.getFlight(id).catch(() => null))
        );
        if (!mounted) return;
        const map = {};
        fetched.forEach((f) => { if (f) map[f.id] = f; });
        setFlightsById(map);
      } catch (err) {
        if (mounted) setError(err.message);
      } finally {
        if (mounted) setLoading(false);
      }
    })();
    return () => { mounted = false; };
  }, [user?.passenger_id]);

  if (!user?.passenger_id) {
    return (
      <>
        <header className="page-header">
          <div>
            <h1>Мои брони</h1>
            <p className="subtitle">Страница доступна только пассажирам.</p>
          </div>
        </header>
      </>
    );
  }

  return (
    <>
      <header className="page-header">
        <div>
          <h1>Мои брони</h1>
          <p className="subtitle">Все ваши забронированные рейсы и PNR-коды.</p>
        </div>
        <Link to="/search" className="primary-btn">Новое бронирование →</Link>
      </header>

      {error && <div className="alert error">{error}</div>}
      {info && <div className="alert success">{info}</div>}

      {loading && <div className="empty-state">Загружаем ваши брони…</div>}

      {!loading && items.length === 0 && (
        <div className="empty-state">
          У вас пока нет бронирований. Выберите рейс на <Link to="/search" className="link">странице поиска</Link>.
        </div>
      )}

      <section className="my-bookings">
        {items.map((b) => {
          const f = flightsById[b.flight_id];
          const departure = f ? new Date(f.departure_time) : null;
          const hoursUntil = departure ? (departure - new Date()) / 36e5 : Infinity;
          const cancelEligible = b.status !== 'CANCELLED' && hoursUntil >= 24;
          const cancel = async () => {
            if (!window.confirm(`Отменить бронь по PNR ${b.pnr_code}? Возврат будет оформлен автоматически.`)) return;
            setCancellingId(b.id);
            setError('');
            setInfo('');
            try {
              await api.cancelOwnBooking(b.id);
              setItems((prev) => prev.map((x) => x.id === b.id
                ? { ...x, status: 'CANCELLED', payment_status: 'REFUNDED' }
                : x));
              setInfo(`Бронь ${b.pnr_code} отменена. Уведомление о возврате отправлено.`);
            } catch (err) {
              setError(err.message || 'Не удалось отменить бронь');
            } finally {
              setCancellingId('');
            }
          };
          return (
            <article key={b.id} className="card glass-effect my-booking animate-in">
              <div className="mb-head">
                <div>
                  <small>PNR</small>
                  <strong className="pnr">{b.pnr_code}</strong>
                </div>
                <span className={`tag tone-${(b.status || '').toLowerCase()}`}>{bookingStatusLabel(b.status)}</span>
                <div className="price">{formatPrice(b.price, b.currency)}</div>
              </div>
              {f ? (
                <div className="mb-route">
                  <div className="bp-side">
                    <small>{airportShort(f.origin)}</small>
                    <strong>{f.origin}</strong>
                    <span>{formatTime(f.departure_time)} · {formatDate(f.departure_time)}</span>
                  </div>
                  <div className="bp-arrow">→ {durationLabel(f.departure_time, f.arrival_time)} →</div>
                  <div className="bp-side">
                    <small>{airportShort(f.destination)}</small>
                    <strong>{f.destination}</strong>
                    <span>{formatTime(f.arrival_time)} · {formatDate(f.arrival_time)}</span>
                  </div>
                </div>
              ) : (
                <div className="muted">Рейс {b.flight_id}</div>
              )}
              <div className="mb-meta">
                {f && <span>Рейс <strong>{f.flight_number}</strong></span>}
                {f?.gate && <span>Выход <strong>{f.gate}</strong></span>}
                <span>Создано {formatDate(b.created_at)} {formatTime(b.created_at)}</span>
              </div>
              <div className="mb-actions">
                {b.status === 'CANCELLED' ? (
                  <span className="muted small">Бронь отменена. Возврат оформлен.</span>
                ) : cancelEligible ? (
                  <button
                    className="ghost-btn"
                    onClick={cancel}
                    disabled={cancellingId === b.id}
                    title="Бронь можно отменить не позднее, чем за 24 часа до вылета"
                  >
                    {cancellingId === b.id ? '…' : 'Отменить бронь и оформить возврат'}
                  </button>
                ) : (
                  <span className="muted small">До вылета меньше 24 часов — отмена через сотрудника аэропорта.</span>
                )}
              </div>
            </article>
          );
        })}
      </section>
    </>
  );
}
