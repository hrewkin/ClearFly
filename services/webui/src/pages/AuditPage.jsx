import React, { useEffect, useMemo, useState } from 'react';
import { api, formatDate, formatTime } from '../api';

const ACTION_LABEL = {
  BOOKING_REFUND: 'Возврат брони (сотрудник)',
  BOOKING_SELF_CANCEL: 'Отмена брони пассажиром',
};

const ROLE_LABEL = {
  admin: 'Админ',
  staff: 'Сотрудник',
  passenger: 'Пассажир',
};

export default function AuditPage() {
  const [entries, setEntries] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [actorFilter, setActorFilter] = useState('');
  const [actionFilter, setActionFilter] = useState('');

  useEffect(() => {
    let mounted = true;
    api.staffAudit()
      .then((list) => { if (mounted) setEntries(list || []); })
      .catch((err) => { if (mounted) setError(err.message); })
      .finally(() => { if (mounted) setLoading(false); });
    return () => { mounted = false; };
  }, []);

  const filtered = useMemo(() => {
    return (entries || []).filter((e) => {
      if (actorFilter && !e.actor_name.toLowerCase().includes(actorFilter.toLowerCase())) return false;
      if (actionFilter && e.action !== actionFilter) return false;
      return true;
    });
  }, [entries, actorFilter, actionFilter]);

  const actions = useMemo(() => Array.from(new Set((entries || []).map((e) => e.action))), [entries]);

  return (
    <>
      <header className="page-header">
        <div>
          <h1>Журнал аудита</h1>
          <p className="subtitle">Действия сотрудников и пассажиров над бронями и профилями.</p>
        </div>
      </header>

      {error && <div className="alert error">{error}</div>}

      <section className="card glass-effect" style={{ padding: '20px', marginBottom: 20 }}>
        <div className="audit-filters">
          <label className="field">
            <span>Поиск по актёру</span>
            <input value={actorFilter} onChange={(e) => setActorFilter(e.target.value)} placeholder="ФИО или часть имени" />
          </label>
          <label className="field">
            <span>Действие</span>
            <select value={actionFilter} onChange={(e) => setActionFilter(e.target.value)}>
              <option value="">Все</option>
              {actions.map((a) => <option key={a} value={a}>{ACTION_LABEL[a] || a}</option>)}
            </select>
          </label>
          <div className="audit-counter muted small">Всего: {filtered.length}</div>
        </div>
      </section>

      <section className="card glass-effect" style={{ padding: '0', overflow: 'hidden' }}>
        {loading && <div className="empty-state">Загружаем журнал…</div>}
        {!loading && filtered.length === 0 && (
          <div className="empty-state">Записей пока нет.</div>
        )}
        {!loading && filtered.length > 0 && (
          <div className="manifest-table-wrap">
            <table className="manifest-table">
              <thead>
                <tr>
                  <th>Когда</th>
                  <th>Кто</th>
                  <th>Роль</th>
                  <th>Действие</th>
                  <th>Объект</th>
                  <th>Детали</th>
                  <th>IP</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((e) => (
                  <tr key={e.id}>
                    <td>
                      <div>{formatDate(e.created_at)} {formatTime(e.created_at)}</div>
                    </td>
                    <td><strong>{e.actor_name}</strong></td>
                    <td><span className="tag">{ROLE_LABEL[e.actor_role] || e.actor_role}</span></td>
                    <td>{ACTION_LABEL[e.action] || e.action}</td>
                    <td>
                      {e.target_type ? <div className="muted small">{e.target_type}</div> : null}
                      {e.target_id ? <div><code>{e.target_id.slice(0, 8)}</code></div> : null}
                    </td>
                    <td className="muted small">{e.details}</td>
                    <td className="muted small">{e.ip_address || '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </>
  );
}
