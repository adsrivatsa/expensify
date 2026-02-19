import { useState } from 'react';
import { Navigate } from 'react-router-dom';
import { Navbar } from '../components/Navbar';
import { TransactionList } from '../components/TransactionList';
import { TransactionForm } from '../components/TransactionForm';
import { CategoryForm } from '../components/CategoryForm';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { ChartsView } from '../components/ChartsView';
import { useAuth } from '../hooks/useAuth';
import { useCategories, useCreateCategory, useDeleteCategory } from '../hooks/useCategories';
import { useCreateTransaction } from '../hooks/useTransactions';
import type { CreateTransactionPayload, CreateCategoryPayload, Category } from '../types';

export function DashboardPage() {
  const { user, isLoading, isAuthenticated } = useAuth();
  const { data: categories = [], isLoading: catsLoading } = useCategories();
  const createTxMutation = useCreateTransaction();
  const createCatMutation = useCreateCategory();
  const deleteCatMutation = useDeleteCategory();

  const [showAddTx, setShowAddTx] = useState(false);
  const [showAddCat, setShowAddCat] = useState(false);
  const [activeTab, setActiveTab] = useState<'charts' | 'cashflow' | 'categories'>('charts');
  const [confirmDeleteCat, setConfirmDeleteCat] = useState<Category | null>(null);
  const [deleteCatError, setDeleteCatError] = useState('');

  if (isLoading || catsLoading) {
    return <div className="spinner" style={{ marginTop: 80 }} />;
  }

  if (!isAuthenticated || !user) {
    return <Navigate to="/login" replace />;
  }

  async function handleCreateTransaction(payload: CreateTransactionPayload) {
    await createTxMutation.mutateAsync(payload);
  }

  function handleAddEntry() {
    setActiveTab('cashflow');
    setShowAddTx(true);
  }

  async function handleCreateCategory(payload: CreateCategoryPayload) {
    await createCatMutation.mutateAsync(payload);
  }

  async function handleDeleteCategory() {
    if (!confirmDeleteCat) return;
    setDeleteCatError('');
    try {
      await deleteCatMutation.mutateAsync(confirmDeleteCat.id);
      setConfirmDeleteCat(null);
    } catch (err: unknown) {
      setConfirmDeleteCat(null);
      const status = (err as { response?: { status?: number } })?.response?.status;
      if (status === 409) {
        setDeleteCatError(`"${confirmDeleteCat.name}" has existing transactions and cannot be deleted.`);
      } else {
        setDeleteCatError('Failed to delete category. Please try again.');
      }
    }
  }

  const customCategories = categories.filter((c) => !c.is_default);

  return (
    <>
      <Navbar user={user} />

      <main style={{ flex: 1 }}>
        <div className="container" style={{ paddingTop: 32, paddingBottom: 48 }}>
          {/* Tab bar + Add Entry */}
          <div className="tab-bar-row">
            <div className="tab-group">
              <button
                className="btn btn-flat"
                style={{
                  background: activeTab === 'charts' ? 'var(--black)' : 'var(--white)',
                  color: activeTab === 'charts' ? 'var(--white)' : 'var(--black)',
                  borderRight: 'var(--border)',
                }}
                onClick={() => setActiveTab('charts')}
              >
                Charts
              </button>
              <button
                className="btn btn-flat"
                style={{
                  background: activeTab === 'cashflow' ? 'var(--black)' : 'var(--white)',
                  color: activeTab === 'cashflow' ? 'var(--white)' : 'var(--black)',
                  borderRight: 'var(--border)',
                }}
                onClick={() => setActiveTab('cashflow')}
              >
                Cashflow
              </button>
              <button
                className="btn btn-flat"
                style={{
                  background: activeTab === 'categories' ? 'var(--black)' : 'var(--white)',
                  color: activeTab === 'categories' ? 'var(--white)' : 'var(--black)',
                }}
                onClick={() => setActiveTab('categories')}
              >
                Categories
              </button>
            </div>

            {activeTab !== 'categories' && (
              <button className="btn btn-primary" onClick={handleAddEntry}>
                + Add Entry
              </button>
            )}
            {activeTab === 'categories' && (
              <button className="btn btn-primary" onClick={() => setShowAddCat(true)}>
                + New Category
              </button>
            )}
          </div>

          {activeTab === 'cashflow' && (
            <TransactionList
              categories={categories}
              onAddClick={() => setShowAddTx(true)}
            />
          )}

          {activeTab === 'charts' && (
            <ChartsView onAddClick={() => { setActiveTab('cashflow'); setShowAddTx(true); }} />
          )}

          {activeTab === 'categories' && (
            <div>
              <h1 style={{ fontSize: 28, marginBottom: 28 }}>Categories</h1>

              {/* Default categories */}
              <section style={{ marginBottom: 40 }}>
                <h2 style={{ fontSize: 16, marginBottom: 16, textTransform: 'uppercase', letterSpacing: '0.06em' }}>
                  Default
                </h2>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: 12 }}>
                  {categories
                    .filter((c) => c.is_default)
                    .map((c) => (
                      <div
                        key={c.id}
                        className="card"
                        style={{ padding: '14px 16px', display: 'flex', alignItems: 'center', gap: 10 }}
                      >
                        <span
                          style={{
                            width: 36,
                            height: 36,
                            background: c.color + '33',
                            border: `2px solid ${c.color}`,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            fontSize: 18,
                            flexShrink: 0,
                          }}
                        >
                          {c.icon}
                        </span>
                        <span style={{ fontWeight: 600 }}>{c.name}</span>
                      </div>
                    ))}
                </div>
              </section>

              {/* Custom categories */}
              <section>
                <h2 style={{ fontSize: 16, marginBottom: 16, textTransform: 'uppercase', letterSpacing: '0.06em' }}>
                  Custom ({customCategories.length})
                </h2>
                {deleteCatError && (
                  <div className="error-box" style={{ marginBottom: 16 }}>{deleteCatError}</div>
                )}

                {customCategories.length === 0 ? (
                  <div className="empty-state">
                    <div className="empty-state-icon">🏷️</div>
                    <p style={{ color: 'var(--gray-500)', fontWeight: 500 }}>
                      No custom categories yet. Create one above.
                    </p>
                  </div>
                ) : (
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))', gap: 12 }}>
                    {customCategories.map((c) => (
                      <div
                        key={c.id}
                        className="card"
                        style={{ padding: '14px 16px', display: 'flex', alignItems: 'center', gap: 10 }}
                      >
                        <span
                          style={{
                            width: 36,
                            height: 36,
                            background: c.color + '33',
                            border: `2px solid ${c.color}`,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            fontSize: 18,
                            flexShrink: 0,
                          }}
                        >
                          {c.icon}
                        </span>
                        <span style={{ fontWeight: 600, flex: 1 }}>{c.name}</span>
                        <button
                          className="btn btn-danger btn-sm"
                          onClick={() => { setDeleteCatError(''); setConfirmDeleteCat(c); }}
                          disabled={deleteCatMutation.isPending}
                          title="Delete category"
                        >
                          ✕
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </section>
            </div>
          )}
        </div>
      </main>

      {showAddTx && (
        <TransactionForm
          categories={categories}
          mode={{ kind: 'create' }}
          onSubmit={handleCreateTransaction}
          onClose={() => setShowAddTx(false)}
        />
      )}

      {showAddCat && (
        <CategoryForm
          onSubmit={handleCreateCategory}
          onClose={() => setShowAddCat(false)}
        />
      )}

      {confirmDeleteCat && (
        <ConfirmDialog
          message={`Delete the "${confirmDeleteCat.name}" category? This action cannot be undone.`}
          onConfirm={handleDeleteCategory}
          onClose={() => setConfirmDeleteCat(null)}
        />
      )}
    </>
  );
}
