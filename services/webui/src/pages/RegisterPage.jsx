import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth';

const PASSWORD_HINT = 'От 8 символов, минимум одна буква и одна цифра';

export default function RegisterPage() {
  const { register } = useAuth();
  const navigate = useNavigate();

  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
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
      await register(email.trim(), password, fullName.trim());
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
        <h2>Регистрация пассажира</h2>
        <p className="muted small">
          Аккаунты сотрудников создаются администратором — обратитесь в кадровую службу авиакомпании.
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
