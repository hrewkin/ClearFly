// Centralised API helpers.
//
// All requests go through the API gateway under /api/v1.
//
// In development the WebUI is served from Vite (default :3000) and the
// gateway from Docker on http://localhost:8080. In production the SPA is
// served by Nginx alongside the gateway, so we prefer same-origin paths
// which the reverse proxy can forward.

const ENV_BASE = import.meta?.env?.VITE_API_BASE;
const DEFAULT_BASE = window.location.port === '3000' ? 'http://localhost:8080' : '';
export const API_BASE = (ENV_BASE ?? DEFAULT_BASE).replace(/\/$/, '');

const AIRPORT_NAMES = {
  SVO: 'Москва (Шереметьево)',
  LED: 'Санкт-Петербург (Пулково)',
  AER: 'Сочи (Адлер)',
  KJA: 'Красноярск',
  KZN: 'Казань',
  DME: 'Москва (Домодедово)',
  VKO: 'Москва (Внуково)',
  SVX: 'Екатеринбург',
};

export function airportLabel(code) {
  if (!code) return '';
  return AIRPORT_NAMES[code] ? `${code} · ${AIRPORT_NAMES[code]}` : code;
}

export function airportShort(code) {
  return AIRPORT_NAMES[code]?.split(' ')[0] ?? code;
}

async function request(method, path, body) {
  const res = await fetch(`${API_BASE}/api/v1${path}`, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(text || `${method} ${path} failed: ${res.status}`);
  }
  if (res.status === 204) return null;
  return res.json();
}

export const api = {
  searchFlights: ({ origin, destination, date }) => {
    const qs = new URLSearchParams();
    if (origin) qs.set('origin', origin);
    if (destination) qs.set('destination', destination);
    if (date) qs.set('date', date);
    return request('GET', `/flights/search?${qs.toString()}`);
  },
  upcomingFlights: () => request('GET', '/flights/upcoming'),
  getFlight: (id) => request('GET', `/flights/${id}`),
  getSeats: (id) => request('GET', `/flights/${id}/seats`),
  getTariffs: (id) => request('GET', `/flights/${id}/tariffs`),

  bookSeat: (payload) => request('POST', '/bookings/book', payload),
  getBooking: (id) => request('GET', `/bookings/${id}`),
  checkIn: (id) => request('POST', `/bookings/${id}/checkin`),
  listBookingsByPassenger: (id) => request('GET', `/bookings/passenger/${id}`),

  createIncident: (payload) => request('POST', '/incidents', payload),

  listNotifications: () => request('GET', '/notifications?limit=30'),

  createPassenger: (payload) => request('POST', '/passengers', payload),
  getPassenger: (id) => request('GET', `/passengers/${id}`),
  updatePassenger: (id, payload) => request('PUT', `/passengers/${id}`, payload),
  updatePreferences: (id, payload) => request('PATCH', `/passengers/${id}/preferences`, payload),

  flightLoadFactor: (id) => request('GET', `/analytics/load-factor/${id}`),
};

export function formatPrice(amount, currency = 'RUB') {
  if (amount == null) return '';
  return new Intl.NumberFormat('ru-RU', { style: 'currency', currency, maximumFractionDigits: 0 }).format(amount);
}

export function formatTime(value) {
  if (!value) return '';
  const d = new Date(value);
  return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
}

export function formatDate(value) {
  if (!value) return '';
  const d = new Date(value);
  return d.toLocaleDateString('ru-RU', { day: '2-digit', month: 'short' });
}

export function durationLabel(start, end) {
  if (!start || !end) return '';
  const ms = new Date(end) - new Date(start);
  const totalMin = Math.round(ms / 60000);
  const hours = Math.floor(totalMin / 60);
  const minutes = totalMin % 60;
  if (hours <= 0) return `${minutes}м`;
  return `${hours}ч ${minutes.toString().padStart(2, '0')}м`;
}
