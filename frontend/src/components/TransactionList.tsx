import { useState } from 'react';
import { CategoryBadge } from './CategoryBadge';
import { TransactionForm } from './TransactionForm';
import { ConfirmDialog } from './ConfirmDialog';
import { Pagination } from './Pagination';
import { useTransactions, useDeleteTransaction, useUpdateTransaction } from '../hooks/useTransactions';
import type { Category, Transaction, UpdateTransactionPayload } from '../types';

interface TransactionListProps {
  categories: Category[];
  onAddClick: () => void;
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
}

function formatAmount(amount: number): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(amount);
}

export function TransactionList({ categories, onAddClick }: TransactionListProps) {
  const [page, setPage] = useState(1);
  const [editingTx, setEditingTx] = useState<Transaction | null>(null);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  const { data, isLoading, isError } = useTransactions(page);
  const deleteMutation = useDeleteTransaction();
  const updateMutation = useUpdateTransaction();

  async function handleConfirmDelete() {
    if (!confirmDeleteId) return;
    await deleteMutation.mutateAsync(confirmDeleteId);
    setConfirmDeleteId(null);
  }

  async function handleUpdate(payload: UpdateTransactionPayload) {
    if (!editingTx) return;
    await updateMutation.mutateAsync({ id: editingTx.id, payload });
  }

  if (isLoading) return <div className="spinner" />;
  if (isError) return <div className="error-box">Failed to load transactions.</div>;

  const { items = [], total_pages = 1, total = 0 } = data ?? {};

  return (
    <>
      <div style={{ marginBottom: 28 }}>
        <h1 style={{ fontSize: 28 }}>Cashflow</h1>
        {total > 0 && (
          <p style={{ color: 'var(--gray-500)', marginTop: 4, fontWeight: 500 }}>
            {total} total {total === 1 ? 'entry' : 'entries'}
          </p>
        )}
      </div>

      {items.length === 0 ? (
        <div className="empty-state">
          <div className="empty-state-icon">💸</div>
          <h2 style={{ marginBottom: 8 }}>No transactions yet</h2>
          <p style={{ color: 'var(--gray-500)', marginBottom: 20 }}>
            Start tracking your spending by adding your first transaction.
          </p>
          <button className="btn btn-primary" onClick={onAddClick}>
            + Add Entry
          </button>
        </div>
      ) : (
        <>
          <div className="card tx-table-wrap" style={{ padding: 0 }}>
            <table className="tx-table">
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Category</th>
                  <th>Description</th>
                  <th>Type</th>
                  <th style={{ textAlign: 'right' }}>Amount</th>
                  <th style={{ textAlign: 'center' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {items.map((tx) => (
                  <tr key={tx.id}>
                    <td style={{ whiteSpace: 'nowrap', color: 'var(--gray-500)', fontSize: 13 }}>
                      {formatDate(tx.date)}
                    </td>
                    <td>
                      <CategoryBadge
                        icon={tx.category_icon || '📦'}
                        name={tx.category_name || 'Unknown'}
                        color={tx.category_color || '#B2BEC3'}
                      />
                    </td>
                    <td style={{ maxWidth: 280, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                      {tx.description || <span style={{ color: 'var(--gray-500)' }}>—</span>}
                    </td>
                    <td>
                      {tx.type === 'inflow' ? (
                        <span style={{ fontWeight: 700, color: '#51cf66', fontSize: 13 }}>💰 Inflow</span>
                      ) : (
                        <span style={{ fontWeight: 700, color: '#ff6b6b', fontSize: 13 }}>💸 Outflow</span>
                      )}
                    </td>
                    <td style={{ textAlign: 'right' }}>
                      <span className={tx.type === 'inflow' ? 'amount-positive' : 'amount-negative'}>
                        {formatAmount(tx.amount)}
                      </span>
                    </td>
                    <td>
                      <div className="action-row" style={{ justifyContent: 'center' }}>
                        <button
                          className="btn btn-ghost btn-sm"
                          onClick={() => setEditingTx(tx)}
                          title="Edit"
                        >
                          ✏️
                        </button>
                        <button
                          className="btn btn-danger btn-sm"
                          onClick={() => setConfirmDeleteId(tx.id)}
                          disabled={deleteMutation.isPending}
                          title="Delete"
                        >
                          🗑
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <Pagination page={page} totalPages={total_pages} onPageChange={setPage} />
        </>
      )}

      {editingTx && (
        <TransactionForm
          categories={categories}
          mode={{ kind: 'edit', transaction: editingTx }}
          onSubmit={handleUpdate}
          onClose={() => setEditingTx(null)}
        />
      )}

      {confirmDeleteId && (
        <ConfirmDialog
          message="Delete this transaction? This action cannot be undone."
          onConfirm={handleConfirmDelete}
          onClose={() => setConfirmDeleteId(null)}
        />
      )}
    </>
  );
}
