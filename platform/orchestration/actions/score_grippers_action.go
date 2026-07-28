// FILE: platform/orchestration/actions/score_grippers_action.go
//
// ScoreGrippersAction is the deterministic, server-side port of the
// MatchMatrix v2 scoring physics that ships client-side on robot-hands.com
// (content_components function='tool-matchmatrix'). It reads the normalised
// per-gripper spec blocks from products.content_data->'matchmatrix' (seeded
// from the same verified figures the tool carries — see
// docs/agent_docs/sql_for_agents/204_robot_hands_matchmatrix_normalized_specs.sql)
// and scores every active gripper for one visitor application.
//
// NO LLM is involved. Published figures are compared, never invented; a
// missing figure is reported as "not published" and can only ever soften a
// verdict to "Insufficient data", never harden it to a pass. This action is
// the sole source of numbers for the report-builder's prose steps: its
// fact_block output is injected into the writer prompt as the ONLY numbers
// and product names the writer may assert (the per-request analogue of the
// evidence_base writer_block convention, bugs_open/043 lane).
//
// Physics (identical to the tool, deliberately):
//   dyn  = m * a * S
//   fJaw = dyn / (mu * n)     — friction-grip force per surface (jaw)
//   fDir = dyn                — direct normal holding force (magnetic)
//   mEq  = dyn / 9.81         — dynamics-adjusted payload (vacuum/adhesive/soft)
// Verdicts: any failed criterion -> "No match"; else any unpublished
// checked criterion -> "Insufficient data"; else capacity < need*1.25 ->
// "Marginal"; else "Match". Sort: verdict rank asc, then headroom desc.
//
// Config / data inputs (via ActionInputSpec, all values arrive as strings
// from query_database extraction of the work item spec):
//   - site_id          (required) — the site whose gripper index to score
//   - mass_kg          (required) — workpiece mass
//   - travel_mm        (required) — required jaw opening / part size
//   - surface_material (required) — mu key ("0.15") or material name ("steel")
//   - surfaces_n       (optional, default 2)  — gripping surfaces
//   - accel_ms2        (optional, default 9.81)
//   - safety_factor    (optional) — if absent, derived from cycle_rate tier
//   - cycle_rate       (optional, picks/min) — tier map: <=10 -> S=2,
//                       <=30 -> S=3, >30 -> S=4 (DESIGN §5.3, owner-approved)
//   - ip_min           (optional, default 0 = no requirement)
//   - check_payload    (optional, default true)

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpec
// ============================================================================

var ScoreGrippersInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id", "mass_kg", "travel_mm", "surface_material"},
	Optional: []string{"surfaces_n", "accel_ms2", "safety_factor", "cycle_rate",
		"ip_min", "check_payload"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("score_grippers", ScoreGrippersInputSpec)
}

// ============================================================================
// Material table — mirrors the tool's MATERIALS map exactly. Keys are the
// friction coefficient as a string (the tool's option values); the alias map
// lets the chat intake submit a plain material name instead.
// ============================================================================

type gripperMaterial struct {
	Mu      float64
	Name    string
	Ferrous bool
}

var gripperMaterials = map[string]gripperMaterial{
	"0.10": {0.10, "Glass, smooth", false},
	"0.15": {0.15, "Steel, dry", true},
	"0.20": {0.20, "Aluminium, machined", false},
	"0.25": {0.25, "Plastic / ABS", false},
	"0.30": {0.30, "Cardboard", false},
	"0.50": {0.50, "Rubber", false},
}

var gripperMaterialAliases = map[string]string{
	"glass": "0.10", "steel": "0.15", "aluminium": "0.20", "aluminum": "0.20",
	"plastic": "0.25", "abs": "0.25", "cardboard": "0.30", "rubber": "0.50",
}

func resolveMaterial(s string) (key string, m gripperMaterial, ok bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if m, ok := gripperMaterials[s]; ok {
		return s, m, true
	}
	for alias, key := range gripperMaterialAliases {
		if strings.Contains(s, alias) {
			return key, gripperMaterials[key], true
		}
	}
	return "", gripperMaterial{}, false
}

// ============================================================================
// Normalised gripper spec block: products.content_data->'matchmatrix'.
// A nil pointer means the manufacturer publishes no such figure — it is
// reported as unpublished, never inferred (the tool's exact contract).
// ============================================================================

type matchmatrixSpec struct {
	Tech          string     `json:"tech"` // jaw | vacuum | magnetic | adhesive | soft
	TechLabel     string     `json:"tech_label"`
	Maker         string     `json:"maker"`
	ForceN        *float64   `json:"force_n"`
	ForceText     string     `json:"force_text"`
	StrokeMMTotal *float64   `json:"stroke_mm_total"`
	StrokeText    string     `json:"stroke_text"`
	PayloadKg     *float64   `json:"payload_kg"`
	PayloadText   string     `json:"payload_text"`
	IP            *int       `json:"ip"`
	IPText        string     `json:"ip_text"`
	GripMinMM     *float64   `json:"grip_min_mm"`
	GripMaxMM     *float64   `json:"grip_max_mm"`
	GripText      string     `json:"grip_text"`
	Note          string     `json:"note"`
	Extras        [][]string `json:"extras"`
}

type gripperCandidate struct {
	Name         string
	Spec         matchmatrixSpec
	SourceURL    string
	VerifiedDate string
}

type scoreRequest struct {
	Mass         float64
	Travel       float64
	Accel        float64
	Safety       float64
	Mu           float64
	MuKey        string
	Material     gripperMaterial
	N            float64
	IPMin        int
	CheckPayload bool
}

type criteriaRow struct {
	Label string  `json:"label"`
	Text  *string `json:"published_text"` // nil = not published by manufacturer
	Flag  string  `json:"flag"`           // pass | fail | none | ""
	Note  string  `json:"note"`
}

type assessment struct {
	Name         string        `json:"name"`
	Maker        string        `json:"maker"`
	Tech         string        `json:"tech"`
	TechLabel    string        `json:"tech_label"`
	Verdict      string        `json:"verdict"`
	Rank         int           `json:"rank"`
	Headroom     float64       `json:"headroom"`
	Criteria     []criteriaRow `json:"criteria"`
	Extras       [][]string    `json:"extras"`
	Conflict     string        `json:"conflict,omitempty"`
	TechNote     string        `json:"tech_note,omitempty"`
	SourceURL    string        `json:"source_url"`
	VerifiedDate string        `json:"verified_date"`
}

// ============================================================================
// ACTION: score_grippers
// ============================================================================

func ScoreGrippersAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "score_grippers"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, ScoreGrippersInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	req, err := buildScoreRequest(inputs)
	if err != nil {
		return nil, err // deliberate hard failure: a malformed spec must route to error_step
	}

	candidates, err := loadGripperCandidates(ctx, params, siteID)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no active grippers with matchmatrix spec blocks for site %s (seed 204 applied?)", siteID)
	}

	result := scoreGrippers(req, candidates)

	logger.Info("ScoreGrippersAction: scored",
		zap.Int("candidates", len(candidates)),
		zap.Int("match_count", result["match_count"].(int)))
	return result, nil
}

func buildScoreRequest(inputs *datahelpers.ActionInputs) (scoreRequest, error) {
	var req scoreRequest

	parseF := func(field string, min float64, required bool, def float64) (float64, error) {
		s := strings.TrimSpace(inputs.Get(field))
		if s == "" {
			if required {
				return 0, fmt.Errorf("required field %s is empty", field)
			}
			return def, nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < min {
			return 0, fmt.Errorf("field %s: invalid value %q", field, s)
		}
		return v, nil
	}

	var err error
	if req.Mass, err = parseF("mass_kg", 1e-9, true, 0); err != nil {
		return req, err
	}
	if req.Travel, err = parseF("travel_mm", 0, true, 0); err != nil {
		return req, err
	}
	if req.Accel, err = parseF("accel_ms2", 0, false, 9.81); err != nil {
		return req, err
	}
	if req.N, err = parseF("surfaces_n", 1, false, 2); err != nil {
		return req, err
	}

	// Safety factor: explicit wins; otherwise derive from the cycle-rate tier
	// (DESIGN §5.3 mapping, owner-approved 2026-07-24); otherwise default 2.
	if s := strings.TrimSpace(inputs.Get("safety_factor")); s != "" {
		if req.Safety, err = parseF("safety_factor", 1, true, 0); err != nil {
			return req, err
		}
	} else if c := strings.TrimSpace(inputs.Get("cycle_rate")); c != "" {
		rate, cerr := strconv.ParseFloat(c, 64)
		if cerr != nil || rate < 0 {
			return req, fmt.Errorf("field cycle_rate: invalid value %q", c)
		}
		switch {
		case rate <= 10:
			req.Safety = 2
		case rate <= 30:
			req.Safety = 3
		default:
			req.Safety = 4
		}
	} else {
		req.Safety = 2
	}

	key, mat, ok := resolveMaterial(inputs.Get("surface_material"))
	if !ok {
		return req, fmt.Errorf("field surface_material: %q matches no material in the index", inputs.Get("surface_material"))
	}
	req.MuKey, req.Mu, req.Material = key, mat.Mu, mat

	if s := strings.TrimSpace(inputs.Get("ip_min")); s != "" {
		ip, ierr := strconv.Atoi(strings.TrimPrefix(strings.ToUpper(s), "IP"))
		if ierr != nil || ip < 0 {
			return req, fmt.Errorf("field ip_min: invalid value %q", s)
		}
		req.IPMin = ip
	}

	req.CheckPayload = true
	if s := strings.ToLower(strings.TrimSpace(inputs.Get("check_payload"))); s == "false" || s == "0" {
		req.CheckPayload = false
	}
	return req, nil
}

func loadGripperCandidates(ctx context.Context, params ActionParams, siteID uuid.UUID) ([]gripperCandidate, error) {
	rows, err := params.DB.QueryContext(ctx, `
		SELECT name,
		       content_data->'matchmatrix',
		       COALESCE(content_data->>'source_url', ''),
		       COALESCE(content_data->>'verified_date', '')
		FROM products
		WHERE site_id = $1 AND category = 'gripper' AND status = 'active'
		  AND content_data ? 'matchmatrix'
		ORDER BY name`, siteID)
	if err != nil {
		return nil, fmt.Errorf("loading grippers: %w", err)
	}
	defer rows.Close()

	var out []gripperCandidate
	for rows.Next() {
		var c gripperCandidate
		var raw []byte
		if err := rows.Scan(&c.Name, &raw, &c.SourceURL, &c.VerifiedDate); err != nil {
			return nil, fmt.Errorf("scanning gripper row: %w", err)
		}
		if err := json.Unmarshal(raw, &c.Spec); err != nil {
			return nil, fmt.Errorf("gripper %q: malformed matchmatrix block: %w", c.Name, err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ============================================================================
// The physics. Each assess* function mirrors its JS namesake line for line.
// ============================================================================

type assessState struct {
	failed   bool
	unknown  bool
	capacity float64
	need     float64
}

func fmtF(v float64, dp int) string { return strconv.FormatFloat(v, 'f', dp, 64) }

// trimNum renders a float the way the tool's JS string-concatenation does:
// "2.5" not "2.50", "12" not "12.0".
func trimNum(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func ipRow(spec matchmatrixSpec, req scoreRequest, state *assessState) *criteriaRow {
	if req.IPMin > 0 {
		if spec.IP != nil {
			ok := *spec.IP >= req.IPMin
			if !ok {
				state.failed = true
			}
			note := ""
			if !ok {
				note = fmt.Sprintf("needs IP%d", req.IPMin)
			}
			t := spec.IPText
			return &criteriaRow{Label: "IP rating", Text: &t, Flag: flagOf(ok), Note: note}
		}
		state.unknown = true
		return &criteriaRow{Label: "IP rating", Text: nil, Flag: "none"}
	}
	if spec.IP != nil {
		t := spec.IPText
		return &criteriaRow{Label: "IP rating", Text: &t, Flag: "", Note: "no requirement set"}
	}
	return nil
}

func flagOf(ok bool) string {
	if ok {
		return "pass"
	}
	return "fail"
}

func assessJaw(spec matchmatrixSpec, req scoreRequest, fJaw float64, state *assessState) []criteriaRow {
	var rows []criteriaRow

	if spec.ForceN != nil {
		ok := *spec.ForceN >= fJaw
		if !ok {
			state.failed = true
		}
		note := ""
		if ok {
			if fJaw > 0 {
				margin := (*spec.ForceN - fJaw) / fJaw * 100
				if margin >= 0 {
					note = fmt.Sprintf("+%s%% margin", fmtF(margin, 0))
				}
			}
		} else {
			note = fmt.Sprintf("needs %s N", fmtF(fJaw, 1))
		}
		t := spec.ForceText
		rows = append(rows, criteriaRow{Label: "Gripping force", Text: &t, Flag: flagOf(ok), Note: note})
	} else {
		state.unknown = true
		rows = append(rows, criteriaRow{Label: "Gripping force", Text: nil, Flag: "none"})
	}

	if spec.StrokeMMTotal != nil {
		ok := *spec.StrokeMMTotal >= req.Travel
		if !ok {
			state.failed = true
		}
		note := ""
		if !ok {
			note = fmt.Sprintf("needs %s mm", trimNum(req.Travel))
		}
		t := spec.StrokeText
		rows = append(rows, criteriaRow{Label: "Jaw travel", Text: &t, Flag: flagOf(ok), Note: note})
	} else {
		state.unknown = true
		rows = append(rows, criteriaRow{Label: "Jaw travel", Text: nil, Flag: "none"})
	}

	if req.CheckPayload {
		if spec.PayloadKg != nil {
			ok := *spec.PayloadKg >= req.Mass
			if !ok {
				state.failed = true
			}
			note := ""
			if !ok {
				note = fmt.Sprintf("needs %s kg", trimNum(req.Mass))
			}
			t := spec.PayloadText
			rows = append(rows, criteriaRow{Label: "Rated payload", Text: &t, Flag: flagOf(ok), Note: note})
		} else {
			state.unknown = true
			rows = append(rows, criteriaRow{Label: "Rated payload", Text: nil, Flag: "none"})
		}
	} else if spec.PayloadKg != nil {
		t := spec.PayloadText
		rows = append(rows, criteriaRow{Label: "Rated payload", Text: &t, Flag: "", Note: "not checked"})
	}

	if ip := ipRow(spec, req, state); ip != nil {
		rows = append(rows, *ip)
	}

	if spec.ForceN != nil {
		state.capacity = *spec.ForceN
	}
	state.need = fJaw
	return rows
}

func assessMagnetic(spec matchmatrixSpec, req scoreRequest, fDir float64, state *assessState) []criteriaRow {
	var rows []criteriaRow

	okM := req.Material.Ferrous
	if !okM {
		state.failed = true
	}
	matNote := "requires a ferromagnetic workpiece"
	if okM {
		matNote = "ferromagnetic"
	}
	matName := req.Material.Name
	rows = append(rows, criteriaRow{Label: "Workpiece material", Text: &matName, Flag: flagOf(okM), Note: matNote})

	// The magnetic gripper in this index always publishes a holding force;
	// guard anyway so a mis-seeded row degrades to unknown, not a panic.
	if spec.ForceN != nil {
		ok := *spec.ForceN >= fDir
		if !ok {
			state.failed = true
		}
		note := fmt.Sprintf("needs %s N direct hold", fmtF(fDir, 1))
		if ok && fDir > 0 {
			note = fmt.Sprintf("+%s%% vs %s N direct hold", fmtF((*spec.ForceN-fDir)/fDir*100, 0), fmtF(fDir, 1))
		}
		t := spec.ForceText
		rows = append(rows, criteriaRow{Label: "Holding force", Text: &t, Flag: flagOf(ok), Note: note})
		state.capacity = *spec.ForceN
	} else {
		state.unknown = true
		rows = append(rows, criteriaRow{Label: "Holding force", Text: nil, Flag: "none"})
	}

	na := "Not applicable — surface hold, no jaws"
	rows = append(rows, criteriaRow{Label: "Jaw travel", Text: &na, Flag: ""})

	if ip := ipRow(spec, req, state); ip != nil {
		rows = append(rows, *ip)
	}

	state.need = fDir
	return rows
}

func assessPayloadRated(spec matchmatrixSpec, req scoreRequest, mEq float64, state *assessState) []criteriaRow {
	var rows []criteriaRow

	// Payload assessed against the dynamics-adjusted equivalent mass, so a
	// rating earned at rest is not silently credited with surviving the
	// visitor's acceleration and safety factor.
	if spec.PayloadKg != nil {
		ok := *spec.PayloadKg >= mEq
		if !ok {
			state.failed = true
		}
		note := fmt.Sprintf("needs %s kg equivalent (mass × a × S ÷ g)", fmtF(mEq, 2))
		if ok {
			margin := 0.0
			if mEq > 0 {
				margin = (*spec.PayloadKg - mEq) / mEq * 100
			}
			note = fmt.Sprintf("+%s%% vs %s kg equivalent", fmtF(margin, 0), fmtF(mEq, 2))
		}
		t := spec.PayloadText
		rows = append(rows, criteriaRow{Label: "Rated payload", Text: &t, Flag: flagOf(ok), Note: note})
		state.capacity = *spec.PayloadKg
	} else {
		state.unknown = true
		rows = append(rows, criteriaRow{Label: "Rated payload", Text: nil, Flag: "none"})
	}

	if spec.Tech == "soft" && (spec.GripMinMM == nil || spec.GripMaxMM == nil) {
		// A soft gripper whose cup range the manufacturer never published.
		//
		// This branch used to fall through to the "Not applicable — surface
		// hold, no jaws" else below, which is correct for a VACUUM or MAGNETIC
		// gripper (they genuinely have no jaw travel) but silently wrong here:
		// a soft gripper DOES have a size window, we simply do not know it, and
		// treating unknown as "no constraint" let the candidate score Match on
		// its remaining criteria. A paying customer was told the part fits a cup
		// nobody has published the size of.
		//
		// Every other unpublished figure in this file sets state.unknown; this
		// one did not. Found by TestUnknownNeverPasses on its first real run.
		state.unknown = true
		rows = append(rows, criteriaRow{Label: "Grip range", Text: nil, Flag: "none"})
	} else if spec.Tech == "soft" {
		if req.Travel > 0 {
			ok := req.Travel >= *spec.GripMinMM && req.Travel <= *spec.GripMaxMM
			if !ok {
				state.failed = true
			}
			note := ""
			if !ok {
				if req.Travel < *spec.GripMinMM {
					note = fmt.Sprintf("part smaller than the smallest cup (%s mm)", trimNum(*spec.GripMinMM))
				} else {
					note = fmt.Sprintf("part exceeds the largest cup (%s mm)", trimNum(*spec.GripMaxMM))
				}
			}
			t := spec.GripText
			rows = append(rows, criteriaRow{Label: "Grip range", Text: &t, Flag: flagOf(ok), Note: note})
		} else {
			t := spec.GripText
			rows = append(rows, criteriaRow{Label: "Grip range", Text: &t, Flag: "", Note: "no part size set"})
		}
	} else {
		na := "Not applicable — surface hold, no jaws"
		rows = append(rows, criteriaRow{Label: "Jaw travel", Text: &na, Flag: ""})
	}

	if ip := ipRow(spec, req, state); ip != nil {
		rows = append(rows, *ip)
	}

	state.need = mEq
	return rows
}

func assessGripper(c gripperCandidate, req scoreRequest, fJaw, fDir, mEq float64) assessment {
	state := assessState{need: 1}
	var rows []criteriaRow

	switch c.Spec.Tech {
	case "magnetic":
		rows = assessMagnetic(c.Spec, req, fDir, &state)
	case "vacuum", "adhesive", "soft":
		rows = assessPayloadRated(c.Spec, req, mEq, &state)
	default: // jaw
		rows = assessJaw(c.Spec, req, fJaw, &state)
	}

	// A jaw gripper can pass on its published payload rating and still fail
	// the force calculation, because headline payload figures assume the
	// manufacturer's own friction assumptions. Say so rather than leaving the
	// reader to reconcile two rows that appear to contradict each other.
	conflict := ""
	if c.Spec.Tech == "jaw" && c.Spec.ForceN != nil && c.Spec.PayloadKg != nil &&
		*c.Spec.ForceN < fJaw && *c.Spec.PayloadKg >= req.Mass {
		holds := (*c.Spec.ForceN * req.Mu * req.N) / (req.Accel * req.Safety)
		impliedMu := (*c.Spec.PayloadKg * req.Accel * req.Safety) / (*c.Spec.ForceN * req.N)
		conflict = fmt.Sprintf(
			"Rated for %s, but %s N holds only %s kg on your surface (μ %s). "+
				"That payload rating implies μ ≈ %s — high-friction or form-fit fingers, not a bare machined surface.",
			c.Spec.PayloadText, trimNum(*c.Spec.ForceN), fmtF(holds, 1), trimNum(req.Mu), fmtF(impliedMu, 2))
	}

	var verdict string
	var rank int
	switch {
	case state.failed:
		verdict, rank = "No match", 3
	case state.unknown:
		verdict, rank = "Insufficient data", 2
	case state.capacity < state.need*1.25:
		verdict, rank = "Marginal", 1
	default:
		verdict, rank = "Match", 0
	}

	headroom := 0.0
	if state.need > 0 {
		headroom = state.capacity / state.need
	}

	return assessment{
		Name: c.Name, Maker: c.Spec.Maker, Tech: c.Spec.Tech, TechLabel: c.Spec.TechLabel,
		Verdict: verdict, Rank: rank, Headroom: headroom, Criteria: rows,
		Extras: c.Spec.Extras, Conflict: conflict, TechNote: c.Spec.Note,
		SourceURL: c.SourceURL, VerifiedDate: c.VerifiedDate,
	}
}

const noMatchSentence = "No gripper in this index meets the requirement."

func scoreGrippers(req scoreRequest, candidates []gripperCandidate) map[string]interface{} {
	dyn := req.Mass * req.Accel * req.Safety
	fJaw := dyn / (req.Mu * req.N)
	fDir := dyn
	mEq := dyn / 9.81

	formulas := []string{
		fmt.Sprintf("Friction grip (jaw): F = (m × a × S) ÷ (μ × n) = (%s × %s × %s) ÷ (%s × %s) = %s N",
			trimNum(req.Mass), trimNum(req.Accel), trimNum(req.Safety),
			trimNum(req.Mu), trimNum(req.N), fmtF(fJaw, 1)),
		fmt.Sprintf("Direct hold (magnetic): F = m × a × S = %s N", fmtF(fDir, 1)),
		fmt.Sprintf("Equivalent payload (vacuum / adhesive / soft): m′ = (m × a × S) ÷ g = %s kg", fmtF(mEq, 2)),
	}

	assessed := make([]assessment, 0, len(candidates))
	for _, c := range candidates {
		assessed = append(assessed, assessGripper(c, req, fJaw, fDir, mEq))
	}
	sort.SliceStable(assessed, func(i, j int) bool {
		if assessed[i].Rank != assessed[j].Rank {
			return assessed[i].Rank < assessed[j].Rank
		}
		return assessed[i].Headroom > assessed[j].Headroom
	})

	matchCount := 0
	for _, a := range assessed {
		if a.Rank == 0 || a.Rank == 1 {
			matchCount++
		}
	}
	summary := fmt.Sprintf("%d of %d indexed grippers meet the requirement.", matchCount, len(assessed))
	if matchCount == 0 {
		summary = noMatchSentence + " The closest are shown first."
	}

	requestEcho := map[string]interface{}{
		"mass_kg": req.Mass, "travel_mm": req.Travel, "accel_ms2": req.Accel,
		"safety_factor": req.Safety, "mu": req.Mu, "mu_key": req.MuKey,
		"material": req.Material.Name, "surfaces_n": req.N,
		"ip_min": req.IPMin, "check_payload": req.CheckPayload,
	}

	// Marshal assessments through JSON so downstream template/prose steps see
	// plain maps, matching how CollectedData round-trips through the saga.
	var candidateMaps []interface{}
	b, _ := json.Marshal(assessed)
	_ = json.Unmarshal(b, &candidateMaps)

	return map[string]interface{}{
		"request": requestEcho,
		"requirements": map[string]interface{}{
			"f_jaw_n": fJaw, "f_dir_n": fDir, "m_eq_kg": mEq, "formulas": formulas,
		},
		"candidates":       candidateMaps,
		"match_count":      matchCount,
		"summary_sentence": summary,
		"fact_block":       buildFactBlock(req, formulas, assessed, matchCount),

		// The report contract the honesty gate enforces, carried WITH the data
		// that produced it rather than shared as a package const.
		//
		// verify_report_prose used to reach for this file's `noMatchSentence`
		// const and a hardcoded section list. That works only while exactly one
		// report type exists in this package — the moment a second scorer is
		// added, the gate silently checks THIS report's sentence against THAT
		// report's prose, and check (3), the honest-no-match contract, becomes
		// decorative while still reporting success. Same shape as the
		// render_directory incident (bugs_open/042): a value that should have
		// been carried explicitly was resolved from ambient state instead, and a
		// default silently won.
		//
		// A scorer therefore declares the sentence its writer was told to use and
		// the prose sections it expects, exactly as it already declares
		// fact_block. The gate REFUSES when either is absent — it must never
		// fall back to a default, because a gate that defaults its own contract
		// is not a gate.
		"no_match_sentence": noMatchSentence,
		"prose_sections":    reportProseSections(),
	}
}

// reportProseSections is the gripper dossier's prose section list, in render
// order. Returned as a fresh slice per call so a consumer cannot mutate the
// contract for every later run in the same process.
func reportProseSections() []string {
	return []string{"summary_html", "candidates_html", "integration_html", "vendor_questions_html"}
}

// buildFactBlock renders the plain-text whitelist of every number and name
// the report writer is allowed to assert — the per-request analogue of the
// evidence_base writer_block already live on this site.
func buildFactBlock(req scoreRequest, formulas []string, assessed []assessment, matchCount int) string {
	var b strings.Builder
	if matchCount == 0 {
		b.WriteString(noMatchSentence + "\n\n")
	}
	b.WriteString("APPLICATION (as submitted):\n")
	fmt.Fprintf(&b, "- mass %s kg; required opening/part size %s mm; acceleration %s m/s²; safety factor %s; surface %s (μ %s); gripping surfaces %s",
		trimNum(req.Mass), trimNum(req.Travel), trimNum(req.Accel),
		trimNum(req.Safety), req.Material.Name, trimNum(req.Mu), trimNum(req.N))
	if req.IPMin > 0 {
		fmt.Fprintf(&b, "; required protection IP%d", req.IPMin)
	}
	b.WriteString("\n\nREQUIREMENT FIGURES (computed, with formulas):\n")
	for _, f := range formulas {
		b.WriteString("- " + f + "\n")
	}
	b.WriteString("\nCANDIDATES (verdicts and published figures — the ONLY product facts that may be asserted):\n")
	for _, a := range assessed {
		fmt.Fprintf(&b, "\n%s (%s, %s) — VERDICT: %s (capacity/need headroom %.2f)\n",
			a.Name, a.Maker, a.TechLabel, a.Verdict, a.Headroom)
		for _, r := range a.Criteria {
			if r.Text == nil {
				fmt.Fprintf(&b, "  - %s: NOT PUBLISHED by the manufacturer — say so if mentioned; never estimate it\n", r.Label)
				continue
			}
			line := fmt.Sprintf("  - %s: %s", r.Label, *r.Text)
			if r.Flag == "pass" || r.Flag == "fail" {
				line += " [" + strings.ToUpper(r.Flag) + "]"
			}
			if r.Note != "" {
				line += " (" + r.Note + ")"
			}
			b.WriteString(line + "\n")
		}
		if a.Conflict != "" {
			b.WriteString("  - CONFLICT NOTE (quote or paraphrase only this): " + a.Conflict + "\n")
		}
		if a.TechNote != "" {
			b.WriteString("  - TECHNOLOGY NOTE: " + a.TechNote + "\n")
		}
		if a.SourceURL != "" {
			fmt.Fprintf(&b, "  - Source: %s (specs as published; verified %s)\n", a.SourceURL, a.VerifiedDate)
		}
	}
	return b.String()
}
