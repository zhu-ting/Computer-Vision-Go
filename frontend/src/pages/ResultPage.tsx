import { useLocation, useNavigate } from 'react-router-dom';
import type { ExamResult, QuestionResult } from '../types';

export default function ResultPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const result = (location.state as { result?: ExamResult })?.result;

  if (!result) {
    return (
      <main className="mx-auto max-w-3xl px-4 py-16 text-center">
        <p className="text-gray-500">No result data available.</p>
        <button
          onClick={() => navigate('/')}
          className="mt-4 text-brand-600 hover:underline"
        >
          Back to home
        </button>
      </main>
    );
  }

  const percentage = Math.round(result.score);

  return (
    <main className="mx-auto max-w-3xl px-4 py-8">
      {/* Score card */}
      <div className="rounded-xl border bg-white p-8 text-center shadow-sm">
        <p className="text-sm font-medium text-gray-500">Your Score</p>
        <p
          className={`mt-2 text-5xl font-bold ${
            percentage >= 70 ? 'text-green-600' : percentage >= 40 ? 'text-amber-600' : 'text-red-600'
          }`}
        >
          {percentage}%
        </p>
        <p className="mt-2 text-sm text-gray-500">
          {result.correct_count} / {result.total_questions} correct
        </p>
      </div>

      {/* Questions review */}
      <h2 className="mt-10 text-xl font-bold text-gray-900">Review</h2>
      <div className="mt-4 space-y-4">
        {result.questions.map((q, idx) => (
          <ResultCard key={q.exam_question_id} question={q} index={idx + 1} />
        ))}
      </div>

      <div className="mt-8 text-center">
        <button
          onClick={() => navigate('/')}
          className="text-brand-600 hover:underline"
        >
          Back to home
        </button>
      </div>
    </main>
  );
}

function ResultCard({ question, index }: { question: QuestionResult; index: number }) {
  return (
    <div
      className={`rounded-xl border bg-white p-5 shadow-sm ${
        question.is_correct ? 'border-l-4 border-l-green-500' : 'border-l-4 border-l-red-400'
      }`}
    >
      <div className="flex items-start gap-3">
        <span
          className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-xs font-bold text-white ${
            question.is_correct ? 'bg-green-500' : 'bg-red-400'
          }`}
        >
          {question.is_correct ? '✓' : '✗'}
        </span>
        <div className="flex-1">
          <p className="text-sm font-medium text-gray-900">
            {index}. {question.content}
          </p>

          <div className="mt-3 space-y-1">
            {question.options.map((opt, i) => {
              const isUserPick = opt.id === question.selected_option_id;
              const isCorrect = opt.is_correct;

              let className = 'rounded-md border p-2 text-sm ';
              if (isCorrect) {
                className += 'border-green-300 bg-green-50 text-green-800';
              } else if (isUserPick && !isCorrect) {
                className += 'border-red-300 bg-red-50 text-red-800';
              } else {
                className += 'border-gray-100 text-gray-500';
              }

              return (
                <div key={opt.id} className={className}>
                  <span className="font-medium">{String.fromCharCode(65 + i)}.</span>{' '}
                  {opt.content}
                  {isUserPick && <span className="ml-2 text-xs">← your answer</span>}
                  {isCorrect && <span className="ml-2 text-xs">← correct</span>}
                </div>
              );
            })}
          </div>

          {!question.is_correct && question.selected_option_id === null && (
            <p className="mt-2 text-xs text-gray-400">Not answered</p>
          )}

          {/* Analysis — only revealed on the result page */}
          <details className="mt-3">
            <summary className="cursor-pointer text-sm font-medium text-brand-600 hover:underline">
              Show explanation
            </summary>
            <p className="mt-2 rounded-lg bg-gray-50 p-3 text-sm text-gray-700">
              {question.analysis}
            </p>
          </details>
        </div>
      </div>
    </div>
  );
}
