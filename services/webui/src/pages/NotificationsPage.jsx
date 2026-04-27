import React, { useEffect, useRef, useState } from 'react';
import { api, formatTime, formatDate } from '../api';
import { isAdmin, useAuth } from '../auth';

const TYPE_META = {
  BOOKING_CONFIRMED: { icon: '🎟️', tone: 'success', label: 'Бронирование' },
  CHECKED_IN: { icon: '🪪', tone: 'success', label: 'Регистрация' },
  FLIGHT_DELAYED: { icon: '⏱️', tone: 'warning', label: 'Задержка' },
  FLIGHT_CANCELLED: { icon: '❌', tone: 'danger', label: 'Отмена' },
  GATE_CHANGED: { icon: '🚪', tone: 'info', label: 'Смена выхода' },
};

function meta(type) {
  return TYPE_META[type] || { icon: '🔔', tone: 'info', label: type };
}

export default function NotificationsPage() {
  const { user } = useAuth();
  const admin = isAdmin(user);
  const passengerId = user?.passenger_id;
  const [items, setItems] = useState([]);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const timerRef = useRef(null);

  const load = async () => {
    try {
      const data = admin
        ? await api.listNotifications()
        : passengerId
          ? await api.listNotificationsByPassenger(passengerId)
          : [];
      setItems(data || []);
      setError('');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
    timerRef.current = setInterval(load, 4000);
    return () => clearInterval(timerRef.current);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [admin, passengerId]);

  return (
    <>
      <header className="page-header">
        <div>
          <h1>Уведомления</h1>
          <p className="subtitle">
            {admin
              ? 'Все события по рейсам и бронированиям. Обновляется автоматически.'
              : 'События по вашим бронированиям и рейсам. Обновляется автоматически.'}
          </p>
        </div>
        <button className="ghost-btn" onClick={load} disabled={loading}>↻ Обновить</button>
      </header>

      {error && <div className="alert error">{error}</div>}

      <section className="notifications">
        {items.length === 0 && !loading && (
          <div className="empty-state">Пока нет уведомлений. Создайте инцидент на странице «Операции» или забронируйте рейс — событие появится здесь.</div>
        )}
        {items.map((item) => {
          const m = meta(item.type);
          return (
            <article key={item.id} className={`notification tone-${m.tone} animate-in`}>
              <div className="notification-icon">{m.icon}</div>
              <div className="notification-body">
                <div className="notification-meta">
                  <span className={`tag tone-${m.tone}`}>{m.label}</span>
                  <span className="muted">{item.channel}</span>
                  <span className="muted">{formatDate(item.sent_at)} · {formatTime(item.sent_at)}</span>
                </div>
                <h4>{item.title}</h4>
                <p>{item.content}</p>
                {item.flight_id && <small className="muted">Рейс {item.flight_id}</small>}
              </div>
            </article>
          );
        })}
      </section>
    </>
  );
}
