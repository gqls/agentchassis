// FILE: platform/orchestration/actions/page_role_upsert.go
//
// UpsertPageForRole — one answer, in one place, to "the name I want to create is
// already taken", for every arm whose page role is a compile-time constant.
//
// WHY THIS EXISTS (bugs_open/175, and bugs_closed/081 before it)
//
// The shape it removes:
//
//	INSERT INTO pages (site_id, name, url, title, page_type, sections, ...)
//	VALUES ($1, $2, $3, $4, 'tool', $5::jsonb, ...)
//	ON CONFLICT (site_id, name) DO UPDATE SET
//	    url = EXCLUDED.url, title = EXCLUDED.title,
//	    sections = EXCLUDED.sections, updated_at = NOW()
//	RETURNING id
//
// `page_type` is in the INSERT and NOT in the `DO UPDATE SET`. So on a name
// collision the statement silently stops being a CREATE and becomes a PARTIAL
// update: this arm's content lands under the EXISTING row's role. Nothing errors,
// `RETURNING id` yields an id either way, and the caller cannot tell a create from
// an overwrite. On a page that is already live that is two defects at once — the
// live content is replaced by content written for a role the page does not hold,
// and the routing key stays wrong, so whatever asked for the page asks again on
// the next sweep. bugs_closed/081 measured that loop running for three months on
// ai-agent-orchestration.com before anyone noticed.
//
// 081 fixed one arm. The council's `bug_historian` seat objected that nothing had
// audited the siblings, and it was right: four more carried the identical shape
// (bugs_open/175's census). This helper is 175's fix candidate 2 — the only one
// that also stops a SEVENTH arm being written with the same bug, because the next
// arm reaches for a function rather than for a fresh INSERT.
//
// # THE CONTRACT, AND THE ONE THING THAT MAKES IT SAFE
//
// This is for an arm whose `PageType` is a CONSTANT OF THE ARM ITSELF — the tool
// deployer always writes `tool`, the report builder always writes `report`, the
// companion-guide arms always write `blog-post`. That is what licenses the ADOPT
// branch below. An arm taking its role from an LLM plan must NOT use this helper:
// giving a generic, model-steered arm the authority to re-type a page is exactly
// what 081 declined to do, and `applyNewPage` keeps its own resolver for that
// reason (see apply_gap_plan_action.go).
//
// THE FOUR OUTCOMES
//
//	no collision                          → CREATE, every declared column
//	collision, SAME role                  → REFRESH the caller's declared subset
//	collision, DIFFERENT role, has shipped → REFUSE: touch nothing, file for a human
//	collision, DIFFERENT role, unshipped,
//	    AdoptUnshippedRows declared        → ADOPT: take the row over completely
//	collision, DIFFERENT role, unshipped,
//	    NOT declared                       → REFUSE: touch nothing, no item filed
//
// The last two are one branch split by a caller's declaration, and it is opt-in
// with the default OFF by OWNER RULING 2026-08-02
// (architecture_review/RFC_010 decision 1). Five council seats independently
// observed that ADOPT was the widest new authority here and was licensed only by
// a doc comment; a field makes it visible to a reviewer of the CALLER, which a
// comment in this file never could.
//
// "Has shipped" is `NOT (datahelpers.NeverDeployedPagePredicate)` — the estate's
// ONE definition of "this page has been served", reused rather than restated. It
// is wider than 081's `build_status = 'deployed'`, which is the point:
// bugs_closed/037 is an entire case about `needs_rebuild` falling outside a
// `= 'deployed'` guard, and 35 of 46 `needs_rebuild` rows carry a non-null
// `deployed_at` — they have shipped and are being served. The shared predicate
// catches them through `deployed_at IS NULL`, not by naming the status.
//
// > **CORRECTED 2026-08-02, hours after this file first shipped, and the mistake
// > is worth more than the fix.** The first version spelled the predicate out
// > here as `everDeployed || build_status == "deployed" || build_status ==
// > "needs_rebuild"`. That third clause is WRONG, and it is wrong in the
// > direction that costs: measured live, **11 rows are `needs_rebuild` with NO
// > `deployed_at`, no `last_built_at` and mostly zero components** — never built,
// > never served — and three of them are `lendzy.co.uk` TOOL pages created the
// > same day, i.e. exactly what the tool arm collides with. The hand-rolled
// > predicate would have REFUSED all eleven: a filed human decision and a hard
// > error on the tool arm, for pages nothing has ever served.
// >
// > `datahelpers.NeverDeployedPagePredicate` already existed, already had a test
// > forbidding precisely the clause I added (*"predicate must not single out
// > needs_rebuild"* — singling it out had produced a 34-page false-positive class
// > for the nav lane), and already had two other consumers. **I answered the
// > council's "did a page-upsert HELPER already exist?" objection and never asked
// > the same question about the PREDICATE.** The predicate is now read from
// > datahelpers, so the two cannot drift.
//
// ADOPT is a WHOLE takeover, not a partial one, and that is a decision rather than
// convenience. Adopting the type but leaving (say) the url would mint a fresh
// hybrid — the same defect wearing different columns. The row has never been
// served, the arm owns the role by construction, and the arm was already writing
// this content into it; taking it over completely is the only shape that cannot
// leave a half-claimed page behind.
//
// ERRORS PROPAGATE, AND THAT IS LOAD-BEARING FOR THE TESTS. Every statement's
// error is checked and returned. bugs_closed/081's first test file was a
// decoration because the production code discarded an Exec error, so an extra
// `UPDATE pages` on the refusal path did not fail anything
// (LANDMINES.md — "mock.ExpectationsWereMet() is NOT 'no database call
// happened'"). Keep it that way: if you add a statement here, check its error.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// PageRoleOutcome names what actually happened, so a caller and a log line can
// tell a create from an overwrite — the thing the raw upsert could not express.
type PageRoleOutcome string

const (
	// PageRoleCreated — the name was free and the row was written.
	PageRoleCreated PageRoleOutcome = "created"
	// PageRoleRefreshed — the name was taken by a page already holding this
	// role, and the caller's declared Refresh columns were re-written.
	PageRoleRefreshed PageRoleOutcome = "refreshed"
	// PageRoleAdopted — the name was taken by a page that has NEVER been live,
	// holding a different role; the arm took the row over completely.
	PageRoleAdopted PageRoleOutcome = "adopted"
	// PageRoleRefused — the name is held by a LIVE page of a different role.
	// Nothing was mutated and a human decision was filed.
	PageRoleRefused PageRoleOutcome = "refused"
)

// PageColumn is one column of the write. Cast is the Postgres type to cast the
// placeholder to ("jsonb", "text[]"), empty for none — the driver cannot infer
// jsonb from a Go string, which is why this is explicit rather than magic.
type PageColumn struct {
	Name  string
	Value interface{}
	Cast  string
}

// JSONCol is the common case: a marshalled column that must be cast to jsonb.
func JSONCol(name, value string) PageColumn {
	return PageColumn{Name: name, Value: value, Cast: "jsonb"}
}

// Col is a plain column.
func Col(name string, value interface{}) PageColumn {
	return PageColumn{Name: name, Value: value}
}

// PageRoleUpsert is one page write by a constant-role arm.
type PageRoleUpsert struct {
	SiteID uuid.UUID
	Domain string // for the refusal message and item spec; may be empty
	Name   string

	// PageType is the role this arm writes. It must be a constant of the arm —
	// see the file comment. It is never taken from the existing row and never
	// silently dropped.
	PageType string

	// Columns are every column written on CREATE, excluding site_id, name and
	// page_type, which this helper owns. Column NAMES must be compile-time
	// literals; they are interpolated into the statement (values never are).
	Columns []PageColumn

	// Refresh names the columns re-written when the row already holds this role.
	// It is the caller's conservative subset — the companion-guide arms refresh
	// only `title`, deliberately, so a built guide does not lose its sections to
	// a redeploy. Empty means "leave a same-role row alone entirely".
	// Every name must appear in Columns.
	Refresh []string

	// AdoptUnshippedRows declares that this arm may take over a row that has
	// NEVER been served but holds a different role — rewriting every declared
	// column INCLUDING page_type. See the ADOPT branch in the file header.
	//
	// OWNER RULING 2026-08-02 (architecture_review/RFC_010, decision 1): this is
	// opt-in, default OFF, and it is a field rather than a doc comment on
	// purpose. The behaviour is right for all four arms that use it today, and
	// the only thing that stood between it and the design bugs_closed/081
	// deliberately REJECTED — a model-steered arm re-typing pages — was a human
	// remembering to read a comment. On a tree this many sessions share, that is
	// not a control. The field puts the authority where a reviewer of the CALLER
	// can see it.
	//
	// Left false, a collision with an unshipped row of another role is REFUSED
	// rather than silently refreshed. Refreshing is what shipped bugs_open/175 in
	// the first place; refusing cannot corrupt anything, and the reason names the
	// missing declaration so it is diagnosable rather than mysterious.
	AdoptUnshippedRows bool

	// Source is the arm's name, used for the refusal item's source/created_by
	// ("tool-deployer", "report-builder").
	Source string

	// OriginalItemID, when set, is BLOCKED on a refusal — mirroring
	// applyNewPage. 'complete' on an item whose defect is untouched is the false
	// green this estate keeps relearning.
	OriginalItemID *uuid.UUID
}

// PageRoleResult reports the id and, crucially, WHICH branch ran.
type PageRoleResult struct {
	PageID  uuid.UUID
	Outcome PageRoleOutcome

	// ExistingType / ExistingBuild describe the row that was collided with;
	// empty when there was no collision.
	ExistingType  string
	ExistingBuild string

	// ItemFiled reports whether a refusal actually filed a new work item (the
	// anti-churn rules in insertWorkItem can suppress it).
	ItemFiled bool

	// Reason is human-readable, set on a refusal.
	Reason string
}

// Refused is a convenience for the common caller branch.
func (r PageRoleResult) Refused() bool { return r.Outcome == PageRoleRefused }

// Created reports whether a NEW page row came into existence.
func (r PageRoleResult) Created() bool { return r.Outcome == PageRoleCreated }

// columnNamePattern guards the only interpolated part of the statement. Column
// names at every call site are string literals, so this is defence in depth
// rather than a live sanitiser — but a shared seam that builds SQL should refuse
// anything it cannot prove is a bare identifier.
var columnNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// reservedPageColumns are owned by the helper and rejected in Columns, so a
// caller cannot re-introduce the bug by passing its own page_type.
var reservedPageColumns = map[string]bool{
	"site_id": true, "name": true, "page_type": true, "id": true, "updated_at": true,
}

// UpsertPageForRole writes a page for an arm that owns a constant role, and
// resolves a name collision explicitly instead of silently. See the file comment
// for the contract; the four outcomes are on PageRoleResult.Outcome.
func UpsertPageForRole(ctx context.Context, db *sql.DB, req PageRoleUpsert, logger *zap.Logger) (PageRoleResult, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	if err := req.validate(); err != nil {
		return PageRoleResult{}, err
	}

	// CREATE, or nothing. DO NOTHING rather than DO UPDATE is the whole point:
	// the conflict becomes a branch this code takes deliberately, instead of a
	// partial update the statement performs on its own.
	cols := []string{"site_id", "name", "page_type"}
	vals := []string{"$1", "$2", "$3"}
	args := []interface{}{req.SiteID, req.Name, req.PageType}
	for _, c := range req.Columns {
		args = append(args, c.Value)
		cols = append(cols, c.Name)
		vals = append(vals, placeholder(len(args), c.Cast))
	}

	insert := fmt.Sprintf(`
		INSERT INTO pages (%s)
		VALUES (%s)
		ON CONFLICT (site_id, name) DO NOTHING
		RETURNING id`,
		strings.Join(cols, ", "), strings.Join(vals, ", "))

	var pageID uuid.UUID
	err := db.QueryRowContext(ctx, insert, args...).Scan(&pageID)
	switch {
	case err == nil:
		return PageRoleResult{PageID: pageID, Outcome: PageRoleCreated}, nil
	case errors.Is(err, sql.ErrNoRows):
		// The name is taken. Everything from here is one transaction.
		return resolvePageRoleConflict(ctx, db, req, logger)
	default:
		return PageRoleResult{}, fmt.Errorf("create %s page %q: %w", req.PageType, req.Name, err)
	}
}

func (req PageRoleUpsert) validate() error {
	if req.SiteID == uuid.Nil {
		return fmt.Errorf("UpsertPageForRole: site_id is required")
	}
	if req.Name == "" {
		return fmt.Errorf("UpsertPageForRole: page name is required")
	}
	if req.PageType == "" {
		// An empty role would make every collision look like a role change and
		// send working pages to the refusal queue.
		return fmt.Errorf("UpsertPageForRole: page_type is required (it is the role this arm owns)")
	}
	seen := make(map[string]bool, len(req.Columns))
	for _, c := range req.Columns {
		if !columnNamePattern.MatchString(c.Name) {
			return fmt.Errorf("UpsertPageForRole: %q is not a bare column identifier", c.Name)
		}
		if reservedPageColumns[c.Name] {
			return fmt.Errorf("UpsertPageForRole: column %q is owned by the helper and must not be passed", c.Name)
		}
		if seen[c.Name] {
			return fmt.Errorf("UpsertPageForRole: column %q passed twice", c.Name)
		}
		if c.Cast != "" && !columnNamePattern.MatchString(strings.TrimSuffix(c.Cast, "[]")) {
			return fmt.Errorf("UpsertPageForRole: %q is not a bare type name", c.Cast)
		}
		seen[c.Name] = true
	}
	for _, r := range req.Refresh {
		if !seen[r] {
			return fmt.Errorf("UpsertPageForRole: Refresh names %q, which is not one of the written columns", r)
		}
	}
	return nil
}

func placeholder(n int, cast string) string {
	if cast == "" {
		return fmt.Sprintf("$%d", n)
	}
	return fmt.Sprintf("$%d::%s", n, cast)
}

// resolvePageRoleConflict runs the three collision branches under one
// transaction, with the row locked FOR UPDATE so a concurrent arm cannot read
// the same row and reach the opposite decision.
func resolvePageRoleConflict(ctx context.Context, db *sql.DB, req PageRoleUpsert, logger *zap.Logger) (PageRoleResult, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return PageRoleResult{}, fmt.Errorf("begin conflict resolution for %q: %w", req.Name, err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// "Has this page ever shipped?" is answered by the estate's ONE definition,
	// datahelpers.NeverDeployedPagePredicate, negated — never by a predicate spelled
	// out here. See the CORRECTION in the file header: the hand-rolled version this
	// replaced added `|| build_status == "needs_rebuild"`, which is wrong on the 11
	// rows that are needs_rebuild and were never built.
	var (
		pageID       uuid.UUID
		existingType string
		existingBld  string
		existingStat string
		hasShipped   bool
	)
	if qerr := tx.QueryRowContext(ctx, `
		SELECT id, COALESCE(page_type, ''), COALESCE(build_status, ''),
		       COALESCE(status, ''), NOT (`+datahelpers.NeverDeployedPagePredicate+`)
		FROM pages WHERE site_id = $1 AND name = $2
		FOR UPDATE
	`, req.SiteID, req.Name).Scan(&pageID, &existingType, &existingBld, &existingStat, &hasShipped); qerr != nil {
		return PageRoleResult{}, fmt.Errorf("resolve conflicting page %q: %w", req.Name, qerr)
	}

	result := PageRoleResult{PageID: pageID, ExistingType: existingType, ExistingBuild: existingBld}

	switch {
	case existingType == req.PageType:
		// The page already does this arm's job. Refreshing the caller's declared
		// subset is what these arms have always done, and it is right.
		if uerr := updatePageColumns(ctx, tx, req, req.Refresh, false); uerr != nil {
			return PageRoleResult{}, uerr
		}
		result.Outcome = PageRoleRefreshed
		if existingStat != "" && existingStat != "active" {
			// Not a behaviour change — a same-role write into a non-active row
			// has always landed here. It is logged because content written into
			// an archived page is invisible, and nothing else says so
			// (cf. bugs_open/098: archiving does not retract the live page).
			logger.Warn("UpsertPageForRole: refreshed a page that is not active",
				zap.String("page_name", req.Name), zap.String("status", existingStat),
				zap.String("page_type", req.PageType), zap.String("source", req.Source))
		}

	case hasShipped:
		refusal, ferr := refusePageRoleConflict(ctx, tx, req, pageID, existingType, existingBld, logger)
		if ferr != nil {
			return PageRoleResult{}, ferr
		}
		result = refusal
		result.PageID = pageID

	case req.AdoptUnshippedRows:
		// Never served, different role, and this arm has DECLARED the authority:
		// it takes the row over completely, page_type included. A partial adopt
		// would leave a fresh hybrid.
		if uerr := updatePageColumns(ctx, tx, req, columnNames(req.Columns), true); uerr != nil {
			return PageRoleResult{}, uerr
		}
		result.Outcome = PageRoleAdopted
		logger.Info("UpsertPageForRole: adopted an unshipped page into this role",
			zap.String("page_name", req.Name),
			zap.String("was_page_type", existingType),
			zap.String("now_page_type", req.PageType),
			zap.String("build_status", existingBld),
			zap.String("source", req.Source))

	default:
		// Never served, different role, and the arm did NOT declare adoption.
		// Refuse rather than refresh: a refresh that leaves page_type alone is
		// exactly the partial update bugs_open/175 is about, and doing it here
		// because a field was left at its default would reintroduce the bug
		// through the back door.
		//
		// No work item is filed. `mistyped_deployed_page` is a decision about a
		// LIVE artefact — which page holds a role a visitor can reach — and this
		// row has never been served, so filing one would put noise in a queue a
		// human reads. The reason names the missing declaration, which is what
		// makes this diagnosable instead of mysterious.
		result.Outcome = PageRoleRefused
		result.Reason = fmt.Sprintf(
			"page %q exists as page_type=%q and has never been served; %s wants %q but did not set AdoptUnshippedRows, so the row was left untouched (architecture_review/RFC_010 decision 1)",
			req.Name, existingType, sourceOrUnknown(req.Source), req.PageType)
		logger.Warn("UpsertPageForRole: refused — unshipped page of another role, and this arm has not declared AdoptUnshippedRows",
			zap.String("page_name", req.Name),
			zap.String("existing_page_type", existingType),
			zap.String("wanted_page_type", req.PageType),
			zap.String("build_status", existingBld),
			zap.String("source", req.Source))
	}

	if cerr := tx.Commit(); cerr != nil {
		return PageRoleResult{}, fmt.Errorf("commit conflict resolution for %q: %w", req.Name, cerr)
	}
	committed = true
	return result, nil
}

func columnNames(cols []PageColumn) []string {
	out := make([]string, 0, len(cols))
	for _, c := range cols {
		out = append(out, c.Name)
	}
	return out
}

// updatePageColumns writes the named subset of the caller's columns. withRole
// adds page_type — only the ADOPT branch passes true, so no other path can
// re-type a page.
func updatePageColumns(ctx context.Context, tx *sql.Tx, req PageRoleUpsert, names []string, withRole bool) error {
	byName := make(map[string]PageColumn, len(req.Columns))
	for _, c := range req.Columns {
		byName[c.Name] = c
	}

	args := []interface{}{req.SiteID, req.Name}
	sets := make([]string, 0, len(names)+2)
	if withRole {
		args = append(args, req.PageType)
		sets = append(sets, fmt.Sprintf("page_type = $%d", len(args)))
	}
	for _, n := range names {
		c, ok := byName[n]
		if !ok {
			// validate() forbids this; a panic-free guard keeps a future edit
			// from silently writing fewer columns than it names.
			return fmt.Errorf("UpsertPageForRole: column %q is not among the written columns", n)
		}
		args = append(args, c.Value)
		sets = append(sets, fmt.Sprintf("%s = %s", c.Name, placeholder(len(args), c.Cast)))
	}
	if len(sets) == 0 {
		return nil // caller asked for nothing to be refreshed
	}
	sets = append(sets, "updated_at = NOW()")

	_, err := tx.ExecContext(ctx, fmt.Sprintf(
		`UPDATE pages SET %s WHERE site_id = $1 AND name = $2`, strings.Join(sets, ", ")), args...)
	if err != nil {
		return fmt.Errorf("update page %q: %w", req.Name, err)
	}
	return nil
}

// refusePageRoleConflict mutates the page not at all and routes the decision to a
// person. It is bugs_closed/081's answer, generalised: re-typing a live page
// changes what it serves the moment it happens, and 081 measured that the choice
// of WHICH page holds a role cannot be made from the data — on robot-hands.com the
// real news listing and the gripper-catalog index were byte-identical at
// sections=["news-listing"]. Guessing wrong breaks a working page, which is worse
// than the loop.
func refusePageRoleConflict(
	ctx context.Context, tx *sql.Tx, req PageRoleUpsert, pageID uuid.UUID,
	existingType, existingBuild string, logger *zap.Logger,
) (PageRoleResult, error) {
	reason := fmt.Sprintf(
		"page %q is live as page_type=%q (build_status=%q); %s wants to write it as %q. Re-typing a live page changes what it serves, so it is not done automatically (bugs_open/175)",
		req.Name, existingType, existingBuild, sourceOrUnknown(req.Source), req.PageType)

	itemFiled, err := fileMistypedLivePageItem(ctx, tx, mistypedPageConflict{
		SiteID:       req.SiteID,
		Domain:       req.Domain,
		PageID:       pageID,
		PageName:     req.Name,
		ExistingType: existingType,
		WantedType:   req.PageType,
		Source:       sourceOrUnknown(req.Source),
		Bug:          "bugs_open/175",
	}, logger)
	if err != nil {
		return PageRoleResult{}, err
	}

	if req.OriginalItemID != nil {
		if _, uerr := tx.ExecContext(ctx, `
			UPDATE site_work_items
			SET status = 'blocked', error = $2, updated_at = NOW()
			WHERE id = $1
		`, *req.OriginalItemID, reason); uerr != nil {
			return PageRoleResult{}, fmt.Errorf("block originating item for %q: %w", req.Name, uerr)
		}
	}

	logger.Warn("UpsertPageForRole: refused — a live page holds this name under a different role",
		zap.String("domain", req.Domain),
		zap.String("page_name", req.Name),
		zap.String("existing_page_type", existingType),
		zap.String("existing_build_status", existingBuild),
		zap.String("wanted_page_type", req.PageType),
		zap.String("source", req.Source),
		zap.String("page_id", pageID.String()),
		zap.Bool("item_created", itemFiled))

	return PageRoleResult{
		PageID:        pageID,
		Outcome:       PageRoleRefused,
		ExistingType:  existingType,
		ExistingBuild: existingBuild,
		ItemFiled:     itemFiled,
		Reason:        reason,
	}, nil
}

func sourceOrUnknown(s string) string {
	if s == "" {
		return "an unnamed arm"
	}
	return s
}

// mistypedPageConflict is the shared shape of the human decision filed by BOTH
// refusal sites — this helper and applyNewPage's resolveNewPageConflict. One
// item_type and one item_key shape means a gap-plan refusal and a tool-deploy
// refusal on the same page dedupe onto ONE open decision rather than two.
type mistypedPageConflict struct {
	SiteID       uuid.UUID
	Domain       string
	PageID       uuid.UUID
	PageName     string
	ExistingType string
	WantedType   string
	Source       string
	Bug          string
}

// fileMistypedLivePageItem files the `mistyped_deployed_page` decision.
//
// The item is `needs_human_review` and carries NO handler agent, ON PURPOSE:
// nothing can resolve it but a person. Measured 2026-08-01 —
// claim_work_item_action.go and load_work_item_actions.go both claim
// `status IN ('triaged','approved')`, so such a row cannot be picked up by the
// dispatch loop and re-run.
//
// `recurrenceExpected` is left false deliberately: this is a DETECTED DEFECT, not
// an action request. While the decision is outstanding the row is non-terminal,
// so idx_swi_dedup blocks a duplicate; if a human resolves it and the collision
// returns, that genuinely is "we fixed this and it came back", which is what the
// two-strike label is for.
func fileMistypedLivePageItem(ctx context.Context, tx *sql.Tx, c mistypedPageConflict, logger *zap.Logger) (bool, error) {
	bug := c.Bug
	if bug == "" {
		bug = "bugs_open/175"
	}
	spec, _ := json.Marshal(map[string]interface{}{
		"page_name":     c.PageName,
		"existing_type": c.ExistingType,
		"wanted_type":   c.WantedType,
		"domain":        c.Domain,
		"source":        c.Source,
		"bug":           bug,
		"resolution": fmt.Sprintf(
			"If this page should hold the %q role: UPDATE pages SET page_type='%s' WHERE id='%s'. If it should not, the site needs a different page for that role.",
			c.WantedType, c.WantedType, c.PageID),
	})

	summary := fmt.Sprintf("Deployed page %q occupies the %q role under page_type=%q — needs a human decision",
		c.PageName, c.WantedType, c.ExistingType)

	pageID := c.PageID
	filed, err := insertWorkItem(ctx, tx, workItem{
		siteID:   c.SiteID,
		source:   c.Source,
		pipeline: "build",
		itemType: "mistyped_deployed_page",
		severity: "medium",
		summary:  summary,
		spec:     string(spec),
		pageID:   &pageID,
		priority: 40,
		status:   "needs_human_review",
		// handlerAgent deliberately empty — see the doc comment.
		createdBy: c.Source,
		itemKey:   fmt.Sprintf("mistyped_deployed_page:%s:%s", c.PageName, c.WantedType),
	}, logger)
	if err != nil {
		return false, fmt.Errorf("file mistyped_deployed_page for %q: %w", c.PageName, err)
	}
	return filed, nil
}
