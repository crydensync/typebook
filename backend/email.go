package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// resendEmailSender sends real verification emails via Resend's HTTP API
// (https://resend.com/docs/api-reference/emails/send-email) — a single
// plain HTTP POST, no SDK or extra Go dependency needed, consistent with
// this backend's otherwise stdlib-only, dependency-light approach.
//
// fromAddress must be on a domain you've verified in your Resend account.
// Until you verify a domain, Resend only lets you send from their test
// address (onboarding@resend.dev) and only to the email you signed up
// with — fine for smoke-testing this integration, not for real users.
type resendEmailSender struct {
	apiKey      string
	fromAddress string
	frontendURL string // e.g. https://typebook.vercel.app — no trailing slash
}

type resendEmailPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

func (s *resendEmailSender) SendVerification(ctx context.Context, to string, rawToken string) error {
	link := fmt.Sprintf("%s/confirm-email?token=%s", s.frontendURL, rawToken)

	payload := resendEmailPayload{
		From:    s.fromAddress,
		To:      []string{to},
		Subject: "Confirm your new email address — typebook",
		HTML: fmt.Sprintf(
			`<p>Click the link below to confirm this email address for your typebook account.</p>`+
				`<p><a href="%s">%s</a></p>`+
				`<p>If you didn't request this, you can safely ignore this email — your account is unaffected.</p>`,
			link, link,
		),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("resend: failed to encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("resend: failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("resend: request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("resend: API returned %d: %s", res.StatusCode, string(respBody))
	}

	return nil
}

// consoleEmailSender is a dev stand-in implementing notify.EmailSender —
// used automatically when RESEND_API_KEY isn't set (see main.go), so
// local development still works without a real email account. This is
// the concrete proof of CrydenSync's zero-telemetry design: the engine
// never sends email itself — it hands us a raw token, and what we do
// with it is entirely our choice.
type consoleEmailSender struct {
	frontendURL string
}

func (s *consoleEmailSender) SendVerification(ctx context.Context, to string, rawToken string) error {
	log.Printf("[EMAIL] (RESEND_API_KEY not set — dev mode) Verification link for %s: %s/confirm-email?token=%s", to, s.frontendURL, rawToken)
	return nil
}
