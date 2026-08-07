import { useState, useEffect, useCallback } from 'react';
import {
  listModules,
  listQuestionGroups,
  createQuestionGroup,
  listQuestions,
  createQuestion,
} from '../api/client';
import type { Module, QuestionGroupSummary, AdminQuestion } from '../types';

const DIFFICULTIES = ['easy', 'medium', 'hard'] as const;

export default function AdminDataEntryPage() {
  // ── Modules ───────────────────────────────────────────────────
  const [modules, setModules] = useState<Module[]>([]);
  const [selectedModuleId, setSelectedModuleId] = useState<number | ''>('');
  const [groups, setGroups] = useState<QuestionGroupSummary[]>([]);
  const [loadingModules, setLoadingModules] = useState(true);

  // ── Question group form ───────────────────────────────────────
  const [groupTitle, setGroupTitle] = useState('');
  const [groupTopic, setGroupTopic] = useState('');
  const [groupDifficulty, setGroupDifficulty] = useState<string>('easy');
  const [groupSubmitting, setGroupSubmitting] = useState(false);
  const [groupError, setGroupError] = useState<string | null>(null);
  const [groupSuccess, setGroupSuccess] = useState<string | null>(null);

  // ── Question form ─────────────────────────────────────────────
  const [selectedGroupId, setSelectedGroupId] = useState<number | ''>('');
  const [questions, setQuestions] = useState<AdminQuestion[]>([]);
  const [qContent, setQContent] = useState('');
  const [qAnalysis, setQAnalysis] = useState('');
  const [options, setOptions] = useState<{ content: string; is_correct: boolean }[]>([
    { content: '', is_correct: true },
    { content: '', is_correct: false },
  ]);
  const [qSubmitting, setQSubmitting] = useState(false);
  const [qError, setQError] = useState<string | null>(null);
  const [qSuccess, setQSuccess] = useState<string | null>(null);

  // ── Load modules ──────────────────────────────────────────────
  const fetchModules = useCallback(async () => {
    setLoadingModules(true);
    try {
      setModules(await listModules());
    } catch {
      // silent
    } finally {
      setLoadingModules(false);
    }
  }, []);

  useEffect(() => {
    fetchModules();
  }, [fetchModules]);

  // ── Load groups when module changes ───────────────────────────
  const fetchGroups = useCallback(async (moduleId: number) => {
    try {
      setGroups(await listQuestionGroups(moduleId));
    } catch {
      setGroups([]);
    }
  }, []);

  useEffect(() => {
    if (selectedModuleId !== '') {
      fetchGroups(selectedModuleId);
    } else {
      setGroups([]);
    }
  }, [selectedModuleId, fetchGroups]);

  // ── Load questions when group changes ─────────────────────────
  const fetchQuestions = useCallback(async (groupId: number) => {
    try {
      setQuestions(await listQuestions(groupId));
    } catch {
      setQuestions([]);
    }
  }, []);

  useEffect(() => {
    if (selectedGroupId !== '') {
      fetchQuestions(selectedGroupId);
    } else {
      setQuestions([]);
    }
  }, [selectedGroupId, fetchQuestions]);

  // ── Create question group ─────────────────────────────────────
  const handleCreateGroup = async () => {
    if (!groupTitle.trim() || !groupTopic.trim() || selectedModuleId === '') return;
    setGroupSubmitting(true);
    setGroupError(null);
    setGroupSuccess(null);
    try {
      const g = await createQuestionGroup({
        module_id: selectedModuleId,
        title: groupTitle.trim(),
        topic: groupTopic.trim(),
        difficulty: groupDifficulty,
      });
      setGroupTitle('');
      setGroupTopic('');
      setGroupDifficulty('easy');
      setGroupSuccess(`Group "${g.title}" created (ID: ${g.id}).`);
      await fetchGroups(selectedModuleId);
      setSelectedGroupId(g.id);
    } catch (err) {
      setGroupError(err instanceof Error ? err.message : 'Failed to create question group');
    } finally {
      setGroupSubmitting(false);
    }
  };

  // ── Create question ───────────────────────────────────────────
  const handleCreateQuestion = async () => {
    if (!qContent.trim() || !qAnalysis.trim() || selectedGroupId === '') return;
    setQSubmitting(true);
    setQError(null);
    setQSuccess(null);
    try {
      await createQuestion({
        group_id: selectedGroupId,
        content: qContent.trim(),
        analysis: qAnalysis.trim(),
        options: options.map((o, i) => ({ ...o, sort_order: i + 1 })),
      });
      setQContent('');
      setQAnalysis('');
      setOptions([
        { content: '', is_correct: true },
        { content: '', is_correct: false },
      ]);
      setQSuccess('Question created successfully.');
      await fetchQuestions(selectedGroupId);
    } catch (err) {
      setQError(err instanceof Error ? err.message : 'Failed to create question');
    } finally {
      setQSubmitting(false);
    }
  };

  const handleOptionChange = (idx: number, field: 'content' | 'is_correct', value: string | boolean) => {
    setOptions((prev) => {
      const next = prev.map((o, i) => {
        if (field === 'is_correct' && value === true) {
          return { ...o, is_correct: i === idx };
        }
        if (i !== idx) return o;
        return { ...o, [field]: value };
      });
      return next;
    });
  };

  const addOption = () => {
    setOptions((prev) => [...prev, { content: '', is_correct: false }]);
  };

  const removeOption = (idx: number) => {
    if (options.length <= 2) return;
    setOptions((prev) => {
      const next = prev.filter((_, i) => i !== idx);
      // If we removed the correct one, make the first one correct
      if (!next.some((o) => o.is_correct)) {
        next[0].is_correct = true;
      }
      return next;
    });
  };

  // ── Render ────────────────────────────────────────────────────
  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900">Data Entry</h1>
      <p className="mt-1 text-sm text-gray-500">
        Create question groups and questions under a module.
      </p>

      {/* ── Module selector ─────────────────────────────────── */}
      <div className="mt-6 rounded-xl border bg-white p-5 shadow-sm">
        <label className="block text-sm font-medium text-gray-700">
          Select a module
        </label>
        {loadingModules ? (
          <p className="mt-2 text-sm text-gray-400">Loading modules...</p>
        ) : (
          <select
            value={selectedModuleId}
            onChange={(e) => {
              const v = e.target.value;
              setSelectedModuleId(v ? Number(v) : '');
              setSelectedGroupId('');
            }}
            className="mt-2 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm
                       focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
          >
            <option value="">-- Choose a module --</option>
            {modules.map((m) => (
              <option key={m.id} value={m.id}>
                {m.name}
              </option>
            ))}
          </select>
        )}

        {selectedModuleId === '' && modules.length === 0 && !loadingModules && (
          <p className="mt-2 text-sm text-gray-400">
            No modules yet.{' '}
            <a href="/admin/modules" className="text-brand-600 hover:underline">
              Create one first →
            </a>
          </p>
        )}
      </div>

      {/* ── Question group form ─────────────────────────────── */}
      {selectedModuleId !== '' && (
        <div className="mt-4 rounded-xl border bg-white p-5 shadow-sm">
          <h2 className="font-semibold text-gray-900">Create Question Group</h2>

          <div className="mt-3 grid gap-3 sm:grid-cols-2">
            <div>
              <label className="block text-xs font-medium text-gray-600">Title</label>
              <input
                type="text"
                value={groupTitle}
                onChange={(e) => setGroupTitle(e.target.value)}
                placeholder="e.g., Backpropagation Basics"
                className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm
                           focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600">Topic</label>
              <input
                type="text"
                value={groupTopic}
                onChange={(e) => setGroupTopic(e.target.value)}
                placeholder="e.g., Backpropagation"
                className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm
                           focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
              />
            </div>
          </div>

          <div className="mt-3">
            <label className="block text-xs font-medium text-gray-600">Difficulty</label>
            <select
              value={groupDifficulty}
              onChange={(e) => setGroupDifficulty(e.target.value)}
              className="mt-1 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm
                         focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
            >
              {DIFFICULTIES.map((d) => (
                <option key={d} value={d}>
                  {d}
                </option>
              ))}
            </select>
          </div>

          <button
            onClick={handleCreateGroup}
            disabled={groupSubmitting || !groupTitle.trim() || !groupTopic.trim()}
            className="mt-4 rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white
                       hover:bg-brand-700 disabled:opacity-50 transition-colors"
          >
            {groupSubmitting ? 'Creating...' : 'Create Group'}
          </button>

          {groupError && (
            <p className="mt-3 rounded-lg bg-red-50 p-2 text-sm text-red-600">{groupError}</p>
          )}
          {groupSuccess && (
            <p className="mt-3 rounded-lg bg-green-50 p-2 text-sm text-green-600">{groupSuccess}</p>
          )}
        </div>
      )}

      {/* ── Group selector + question form ───────────────────── */}
      {groups.length > 0 && (
        <div className="mt-4 rounded-xl border bg-white p-5 shadow-sm">
          <label className="block text-sm font-medium text-gray-700">
            Target question group
          </label>
          <select
            value={selectedGroupId}
            onChange={(e) => {
              const v = e.target.value;
              setSelectedGroupId(v ? Number(v) : '');
            }}
            className="mt-2 w-full rounded-lg border border-gray-200 px-3 py-2 text-sm
                       focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
          >
            <option value="">-- Choose a group --</option>
            {groups.map((g) => (
              <option key={g.id} value={g.id}>
                #{g.id} — {g.title} ({g.difficulty}, {g.question_count} questions)
              </option>
            ))}
          </select>
        </div>
      )}

      {/* ── Question form ────────────────────────────────────── */}
      {selectedGroupId !== '' && (
        <div className="mt-4 rounded-xl border bg-white p-5 shadow-sm">
          <h2 className="font-semibold text-gray-900">Add Question</h2>

          <div className="mt-3 space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-600">Question content</label>
              <textarea
                value={qContent}
                onChange={(e) => setQContent(e.target.value)}
                rows={3}
                placeholder="Enter the question text..."
                className="mt-1 w-full rounded-lg border border-gray-200 p-3 text-sm
                           focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-600">Analysis / Explanation</label>
              <textarea
                value={qAnalysis}
                onChange={(e) => setQAnalysis(e.target.value)}
                rows={3}
                placeholder="Explain the correct answer..."
                className="mt-1 w-full rounded-lg border border-gray-200 p-3 text-sm
                           focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
              />
            </div>

            {/* Options */}
            <div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-medium text-gray-600">
                  Options ({options.length})
                </span>
                <button
                  type="button"
                  onClick={addOption}
                  className="text-xs font-medium text-brand-600 hover:underline"
                >
                  + Add option
                </button>
              </div>
              <div className="mt-2 space-y-2">
                {options.map((opt, idx) => (
                  <div key={idx} className="flex items-center gap-2">
                    {/* Radio for correct answer */}
                    <label className="flex shrink-0 items-center gap-1 text-xs text-gray-500">
                      <input
                        type="radio"
                        name="correct-option"
                        checked={opt.is_correct}
                        onChange={() => handleOptionChange(idx, 'is_correct', true)}
                        className="text-brand-600 focus:ring-brand-500"
                      />
                      Correct
                    </label>
                    <input
                      type="text"
                      value={opt.content}
                      onChange={(e) => handleOptionChange(idx, 'content', e.target.value)}
                      placeholder={`Option ${idx + 1}`}
                      className="flex-1 rounded-lg border border-gray-200 px-3 py-2 text-sm
                                 focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
                    />
                    {options.length > 2 && (
                      <button
                        type="button"
                        onClick={() => removeOption(idx)}
                        className="shrink-0 text-xs text-red-500 hover:underline"
                      >
                        Remove
                      </button>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </div>

          <button
            onClick={handleCreateQuestion}
            disabled={
              qSubmitting ||
              !qContent.trim() ||
              !qAnalysis.trim() ||
              options.some((o) => !o.content.trim())
            }
            className="mt-4 rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white
                       hover:bg-brand-700 disabled:opacity-50 transition-colors"
          >
            {qSubmitting ? 'Creating...' : 'Create Question'}
          </button>

          {qError && (
            <p className="mt-3 rounded-lg bg-red-50 p-2 text-sm text-red-600">{qError}</p>
          )}
          {qSuccess && (
            <p className="mt-3 rounded-lg bg-green-50 p-2 text-sm text-green-600">{qSuccess}</p>
          )}
        </div>
      )}

      {/* ── Existing questions list ──────────────────────────── */}
      {questions.length > 0 && (
        <div className="mt-8">
          <h2 className="text-lg font-bold text-gray-900">
            Existing Questions ({questions.length})
          </h2>
          <div className="mt-3 space-y-3">
            {questions.map((q) => (
              <div key={q.id} className="rounded-xl border bg-white p-4 shadow-sm">
                <div className="flex items-start justify-between">
                  <p className="text-sm font-medium text-gray-900">
                    v{q.version}: {q.content}
                  </p>
                </div>
                <p className="mt-1 text-xs text-gray-500">{q.analysis}</p>
                <div className="mt-2 flex flex-wrap gap-2">
                  {q.options.map((o) => (
                    <span
                      key={o.id}
                      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                        o.is_correct
                          ? 'bg-green-100 text-green-700'
                          : 'bg-gray-100 text-gray-600'
                      }`}
                    >
                      {o.is_correct && '✓ '}
                      {o.content}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
