package main

// server.go — HTTP surface. Session cookie in, JSON out.
//
// Every authenticated route goes through requireAccount, which is the ONLY
// place a session is turned into an account id. Handlers never read the cookie
// themselves, so there is no second, subtly different auth path to get wrong.

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const sessionCookie = "noted_session"

type Server struct {
	Store          *Store
	SecureCookies  bool
	SessionTTL     time.Duration
	MaxUploadBytes int64
	// B2 nil = media bytes live in Postgres (the pre-2026-08-25 behaviour).
	B2 *B2
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("POST /api/register", s.register)
	mux.HandleFunc("POST /api/login", s.login)
	mux.HandleFunc("POST /api/logout", s.logout)
	mux.HandleFunc("GET /api/me", s.auth(s.me))

	mux.HandleFunc("GET /api/notes", s.auth(s.listNotes))
	mux.HandleFunc("POST /api/notes", s.auth(s.saveNote))
	mux.HandleFunc("DELETE /api/notes/{id}", s.auth(s.deleteNote))

	mux.HandleFunc("POST /api/notes/{id}/media", s.auth(s.uploadMedia))
	// Chunked uploads for files one request cannot carry (server_uploads.go).
	mux.HandleFunc("POST /api/notes/{id}/media/uploads", s.auth(s.beginUpload))
	mux.HandleFunc("PUT /api/uploads/{id}/parts/{n}", s.auth(s.uploadPart))
	mux.HandleFunc("POST /api/uploads/{id}/finish", s.auth(s.finishUpload))
	mux.HandleFunc("DELETE /api/uploads/{id}", s.auth(s.abortUpload))
	mux.HandleFunc("GET /api/media/{id}", s.auth(s.getMedia))
	mux.HandleFunc("DELETE /api/media/{id}", s.auth(s.deleteMedia))
	mux.HandleFunc("PATCH /api/media/{id}", s.auth(s.updateMedia))

	mux.HandleFunc("DELETE /api/account", s.auth(s.deleteAccount))

	mux.HandleFunc("POST /api/import", s.auth(s.importBackup))

	return logging(mux)
}

// --- plumbing -------------------------------------------------------------

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		// Deliberately no request bodies and no note content in the log. This
		// process handles people's private writing; a log line is a copy of it
		// that outlives every control we put on the database.
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

type ctxKeyAccount struct{}

func (s *Server) auth(h func(http.ResponseWriter, *http.Request, *Account)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookie)
		if err != nil || c.Value == "" {
			writeErr(w, http.StatusUnauthorized, "not signed in")
			return
		}
		acct, err := s.Store.AccountForSession(r.Context(), c.Value)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "not signed in")
			return
		}
		h(w, r, acct)
	}
}

func (s *Server) setSession(w http.ResponseWriter, r *http.Request, accountID int64) error {
	token, err := s.Store.CreateSession(r.Context(), accountID, s.SessionTTL)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true, // JS cannot read it, so an XSS cannot lift the session
		Secure:   s.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(s.SessionTTL),
	})
	return nil
}

// --- auth handlers --------------------------------------------------------

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if err := s.Store.DB.PingContext(r.Context()); err != nil {
		// Report unhealthy when the database is unreachable. A health endpoint
		// that only proves the process is running tells the monitor nothing
		// about whether anyone's notes can actually be saved.
		writeErr(w, http.StatusServiceUnavailable, "database unreachable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&c); err != nil {
		writeErr(w, http.StatusBadRequest, "could not read that request")
		return
	}
	c.Email = strings.TrimSpace(c.Email)
	if !strings.Contains(c.Email, "@") || len(c.Email) < 3 {
		writeErr(w, http.StatusBadRequest, "that does not look like an email address")
		return
	}
	// A length floor and nothing else. Composition rules push people towards
	// Password1! and away from length, which is the thing that actually helps.
	if len([]rune(c.Password)) < 10 {
		writeErr(w, http.StatusBadRequest, "please use a password of at least 10 characters")
		return
	}

	acct, err := s.Store.CreateAccount(r.Context(), c.Email, c.Password)
	if errors.Is(err, ErrEmailTaken) {
		writeErr(w, http.StatusConflict, "there is already an account with that email address")
		return
	}
	if err != nil {
		log.Printf("register failed: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not create the account")
		return
	}
	if err := s.setSession(w, r, acct.ID); err != nil {
		log.Printf("session create failed: %v", err)
		writeErr(w, http.StatusInternalServerError, "account created but sign-in failed — please sign in")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"email": acct.Email})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&c); err != nil {
		writeErr(w, http.StatusBadRequest, "could not read that request")
		return
	}
	acct, err := s.Store.Authenticate(r.Context(), c.Email, c.Password)
	if err != nil {
		// One message for both "no such account" and "wrong password". Telling
		// them apart lets anyone enumerate who has an account here, and on a
		// notes product the mere fact of an account is worth protecting.
		writeErr(w, http.StatusUnauthorized, "that email address and password do not match")
		return
	}
	if err := s.setSession(w, r, acct.ID); err != nil {
		writeErr(w, http.StatusInternalServerError, "could not start a session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"email": acct.Email})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		_ = s.Store.DeleteSession(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: s.SecureCookies, SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "signed out"})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request, a *Account) {
	// Effective per-account limits (override, or the env default) — reporting
	// them here is what lets the editor refuse an oversized file BEFORE the
	// upload, and route a large one down the chunked path, without hardcoding
	// numbers that can drift from the truth. small_upload_max is the
	// single-request cap: at or under it, the plain POST; above it, chunks.
	quota, maxUpload, err := s.Store.EffectiveLimits(r.Context(), a.ID, s.Store.QuotaBytes, s.MaxUploadBytes)
	if err != nil {
		log.Printf("me: limits read failed: %v", err)
		quota, maxUpload = s.Store.QuotaBytes, s.MaxUploadBytes
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"email":            a.Email,
		"media_bytes":      a.MediaBytes,
		"media_quota":      quota,
		"max_upload":       maxUpload,
		"small_upload_max": s.MaxUploadBytes,
	})
}

// --- notes ----------------------------------------------------------------

func (s *Server) listNotes(w http.ResponseWriter, r *http.Request, a *Account) {
	notes, err := s.Store.ListNotes(r.Context(), a.ID)
	if err != nil {
		log.Printf("list notes failed: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not load your notes")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"notes": notes})
}

func (s *Server) saveNote(w http.ResponseWriter, r *http.Request, a *Account) {
	var n Note
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&n); err != nil {
		writeErr(w, http.StatusBadRequest, "could not read that note")
		return
	}
	// The board layout rides the note save; cap it so a runaway client cannot
	// bloat rows (a real arrangement is a few KB).
	if len(n.Layout) > 256*1024 {
		writeErr(w, http.StatusBadRequest, "that note's board layout is too large")
		return
	}
	saved, err := s.Store.SaveNote(r.Context(), a.ID, n)
	if errors.Is(err, ErrNoAccount) {
		writeErr(w, http.StatusNotFound, "that note does not exist")
		return
	}
	if err != nil {
		log.Printf("save note failed: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not save that note")
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

func (s *Server) deleteNote(w http.ResponseWriter, r *http.Request, a *Account) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad note id")
		return
	}
	if err := s.Store.DeleteNote(r.Context(), a.ID, id); err != nil {
		writeErr(w, http.StatusNotFound, "that note does not exist")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- media ----------------------------------------------------------------

func (s *Server) uploadMedia(w http.ResponseWriter, r *http.Request, a *Account) {
	noteID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad note id")
		return
	}
	kind := r.URL.Query().Get("kind")
	if kind != "audio" && kind != "image" && kind != "video" {
		writeErr(w, http.StatusBadRequest, "kind must be audio, image or video")
		return
	}
	mime := r.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	body := http.MaxBytesReader(w, r.Body, s.MaxUploadBytes)
	data, err := readAll(body)
	if err != nil {
		writeErr(w, http.StatusRequestEntityTooLarge, "that file is too large")
		return
	}
	if len(data) == 0 {
		writeErr(w, http.StatusBadRequest, "empty upload")
		return
	}

	var id int64
	if s.B2 != nil {
		// B2 first (no DB state yet), THEN the quota transaction — the account
		// row lock is held for milliseconds, not for the upload. If the quota
		// refuses, the object is deleted in the same breath, so a refusal
		// cannot leak paid storage. The key is logged BEFORE the upload so a
		// crash between upload and insert leaves a greppable orphan.
		rnd := make([]byte, 16)
		if _, rerr := rand.Read(rnd); rerr != nil {
			writeErr(w, http.StatusInternalServerError, "could not save that file")
			return
		}
		key := fmt.Sprintf("media/acct_%d/%s", a.ID, hex.EncodeToString(rnd))
		sum := sha1.Sum(data)
		log.Printf("b2 upload begin key=%s bytes=%d", key, len(data))
		fileID, uerr := s.B2.Upload(key, mime, data, hex.EncodeToString(sum[:]))
		if uerr != nil {
			log.Printf("b2 upload failed key=%s: %v", key, uerr)
			writeErr(w, http.StatusBadGateway, "could not save that file")
			return
		}
		id, err = s.Store.AddMediaB2(r.Context(), a.ID, noteID, kind, mime, len(data), key, fileID)
		if err != nil {
			if derr := s.B2.Delete(key, fileID); derr != nil {
				log.Printf("b2 orphan: refused insert but delete failed key=%s: %v", key, derr)
			}
		}
	} else {
		id, err = s.Store.AddMedia(r.Context(), a.ID, noteID, kind, mime, data)
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
		log.Printf("media upload failed: %v", err)
		writeErr(w, http.StatusInternalServerError, "could not save that file")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "byte_len": len(data)})
}

func (s *Server) getMedia(w http.ResponseWriter, r *http.Request, a *Account) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad media id")
		return
	}
	mime, data, storageKey, err := s.Store.GetMedia(r.Context(), a.ID, id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if storageKey != "" && s.B2 != nil {
		res, berr := s.B2.Download(storageKey, r.Header.Get("Range"))
		if berr != nil {
			log.Printf("b2 download failed key=%s: %v", storageKey, berr)
			writeErr(w, http.StatusBadGateway, "could not fetch that file")
			return
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusPartialContent {
			log.Printf("b2 download key=%s answered %d", storageKey, res.StatusCode)
			writeErr(w, http.StatusNotFound, "not found")
			return
		}
		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Disposition", "inline")
		w.Header().Set("Accept-Ranges", "bytes")
		for _, h := range []string{"Content-Length", "Content-Range"} {
			if v := res.Header.Get(h); v != "" {
				w.Header().Set(h, v)
			}
		}
		w.WriteHeader(res.StatusCode)
		_, _ = io.Copy(w, res.Body)
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "private, no-store")
	// Anything served back as a download must not be rendered as markup by the
	// browser: an uploaded file is attacker-controlled bytes.
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", "inline")
	// ServeContent for the Range handling: without it a <video> can play but
	// never seek. Zero modtime on purpose — Cache-Control forbids caching, so
	// a Last-Modified would only invite conditional requests it then refuses.
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data))
}

func (s *Server) deleteMedia(w http.ResponseWriter, r *http.Request, a *Account) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad media id")
		return
	}
	storageKey, fileID, err := s.Store.MediaStorage(r.Context(), a.ID, id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if storageKey != "" && s.B2 != nil {
		// Object first, row second: if B2 refuses, the row stays, the UI says
		// NOT removed, and a retry converges ("already gone" counts as done).
		if derr := s.B2.Delete(storageKey, fileID); derr != nil {
			log.Printf("b2 delete failed key=%s: %v", storageKey, derr)
			writeErr(w, http.StatusBadGateway, "could not remove that file — please try again")
			return
		}
	}
	if err := s.Store.DeleteMedia(r.Context(), a.ID, id); err != nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) updateMedia(w http.ResponseWriter, r *http.Request, a *Account) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad media id")
		return
	}
	var in struct {
		Caption *string `json:"caption"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&in); err != nil || in.Caption == nil {
		writeErr(w, http.StatusBadRequest, "could not read that caption")
		return
	}
	if len([]rune(*in.Caption)) > 500 {
		writeErr(w, http.StatusBadRequest, "a caption can be at most 500 characters")
		return
	}
	if err := s.Store.SetMediaCaption(r.Context(), a.ID, id, *in.Caption); err != nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "saved", "caption": *in.Caption})
}

// --- account deletion (immediate — owner ruling 2026-08-25) ----------------

func (s *Server) deleteAccount(w http.ResponseWriter, r *http.Request, a *Account) {
	// The cookie is not enough for an irreversible action: the password typed
	// again proves the person at the keyboard, not just the browser.
	var in struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, "could not read that request")
		return
	}
	if _, err := s.Store.Authenticate(r.Context(), a.Email, in.Password); err != nil {
		writeErr(w, http.StatusUnauthorized, "that password is not right — your account is untouched")
		return
	}

	// B2 objects BEFORE rows: a row without an object is a 404; an object
	// without a row is invisible, paid storage. "Already gone" counts as
	// success, so a retry after a half-done attempt converges.
	refs, err := s.Store.ListAccountMediaStorage(r.Context(), a.ID)
	if err != nil {
		log.Printf("delete account %d: list storage failed: %v", a.ID, err)
		writeErr(w, http.StatusInternalServerError, "could not delete the account — nothing has been removed")
		return
	}
	if len(refs) > 0 && s.B2 == nil {
		// Refusing beats orphaning: with the B2 keys absent the objects could
		// not be removed, and a deletion that silently kept files would make
		// the privacy promise false.
		log.Printf("delete account %d: %d B2 objects but no B2 client configured", a.ID, len(refs))
		writeErr(w, http.StatusInternalServerError, "could not delete the account — nothing has been removed")
		return
	}
	for _, ref := range refs {
		if err := s.B2.Delete(ref.Key, ref.FileID); err != nil {
			log.Printf("delete account %d: b2 delete %s failed: %v", a.ID, ref.Key, err)
			writeErr(w, http.StatusBadGateway,
				"could not remove your files — nothing has been deleted; please try again")
			return
		}
	}

	// Unfinished chunked uploads hold bytes too (B2 bills their parts).
	// "Goodbye only after everything is gone" includes them: cancel each
	// before the rows cascade, with the same refuse-not-orphan posture.
	pending, err := s.Store.AccountPendingUploads(r.Context(), a.ID)
	if err != nil {
		log.Printf("delete account %d: pending uploads list failed: %v", a.ID, err)
		writeErr(w, http.StatusInternalServerError, "could not delete the account — nothing has been removed")
		return
	}
	for _, p := range pending {
		if s.B2 == nil {
			log.Printf("delete account %d: pending upload but no B2 client configured", a.ID)
			writeErr(w, http.StatusInternalServerError, "could not delete the account — nothing has been removed")
			return
		}
		if err := s.B2.CancelLargeFile(p.B2FileID); err != nil {
			log.Printf("delete account %d: cancel pending upload %s failed: %v", a.ID, p.B2FileID, err)
			writeErr(w, http.StatusBadGateway,
				"could not remove your files — nothing has been deleted; please try again")
			return
		}
	}

	if err := s.Store.DeleteAccount(r.Context(), a.ID); err != nil {
		log.Printf("delete account %d: %v", a.ID, err)
		writeErr(w, http.StatusInternalServerError, "could not delete the account")
		return
	}
	log.Printf("account %d deleted (%d B2 objects removed)", a.ID, len(refs))
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: s.SecureCookies, SameSite: http.SameSiteLaxMode,
	})
	// ONLY now — everything is actually gone. No optimistic goodbyes.
	writeJSON(w, http.StatusOK, map[string]string{"status": "account deleted"})
}

// --- import ---------------------------------------------------------------

// backupFile is the format the browser app has been handing users since
// 2026-08-10 ("Save everything"). Accepting exactly this file is the whole
// migration path off the local-only app, so the shape is fixed by what is
// already on people's disks — it cannot be tidied up later.
type backupFile struct {
	Format  string              `json:"format"`
	Version int                 `json:"version"`
	Notes   []backupNote        `json:"notes"`
	Audio   map[string][]string `json:"audio"`
	Images  map[string][]string `json:"images"`
}

type backupNote struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (s *Server) importBackup(w http.ResponseWriter, r *http.Request, a *Account) {
	var bf backupFile
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512<<20))
	if err := dec.Decode(&bf); err != nil {
		writeErr(w, http.StatusBadRequest, "that file could not be read as a Noted backup")
		return
	}
	// Accept the old text-only export too: it is a bare array, which fails the
	// decode above, so the client sends it under {"notes": [...]}. Both shapes
	// are in users' hands and neither can be withdrawn.
	if len(bf.Notes) == 0 {
		writeErr(w, http.StatusBadRequest, "there are no notes in that file")
		return
	}

	var notesIn, audioIn, imagesIn, skipped int
	for _, bn := range bf.Notes {
		saved, err := s.Store.SaveNote(r.Context(), a.ID, Note{
			ClientID: bn.ID, Title: bn.Title, Content: bn.Content,
		})
		if err != nil {
			log.Printf("import: note %q failed: %v", bn.ID, err)
			skipped++
			continue
		}
		notesIn++

		for kind, bucket := range map[string]map[string][]string{"audio": bf.Audio, "image": bf.Images} {
			for _, dataURL := range bucket[bn.ID] {
				mime, raw, err := decodeDataURL(dataURL)
				if err != nil {
					skipped++
					continue
				}
				if _, err := s.Store.AddMedia(r.Context(), a.ID, saved.ID, kind, mime, raw); err != nil {
					// Quota is the expected failure here, and it must not abort
					// the whole import — the notes that fit are still worth
					// having, and the count below tells the truth about the rest.
					skipped++
					continue
				}
				if kind == "audio" {
					audioIn++
				} else {
					imagesIn++
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"notes": notesIn, "recordings": audioIn, "photos": imagesIn, "skipped": skipped,
	})
}

func decodeDataURL(s string) (string, []byte, error) {
	if !strings.HasPrefix(s, "data:") {
		return "", nil, errors.New("not a data url")
	}
	comma := strings.Index(s, ",")
	if comma < 0 {
		return "", nil, errors.New("malformed data url")
	}
	header := s[5:comma]
	mime, isB64 := header, false
	if strings.HasSuffix(header, ";base64") {
		mime, isB64 = strings.TrimSuffix(header, ";base64"), true
	}
	if mime == "" {
		mime = "application/octet-stream"
	}
	if !isB64 {
		return mime, []byte(s[comma+1:]), nil
	}
	raw, err := base64.StdEncoding.DecodeString(s[comma+1:])
	return mime, raw, err
}
