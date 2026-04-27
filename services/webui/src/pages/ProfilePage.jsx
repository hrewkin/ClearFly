import React, { useEffect, useState } from 'react';
import { api } from '../api';
import { useAuth, isAdmin } from '../auth';

const MEAL_OPTIONS = [
  { value: 'STANDARD', label: 'Стандарт', icon: '🍽️' },
  { value: 'VEGETARIAN', label: 'Вегетарианское', icon: '🥗' },
  { value: 'VEGAN', label: 'Веганское', icon: '🌱' },
  { value: 'HALAL', label: 'Халяль', icon: '☪️' },
  { value: 'KOSHER', label: 'Кошер', icon: '✡️' },
  { value: 'GLUTEN_FREE', label: 'Без глютена', icon: '🌾' },
  { value: 'DIABETIC', label: 'Диабетическое', icon: '💉' },
];

const SPECIAL_NEEDS_OPTIONS = [
  { value: '', label: 'Нет' },
  { value: 'WHEELCHAIR', label: 'Инвалидная коляска' },
  { value: 'EXTRA_LEGROOM', label: 'Доп. место для ног' },
  { value: 'INFANT', label: 'С младенцем' },
  { value: 'UNACCOMPANIED_MINOR', label: 'Несопровождаемый ребёнок' },
  { value: 'VISUALLY_IMPAIRED', label: 'Слабовидящий' },
  { value: 'HEARING_IMPAIRED', label: 'Слабослышащий' },
];

const LOYALTY_TIERS = {
  STANDARD: { label: 'Standard', tone: 'info', icon: '✦' },
  SILVER: { label: 'Silver', tone: 'scheduled', icon: '🥈' },
  GOLD: { label: 'Gold', tone: 'warning', icon: '🥇' },
  PLATINUM: { label: 'Platinum', tone: 'success', icon: '💎' },
};

const DEMO_PROFILES = [
  { name: 'Анна Иванова', email: 'anna.ivanova@clearfly.ru', phone: '+7 903 111 22 33', passport_number: '4510 123456', tier: 'GOLD', meal: 'VEGETARIAN' },
  { name: 'Михаил Петров', email: 'mikhail.petrov@clearfly.ru', phone: '+7 915 222 33 44', passport_number: '4518 654321', tier: 'SILVER', meal: 'HALAL' },
  { name: 'Елена Смирнова', email: 'elena.smirnova@clearfly.ru', phone: '+7 926 333 44 55', passport_number: '4520 112233', tier: 'PLATINUM', meal: 'STANDARD' },
];

function loyaltyMeta(tier) {
  return LOYALTY_TIERS[tier] || LOYALTY_TIERS.STANDARD;
}

export default function ProfilePage() {
  const { user } = useAuth();
  const admin = isAdmin(user);
  const [searchId, setSearchId] = useState('');
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({ name: '', email: '', phone: '', passport_number: '' });
  const [createdRecently, setCreatedRecently] = useState([]);

  useEffect(() => {
    if (!success) return;
    const t = setTimeout(() => setSuccess(''), 3000);
    return () => clearTimeout(t);
  }, [success]);

  const loadProfile = async (id) => {
    setLoading(true);
    setError('');
    try {
      const data = await api.getPassenger(id);
      setProfile(data);
      setSearchId(id);
    } catch (err) {
      setError(err.message);
      setProfile(null);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (user?.passenger_id && !profile) {
      loadProfile(user.passenger_id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user?.passenger_id]);

  const onSearch = (e) => {
    e.preventDefault();
    if (!searchId) return;
    loadProfile(searchId.trim());
  };

  const onCreate = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const created = await api.createPassenger(form);
      setProfile(created);
      setSearchId(created.id);
      setCreatedRecently((prev) => [created, ...prev].slice(0, 5));
      setShowCreate(false);
      setSuccess(`Пассажир зарегистрирован. ID: ${created.id.slice(0, 8)}…`);
      setForm({ name: '', email: '', phone: '', passport_number: '' });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const onQuickFill = (p) => {
    setForm({ name: p.name, email: p.email, phone: p.phone, passport_number: p.passport_number });
    setShowCreate(true);
  };

  const onMealChange = async (value) => {
    if (!profile) return;
    setLoading(true);
    try {
      const updated = await api.updatePreferences(profile.id, {
        meal_preference: value,
        special_needs: profile.special_needs || '',
      });
      setProfile(updated);
      setSuccess('Предпочтение по питанию обновлено');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const onSpecialNeedsChange = async (value) => {
    if (!profile) return;
    setLoading(true);
    try {
      const updated = await api.updatePreferences(profile.id, { special_needs: value });
      setProfile(updated);
      setSuccess('Особые потребности обновлены');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const onTierBump = async (newTier) => {
    if (!profile) return;
    setLoading(true);
    try {
      const updated = await api.updatePreferences(profile.id, {
        loyalty_tier: newTier,
        special_needs: profile.special_needs || '',
      });
      setProfile(updated);
      setSuccess(`Статус лояльности: ${loyaltyMeta(newTier).label}`);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const tier = profile ? loyaltyMeta(profile.loyalty_tier) : null;

  return (
    <>
      <header className="page-header">
        <div>
          <h1>Профиль пассажира</h1>
          <p className="subtitle">Лояльность, предпочтения по питанию и особые потребности — всё в одном месте.</p>
        </div>
        {admin && (
          <button className="ghost-btn" onClick={() => setShowCreate((v) => !v)}>
            {showCreate ? 'Скрыть форму' : '+ Новый пассажир'}
          </button>
        )}
      </header>

      {error && <div className="alert error">{error}</div>}
      {success && <div className="alert success">{success}</div>}

      {admin && (
      <section className="card glass-effect profile-search-card">
        <form onSubmit={onSearch} className="profile-search-form">
          <label className="field">
            <span>ID пассажира</span>
            <input
              type="text"
              placeholder="UUID, напр. 5e5c28d4-…"
              value={searchId}
              onChange={(e) => setSearchId(e.target.value)}
            />
          </label>
          <button type="submit" className="primary-btn" disabled={loading || !searchId}>
            {loading ? 'Загружаем…' : 'Открыть профиль'}
          </button>
        </form>
        {createdRecently.length > 0 && (
          <div className="recent-chips">
            <span className="muted">Созданные в этой сессии:</span>
            {createdRecently.map((p) => (
              <button key={p.id} className="chip" onClick={() => loadProfile(p.id)}>
                {p.name} · {p.id.slice(0, 8)}
              </button>
            ))}
          </div>
        )}
      </section>
      )}

      {admin && showCreate && (
        <section className="card glass-effect animate-in">
          <h3>Регистрация пассажира</h3>
          <div className="quick-fill">
            <span className="muted">Быстрый ввод:</span>
            {DEMO_PROFILES.map((p) => (
              <button key={p.email} className="chip" onClick={() => onQuickFill(p)}>
                {p.name.split(' ')[0]}
              </button>
            ))}
          </div>
          <form onSubmit={onCreate} className="profile-grid">
            <label className="field">
              <span>ФИО</span>
              <input required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="Иванов Иван" />
            </label>
            <label className="field">
              <span>Email</span>
              <input required type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} placeholder="ivan@example.com" />
            </label>
            <label className="field">
              <span>Телефон</span>
              <input required value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} placeholder="+7 900 000 00 00" />
            </label>
            <label className="field">
              <span>Номер паспорта</span>
              <input required value={form.passport_number} onChange={(e) => setForm({ ...form, passport_number: e.target.value })} placeholder="4510 123456" />
            </label>
            <button className="primary-btn full" type="submit" disabled={loading}>
              {loading ? 'Создаём…' : 'Зарегистрировать пассажира'}
            </button>
          </form>
        </section>
      )}

      {profile && (
        <section className="profile-card animate-in">
          <div className="profile-header">
            <div className="profile-avatar">{(profile.name || '?').slice(0, 1).toUpperCase()}</div>
            <div className="profile-title">
              <div className="profile-name-row">
                <h2>{profile.name}</h2>
                <span className={`tag tone-${tier.tone} loyalty-tag`}>
                  <span className="loyalty-icon">{tier.icon}</span>{tier.label}
                  <small>{profile.loyalty_points ?? 0} б</small>
                </span>
              </div>
              <small className="muted">ID: {profile.id}</small>
            </div>
          </div>

          <div className="profile-info-grid">
            <div className="info-cell">
              <small>Email</small>
              <span>{profile.email}</span>
            </div>
            <div className="info-cell">
              <small>Телефон</small>
              <span>{profile.phone}</span>
            </div>
            <div className="info-cell">
              <small>Паспорт</small>
              <span>{profile.passport_number}</span>
            </div>
          </div>

          <div className="profile-prefs">
            <div className="pref-group">
              <small>Предпочтение по питанию</small>
              <div className="pref-options">
                {MEAL_OPTIONS.map((m) => (
                  <button
                    key={m.value}
                    className={`pref-option ${profile.meal_preference === m.value ? 'selected' : ''}`}
                    onClick={() => onMealChange(m.value)}
                    disabled={loading}
                    type="button"
                  >
                    <span className="pref-icon">{m.icon}</span>
                    <span>{m.label}</span>
                  </button>
                ))}
              </div>
            </div>

            <div className="pref-group">
              <small>Особые потребности</small>
              <select
                value={profile.special_needs || ''}
                onChange={(e) => onSpecialNeedsChange(e.target.value)}
                disabled={loading}
                className="pref-select"
              >
                {SPECIAL_NEEDS_OPTIONS.map((o) => (
                  <option key={o.value || 'none'} value={o.value}>{o.label}</option>
                ))}
              </select>
            </div>

            <div className="pref-group">
              <small>Статус лояльности</small>
              {admin ? (
                <div className="pref-options tier-options">
                  {Object.entries(LOYALTY_TIERS).map(([key, meta]) => (
                    <button
                      key={key}
                      className={`pref-option ${profile.loyalty_tier === key ? 'selected' : ''}`}
                      onClick={() => onTierBump(key)}
                      disabled={loading}
                      type="button"
                    >
                      <span className="pref-icon">{meta.icon}</span>
                      <span>{meta.label}</span>
                    </button>
                  ))}
                </div>
              ) : (
                <div className="loyalty-readonly">
                  <span className={`tag tone-${tier.tone} loyalty-tag`}>
                    <span className="loyalty-icon">{tier.icon}</span>{tier.label}
                    <small>{profile.loyalty_points ?? 0} б</small>
                  </span>
                </div>
              )}
            </div>
          </div>
        </section>
      )}

      {!profile && !loading && !error && (
        <div className="empty-state">
          {admin
            ? 'Введите ID пассажира или создайте нового. Данные профиля используются во время бронирования и в посадочных ведомостях экипажа.'
            : 'У вас пока не заполнен профиль пассажира. Забронируйте рейс — и данные появятся здесь.'}
        </div>
      )}
    </>
  );
}
