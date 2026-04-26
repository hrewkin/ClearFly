import React, { useState } from 'react';

function AnalyticsPage() {
  const [flightId, setFlightId] = useState('');
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchAnalytics = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setResult(null);

    try {
      const res = await fetch(`http://localhost:8080/api/v1/analytics/load-factor/${flightId}`);
      if (!res.ok) {
        throw new Error('Данные не найдены');
      }
      const data = await res.json();
      setResult(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const getLoadColor = (factor) => {
    if (factor > 80) return '#ef4444';
    if (factor > 50) return '#f59e0b';
    return '#10b981';
  };

  return (
    <>
      <header>
        <h1>Аналитика рейсов</h1>
        <p className="subtitle">Загрузка рейсов и рекомендации по ценообразованию</p>
      </header>

      <section className="dashboard-cards">
        <div className="card glass-effect animate-in">
          <h3>Анализ загрузки рейса</h3>
          <p>Введите ID рейса для получения коэффициента загрузки и рекомендуемой цены.</p>
          <form onSubmit={fetchAnalytics} className="search-form">
            <input
              type="text"
              placeholder="e.g. 123e4567-e89b-12d3-..."
              value={flightId}
              onChange={(e) => setFlightId(e.target.value)}
              required
            />
            <button type="submit" disabled={loading}>
              {loading ? 'Анализ...' : 'Анализировать'}
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
          <h2>Результаты анализа</h2>
          <div className="analytics-grid">
            <div className="card glass-effect">
              <h3>Загрузка рейса</h3>
              <div className="metric-value" style={{ color: getLoadColor(result.analytics?.load_factor || 0) }}>
                {result.analytics?.load_factor || 0}%
              </div>
              <p>Забронировано мест: {result.analytics?.total_bookings || 0} / 150</p>
            </div>
            <div className="card glass-effect">
              <h3>Рекомендуемая цена</h3>
              <div className="metric-value" style={{ color: 'var(--accent)' }}>
                ${result.suggested_price?.toFixed(2)}
              </div>
              <p>
                {result.analytics?.load_factor > 80 && 'Высокий спрос — коэффициент ×1.5'}
                {result.analytics?.load_factor > 50 && result.analytics?.load_factor <= 80 && 'Средний спрос — коэффициент ×1.2'}
                {result.analytics?.load_factor <= 50 && 'Базовая цена'}
              </p>
            </div>
          </div>
        </section>
      )}
    </>
  );
}

export default AnalyticsPage;
