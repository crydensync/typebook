const API_BASE = import.meta.env.VITE_API_BASE || "http://localhost:8080";

function getStoredTokens() {
  const raw = localStorage.getItem("typebook_tokens");
  return raw ? JSON.parse(raw) : null;
}

function storeTokens(tokens) {
  localStorage.setItem("typebook_tokens", JSON.stringify(tokens));
}

function clearTokens() {
  localStorage.removeItem("typebook_tokens");
}

async function request(path, { method = "GET", body, auth = false } = {}) {
  const headers = { "Content-Type": "application/json" };
  if (auth) {
    const tokens = getStoredTokens();
    if (!tokens) throw new Error("not authenticated");
    headers["Authorization"] = `Bearer ${tokens.AccessToken}`;
  }

  let res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  // Access token expired — try one silent refresh, then retry once.
  // This is the frontend's real-world proof of RefreshToken working.
  if (res.status === 401 && auth) {
    const refreshed = await tryRefresh();
    if (refreshed) {
      headers["Authorization"] = `Bearer ${refreshed.AccessToken}`;
      res = await fetch(`${API_BASE}${path}`, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      });
    }
  }

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || `request failed: ${res.status}`);
  }
  return data;
}

async function tryRefresh() {
  const tokens = getStoredTokens();
  if (!tokens) return null;
  try {
    const res = await fetch(`${API_BASE}/api/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: tokens.RefreshToken }),
    });
    if (!res.ok) {
      // Refresh failed — could be genuine expiry, or CrydenSync's
      // reuse-detection killing the whole session family. Either way,
      // the client must not keep retrying with a dead token.
      clearTokens();
      return null;
    }
    const newTokens = await res.json();
    storeTokens(newTokens);
    return newTokens;
  } catch {
    clearTokens();
    return null;
  }
}

export const api = {
  signUp: (email, password) =>
    request("/api/signup", { method: "POST", body: { email, password } }),

  login: async (email, password) => {
    const tokens = await request("/api/login", { method: "POST", body: { email, password } });
    storeTokens(tokens);
    return tokens;
  },

  logout: async (sessionId) => {
    await request("/api/logout", { method: "POST", body: { session_id: sessionId }, auth: true });
    clearTokens();
  },

  logoutAll: async () => {
    await request("/api/logout-all", { method: "POST", auth: true });
    clearTokens();
  },

  listSessions: () => request("/api/sessions", { auth: true }),

  revokeSession: (id) => request(`/api/sessions/${id}`, { method: "DELETE", auth: true }),

  changePassword: async (currentPassword, newPassword) => {
    const result = await request("/api/change-password", {
      method: "POST",
      body: { current_password: currentPassword, new_password: newPassword },
      auth: true,
    });
    // Backend just revoked ALL sessions including this one.
    clearTokens();
    return result;
  },

  deleteAccount: async (currentPassword) => {
    const result = await request("/api/delete-account", {
      method: "POST",
      body: { current_password: currentPassword },
      auth: true,
    });
    clearTokens();
    return result;
  },

  requestEmailChange: (newEmail) =>
    request("/api/email/request-change", { method: "POST", body: { new_email: newEmail }, auth: true }),

  confirmEmailChange: (token) =>
    request("/api/email/confirm-change", { method: "POST", body: { token } }),

  listNotes: () => request("/api/notes", { auth: true }),
  createNote: (note) => request("/api/notes", { method: "POST", body: note, auth: true }),
  updateNote: (id, note) => request(`/api/notes/${id}`, { method: "PUT", body: note, auth: true }),
  deleteNote: (id) => request(`/api/notes/${id}`, { method: "DELETE", auth: true }),

  isAuthenticated: () => !!getStoredTokens(),
  logoutLocal: clearTokens,
};
