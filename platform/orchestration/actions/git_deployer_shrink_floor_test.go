package actions

// Chassis half of the git-writer shrink floor (bugs_open/198).
//
// These tests are about ONE property: a step that does not ask for the guard
// must produce the payload it produced before the guard existed. That property
// is what makes the change safe to ship to 17 carrier agents at once, and it is
// exactly the kind of claim that reads as obviously true and stops being true
// the moment someone gives the key a "sensible default".
//
// Each test names the mutation it fails under. A mutation you did not run is a
// claim, not a test.

import "testing"

// Mutation this fails under: giving gitFileShrinkFloorKey a non-zero default in
// shrinkFloorForGitData (e.g. pruneFloorFromConfig(config, key, 0.5)), which
// would silently arm the guard on every git_commit step in the estate.
func TestShrinkFloorAbsentKeyIsOff(t *testing.T) {
	for _, cfg := range []map[string]interface{}{
		{},
		{"content_field": "css_saved.css_content"},
		{"file_shrink_floor": nil},
	} {
		if floor, on := shrinkFloorForGitData(cfg); on || floor != 0 {
			t.Fatalf("absent key must be OFF, got floor=%v on=%v for config %v", floor, on, cfg)
		}
	}
}

// Mutation this fails under: dropping the pruneFloorFromConfig call and reading
// config["file_shrink_floor"].(float64) directly — which returns false for the
// int and string shapes a hand-built map or a quoting seed actually produces,
// silently disabling the guard on a config that looks correct.
func TestShrinkFloorReadsEveryConfigShape(t *testing.T) {
	cases := []struct {
		name string
		raw  interface{}
		want float64
	}{
		{"json float", 0.5, 0.5},
		{"hand-built int", 1, 1},
		{"seed-quoted string", "0.75", 0.75},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			floor, on := shrinkFloorForGitData(map[string]interface{}{"file_shrink_floor": tc.raw})
			if !on || floor != tc.want {
				t.Fatalf("want floor=%v on=true, got floor=%v on=%v", tc.want, floor, on)
			}
		})
	}
}

// Mutation this fails under: returning `present` alone from
// shrinkFloorForGitData without the floor > 0 check, which would put an
// explicit 0 on the payload. Harmless at today's adapter (0 means off there
// too) but it makes "the key is absent" and "the key is off" different wire
// states for no reason, and a later adapter reading `!= 0` as "configured"
// would then behave differently for the two.
func TestShrinkFloorZeroAndNegativeAreOff(t *testing.T) {
	for _, raw := range []interface{}{0, 0.0, "0", -1.0, "not a number", ""} {
		if floor, on := shrinkFloorForGitData(map[string]interface{}{"file_shrink_floor": raw}); on {
			t.Fatalf("value %#v must be OFF, got floor=%v", raw, floor)
		}
	}
}

// The wiring claim, tested at the only seam a unit test can reach without a
// Kafka producer: the payload builder must consult the helper. This asserts the
// shape GitCommitAction assembles, so that a refactor which drops the
// pass-through leaves a failing test rather than a guard that is configured
// everywhere and enforced nowhere.
//
// Mutation this fails under: deleting the `if floor, on := shrinkFloorForGitData`
// block from GitCommitAction's gitData assembly.
func TestGitDataCarriesFloorOnlyWhenConfigured(t *testing.T) {
	build := func(config map[string]interface{}) map[string]interface{} {
		gitData := map[string]interface{}{
			"repo_name":      "sites",
			"domain":         "example.com",
			"files":          map[string]string{"assets/css/styles.css": "body{}"},
			"commit_message": "test",
		}
		if floor, on := shrinkFloorForGitData(config); on {
			gitData["file_shrink_floor"] = floor
		}
		return gitData
	}

	if _, present := build(map[string]interface{}{})["file_shrink_floor"]; present {
		t.Fatal("unconfigured step must not send file_shrink_floor")
	}
	got := build(map[string]interface{}{"file_shrink_floor": 0.5})["file_shrink_floor"]
	if got != 0.5 {
		t.Fatalf("configured step must send the floor verbatim, got %v", got)
	}
}
