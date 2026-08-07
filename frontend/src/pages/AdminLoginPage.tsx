import { useState, type FormEvent } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';

const ADMIN_PASSWORD = import.meta.env.VITE_ADMIN_PASSWORD ?? 'admin123';
const ADMIN_STORAGE_KEY = 'cv_admin';

export default function AdminLoginPage() {
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();
  const location = useLocation();

  const from = (location.state as { from?: string })?.from ?? '/admin/modules';

  const handleSubmit = (e: FormEvent) => {
    e.preventDefault();
    if (password === ADMIN_PASSWORD) {
      localStorage.setItem(ADMIN_STORAGE_KEY, 'true');
      navigate(from, { replace: true });
    } else {
      setError('Incorrect password');
      setPassword('');
    }
  };

  return (
    <main className="mx-auto max-w-sm px-4 py-16">
      <div className="rounded-xl border bg-white p-6 shadow-sm">
        <h1 className="text-xl font-bold text-gray-900">Admin Login</h1>
        <p className="mt-1 text-sm text-gray-500">
          Enter the admin password to continue.
        </p>

        <form onSubmit={handleSubmit} className="mt-6 space-y-4">
          {error && (
            <p className="rounded-lg bg-red-50 p-3 text-sm text-red-600">{error}</p>
          )}

          <div>
            <label htmlFor="password" className="block text-sm font-medium text-gray-700">
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoFocus
              className="mt-1 w-full rounded-lg border border-gray-200 p-3 text-sm
                         focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
              placeholder="Enter password"
            />
          </div>

          <button
            type="submit"
            disabled={!password}
            className="w-full rounded-lg bg-brand-600 px-4 py-3 text-sm font-medium text-white
                       hover:bg-brand-700 disabled:opacity-50 transition-colors"
          >
            Sign In
          </button>
        </form>
      </div>
    </main>
  );
}
