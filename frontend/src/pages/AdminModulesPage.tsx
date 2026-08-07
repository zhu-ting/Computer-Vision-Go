import { useState, useEffect, useCallback } from 'react';
import { listModules, createModule } from '../api/client';
import type { Module } from '../types';

export default function AdminModulesPage() {
  const [modules, setModules] = useState<Module[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Form state
  const [name, setName] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [success, setSuccess] = useState<string | null>(null);

  const fetchModules = useCallback(async () => {
    setLoading(true);
    try {
      setModules(await listModules());
    } catch {
      setError('Failed to load modules');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchModules();
  }, [fetchModules]);

  const handleAdd = async () => {
    if (!name.trim()) return;
    setSubmitting(true);
    setError(null);
    setSuccess(null);
    try {
      await createModule({ name: name.trim() });
      setName('');
      setSuccess(`Module "${name.trim()}" created.`);
      await fetchModules();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create module');
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="text-center py-16">
        <p className="text-gray-500">Loading modules...</p>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900">Modules</h1>
      <p className="mt-1 text-sm text-gray-500">
        Modules are top-level themes (e.g., "week7_deep_learning"). Each module
        contains question groups with related questions.
      </p>

      {/* Add form */}
      <div className="mt-6 rounded-xl border bg-white p-5 shadow-sm">
        <label htmlFor="module-name" className="block text-sm font-medium text-gray-700">
          New module name
        </label>
        <div className="mt-2 flex gap-3">
          <input
            id="module-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
            placeholder="e.g., week7_deep_learning"
            className="flex-1 rounded-lg border border-gray-200 px-3 py-2 text-sm
                       focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
          />
          <button
            onClick={handleAdd}
            disabled={submitting || !name.trim()}
            className="rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white
                       hover:bg-brand-700 disabled:opacity-50 transition-colors"
          >
            {submitting ? 'Adding...' : 'Add Module'}
          </button>
        </div>

        {error && (
          <p className="mt-3 rounded-lg bg-red-50 p-2 text-sm text-red-600">{error}</p>
        )}
        {success && (
          <p className="mt-3 rounded-lg bg-green-50 p-2 text-sm text-green-600">{success}</p>
        )}
      </div>

      {/* Modules list */}
      {modules.length === 0 ? (
        <p className="mt-8 text-center text-gray-400">
          No modules yet — create one to get started.
        </p>
      ) : (
        <div className="mt-4 space-y-3">
          {modules.map((m) => (
            <div key={m.id} className="rounded-xl border bg-white p-5 shadow-sm">
              <div className="flex items-center justify-between">
                <span className="font-medium text-gray-900">{m.name}</span>
                <span className="text-xs text-gray-400">
                  {new Date(m.created_at).toLocaleDateString()}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
