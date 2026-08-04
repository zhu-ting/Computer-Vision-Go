import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { createExam } from '../api/client';

const QUESTION_COUNTS = [10, 20, 30, 40, 50];

export default function HomePage() {
  const [count, setCount] = useState(20);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const handleStart = async () => {
    setLoading(true);
    setError(null);
    try {
      const exam = await createExam({ question_count: count });
      // Pass exam data to the exam page via navigation state.
      // On page refresh the state is lost — a production app would
      // have a GET /api/v1/exams/:id endpoint to rehydrate.
      navigate(`/exam/${exam.exam_id}`, { state: { exam } });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create exam');
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="mx-auto max-w-2xl px-4 py-16">
      <h1 className="text-4xl font-bold tracking-tight text-brand-800">
        Computer Vision · Exam Review
      </h1>
      <p className="mt-4 text-lg text-gray-600">
        Select the number of questions, then start practicing.
        Your progress is saved automatically — submit whenever you're ready.
      </p>

      <div className="mt-10 rounded-xl border bg-white p-6 shadow-sm">
        <label className="block text-sm font-medium text-gray-700">
          Number of questions
        </label>
        <div className="mt-3 flex flex-wrap gap-2">
          {QUESTION_COUNTS.map((n) => (
            <button
              key={n}
              onClick={() => setCount(n)}
              className={`rounded-lg border px-4 py-2 text-sm font-medium transition-colors ${
                count === n
                  ? 'border-brand-600 bg-brand-50 text-brand-700'
                  : 'border-gray-200 text-gray-600 hover:border-gray-300'
              }`}
            >
              {n}
            </button>
          ))}
        </div>

        {error && (
          <p className="mt-4 text-sm text-red-600">{error}</p>
        )}

        <button
          onClick={handleStart}
          disabled={loading}
          className="mt-6 w-full rounded-lg bg-brand-600 px-6 py-3 font-medium text-white
                     shadow hover:bg-brand-700 disabled:opacity-50 transition-colors"
        >
          {loading ? 'Creating exam...' : 'Start Exam'}
        </button>
      </div>

      <footer className="mt-8 text-center">
        <a href="/notes" className="text-sm text-brand-600 hover:underline">
          View my notes →
        </a>
      </footer>
    </main>
  );
}
