package main

import (
	"context"
	"log"
)

// consoleEmailSender is a dev stand-in implementing notify.EmailSender.
// This is the concrete proof of CrydenSync's zero-telemetry design:
// the engine never sends email itself — it handed us a raw token, and
// what we do with it is entirely our choice. For a real deployment,
// replace this with a real provider (SES, SendGrid, Postmark, etc.)
// that builds an actual clickable URL using YOUR domain/routes.
type consoleEmailSender struct{}

func (s *consoleEmailSender) SendVerification(ctx context.Context, to string, rawToken string) error {
	log.Printf("[EMAIL] Verification link for %s: http://localhost:5173/confirm-email?token=%s", to, rawToken)
	return nil
}