import { useState, useEffect } from "react";
import { api } from "./api";
import { useTheme } from "./useTheme";
import LandingView from "./components/LandingView";
import LoginForm from "./components/LoginForm";
import SignupForm from "./components/SignupForm";
import NotesView from "./components/NotesView";
import SettingsView from "./components/SettingsView";
import ConfirmEmailView from "./components/ConfirmEmailView";

export default function App() {
  const { theme, toggle } = useTheme();
  const [authed, setAuthed] = useState(api.isAuthenticated());
  // "landing" | "login" | "signup" — landing is the entry point for anyone
  // not already authenticated, so first-time visitors see what the app is
  // before being dropped straight into a login form.
  const [authView, setAuthView] = useState("landing");
  const [page, setPage] = useState("notes"); // "notes" | "settings"

  // Simple query-param check for the email confirmation link —
  // deliberately not pulling in a router library for one route.
  const params = new URLSearchParams(window.location.search);
  const confirmToken = window.location.pathname === "/confirm-email" ? params.get("token") : null;

  function handleLoggedOut() {
    api.logoutLocal();
    setAuthed(false);
    setAuthView("landing");
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
    if (authView === "landing") {
      return (
        <LandingView
          onGetStarted={() => setAuthView("signup")}
          onLogin={() => setAuthView("login")}
        />
      );
    }
    return (
      <div className="auth-screen-wrap">
        <button className="btn btn-text auth-back" onClick={() => setAuthView("landing")}>
          ← Back
        </button>
        {authView === "login" ? (
          <LoginForm onLoggedIn={() => setAuthed(true)} onSwitchToSignup={() => setAuthView("signup")} />
        ) : (
          <SignupForm onSignedUp={() => setAuthed(true)} onSwitchToLogin={() => setAuthView("login")} />
        )}
      </div>
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
