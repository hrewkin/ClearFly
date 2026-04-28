import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth';

const PASSWORD_HINT = 'От 8 символов, минимум одна буква и одна цифра';

export default function RegisterPage() {
  const { register, registerStaff } = useAuth();
  const navigate = useNavigate();

  const [mode, setMode] = useState('passenger'); // 'passenger' | 'staff'
  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [employeeId, setEmployeeId] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState('');
  const [errors, setErrors] = useState({});
  const [loading, setLoading] = useState(false);

  const validate = () => {
    const errs = {};
    const name = fullName.trim();
    const words = name.split(/\s+/).filter(Boolean);
    if (!name) errs.fullName = 'Укажите ФИО';
    else if (words.length < 2) errs.fullName = 'Введите имя и фамилию (минимум)';
    else if (!/^[А-Яа-яЁёA-Za-z\-\s]+$/.test(name)) errs.fullName = 'Только буквы, дефис и пробел';

    const mail = email.trim();
    if (!mail) errs.email = 'Укажите email';
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(mail)) errs.email = 'Некорректный email';

    if (mode === 'staff') {
      const emp = employeeId.trim();
      if (!emp) errs.employeeId = 'Укажите табельный номер';
      else if (!/^[A-Za-zА-Яа-я0-9\-]{3,20}$/.test(emp)) errs.employeeId = 'От 3 до 20 символов: буквы, цифры и дефис';
    }

    if (!password) errs.password = 'Введите пароль';
    else if (password.length < 8) errs.password = PASSWORD_HINT;
    else if (!/[A-Za-zА-Яа-я]/.test(password) || !/\d/.test(password)) errs.password = PASSWORD_HINT;

    if (password !== confirm) errs.confirm = 'Пароли не совпадают';

    return errs;
  };

  const submit = async (e) => {
    e.preventDefault();
    const errs = validate();
    setErrors(errs);
    if (Object.keys(errs).length) return;
    setError('');
    setLoading(true);
    try {
      if (mode === 'staff') {
        await registerStaff(email.trim(), password, fullName.trim(), employeeId.trim());
      } else {
        await register(email.trim(), password, fullName.trim());
      }
      navigate('/', { replace: true });
    } catch (err) {
      setError(err.message || 'Не удалось зарегистрироваться');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-wrap">
      <div className="auth-card card glass-effect">
        <div className="auth-brand">
          <div className="logo-icon">✈</div>
          <div>
            <h1>ClearFly</h1>
            <small>Чистое небо</small>
          </div>
        </div>
        <h2>Регистрация</h2>

        <div className="role-tabs" role="tablist">
          <button
            type="button"
            role="tab"
            aria-selected={mode === 'passenger'}
            className={`role-tab${mode === 'passenger' ? ' active' : ''}`}
            onClick={() => setMode('passenger')}
          >
            Пассажир
          </button>
          <button
            type="button"
            role="tab"
            aria-selected={mode === 'staff'}
            className={`role-tab${mode === 'staff' ? ' active' : ''}`}
            onClick={() => setMode('staff')}
          >
            Сотрудник
          </button>
        </div>

        <p className="muted small">
          {mode === 'staff'
            ? 'Аккаунт сотрудника авиакомпании. Требуется внутренний табельный номер.'
            : 'Аккаунт пассажира. Профиль и бронирования будут закреплены за ним.'}
        </p>

        <form onSubmit={submit}>
          <label className={`field${errors.fullName ? ' field-error' : ''}`}>
            <span>ФИО</span>
            <input value={fullName} onChange={(e) => { setFullName(e.target.value); if (errors.fullName) setErrors({ ...errors, fullName: undefined }); }} placeholder="Иван Петров" required />
            {errors.fullName && <small className="field-error-msg">{errors.fullName}</small>}
          </label>

          <label className={`field${errors.email ? ' field-error' : ''}`}>
            <span>Email / логин</span>
            <input type="email" value={email} onChange={(e) => { setEmail(e.target.value); if (errors.email) setErrors({ ...errors, email: undefined }); }} placeholder="ivan@example.com" required />
            {errors.email && <small className="field-error-msg">{errors.email}</small>}
          </label>

          {mode === 'staff' && (
            <label className={`field${errors.employeeId ? ' field-error' : ''}`}>
              <span>Табельный номер</span>
              <input value={employeeId} onChange={(e) => { setEmployeeId(e.target.value); if (errors.employeeId) setErrors({ ...errors, employeeId: undefined }); }} placeholder="EMP-12345" required />
              {errors.employeeId && <small className="field-error-msg">{errors.employeeId}</small>}
            </label>
          )}

          <label className={`field${errors.password ? ' field-error' : ''}`}>
            <span>Пароль</span>
            <input type="password" value={password} onChange={(e) => { setPassword(e.target.value); if (errors.password) setErrors({ ...errors, password: undefined }); }} required />
            <small className="field-hint">{PASSWORD_HINT}</small>
            {errors.password && <small className="field-error-msg">{errors.password}</small>}
          </label>

          <label className={`field${errors.confirm ? ' field-error' : ''}`}>
            <span>Подтверждение пароля</span>
            <input type="password" value={confirm} onChange={(e) => { setConfirm(e.target.value); if (errors.confirm) setErrors({ ...errors, confirm: undefined }); }} required />
            {errors.confirm && <small className="field-error-msg">{errors.confirm}</small>}
          </label>

          {error && <div className="alert error">{error}</div>}
          <button className="primary-btn full" type="submit" disabled={loading}>
            {loading ? 'Создаём…' : 'Зарегистрироваться'}
          </button>
        </form>
        <p className="auth-footer">
          Уже есть аккаунт? <Link to="/login">Войти</Link>
        </p>
      </div>
    </div>
  );
}
