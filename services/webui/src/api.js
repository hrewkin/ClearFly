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

const AUTH_TOKEN_KEY = 'clearfly_token';

export function getAuthToken() {
  try { return localStorage.getItem(AUTH_TOKEN_KEY) || ''; } catch { return ''; }
}

export function setAuthToken(token) {
  try {
    if (token) localStorage.setItem(AUTH_TOKEN_KEY, token);
    else localStorage.removeItem(AUTH_TOKEN_KEY);
  } catch { /* ignore storage errors */ }
}

async function request(method, path, body) {
  const headers = {};
  if (body) headers['Content-Type'] = 'application/json';
  const token = getAuthToken();
  if (token) headers['Authorization'] = `Bearer ${token}`;
  const res = await fetch(`${API_BASE}/api/v1${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    let msg = res.statusText;
    try {
      const text = await res.text();
      if (text) {
        try {
          const j = JSON.parse(text);
          msg = j.error || j.message || text;
        } catch { msg = text; }
      }
    } catch { /* ignore body parse errors */ }
    throw new Error(msg || `${method} ${path} failed: ${res.status}`);
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

  listBaggage: (params = {}) => {
    const qs = new URLSearchParams();
    if (params.passenger_id) qs.set('passenger_id', params.passenger_id);
    if (params.flight_id) qs.set('flight_id', params.flight_id);
    if (params.limit) qs.set('limit', params.limit);
    const s = qs.toString();
    return request('GET', `/baggage${s ? `?${s}` : ''}`);
  },
  createBaggage: (payload) => request('POST', '/baggage', payload),
  scanBaggage: (id, payload = {}) => request('POST', `/baggage/${id}/scan`, payload),

  flightLoadFactor: (id) => request('GET', `/analytics/load-factor/${id}`),

  authRegister: (payload) => request('POST', '/auth/register', payload),
  authRegisterStaff: (payload) => request('POST', '/auth/register-staff', payload),
  authLogin: (payload) => request('POST', '/auth/login', payload),
  authMe: () => request('GET', '/auth/me'),

  listNotificationsByPassenger: (id) => request('GET', `/notifications/${id}`),

  listFlightBookings: (flightId) => request('GET', `/bookings/flight/${flightId}`),
  staffRefund: (payload) => request('POST', '/staff/refund', payload),
  cancelOwnBooking: (id) => request('POST', `/staff/bookings/${id}/cancel`),
  staffAudit: () => request('GET', '/staff/audit'),
};

export const BAGGAGE_STAGES = [
  { key: 'CHECKED_IN', label: 'Сдан', icon: '🧳', location: 'Стойка регистрации' },
  { key: 'SCREENED', label: 'Досмотрен', icon: '🛂', location: 'Интроскоп' },
  { key: 'LOADED', label: 'Загружен', icon: '📦', location: 'Багажный люк' },
  { key: 'IN_FLIGHT', label: 'В полёте', icon: '✈️', location: 'На борту' },
  { key: 'UNLOADED', label: 'Выгружен', icon: '📤', location: 'Багажная лента' },
  { key: 'CLAIMED', label: 'Получен', icon: '✅', location: 'Выдан пассажиру' },
];

export function baggageStageIndex(status) {
  return BAGGAGE_STAGES.findIndex((s) => s.key === status);
}

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

const FLIGHT_STATUS_LABELS = {
  SCHEDULED: 'Запланирован',
  DELAYED: 'Задержан',
  CANCELLED: 'Отменён',
  BOARDING: 'Посадка',
  DEPARTED: 'Вылетел',
  ARRIVED: 'Прибыл',
  COMPLETED: 'Завершён',
  GATE_CHANGED: 'Смена выхода',
};

export function flightStatusLabel(status) {
  if (!status) return FLIGHT_STATUS_LABELS.SCHEDULED;
  return FLIGHT_STATUS_LABELS[status] || status;
}

const BOOKING_STATUS_LABELS = {
  CONFIRMED: 'Подтверждено',
  CANCELLED: 'Отменено',
  PENDING: 'Ожидает оплаты',
  CHECKED_IN: 'Регистрация пройдена',
};

export function bookingStatusLabel(status) {
  if (!status) return '';
  return BOOKING_STATUS_LABELS[status] || status;
}

const INCIDENT_LABELS = {
  FLIGHT_DELAYED: 'Задержка',
  FLIGHT_CANCELLED: 'Отмена',
  GATE_CHANGED: 'Смена выхода',
  BOARDING: 'Посадка',
  DEPARTURE: 'Вылет',
};

export function incidentLabel(type) {
  if (!type) return '';
  return INCIDENT_LABELS[type] || type;
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
