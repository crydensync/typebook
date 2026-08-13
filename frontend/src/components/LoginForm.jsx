import { useState } from "react";
import { api } from "../api";

export default function LoginForm({ onLoggedIn, onSwitchToSignup }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await api.login(email, password);
      onLoggedIn();
    } catch (err) {
      // Surfaces CrydenSync's actual errors as written — including
      // "account temporarily locked" after repeated failed attempts.
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="auth-screen">
      <h2>Welcome back</h2>
      <p className="subtitle">Log in to typebook</p>
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
          {loading ? "Logging in..." : "Log in"}
        </button>
      </form>
      <div className="switch-link">
        No account yet? <button onClick={onSwitchToSignup}>Sign up</button>
      </div>
    </div>
  );
}
