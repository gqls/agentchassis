// FILE: platform/orchestration/actions/agent_image_test.go
package actions

import "testing"

// bugs_open/066. The cases that matter are the ones where the row and the
// running chassis DISAGREE — a test where they already agree cannot fail for
// the reason this fix exists.

const chassisRepo = "docker.io/aqls/agent-chassis"

func def(repo, tag string, cfg map[string]interface{}) *AgentDefinition {
	return &AgentDefinition{
		Type:            "feature-implementer",
		ImageRepository: repo,
		ImageTag:        tag,
		DefaultConfig:   cfg,
	}
}

func TestChooseAgentImage(t *testing.T) {
	cases := []struct {
		name        string
		selfImages  []string
		agentDef    *AgentDefinition
		wantRef     string
		wantSource  string
		wantDrifted string
	}{
		{
			// The reported bug: the row trails the running chassis by four
			// tags, and the spawned pod used to run the row's tag.
			name:        "row trails the running chassis",
			selfImages:  []string{chassisRepo + ":v1.0.1155"},
			agentDef:    def(chassisRepo, "v1.0.1151", nil),
			wantRef:     chassisRepo + ":v1.0.1155",
			wantSource:  imageSourceRunningChassis,
			wantDrifted: "v1.0.1151",
		},
		{
			// The direction observed live on 2026-07-27: the row was updated
			// 35s BEFORE the Deployment rolled, so the row LED the chassis. A
			// spawned pod must still follow the spawner, not the row.
			name:        "row leads the running chassis",
			selfImages:  []string{chassisRepo + ":v1.0.1172"},
			agentDef:    def(chassisRepo, "v1.0.1173", nil),
			wantRef:     chassisRepo + ":v1.0.1172",
			wantSource:  imageSourceRunningChassis,
			wantDrifted: "v1.0.1173",
		},
		{
			name:       "row agrees with the running chassis: no drift reported",
			selfImages: []string{chassisRepo + ":v1.0.1173"},
			agentDef:   def(chassisRepo, "v1.0.1173", nil),
			wantRef:    chassisRepo + ":v1.0.1173",
			wantSource: imageSourceRunningChassis,
		},
		{
			// The escape hatch, and the supported form of 066's interim rule.
			name:       "an explicit pin is honoured over the running chassis",
			selfImages: []string{chassisRepo + ":v1.0.1173"},
			agentDef:   def(chassisRepo, "v1.0.1151", map[string]interface{}{"pin_image_tag": true}),
			wantRef:    chassisRepo + ":v1.0.1151",
			wantSource: imageSourcePinned,
		},
		{
			name:       "pin_image_tag false behaves as unpinned",
			selfImages: []string{chassisRepo + ":v1.0.1173"},
			agentDef:   def(chassisRepo, "v1.0.1151", map[string]interface{}{"pin_image_tag": false}),
			wantRef:    chassisRepo + ":v1.0.1173",
			wantSource: imageSourceRunningChassis,
			// drift still reported — the row is still wrong
			wantDrifted: "v1.0.1151",
		},
		{
			// A non-boolean value must not be read as a pin, and must not panic.
			name:        "pin_image_tag as a string is not a pin",
			selfImages:  []string{chassisRepo + ":v1.0.1173"},
			agentDef:    def(chassisRepo, "v1.0.1151", map[string]interface{}{"pin_image_tag": "true"}),
			wantRef:     chassisRepo + ":v1.0.1173",
			wantSource:  imageSourceRunningChassis,
			wantDrifted: "v1.0.1151",
		},
		{
			// The honesty test: an agent that deliberately runs some other
			// image must be left completely alone.
			name:       "a different repository is never rewritten",
			selfImages: []string{chassisRepo + ":v1.0.1173"},
			agentDef:   def("docker.io/aqls/some-other-agent", "v2.3.4", nil),
			wantRef:    "docker.io/aqls/some-other-agent:v2.3.4",
			wantSource: imageSourceDefinition,
		},
		{
			name:        "registry prefix differences still match",
			selfImages:  []string{"aqls/agent-chassis:v1.0.1173"},
			agentDef:    def(chassisRepo, "v1.0.1151", nil),
			wantRef:     "aqls/agent-chassis:v1.0.1173",
			wantSource:  imageSourceRunningChassis,
			wantDrifted: "v1.0.1151",
		},
		{
			// Sidecars: pick the container matching the requested repository,
			// not simply the first one.
			name:        "a sidecar image is skipped in favour of the matching repository",
			selfImages:  []string{"docker.io/istio/proxyv2:1.20.0", chassisRepo + ":v1.0.1173"},
			agentDef:    def(chassisRepo, "v1.0.1151", nil),
			wantRef:     chassisRepo + ":v1.0.1173",
			wantSource:  imageSourceRunningChassis,
			wantDrifted: "v1.0.1151",
		},
		{
			// The fallback that makes this safe to ship: no self-lookup, old
			// behaviour exactly.
			name:       "no self images falls back to the row",
			selfImages: nil,
			agentDef:   def(chassisRepo, "v1.0.1151", nil),
			wantRef:    chassisRepo + ":v1.0.1151",
			wantSource: imageSourceDefinition,
		},
		{
			name:       "a digest-pinned self image falls back to the row",
			selfImages: []string{chassisRepo + "@sha256:0123456789abcdef"},
			agentDef:   def(chassisRepo, "v1.0.1151", nil),
			wantRef:    chassisRepo + ":v1.0.1151",
			wantSource: imageSourceDefinition,
		},
		{
			// image_repository is nullable; an empty row means "whatever the
			// chassis runs".
			name:        "an empty repository on the row adopts the running image",
			selfImages:  []string{chassisRepo + ":v1.0.1173"},
			agentDef:    def("", "", nil),
			wantRef:     chassisRepo + ":v1.0.1173",
			wantSource:  imageSourceRunningChassis,
			wantDrifted: "(null)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := chooseAgentImage(tc.selfImages, tc.agentDef)
			if got.Ref() != tc.wantRef {
				t.Errorf("Ref() = %q, want %q", got.Ref(), tc.wantRef)
			}
			if got.Source != tc.wantSource {
				t.Errorf("Source = %q, want %q", got.Source, tc.wantSource)
			}
			if got.DriftedFrom != tc.wantDrifted {
				t.Errorf("DriftedFrom = %q, want %q", got.DriftedFrom, tc.wantDrifted)
			}
		})
	}
}

func TestParseImageRef(t *testing.T) {
	cases := []struct {
		ref      string
		wantRepo string
		wantTag  string
		wantOK   bool
	}{
		{"docker.io/aqls/agent-chassis:v1.0.1173", "docker.io/aqls/agent-chassis", "v1.0.1173", true},
		{"aqls/agent-chassis:latest", "aqls/agent-chassis", "latest", true},
		// A colon before the last slash is a registry port, not a tag.
		{"registry.local:5000/aqls/agent-chassis:v1.0.1", "registry.local:5000/aqls/agent-chassis", "v1.0.1", true},
		{"registry.local:5000/aqls/agent-chassis", "registry.local:5000/aqls/agent-chassis", "latest", true},
		{"agent-chassis", "agent-chassis", "latest", true},
		// Digests are declined rather than guessed at.
		{"docker.io/aqls/agent-chassis@sha256:abc", "", "", false},
		{"", "", "", false},
		{"   ", "", "", false},
		{"docker.io/aqls/agent-chassis:", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.ref, func(t *testing.T) {
			repo, tag, ok := parseImageRef(tc.ref)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if repo != tc.wantRepo || tag != tc.wantTag {
				t.Errorf("= (%q, %q), want (%q, %q)", repo, tag, tc.wantRepo, tc.wantTag)
			}
		})
	}
}

func TestSameRepository(t *testing.T) {
	same := [][2]string{
		{"docker.io/aqls/agent-chassis", "aqls/agent-chassis"},
		{"index.docker.io/aqls/agent-chassis", "docker.io/aqls/agent-chassis"},
		{" docker.io/aqls/agent-chassis ", "aqls/agent-chassis"},
	}
	for _, p := range same {
		if !sameRepository(p[0], p[1]) {
			t.Errorf("sameRepository(%q, %q) = false, want true", p[0], p[1])
		}
	}

	differ := [][2]string{
		{"docker.io/aqls/agent-chassis", "docker.io/aqls/agent-chassis-dev"},
		{"docker.io/aqls/agent-chassis", "ghcr.io/aqls/agent-chassis"},
		{"docker.io/aqls/agent-chassis", "docker.io/other/agent-chassis"},
	}
	for _, p := range differ {
		if sameRepository(p[0], p[1]) {
			t.Errorf("sameRepository(%q, %q) = true, want false", p[0], p[1])
		}
	}
}
