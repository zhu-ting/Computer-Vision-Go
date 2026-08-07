import { Routes, Route, Link } from 'react-router-dom';
import HomePage from './pages/HomePage';
import ExamPage from './pages/ExamPage';
import ResultPage from './pages/ResultPage';
import NotesPage from './pages/NotesPage';
import AdminLoginPage from './pages/AdminLoginPage';
import AdminModulesPage from './pages/AdminModulesPage';
import AdminDataEntryPage from './pages/AdminDataEntryPage';
import ProtectedRoute from './components/ProtectedRoute';
import AdminLayout from './components/AdminLayout';

export default function App() {
  return (
    <div className="min-h-screen bg-gray-50">
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/exam/:examId" element={<ExamPage />} />
        <Route path="/result/:examId" element={<ResultPage />} />
        <Route path="/notes" element={<NotesPage />} />

        {/* Admin routes */}
        <Route path="/admin/login" element={<AdminLoginPage />} />
        <Route
          path="/admin"
          element={
            <ProtectedRoute>
              <AdminLayout />
            </ProtectedRoute>
          }
        >
          <Route path="modules" element={<AdminModulesPage />} />
          <Route path="data-entry" element={<AdminDataEntryPage />} />
        </Route>

        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </div>
  );
}

function NotFoundPage() {
  return (
    <main className="mx-auto max-w-4xl px-4 py-16 text-center">
      <h1 className="text-6xl font-bold text-gray-300">404</h1>
      <p className="mt-4 text-gray-500">Page not found</p>
      <Link to="/" className="mt-4 inline-block text-brand-600 hover:underline">
        Back to home
      </Link>
    </main>
  );
}
