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

/** Navigates to the Google OAuth flow (server-side redirect). */
export function loginWithGoogle(): void {
  window.location.href = '/auth/google';
}
