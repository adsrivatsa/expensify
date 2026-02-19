import client from './client';
import type {
  PaginatedTransactions,
  ApiEnvelope,
  CreateTransactionPayload,
  UpdateTransactionPayload,
  Transaction,
  CashflowSummary,
} from '../types';

export async function fetchTransactions(page: number, pageSize: number): Promise<PaginatedTransactions> {
  const res = await client.get<ApiEnvelope<PaginatedTransactions>>('/api/transactions', {
    params: { page, page_size: pageSize },
  });
  return res.data.data ?? { items: [], total: 0, page, page_size: pageSize, total_pages: 0 };
}

export async function createTransaction(payload: CreateTransactionPayload): Promise<Transaction> {
  const res = await client.post<ApiEnvelope<Transaction>>('/api/transactions', payload);
  if (!res.data.data) throw new Error('No transaction data returned');
  return res.data.data;
}

export async function updateTransaction(id: string, payload: UpdateTransactionPayload): Promise<Transaction> {
  const res = await client.put<ApiEnvelope<Transaction>>(`/api/transactions/${id}`, payload);
  if (!res.data.data) throw new Error('No transaction data returned');
  return res.data.data;
}

export async function deleteTransaction(id: string): Promise<void> {
  await client.delete(`/api/transactions/${id}`);
}

export async function fetchCashflowSummary(params: { year?: number; months?: number }): Promise<CashflowSummary> {
  const res = await client.get<ApiEnvelope<CashflowSummary>>('/api/cashflow/summary', {
    params,
  });
  return res.data.data ?? { monthly: [], by_category: [] };
}
