import React, { useState } from 'react';

function ProfilePage() {
  const [passengerId, setPassengerId] = useState('');
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [isCreating, setIsCreating] = useState(false);

  const [formData, setFormData] = useState({
    name: '',
    email: '',
    phone: '',
    passport_number: ''
  });

  const fetchProfile = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setProfile(null);

    try {
      const res = await fetch(`http://localhost:8080/api/v1/passengers/${passengerId}`);
      if (!res.ok) {
        throw new Error('Пассажир не найден');
      }
      const data = await res.json();
      setProfile(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const createProfile = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const res = await fetch('http://localhost:8080/api/v1/passengers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(formData)
      });
      if (!res.ok) {
        throw new Error('Не удалось создать профиль');
      }
      const data = await res.json();
      setProfile(data);
      setIsCreating(false);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <header>
        <h1>Профиль пассажира</h1>
        <p className="subtitle">Управление данными вашего профиля</p>
      </header>

      <section className="dashboard-cards">
        <div className="card glass-effect animate-in">
          <h3>Найти профиль</h3>
          <p>Введите ID пассажира (UUID) для просмотра профиля.</p>
          <form onSubmit={fetchProfile} className="search-form">
            <input
              type="text"
              placeholder="e.g. 123e4567-e89b-12d3-..."
              value={passengerId}
              onChange={(e) => setPassengerId(e.target.value)}
              required
            />
            <button type="submit" disabled={loading}>
              {loading ? 'Поиск...' : 'Найти'}
            </button>
          </form>
        </div>

        <div className="card glass-effect highlight animate-in" style={{ animationDelay: '0.1s' }}>
          <h3>Новый профиль</h3>
          <p>Создайте новый профиль пассажира в системе.</p>
          <button className="primary-btn" onClick={() => setIsCreating(!isCreating)}>
            {isCreating ? 'Отмена' : 'Создать профиль'}
          </button>
        </div>
      </section>

      {isCreating && (
        <section className="results-section animate-in">
          <h2>Регистрация пассажира</h2>
          <form onSubmit={createProfile} className="profile-form">
            <div className="form-group">
              <label>ФИО</label>
              <input
                type="text"
                placeholder="Иванов Иван Иванович"
                value={formData.name}
                onChange={(e) => setFormData({...formData, name: e.target.value})}
                required
              />
            </div>
            <div className="form-group">
              <label>Email</label>
              <input
                type="email"
                placeholder="ivan@example.com"
                value={formData.email}
                onChange={(e) => setFormData({...formData, email: e.target.value})}
                required
              />
            </div>
            <div className="form-group">
              <label>Телефон</label>
              <input
                type="tel"
                placeholder="+7 900 123 45 67"
                value={formData.phone}
                onChange={(e) => setFormData({...formData, phone: e.target.value})}
                required
              />
            </div>
            <div className="form-group">
              <label>Номер паспорта</label>
              <input
                type="text"
                placeholder="AB1234567"
                value={formData.passport_number}
                onChange={(e) => setFormData({...formData, passport_number: e.target.value})}
                required
              />
            </div>
            <button type="submit" className="primary-btn" disabled={loading}>
              {loading ? 'Создание...' : 'Зарегистрировать'}
            </button>
          </form>
        </section>
      )}

      {error && (
        <div className="alert error animate-in">
          <strong>Ошибка:</strong> {error}
        </div>
      )}

      {profile && (
        <section className="results-section animate-in">
          <h2>Данные профиля</h2>
          <div className="ticket">
            <div className="ticket-header">
              <span className="status">ACTIVE</span>
              <span className="booking-id">ID: {profile.id}</span>
            </div>
            <div className="ticket-body">
              <div className="info-group">
                <label>ФИО</label>
                <span>{profile.name}</span>
              </div>
              <div className="info-group">
                <label>Email</label>
                <span>{profile.email}</span>
              </div>
              <div className="info-group">
                <label>Телефон</label>
                <span>{profile.phone}</span>
              </div>
              <div className="info-group">
                <label>Паспорт</label>
                <span>{profile.passport_number}</span>
              </div>
            </div>
          </div>
        </section>
      )}
    </>
  );
}

export default ProfilePage;
