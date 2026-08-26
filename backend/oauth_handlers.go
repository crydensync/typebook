package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/crydensync/cryden/v2"
	"github.com/crydensync/cryden/v2/auth"
)

// oauthConfig holds everything the OAuth handlers need, read once
// from env in main.go and passed in — same style as jwtSecret/dsn
// elsewhere in this file, no separate config package like `api` has.
type oauthConfig struct {
	baseURL     string // this backend's own public URL, for building provider callback URLs
	frontendURL string // where the browser lands after a successful (or failed) login, same var main.go already uses for email links

	googleClientID     string
	googleClientSecret string
	githubClientID     string
	githubClientSecret string
}

type oauthProvider struct {
	name         string
	clientID     string
	clientSecret string
	authURL      string
	tokenURL     string
	userInfoURL  string
	scope        string
}

type oauthHandlers struct {
	engine *cryden.Engine
	cfg    oauthConfig
}

const oauthStateCookie = "typebook_oauth_state"
const oauthLinkUserCookie = "typebook_oauth_link_user"

func (h *oauthHandlers) provider(name string) (oauthProvider, bool) {
	switch name {
	case "google":
		if h.cfg.googleClientID == "" || h.cfg.googleClientSecret == "" {
			return oauthProvider{}, false
		}
		return oauthProvider{
			name:         "google",
			clientID:     h.cfg.googleClientID,
			clientSecret: h.cfg.googleClientSecret,
			authURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			tokenURL:     "https://oauth2.googleapis.com/token",
			userInfoURL:  "https://www.googleapis.com/oauth2/v3/userinfo",
			scope:        "openid email",
		}, true
	case "github":
		if h.cfg.githubClientID == "" || h.cfg.githubClientSecret == "" {
			return oauthProvider{}, false
		}
		return oauthProvider{
			name:         "github",
			clientID:     h.cfg.githubClientID,
			clientSecret: h.cfg.githubClientSecret,
			authURL:      "https://github.com/login/oauth/authorize",
			tokenURL:     "https://github.com/login/oauth/access_token",
			userInfoURL:  "https://api.github.com/user",
			scope:        "read:user user:email",
		}, true
	default:
		return oauthProvider{}, false
	}
}

func (h *oauthHandlers) callbackURL(providerName string) string {
	return h.cfg.baseURL + "/api/oauth/" + providerName + "/callback"
}

func (h *oauthHandlers) linkCallbackURL(providerName string) string {
	return h.cfg.baseURL + "/api/oauth/" + providerName + "/link/callback"
}

// redirectToFrontendWithError sends the browser back to the frontend
// with an error code in the fragment — same handoff shape as the
// success path, so the frontend has one place (not two) to check for
// the outcome of an OAuth attempt.
func (h *oauthHandlers) redirectToFrontendWithError(w http.ResponseWriter, r *http.Request, code string) {
	http.Redirect(w, r, h.cfg.frontendURL+"/#oauth_error="+url.QueryEscape(code), http.StatusFound)
}

// redirectToFrontendWithErrorLogged is the same, but for the paths
// where an actual Go error caused the failure — logs it server-side
// first. The frontend only ever sees a generic code (never leak
// internal error text to the browser), but the terminal running this
// server sees the real reason, which is the only way to debug a
// failed provider round trip from outside.
func (h *oauthHandlers) redirectToFrontendWithErrorLogged(w http.ResponseWriter, r *http.Request, provider, stage, code string, err error) {
	log.Printf("oauth %s (%s): %s: %v", stage, provider, code, err)
	h.redirectToFrontendWithError(w, r, code)
}

// Start redirects the browser to the provider's consent screen. Full
// browser navigation (GET), not something the frontend calls with
// fetch — the button in LoginForm/SignupForm should just set
// window.location to this URL directly.
func (h *oauthHandlers) start(w http.ResponseWriter, r *http.Request, providerName string) {
	p, ok := h.provider(providerName)
	if !ok {
		h.redirectToFrontendWithError(w, r, "provider_not_configured")
		return
	}
	state, err := randomState()
	if err != nil {
		h.redirectToFrontendWithError(w, r, "internal_error")
		return
	}
	setShortCookie(w, oauthStateCookie, state)

	q := url.Values{
		"client_id":     {p.clientID},
		"redirect_uri":  {h.callbackURL(p.name)},
		"response_type": {"code"},
		"scope":         {p.scope},
		"state":         {state},
	}
	http.Redirect(w, r, p.authURL+"?"+q.Encode(), http.StatusFound)
}

// callback receives the provider's redirect, exchanges the code,
// fetches the confirmed identity, calls cryden.LoginWithOAuth, then
// hands tokens to the frontend via a URL FRAGMENT (#access_token=...),
// never a query param — fragments never reach the server or get
// logged by an intervening proxy/analytics tool, and the frontend
// clears it from the address bar immediately after reading it.
func (h *oauthHandlers) callback(w http.ResponseWriter, r *http.Request, providerName string) {
	p, ok := h.provider(providerName)
	if !ok {
		h.redirectToFrontendWithError(w, r, "provider_not_configured")
		return
	}
	if !verifyState(r) {
		h.redirectToFrontendWithError(w, r, "state_mismatch")
		return
	}
	clearCookie(w, oauthStateCookie)

	code := r.URL.Query().Get("code")
	if code == "" {
		h.redirectToFrontendWithError(w, r, "missing_code")
		return
	}

	externalID, email, err := exchangeAndFetchIdentity(r, p, h.callbackURL(p.name), code)
	if err != nil {
		h.redirectToFrontendWithErrorLogged(w, r, p.name, "callback", "provider_error", err)
		return
	}

	tokens, err := cryden.LoginWithOAuth(r.Context(), h.engine, p.name, externalID, email, callerIP(r), userAgent(r))
	if err != nil {
		var conflict *auth.ErrOAuthEmailConflict
		if errors.As(err, &conflict) {
			// The confirmed decision: never auto-link. Send the
			// frontend a distinct code + the email involved, so it
			// can show "an account with this email already exists —
			// log in with your password, then connect Google from
			// Settings" instead of a generic failure.
			http.Redirect(w, r, h.cfg.frontendURL+"/#oauth_error=email_conflict&oauth_email="+url.QueryEscape(conflict.Email), http.StatusFound)
			return
		}
		h.redirectToFrontendWithErrorLogged(w, r, p.name, "callback", "login_failed", err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/#access_token=%s&refresh_token=%s",
		h.cfg.frontendURL, url.QueryEscape(tokens.AccessToken), url.QueryEscape(tokens.RefreshToken)), http.StatusFound)
}

// linkInit is called via an authenticated fetch() (NOT a plain
// navigation — a top-level browser navigation can't carry a custom
// Authorization header, and a redirect response to fetch() never
// moves the actual browser tab, so the user would never see the
// provider's real consent screen). It only verifies the Bearer token
// and signs the caller's user ID into a cookie — no redirect here.
// The frontend follows this with a PLAIN navigation to linkStart,
// which is what actually sends the browser to the provider.
func (h *oauthHandlers) linkInit(w http.ResponseWriter, r *http.Request, providerName string) {
	if _, ok := h.provider(providerName); !ok {
		writeError(w, http.StatusNotFound, "oauth provider not configured")
		return
	}
	userID := userIDFromContext(r)
	signed, err := signLinkUserID(jwtSecretForSigning, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not start link flow")
		return
	}
	setShortCookie(w, oauthLinkUserCookie, signed)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// linkStart is a PLAIN GET — no Authorization header involved. It
// reads the user ID from the cookie linkInit just set (sent
// automatically since this is a same-origin follow-up request) and
// performs the actual redirect to the provider's consent screen. This
// is the request the frontend reaches via window.location, not fetch.
func (h *oauthHandlers) linkStart(w http.ResponseWriter, r *http.Request, providerName string) {
	p, ok := h.provider(providerName)
	if !ok {
		h.redirectToFrontendWithError(w, r, "provider_not_configured")
		return
	}
	if _, err := r.Cookie(oauthLinkUserCookie); err != nil {
		// linkInit was never called first, or the cookie already
		// expired — send the user back rather than starting a
		// provider redirect that linkCallback can't complete.
		h.redirectToFrontendWithError(w, r, "link_session_missing")
		return
	}

	state, err := randomState()
	if err != nil {
		h.redirectToFrontendWithError(w, r, "internal_error")
		return
	}
	setShortCookie(w, oauthStateCookie, state)

	q := url.Values{
		"client_id":     {p.clientID},
		"redirect_uri":  {h.linkCallbackURL(p.name)},
		"response_type": {"code"},
		"scope":         {p.scope},
		"state":         {state},
	}
	http.Redirect(w, r, p.authURL+"?"+q.Encode(), http.StatusFound)
}

// linkCallback is deliberately NOT behind requireAuth — a browser
// redirect carries no Authorization header. The linking user's
// identity instead comes from the signed cookie linkStart set.
func (h *oauthHandlers) linkCallback(w http.ResponseWriter, r *http.Request, providerName string) {
	p, ok := h.provider(providerName)
	if !ok {
		h.redirectToFrontendWithError(w, r, "provider_not_configured")
		return
	}
	if !verifyState(r) {
		h.redirectToFrontendWithError(w, r, "state_mismatch")
		return
	}
	clearCookie(w, oauthStateCookie)

	cookie, err := r.Cookie(oauthLinkUserCookie)
	if err != nil {
		h.redirectToFrontendWithError(w, r, "link_session_missing")
		return
	}
	userID, err := verifyLinkUserID(jwtSecretForSigning, cookie.Value)
	if err != nil {
		h.redirectToFrontendWithError(w, r, "link_session_missing")
		return
	}
	clearCookie(w, oauthLinkUserCookie)

	code := r.URL.Query().Get("code")
	if code == "" {
		h.redirectToFrontendWithError(w, r, "missing_code")
		return
	}

	externalID, email, err := exchangeAndFetchIdentity(r, p, h.linkCallbackURL(p.name), code)
	if err != nil {
		h.redirectToFrontendWithErrorLogged(w, r, p.name, "linkCallback", "provider_error", err)
		return
	}

	if err := cryden.LinkOAuthIdentity(r.Context(), h.engine, userID, p.name, externalID, email, callerIP(r)); err != nil {
		h.redirectToFrontendWithErrorLogged(w, r, p.name, "linkCallback", "link_failed", err)
		return
	}
	http.Redirect(w, r, h.cfg.frontendURL+"/#oauth_linked="+p.name, http.StatusFound)
}

// jwtSecretForSigning is set once from main.go's jwtSecret at
// startup — reused to sign the link-flow cookie rather than adding a
// second secret to configure, since it's already a real, private
// value with the right properties for HMAC signing.
var jwtSecretForSigning string

func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func verifyState(r *http.Request) bool {
	cookie, err := r.Cookie(oauthStateCookie)
	if err != nil || cookie.Value == "" {
		return false
	}
	return r.URL.Query().Get("state") == cookie.Value
}

func setShortCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: value, Path: "/api/oauth",
		HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, MaxAge: 600,
	})
}

func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name: name, Value: "", Path: "/api/oauth",
		HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, MaxAge: -1,
	})
}

func signLinkUserID(secret, userID string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("no signing secret configured")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(userID))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(userID)) + "." + sig, nil
}

func verifyLinkUserID(secret, value string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("no signing secret configured")
	}
	i := -1
	for idx := 0; idx < len(value); idx++ {
		if value[idx] == '.' {
			i = idx
			break
		}
	}
	if i < 0 {
		return "", fmt.Errorf("malformed link cookie")
	}
	userIDBytes, err := base64.RawURLEncoding.DecodeString(value[:i])
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(userIDBytes)
	expected := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(value[i+1:])) {
		return "", fmt.Errorf("link cookie signature mismatch")
	}
	return string(userIDBytes), nil
}

func exchangeAndFetchIdentity(r *http.Request, p oauthProvider, redirectURI, code string) (externalID, email string, err error) {
	token, err := exchangeCode(r, p, redirectURI, code)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, p.userInfoURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("oauth userinfo request failed: status %d", resp.StatusCode)
	}

	switch p.name {
	case "google":
		var info struct {
			Sub   string `json:"sub"`
			Email string `json:"email"`
		}
		if err := json.Unmarshal(body, &info); err != nil {
			return "", "", err
		}
		return info.Sub, info.Email, nil
	case "github":
		var info struct {
			ID    int64  `json:"id"`
			Email string `json:"email"`
		}
		if err := json.Unmarshal(body, &info); err != nil {
			return "", "", err
		}
		if info.Email == "" {
			email, err := fetchGitHubPrimaryEmail(r, token)
			if err != nil {
				return "", "", err
			}
			return fmt.Sprintf("%d", info.ID), email, nil
		}
		return fmt.Sprintf("%d", info.ID), info.Email, nil
	default:
		return "", "", fmt.Errorf("unknown provider %q", p.name)
	}
}

func fetchGitHubPrimaryEmail(r *http.Request, token string) (string, error) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github /user/emails request failed: status %d", resp.StatusCode)
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("no primary verified email available from github")
}

func exchangeCode(r *http.Request, p oauthProvider, redirectURI, code string) (string, error) {
	form := url.Values{
		"client_id":     {p.clientID},
		"client_secret": {p.clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	}
	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, p.tokenURL, nil)
	if err != nil {
		return "", err
	}
	req.URL.RawQuery = form.Encode()
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oauth token exchange failed: status %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("oauth token exchange returned no access_token")
	}
	return tokenResp.AccessToken, nil
}
