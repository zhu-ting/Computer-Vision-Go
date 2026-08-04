// ── Exam creation (no correct answers exposed) ──────────────

export interface Exam {
  exam_id: number;
  total_questions: number;
  questions: ExamQuestion[];
}

export interface ExamQuestion {
  id: number;
  exam_question_id: number;
  group_id: number;
  content: string;
  options: ExamOption[];
}

export interface ExamOption {
  id: number;
  content: string;
  // is_correct is NOT present during an active exam
}

// ── Progress saving ─────────────────────────────────────────

export interface AnswerInput {
  exam_question_id: number;
  selected_option_id: number | null;
}

// ── Submission / result (correct answers revealed) ──────────

export interface ExamResult {
  exam_id: number;
  total_questions: number;
  correct_count: number;
  score: number;
  status: string;
  questions: QuestionResult[];
}

export interface QuestionResult {
  id: number;
  exam_question_id: number;
  group_id: number;
  content: string;
  analysis: string;
  selected_option_id: number | null;
  is_correct: boolean;
  options: OptionResult[];
}

export interface OptionResult {
  id: number;
  content: string;
  is_correct: boolean;
}

// ── Notes ───────────────────────────────────────────────────

export interface Note {
  id: number;
  group_id: number;
  content: string;
  created_at: string;
  updated_at: string;
}

// ── API request bodies ──────────────────────────────────────

export interface CreateExamRequest {
  question_count: number;
}

export interface SaveAnswersRequest {
  answers: AnswerInput[];
}

export interface SaveNoteRequest {
  content: string;
}
