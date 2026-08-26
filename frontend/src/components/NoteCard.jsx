export default function NoteCard({ note, onOpen }) {
  return (
    <div className="note-card" onClick={() => onOpen(note)}>
      {note.title && <h3>{note.title}</h3>}
      {note.body && <p>{note.body}</p>}
    </div>
  );
}
