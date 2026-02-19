import { useState } from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
} from 'recharts';
import { useCashflowSummary } from '../hooks/useTransactions';

const MONTH_NAMES = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

function formatCurrency(value: number): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(value);
}

function monthLabel(year: number, month: number): string {
  return `${MONTH_NAMES[month - 1]} '${String(year).slice(2)}`;
}

const FALLBACK_COLORS = ['#ff6b6b', '#51cf66', '#339af0', '#fcc419', '#cc5de8', '#20c997', '#ff922b', '#74c0fc'];

// Trailing-12-months label: "Feb '25 – Jan '26" based on today
function trailingLabel(): string {
  const end = new Date();
  const start = new Date(end);
  start.setMonth(start.getMonth() - 12);
  const fmt = (d: Date) => `${MONTH_NAMES[d.getMonth()]} '${String(d.getFullYear()).slice(2)}`;
  return `${fmt(start)} – ${fmt(end)}`;
}

interface ChartsViewProps {
  onAddClick: () => void;
}

export function ChartsView({ onAddClick }: ChartsViewProps) {
  const currentYear = new Date().getFullYear();
  const isMobile = window.matchMedia('(max-width: 640px)').matches;

  // null = trailing 12 months; number = that calendar year
  const [selectedYear, setSelectedYear] = useState<number | null>(null);

  const queryParams = selectedYear !== null ? { year: selectedYear } : { months: 12 };
  const { data, isLoading, isError } = useCashflowSummary(queryParams);

  // Navigation helpers
  function goPrev() {
    if (selectedYear === null) {
      setSelectedYear(currentYear - 1);
    } else {
      setSelectedYear(selectedYear - 1);
    }
  }

  function goNext() {
    if (selectedYear === null) return; // already at the most recent
    if (selectedYear + 1 >= currentYear) {
      setSelectedYear(null); // go back to trailing view
    } else {
      setSelectedYear(selectedYear + 1);
    }
  }

  const periodLabel = selectedYear !== null ? String(selectedYear) : 'Last 12 Months';
  const periodSub = selectedYear !== null ? `Jan – Dec ${selectedYear}` : trailingLabel();
  const canGoNext = selectedYear !== null;

  if (isLoading) return <div className="spinner" />;
  if (isError) return <div className="error-box">Failed to load chart data.</div>;

  const { monthly = [], by_category = [] } = data ?? {};
  const hasData = monthly.length > 0 || by_category.length > 0;

  if (!hasData) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 32 }}>
        <PeriodNav
          label={periodLabel}
          sub={periodSub}
          canGoNext={canGoNext}
          onPrev={goPrev}
          onNext={goNext}
        />
        <div className="empty-state">
          <div className="empty-state-icon">📊</div>
          <h2 style={{ marginBottom: 8 }}>No data for this period</h2>
          <p style={{ color: 'var(--gray-500)', marginBottom: 20 }}>
            {selectedYear !== null
              ? `No cashflow entries found for ${selectedYear}.`
              : 'Add some cashflow entries to see your charts here.'}
          </p>
          {selectedYear === null && (
            <button className="btn btn-primary" onClick={onAddClick}>+ Add Entry</button>
          )}
        </div>
      </div>
    );
  }

  const totalInflow = monthly.reduce((s, m) => s + m.inflow, 0);
  const totalOutflow = monthly.reduce((s, m) => s + m.outflow, 0);
  const netBalance = totalInflow - totalOutflow;

  const barData = monthly.map((m) => ({
    name: monthLabel(m.year, m.month),
    Inflow: m.inflow,
    Outflow: m.outflow,
  }));

  const pieData = by_category.map((c, i) => ({
    name: c.category_name || 'Unknown',
    value: c.total,
    color: c.category_color || FALLBACK_COLORS[i % FALLBACK_COLORS.length],
    icon: c.category_icon || '📦',
  }));

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 32 }}>
      <PeriodNav
        label={periodLabel}
        sub={periodSub}
        canGoNext={canGoNext}
        onPrev={goPrev}
        onNext={goNext}
      />

      {/* Stat cards */}
      <div className="stats-grid">
        <div className="card" style={{ padding: '20px 24px' }}>
          <p style={{ fontSize: 12, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--gray-500)', marginBottom: 8 }}>
            Total Inflow
          </p>
          <p className="stat-value" style={{ color: '#51cf66' }}>{formatCurrency(totalInflow)}</p>
        </div>
        <div className="card" style={{ padding: '20px 24px' }}>
          <p style={{ fontSize: 12, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--gray-500)', marginBottom: 8 }}>
            Total Outflow
          </p>
          <p className="stat-value" style={{ color: '#ff6b6b' }}>{formatCurrency(totalOutflow)}</p>
        </div>
        <div className="card" style={{ padding: '20px 24px' }}>
          <p style={{ fontSize: 12, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--gray-500)', marginBottom: 8 }}>
            Net Balance
          </p>
          <p className="stat-value" style={{ color: netBalance >= 0 ? '#51cf66' : '#ff6b6b' }}>
            {formatCurrency(netBalance)}
          </p>
        </div>
      </div>

      {/* Monthly bar chart */}
      {monthly.length > 0 && (
        <div className="card" style={{ padding: 24 }}>
          <h2 style={{ fontSize: 16, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: 24 }}>
            Monthly Overview
          </h2>
          <ResponsiveContainer width="100%" height={280}>
            <BarChart data={barData} margin={{ top: 0, right: isMobile ? 4 : 0, left: 0, bottom: 0 }}>
              <XAxis dataKey="name" tick={{ fontFamily: 'var(--font)', fontSize: isMobile ? 10 : 13, fontWeight: 600 }} />
              <YAxis
                tickFormatter={(v: number) => v >= 1000 ? `$${v / 1000}k` : `$${v}`}
                tick={{ fontFamily: 'var(--font)', fontSize: isMobile ? 10 : 12 }}
                width={isMobile ? 36 : 60}
              />
              <Tooltip formatter={(value: number) => formatCurrency(value)} />
              <Bar dataKey="Inflow" fill="#51cf66" stroke="#0a0a0a" strokeWidth={2} />
              <Bar dataKey="Outflow" fill="#ff6b6b" stroke="#0a0a0a" strokeWidth={2} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Category pie chart */}
      {pieData.length > 0 && (
        <div className="card" style={{ padding: 24 }}>
          <h2 style={{ fontSize: 16, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: 24 }}>
            Spending by Category
          </h2>
          <div className="pie-layout">
            <PieChart width={240} height={240}>
              <Pie
                data={pieData}
                cx={110}
                cy={110}
                outerRadius={100}
                dataKey="value"
                strokeWidth={3}
                stroke="#0a0a0a"
              >
                {pieData.map((entry, i) => (
                  <Cell key={i} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip formatter={(value: number) => formatCurrency(value)} />
            </PieChart>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {pieData.map((entry, i) => (
                <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                  <span
                    style={{
                      width: 16,
                      height: 16,
                      background: entry.color,
                      border: '2px solid #0a0a0a',
                      flexShrink: 0,
                    }}
                  />
                  <span style={{ fontWeight: 600, fontSize: 14 }}>
                    {entry.icon} {entry.name}
                  </span>
                  <span style={{ fontSize: 14, color: 'var(--gray-500)', marginLeft: 'auto', paddingLeft: 16 }}>
                    {formatCurrency(entry.value)}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// ── Period navigation bar ─────────────────────────────────────────────────────

interface PeriodNavProps {
  label: string;
  sub: string;
  canGoNext: boolean;
  onPrev: () => void;
  onNext: () => void;
}

function PeriodNav({ label, sub, canGoNext, onPrev, onNext }: PeriodNavProps) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 12 }}>
      <button className="page-btn" onClick={onPrev} title="Previous period">
        ←
      </button>
      <div style={{ textAlign: 'center', minWidth: 200 }}>
        <div style={{ fontWeight: 800, fontSize: 20, letterSpacing: '-0.02em' }}>{label}</div>
        <div style={{ fontSize: 13, color: 'var(--gray-500)', fontWeight: 500, marginTop: 2 }}>{sub}</div>
      </div>
      <button
        className="page-btn"
        onClick={onNext}
        disabled={!canGoNext}
        title={canGoNext ? 'Next period' : 'Already at most recent'}
      >
        →
      </button>
    </div>
  );
}
