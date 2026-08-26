package main

// server_uploads.go — the chunked-upload HTTP surface
// (PLAN_2026-08-26_large_uploads.md). Files above the single-request cap
// arrive as parts: begin reserves quota and opens a B2 large file, each part
// is one bounded request, finish assembles and records the media row, abort
// (or the reaper) releases everything. The honesty contract is the small
// path's, extended: the media row exists — and the editor may show STORED —
// only after finish's 2xx, and a refusal at any step leaves the account's
// stored media untouched.

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	// One part must clear the Cloudflare tunnel's ~100 MB request cap with
	// margin; B2 needs every part but the last to be at least 5 MB.
	maxPartSize = 64 << 20
	minPartSize = 5 << 20

	// Abandoned reservations (closed laptops) are released after this long.
	uploadReapAge = 24 * time.Hour
)

// partSlots bounds engine memory: at most this many part buffers exist at
// once, fleet-of-one box arithmetic — 2 × 64 MiB alongside everything else the
// box runs. Uploads beyond the bound queue briefly rather than OOMing the box.
var partSlots = make(chan struct{}, 2)

// choosePartSize picks the per-upload part size: as large as allowed, but
// never so large that the upload has fewer than 2 parts — B2 refuses a
// single-part large file (its minimum is 5 MB + 1 byte across ≥2 parts).
func choosePartSize(declared int64) int64 {
	half := (declared + 1) / 2
	size := int64(maxPartSize)
	if half < size {
		size = half
	}
	if size < minPartSize {
		size = minPartSize
	}
	return size
}

// beginUpload: POST /api/notes/{id}/media/uploads {kind, mime, size}
// B2 first, THEN the reservation — the same orphan discipline as the small
// path: if the reservation refuses, the just-started large file is cancelled
// in the same breath, so a refusal cannot leak billed storage.
func (s *Server) beginUpload(w http.ResponseWriter, r *http.Request, a *Account) {
	noteID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad note id")
		return
	}
	if s.B2 == nil {
		writeErr(w, http.StatusServiceUnavailable, "large uploads are not available right now")
		return
	}
	var in struct {
		Kind string `json:"kind"`
		Mime string `json:"mime"`
		Size int64  `json:"size"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, "could not read that request")
		return
	}
	if in.Kind != "audio" && in.Kind != "image" && in.Kind != "video" {
		writeErr(w, http.StatusBadRequest, "kind must be audio, image or video")
		return
	}
	if in.Mime == "" {
		in.Mime = "application/octet-stream"
	}
	quota, maxUpload, err := s.Store.EffectiveLimits(r.Context(), a.ID, s.Store.QuotaBytes, s.MaxUploadBytes)
	if err != nil {
		log.Printf("begin upload: limits read failed: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not start that upload")
		return
	}
	// The chunked path exists for files the single request cannot carry;
	// anything at or under that cap takes the proven small path instead.
	if in.Size <= s.MaxUploadBytes {
		writeErr(w, http.StatusBadRequest, "that file fits a normal upload — send it in one request")
		return
	}
	if in.Size > maxUpload {
		writeErr(w, http.StatusRequestEntityTooLarge, "that file is too large")
		return
	}

	rnd := make([]byte, 16)
	if _, rerr := rand.Read(rnd); rerr != nil {
		writeErr(w, http.StatusInternalServerError, "could not start that upload")
		return
	}
	key := fmt.Sprintf("media/acct_%d/%s", a.ID, hex.EncodeToString(rnd))
	partSize := choosePartSize(in.Size)
	log.Printf("b2 large upload begin key=%s bytes=%d part_size=%d", key, in.Size, partSize)
	fileID, berr := s.B2.StartLargeFile(key, in.Mime)
	if berr != nil {
		log.Printf("b2 start_large_file failed key=%s: %v", key, berr)
		writeErr(w, http.StatusBadGateway, "could not start that upload")
		return
	}
	id, err := s.Store.BeginPendingUpload(r.Context(), a.ID, noteID, in.Kind, in.Mime, in.Size, partSize, quota, key, fileID)
	if err != nil {
		if cerr := s.B2.CancelLargeFile(fileID); cerr != nil {
			log.Printf("b2 orphan: refused begin but cancel failed key=%s: %v", key, cerr)
		}
	}
	if errors.Is(err, ErrQuotaExceeded) {
		writeErr(w, http.StatusInsufficientStorage,
			"you have used all your storage for recordings and photos")
		return
	}
	if errors.Is(err, ErrNoAccount) {
		writeErr(w, http.StatusNotFound, "that note does not exist")
		return
	}
	if err != nil {
		log.Printf("begin upload: reservation failed: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not start that upload")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"upload_id":   id,
		"part_size":   partSize,
		"parts_total": (in.Size + partSize - 1) / partSize,
	})
}

// uploadPart: PUT /api/uploads/{id}/parts/{n} — raw body, exactly the expected
// size for that part number. The server computes the sha1 itself; B2 then
// re-verifies it, so a 2xx here is a byte-level receipt, not a claim.
// Re-sending a part replaces it on both sides, which is what makes the
// editor's retry safe.
func (s *Server) uploadPart(w http.ResponseWriter, r *http.Request, a *Account) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad upload id")
		return
	}
	n, err := strconv.Atoi(r.PathValue("n"))
	if err != nil || n < 1 {
		writeErr(w, http.StatusBadRequest, "bad part number")
		return
	}
	if s.B2 == nil {
		writeErr(w, http.StatusServiceUnavailable, "large uploads are not available right now")
		return
	}
	p, err := s.Store.GetPendingUpload(r.Context(), id, a.ID)
	if errors.Is(err, ErrNoUpload) {
		writeErr(w, http.StatusNotFound, "no such upload")
		return
	}
	if err != nil {
		log.Printf("upload part: read pending failed: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not save that part")
		return
	}
	if n > p.PartsTotal() {
		writeErr(w, http.StatusBadRequest, "part number beyond the declared size")
		return
	}
	expected := p.ExpectedPartSize(n)

	// The slot is held for the buffer's whole life: read, hash, send.
	partSlots <- struct{}{}
	defer func() { <-partSlots }()

	body := http.MaxBytesReader(w, r.Body, expected)
	data, err := readAll(body)
	if err != nil {
		writeErr(w, http.StatusRequestEntityTooLarge, "that part is larger than declared")
		return
	}
	if int64(len(data)) != expected {
		writeErr(w, http.StatusBadRequest,
			fmt.Sprintf("part %d must be exactly %d bytes (got %d)", n, expected, len(data)))
		return
	}
	sum := sha1.Sum(data)
	sha := hex.EncodeToString(sum[:])
	if err := s.B2.UploadPart(p.B2FileID, n, data, sha); err != nil {
		log.Printf("b2 upload_part failed key=%s n=%d: %v", p.StorageKey, n, err)
		writeErr(w, http.StatusBadGateway, "could not save that part")
		return
	}
	if err := s.Store.RecordUploadPart(r.Context(), id, a.ID, n, expected, sha); err != nil {
		// B2 has the part; our record does not. The client's retry re-sends
		// it, replacing identically — nothing is lost and nothing double-counts.
		log.Printf("upload part: record failed id=%d n=%d: %v", id, n, err)
		writeErr(w, http.StatusInternalServerError, "could not save that part")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"part": n, "size": expected})
}

// finishUpload: POST /api/uploads/{id}/finish — verifies every part is
// recorded and the sizes sum to the declared total, assembles on B2, and only
// then records the media row. Any gap refuses with the missing part named, so
// the editor can resend precisely rather than starting over.
func (s *Server) finishUpload(w http.ResponseWriter, r *http.Request, a *Account) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad upload id")
		return
	}
	if s.B2 == nil {
		writeErr(w, http.StatusServiceUnavailable, "large uploads are not available right now")
		return
	}
	p, err := s.Store.GetPendingUpload(r.Context(), id, a.ID)
	if errors.Is(err, ErrNoUpload) {
		writeErr(w, http.StatusNotFound, "no such upload")
		return
	}
	if err != nil {
		log.Printf("finish upload: read pending failed: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not finish that upload")
		return
	}
	total := p.PartsTotal()
	sha1s := make([]string, 0, total)
	var sum int64
	for n := 1; n <= total; n++ {
		part, ok := p.Parts[strconv.Itoa(n)]
		if !ok {
			writeErr(w, http.StatusConflict, fmt.Sprintf("part %d has not arrived", n))
			return
		}
		if part.Size != p.ExpectedPartSize(n) {
			writeErr(w, http.StatusConflict, fmt.Sprintf("part %d has the wrong size", n))
			return
		}
		sha1s = append(sha1s, part.Sha1)
		sum += part.Size
	}
	if sum != p.DeclaredBytes {
		writeErr(w, http.StatusConflict, "the parts do not add up to the declared size")
		return
	}
	if err := s.B2.FinishLargeFile(p.B2FileID, sha1s); err != nil {
		log.Printf("b2 finish_large_file failed key=%s: %v", p.StorageKey, err)
		writeErr(w, http.StatusBadGateway, "could not finish that upload")
		return
	}
	mediaID, err := s.Store.FinalizePendingUpload(r.Context(), id, a.ID)
	if err != nil {
		// B2 now holds a FINISHED file with no row — greppable, and the
		// account-deletion sweep and any retry of finish cannot reach it, so
		// say so loudly rather than quietly billing it.
		log.Printf("b2 orphan: finished file but finalize failed key=%s fileId=%s: %v", p.StorageKey, p.B2FileID, err)
		writeErr(w, http.StatusInternalServerError, "could not finish that upload")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": mediaID, "byte_len": p.DeclaredBytes})
}

// abortUpload: DELETE /api/uploads/{id} — row first (the reservation releases
// immediately), cancel second; a failed cancel is an orphan the reaper's
// B2-side listing finds, because B2, not our table, is the truth about what is
// pending.
func (s *Server) abortUpload(w http.ResponseWriter, r *http.Request, a *Account) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad upload id")
		return
	}
	_, fileID, err := s.Store.DeletePendingUpload(r.Context(), id, a.ID)
	if errors.Is(err, ErrNoUpload) {
		writeErr(w, http.StatusNotFound, "no such upload")
		return
	}
	if err != nil {
		log.Printf("abort upload: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not discard that upload")
		return
	}
	if s.B2 != nil {
		if cerr := s.B2.CancelLargeFile(fileID); cerr != nil {
			log.Printf("b2 orphan: abort released row but cancel failed fileId=%s: %v", fileID, cerr)
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "discarded"})
}

// ReapAbandonedUploads releases what nobody will finish: stale reservations
// (DB-side) and unfinished B2 large files with no row (crash orphans). The
// B2 listing is the authority — a tick that reads only its own bookkeeping has
// never reaped — and the age guard on rowless files is what makes the sweep
// safe against a begin whose insert is milliseconds away.
func (s *Server) ReapAbandonedUploads(ctx context.Context) {
	if s.B2 == nil {
		return
	}
	stale, err := s.Store.StalePendingUploads(ctx, uploadReapAge)
	if err != nil {
		log.Printf("upload reaper: stale list failed: %v", err)
	}
	for _, row := range stale {
		if _, _, err := s.Store.DeletePendingUpload(ctx, row.ID, row.AccountID); err != nil && !errors.Is(err, ErrNoUpload) {
			log.Printf("upload reaper: release row %d failed: %v", row.ID, err)
			continue
		}
		if err := s.B2.CancelLargeFile(row.B2FileID); err != nil {
			log.Printf("upload reaper: cancel fileId=%s failed: %v", row.B2FileID, err)
		}
	}

	unfinished, err := s.B2.ListUnfinishedLargeFiles("media/")
	if err != nil {
		log.Printf("upload reaper: b2 unfinished list failed: %v", err)
		return
	}
	known, err := s.Store.PendingB2FileIDs(ctx)
	if err != nil {
		log.Printf("upload reaper: pending ids failed: %v", err)
		return
	}
	cutoff := time.Now().Add(-uploadReapAge).UnixMilli()
	for _, f := range unfinished {
		if known[f.FileID] {
			continue // a live upload; its row governs it
		}
		if f.UploadTimestamp > cutoff {
			continue // rowless but young — possibly a begin mid-flight
		}
		if err := s.B2.CancelLargeFile(f.FileID); err != nil {
			log.Printf("upload reaper: cancel orphan %s failed: %v", f.FileName, err)
		} else {
			log.Printf("upload reaper: cancelled rowless unfinished file %s", f.FileName)
		}
	}
}
