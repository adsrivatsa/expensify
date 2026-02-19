import { useState } from 'react';
import { Modal } from './Modal';
import type { Category, Transaction } from '../types';
import type { CreateTransactionPayload, UpdateTransactionPayload } from '../types';

type FormMode =
  | { kind: 'create' }
  | { kind: 'edit'; transaction: Transaction };

interface TransactionFormProps {
  categories: Category[];
  mode: FormMode;
  onSubmit: (payload: CreateTransactionPayload | UpdateTransactionPayload) => Promise<void>;
  onClose: () => void;
}

function toDateInputValue(iso: string): string {
  return iso.split('T')[0];
}

export function TransactionForm({ categories, mode, onSubmit, onClose }: TransactionFormProps) {
  const isEdit = mode.kind === 'edit';
  const tx = isEdit ? mode.transaction : null;

  const today = new Date().toISOString().split('T')[0];

  const otherCategory = categories.find((c) => c.name.toLowerCase() === 'other');
  const firstCategory = categories[0];

  const initialCategory = tx?.category_id
    ?? ((tx?.type ?? 'outflow') === 'inflow' ? (otherCategory?.id ?? firstCategory?.id ?? '') : (firstCategory?.id ?? ''));

  const [categoryId, setCategoryId] = useState(initialCategory);
  const [txType, setTxType] = useState<'inflow' | 'outflow'>(tx?.type ?? 'outflow');

  function handleSetType(newType: 'inflow' | 'outflow') {
    setTxType(newType);
    if (!isEdit) {
      if (newType === 'inflow') {
        setCategoryId(otherCategory?.id ?? firstCategory?.id ?? '');
      } else {
        setCategoryId(firstCategory?.id ?? '');
      }
    }
  }
  const [amount, setAmount] = useState(tx ? String(tx.amount) : '');
  const [description, setDescription] = useState(tx?.description ?? '');
  const [date, setDate] = useState(tx ? toDateInputValue(tx.date) : today);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');

    const parsedAmount = parseFloat(amount);
    if (isNaN(parsedAmount) || parsedAmount <= 0) {
      setError('Amount must be a positive number.');
      return;
    }
    if (!categoryId) {
      setError('Please select a category.');
      return;
    }
    if (!date) {
      setError('Please select a date.');
      return;
    }

    const payload: CreateTransactionPayload = {
      category_id: categoryId,
      type: txType,
      amount: parsedAmount,
      description: description.trim(),
      date: new Date(date).toISOString(),
    };

    setSubmitting(true);
    try {
      await onSubmit(payload);
      onClose();
    } catch {
      setError('Something went wrong. Please try again.');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal
      title={isEdit ? 'Edit Transaction' : 'Add Transaction'}
      onClose={onClose}
      footer={
        <>
          <button type="button" className="btn btn-ghost" onClick={onClose}>
            Cancel
          </button>
          <button
            type="submit"
            form="tx-form"
            className="btn btn-primary"
            disabled={submitting}
          >
            {submitting ? 'Saving…' : isEdit ? 'Save Changes' : 'Add Transaction'}
          </button>
        </>
      }
    >
      {error && <div className="error-box">{error}</div>}

      <form id="tx-form" onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div className="form-group">
          <label className="label">Type</label>
          <div style={{ display: 'flex', gap: 0, border: 'var(--border)', width: 'fit-content', boxShadow: 'var(--shadow)' }}>
            <button
              type="button"
              onClick={() => handleSetType('outflow')}
              style={{
                padding: '8px 20px',
                fontWeight: 700,
                border: 'none',
                borderRight: 'var(--border)',
                cursor: 'pointer',
                background: txType === 'outflow' ? 'var(--black)' : 'var(--white)',
                color: txType === 'outflow' ? '#ff6b6b' : 'var(--black)',
              }}
            >
              💸 Outflow
            </button>
            <button
              type="button"
              onClick={() => handleSetType('inflow')}
              style={{
                padding: '8px 20px',
                fontWeight: 700,
                border: 'none',
                cursor: 'pointer',
                background: txType === 'inflow' ? 'var(--black)' : 'var(--white)',
                color: txType === 'inflow' ? '#51cf66' : 'var(--black)',
              }}
            >
              💰 Inflow
            </button>
          </div>
        </div>

        <div className="form-group">
          <label className="label" htmlFor="tx-amount">Amount ($)</label>
          <input
            id="tx-amount"
            className="input"
            type="number"
            min="0.01"
            step="0.01"
            placeholder="0.00"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            required
            autoFocus
          />
        </div>

        <div className="form-group">
          <label className="label" htmlFor="tx-category">Category</label>
          <select
            id="tx-category"
            className="select"
            value={categoryId}
            onChange={(e) => setCategoryId(e.target.value)}
            required
          >
            {categories.map((c) => (
              <option key={c.id} value={c.id}>
                {c.icon} {c.name}
              </option>
            ))}
          </select>
        </div>

        <div className="form-group">
          <label className="label" htmlFor="tx-description">Description</label>
          <input
            id="tx-description"
            className="input"
            type="text"
            placeholder="What was this for?"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            maxLength={200}
          />
        </div>

        <div className="form-group">
          <label className="label" htmlFor="tx-date">Date</label>
          <input
            id="tx-date"
            className="input"
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
            required
          />
        </div>
      </form>
    </Modal>
  );
}
