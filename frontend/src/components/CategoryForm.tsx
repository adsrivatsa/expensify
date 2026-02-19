import { useState } from 'react';
import { Modal } from './Modal';
import type { CreateCategoryPayload } from '../types';

// Colors deliberately chosen to not overlap with any default category colors.
const PRESET_COLORS = [
  '#E17055', '#00B894', '#0984E3', '#6C5CE7',
  '#FDCB6E', '#D63031', '#2D3436', '#FF7F50',
  '#5F27CD', '#C44569', '#FF9F43', '#EE5A24',
];

// Icons deliberately chosen to not overlap with any default category icons.
const PRESET_ICONS = ['☕', '🏋️', '🎮', '🐾', '💼', '🎵', '🎨', '🏖️', '💊', '🍷', '🚀', '🌿', '🎓', '💰', '🎯', '🧘'];

interface CategoryFormProps {
  onSubmit: (payload: CreateCategoryPayload) => Promise<void>;
  onClose: () => void;
}

export function CategoryForm({ onSubmit, onClose }: CategoryFormProps) {
  const [name, setName] = useState('');
  const [icon, setIcon] = useState('☕');
  const [color, setColor] = useState('#0984E3');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    if (!name.trim()) {
      setError('Category name is required.');
      return;
    }
    setSubmitting(true);
    try {
      await onSubmit({ name: name.trim(), icon, color });
      onClose();
    } catch {
      setError('Failed to create category. Please try again.');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal
      title="New Category"
      onClose={onClose}
      footer={
        <>
          <button type="button" className="btn btn-ghost" onClick={onClose}>
            Cancel
          </button>
          <button type="submit" form="cat-form" className="btn btn-primary" disabled={submitting}>
            {submitting ? 'Creating…' : 'Create'}
          </button>
        </>
      }
    >
      {error && <div className="error-box">{error}</div>}

      <form id="cat-form" onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div className="form-group">
          <label className="label" htmlFor="cat-name">Name</label>
          <input
            id="cat-name"
            className="input"
            type="text"
            placeholder="e.g. Coffee"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
            autoFocus
            maxLength={50}
          />
        </div>

        <div className="form-group">
          <span className="label">Icon</span>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            {PRESET_ICONS.map((i) => (
              <button
                key={i}
                type="button"
                onClick={() => setIcon(i)}
                style={{
                  width: 40,
                  height: 40,
                  fontSize: 20,
                  border: icon === i ? '3px solid #0a0a0a' : '2px solid #ddd8cc',
                  background: icon === i ? '#ffe500' : '#fff',
                  cursor: 'pointer',
                  boxShadow: icon === i ? '2px 2px 0 #0a0a0a' : 'none',
                }}
              >
                {i}
              </button>
            ))}
          </div>
        </div>

        <div className="form-group">
          <span className="label">Color</span>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
            {PRESET_COLORS.map((c) => (
              <button
                key={c}
                type="button"
                onClick={() => setColor(c)}
                style={{
                  width: 32,
                  height: 32,
                  background: c,
                  border: color === c ? '3px solid #0a0a0a' : '2px solid #ddd8cc',
                  cursor: 'pointer',
                  boxShadow: color === c ? '2px 2px 0 #0a0a0a' : 'none',
                }}
                aria-label={c}
              />
            ))}
          </div>
        </div>

        <div className="form-group">
          <span className="label">Preview</span>
          <span
            className="badge"
            style={{ backgroundColor: color + '33', borderColor: color, fontSize: 14 }}
          >
            {icon} {name || 'Category Name'}
          </span>
        </div>
      </form>
    </Modal>
  );
}
