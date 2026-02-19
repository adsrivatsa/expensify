import axios from 'axios';

// In development the Vite proxy rewrites /api and /auth to the backend,
// so baseURL '/' works. In production set VITE_API_BASE_URL to the backend
// origin (e.g. https://expensify-backend.domain.com).
const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/',
  withCredentials: true, // Always send session cookie
  headers: { 'Content-Type': 'application/json' },
});

// If the server returns 401, the caller (React Query) will surface it as an error.
// We don't do a global redirect here; that's handled in the AuthContext.
client.interceptors.response.use(
  (res) => res,
  (err) => Promise.reject(err),
);

export default client;
