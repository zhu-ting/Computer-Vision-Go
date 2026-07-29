import { useState, useEffect, useCallback } from 'react';
import { listNotes, saveNote, deleteNote, getNote } from '../api/client';
import type { Note } from '../types';

export default function NotesPage() {
  const [notes, setNotes] = useState<Note[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // ── Edit state ───────────────────────────────────────────────
  const [editingGroupId, setEditingGroupId] = useState<number | null>(null);
  const [editContent, setEditContent] = useState('');

  const fetchNotes = useCallback(async () => {
    setLoading(true);
    try {
      setNotes(await listNotes());
    } catch {
      setError('Failed to load notes');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchNotes();
  }, [fetchNotes]);

  const handleSave = async (groupId: number) => {
    if (!editContent.trim()) return;
    try {
      const updated = await saveNote(groupId, { content: editContent.trim() });
      setNotes((prev) => {
        const idx = prev.findIndex((n) => n.group_id === groupId);
        if (idx >= 0) {
          const copy = [...prev];
          copy[idx] = updated;
          return copy;
        }
        return [updated, ...prev];
      });
      setEditingGroupId(null);
    } catch {
      setError('Failed to save note');
    }
  };

  const handleDelete = async (groupId: number) => {
    try {
      await deleteNote(groupId);
      setNotes((prev) => prev.filter((n) => n.group_id !== groupId));
    } catch {
      setError('Failed to delete note');
    }
  };

  const startEditing = async (groupId: number) => {
    // Try to load existing content
    try {
      const existing = await getNote(groupId);
      setEditContent(existing.content);
    } catch {
      setEditContent('');
    }
    setEditingGroupId(groupId);
  };

  if (loading) {
    return (
      <main className="mx-auto max-w-2xl px-4 py-16 text-center">
        <p className="text-gray-500">Loading notes...</p>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-2xl px-4 py-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">My Notes</h1>
        <button
          onClick={() => {
            setEditContent('');
            setEditingGroupId(-1);
          }}
          className="rounded-lg bg-brand-600 px-4 py-2 text-sm font-medium text-white
                     hover:bg-brand-700 transition-colors"
        >
          New Note
        </button>
      </div>

      {error && (
        <p className="mt-4 rounded-lg bg-red-50 p-3 text-sm text-red-600">{error}</p>
      )}

      {/* New note form */}
      {editingGroupId === -1 && (
        <NoteEditor
          groupId={0}
          content={editContent}
          onChange={setEditContent}
          onSave={() => {
            // New notes need a group_id — prompt the user.
            const id = prompt('Enter question group ID:');
            if (id) handleSave(Number(id));
          }}
          onCancel={() => setEditingGroupId(null)}
          isNew
        />
      )}

      {notes.length === 0 && editingGroupId !== -1 && (
        <p className="mt-8 text-center text-gray-400">
          No notes yet. Notes are attached to question groups — add one from the
          exam review screen.
        </p>
      )}

      <div className="mt-6 space-y-4">
        {notes.map((note) => (
          <div key={note.id} className="rounded-xl border bg-white p-5 shadow-sm">
            {editingGroupId === note.group_id ? (
              <NoteEditor
                groupId={note.group_id}
                content={editContent}
                onChange={setEditContent}
                onSave={() => handleSave(note.group_id)}
                onCancel={() => setEditingGroupId(null)}
              />
            ) : (
              <>
                <div className="flex items-start justify-between">
                  <span className="text-xs font-medium text-gray-400">
                    Group #{note.group_id}
                  </span>
                  <span className="text-xs text-gray-300">
                    {new Date(note.updated_at).toLocaleDateString()}
                  </span>
                </div>
                <p className="mt-2 whitespace-pre-wrap text-sm text-gray-700">
                  {note.content}
                </p>
                <div className="mt-3 flex gap-2">
                  <button
                    onClick={() => startEditing(note.group_id)}
                    className="text-xs font-medium text-brand-600 hover:underline"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => handleDelete(note.group_id)}
                    className="text-xs font-medium text-red-500 hover:underline"
                  >
                    Delete
                  </button>
                </div>
              </>
            )}
          </div>
        ))}
      </div>
    </main>
  );
}

// ── Inline editor sub-component ────────────────────────────────

function NoteEditor({
  groupId,
  content,
  onChange,
  onSave,
  onCancel,
  isNew = false,
}: {
  groupId: number;
  content: string;
  onChange: (v: string) => void;
  onSave: () => void;
  onCancel: () => void;
  isNew?: boolean;
}) {
  return (
    <div className="space-y-3">
      {isNew && (
        <p className="text-xs text-gray-400">
          Click save, then enter the question group ID in the prompt.
        </p>
      )}
      {!isNew && (
        <p className="text-xs font-medium text-gray-400">Group #{groupId}</p>
      )}
      <textarea
        value={content}
        onChange={(e) => onChange(e.target.value)}
        rows={4}
        className="w-full rounded-lg border border-gray-200 p-3 text-sm
                   focus:border-brand-400 focus:outline-none focus:ring-1 focus:ring-brand-400"
        placeholder="Write your note..."
      />
      <div className="flex gap-2">
        <button
          onClick={onSave}
          disabled={!content.trim()}
          className="rounded-lg bg-brand-600 px-4 py-1.5 text-sm font-medium text-white
                     hover:bg-brand-700 disabled:opacity-50 transition-colors"
        >
          Save
        </button>
        <button
          onClick={onCancel}
          className="rounded-lg border border-gray-200 px-4 py-1.5 text-sm text-gray-600
                     hover:bg-gray-50 transition-colors"
        >
          Cancel
        </button>
      </div>
    </div>
  );
}
