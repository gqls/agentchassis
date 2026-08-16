package gripper

import (
	"net/mail"
	"strings"

	"github.com/gqls/agentchassis/platform/mailer"
)

// The two emails the poller sends. Plain text only — a link and a sentence or
// two, the way a person would write it. Copy is code so it ships with the
// behaviour and is reviewed with it; it was NOT signed off by the owner at the
// time of writing (no copy existed anywhere in the lane's docs), so treat the
// wording as a first draft that is easy to change and cheap to review.

// LinkMessage is the "your dossier is ready" email.
func LinkMessage(to, link string) mailer.Message {
	return mailer.Message{
		To:      []string{to},
		Subject: "Your gripper selection dossier is ready",
		Text: "Hello,\n\n" +
			"The gripper selection dossier you asked for on robot-hands.com is ready:\n\n" +
			link + "\n\n" +
			"It was built from the details you gave in the chat, scored against the published " +
			"specifications of the grippers in our index. Where a manufacturer does not publish a " +
			"figure the dossier says so rather than guessing.\n\n" +
			"If anything in it looks wrong, or you would like a person to look at the application, " +
			"reply to this email.\n\n" +
			"Robot Hands\n",
	}
}

// ApologyMessage is sent when the report failed to build or nothing arrived
// before the request expired.
func ApologyMessage(to string) mailer.Message {
	return mailer.Message{
		To:      []string{to},
		Subject: "We could not produce your gripper dossier",
		Text: "Hello,\n\n" +
			"Sorry — we were not able to produce the gripper selection dossier you asked for on " +
			"robot-hands.com. That is our failure, not anything you did.\n\n" +
			"If you would like, reply to this email with a line or two about the part and the cell " +
			"and we will look at it by hand.\n\n" +
			"Robot Hands\n",
	}
}

// ValidEmail reports whether s is a single bare address the mailer will accept
// (mailer.validateAddr refuses display names, angle brackets, commas and
// spaces). Returns the normalised address on success.
func ValidEmail(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 254 {
		return "", false
	}
	a, err := mail.ParseAddress(s)
	if err != nil {
		return "", false
	}
	// ParseAddress accepts "Name <x@y>"; the mailer does not. Insist the input
	// was the bare addr-spec.
	if a.Address != s || a.Name != "" {
		return "", false
	}
	if strings.ContainsAny(s, " <>,\r\n") {
		return "", false
	}
	return s, true
}
