import { useQueryClient } from '@tanstack/react-query';
import { logout } from '../api/auth';
import type { User } from '../types';

interface NavbarProps {
  user: User;
}

export function Navbar({ user }: NavbarProps) {
  const qc = useQueryClient();

  async function handleLogout() {
    await logout();
    qc.clear();
    window.location.href = '/login';
  }

  return (
    <nav className="navbar">
      <a href="/dashboard" className="navbar-brand">
        💸 Expensify
      </a>
      <div className="navbar-right">
        <span className="navbar-username" style={{ fontWeight: 600, fontSize: 14 }}>{user.name}</span>
        {user.picture && (
          <img src={user.picture} alt={user.name} className="user-avatar" referrerPolicy="no-referrer" />
        )}
        <button className="btn btn-ghost btn-sm" onClick={handleLogout}>
          Logout
        </button>
      </div>
    </nav>
  );
}
