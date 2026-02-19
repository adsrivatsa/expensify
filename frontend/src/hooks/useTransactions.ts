import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  fetchTransactions,
  createTransaction,
  updateTransaction,
  deleteTransaction,
  fetchCashflowSummary,
} from '../api/transactions';
import type { CreateTransactionPayload, UpdateTransactionPayload } from '../types';

const PAGE_SIZE = 20;

export function useTransactions(page: number) {
  return useQuery({
    queryKey: ['transactions', page],
    queryFn: () => fetchTransactions(page, PAGE_SIZE),
    placeholderData: (prev) => prev, // Keep previous page visible while loading next
  });
}

export function useCreateTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateTransactionPayload) => createTransaction(payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['transactions'] }),
  });
}

export function useUpdateTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload: UpdateTransactionPayload }) =>
      updateTransaction(id, payload),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['transactions'] }),
  });
}

export function useDeleteTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteTransaction(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['transactions'] }),
  });
}

export function useCashflowSummary(params: { year?: number; months?: number }) {
  return useQuery({
    queryKey: ['cashflow', 'summary', params.year ?? `months-${params.months ?? 12}`],
    queryFn: () => fetchCashflowSummary(params),
  });
}
