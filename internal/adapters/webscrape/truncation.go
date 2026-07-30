// FILE: internal/adapters/webscrape/truncation.go
package webscrape

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// Transport truncation for the single-URL scrape path.
//
// bugs_open/133 defect A: this path cut three content fields at 50,000 chars and
// appended
//
//	[Content truncated for Kafka transport - full version in S3]
//
// unconditionally, justified by a comment reading "Full content is already in S3
// (via storage URIs)". The upload that makes that true is guarded by
// `if uploadResults && a.storageClient != nil`, and four of the six live
// single-URL scrape steps do not set `upload_results`. So the reader — human or
// model — was told a copy existed when the bytes had just been destroyed.
// MEASURED 2026-07-28: one scrape discarded 3,805 characters of
// vetcomparison.uk and pointed at a file that was never written.
//
// THE FIX IS STRUCTURAL, NOT A REWORDED STRING. A claim of a stored copy is now
// only reachable by NAMING the URI, and the URI is resolved per field from what
// the upload actually recorded. There is deliberately no way to spell "full
// version in S3" without one in hand, so a later edit cannot reintroduce the lie
// without deleting a parameter.
//
// Per-field resolution is not gold-plating: uploadScrapingResults uploads each
// field separately and best-effort — every one is `logger.Warn(...)` on failure
// and then carries on — so even with `upload_results: true` a single failed
// upload left the old marker lying about that field.

// maxTransportContentLen caps each content-bearing field in a reply. The whole
// message still has to fit the broker's limit, which this cap does not
// guarantee on its own (nothing bounds `links`, `metadata`, or the NUMBER of
// pages) — that is what the degrade-or-error path in sendSuccessResponse is
// for. The two mechanisms are complements, not alternatives.
const maxTransportContentLen = 50000 // 50KB per field

// topLevelFieldURIKeys maps each truncatable top-level field to the key under
// `storage` where uploadScrapingResults records its URI.
//
// This is one of TWO lists that have to agree — the other is the set of fields
// actually truncated. Two lists that must agree is the drift class this repo
// keeps getting bitten by (bugs_closed/144 was two hand-written traversals
// disagreeing), so TestEveryTruncatedFieldHasAURIRuling fails if a field is
// added to one and not the other. Mapping a field to the wrong key would put
// ANOTHER field's URI in this field's marker, which is the same false claim
// this bug is about, so the test also asserts each key is one the uploader
// really writes.
var topLevelFieldURIKeys = map[string]string{
	"markdown_content": "markdown_uri",
	"html_content":     "html_uri",
	"raw_html":         "raw_html_uri",
}

// pageFieldURIKeys maps truncatable per-page fields to their URI key within
// `storage.pages[]`.
//
// Only `markdown` is here, and the asymmetry is real rather than an oversight:
// the uploader stores `pageMap["html"]` and `pageMap["markdown"]`, while the
// truncator cuts content/markdown_content/html_content/markdown/rawHtml/raw_html.
// `markdown` is the only overlap. The other five have no stored copy at any
// setting of upload_results, so their marker can never truthfully claim one —
// pageFieldsNeverUploaded records that as a decision rather than leaving it to
// be rediscovered.
var pageFieldURIKeys = map[string]string{
	"markdown": "markdown_uri",
}

// pageFieldsNeverUploaded are per-page fields that are truncated but which
// uploadScrapingResults never stores. Listed explicitly so the drift test can
// tell "no URI is possible" apart from "somebody forgot to map it".
var pageFieldsNeverUploaded = map[string]bool{
	"content":          true,
	"markdown_content": true,
	"html_content":     true,
	"rawHtml":          true,
	"raw_html":         true,
}

// truncatableTopLevelFields is the order in which top-level fields are cut.
var truncatableTopLevelFields = []string{"markdown_content", "html_content", "raw_html"}

// truncatablePageFields is the order in which per-page fields are cut.
var truncatablePageFields = []string{"content", "markdown_content", "html_content", "markdown", "rawHtml", "raw_html"}

// transportTruncationMarker returns the text appended to a field cut down for
// transport.
//
// A stored copy can only be CLAIMED by NAMING it. Passing "" — no URI, because
// nothing was uploaded or that field's upload failed — produces a marker saying
// the remainder is gone, which is the truth in that case. This function is the
// only place either sentence exists.
func transportTruncationMarker(uri string) string {
	if uri == "" {
		return "\n\n[Content truncated for Kafka transport at " +
			fmt.Sprint(maxTransportContentLen) +
			" chars - the remainder was DISCARDED and no copy was stored]"
	}
	return "\n\n[Content truncated for Kafka transport at " +
		fmt.Sprint(maxTransportContentLen) +
		" chars - full version stored at " + uri + "]"
}

// storageInfoOf returns the upload record merged onto a result by
// handleMessage (`resultMap["storage"] = uploadedResult`), or nil when nothing
// was uploaded.
//
// The result map is the single source: the truncator and the degrader both read
// the URIs from the same place a downstream consumer does, so a marker cannot
// name a URI the reply does not also carry.
func storageInfoOf(result map[string]interface{}) map[string]interface{} {
	if result == nil {
		return nil
	}
	storage, _ := result["storage"].(map[string]interface{})
	return storage
}

// storageURIFor returns the URI recorded for a top-level field, or "" when the
// upload did not happen or did not record one.
func storageURIFor(storage map[string]interface{}, field string) string {
	key, mapped := topLevelFieldURIKeys[field]
	if !mapped || storage == nil {
		return ""
	}
	uri, _ := storage[key].(string)
	return uri
}

// pageStorageURIFor returns the URI recorded for one field of page `index`.
//
// It finds the page by the index embedded IN the URI rather than by position in
// storage.pages, because that list is COMPACTED: uploadScrapingResults appends
// only pages that uploaded something (`if len(pageInfo) > 0`), so
// storage.pages[i] does not correspond to result.pages[i] whenever any page
// stored nothing. Indexing by position would attach another page's URI to this
// page's marker — a false claim reached by fixing a false claim carelessly.
//
// The URI carries its own index because the uploader builds it from
// "%s/pages/page_%d.md", so searching for "/page_<i>." is position-independent.
// If that naming convention ever changes, this returns "" and the marker
// degrades to the honest "DISCARDED" form. It cannot degrade into a wrong URI.
func pageStorageURIFor(storage map[string]interface{}, field string, index int) string {
	key, mapped := pageFieldURIKeys[field]
	if !mapped || storage == nil {
		return ""
	}
	needle := fmt.Sprintf("/page_%d.", index)
	for _, entry := range storagePageEntries(storage) {
		uri := entry[key]
		if uri != "" && strings.Contains(uri, needle) {
			return uri
		}
	}
	return ""
}

// storagePageEntries normalises storage.pages, which is a []map[string]string
// when built in-process and a []interface{} if it has been through JSON.
func storagePageEntries(storage map[string]interface{}) []map[string]string {
	switch pages := storage["pages"].(type) {
	case []map[string]string:
		return pages
	case []interface{}:
		out := make([]map[string]string, 0, len(pages))
		for _, p := range pages {
			entry := map[string]string{}
			switch m := p.(type) {
			case map[string]string:
				entry = m
			case map[string]interface{}:
				for k, v := range m {
					if s, ok := v.(string); ok {
						entry[k] = s
					}
				}
			}
			out = append(out, entry)
		}
		return out
	}
	return nil
}

// truncateResultForTransport caps every content-bearing field in a scrape
// result, marking each cut with a marker that can only claim a stored copy when
// one was actually recorded. It returns the names of the fields it cut, newest
// caller-visible signal first, and records them ON the result so a consumer can
// detect truncation without parsing English out of the content.
//
// `storage` is the map uploadScrapingResults returned, or nil when
// upload_results was false — which is exactly the case the old code got wrong.
func truncateResultForTransport(result map[string]interface{}, storage map[string]interface{}, logger *zap.Logger) []string {
	if result == nil {
		return nil
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	var cut []string

	for _, field := range truncatableTopLevelFields {
		content, ok := result[field].(string)
		if !ok || len(content) <= maxTransportContentLen {
			continue
		}
		uri := storageURIFor(storage, field)
		result[field] = content[:maxTransportContentLen] + transportTruncationMarker(uri)
		cut = append(cut, field)
		logger.Info("Truncating large field for Kafka",
			zap.String("field", field),
			zap.Int("original_len", len(content)),
			zap.Int("truncated_to", maxTransportContentLen),
			zap.Bool("stored_copy", uri != ""),
			zap.String("storage_uri", uri))
		if uri == "" {
			logger.Warn("Truncated content was DISCARDED, not stored — the reply carries less than was scraped",
				zap.String("field", field),
				zap.Int("discarded_chars", len(content)-maxTransportContentLen),
				zap.String("ref", "bugs_open/133"))
		}
	}

	// Screenshots go entirely; the URI is the useful form and the base64 is the
	// single biggest contributor to an oversized reply.
	if _, present := result["screenshot_base64"]; present {
		delete(result, "screenshot_base64")
	}

	// Multi-page results (crawl) carry content per page.
	if pages, ok := result["pages"].([]interface{}); ok {
		for i, page := range pages {
			pageMap, ok := page.(map[string]interface{})
			if !ok {
				continue
			}
			for _, field := range truncatablePageFields {
				content, ok := pageMap[field].(string)
				if !ok || len(content) <= maxTransportContentLen {
					continue
				}
				uri := pageStorageURIFor(storage, field, i)
				pageMap[field] = content[:maxTransportContentLen] + transportTruncationMarker(uri)
				cut = append(cut, fmt.Sprintf("pages[%d].%s", i, field))
				logger.Info("Truncating large page field for Kafka",
					zap.Int("page", i),
					zap.String("field", field),
					zap.Int("original_len", len(content)),
					zap.Int("truncated_to", maxTransportContentLen),
					zap.Bool("stored_copy", uri != ""),
					zap.String("storage_uri", uri))
			}
		}
	}

	if len(cut) > 0 {
		// Machine-readable, so a consumer never has to read the marker prose to
		// know the payload is short. `truncated` matches the batch path's key.
		result["truncated"] = true
		result["truncated_fields"] = cut
	}

	return cut
}

// stripResultForRetry is the degraded form of a single-URL reply, for the one
// resend after the broker refuses the full one. Raw HTML goes outright — if the
// reply did not fit WITH it, it is what did not fit — and what remains is cut
// hard.
//
// It marks the result so the degradation is visible in the reply itself: a
// caller that receives a stub must be able to tell it from a genuinely small
// page, or the next investigation starts from a page that "scraped fine".
func stripResultForRetry(result map[string]interface{}, storage map[string]interface{}) {
	if result == nil {
		return
	}
	delete(result, "raw_html")
	delete(result, "screenshot_base64")

	for _, field := range []string{"markdown_content", "html_content"} {
		if content, ok := result[field].(string); ok && len(content) > oversizeStripContentCap {
			result[field] = content[:oversizeStripContentCap] +
				transportTruncationMarker(storageURIFor(storage, field))
		}
	}

	if pages, ok := result["pages"].([]interface{}); ok {
		for i, page := range pages {
			pageMap, ok := page.(map[string]interface{})
			if !ok {
				continue
			}
			delete(pageMap, "rawHtml")
			delete(pageMap, "raw_html")
			delete(pageMap, "html_content")
			for _, field := range []string{"content", "markdown_content", "markdown"} {
				if content, ok := pageMap[field].(string); ok && len(content) > oversizeStripContentCap {
					pageMap[field] = content[:oversizeStripContentCap] +
						transportTruncationMarker(pageStorageURIFor(storage, field, i))
				}
			}
		}
	}

	result["truncated"] = true
	result["degraded_for_transport"] = true
}
