import React, { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { airportShort, api, durationLabel, formatPrice, formatTime } from '../api';
import { useAuth } from '../auth';

const ECONOMY_LAYOUT = ['A', 'B', 'C', 'D', 'E', 'F'];
const BUSINESS_LAYOUT = ['A', 'B', 'C', 'D'];
const ECONOMY_AISLE_AFTER = 'C';
const BUSINESS_AISLE_AFTER = 'B';

function groupSeatsByRow(seats) {
  const map = new Map();
  seats.forEach((seat) => {
    const row = seat.seat_number.replace(/\D+$/, '');
    if (!map.has(row)) map.set(row, []);
    map.get(row).push(seat);
  });
  return Array.from(map.entries())
    .map(([row, list]) => ({
      row,
      seats: list.sort((a, b) => a.seat_number.localeCompare(b.seat_number)),
      class: list[0]?.class ?? 'ECONOMY',
    }))
    .sort((a, b) => Number(a.row) - Number(b.row));
}

function classLayout(rowClass) {
  if (rowClass === 'BUSINESS') {
    return { columns: BUSINESS_LAYOUT, aisleAfter: BUSINESS_AISLE_AFTER };
  }
  return { columns: ECONOMY_LAYOUT, aisleAfter: ECONOMY_AISLE_AFTER };
}

export default function BookingFlowPage() {
  const { flightId } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [flight, setFlight] = useState(null);
  const [seats, setSeats] = useState([]);
  const [tariffs, setTariffs] = useState([]);
  const [selectedSeat, setSelectedSeat] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [confirmation, setConfirmation] = useState(null);

  const [passengerName, setPassengerName] = useState(user?.full_name || '');
  const [email, setEmail] = useState(user?.email || '');
  const [phone, setPhone] = useState('');
  const [passport, setPassport] = useState('');
  const [fieldErrors, setFieldErrors] = useState({});

  useEffect(() => {
    if (user?.full_name && !passengerName) setPassengerName(user.full_name);
    if (user?.email && !email) setEmail(user.email);
  }, [user]); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    let mounted = true;
    setLoading(true);
    Promise.all([
      api.getFlight(flightId),
      api.getSeats(flightId),
      api.getTariffs(flightId).catch(() => []),
    ])
      .then(([flightData, seatData, tariffData]) => {
        if (!mounted) return;
        setFlight(flightData);
        setSeats(seatData || []);
        setTariffs(tariffData || []);
      })
      .catch((err) => mounted && setError(err.message))
      .finally(() => mounted && setLoading(false));
    return () => { mounted = false; };
  }, [flightId]);

  const rows = useMemo(() => groupSeatsByRow(seats), [seats]);
  const tariffByClass = useMemo(() => {
    const map = {};
    tariffs.forEach((t) => { map[t.class] = t; });
    return map;
  }, [tariffs]);

  const selectedTariff = selectedSeat ? tariffByClass[selectedSeat.class] : null;

  const validate = () => {
    const errs = {};
    const name = passengerName.trim();
    const words = name.split(/\s+/).filter(Boolean);
    if (!name) errs.name = 'Укажите ФИО';
    else if (words.length < 2) errs.name = 'Введите имя и фамилию (минимум)';
    else if (!/^[А-Яа-яЁёA-Za-z\-\s]+$/.test(name)) errs.name = 'Только буквы, дефис и пробел';
    else if (name.length < 4) errs.name = 'Слишком короткое ФИО';

    const mail = email.trim();
    if (!mail) errs.email = 'Укажите email';
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(mail)) errs.email = 'Некорректный email';

    const phoneDigits = phone.replace(/\D/g, '');
    if (!phone.trim()) errs.phone = 'Укажите телефон';
    else if (phoneDigits.length < 10 || phoneDigits.length > 15) errs.phone = 'Телефон должен содержать 10–15 цифр';

    const pass = passport.replace(/\s/g, '');
    if (!pass) errs.passport = 'Укажите паспорт';
    else if (!/^\d{10}$/.test(pass)) errs.passport = 'Паспорт РФ: серия 4 цифры + номер 6 цифр';

    return errs;
  };

  const onConfirm = async () => {
    if (!selectedSeat || !flight) return;
    const errs = validate();
    setFieldErrors(errs);
    if (Object.keys(errs).length > 0) {
      setError('Проверьте корректность полей пассажира.');
      return;
    }
    setSubmitting(true);
    setError('');
    try {
      let passenger;
      if (user?.passenger_id) {
        const existing = await api.getPassenger(user.passenger_id);
        passenger = await api.updatePassenger(user.passenger_id, {
          ...existing,
          name: passengerName,
          email,
          phone,
          passport_number: passport,
        });
      } else {
        passenger = await api.createPassenger({
          name: passengerName,
          email,
          phone,
          passport_number: passport,
        });
      }
      const booking = await api.bookSeat({
        flight_id: flight.id,
        passenger_id: passenger.id,
        seat_id: selectedSeat.id,
      });
      setConfirmation({ booking, passenger, seat: selectedSeat });
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return <div className="loading">Загружаем рейс…</div>;
  }
  if (error && !flight) {
    return <div className="alert error">Не удалось загрузить рейс: {error}</div>;
  }
  if (!flight) {
    return <div className="alert error">Рейс не найден.</div>;
  }

  if (confirmation) {
    const { booking, passenger, seat } = confirmation;
    return (
      <>
        <header className="page-header">
          <div>
            <h1>Бронирование подтверждено</h1>
            <p className="subtitle">Поздравляем! Сохраните PNR — он понадобится при регистрации.</p>
          </div>
        </header>
        <section className="boarding-pass animate-in">
          <div className="boarding-pass-stub">
            <div className="brand-line">
              <span className="logo-mini">✈</span>
              ClearFly
            </div>
            <div className="pnr">
              <small>PNR</small>
              <strong>{booking.pnr_code}</strong>
            </div>
            <div className="passenger-row">
              <div>
                <small>Пассажир</small>
                <strong>{passenger.name}</strong>
              </div>
              <div>
                <small>Класс</small>
                <strong>{seat.class === 'BUSINESS' ? 'Бизнес' : 'Эконом'}</strong>
              </div>
            </div>
          </div>
          <div className="boarding-pass-main">
            <div className="bp-route">
              <div className="bp-side">
                <small>{airportShort(flight.origin)}</small>
                <strong>{flight.origin}</strong>
                <span>{formatTime(flight.departure_time)}</span>
              </div>
              <div className="bp-arrow">→</div>
              <div className="bp-side">
                <small>{airportShort(flight.destination)}</small>
                <strong>{flight.destination}</strong>
                <span>{formatTime(flight.arrival_time)}</span>
              </div>
            </div>
            <div className="bp-meta">
              <div><small>Рейс</small><strong>{flight.flight_number}</strong></div>
              <div><small>Выход</small><strong>{flight.gate || '—'}</strong></div>
              <div><small>Место</small><strong>{seat.seat_number}</strong></div>
              <div><small>Цена</small><strong>{formatPrice(booking.price, booking.currency)}</strong></div>
            </div>
          </div>
        </section>
        <div className="actions-row">
          <button className="ghost-btn" onClick={() => navigate('/notifications')}>Открыть уведомления</button>
          <button className="primary-btn" onClick={() => navigate('/search')}>Забронировать ещё рейс</button>
        </div>
      </>
    );
  }

  return (
    <>
      <header className="page-header">
        <div>
          <h1>Бронирование рейса {flight.flight_number}</h1>
          <p className="subtitle">
            {airportShort(flight.origin)} → {airportShort(flight.destination)} · {formatTime(flight.departure_time)} · {durationLabel(flight.departure_time, flight.arrival_time)} в пути
          </p>
        </div>
        <button className="ghost-btn" onClick={() => navigate('/search')}>← К поиску</button>
      </header>

      <section className="booking-grid">
        <div className="card glass-effect">
          <h3>Выбор места</h3>
          <p className="muted">Кликните по доступному месту, чтобы выбрать. Зелёные — свободно, синие — выбрано, серые — заняты.</p>
          <div className="seat-map">
            <div className="seat-map-cabin">Бизнес-класс</div>
            {rows.filter((row) => row.class === 'BUSINESS').map((row) => (
              <SeatRow key={`bus-${row.row}`} row={row} selected={selectedSeat} onSelect={setSelectedSeat} />
            ))}
            <div className="seat-map-cabin">Эконом-класс</div>
            {rows.filter((row) => row.class !== 'BUSINESS').map((row) => (
              <SeatRow key={`eco-${row.row}`} row={row} selected={selectedSeat} onSelect={setSelectedSeat} />
            ))}
          </div>
          <div className="seat-legend">
            <span><i className="seat available" /> Свободно</span>
            <span><i className="seat selected" /> Выбрано</span>
            <span><i className="seat blocked" /> Занято</span>
            <span><i className="seat business" /> Бизнес</span>
          </div>
        </div>

        <aside className="card glass-effect">
          <h3>Данные пассажира</h3>
          <label className={`field${fieldErrors.name ? ' field-error' : ''}`}>
            <span>ФИО</span>
            <input value={passengerName} onChange={(e) => { setPassengerName(e.target.value); if (fieldErrors.name) setFieldErrors({ ...fieldErrors, name: undefined }); }} placeholder="Иван Петров" />
            {fieldErrors.name && <small className="field-error-msg">{fieldErrors.name}</small>}
          </label>
          <label className={`field${fieldErrors.email ? ' field-error' : ''}`}>
            <span>Email</span>
            <input type="email" value={email} onChange={(e) => { setEmail(e.target.value); if (fieldErrors.email) setFieldErrors({ ...fieldErrors, email: undefined }); }} placeholder="name@example.com" />
            {fieldErrors.email && <small className="field-error-msg">{fieldErrors.email}</small>}
          </label>
          <label className={`field${fieldErrors.phone ? ' field-error' : ''}`}>
            <span>Телефон</span>
            <input value={phone} onChange={(e) => { setPhone(e.target.value); if (fieldErrors.phone) setFieldErrors({ ...fieldErrors, phone: undefined }); }} placeholder="+7 999 123-45-67" />
            {fieldErrors.phone && <small className="field-error-msg">{fieldErrors.phone}</small>}
          </label>
          <label className={`field${fieldErrors.passport ? ' field-error' : ''}`}>
            <span>Паспорт РФ (серия и номер)</span>
            <input value={passport} onChange={(e) => { setPassport(e.target.value); if (fieldErrors.passport) setFieldErrors({ ...fieldErrors, passport: undefined }); }} placeholder="45 12 678901" />
            {fieldErrors.passport && <small className="field-error-msg">{fieldErrors.passport}</small>}
          </label>

          <div className="summary">
            <div className="summary-row">
              <span>Место</span>
              <strong>{selectedSeat ? selectedSeat.seat_number : '—'}</strong>
            </div>
            <div className="summary-row">
              <span>Класс</span>
              <strong>{selectedSeat ? (selectedSeat.class === 'BUSINESS' ? 'Бизнес' : 'Эконом') : '—'}</strong>
            </div>
            <div className="summary-row total">
              <span>К оплате</span>
              <strong>{selectedTariff ? formatPrice(selectedTariff.base_price, selectedTariff.currency) : '—'}</strong>
            </div>
          </div>

          {error && <div className="alert error">{error}</div>}

          <button className="primary-btn full" onClick={onConfirm} disabled={!selectedSeat || submitting}>
            {submitting ? 'Бронируем…' : 'Подтвердить и оплатить'}
          </button>
          <p className="muted small">Демо-оплата: средства не списываются.</p>
        </aside>
      </section>
    </>
  );
}

function SeatRow({ row, selected, onSelect }) {
  const layout = classLayout(row.class);
  const seatByCol = {};
  row.seats.forEach((s) => {
    const col = s.seat_number.slice(-1);
    seatByCol[col] = s;
  });

  return (
    <div className={`seat-row ${row.class === 'BUSINESS' ? 'row-business' : ''}`}>
      <div className="row-label">{row.row}</div>
      {layout.columns.map((col, idx) => {
        const seat = seatByCol[col];
        const aisle = col === layout.aisleAfter && idx !== layout.columns.length - 1;
        return (
          <React.Fragment key={col}>
            {seat ? (
              <button
                type="button"
                disabled={seat.status !== 'AVAILABLE'}
                className={`seat ${row.class === 'BUSINESS' ? 'business' : ''}
                  ${seat.status === 'AVAILABLE' ? 'available' : 'blocked'}
                  ${selected?.id === seat.id ? 'selected' : ''}`}
                onClick={() => onSelect(seat)}
                title={`${seat.seat_number} · ${seat.class === 'BUSINESS' ? 'Бизнес' : 'Эконом'}`}
              >
                {col}
              </button>
            ) : (
              <span className="seat empty" />
            )}
            {aisle && <span className="aisle" aria-hidden />}
          </React.Fragment>
        );
      })}
    </div>
  );
}
