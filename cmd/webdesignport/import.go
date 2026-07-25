package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

// The importer: manifest -> pages + page_components rows.
//
// WHY GO AND A DRIVER, not psql. This writes ~97 HTML blobs, several of them
// tens of kilobytes with inline scripts, quotes, backslashes and unicode. Sent
// through psql they would ride a shell heredoc, which is precisely the
// escape-mangling trap this repo has been bitten by before (a NUL byte once
// reached a Go source file that way). As driver parameters they are bytes and
// never touch a shell.
//
// The row shapes follow create_report_page_action.go, the platform's own
// reference for an owned page: page_type set, in_header/in_footer false so the
// generic nav machinery ignores it, rebuild_policy='owned' so
// save_page_sections refuses to overwrite it, and the component instance
// upserted by (page_id, slot_name) rather than by a remembered id — page
// component ids are NOT stable across re-renders.

const portedSlot = "ported-page"

func runImport(outDir, dsn, domain string, dryRun bool, only string) error {
	if dsn == "" {
		return fmt.Errorf("--dsn is required (run over `kubectl -n ai-persona-system port-forward svc/postgres-clients 5433:5432`)")
	}

	var man Manifest
	if err := readJSON(filepath.Join(outDir, "manifest.json"), &man); err != nil {
		return fmt.Errorf("manifest: %w (run `webdesignport transform` first)", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	var siteID string
	err = db.QueryRow(`SELECT id::text FROM sites WHERE domain = $1`, domain).Scan(&siteID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no sites row for %s — the site must be onboarded first", domain)
	} else if err != nil {
		return err
	}

	// The shared component every ported page instantiates. Seeded separately so
	// this tool never invents library rows.
	var componentID string
	err = db.QueryRow(`
		SELECT id::text FROM content_components
		 WHERE function = $1 AND is_active
		 ORDER BY created_at DESC LIMIT 1`, portedSlot).Scan(&componentID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no active content_components row with function=%q — apply the seed first", portedSlot)
	} else if err != nil {
		return err
	}

	fmt.Printf("site %s (%s), component %s, %d pages in manifest%s\n\n",
		domain, siteID, componentID, len(man.Pages), map[bool]string{true: "  [DRY RUN]"}[dryRun])

	var created, updated, unchanged, locked int

	for _, p := range man.Pages {
		if only != "" && !strings.Contains(p.Name, only) {
			continue
		}

		htmlBytes, err := os.ReadFile(filepath.Join(outDir, p.Fragment))
		if err != nil {
			return fmt.Errorf("%s: %w", p.Name, err)
		}
		body := string(htmlBytes)

		// What is already there? Compare on the sha we stored last time rather
		// than on the HTML itself, so an unchanged page costs one small query
		// instead of dragging every blob back over the wire.
		var (
			existingPageID sql.NullString
			existingSHA    sql.NullString
			isLocked       sql.NullBool
		)
		err = db.QueryRow(`
			SELECT pg.id::text,
			       pc.content_data->>'sha256',
			       (pc.id IS NOT NULL AND NOT (`+agentWritable("pc.")+`))
			  FROM pages pg
			  LEFT JOIN page_components pc
			         ON pc.page_id = pg.id AND pc.slot_name = $3
			 WHERE pg.site_id = $1 AND pg.name = $2`,
			siteID, p.Name, portedSlot).Scan(&existingPageID, &existingSHA, &isLocked)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("%s: lookup: %w", p.Name, err)
		}

		switch {
		case isLocked.Valid && isLocked.Bool:
			// A human lock beats the importer. Report it rather than working
			// around it — bugs_open/058 semantics.
			fmt.Printf("  LOCKED    %-38s (human lock — skipped)\n", p.Name)
			locked++
			continue
		case existingSHA.Valid && existingSHA.String == p.SHA256:
			unchanged++
			continue
		case existingPageID.Valid:
			updated++
		default:
			created++
		}

		if dryRun {
			verb := "CREATE"
			if existingPageID.Valid {
				verb = "UPDATE"
			}
			fmt.Printf("  %-9s %-38s %6d bytes  %s\n", verb, p.Name, len(body), p.URL)
			continue
		}

		if err := upsertPage(db, siteID, componentID, p, body); err != nil {
			return fmt.Errorf("%s: %w", p.Name, err)
		}
		verb := "created"
		if existingPageID.Valid {
			verb = "updated"
		}
		fmt.Printf("  %-9s %-38s %6d bytes  %s\n", verb, p.Name, len(body), p.URL)
	}

	fmt.Printf("\n%d created, %d updated, %d unchanged, %d locked\n",
		created, updated, unchanged, locked)
	if dryRun {
		fmt.Println("dry run — nothing was written")
	}
	return nil
}

// agentWritable mirrors pageComponentAgentWritableSQL from the platform: a row
// is writable by automation iff it carries no lock, or only a timed lock that
// has already expired. Duplicated rather than imported because this is a
// standalone cmd tool, so it is spelled out here and must be kept in step.
func agentWritable(alias string) string {
	return "(" + alias + "locked_at IS NULL OR (" +
		alias + "lock_type = 'timed' AND " +
		alias + "lock_expires_at IS NOT NULL AND " +
		alias + "lock_expires_at < NOW()))"
}

func upsertPage(db *sql.DB, siteID, componentID string, p ManifestPage, body string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	sections, _ := json.Marshal([]string{portedSlot})

	// build_status='deployed' rather than 'planned': a 'planned' page is picked
	// up by write_build_items and handed to the generic build pipeline, which is
	// exactly what an owned page must never enter. in_header/in_footer stay
	// false — nav is set separately, and only for the three top-level pages.
	var pageID string
	err = tx.QueryRow(`
		INSERT INTO pages (site_id, name, url, title, page_type, status,
		                   meta_description, sections, rebuild_policy,
		                   build_status, in_header, in_footer, updated_at)
		VALUES ($1,$2,$3,$4,$5,'active',$6,$7,'owned','deployed',false,false,NOW())
		ON CONFLICT (site_id, name) DO UPDATE
		   SET url = EXCLUDED.url,
		       title = EXCLUDED.title,
		       page_type = EXCLUDED.page_type,
		       meta_description = EXCLUDED.meta_description,
		       sections = EXCLUDED.sections,
		       rebuild_policy = 'owned',
		       updated_at = NOW()
		RETURNING id::text`,
		siteID, p.Name, p.URL, p.Title, p.PageType, p.MetaDescription, sections).Scan(&pageID)
	if err != nil {
		return fmt.Errorf("upsert page: %w", err)
	}

	contentData, _ := json.Marshal(map[string]any{
		"schema":    "ported-page.v1",
		"source":    p.Source,
		"sha256":    p.SHA256,
		"generator": "webdesignport/1",
		"qa_tier":   p.QATier,
	})

	// Keyed by (page_id, slot_name) — never by a remembered page_components.id,
	// which is not stable across re-renders.
	var instanceID sql.NullString
	if err := tx.QueryRow(
		`SELECT id::text FROM page_components WHERE page_id = $1 AND slot_name = $2`,
		pageID, portedSlot).Scan(&instanceID); err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("find instance: %w", err)
	}

	if instanceID.Valid {
		res, err := tx.Exec(`
			UPDATE page_components
			   SET rendered_html = $1, content_data = $2, component_id = $3,
			       build_status = 'approved', updated_at = NOW()
			 WHERE id = $4 AND `+agentWritable(""),
			body, contentData, componentID, instanceID.String)
		if err != nil {
			return fmt.Errorf("update instance: %w", err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			// The lock predicate matched nothing: someone locked the row between
			// the pre-check and here. Fail rather than report a phantom success.
			return fmt.Errorf("component became locked mid-import — re-run")
		}
	} else {
		if _, err := tx.Exec(`
			INSERT INTO page_components (page_id, component_id, position, slot_name,
			                             rendered_html, content_data, build_status)
			VALUES ($1,$2,0,$3,$4,$5,'approved')`,
			pageID, componentID, portedSlot, body, contentData); err != nil {
			return fmt.Errorf("insert instance: %w", err)
		}
	}

	return tx.Commit()
}
