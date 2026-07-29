import type {
  Exam,
  ExamResult,
  Note,
  CreateExamRequest,
  SaveAnswersRequest,
  SaveNoteRequest,
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
