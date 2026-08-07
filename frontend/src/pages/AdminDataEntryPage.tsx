import { useState, useEffect, useCallback } from 'react';
import {
  listModules,
  listQuestions,
  createQuestion,
  deleteQuestion,
} from '../api/client';
import type { Module, AdminQuestion } from '../types';

export default function AdminDataEntryPage() {
  // ── Modules ───────────────────────────────────────────────────
  const [modules, setModules] = useState<Module[]>([]);
  const [selectedModuleId, setSelectedModuleId] = useState<number | ''>('');
  const [loadingModules, setLoadingModules] = useState(true);

  // ── Question form ─────────────────────────────────────────────
  const [questions, setQuestions] = useState<AdminQuestion[]>([]);
  const [qContent, setQContent] = useState('');
  const [qAnalysis, setQAnalysis] = useState('');
  const [options, setOptions] = useState<{ content: string; is_correct: boolean }[]>([
    { content: '', is_correct: true },
    { content: '', is_correct: false },
    { content: '', is_correct: false },
    { content: '', is_correct: false },
  ]);
  const [qSubmitting, setQSubmitting] = useState(false);
  const [qError, setQError] = useState<string | null>(null);
  const [qSuccess, setQSuccess] = useState<string | null>(null);
  const [deletingId, setDeletingId] = useState<number | null>(null);

  // ── Load modules ──────────────────────────────────────────────
  const fetchModules = useCallback(async () => {
    setLoadingModules(true);
    try {
      setModules(await listModules());
    } catch {
      // silent
    } finally {
      setLoadingModules(false);
    }
  }, []);

  useEffect(() => {
    fetchModules();
  }, [fetchModules]);

  // ── Load questions when module changes ────────────────────────
  const fetchQuestions = useCallback(async (moduleId: number) => {
    try {
      setQuestions(await listQuestions({ moduleId }));
    } catch {
      setQuestions([]);
    }
  }, []);

  useEffect(() => {
    if (selectedModuleId !== '') {
      fetchQuestions(selectedModuleId);
    } else {
      setQuestions([]);
    }
  }, [selectedModuleId, fetchQuestions]);

  // ── Create question ───────────────────────────────────────────
  const handleCreateQuestion = async () => {
    if (
      !qContent.trim() ||
      !qAnalysis.trim() ||
      selectedModuleId === '' ||
      options.some((o) => !o.content.trim()) ||
      options.filter((o) => o.is_correct).length !== 1
    )
      return;

    setQSubmitting(true);
    setQError(null);
    setQSuccess(null);
    try {
      await createQuestion({
        module_id: selectedModuleId,
        content: qContent.trim(),
        analysis: qAnalysis.trim(),
        options: options.map((o, i) => ({ ...o, sort_order: i + 1 })),
      });
      setQContent('');
      setQAnalysis('');
      setOptions([
        { content: '', is_correct: true },
        { content: '', is_correct: false },
        { content: '', is_correct: false },
        { content: '', is_correct: false },
      ]);
      setQSuccess('Question created successfully.');
      await fetchQuestions(selectedModuleId);
    } catch (err) {
      setQError(err instanceof Error ? err.message : 'Failed to create question');
    } finally {
      setQSubmitting(false);
    }
  };

  const handleOptionChange = (idx: number, field: 'content' | 'is_correct', value: string | boolean) => {
    setOptions((prev) => {
      const next = prev.map((o, i) => {
        if (field === 'is_correct' && value === true) {
          return { ...o, is_correct: i === idx };
        }
        if (i !== idx) return o;
        return { ...o, [field]: value };
      });
      return next;
    });
  };

  // ── Delete question ───────────────────────────────────────────
  const handleDeleteQuestion = async (id: number) => {
    setDeletingId(id);
    try {
      await deleteQuestion(id);
      setQuestions((prev) => prev.filter((q) => q.id !== id));
    } catch {
      // silent — question remains in list
    } finally {
      setDeletingId(null);
    }
  };

  // ── Render ────────────────────────────────────────────────────
  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900">Data Entry</h1>
      <p className="mt-1 text-sm text-gray-500">
        Add questions directly under a module.
      </p>

      {/* ── Module selector ─────────────────────────────────── */}
      <div className="mt-6 rounded-xl border bg-white p-5 shadow-sm">
        <label className="block text-sm font-medium text-gray-700">
          Select a module
        </label>
        {loadingModules ? (
          <p className="mt-2 text-sm text-gray-400">Loading modules...</p>
        ) : (
          <select
            value={selectedModuleId}
            onChange={(e) => {
              const v = e.target.value;
              setSelectedModuleId(v ? Number(v) : '');
            }}
            className="mt-2 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm
                       focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
          >
            <option value="">-- Choose a module --</option>
            {modules.map((m) => (
              <option key={m.id} value={m.id}>
                {m.name}
              </option>
            ))}
          </select>
        )}

        {selectedModuleId === '' && modules.length === 0 && !loadingModules && (
          <p className="mt-2 text-sm text-gray-400">
            No modules yet.{' '}
            <a href="/admin/modules" className="text-brand-600 hover:underline">
              Create one first →
            </a>
          </p>
        )}
      </div>

      {/* ── Question form ────────────────────────────────────── */}
      {selectedModuleId !== '' && (
        <div className="mt-4 rounded-xl border bg-white p-5 shadow-sm">
          <h2 className="font-semibold text-gray-900">Add Question</h2>

          <div className="mt-3 space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-600">Question content</label>
              <textarea
                value={qContent}
                onChange={(e) => setQContent(e.target.value)}
                rows={3}
                placeholder="Enter the question text..."
                className="mt-1 w-full rounded-lg border border-gray-200 p-3 text-sm
                           focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600">Analysis / Explanation</label>
              <textarea
                value={qAnalysis}
                onChange={(e) => setQAnalysis(e.target.value)}
                rows={3}
                placeholder="Explain the correct answer..."
                className="mt-1 w-full rounded-lg border border-gray-200 p-3 text-sm
                           focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
              />
            </div>

            {/* Options — fixed at 4 (1 correct + 3 incorrect) */}
            <div>
              <span className="text-xs font-medium text-gray-600">
                Options (1 correct + 3 incorrect)
              </span>
              <div className="mt-2 space-y-2">
                {options.map((opt, idx) => (
                  <div key={idx} className="flex items-center gap-2">
                    {/* Radio for correct answer */}
                    <label className="flex shrink-0 items-center gap-1 text-xs text-gray-500">
                      <input
                        type="radio"
                        name="correct-option"
                        checked={opt.is_correct}
                        onChange={() => handleOptionChange(idx, 'is_correct', true)}
                        className="text-brand-600 focus:ring-brand-500"
                      />
                      Correct
                    </label>
                    <input
                      type="text"
                      value={opt.content}
                      onChange={(e) => handleOptionChange(idx, 'content', e.target.value)}
                      placeholder={`Option ${idx + 1}`}
                      className="flex-1 rounded-lg border border-gray-200 px-3 py-2 text-sm
                                 focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
                    />
                  </div>
                ))}
              </div>
            </div>
          </div>

          <button
            onClick={handleCreateQuestion}
            disabled={
              qSubmitting ||
              !qContent.trim() ||
              !qAnalysis.trim() ||
              options.some((o) => !o.content.trim())
            }
            className="mt-4 rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white
                       hover:bg-brand-700 disabled:opacity-50 transition-colors"
          >
            {qSubmitting ? 'Creating...' : 'Create Question'}
          </button>

          {qError && (
            <p className="mt-3 rounded-lg bg-red-50 p-2 text-sm text-red-600">{qError}</p>
          )}
          {qSuccess && (
            <p className="mt-3 rounded-lg bg-green-50 p-2 text-sm text-green-600">{qSuccess}</p>
          )}
        </div>
      )}

      {/* ── Existing questions list ──────────────────────────── */}
      {questions.length > 0 && (
        <div className="mt-8">
          <h2 className="text-lg font-bold text-gray-900">
            Existing Questions ({questions.length})
          </h2>
          <div className="mt-3 space-y-3">
            {questions.map((q) => (
              <div key={q.id} className="rounded-xl border bg-white p-4 shadow-sm">
                <div className="flex items-start justify-between">
                  <p className="text-sm font-medium text-gray-900">
                    v{q.version}: {q.content}
                  </p>
                  <button
                    onClick={() => handleDeleteQuestion(q.id)}
                    disabled={deletingId === q.id}
                    className="shrink-0 ml-4 rounded-lg px-3 py-1 text-xs font-medium text-red-600
                               hover:bg-red-50 disabled:opacity-50 transition-colors"
                  >
                    {deletingId === q.id ? 'Deleting...' : 'Delete'}
                  </button>
                </div>
                <p className="mt-1 text-xs text-gray-500">{q.analysis}</p>
                <div className="mt-2 flex flex-wrap gap-2">
                  {q.options.map((o) => (
                    <span
                      key={o.id}
                      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                        o.is_correct
                          ? 'bg-green-100 text-green-700'
                          : 'bg-gray-100 text-gray-600'
                      }`}
                    >
                      {o.is_correct && '✓ '}
                      {o.content}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
