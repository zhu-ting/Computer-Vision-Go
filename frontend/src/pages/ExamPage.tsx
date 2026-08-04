import { useState, useCallback, useEffect } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import { saveAnswers, submitExam } from '../api/client';
import QuestionCard from '../components/QuestionCard';
import Pagination from '../components/Pagination';
import type { Exam, AnswerInput } from '../types';

const PAGE_SIZE = 10;

export default function ExamPage() {
  const { examId } = useParams<{ examId: string }>();
  const location = useLocation();
  const navigate = useNavigate();

  // Exam data is passed via navigation state from HomePage.
  // On direct page load (refresh), there's no state — the user
  // would need a GET /api/v1/exams/:id endpoint to rehydrate.
  const exam = (location.state as { exam?: Exam })?.exam;

  const [answers, setAnswers] = useState<Record<number, number | null>>({});
  const [currentPage, setCurrentPage] = useState(1);
  const [saving, setSaving] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // If exam data is missing (e.g., user refreshed the page),
  // redirect back to home.
  useEffect(() => {
    if (!exam) {
      navigate('/', { replace: true });
    }
  }, [exam, navigate]);

  if (!exam) return null;

  const totalPages = Math.ceil(exam.total_questions / PAGE_SIZE);
  const startIdx = (currentPage - 1) * PAGE_SIZE;
  const pageQuestions = exam.questions.slice(startIdx, startIdx + PAGE_SIZE);

  // ── Auto-save on page flip ───────────────────────────────────

  const flushPage = useCallback(
    async (page: number) => {
      const pageStart = (page - 1) * PAGE_SIZE;
      const pageEqs = exam.questions.slice(pageStart, pageStart + PAGE_SIZE);

      const inputs: AnswerInput[] = pageEqs
        .filter((q) => answers[q.exam_question_id] !== undefined)
        .map((q) => ({
          exam_question_id: q.exam_question_id,
          selected_option_id: answers[q.exam_question_id],
        }));

      if (inputs.length === 0) return;

      setSaving(true);
      try {
        await saveAnswers(Number(examId), { answers: inputs });
      } catch {
        // Auto-save failures are silent — the user can still submit.
        // A production app would queue failed saves and retry.
      } finally {
        setSaving(false);
      }
    },
    [examId, exam.questions, answers],
  );

  const handlePageChange = useCallback(
    async (page: number) => {
      await flushPage(currentPage);
      setCurrentPage(page);
      setError(null);
    },
    [currentPage, flushPage],
  );

  // ── Option selection ─────────────────────────────────────────

  const handleOptionChange = (examQuestionId: number, optionId: number) => {
    setAnswers((prev) => ({ ...prev, [examQuestionId]: optionId }));
  };

  // ── Submit ───────────────────────────────────────────────────

  const handleSubmit = async () => {
    if (submitting) return;

    // Flush the current page first so the backend has all answers.
    await flushPage(currentPage);

    setSubmitting(true);
    setError(null);
    try {
      const result = await submitExam(Number(examId));
      navigate(`/result/${examId}`, { state: { result } });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit exam');
      setSubmitting(false);
    }
  };

  const isLastPage = currentPage === totalPages;

  return (
    <main className="mx-auto max-w-3xl px-4 py-8">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-xl font-bold text-gray-900">
          Exam #{examId}
        </h1>
        <span className="text-sm text-gray-500">
          {saving ? 'Saving...' : 'Saved'}
        </span>
      </div>

      {/* Progress bar */}
      <div className="mb-8 h-1.5 rounded-full bg-gray-100">
        <div
          className="h-full rounded-full bg-brand-500 transition-all"
          style={{
            width: `${((currentPage - 1) / totalPages) * 100}%`,
          }}
        />
      </div>

      {error && (
        <p className="mb-4 rounded-lg bg-red-50 p-3 text-sm text-red-600">{error}</p>
      )}

      {/* Questions for current page */}
      <div className="space-y-6">
        {pageQuestions.map((q, idx) => (
          <QuestionCard
            key={q.exam_question_id}
            question={q}
            selectedOptionId={answers[q.exam_question_id] ?? null}
            onChange={handleOptionChange}
            questionNumber={startIdx + idx + 1}
          />
        ))}
      </div>

      {/* Pagination + Submit */}
      <div className="mt-8 space-y-4">
        <Pagination
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={handlePageChange}
        />

        {isLastPage && (
          <button
            onClick={handleSubmit}
            disabled={submitting}
            className="w-full rounded-lg bg-green-600 px-6 py-3 font-medium text-white
                       shadow hover:bg-green-700 disabled:opacity-50 transition-colors"
          >
            {submitting ? 'Submitting...' : 'Submit Exam'}
          </button>
        )}
      </div>
    </main>
  );
}
