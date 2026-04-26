import React, { useState } from 'react';

function BaggagePage() {
  const [baggageId, setBaggageId] = useState('');
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchBaggage = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setResult(null);

    try {
      const res = await fetch(`http://localhost:8080/api/v1/baggage/${baggageId}`);
      if (!res.ok) {
        throw new Error('Багаж не найден');
      }
      const data = await res.json();
      setResult(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <header>
        <h1>Трекинг багажа</h1>
        <p className="subtitle">Отслеживайте местоположение вашего багажа в реальном времени</p>
      </header>

      <section className="dashboard-cards">
        <div className="card glass-effect animate-in">
          <h3>Поиск багажа</h3>
          <p>Введите ID вашего багажа (UUID) для получения актуального статуса.</p>
          <form onSubmit={fetchBaggage} className="search-form">
            <input
              type="text"
              placeholder="e.g. 123e4567-e89b-12d3-..."
              value={baggageId}
              onChange={(e) => setBaggageId(e.target.value)}
              required
            />
            <button type="submit" disabled={loading}>
              {loading ? 'Поиск...' : 'Найти'}
            </button>
          </form>
        </div>
      </section>

      {error && (
        <div className="alert error animate-in">
          <strong>Ошибка:</strong> {error}
        </div>
      )}

      {result && (
        <section className="results-section animate-in">
          <h2>Статус багажа</h2>
          <div className="ticket">
            <div className="ticket-header">
              <span className="status">{result.status}</span>
              <span className="booking-id">ID: {result.id}</span>
            </div>
            <div className="ticket-body">
              <div className="info-group">
                <label>Местоположение</label>
                <span>{result.location}</span>
              </div>
              <div className="info-group">
                <label>ID Пассажира</label>
                <span>{result.passenger_id}</span>
              </div>
              <div className="info-group">
                <label>Последнее обновление</label>
                <span>{new Date(result.updated_at).toLocaleString('ru-RU')}</span>
              </div>
            </div>
          </div>
        </section>
      )}
    </>
  );
}

export default BaggagePage;
