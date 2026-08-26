import { useState, useEffect } from "react";
import { api } from "../api";

export default function SettingsView({ onLoggedOut, oauthNotice }) {
  const [sessions, setSessions] = useState([]);
  const [sessionsError, setSessionsError] = useState("");

  const [currentPw, setCurrentPw] = useState("");
  const [newPw, setNewPw] = useState("");
  const [pwMsg, setPwMsg] = useState("");
  const [pwError, setPwError] = useState("");

  const [newEmail, setNewEmail] = useState("");
  const [emailMsg, setEmailMsg] = useState("");
  const [emailError, setEmailError] = useState("");

  const [deletePw, setDeletePw] = useState("");
  const [deleteError, setDeleteError] = useState("");
  const [confirmingDelete, setConfirmingDelete] = useState(false);

  useEffect(() => {
    loadSessions();
  }, []);

  async function loadSessions() {
    try {
      setSessions(await api.listSessions());
    } catch (err) {
      setSessionsError(err.message);
    }
  }

  async function handleRevoke(id) {
    try {
      await api.revokeSession(id);
      setSessions(sessions.filter((s) => s.id !== id));
    } catch (err) {
      setSessionsError(err.message);
    }
  }

  async function handleLogoutAll() {
    await api.logoutAll();
    onLoggedOut();
  }

  async function handleChangePassword(e) {
    e.preventDefault();
    setPwError("");
    setPwMsg("");
    try {
      await api.changePassword(currentPw, newPw);
      // ChangePassword revokes ALL sessions including this one — the
      // user must log in again. This is CrydenSync's actual behavior,
      // not a frontend choice.
      setPwMsg("Password changed. Logging you out for security...");
      setTimeout(onLoggedOut, 1500);
    } catch (err) {
      setPwError(err.message);
    }
  }

  async function handleRequestEmailChange(e) {
    e.preventDefault();
    setEmailError("");
    setEmailMsg("");
    try {
      await api.requestEmailChange(newEmail);
      setEmailMsg(`Verification link sent to ${newEmail}. Check the backend console log (dev mode) for the link.`);
    } catch (err) {
      setEmailError(err.message);
    }
  }

  async function handleDeleteAccount(e) {
    e.preventDefault();
    setDeleteError("");
    try {
      await api.deleteAccount(deletePw);
      onLoggedOut();
    } catch (err) {
      setDeleteError(err.message);
    }
  }

  return (
    <div className="container">
      {oauthNotice && (
        <div className={oauthNotice.kind === "error" ? "error-msg" : "oauth-notice"}>{oauthNotice.text}</div>
      )}

      <div className="settings-section">
        <h3>Connected accounts</h3>
        <p className="desc">
          Sign in with Google or GitHub too. Connecting doesn't affect your
          password — you can still log in either way.
        </p>
        <div style={{ display: "flex", gap: "8px" }}>
          <button className="btn btn-oauth" onClick={() => api.linkOAuthProvider("google")}>
            Connect Google
          </button>
          <button className="btn btn-oauth" onClick={() => api.linkOAuthProvider("github")}>
            Connect GitHub
          </button>
        </div>
      </div>

      <div className="settings-section">
        <h3>Active sessions</h3>
        <p className="desc">Every device currently logged into your account.</p>
        {sessionsError && <div className="error-msg">{sessionsError}</div>}
        {sessions.length === 0 ? (
          <p className="session-meta">No active sessions.</p>
        ) : (
          sessions.map((s) => (
            <div className="session-row" key={s.id}>
              <div>
                <div>{s.user_agent || "Unknown device"}</div>
                <div className="session-meta">{s.ip} · {s.created_at}</div>
              </div>
              <button className="btn btn-text" onClick={() => handleRevoke(s.id)}>
                Revoke
              </button>
            </div>
          ))
        )}
        <div style={{ marginTop: "14px" }}>
          <button className="btn btn-danger" onClick={handleLogoutAll}>
            Log out of all devices
          </button>
        </div>
      </div>

      <div className="settings-section">
        <h3>Change password</h3>
        <p className="desc">This will log you out of every device, including this one.</p>
        <form onSubmit={handleChangePassword}>
          <div className="field">
            <label>Current password</label>
            <input type="password" required value={currentPw} onChange={(e) => setCurrentPw(e.target.value)} />
          </div>
          <div className="field">
            <label>New password</label>
            <input type="password" required value={newPw} onChange={(e) => setNewPw(e.target.value)} />
          </div>
          {pwError && <div className="error-msg">{pwError}</div>}
          {pwMsg && <div className="session-meta">{pwMsg}</div>}
          <button className="btn btn-primary" type="submit">
            Change password
          </button>
        </form>
      </div>

      <div className="settings-section">
        <h3>Change email</h3>
        <p className="desc">Your email won't change until you confirm it via the verification link.</p>
        <form onSubmit={handleRequestEmailChange}>
          <div className="field">
            <label>New email</label>
            <input type="email" required value={newEmail} onChange={(e) => setNewEmail(e.target.value)} />
          </div>
          {emailError && <div className="error-msg">{emailError}</div>}
          {emailMsg && <div className="session-meta">{emailMsg}</div>}
          <button className="btn btn-primary" type="submit">
            Send verification link
          </button>
        </form>
      </div>

      <div className="settings-section">
        <h3>Delete account</h3>
        <p className="desc">This is permanent. All your notes and sessions will be deleted.</p>
        {!confirmingDelete ? (
          <button className="btn btn-danger" onClick={() => setConfirmingDelete(true)}>
            Delete my account
          </button>
        ) : (
          <form onSubmit={handleDeleteAccount}>
            <div className="field">
              <label>Confirm your password to delete your account</label>
              <input type="password" required value={deletePw} onChange={(e) => setDeletePw(e.target.value)} />
            </div>
            {deleteError && <div className="error-msg">{deleteError}</div>}
            <div style={{ display: "flex", gap: "8px" }}>
              <button className="btn btn-text" type="button" onClick={() => setConfirmingDelete(false)}>
                Cancel
              </button>
              <button className="btn btn-danger" type="submit">
                Permanently delete
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
