import { useState, useEffect } from "react";
import { api } from "./api";
import { useTheme } from "./useTheme";
import LoginForm from "./components/LoginForm";
import SignupForm from "./components/SignupForm";
import NotesView from "./components/NotesView";
import SettingsView from "./components/SettingsView";
import ConfirmEmailView from "./components/ConfirmEmailView";

export default function App() {
  const { theme, toggle } = useTheme();
  const [authed, setAuthed] = useState(api.isAuthenticated());
  const [authView, setAuthView] = useState("login"); // "login" | "signup"
  const [page, setPage] = useState("notes"); // "notes" | "settings"

  // Simple query-param check for the email confirmation link —
  // deliberately not pulling in a router library for one route.
  const params = new URLSearchParams(window.location.search);
  const confirmToken = window.location.pathname === "/confirm-email" ? params.get("token") : null;

  function handleLoggedOut() {
    api.logoutLocal();
    setAuthed(false);
    setAuthView("login");
    setPage("notes");
  }

  if (confirmToken) {
    return (
      <ConfirmEmailView
        token={confirmToken}
        onDone={() => {
          window.history.pushState({}, "", "/");
          window.location.reload();
        }}
      />
    );
  }

  if (!authed) {
    return authView === "login" ? (
      <LoginForm onLoggedIn={() => setAuthed(true)} onSwitchToSignup={() => setAuthView("signup")} />
    ) : (
      <SignupForm onSignedUp={() => setAuthed(true)} onSwitchToLogin={() => setAuthView("login")} />
    );
  }

  return (
    <>
      <header className="app-header">
        <h1>typebook</h1>
        <div className="header-actions">
          <button className="icon-btn" onClick={() => setPage(page === "notes" ? "settings" : "notes")} title="Settings">
            {page === "notes" ? "⚙" : "📝"}
          </button>
          <button className="icon-btn" onClick={toggle} title="Toggle theme">
            {theme === "light" ? "🌙" : "☀"}
          </button>
          <button className="icon-btn" onClick={handleLoggedOut} title="Log out">
            ⎋
          </button>
        </div>
      </header>
      {page === "notes" ? <NotesView /> : <SettingsView onLoggedOut={handleLoggedOut} />}
    </>
  );
}
