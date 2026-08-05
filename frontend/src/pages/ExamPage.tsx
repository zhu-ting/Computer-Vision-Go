import { useState, useCallback, useEffect, useRef } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import { saveAnswers, submitExam } from '../api/client';
import QuestionCard from '../components/QuestionCard';
import Pagination from '../components/Pagination';
import type { Exam, AnswerInput } from '../types';

const PAGE_SIZE = 10;
const AUTO_SAVE_INTERVAL_MS = 30_000; // 30 seconds
const SECONDS_PER_QUESTION = 120; // 2 minutes per question

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
  const [lastSavedAt, setLastSavedAt] = useState<Date | null>(null);
  const [hasUnsaved, setHasUnsaved] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showSubmitConfirm, setShowSubmitConfirm] = useState(false);

  // Countdown timer (seconds remaining).
  const totalSeconds = (exam?.total_questions ?? 0) * SECONDS_PER_QUESTION;
  const [timeLeft, setTimeLeft] = useState(totalSeconds);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const autoSaveRef = useRef<ReturnType<typeof setInterval> | null>(null);
  // Refs to avoid stale closures in interval/timeout callbacks.
  const dirtyRef = useRef(false);
  const answersRef = useRef(answers);
  const currentPageRef = useRef(currentPage);
  const flushPageRef = useRef<(page: number) => Promise<void>>(async () => {});
  const handleSubmitRef = useRef<() => Promise<void>>(async () => {});

  // Keep refs in sync so callbacks always see the latest values.
  answersRef.current = answers;
  currentPageRef.current = currentPage;

  // ── Redirect if exam data is missing ──────────────────────────

  useEffect(() => {
    if (!exam) {
      navigate('/', { replace: true });
    }
  }, [exam, navigate]);

  // ── Countdown timer ───────────────────────────────────────────

  useEffect(() => {
    if (!exam) return;

    timerRef.current = setInterval(() => {
      setTimeLeft((prev) => {
        if (prev <= 1) {
          // Time's up — auto-submit via the ref to avoid stale closures.
          // Use setTimeout to avoid state-update-inside-render warning.
          setTimeout(() => handleSubmitRef.current(), 0);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [exam]);

  // ── Interval auto-save ────────────────────────────────────────

  const flushPage = useCallback(
    async (page: number) => {
      const pageStart = (page - 1) * PAGE_SIZE;
      const pageEqs = exam!.questions.slice(pageStart, pageStart + PAGE_SIZE);

      const inputs: AnswerInput[] = pageEqs
        .filter((q) => answersRef.current[q.exam_question_id] !== undefined)
        .map((q) => ({
          exam_question_id: q.exam_question_id,
          selected_option_id: answersRef.current[q.exam_question_id],
        }));

      if (inputs.length === 0) return;

      setSaving(true);
      try {
        await saveAnswers(Number(examId), { answers: inputs });
        setLastSavedAt(new Date());
        setHasUnsaved(false);
        dirtyRef.current = false;
      } catch {
        // Auto-save failures are silent — the user can still submit.
        // A production app would queue failed saves and retry.
      } finally {
        setSaving(false);
      }
    },
    [examId, exam],
  );

  // Keep flushPageRef current for use in the interval callback.
  flushPageRef.current = flushPage;

  useEffect(() => {
    if (!exam) return;

    autoSaveRef.current = setInterval(() => {
      if (dirtyRef.current) {
        flushPageRef.current(currentPageRef.current);
      }
    }, AUTO_SAVE_INTERVAL_MS);

    return () => {
      if (autoSaveRef.current) clearInterval(autoSaveRef.current);
    };
    // Only set up once when exam is ready; refs keep the callback fresh.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [exam]);

  // ── Save on beforeunload ──────────────────────────────────────

  useEffect(() => {
    const handleBeforeUnload = () => {
      // Attempt to flush dirty answers synchronously via fetch with keepalive.
      // sendBeacon is POST-only, but our endpoint expects PATCH.
      if (dirtyRef.current && exam) {
        const page = currentPageRef.current;
        const pageStart = (page - 1) * PAGE_SIZE;
        const pageEqs = exam.questions.slice(pageStart, pageStart + PAGE_SIZE);
        const inputs: AnswerInput[] = pageEqs
          .filter((q) => answersRef.current[q.exam_question_id] !== undefined)
          .map((q) => ({
            exam_question_id: q.exam_question_id,
            selected_option_id: answersRef.current[q.exam_question_id],
          }));

        if (inputs.length > 0) {
          const baseUrl = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';
          // fetch with keepalive allows PATCH (unlike sendBeacon which is POST-only).
          fetch(`${baseUrl}/exams/${examId}/answers`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ answers: inputs }),
            keepalive: true,
          }).catch(() => {
            // Silently ignore — best effort on page close.
          });
        }
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [examId, exam]);

  if (!exam) return null;

  const totalPages = Math.ceil(exam.total_questions / PAGE_SIZE);
  const startIdx = (currentPage - 1) * PAGE_SIZE;
  const pageQuestions = exam.questions.slice(startIdx, startIdx + PAGE_SIZE);

  // ── Derived stats ─────────────────────────────────────────────

  const answeredCount = exam.questions.filter(
    (q) => answers[q.exam_question_id] !== undefined,
  ).length;
  const totalQuestions = exam.total_questions;
  const progressPct = totalQuestions > 0 ? (answeredCount / totalQuestions) * 100 : 0;

  // ── Navigation ────────────────────────────────────────────────

  const handlePageChange = useCallback(
    async (page: number) => {
      if (page === currentPage) return;
      await flushPage(currentPage);
      setCurrentPage(page);
      setError(null);
    },
    [currentPage, flushPage],
  );

  // ── Option selection ─────────────────────────────────────────

  const handleOptionChange = (examQuestionId: number, optionId: number) => {
    setAnswers((prev) => ({ ...prev, [examQuestionId]: optionId }));
    setHasUnsaved(true);
    dirtyRef.current = true;
  };

  // ── Keyboard shortcuts ───────────────────────────────────────

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't capture when typing in inputs.
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      // Arrow keys for page navigation.
      if (e.key === 'ArrowLeft' && currentPage > 1) {
        e.preventDefault();
        handlePageChange(currentPage - 1);
      } else if (e.key === 'ArrowRight' && currentPage < totalPages) {
        e.preventDefault();
        handlePageChange(currentPage + 1);
      }

      // A/B/C/D for selecting options on the current page.
      const optionKeys = ['a', 'b', 'c', 'd'];
      const keyIdx = optionKeys.indexOf(e.key.toLowerCase());
      if (keyIdx >= 0 && pageQuestions.length > 0) {
        e.preventDefault();
        // Apply to the first unanswered question on the page, or the first question.
        const target =
          pageQuestions.find((q) => answers[q.exam_question_id] === undefined) ??
          pageQuestions[0];
        if (target && target.options[keyIdx]) {
          handleOptionChange(target.exam_question_id, target.options[keyIdx].id);
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [currentPage, totalPages, pageQuestions, answers, handlePageChange]);

  // ── Submit ───────────────────────────────────────────────────

  const handleSubmitClick = () => {
    const unanswered = totalQuestions - answeredCount;
    if (unanswered > 0) {
      setShowSubmitConfirm(true);
    } else {
      handleSubmit();
    }
  };

  const handleSubmit = async () => {
    if (submitting) return;

    // Flush the current page first.
    await flushPage(currentPage);

    // Stop timers.
    if (timerRef.current) clearInterval(timerRef.current);
    if (autoSaveRef.current) clearInterval(autoSaveRef.current);

    setSubmitting(true);
    setError(null);
    setShowSubmitConfirm(false);
    try {
      const result = await submitExam(Number(examId));
      navigate(`/result/${examId}`, { state: { result } });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit exam');
      setSubmitting(false);

      // Restart timers on failure.
      timerRef.current = setInterval(() => {
        setTimeLeft((prev) => Math.max(0, prev - 1));
      }, 1000);
    }
  };

  // Keep submit ref current for the timer callback.
  handleSubmitRef.current = handleSubmit;

  // ── Timer formatting ─────────────────────────────────────────

  const formatTime = (seconds: number) => {
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;
    return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  };

  const timerUrgent = timeLeft < 300; // less than 5 minutes

  // ── Save status text ──────────────────────────────────────────

  const saveStatus = saving
    ? 'Saving...'
    : hasUnsaved
      ? 'Unsaved changes'
      : lastSavedAt
        ? `Saved at ${lastSavedAt.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
        : 'No answers yet';

  const isLastPage = currentPage === totalPages;
  const unansweredOnSubmit = totalQuestions - answeredCount;

  // Shorter save status for narrow screens.
  const saveStatusShort = saving ? 'Saving...' : hasUnsaved ? 'Unsaved' : 'Saved';

  return (
    <main className="mx-auto max-w-3xl px-4 py-6 sm:py-8">
      {/* ── Header bar ──────────────────────────────────────── */}
      <div className="mb-6 flex flex-wrap items-center justify-between gap-2">
        <h1 className="text-lg sm:text-xl font-bold text-gray-900">Exam #{examId}</h1>

        <div className="flex items-center gap-2 sm:gap-4">
          {/* Timer */}
          <span
            className={`rounded-full px-2 sm:px-3 py-1 text-xs sm:text-sm font-mono font-semibold ${
              timerUrgent
                ? 'bg-red-100 text-red-700 animate-pulse'
                : 'bg-gray-100 text-gray-700'
            }`}
          >
            {formatTime(timeLeft)}
          </span>

          {/* Save status — compact on mobile, verbose on desktop */}
          <span
            className={`text-xs sm:text-sm ${
              saving ? 'text-brand-600' : hasUnsaved ? 'text-amber-600' : 'text-gray-400'
            }`}
            title={saveStatus}
          >
            <span className="hidden sm:inline">{saveStatus}</span>
            <span className="sm:hidden">{saveStatusShort}</span>
          </span>
        </div>
      </div>

      {/* ── Progress ────────────────────────────────────────── */}
      <div className="mb-2 flex items-center justify-between text-xs sm:text-sm text-gray-500">
        <span>
          {answeredCount}/{totalQuestions} answered
        </span>
        <span>{Math.round(progressPct)}%</span>
      </div>

      <div className="mb-8 h-2 rounded-full bg-gray-100">
        <div
          className="h-full rounded-full bg-brand-500 transition-all duration-500"
          style={{ width: `${progressPct}%` }}
        />
      </div>

      {error && (
        <p className="mb-4 rounded-lg bg-red-50 p-3 text-sm text-red-600">{error}</p>
      )}

      {/* ── Questions ───────────────────────────────────────── */}
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

      {/* ── Pagination + Submit ─────────────────────────────── */}
      <div className="mt-8 space-y-4">
        <Pagination
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={handlePageChange}
        />

        {isLastPage && (
          <button
            onClick={handleSubmitClick}
            disabled={submitting}
            className="w-full rounded-lg bg-green-600 px-4 sm:px-6 py-3.5 sm:py-3
                       text-base sm:text-sm font-medium text-white shadow
                       hover:bg-green-700 disabled:opacity-50 transition-colors"
          >
            {submitting ? 'Submitting...' : 'Submit Exam'}
          </button>
        )}
      </div>

      {/* ── Keyboard shortcuts hint (desktop only) ──────────── */}
      <p className="mt-6 hidden sm:block text-center text-xs text-gray-400">
        Keyboard: ← → to navigate pages · A B C D to select options
      </p>

      {/* ── Submit confirmation modal ───────────────────────── */}
      {showSubmitConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="mx-4 w-full max-w-sm rounded-xl bg-white p-6 shadow-xl">
            <h3 className="text-lg font-semibold text-gray-900">Submit exam?</h3>
            <p className="mt-2 text-sm text-gray-600">
              You have{' '}
              <span className="font-semibold text-amber-600">
                {unansweredOnSubmit} unanswered question{unansweredOnSubmit !== 1 ? 's' : ''}
              </span>
              . Unanswered questions will be marked as incorrect.
            </p>
            <div className="mt-6 flex flex-col-reverse sm:flex-row gap-2 sm:gap-3">
              <button
                onClick={() => setShowSubmitConfirm(false)}
                className="flex-1 rounded-lg border border-gray-200 px-4 py-2.5 sm:py-2
                           text-sm font-medium text-gray-600 hover:bg-gray-50 transition-colors"
              >
                Go Back
              </button>
              <button
                onClick={handleSubmit}
                className="flex-1 rounded-lg bg-green-600 px-4 py-2.5 sm:py-2 text-sm
                           font-medium text-white hover:bg-green-700 transition-colors"
              >
                Submit Anyway
              </button>
            </div>
          </div>
        </div>
      )}
    </main>
  );
}
