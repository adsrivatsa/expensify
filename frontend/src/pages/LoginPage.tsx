import { loginWithGoogle } from '../api/auth';

export function LoginPage() {
  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'var(--bg)',
        padding: 24,
      }}
    >
      <div style={{ textAlign: 'center', maxWidth: 460, width: '100%' }}>
        {/* Brand mark */}
        <div
          style={{
            display: 'inline-block',
            background: 'var(--yellow)',
            border: 'var(--border)',
            boxShadow: 'var(--shadow-lg)',
            padding: '16px 28px',
            marginBottom: 32,
          }}
        >
          <span style={{ fontSize: 48 }}>💸</span>
        </div>

        {/* Heading */}
        <h1 style={{ fontSize: 40, marginBottom: 12, letterSpacing: '-0.03em' }}>
          Expensify
        </h1>
        <p
          style={{
            color: 'var(--gray-500)',
            fontSize: 16,
            fontWeight: 500,
            marginBottom: 40,
          }}
        >
          Track your spending. Know where your money goes.
        </p>

        {/* Sign in card */}
        <div className="card" style={{ marginBottom: 0 }}>
          <h2 style={{ fontSize: 18, marginBottom: 8 }}>Sign in to get started</h2>
          <p style={{ color: 'var(--gray-500)', fontSize: 14, marginBottom: 28 }}>
            We use Google OAuth — no passwords, ever.
          </p>

          <button
            className="btn btn-primary btn-lg"
            onClick={loginWithGoogle}
            style={{ width: '100%' }}
          >
            <GoogleIcon />
            Continue with Google
          </button>
        </div>

        <p style={{ marginTop: 20, fontSize: 12, color: 'var(--gray-500)' }}>
          By signing in you agree to use this app responsibly.
        </p>
      </div>
    </div>
  );
}

function GoogleIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" aria-hidden="true">
      <path
        fill="#4285F4"
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
      />
      <path
        fill="#34A853"
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
      />
      <path
        fill="#FBBC05"
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
      />
      <path
        fill="#EA4335"
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
      />
    </svg>
  );
}
