// FILE: platform/orchestration/actions/copy_gate_annotation_test.go
//
// The wrapper's contract is "adds keys, changes nothing else". These tests pin
// both halves of that, including the arms where it must keep its hands off: an
// error, a non-map result, and clean copy.
//
// Mutation checks (by hand, recorded in the lane NOTES):
//   - drop the `if err != nil` early return -> TestAnnotationPassesErrorsThrough fails
//   - count from copy_gate_findings instead of re-scanning content_data in
//     annotatePageNegation -> TestPageAnnotationCountsTemplateSectionsToo fails

package actions

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
)

func annotationParams() ActionParams {
	return ActionParams{Logger: zap.NewNop(), CollectedData: map[string]interface{}{}}
}

func TestSectionAnnotationAttachesFindings(t *testing.T) {
	inner := func(ctx context.Context, p ActionParams) (interface{}, error) {
		return map[string]interface{}{
			"rendered_html":      "<section>…</section>",
			"component_function": "call-to-action",
			"content_data": map[string]interface{}{
				"headline": "The registry shows you what's possible, not what survives production.",
				"cta_url":  "/contact.html",
			},
		}, nil
	}
	out, err := annotateSectionNegation(inner)(context.Background(), annotationParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := out.(map[string]interface{})
	findings, ok := res["copy_gate_findings"].([]map[string]interface{})
	if !ok || len(findings) == 0 {
		t.Fatalf("expected copy_gate_findings, got %v", res["copy_gate_findings"])
	}
	if findings[0]["field"] != "headline" || findings[0]["headline"] != true {
		t.Errorf("finding should name the headline field and mark it headline-class: %v", findings[0])
	}
	if res["rendered_html"] != "<section>…</section>" {
		t.Error("the wrapper must not touch the rendered html")
	}
}

func TestSectionAnnotationSilentOnCleanCopy(t *testing.T) {
	inner := func(ctx context.Context, p ActionParams) (interface{}, error) {
		return map[string]interface{}{
			"content_data": map[string]interface{}{
				"headline": "Every agent definition running in our production fleet",
				"body":     "<p>We run 1,600 orchestrations a day across 13 live systems.</p>",
			},
		}, nil
	}
	out, _ := annotateSectionNegation(inner)(context.Background(), annotationParams())
	if _, present := out.(map[string]interface{})["copy_gate_findings"]; present {
		t.Error("clean copy must attach no key at all — presence of the key is the signal")
	}
}

func TestAnnotationPassesErrorsThrough(t *testing.T) {
	want := errors.New("component \"x\" is missing required content field(s)")
	inner := func(ctx context.Context, p ActionParams) (interface{}, error) { return nil, want }
	out, err := annotateSectionNegation(inner)(context.Background(), annotationParams())
	if !errors.Is(err, want) {
		t.Errorf("the wrapper must not swallow or alter the action's error, got %v", err)
	}
	if out != nil {
		t.Errorf("expected the inner result to pass through untouched, got %v", out)
	}
	// A handler that returns something other than a map must not panic the wrapper.
	strInner := func(ctx context.Context, p ActionParams) (interface{}, error) { return "not a map", nil }
	if o, e := annotateSectionNegation(strInner)(context.Background(), annotationParams()); e != nil || o != "not a map" {
		t.Errorf("non-map result must pass through: %v/%v", o, e)
	}
}

// The page count re-scans content_data rather than summing the per-section
// annotations, so a section rendered from a template (which never carries an
// annotation) is still counted. This is that arm.
func TestPageAnnotationCountsTemplateSectionsToo(t *testing.T) {
	inner := func(ctx context.Context, p ActionParams) (interface{}, error) {
		return map[string]interface{}{
			"page_name":     "model-directory",
			"section_count": 3,
			"sections_metadata": []map[string]interface{}{
				{ // annotated section
					"stored_slot_name": "hero",
					"content_data": map[string]interface{}{
						"headline": "Multi-agent systems deployed to production in days, not months",
					},
					"copy_gate_findings": []map[string]interface{}{{"field": "headline"}},
				},
				{ // template-rendered: content_data present, no annotation key
					"stored_slot_name": "listing",
					"content_data": map[string]interface{}{
						"intro": "This list is pulled from the production registry, not from provider marketing pages.",
					},
				},
				{ // clean
					"stored_slot_name": "call-to-action",
					"content_data": map[string]interface{}{
						"headline": "Book a call with the people who run the pipeline",
					},
				},
			},
		}, nil
	}
	out, err := annotatePageNegation(inner)(context.Background(), annotationParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := out.(map[string]interface{})
	if got := res["copy_gate_page_hits"]; got != 2 {
		t.Errorf("expected 2 page hits (one of them from an UNannotated template section), got %v", got)
	}
	fields, _ := res["copy_gate_page_fields"].([]string)
	if len(fields) != 2 {
		t.Fatalf("expected two field labels, got %v", fields)
	}
	if fields[0] != "0:hero:headline" || fields[1] != "1:listing:intro" {
		t.Errorf("field labels should locate the hit by section index, slot and field: %v", fields)
	}
}

func TestPageAnnotationSilentOnCleanPage(t *testing.T) {
	inner := func(ctx context.Context, p ActionParams) (interface{}, error) {
		return map[string]interface{}{
			"sections_metadata": []map[string]interface{}{
				{"content_data": map[string]interface{}{"headline": "Agent definitions running in production"}},
			},
		}, nil
	}
	out, _ := annotatePageNegation(inner)(context.Background(), annotationParams())
	if _, present := out.(map[string]interface{})["copy_gate_page_hits"]; present {
		t.Error("a clean page must attach no count key")
	}
}
