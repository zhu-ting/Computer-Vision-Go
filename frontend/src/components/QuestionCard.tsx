import type { ExamQuestion } from '../types';

interface Props {
  question: ExamQuestion;
  selectedOptionId: number | null;
  onChange: (examQuestionId: number, optionId: number) => void;
  questionNumber: number;
}

export default function QuestionCard({
  question,
  selectedOptionId,
  onChange,
  questionNumber,
}: Props) {
  return (
    <div className="rounded-xl border bg-white p-4 sm:p-6 shadow-sm">
      <div className="flex items-start gap-2 sm:gap-3">
        <span className="flex h-6 w-6 sm:h-7 sm:w-7 shrink-0 items-center justify-center
                         rounded-full bg-brand-100 text-xs sm:text-sm font-semibold text-brand-700">
          {questionNumber}
        </span>
        <p className="text-base sm:text-lg font-medium text-gray-900">{question.content}</p>
      </div>

      <div className="mt-4 sm:mt-5 space-y-2">
        {question.options.map((opt, idx) => {
          const isSelected = selectedOptionId === opt.id;
          return (
            <label
              key={opt.id}
              className={`flex cursor-pointer items-center gap-2 sm:gap-3 rounded-lg border
                          py-3 px-3 sm:p-3 transition-colors min-h-[48px] sm:min-h-0 ${
                            isSelected
                              ? 'border-brand-500 bg-brand-50'
                              : 'border-gray-200 hover:border-gray-300 active:bg-gray-50'
                          }`}
            >
              <input
                type="radio"
                name={`question-${question.exam_question_id}`}
                value={opt.id}
                checked={isSelected}
                onChange={() => onChange(question.exam_question_id, opt.id)}
                className="h-5 w-5 sm:h-4 sm:w-4 text-brand-600 focus:ring-brand-500 shrink-0"
              />
              <span className="text-xs sm:text-sm font-medium text-gray-500 shrink-0">
                {String.fromCharCode(65 + idx)}.
              </span>
              <span className="text-sm text-gray-800 leading-snug">{opt.content}</span>
            </label>
          );
        })}
      </div>
    </div>
  );
}
