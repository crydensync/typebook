import { useState } from "react";

export default function NoteEditorModal({ note, onSave, onDelete, onClose }) {
  const [title, setTitle] = useState(note.title);
  const [body, setBody] = useState(note.body);

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <input
          placeholder="Title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          style={{
            width: "100%",
            border: "none",
            outline: "none",
            fontSize: "17px",
            fontWeight: 500,
            marginBottom: "10px",
            background: "transparent",
            color: "var(--text)",
          }}
        />
        <textarea
          placeholder="Note"
          value={body}
          onChange={(e) => setBody(e.target.value)}
          rows={8}
          style={{
            width: "100%",
            border: "none",
            outline: "none",
            resize: "vertical",
            fontSize: "14px",
            background: "transparent",
            color: "var(--text)",
          }}
        />
        <div className="modal-actions">
          <button className="btn btn-danger" onClick={() => onDelete(note.id)}>
            Delete
          </button>
          <button className="btn btn-text" onClick={onClose}>
            Cancel
          </button>
          <button className="btn btn-primary" onClick={() => onSave(note.id, { title, body })}>
            Save
          </button>
        </div>
      </div>
    </div>
  );
}
