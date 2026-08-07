import { Outlet, NavLink, Link, useNavigate } from 'react-router-dom';

const ADMIN_STORAGE_KEY = 'cv_admin';

export default function AdminLayout() {
  const navigate = useNavigate();

  const handleLogout = () => {
    localStorage.removeItem(ADMIN_STORAGE_KEY);
    navigate('/admin/login', { replace: true });
  };

  const linkClass = ({ isActive }: { isActive: boolean }) =>
    `text-sm font-medium transition-colors ${
      isActive ? 'text-brand-700 border-b-2 border-brand-600' : 'text-gray-500 hover:text-gray-700'
    }`;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="border-b bg-white">
        <div className="mx-auto flex max-w-4xl items-center justify-between px-4 py-3">
          <div className="flex items-center gap-6">
            <span className="text-lg font-bold text-gray-900">Admin</span>
            <nav className="flex gap-4">
              <NavLink to="/admin/modules" className={linkClass} end>
                Modules
              </NavLink>
              <NavLink to="/admin/data-entry" className={linkClass}>
                Data Entry
              </NavLink>
            </nav>
          </div>
          <div className="flex items-center gap-4">
            <Link to="/" className="text-xs text-gray-400 hover:text-brand-600 transition-colors">
              View Site →
            </Link>
            <button
              onClick={handleLogout}
              className="text-sm text-red-500 hover:underline"
            >
              Logout
            </button>
          </div>
        </div>
      </header>

      {/* Page content */}
      <main className="mx-auto max-w-3xl px-4 py-8">
        <Outlet />
      </main>
    </div>
  );
}
