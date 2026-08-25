package main

// b2.go — Backblaze B2 native REST client, stdlib only.
//
// WHY NOT THE SDK: this binary's dependency stance is "the database driver and
// nothing else" (store.go's header), because it runs on a box that also serves
// a live commercial site. The four B2 calls the engine needs — authorize,
// get-upload-url, upload, download, delete-version — are plain HTTP and fit in
// this file.
//
// OPT-IN, DEFAULT OFF: NewB2FromEnv returns nil when NOTED_B2_KEY_ID /
// NOTED_B2_APP_KEY are absent, and every caller treats nil as "media lives in
// Postgres exactly as before". The binary rolls safely before the keys exist.
//
// The key is BUCKET-SCOPED (created 2026-08-25: bucket personae-noted-media,
// capabilities listFiles/readFiles/writeFiles/deleteFiles only). The authorize
// response names the bucket the key is allowed, and the engine reads it from
// there rather than trusting a separate env var to agree with the key.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type B2 struct {
	keyID, appKey string
	apiBase       string // override in tests; default https://api.backblazeb2.com

	mu          sync.Mutex
	authToken   string
	apiURL      string
	downloadURL string
	bucketID    string
	bucketName  string
	authedAt    time.Time

	HTTP *http.Client
}

func NewB2FromEnv() *B2 {
	keyID, appKey := os.Getenv("NOTED_B2_KEY_ID"), os.Getenv("NOTED_B2_APP_KEY")
	if keyID == "" || appKey == "" {
		return nil
	}
	base := os.Getenv("NOTED_B2_API_BASE")
	if base == "" {
		base = "https://api.backblazeb2.com"
	}
	return &B2{keyID: keyID, appKey: appKey, apiBase: base,
		HTTP: &http.Client{Timeout: 2 * time.Minute}}
}

// authorize refreshes the account auth. B2 tokens last 24h; refresh at 12h or
// on a 401 from any call.
//
// v4, NOT v2/v3: api.backblazeb2.com refuses both older versions of this call
// outright ("not currently supported on API version number N" — measured
// 2026-08-25 with the real key). v4 moved the interesting fields under
// apiInfo.storageApi and turned `allowed` into a bucket LIST.
func (b *B2) authorize() error {
	req, err := http.NewRequest("GET", b.apiBase+"/b2api/v4/b2_authorize_account", nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(b.keyID, b.appKey)
	res, err := b.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("b2 authorize: %d %s", res.StatusCode, body)
	}
	var out struct {
		AuthorizationToken string `json:"authorizationToken"`
		APIInfo            struct {
			StorageAPI struct {
				APIURL      string `json:"apiUrl"`
				DownloadURL string `json:"downloadUrl"`
				Allowed     struct {
					Buckets []struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"buckets"`
				} `json:"allowed"`
			} `json:"storageApi"`
		} `json:"apiInfo"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return err
	}
	sa := out.APIInfo.StorageAPI
	if len(sa.Allowed.Buckets) != 1 {
		// A full-account key lists no bucket restriction. Refusing here keeps
		// the engine incapable of writing outside its own bucket even if
		// someone installs the wrong key.
		return fmt.Errorf("b2 key must be scoped to exactly one bucket (got %d) — refusing to use it", len(sa.Allowed.Buckets))
	}
	b.authToken, b.apiURL, b.downloadURL = out.AuthorizationToken, sa.APIURL, sa.DownloadURL
	b.bucketID, b.bucketName = sa.Allowed.Buckets[0].ID, sa.Allowed.Buckets[0].Name
	b.authedAt = time.Now()
	return nil
}

func (b *B2) ensureAuth() (token, apiURL, dlURL, bucketID, bucketName string, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.authToken == "" || time.Since(b.authedAt) > 12*time.Hour {
		if err = b.authorize(); err != nil {
			return
		}
	}
	return b.authToken, b.apiURL, b.downloadURL, b.bucketID, b.bucketName, nil
}

func (b *B2) invalidate() { b.mu.Lock(); b.authToken = ""; b.mu.Unlock() }

// Upload sends data under the given key and returns B2's fileId.
// A fresh upload URL is fetched per call — B2 documents that upload URLs fail
// sporadically and must simply be re-fetched; one retry covers that.
func (b *B2) Upload(key, contentType string, data []byte, sha1hex string) (fileID string, err error) {
	for attempt := 0; attempt < 2; attempt++ {
		token, apiURL, _, bucketID, _, aerr := b.ensureAuth()
		if aerr != nil {
			return "", aerr
		}
		upURL, upTok, uerr := b.getUploadURL(token, apiURL, bucketID)
		if uerr != nil {
			b.invalidate()
			err = uerr
			continue
		}
		req, _ := http.NewRequest("POST", upURL, strings.NewReader(string(data)))
		req.Header.Set("Authorization", upTok)
		req.Header.Set("X-Bz-File-Name", key)
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("X-Bz-Content-Sha1", sha1hex)
		req.ContentLength = int64(len(data))
		res, derr := b.HTTP.Do(req)
		if derr != nil {
			err = derr
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<16))
		res.Body.Close()
		if res.StatusCode == 401 {
			b.invalidate()
			err = fmt.Errorf("b2 upload: 401 %s", body)
			continue
		}
		if res.StatusCode != 200 {
			err = fmt.Errorf("b2 upload: %d %s", res.StatusCode, body)
			continue
		}
		var out struct {
			FileID string `json:"fileId"`
		}
		if jerr := json.Unmarshal(body, &out); jerr != nil || out.FileID == "" {
			err = fmt.Errorf("b2 upload: malformed response %s", body)
			continue
		}
		return out.FileID, nil
	}
	return "", err
}

func (b *B2) getUploadURL(token, apiURL, bucketID string) (string, string, error) {
	payload := fmt.Sprintf(`{"bucketId":%q}`, bucketID)
	req, _ := http.NewRequest("POST", apiURL+"/b2api/v4/b2_get_upload_url", strings.NewReader(payload))
	req.Header.Set("Authorization", token)
	res, err := b.HTTP.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return "", "", fmt.Errorf("b2 get_upload_url: %d %s", res.StatusCode, body)
	}
	var out struct {
		UploadURL          string `json:"uploadUrl"`
		AuthorizationToken string `json:"authorizationToken"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", "", err
	}
	return out.UploadURL, out.AuthorizationToken, nil
}

// Download fetches the object, forwarding an optional Range header, and
// returns the response for the caller to stream from. The caller closes Body.
func (b *B2) Download(key, rangeHeader string) (*http.Response, error) {
	for attempt := 0; attempt < 2; attempt++ {
		token, _, dlURL, _, bucketName, err := b.ensureAuth()
		if err != nil {
			return nil, err
		}
		req, _ := http.NewRequest("GET", dlURL+"/file/"+bucketName+"/"+key, nil)
		req.Header.Set("Authorization", token)
		if rangeHeader != "" {
			req.Header.Set("Range", rangeHeader)
		}
		res, err := b.HTTP.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode == 401 {
			res.Body.Close()
			b.invalidate()
			continue
		}
		return res, nil
	}
	return nil, errors.New("b2 download: repeated 401")
}

// Delete removes one file version. "Already gone" counts as success, so a
// retried delete converges instead of wedging the row.
func (b *B2) Delete(key, fileID string) error {
	for attempt := 0; attempt < 2; attempt++ {
		token, apiURL, _, _, _, err := b.ensureAuth()
		if err != nil {
			return err
		}
		payload := fmt.Sprintf(`{"fileName":%q,"fileId":%q}`, key, fileID)
		req, _ := http.NewRequest("POST", apiURL+"/b2api/v4/b2_delete_file_version", strings.NewReader(payload))
		req.Header.Set("Authorization", token)
		res, derr := b.HTTP.Do(req)
		if derr != nil {
			return derr
		}
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<14))
		res.Body.Close()
		switch {
		case res.StatusCode == 200:
			return nil
		case res.StatusCode == 400 && strings.Contains(string(body), "file_not_present"):
			return nil
		case res.StatusCode == 401:
			b.invalidate()
			continue
		default:
			return fmt.Errorf("b2 delete: %d %s", res.StatusCode, body)
		}
	}
	return errors.New("b2 delete: repeated 401")
}

// LogState says at startup where media bytes will go — the honest signal line,
// greppable, so "is B2 actually on?" is a question the pod answers.
func (b *B2) LogState() {
	if b == nil {
		log.Print("media storage: POSTGRES (B2 keys not configured)")
		return
	}
	if err := func() error { _, _, _, _, _, e := b.ensureAuth(); return e }(); err != nil {
		log.Printf("media storage: B2 configured but authorize FAILED (%v) — uploads will fail loudly rather than fall back silently", err)
		return
	}
	b.mu.Lock()
	log.Printf("media storage: B2 bucket %s", b.bucketName)
	b.mu.Unlock()
}
