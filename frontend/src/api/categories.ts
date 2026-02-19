import client from './client';
import type { Category, ApiEnvelope, CreateCategoryPayload } from '../types';

export async function fetchCategories(): Promise<Category[]> {
  const res = await client.get<ApiEnvelope<Category[]>>('/api/categories');
  return res.data.data ?? [];
}

export async function createCategory(payload: CreateCategoryPayload): Promise<Category> {
  const res = await client.post<ApiEnvelope<Category>>('/api/categories', payload);
  if (!res.data.data) throw new Error('No category data returned');
  return res.data.data;
}

export async function deleteCategory(id: string): Promise<void> {
  await client.delete(`/api/categories/${id}`);
}
