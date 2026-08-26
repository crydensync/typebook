import { useState, useEffect } from "react";
import { api } from "../api";

export default function ConfirmEmailView({ token, onDone }) {
  const [status, setStatus] = useState("confirming");
  const [error, setError] = useState("");

  useEffect(() => {
    api
      .confirmEmailChange(token)
      .then(() => setStatus("done"))
      .catch((err) => {
        setError(err.message);
        setStatus("error");
      });
  }, [token]);

  return (
    <div className="auth-screen">
      <h2>Email confirmation</h2>
      {status === "confirming" && <p className="subtitle">Confirming your new email...</p>}
      {status === "done" && <p className="subtitle">Your email has been updated. You can close this page.</p>}
      {status === "error" && <div className="error-msg">{error}</div>}
      <button className="btn btn-text" onClick={onDone}>
        Back to typebook
      </button>
    </div>
  );
}
