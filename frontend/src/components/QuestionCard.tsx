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
    <div className="rounded-xl border bg-white p-6 shadow-sm">
      <div className="flex items-start gap-3">
        <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full
                         bg-brand-100 text-sm font-semibold text-brand-700">
          {questionNumber}
        </span>
        <p className="text-lg font-medium text-gray-900">{question.content}</p>
      </div>

      <div className="mt-5 space-y-2">
        {question.options.map((opt, idx) => {
          const isSelected = selectedOptionId === opt.id;
          return (
            <label
              key={opt.id}
              className={`flex cursor-pointer items-center gap-3 rounded-lg border p-3
                          transition-colors ${
                            isSelected
                              ? 'border-brand-500 bg-brand-50'
                              : 'border-gray-200 hover:border-gray-300'
                          }`}
            >
              <input
                type="radio"
                name={`question-${question.exam_question_id}`}
                value={opt.id}
                checked={isSelected}
                onChange={() => onChange(question.exam_question_id, opt.id)}
                className="h-4 w-4 text-brand-600 focus:ring-brand-500"
              />
              <span className="text-sm font-medium text-gray-500">
                {String.fromCharCode(65 + idx)}
              </span>
              <span className="text-sm text-gray-800">{opt.content}</span>
            </label>
          );
        })}
      </div>
    </div>
  );
}
