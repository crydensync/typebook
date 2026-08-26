import { useState, useEffect } from "react";
import { api } from "../api";
import NoteCard from "./NoteCard";
import NoteEditorModal from "./NoteEditorModal";

export default function NotesView() {
  const [notes, setNotes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [composerTitle, setComposerTitle] = useState("");
  const [composerBody, setComposerBody] = useState("");
  const [editing, setEditing] = useState(null);

  useEffect(() => {
    loadNotes();
  }, []);

  async function loadNotes() {
    setLoading(true);
    try {
      const data = await api.listNotes();
      setNotes(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function handleCreate(e) {
    e.preventDefault();
    if (!composerTitle && !composerBody) return;
    try {
      const created = await api.createNote({ title: composerTitle, body: composerBody, color: "default" });
      setNotes([created, ...notes]);
      setComposerTitle("");
      setComposerBody("");
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleSave(id, updates) {
    try {
      await api.updateNote(id, updates);
      setNotes(notes.map((n) => (n.id === id ? { ...n, ...updates } : n)));
      setEditing(null);
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleDelete(id) {
    try {
      await api.deleteNote(id);
      setNotes(notes.filter((n) => n.id !== id));
      setEditing(null);
    } catch (err) {
      setError(err.message);
    }
  }

  return (
    <div className="container">
      <form className="note-composer" onSubmit={handleCreate}>
        <input
          placeholder="Title"
          value={composerTitle}
          onChange={(e) => setComposerTitle(e.target.value)}
        />
        <textarea
          placeholder="Take a note..."
          rows={2}
          value={composerBody}
          onChange={(e) => setComposerBody(e.target.value)}
        />
        <div className="composer-actions">
          <button type="submit" className="btn btn-primary">
            Add
          </button>
        </div>
      </form>

      {error && <div className="error-msg">{error}</div>}

      {loading ? (
        <div className="empty-state">Loading...</div>
      ) : notes.length === 0 ? (
        <div className="empty-state">No notes yet — add your first one above.</div>
      ) : (
        <div className="note-grid">
          {notes.map((note) => (
            <NoteCard key={note.id} note={note} onOpen={setEditing} />
          ))}
        </div>
      )}

      {editing && (
        <NoteEditorModal
          note={editing}
          onSave={handleSave}
          onDelete={handleDelete}
          onClose={() => setEditing(null)}
        />
      )}
    </div>
  );
}
