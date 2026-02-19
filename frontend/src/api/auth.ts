import client from './client';
import type { User, ApiEnvelope } from '../types';

export async function fetchCurrentUser(): Promise<User> {
  const res = await client.get<ApiEnvelope<User>>('/auth/me');
  if (!res.data.data) throw new Error('No user data returned');
  return res.data.data;
}

export async function logout(): Promise<void> {
  await client.post('/auth/logout');
}

/** Navigates to the Google OAuth flow (server-side redirect).
 *  Must go directly to the backend origin so the oauth_state cookie is set
 *  on the same domain as the callback, bypassing the Vite proxy.
 */
export function loginWithGoogle(): void {
  const base = import.meta.env.VITE_API_BASE_URL || '';
  window.location.href = `${base}/auth/google`;
}
