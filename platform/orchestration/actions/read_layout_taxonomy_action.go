// FILE: platform/orchestration/actions/read_layout_taxonomy_action.go
//
// ReadLayoutTaxonomyAction loads the current taxonomy of the layouts
// library — the set of distinct categories and the flattened set of
// distinct industry_tags across all active layouts — and returns it as
// a map for downstream prompt templates to render.
//
// Rationale: the domain-research-classifier prompt used to hardcode
// lists of valid categories and example industry_tags. That guidance
// went stale every time a new layout was seeded or forked. This action
// replaces the hardcoded lists with a live read from the library, so
// the classifier's prompt sees the current taxonomy at render time.
//
// Used by:
//   - domain-research-classifier workflow, as a step between
//     read_site_specs and classify_and_extract. The rendered prompt
//     references {{.layout_taxonomy.categories | toJSON}} and
//     {{.layout_taxonomy.industry_tags | toJSON}}.
//
// Output shape in collected_data under the step's output_field:
//
//	{
//	  "categories":    []string,  // deduped, sorted, non-empty only
//	  "industry_tags": []string,  // deduped, sorted, flattened from text[]
//	  "layout_count":  int,       // count of is_active=true layouts
//	  "captured_at":   string,    // RFC3339 timestamp
//	}
//
// Input spec: no required or optional inputs — action queries the
// layouts table directly. The empty spec is still registered so the
// action dispatcher validates calls consistently with every other
// registered action.

package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ReadLayoutTaxonomyInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec(
		"read_layout_taxonomy",
		ReadLayoutTaxonomyInputSpec,
	)
}

// ReadLayoutTaxonomyAction is the workflow entry point.
//
// Implementation notes:
//   - Reads are deterministically ordered (ORDER BY on the deduped values)
//     so the rendered prompt is stable across runs when the library hasn't
//     changed. Matters for LLM prompt caching and for reproducibility
//     when debugging a specific classification decision.
//   - Only layouts with is_active = true are included, so draft or
//     retired layouts don't leak into the taxonomy the classifier sees.
//   - Non-empty filters on category and tags (IS NOT NULL AND <> ”)
//     guard against legacy rows that may predate those columns being
//     populated or that have accidental blanks.
//   - Slices are initialised with make([]string, 0) rather than var-
//     declared, so a library with zero matches still JSON-marshals as
//     []  rather than null when the template renders with toJSON.
//   - layout_count is best-effort; a failure to read it does not fail
//     the action.
//   - captured_at is pre-formatted as an RFC3339 string so the prompt
//     rendering doesn't surface Go's default time.Time Stringer format.
func ReadLayoutTaxonomyAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "read_layout_taxonomy"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- Categories ----------------------------------------------------------
	categories := make([]string, 0)
	catRows, err := params.DB.QueryContext(ctx, `
		SELECT DISTINCT category
		FROM layouts
		WHERE is_active = true
		  AND category IS NOT NULL
		  AND category <> ''
		ORDER BY category
	`)
	if err != nil {
		return nil, fmt.Errorf("read_layout_taxonomy: query categories: %w", err)
	}
	defer catRows.Close()

	for catRows.Next() {
		var c string
		if err := catRows.Scan(&c); err != nil {
			return nil, fmt.Errorf("read_layout_taxonomy: scan category: %w", err)
		}
		categories = append(categories, c)
	}
	if err := catRows.Err(); err != nil {
		return nil, fmt.Errorf("read_layout_taxonomy: iterate categories: %w", err)
	}

	// --- Industry tags -------------------------------------------------------
	// unnest() flattens the text[] column into rows so DISTINCT works
	// across arrays. Empty strings are filtered out for the same reason
	// as category — legacy rows shouldn't pollute the taxonomy.
	tags := make([]string, 0)
	tagRows, err := params.DB.QueryContext(ctx, `
		SELECT DISTINCT t
		FROM layouts l, unnest(l.industry_tags) AS t
		WHERE l.is_active = true
		  AND t IS NOT NULL
		  AND t <> ''
		ORDER BY t
	`)
	if err != nil {
		return nil, fmt.Errorf("read_layout_taxonomy: query industry_tags: %w", err)
	}
	defer tagRows.Close()

	for tagRows.Next() {
		var t string
		if err := tagRows.Scan(&t); err != nil {
			return nil, fmt.Errorf("read_layout_taxonomy: scan industry_tag: %w", err)
		}
		tags = append(tags, t)
	}
	if err := tagRows.Err(); err != nil {
		return nil, fmt.Errorf("read_layout_taxonomy: iterate industry_tags: %w", err)
	}

	// --- Count (best-effort) -------------------------------------------------
	var layoutCount int
	if err := params.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM layouts WHERE is_active = true`,
	).Scan(&layoutCount); err != nil {
		logger.Warn("read_layout_taxonomy: layout_count query failed, proceeding with 0",
			zap.Error(err),
		)
		layoutCount = 0
	}

	logger.Info("read_layout_taxonomy: loaded taxonomy",
		zap.Int("categories", len(categories)),
		zap.Int("industry_tags", len(tags)),
		zap.Int("layout_count", layoutCount),
	)

	// Empty arrays are not an error — they just mean the library is
	// empty or everything is filtered out. The downstream prompt will
	// render guidance but with empty JSON arrays; the classifier falls
	// back on adoption/archetype signals. Warn so the condition shows
	// up in logs without blocking the classifier.
	if len(categories) == 0 || len(tags) == 0 {
		logger.Warn("read_layout_taxonomy: taxonomy is partially or fully empty — downstream prompt will have reduced guidance",
			zap.Int("categories", len(categories)),
			zap.Int("industry_tags", len(tags)),
		)
	}

	return map[string]interface{}{
		"categories":    categories,
		"industry_tags": tags,
		"layout_count":  layoutCount,
		"captured_at":   time.Now().UTC().Format(time.RFC3339),
	}, nil
}
