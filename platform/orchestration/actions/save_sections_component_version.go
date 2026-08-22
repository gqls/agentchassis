package actions

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RFC_046 (owner ruling 2026-08-22) — stamp component identity at the point of
// production instead of inferring it.
//
// THE PROBLEM THIS IS HALF OF. A page_components row carries an identity
// (component_id, slot_name), the bytes it serves, and the content those bytes
// were made from — and nothing asserts the three agree. Identity is re-derived
// by inference at every hop: from a data-component attribute, from a sentinel
// meaning unknown, from POSITION in a plan, from fuzzy name matching, and from
// slot-name equality during a rebuild. bugs_open/357 is the bill: 22 live rows
// (as of 2026-08-22) declare themselves the shared `hero` while storing a whole
// interactive tool, because `hero` is first in those pages' plans.
//
// WHY THE VERSION AND NOT JUST THE COMPONENT. component_id already records which
// component was ASKED to render. It does not record which template text actually
// ran, and the two stop agreeing the moment a template is edited. That gap is not
// hypothetical: bugs_closed/277 closed with 15 rows permanently unrepairable
// because the templates that produced their bytes no longer exist anywhere —
// `component_versions` held zero rows for any of the nine components involved.
// A row that names its version can always be re-rendered, or shown to have
// drifted. A row that does not is a guess with a timestamp.
//
// REUSE, NOT ADDITION. page_components.component_version_id already exists and is
// purpose-built for this. Measured 2026-08-22: 0 of 1,930 rows populated, no Go
// code writing it, no Go code reading it — dormant machinery, wired here rather
// than duplicated. (⚠ site_plan_sections has an identically-named column which
// IS live and accounts for nearly every grep hit; qualify by table.)
//
// THIS FILE ONLY WRITES. Nothing reads component_version_id, so this change is
// inert by construction: it cannot alter what any page serves. Making readers
// stop guessing is a separate, behaviour-changing round — and saying otherwise
// would be the "detected but never blocked" dishonesty the council gate caught
// in this lane's round 2.

// componentVersionSource is what a section knows about where its bytes came
// from. Exactly one of these is normally set; NEITHER is the honest and common
// case for content this platform did not render itself.
type componentVersionSource struct {
	// carriedVersionID is a stamp an earlier render already earned, travelling
	// with bytes that were carried rather than regenerated.
	carriedVersionID string
	// renderedSHA is the digest of the template text RenderTemplate executed.
	renderedSHA string
}

// resolveComponentVersionID turns what a section knows about its own provenance
// into a component_versions row id, or into nothing.
//
// It returns ("", false) far more readily than it returns an id, and that is the
// design rather than a limitation. The only outcomes that produce a stamp are the
// two where provenance is a FACT: a carried stamp, or a digest that still matches
// the component's current template. Everything else — a template edited between
// render and save, a component that has since been deleted, bytes nobody rendered
// — is unknown, and unknown is written as NULL.
//
// NEVER infer a version from the component alone (e.g. "take its newest"). That
// would reproduce, one layer down, precisely the defect RFC_046 exists to end:
// a confident answer manufactured from something that merely correlates.
func resolveComponentVersionID(
	ctx context.Context,
	db *sql.DB,
	componentID uuid.UUID,
	src componentVersionSource,
	logger *zap.Logger,
) (string, bool) {
	if db == nil {
		return "", false
	}

	// A carried stamp is already a fact about these exact bytes. The bytes did
	// not change, so neither did what produced them — re-resolving would be a
	// chance to be wrong for no gain.
	if src.carriedVersionID != "" {
		if _, err := uuid.Parse(src.carriedVersionID); err == nil {
			return src.carriedVersionID, true
		}
		logger.Warn("component version: carried stamp is not a uuid, dropping it rather than guessing",
			zap.String("carried", src.carriedVersionID))
	}

	if src.renderedSHA == "" || componentID == uuid.Nil {
		return "", false
	}

	var currentTemplate string
	var schemaJSON []byte
	err := db.QueryRowContext(ctx, `
		SELECT html_template, input_schema FROM content_components WHERE id = $1
	`, componentID).Scan(&currentTemplate, &schemaJSON)
	if err != nil {
		logger.Warn("component version: could not read the component's template; storing no stamp",
			zap.String("component_id", componentID.String()), zap.Error(err))
		return "", false
	}

	// The render happened before this save. If the template has changed since,
	// the digest is a fact about a template text we no longer hold — we cannot
	// create a version row for bytes we cannot reproduce, and we must not point
	// at the new one, which did NOT produce these bytes.
	sum := sha256.Sum256([]byte(currentTemplate))
	if hex.EncodeToString(sum[:]) != src.renderedSHA {
		logger.Info("component version: template changed between render and save; storing no stamp",
			zap.String("component_id", componentID.String()))
		return "", false
	}

	// Get-or-create. The template text is the identity here, not the version
	// number: two saves of the same unchanged template must resolve to the SAME
	// row, or every render mints a version and the table becomes a log.
	var versionID string
	err = db.QueryRowContext(ctx, `
		SELECT id FROM component_versions
		WHERE component_id = $1 AND html_template = $2
		ORDER BY version_number DESC LIMIT 1
	`, componentID, currentTemplate).Scan(&versionID)
	if err == nil {
		return versionID, true
	}
	if err != sql.ErrNoRows {
		logger.Warn("component version: lookup failed; storing no stamp",
			zap.String("component_id", componentID.String()), zap.Error(err))
		return "", false
	}

	// First time this exact text has been rendered through here. ON CONFLICT
	// covers the race with a concurrent save on another pod — there is a UNIQUE
	// (component_id, version_number), and this tree runs many sessions and many
	// pods at once, so the losing insert must be a no-op rather than an error.
	err = db.QueryRowContext(ctx, `
		INSERT INTO component_versions
			(component_id, version_number, html_template, input_schema, change_description, change_source)
		VALUES (
			$1,
			(SELECT COALESCE(MAX(version_number), 0) + 1 FROM component_versions WHERE component_id = $1),
			$2, $3,
			'observed at render — provenance stamp, not an edit (RFC_046)',
			'render_stamp'
		)
		ON CONFLICT (component_id, version_number) DO NOTHING
		RETURNING id
	`, componentID, currentTemplate, schemaJSON).Scan(&versionID)
	if err == nil {
		return versionID, true
	}

	// Lost the race (or the insert was a no-op): the row the winner wrote is the
	// one we want, and it is keyed by the same template text.
	if err == sql.ErrNoRows {
		if reErr := db.QueryRowContext(ctx, `
			SELECT id FROM component_versions
			WHERE component_id = $1 AND html_template = $2
			ORDER BY version_number DESC LIMIT 1
		`, componentID, currentTemplate).Scan(&versionID); reErr == nil {
			return versionID, true
		}
	}
	logger.Warn("component version: could not record the rendered template; storing no stamp",
		zap.String("component_id", componentID.String()), zap.Error(err))
	return "", false
}
