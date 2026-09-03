// FILE: platform/orchestration/actions/content_write_stamp.go
//
// bugs_open/355 candidate A1 — writes to page_components announce WHO is
// writing, so the archive triggers' application_name column records the call
// site instead of the connection's socket identity.
//
// WHY set_config, AND WHY IT MUST SIT INSIDE A TRANSACTION.
// page_component_history.application_name is filled by
// current_setting('application_name') inside the 357/552 archive triggers.
// Today every application-side row reads `app - <ip>:<port>` — pgbouncer's
// application_name_add_host stamp, one value per CONNECTION, several writers
// per binary — so it attributes nothing (LANDMINES: "application_name looks
// like it names the writer and is a SOCKET"). Two facts dictate the form:
//
//   - SET / SET LOCAL take no bind parameters in the extended protocol;
//     set_config('application_name', $1, true) is the parameterised
//     equivalent, and is_local=true scopes it to the CURRENT TRANSACTION —
//     on an autocommit statement it is a statement-scoped no-op, which is
//     why the helper below opens a transaction rather than issuing two
//     bare statements (which the pool could also route to different
//     server connections).
//   - the pool sits behind pgbouncer in transaction mode
//     (pgbouncer-configmap: pool_mode=transaction), so a session-level SET
//     would leak the stamp onto OTHER clients' work on the shared server
//     connection. Transaction-scoped is the only safe form: the trigger
//     sees it, and it dies at COMMIT before the pool can hand the
//     connection to anyone else.
//
// BEST-EFFORT BY CONSTRUCTION: attribution must never break the write. If the
// wrapping transaction cannot be opened, or the stamp statement itself errors,
// the write runs unstamped on the plain connection — the archive row then
// carries the socket identity, exactly what it carried before A1 existed.
//
// ACCEPTANCE PREDICATE (bugs_open/355 §6): nothing else in this codebase sets
// application_name — verified 2026-08-22, and any `action:*` / `admin-edit` /
// `cli:*` value appearing in page_component_history.application_name is
// therefore provably this mechanism working:
//
//	SELECT DISTINCT application_name FROM page_component_history
//	 WHERE created_at > now() - interval '1 day';
//
// WRITERS DELIBERATELY NOT STAMPED, so nobody "finishes" this later:
//
//   - save_page_sections — the funnel. Its archive rows are already
//     discriminated by op='delete', and its archive/DELETE/INSERT run as
//     separate autocommit statements today; wrapping the whole save in one
//     transaction changes the funnel's crash behaviour and is out of A1's
//     scope (recorded in bugs_open/355).
//   - the INSERT-only writers (deploy_tool, create_tool_component, the
//     insert arms of create_report_page / rebuild_blog_listing /
//     webdesignport import) — INSERT is never archived (357 and 552 fire on
//     UPDATE and DELETE only), so there is no archive row to attribute.
package actions

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
)

// Writer names, one per stamped call site. Grep any of these against
// page_component_history.application_name to see that writer's archive rows.
const (
	contentWriterSectionEditorUpdate = "action:section_editor.update"
	contentWriterSectionEditorSwap   = "action:section_editor.swap"
	contentWriterCreateReportPage    = "action:create_report_page"
	contentWriterRebuildBlogListing  = "action:rebuild_blog_listing"
	contentWriterApplyAdoptionPlan   = "action:apply_adoption_plan"
	contentWriterToolRegenerate      = "action:create_tool_component.regenerate"
	// features_open/035 P1 direction 2. Its archive rows are the only evidence
	// that a child edit propagated upward, so they must name this writer rather
	// than the socket — and it is a distinct name from section_editor.update
	// deliberately: the child write and the ancestor rewrite are two different
	// decisions in one action, and an archive that conflated them could not
	// answer "who rewrote the parent".
	contentWriterRecomposeAncestors = "action:recompose_ancestors"
)

// stampWriterSQL is the one statement both helpers issue. Kept as a const so
// the test can assert the exact form — the broken sketch this replaced
// (`SET LOCAL application_name = $1`) was also one plausible-looking line.
const stampWriterSQL = `SELECT set_config('application_name', $1, true)`

// stampedExecContext runs one write statement with the caller's name announced
// as application_name for the duration of a wrapping transaction. For call
// sites that today write on the bare autocommit connection: the wrap adds no
// second write statement, so atomicity of the write itself is unchanged.
//
// Fallback order, per the best-effort contract in the file header:
// Begin fails → run unstamped; stamp fails → roll the empty tx back and run
// unstamped. The write's error, and its sql.Result (RowsAffected is read by
// the lock-refusal paths), pass through untouched.
func stampedExecContext(ctx context.Context, db *sql.DB, writer, query string, args ...interface{}) (sql.Result, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return db.ExecContext(ctx, query, args...)
	}
	if _, err := tx.ExecContext(ctx, stampWriterSQL, writer); err != nil {
		_ = tx.Rollback()
		return db.ExecContext(ctx, query, args...)
	}
	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return res, nil
}

// stampWriterTx announces the writer inside an ALREADY-OPEN transaction — for
// the call sites that hold one (apply_adoption_plan, the tool regeneration).
// One call after BeginTx covers every archived write in that transaction.
//
// The error is logged, never returned: the only realistic failure is a dead
// connection, which the caller's next statement will surface as itself, and
// a stamping failure must not abort work the caller has already begun. (The
// statement is parameterised and syntactically fixed, so "invalid statement"
// — the silent-failure mode a discarded error would hide — cannot occur.)
func stampWriterTx(ctx context.Context, tx *sql.Tx, writer string, logger *zap.Logger) {
	if _, err := tx.ExecContext(ctx, stampWriterSQL, writer); err != nil && logger != nil {
		logger.Warn("content write stamp failed — archive rows from this transaction keep the socket identity (bugs_open/355 A1)",
			zap.String("writer", writer), zap.Error(err))
	}
}
