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
  const [oauthNotice, setOauthNotice] = useState(null); // { kind: "error" | "info", text }

  // The OAuth callback redirects here with the outcome in the URL
  // FRAGMENT (#access_token=...), never a query param — fragments
  // never reach the server or get logged by an intervening proxy.
  // Runs once on mount; a full page load already happened to get here.
  useEffect(() => {
    if (!window.location.hash) return;
    const params = new URLSearchParams(window.location.hash.slice(1));

    const accessToken = params.get("access_token");
    const refreshToken = params.get("refresh_token");
    if (accessToken && refreshToken) {
      api.storeOAuthTokens(accessToken, refreshToken);
      setAuthed(true);
    }

    const linked = params.get("oauth_linked");
    if (linked) {
      setOauthNotice({ kind: "info", text: `${capitalize(linked)} connected.` });
    }

    const error = params.get("oauth_error");
    if (error === "email_conflict") {
      const email = params.get("oauth_email") || "that email";
      setOauthNotice({
        kind: "error",
        text: `An account with ${email} already exists. Log in with your password, then connect this provider from Settings.`,
      });
      setAuthView("login");
    } else if (error) {
      setOauthNotice({ kind: "error", text: oauthErrorMessage(error) });
    }

    // Never leave tokens (or anything else) sitting in the address bar.
    window.history.replaceState({}, "", window.location.pathname);
  }, []);

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
        {oauthNotice && (
          <div className={oauthNotice.kind === "error" ? "error-msg" : "oauth-notice"}>{oauthNotice.text}</div>
        )}
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
        <h1>typebook<span className="brand-badge">secured by CrydenSync</span></h1>
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
      {page === "notes" ? <NotesView /> : <SettingsView onLoggedOut={handleLoggedOut} oauthNotice={oauthNotice} />}
    </>
  );
}

function capitalize(s) {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

function oauthErrorMessage(code) {
  switch (code) {
    case "provider_not_configured":
      return "That sign-in provider isn't set up on this deployment yet.";
    case "state_mismatch":
      return "That login attempt couldn't be verified — please try again.";
    case "link_session_missing":
      return "That linking attempt expired — please try connecting again from Settings.";
    case "provider_error":
      return "Couldn't complete sign-in with that provider — please try again.";
    default:
      return "Something went wrong during sign-in — please try again.";
  }
}
