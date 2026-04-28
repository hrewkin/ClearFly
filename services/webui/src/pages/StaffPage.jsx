import React, { useEffect, useMemo, useState } from 'react';
import {
  airportShort, api, BAGGAGE_STAGES, baggageStageIndex,
  bookingStatusLabel, durationLabel, flightStatusLabel, formatDate,
  formatPrice, formatTime,
} from '../api';

const TIER_LABEL = {
  STANDARD: 'Standard',
  SILVER: 'Silver',
  GOLD: 'Gold',
  PLATINUM: 'Platinum',
};

const MEAL_LABEL = {
  STANDARD: 'Стандарт',
  VEGETARIAN: 'Вегетарианское',
  VEGAN: 'Веганское',
  KOSHER: 'Кошерное',
  HALAL: 'Халяль',
  GLUTEN_FREE: 'Без глютена',
  DIABETIC: 'Диабетическое',
};

export default function StaffPage() {
  const [flights, setFlights] = useState([]);
  const [selectedFlightId, setSelectedFlightId] = useState('');
  const [bookings, setBookings] = useState([]);
  const [passengers, setPassengers] = useState({});
  const [baggage, setBaggage] = useState([]);
  const [loadingFlights, setLoadingFlights] = useState(true);
  const [loadingManifest, setLoadingManifest] = useState(false);
  const [error, setError] = useState('');
  const [refundingId, setRefundingId] = useState('');
  const [statusMessage, setStatusMessage] = useState('');

  useEffect(() => {
    let mounted = true;
    api.upcomingFlights()
      .then((list) => {
        if (!mounted) return;
        setFlights(list || []);
        if ((list || []).length && !selectedFlightId) {
          setSelectedFlightId(list[0].id);
        }
      })
      .catch((err) => mounted && setError(err.message))
      .finally(() => mounted && setLoadingFlights(false));
    return () => { mounted = false; };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!selectedFlightId) return;
    let mounted = true;
    setLoadingManifest(true);
    setError('');
    setStatusMessage('');

    (async () => {
      try {
        const [bs, bg] = await Promise.all([
          api.listFlightBookings(selectedFlightId),
          api.listBaggage({ flight_id: selectedFlightId }).catch(() => []),
        ]);
        if (!mounted) return;
        setBookings(bs || []);
        setBaggage(bg || []);
        const ids = Array.from(new Set((bs || []).map((b) => b.passenger_id)));
        const fetched = await Promise.all(ids.map((id) => api.getPassenger(id).catch(() => null)));
        if (!mounted) return;
        const map = {};
        fetched.forEach((p) => { if (p) map[p.id] = p; });
        setPassengers(map);
      } catch (err) {
        if (mounted) setError(err.message);
      } finally {
        if (mounted) setLoadingManifest(false);
      }
    })();

    return () => { mounted = false; };
  }, [selectedFlightId]);

  const selectedFlight = useMemo(
    () => flights.find((f) => f.id === selectedFlightId),
    [flights, selectedFlightId],
  );

  const baggageByPassenger = useMemo(() => {
    const map = {};
    (baggage || []).forEach((b) => {
      if (!map[b.passenger_id]) map[b.passenger_id] = [];
      map[b.passenger_id].push(b);
    });
    return map;
  }, [baggage]);

  const refund = async (booking) => {
    if (booking.status === 'CANCELLED') return;
    const passengerName = passengers[booking.passenger_id]?.name || 'пассажиру';
    if (!window.confirm(`Оформить возврат по PNR ${booking.pnr_code} (${passengerName})?\nМесто будет освобождено, бронь отменится.`)) {
      return;
    }
    setRefundingId(booking.id);
    setError('');
    try {
      await api.staffRefund({ booking_id: booking.id, reason: 'Возврат оформлен сотрудником' });
      setBookings((prev) => prev.map((b) => b.id === booking.id
        ? { ...b, status: 'CANCELLED', payment_status: 'REFUNDED' }
        : b));
      setStatusMessage(`Возврат по PNR ${booking.pnr_code} оформлен. Уведомление пассажиру отправлено.`);
    } catch (err) {
      setError(err.message || 'Не удалось оформить возврат');
    } finally {
      setRefundingId('');
    }
  };

  const stageLabel = (status) => {
    const idx = baggageStageIndex(status);
    if (idx < 0) return status;
    return `${idx + 1}/${BAGGAGE_STAGES.length} · ${BAGGAGE_STAGES[idx].label}`;
  };

  return (
    <>
      <header className="page-header">
        <div>
          <h1>Рабочее место сотрудника</h1>
          <p className="subtitle">Манифест рейса, багаж пассажиров и оформление возврата.</p>
        </div>
      </header>

      {error && <div className="alert error">{error}</div>}
      {statusMessage && <div className="alert success">{statusMessage}</div>}

      <section className="card glass-effect" style={{ padding: '20px', marginBottom: '20px' }}>
        <label className="field" style={{ maxWidth: 420 }}>
          <span>Рейс</span>
          <select
            value={selectedFlightId}
            onChange={(e) => setSelectedFlightId(e.target.value)}
            disabled={loadingFlights || flights.length === 0}
          >
            {flights.map((f) => (
              <option key={f.id} value={f.id}>
                {f.flight_number} · {f.origin} → {f.destination} · {formatDate(f.departure_time)} {formatTime(f.departure_time)}
              </option>
            ))}
          </select>
        </label>

        {selectedFlight && (
          <div className="staff-flight-strip">
            <div>
              <small>Маршрут</small>
              <strong>
                {airportShort(selectedFlight.origin)} → {airportShort(selectedFlight.destination)}
              </strong>
              <div className="muted small">
                {formatDate(selectedFlight.departure_time)} {formatTime(selectedFlight.departure_time)}
                {' → '}
                {formatTime(selectedFlight.arrival_time)}
                {' · '}{durationLabel(selectedFlight.departure_time, selectedFlight.arrival_time)}
              </div>
            </div>
            <div>
              <small>Статус</small>
              <strong>{flightStatusLabel(selectedFlight.status)}</strong>
              <div className="muted small">Выход {selectedFlight.gate || '—'}</div>
            </div>
            <div>
              <small>Места</small>
              <strong>{selectedFlight.total_seats - selectedFlight.available_seats} / {selectedFlight.total_seats}</strong>
              <div className="muted small">{selectedFlight.aircraft_type}</div>
            </div>
          </div>
        )}
      </section>

      <section className="card glass-effect" style={{ padding: '20px' }}>
        <h2 style={{ marginTop: 0 }}>Манифест пассажиров</h2>
        {loadingManifest && <div className="empty-state">Загружаем данные…</div>}
        {!loadingManifest && bookings.length === 0 && (
          <div className="empty-state">На этот рейс пока нет бронирований.</div>
        )}
        {!loadingManifest && bookings.length > 0 && (
          <div className="manifest-table-wrap">
            <table className="manifest-table">
              <thead>
                <tr>
                  <th>PNR</th>
                  <th>Пассажир</th>
                  <th>Контакты</th>
                  <th>Тариф</th>
                  <th>Питание</th>
                  <th>Особые потребности</th>
                  <th>Багаж</th>
                  <th>Статус</th>
                  <th>Стоимость</th>
                  <th>Действие</th>
                </tr>
              </thead>
              <tbody>
                {bookings.map((b) => {
                  const p = passengers[b.passenger_id];
                  const bags = baggageByPassenger[b.passenger_id] || [];
                  return (
                    <tr key={b.id}>
                      <td><strong className="pnr">{b.pnr_code || '—'}</strong></td>
                      <td>
                        <div><strong>{p?.name || '…'}</strong></div>
                        <div className="muted small">Паспорт {p?.passport_number || '—'}</div>
                      </td>
                      <td>
                        <div>{p?.email || '—'}</div>
                        <div className="muted small">{p?.phone || '—'}</div>
                      </td>
                      <td>
                        <span className="tag">{TIER_LABEL[p?.loyalty_tier] || '—'}</span>
                        <div className="muted small">{p?.loyalty_points ?? 0} б.</div>
                      </td>
                      <td>{MEAL_LABEL[p?.meal_preference] || '—'}</td>
                      <td>{p?.special_needs ? <span className="tag tone-warning">{p.special_needs}</span> : <span className="muted small">—</span>}</td>
                      <td>
                        {bags.length === 0
                          ? <span className="muted small">нет</span>
                          : bags.map((bag) => (
                              <div key={bag.id} className="muted small">
                                {bag.id.slice(0, 8)} · {stageLabel(bag.status)}
                              </div>
                            ))}
                      </td>
                      <td><span className={`tag tone-${(b.status || '').toLowerCase()}`}>{bookingStatusLabel(b.status)}</span></td>
                      <td>{formatPrice(b.price, b.currency)}</td>
                      <td>
                        <button
                          className="ghost-btn"
                          disabled={b.status === 'CANCELLED' || refundingId === b.id}
                          onClick={() => refund(b)}
                        >
                          {refundingId === b.id ? '…' : (b.status === 'CANCELLED' ? 'Возвращено' : 'Возврат')}
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </>
  );
}
