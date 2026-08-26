import { useState } from "react";
import { api } from "../api";
import OAuthButtons from "./OAuthButtons";

export default function SignupForm({ onSignedUp, onSwitchToLogin }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await api.signUp(email, password);
      // Sign up succeeds, then log in immediately for a smooth flow —
      // the backend's SignUp doesn't itself return tokens (mirrors
      // CrydenSync's own signature: SignUp and Login are separate).
      await api.login(email, password);
      onSignedUp();
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="auth-screen">
      <h2>Create your account</h2>
      <p className="subtitle">Start taking notes with typebook</p>
      <OAuthButtons />
      <form onSubmit={handleSubmit}>
        <div className="field">
          <label>Email</label>
          <input type="email" required value={email} onChange={(e) => setEmail(e.target.value)} />
        </div>
        <div className="field">
          <label>Password</label>
          <input type="password" required value={password} onChange={(e) => setPassword(e.target.value)} />
        </div>
        {error && <div className="error-msg">{error}</div>}
        <button className="btn btn-primary" type="submit" disabled={loading}>
          {loading ? "Creating..." : "Sign up"}
        </button>
      </form>
      <div className="switch-link">
        Already have an account? <button onClick={onSwitchToLogin}>Log in</button>
      </div>
    </div>
  );
}
