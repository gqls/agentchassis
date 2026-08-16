package config

import (
	"strings"
	"testing"
)

func setBase(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("ANTHROPIC_API_KEY", "k")
}

// The gripper group is opt-in by env: with the key unset, Load must succeed
// exactly as before and leave Gripper nil, so an island .env that predates the
// feature boots the same binary unchanged.
func TestGripperIsNilWhenKeyUnset(t *testing.T) {
	setBase(t)
	t.Setenv(GripperAPIKeyEnv, "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Gripper != nil {
		t.Fatal("Gripper config present without the key")
	}
}

// Once opted in, the rest is required and each gap names itself.
func TestGripperRequiresPullKeyAndSMTPOnceOptedIn(t *testing.T) {
	setBase(t)
	t.Setenv(GripperAPIKeyEnv, "sk-test")

	if _, err := Load(); err == nil || !strings.Contains(err.Error(), GripperPullKeyEnv) {
		t.Fatalf("missing pull key: err=%v", err)
	}
	t.Setenv(GripperPullKeyEnv, "short")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), GripperPullKeyEnv) {
		t.Fatalf("short pull key accepted: err=%v", err)
	}
	t.Setenv(GripperPullKeyEnv, "0123456789abcdef0123456789abcdef")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "GRIPPER_SMTP") {
		t.Fatalf("missing smtp: err=%v", err)
	}
	t.Setenv("GRIPPER_SMTP_HOST", "mail.example")
	t.Setenv("GRIPPER_SMTP_FROM", "robot-hands@example")
	t.Setenv("GRIPPER_SMTP_PORT", "465")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	g := cfg.Gripper
	if g == nil || g.Model != GripperDefaultModel || g.DailyTurnCap != GripperDefaultDailyTurnCap ||
		g.MaxBodyBytes != GripperDefaultMaxBodyBytes || g.SMTP.Port != "465" || g.APIKeyEnvVar != GripperAPIKeyEnv {
		t.Fatalf("gripper config = %+v", g)
	}
	t.Setenv(GripperDailyTurnCapEnv, "0")
	if _, err := Load(); err == nil {
		t.Fatal("zero daily cap accepted")
	}
}
