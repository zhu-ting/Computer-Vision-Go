import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import useSWR from 'swr';
import { createExam, listModules } from '../api/client';
import type { Module } from '../types';

const QUESTION_COUNTS = [10, 20, 30, 40, 50];
const DEFAULT_MODULE_NAME = 'week7_deep_learning';

export default function HomePage() {
  const [count, setCount] = useState(10);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedModuleId, setSelectedModuleId] = useState<number | null>(null);
  const navigate = useNavigate();

  // SWR caches the modules list and revalidates on focus (stale-while-revalidate).
  // Navigation back to this page is instant — no re-fetch spinner.
  const { data: modules = [] } = useSWR<Module[]>(
    '/modules',
    () => listModules(),
    {
      revalidateOnFocus: true,
      dedupingInterval: 60_000, // don't re-fetch within 1 min of a prior fetch
    },
  );

  // Auto-select the default module when modules first load.
  useEffect(() => {
    if (modules.length === 0) return;
    const defaultMod = modules.find((m) => m.name === DEFAULT_MODULE_NAME);
    if (defaultMod) {
      setSelectedModuleId(defaultMod.id);
    } else if (modules.length > 0 && selectedModuleId === null) {
      setSelectedModuleId(modules[0].id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [modules]);

  const handleStart = async () => {
    setLoading(true);
    setError(null);
    try {
      const exam = await createExam({
        question_count: count,
        ...(selectedModuleId != null ? { module_id: selectedModuleId } : {}),
      });
      navigate(`/exam/${exam.exam_id}`, { state: { exam } });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create exam');
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="mx-auto max-w-2xl px-4 py-8 sm:py-16">
      <h1 className="text-2xl sm:text-4xl font-bold tracking-tight text-brand-800">
        Computer Vision · Exam Review
      </h1>
      <p className="mt-3 sm:mt-4 text-base sm:text-lg text-gray-600">
        Select the number of questions, then start practicing.
        Your progress is saved automatically — submit whenever you're ready.
      </p>

      <div className="mt-8 sm:mt-10 rounded-xl border bg-white p-4 sm:p-6 shadow-sm">
        <label className="block text-sm font-medium text-gray-700">
          Number of questions
        </label>
        <div className="mt-3 flex flex-wrap gap-2">
          {QUESTION_COUNTS.map((n) => (
            <button
              key={n}
              onClick={() => setCount(n)}
              className={`min-h-[44px] min-w-[44px] rounded-lg border px-4 py-2.5 sm:py-2
                          text-base sm:text-sm font-medium transition-colors ${
                count === n
                  ? 'border-brand-600 bg-brand-50 text-brand-700'
                  : 'border-gray-200 text-gray-600 hover:border-gray-300 active:bg-gray-50'
              }`}
            >
              {n}
            </button>
          ))}
        </div>

        {modules.length > 0 && (
          <div className="mt-6">
            <label className="block text-sm font-medium text-gray-700">
              Module
            </label>
            <div className="mt-3 flex flex-wrap gap-2">
              {modules.map((m) => (
                <button
                  key={m.id}
                  onClick={() => setSelectedModuleId(m.id)}
                  className={`min-h-[44px] min-w-[44px] rounded-lg border px-4 py-2.5 sm:py-2
                              text-base sm:text-sm font-medium transition-colors ${
                    selectedModuleId === m.id
                      ? 'border-brand-600 bg-brand-50 text-brand-700'
                      : 'border-gray-200 text-gray-600 hover:border-gray-300 active:bg-gray-50'
                  }`}
                >
                  {m.name}
                </button>
              ))}
            </div>
          </div>
        )}

        {error && (
          <p className="mt-4 text-sm text-red-600">{error}</p>
        )}

        <button
          onClick={handleStart}
          disabled={loading}
          className="mt-6 w-full rounded-lg bg-brand-600 px-6 py-3.5 sm:py-3
                     text-base sm:text-sm font-medium text-white shadow
                     hover:bg-brand-700 disabled:opacity-50 transition-colors"
        >
          {loading ? 'Creating exam...' : 'Start Exam'}
        </button>
      </div>
    </main>
  );
}
