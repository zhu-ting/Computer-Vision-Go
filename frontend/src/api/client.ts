import type {
  Exam,
  ExamResult,
  Note,
  Module,
  QuestionGroupSummary,
  AdminQuestion,
  CreateExamRequest,
  SaveAnswersRequest,
  SaveNoteRequest,
  CreateModuleRequest,
  CreateQuestionGroupRequest,
  CreateQuestionRequest,
} from '../types';

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';

// Thin wrapper around fetch. For a larger app, consider react-query or swr
// to handle caching, refetching, and optimistic updates.

async function request<T>(
  path: string,
  options?: RequestInit,
): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(res.status, body.error ?? res.statusText);
  }

  return res.json();
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

// ── Exams ────────────────────────────────────────────────────

export function createExam(body: CreateExamRequest): Promise<Exam> {
  return request<Exam>('/exams', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

export function saveAnswers(
  examId: number,
  body: SaveAnswersRequest,
): Promise<{ status: string }> {
  return request(`/exams/${examId}/answers`, {
    method: 'PATCH',
    body: JSON.stringify(body),
  });
}

export function submitExam(examId: number): Promise<ExamResult> {
  return request<ExamResult>(`/exams/${examId}/submit`, {
    method: 'POST',
  });
}

// ── Notes ────────────────────────────────────────────────────

export function listNotes(): Promise<Note[]> {
  return request<Note[]>('/notes');
}

export function getNote(groupId: number): Promise<Note> {
  return request<Note>(`/notes/${groupId}`);
}

export function saveNote(
  groupId: number,
  body: SaveNoteRequest,
): Promise<Note> {
  return request<Note>(`/notes/${groupId}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  });
}

export function deleteNote(groupId: number): Promise<{ status: string }> {
  return request(`/notes/${groupId}`, { method: 'DELETE' });
}

// ── Modules (themes) ────────────────────────────────────────────

export function listModules(): Promise<Module[]> {
  return request<Module[]>('/modules');
}

export function createModule(body: CreateModuleRequest): Promise<Module> {
  return request<Module>('/modules', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

// ── Question groups (catalog) ───────────────────────────────────

export function listQuestionGroups(moduleId?: number): Promise<QuestionGroupSummary[]> {
  const path = moduleId != null ? `/question-groups?module_id=${moduleId}` : '/question-groups';
  return request<QuestionGroupSummary[]>(path);
}

export function createQuestionGroup(
  body: CreateQuestionGroupRequest,
): Promise<QuestionGroupSummary> {
  return request<QuestionGroupSummary>('/question-groups', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}

// ── Questions (admin CRUD) ──────────────────────────────────────

export function listQuestions(groupId?: number): Promise<AdminQuestion[]> {
  const path = groupId != null ? `/questions?group_id=${groupId}` : '/questions';
  return request<AdminQuestion[]>(path);
}

export function createQuestion(body: CreateQuestionRequest): Promise<AdminQuestion> {
  return request<AdminQuestion>('/questions', {
    method: 'POST',
    body: JSON.stringify(body),
  });
}
