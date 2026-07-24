// FILE: platform/orchestration/actions/score_grippers_action_test.go
//
// Table-driven tests for the deterministic MatchMatrix v2 port. The fixture
// set mirrors the live tool's GRIPPERS array (content_components
// fdfeaa7a-be17-46f9-9ecb-3ccba17c8ebc) — the same figures seed 204 writes
// into products.content_data->'matchmatrix'. If the tool's physics and this
// port ever disagree, the tool is the reference; fix the port.

package actions

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func fp(v float64) *float64 { return &v }
func ip(v int) *int         { return &v }

// The full live index, normalised — one fixture per real gripper.
func testGripperIndex() []gripperCandidate {
	mk := func(name string, spec matchmatrixSpec) gripperCandidate {
		return gripperCandidate{Name: name, Spec: spec,
			SourceURL: "https://example.test/" + strings.ToLower(strings.ReplaceAll(name, " ", "-")),
			VerifiedDate: "2026-07-22"}
	}
	return []gripperCandidate{
		mk("Schunk EGP 40-N-S-B", matchmatrixSpec{Tech: "jaw", TechLabel: "Electric parallel-jaw", Maker: "Schunk",
			ForceN: fp(30), ForceText: "30 N",
			StrokeMMTotal: fp(12), StrokeText: "6 mm per jaw (12 mm total)",
			PayloadKg: fp(0.15), PayloadText: "0.15 kg (recommended workpiece weight)",
			IP: ip(30), IPText: "IP30"}),
		mk("OnRobot 2FG7", matchmatrixSpec{Tech: "jaw", TechLabel: "Electric parallel-jaw", Maker: "OnRobot",
			ForceN: fp(140), ForceText: "20 N to 140 N",
			StrokeMMTotal: fp(73), StrokeText: "up to 73 mm external grip range",
			PayloadKg: fp(11), PayloadText: "11 kg (24.3 lb)",
			IP: ip(67), IPText: "IP67"}),
		mk("Robotiq 2F-85", matchmatrixSpec{Tech: "jaw", TechLabel: "Electric parallel-jaw", Maker: "Robotiq",
			ForceN: fp(235), ForceText: "20 to 235 N",
			StrokeMMTotal: fp(85), StrokeText: "85 mm",
			PayloadKg: fp(5), PayloadText: "5 kg"}),
		mk("Zimmer Group GEP5010IO-00-A", matchmatrixSpec{Tech: "jaw", TechLabel: "Electric parallel-jaw", Maker: "Zimmer Group",
			ForceN: fp(1520), ForceText: "1520 N",
			StrokeMMTotal: fp(20), StrokeText: "10 mm per jaw (20 mm total)",
			IP: ip(64), IPText: "IP64"}),
		mk("Festo EHPS-20-A-LK", matchmatrixSpec{Tech: "jaw", TechLabel: "Electric parallel-jaw", Maker: "Festo",
			ForceN: fp(218), ForceText: "218 N",
			StrokeMMTotal: fp(26), StrokeText: "13 mm per jaw (26 mm total)"}),
		mk("Festo DHPS-10-A", matchmatrixSpec{Tech: "jaw", TechLabel: "Pneumatic parallel-jaw", Maker: "Festo",
			ForceN: fp(34.5), ForceText: "34.5 N per jaw closing (39 N opening) at 6 bar",
			StrokeMMTotal: fp(6), StrokeText: "3 mm per jaw (6 mm total)"}),
		mk("OnRobot VG10", matchmatrixSpec{Tech: "vacuum", TechLabel: "Electric vacuum", Maker: "OnRobot",
			PayloadKg: fp(15), PayloadText: "15 kg (35 lb)",
			Note: "Suction hold — needs a surface a vacuum cup can seal against; porous or heavily perforated surfaces reduce holding force. Built-in pump, no external air supply."}),
		mk("OnRobot Gecko SP5", matchmatrixSpec{Tech: "adhesive", TechLabel: "Adhesive (gecko)", Maker: "OnRobot",
			PayloadKg: fp(5), PayloadText: "5 kg",
			Note: "Van der Waals adhesion — requires clean, smooth, dry, flat surfaces; not suitable for greasy, wet or dusty parts. No air and no electricity required."}),
		mk("OnRobot Soft Gripper SG", matchmatrixSpec{Tech: "soft", TechLabel: "Soft silicone", Maker: "OnRobot",
			PayloadKg: fp(2.2), PayloadText: "2.2 kg (depends on shape, softness and friction of the part)",
			GripMinMM: fp(11), GripMaxMM: fp(118), GripText: "11 to 118 mm (cup-dependent)",
			Note: "Payload depends on part geometry — the rating is an upper bound, not a guarantee. Food-grade silicone cups; no external air supply."}),
		mk("Schmalz SGM-HP 50", matchmatrixSpec{Tech: "magnetic", TechLabel: "Permanent magnetic", Maker: "Schmalz",
			ForceN: fp(385), ForceText: "560 N (without friction ring), 385 N (with friction ring)",
			Note: "Permanent-magnet surface hold on ferromagnetic material only. Assessed against the lower published figure (385 N, with friction ring). Workpiece temperatures up to 350 °C."}),
	}
}

func reqFixture(mass, travel, accel, safety, mu float64, muKey string, n float64, ipMin int, checkPayload bool) scoreRequest {
	return scoreRequest{Mass: mass, Travel: travel, Accel: accel, Safety: safety,
		Mu: mu, MuKey: muKey, Material: gripperMaterials[muKey], N: n,
		IPMin: ipMin, CheckPayload: checkPayload}
}

func findCandidate(t *testing.T, result map[string]interface{}, name string) map[string]interface{} {
	t.Helper()
	for _, c := range result["candidates"].([]interface{}) {
		m := c.(map[string]interface{})
		if m["name"] == name {
			return m
		}
	}
	t.Fatalf("candidate %q not in result", name)
	return nil
}

// --- requirements maths + printed formulas -------------------------------

func TestRequirementFiguresAndFormulas(t *testing.T) {
	// 2.5 kg steel housing, a=12, S=2, μ=0.15, n=2: dyn=60, fJaw=200, mEq=6.12
	req := reqFixture(2.5, 40, 12, 2, 0.15, "0.15", 2, 54, true)
	result := scoreGrippers(req, testGripperIndex())

	reqs := result["requirements"].(map[string]interface{})
	if got := reqs["f_jaw_n"].(float64); got != 200 {
		t.Errorf("f_jaw_n = %v, want 200", got)
	}
	if got := reqs["f_dir_n"].(float64); got != 60 {
		t.Errorf("f_dir_n = %v, want 60", got)
	}
	if got := reqs["m_eq_kg"].(float64); got < 6.11 || got > 6.12 {
		t.Errorf("m_eq_kg = %v, want ~6.116", got)
	}
	formulas := reqs["formulas"].([]string)
	// The substituted formula literal is also the discriminating E2E artefact.
	want := "Friction grip (jaw): F = (m × a × S) ÷ (μ × n) = (2.5 × 12 × 2) ÷ (0.15 × 2) = 200.0 N"
	if formulas[0] != want {
		t.Errorf("formula[0] = %q, want %q", formulas[0], want)
	}
}

// --- verdict classes -------------------------------------------------------

func TestVerdicts(t *testing.T) {
	// Same 2.5 kg steel application. Expected, from the tool's own physics:
	//   Zimmer 1520 N: pass force/stroke(20<40? FAIL) — stroke 20 < travel 40 -> No match
	//   2FG7 140 N < 200 -> No match; Robotiq 235>=200 pass, stroke 85 pass,
	//   payload 5>=2.5 pass, no IP published + IP54 required -> unknown -> Insufficient data
	req := reqFixture(2.5, 40, 12, 2, 0.15, "0.15", 2, 54, true)
	result := scoreGrippers(req, testGripperIndex())

	if v := findCandidate(t, result, "Robotiq 2F-85")["verdict"]; v != "Insufficient data" {
		t.Errorf("Robotiq verdict = %v, want Insufficient data (IP not published under an IP requirement)", v)
	}
	if v := findCandidate(t, result, "OnRobot 2FG7")["verdict"]; v != "No match" {
		t.Errorf("2FG7 verdict = %v, want No match (140 N < 200 N)", v)
	}
	if v := findCandidate(t, result, "Zimmer Group GEP5010IO-00-A")["verdict"]; v != "No match" {
		t.Errorf("Zimmer verdict = %v, want No match (20 mm stroke < 40 mm travel)", v)
	}
}

func TestMatchAndMarginal(t *testing.T) {
	// Drop the IP requirement and payload check, travel 15: Zimmer passes
	// force (1520 vs 200, headroom 7.6 -> Match); Robotiq passes at 235 vs
	// 200 -> 235 < 250 (need*1.25) -> Marginal.
	req := reqFixture(2.5, 15, 12, 2, 0.15, "0.15", 2, 0, false)
	result := scoreGrippers(req, testGripperIndex())

	if v := findCandidate(t, result, "Zimmer Group GEP5010IO-00-A")["verdict"]; v != "Match" {
		t.Errorf("Zimmer verdict = %v, want Match", v)
	}
	if v := findCandidate(t, result, "Robotiq 2F-85")["verdict"]; v != "Marginal" {
		t.Errorf("Robotiq verdict = %v, want Marginal (235 < 200*1.25)", v)
	}
	// Sort: first candidate must be rank 0 with the highest headroom (Zimmer).
	first := result["candidates"].([]interface{})[0].(map[string]interface{})
	if first["name"] != "Zimmer Group GEP5010IO-00-A" {
		t.Errorf("first candidate = %v, want Zimmer (rank asc, headroom desc)", first["name"])
	}
}

// --- the honest no-match branch -------------------------------------------

func TestNoMatchIsStatedNotSoftened(t *testing.T) {
	// 500 kg glass part needing IP67: nothing in the index can hold it.
	req := reqFixture(500, 200, 12, 2, 0.10, "0.10", 2, 67, true)
	result := scoreGrippers(req, testGripperIndex())

	if mc := result["match_count"].(int); mc != 0 {
		t.Fatalf("match_count = %d, want 0", mc)
	}
	summary := result["summary_sentence"].(string)
	if !strings.HasPrefix(summary, noMatchSentence) {
		t.Errorf("summary = %q, must start with the mandatory sentence", summary)
	}
	fact := result["fact_block"].(string)
	if !strings.HasPrefix(fact, noMatchSentence) {
		t.Errorf("fact_block must open with the mandatory no-match sentence; got %q", fact[:80])
	}
}

// --- magnetic ferrous gate ---------------------------------------------------

func TestMagneticFerrousGate(t *testing.T) {
	// Steel (ferrous): Schmalz assessed against fDir. 2 kg, a=10, S=2 ->
	// fDir=40 N, 385 published -> pass, headroom 9.6 -> Match.
	reqSteel := reqFixture(2, 0, 10, 2, 0.15, "0.15", 2, 0, false)
	result := scoreGrippers(reqSteel, testGripperIndex())
	if v := findCandidate(t, result, "Schmalz SGM-HP 50")["verdict"]; v != "Match" {
		t.Errorf("Schmalz on steel = %v, want Match", v)
	}

	// Aluminium (non-ferrous): the material gate fails it outright.
	reqAlu := reqFixture(2, 0, 10, 2, 0.20, "0.20", 2, 0, false)
	result = scoreGrippers(reqAlu, testGripperIndex())
	c := findCandidate(t, result, "Schmalz SGM-HP 50")
	if c["verdict"] != "No match" {
		t.Errorf("Schmalz on aluminium = %v, want No match (non-ferromagnetic)", c["verdict"])
	}
}

// --- soft gripper grip-range window ----------------------------------------

func TestSoftGripperRange(t *testing.T) {
	// 8 mm part: smaller than the smallest cup (11 mm) -> fail with the
	// tool's exact wording.
	req := reqFixture(0.5, 8, 9.81, 2, 0.25, "0.25", 2, 0, false)
	result := scoreGrippers(req, testGripperIndex())
	c := findCandidate(t, result, "OnRobot Soft Gripper SG")
	if c["verdict"] != "No match" {
		t.Errorf("Soft SG at 8 mm = %v, want No match", c["verdict"])
	}
	found := false
	for _, r := range c["criteria"].([]interface{}) {
		row := r.(map[string]interface{})
		if row["label"] == "Grip range" && strings.Contains(row["note"].(string), "part smaller than the smallest cup (11 mm)") {
			found = true
		}
	}
	if !found {
		t.Error("grip-range fail note missing or wrong wording")
	}
}

// --- the jaw conflict note ---------------------------------------------------

func TestJawConflictNote(t *testing.T) {
	// 3 kg steel, a=12, S=2, n=2: fJaw = 72/0.3 = 240 N. Robotiq: force 235
	// fails, payload 5 >= 3 passes -> conflict note with the implied-μ maths.
	req := reqFixture(3, 40, 12, 2, 0.15, "0.15", 2, 0, true)
	result := scoreGrippers(req, testGripperIndex())
	c := findCandidate(t, result, "Robotiq 2F-85")
	conflict, _ := c["conflict"].(string)
	if conflict == "" {
		t.Fatal("expected a conflict note for Robotiq (force fails, payload passes)")
	}
	// holds = 235*0.15*2/(12*2) = 2.9 kg;
	// impliedMu = (payload*a*S)/(force*n) = (5*12*2)/(235*2) = 0.26 —
	// the tool derives it from the PUBLISHED payload, not the request mass.
	if !strings.Contains(conflict, "holds only 2.9 kg") || !strings.Contains(conflict, "μ ≈ 0.26") {
		t.Errorf("conflict maths wrong: %q", conflict)
	}
}

// --- fact_block discipline ---------------------------------------------------

func TestFactBlockMarksUnpublishedFigures(t *testing.T) {
	req := reqFixture(2.5, 40, 12, 2, 0.15, "0.15", 2, 54, true)
	result := scoreGrippers(req, testGripperIndex())
	fact := result["fact_block"].(string)

	if !strings.Contains(fact, "NOT PUBLISHED by the manufacturer") {
		t.Error("fact_block must mark unpublished figures explicitly")
	}
	if !strings.Contains(fact, "(2.5 × 12 × 2) ÷ (0.15 × 2)") {
		t.Error("fact_block must carry the substituted formula")
	}
	if !strings.Contains(fact, "Source: https://example.test/robotiq-2f-85") {
		t.Error("fact_block must carry per-candidate source URLs")
	}
}

// --- input parsing / mapping table -------------------------------------------

func TestBuildScoreRequestMappings(t *testing.T) {
	logger := zap.NewNop()
	build := func(overrides map[string]interface{}) (scoreRequest, error) {
		collected := map[string]interface{}{
			"site_id": "00ff3af5-dad8-4770-9f70-3edc267a3c92",
			"mass_kg": "2.5", "travel_mm": "40", "surface_material": "steel",
		}
		for k, v := range overrides {
			collected[k] = v
		}
		inputs, err := datahelpers.ExtractActionInputs(collected, map[string]interface{}{}, ScoreGrippersInputSpec, logger)
		if err != nil {
			t.Fatalf("extract: %v", err)
		}
		return buildScoreRequest(inputs)
	}

	// Defaults: a=9.81, n=2, S=2 (no cycle_rate), checkPayload=true.
	req, err := build(nil)
	if err != nil {
		t.Fatal(err)
	}
	if req.Accel != 9.81 || req.N != 2 || req.Safety != 2 || !req.CheckPayload {
		t.Errorf("defaults wrong: %+v", req)
	}
	if req.MuKey != "0.15" || !req.Material.Ferrous {
		t.Errorf("material alias 'steel' -> %q, want 0.15/ferrous", req.MuKey)
	}

	// Cycle-rate tiers (owner-approved mapping): <=10 -> 2, <=30 -> 3, >30 -> 4.
	for rate, want := range map[string]float64{"5": 2, "20": 3, "40": 4} {
		req, err = build(map[string]interface{}{"cycle_rate": rate})
		if err != nil {
			t.Fatal(err)
		}
		if req.Safety != want {
			t.Errorf("cycle_rate %s -> S=%v, want %v", rate, req.Safety, want)
		}
	}

	// Explicit safety factor beats the tier.
	req, _ = build(map[string]interface{}{"cycle_rate": "40", "safety_factor": "2.5"})
	if req.Safety != 2.5 {
		t.Errorf("explicit safety_factor lost to tier: %v", req.Safety)
	}

	// "machined aluminium" resolves by alias; IP accepts an "IP54" form.
	req, _ = build(map[string]interface{}{"surface_material": "machined aluminium", "ip_min": "IP54"})
	if req.MuKey != "0.20" || req.IPMin != 54 {
		t.Errorf("alias/IP parse wrong: mu=%q ip=%d", req.MuKey, req.IPMin)
	}

	// Malformed mass must be a hard error (routes the workflow to error_step).
	if _, err = build(map[string]interface{}{"mass_kg": "not-a-number"}); err == nil {
		t.Error("mass_kg 'not-a-number' must error")
	}
	// Unknown material must be a hard error, not a silent default.
	if _, err = build(map[string]interface{}{"surface_material": "unobtainium"}); err == nil {
		t.Error("unknown surface_material must error")
	}
}
