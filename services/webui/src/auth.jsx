import React, { createContext, useContext, useEffect, useState } from 'react';
import { api, getAuthToken, setAuthToken } from './api';

const AuthContext = createContext({
  user: null,
  loading: true,
  login: async () => {},
  register: async () => {},
  logout: () => {},
});

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    const token = getAuthToken();
    if (!token) {
      setLoading(false);
      return;
    }
    api.authMe()
      .then((u) => { if (!cancelled) setUser(u); })
      .catch(() => { setAuthToken(''); })
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, []);

  const login = async (email, password) => {
    const { token, user: u } = await api.authLogin({ email, password });
    setAuthToken(token);
    setUser(u);
    return u;
  };

  const register = async (email, password, fullName) => {
    const { token, user: u } = await api.authRegister({ email, password, full_name: fullName });
    setAuthToken(token);
    setUser(u);
    return u;
  };

  const logout = () => {
    setAuthToken('');
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}

export function isAdmin(user) {
  return !!user && user.role === 'admin';
}

export function isPassenger(user) {
  return !!user && user.role === 'passenger';
}
