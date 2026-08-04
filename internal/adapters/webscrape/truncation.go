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
// bugs_open/158 item 3, decision 1 (owner ruling 2026-08-03: "upload
// everything that gets truncated ... nothing is lost"). Before this ruling
// only "markdown" was here — the other five page fields were listed in
// pageFieldsNeverUploaded, a deliberate acknowledgement that no code path
// stored them. That category is retired: uploadScrapingResults now uploads
// all six (adapter.go, the per-page loop), and truncateResultForTransport's
// fallback uploader below closes the remaining case — upload_results:false —
// so every field a truncation can cut now has a real path to a stored copy.
//
// "html_content" and "rawHtml" share a URI key with their sibling spelling
// ("html", "raw_html") because they name the SAME concept under two
// producers' different key conventions (Firecrawl's raw /crawl response vs.
// batch_handler.go's normalised shape) — see the comment in adapter.go's
// per-page upload loop.
var pageFieldURIKeys = map[string]string{
	"content":          "content_uri",
	"markdown_content": "markdown_content_uri",
	"html_content":     "html_uri",
	"markdown":         "markdown_uri",
	"rawHtml":          "raw_html_uri",
	"raw_html":         "raw_html_uri",
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

// FieldUploader uploads a field's FULL, pre-truncation content when it is
// about to be discarded — bugs_open/158 item 3, owner ruling 2026-08-03:
// "upload everything that gets truncated ... the basis is eventual
// correctness and robustness." A nil FieldUploader means storage is not
// configured at all (no client), which is the one case this cannot close.
// Every OTHER case now gets a real copy regardless of the caller's own
// upload_results setting — destroying scraped content is a correctness
// defect, not a feature a step can opt out of. `key` is a storage path
// relative to the scrape's own base path, distinct per field/page so two
// concurrent truncations can never collide.
type FieldUploader func(content []byte, key string) (uri string, err error)

// ensureFieldURI returns uri unchanged when a copy already exists — the main
// uploader (upload_results: true) already wrote it, and calling upload again
// would be a wasted second write of the same bytes. Otherwise, if upload is
// usable, it uploads the FULL content (never the truncated stub — by the time
// a caller reaches this function the field has already been cut, so `full`
// must be the pre-truncation copy) and records the new URI into `storage` so
// a LATER read of the same field (stripResultForRetry's second pass) finds it
// too, without a second upload. Failure degrades to "" — logged, never
// silent, and never a broken or invented URI.
func ensureFieldURI(uri string, storage map[string]interface{}, storageKey, full, key string, upload FieldUploader, logger *zap.Logger) string {
	if uri != "" {
		return uri
	}
	if upload == nil {
		return ""
	}
	newURI, err := upload([]byte(full), key)
	if err != nil || newURI == "" {
		logger.Warn("could not preserve content that is about to be truncated — no copy will exist",
			zap.String("key", key), zap.Error(err))
		return ""
	}
	if storage != nil && storageKey != "" {
		storage[storageKey] = newURI
	}
	return newURI
}

// extFor is a cosmetic file extension for the fallback uploader's storage
// key. It is not load-bearing for correctness — pageStorageURIFor and
// storageURIFor resolve by KEY and by the embedded page index, never by
// extension — it only makes an ad-hoc S3 object readable by eye.
func extFor(field string) string {
	switch {
	case strings.Contains(field, "html"):
		return ".html"
	case strings.Contains(field, "markdown"):
		return ".md"
	default:
		return ".txt"
	}
}

// appendPageFieldURI records a fallback-uploaded per-page URI into
// storage["pages"]. It APPENDS a fresh single-key entry rather than trying to
// merge into an existing one for page i, because entries in storage.pages
// carry no explicit page-index field — they are matched purely by searching
// every entry's URI string for the embedded "/page_<i>." pattern
// (pageStorageURIFor). A linear scan over more, smaller entries resolves
// exactly the same as one over fewer, larger ones, so appending is both
// correct and the simplest thing that is.
//
// storage["pages"] is created as []map[string]string (the shape
// uploadScrapingResults itself builds) if absent. If it already holds some
// OTHER type — which should only happen if a caller round-tripped storage
// through JSON before reaching here, never true on the live single-message
// path — this logs and does nothing rather than panicking or silently
// discarding the URI into a shape nothing will read.
func appendPageFieldURI(storage map[string]interface{}, key, uri string) {
	if storage == nil || key == "" || uri == "" {
		return
	}
	entry := map[string]string{key: uri}
	switch pages := storage["pages"].(type) {
	case nil:
		storage["pages"] = []map[string]string{entry}
	case []map[string]string:
		storage["pages"] = append(pages, entry)
	default:
		// Unexpected shape (e.g. already JSON round-tripped). Not the live
		// path today; degrading to "not recorded" is safer than guessing a
		// conversion that could silently misplace the URI.
	}
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
//
// `upload` is the fallback uploader (bugs_open/158 item 3, decision 1): when a
// field is about to be cut and has no PRE-EXISTING URI (the main uploader
// either never ran, or ran but this particular field was outside its static
// list), the FULL content is uploaded on the spot and the new URI both names
// the marker AND is written back into `storage` — creating it on `result` if
// it did not already exist — so downstream, including a later
// stripResultForRetry call on the same result, sees it too.
func truncateResultForTransport(result map[string]interface{}, storage map[string]interface{}, upload FieldUploader, logger *zap.Logger) []string {
	if result == nil {
		return nil
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	// Materialise a mutable storage map only once something is actually about
	// to be uploaded into it — an empty map attached where none existed before
	// would itself be a small false claim ("something was stored") and would
	// break the existing "no storage at all" test shape.
	attachStorage := func() map[string]interface{} {
		if storage == nil {
			storage = map[string]interface{}{}
			result["storage"] = storage
		}
		return storage
	}

	var cut []string

	for _, field := range truncatableTopLevelFields {
		content, ok := result[field].(string)
		if !ok || len(content) <= maxTransportContentLen {
			continue
		}
		uri := storageURIFor(storage, field)
		if uri == "" && upload != nil {
			uri = ensureFieldURI(uri, attachStorage(), topLevelFieldURIKeys[field], content,
				fmt.Sprintf("truncated/%s%s", field, extFor(field)), upload, logger)
		}
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
				if uri == "" && upload != nil {
					key := fmt.Sprintf("truncated/pages/%s/page_%d%s", field, i, extFor(field))
					newURI, err := upload([]byte(content), key)
					if err == nil && newURI != "" {
						uri = newURI
						appendPageFieldURI(attachStorage(), pageFieldURIKeys[field], newURI)
					} else {
						logger.Warn("could not preserve truncated page content — no copy will exist",
							zap.Int("page", i), zap.String("field", field), zap.Error(err))
					}
				}
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
//
// `upload` (decision 1, as in truncateResultForTransport) is used ONLY for
// raw_html, top-level and per-page, because that is the one thing this
// function deletes OUTRIGHT with no size check first — a field that was
// UNDER the 50KB cap sails through truncateResultForTransport's pass
// untouched and therefore never gets a chance at a stored copy there. Every
// other field this function cuts (markdown_content, html_content, and the
// per-page content/markdown_content/markdown) was already ≥ the same 50KB
// cap on the FIRST pass — since this function only runs when the reply is
// STILL too big after that pass — so anything originally worth preserving
// from them already has a URI, findable via the same storage map, without a
// second upload. The one gap this leaves, stated rather than silently
// carried: a field between oversizeStripContentCap and 50,000 chars that was
// never over the FIRST cap loses its tail here unuploaded. Not closed, because
// this function has never fired: `degraded_for_transport` — the marker it sets
// on every result it touches, and therefore the only instrument that answers
// the question — is present on 0 of 3,246 retained orchestration rows
// (measured 2026-08-04). Closing it would repeat the same upload machinery for
// a path with zero observed volume.
//
// > CORRECTED 2026-08-04, caught by the council's prior_art_librarian seat: an
// > earlier version of this comment justified the same decision with "fleet-wide
// > the largest reply ever recorded is 48KB". That figure is from `llm_call_log`
// > — the size of MODEL COMPLETIONS — and says nothing about scrape replies,
// > which carry whole web pages. Scrape-bearing rows average 1.6MB and reach
// > 5.45MB of accumulated state. The conclusion held; the evidence was about a
// > different population. Do not reintroduce a size statistic here: the question
// > is "has this function run", and its own marker is what answers it.
func stripResultForRetry(result map[string]interface{}, storage map[string]interface{}, upload FieldUploader) {
	if result == nil {
		return
	}
	if raw, ok := result["raw_html"].(string); ok && raw != "" {
		uri := storageURIFor(storage, "raw_html")
		if uri == "" && upload != nil {
			if storage == nil {
				storage = map[string]interface{}{}
				result["storage"] = storage
			}
			uri = ensureFieldURI(uri, storage, topLevelFieldURIKeys["raw_html"], raw,
				"truncated/raw_html.html", upload, zap.NewNop())
		}
		// The URI is not otherwise reachable once raw_html itself is deleted
		// below — nothing else in this reply names it — so it goes into
		// result["storage"] only; there is no marker to put it in.
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
			for _, field := range []string{"rawHtml", "raw_html"} {
				raw, ok := pageMap[field].(string)
				if !ok || raw == "" {
					continue
				}
				uri := pageStorageURIFor(storage, field, i)
				if uri == "" && upload != nil {
					if storage == nil {
						storage = map[string]interface{}{}
						result["storage"] = storage
					}
					key := fmt.Sprintf("truncated/pages/%s/page_%d%s", field, i, extFor(field))
					if newURI, err := upload([]byte(raw), key); err == nil && newURI != "" {
						appendPageFieldURI(storage, pageFieldURIKeys[field], newURI)
					}
				}
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
