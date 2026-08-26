package main

// b2_large.go — the B2 large-file API, same stdlib-only stance as b2.go.
//
// WHY THIS EXISTS: a single-request upload tops out under ~100 MB at the
// Cloudflare tunnel (plan-level, not ours to raise), and b2.go's Upload holds
// the whole file in memory besides. Large files therefore arrive in PARTS —
// the editor slices, each part is one bounded request, and this file streams
// them into B2's large-file flow (start → upload_part×N → finish), so the
// engine never holds more than one part per in-flight request.
//
// B2's own rules, which the server-side validation mirrors: parts are 1-based,
// every part except the last must be at least 5 MB, at most 10,000 parts, and
// finish takes the sha1 of every part IN ORDER and refuses a mismatch.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// StartLargeFile opens a large-file upload and returns its fileId — the handle
// every subsequent part/finish/cancel call names. Nothing is stored until
// finish; an unfinished large file is invisible but BILLED, which is why every
// begin has a matching cancel path and the reaper treats B2, not our table, as
// the truth about what is pending.
func (b *B2) StartLargeFile(key, contentType string) (fileID string, err error) {
	for attempt := 0; attempt < 2; attempt++ {
		token, apiURL, _, bucketID, _, aerr := b.ensureAuth()
		if aerr != nil {
			return "", aerr
		}
		payload := fmt.Sprintf(`{"bucketId":%q,"fileName":%q,"contentType":%q}`, bucketID, key, contentType)
		req, _ := http.NewRequest("POST", apiURL+"/b2api/v4/b2_start_large_file", strings.NewReader(payload))
		req.Header.Set("Authorization", token)
		res, derr := b.HTTP.Do(req)
		if derr != nil {
			err = derr
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<16))
		res.Body.Close()
		if res.StatusCode == 401 {
			b.invalidate()
			err = fmt.Errorf("b2 start_large_file: 401 %s", body)
			continue
		}
		if res.StatusCode != 200 {
			err = fmt.Errorf("b2 start_large_file: %d %s", res.StatusCode, body)
			continue
		}
		var out struct {
			FileID string `json:"fileId"`
		}
		if jerr := json.Unmarshal(body, &out); jerr != nil || out.FileID == "" {
			err = fmt.Errorf("b2 start_large_file: malformed response %s", body)
			continue
		}
		return out.FileID, nil
	}
	return "", err
}

// UploadPart sends one part (1-based). A fresh part-upload URL is fetched per
// call, exactly as b2.go's Upload does for whole files and for the same
// documented reason: part URLs fail sporadically and are simply re-fetched.
// Re-sending a part number REPLACES that part on B2's side, which is what
// makes the editor's per-part retry safe.
func (b *B2) UploadPart(fileID string, partNo int, data []byte, sha1hex string) (err error) {
	for attempt := 0; attempt < 2; attempt++ {
		token, apiURL, _, _, _, aerr := b.ensureAuth()
		if aerr != nil {
			return aerr
		}
		upURL, upTok, uerr := b.getUploadPartURL(token, apiURL, fileID)
		if uerr != nil {
			b.invalidate()
			err = uerr
			continue
		}
		req, _ := http.NewRequest("POST", upURL, bytes.NewReader(data))
		req.Header.Set("Authorization", upTok)
		req.Header.Set("X-Bz-Part-Number", fmt.Sprint(partNo))
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
			err = fmt.Errorf("b2 upload_part %d: 401 %s", partNo, body)
			continue
		}
		if res.StatusCode != 200 {
			err = fmt.Errorf("b2 upload_part %d: %d %s", partNo, res.StatusCode, body)
			continue
		}
		return nil
	}
	return err
}

func (b *B2) getUploadPartURL(token, apiURL, fileID string) (string, string, error) {
	payload := fmt.Sprintf(`{"fileId":%q}`, fileID)
	req, _ := http.NewRequest("POST", apiURL+"/b2api/v4/b2_get_upload_part_url", strings.NewReader(payload))
	req.Header.Set("Authorization", token)
	res, err := b.HTTP.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return "", "", fmt.Errorf("b2 get_upload_part_url: %d %s", res.StatusCode, body)
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

// FinishLargeFile assembles the parts. sha1s MUST be in part order — B2
// re-verifies every one, so a finish that succeeds is a byte-level proof the
// parts arrived intact, which is why the server computes the sha1s itself and
// never trusts the client's.
func (b *B2) FinishLargeFile(fileID string, sha1s []string) (err error) {
	arr, _ := json.Marshal(sha1s)
	for attempt := 0; attempt < 2; attempt++ {
		token, apiURL, _, _, _, aerr := b.ensureAuth()
		if aerr != nil {
			return aerr
		}
		payload := fmt.Sprintf(`{"fileId":%q,"partSha1Array":%s}`, fileID, arr)
		req, _ := http.NewRequest("POST", apiURL+"/b2api/v4/b2_finish_large_file", strings.NewReader(payload))
		req.Header.Set("Authorization", token)
		res, derr := b.HTTP.Do(req)
		if derr != nil {
			err = derr
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<16))
		res.Body.Close()
		if res.StatusCode == 401 {
			b.invalidate()
			err = fmt.Errorf("b2 finish_large_file: 401 %s", body)
			continue
		}
		if res.StatusCode != 200 {
			err = fmt.Errorf("b2 finish_large_file: %d %s", res.StatusCode, body)
			continue
		}
		return nil
	}
	return err
}

// CancelLargeFile abandons an unfinished large file and its parts. "Already
// gone" counts as success for the same reason as Delete: a retried cancel must
// converge, not wedge.
func (b *B2) CancelLargeFile(fileID string) error {
	for attempt := 0; attempt < 2; attempt++ {
		token, apiURL, _, _, _, err := b.ensureAuth()
		if err != nil {
			return err
		}
		payload := fmt.Sprintf(`{"fileId":%q}`, fileID)
		req, _ := http.NewRequest("POST", apiURL+"/b2api/v4/b2_cancel_large_file", strings.NewReader(payload))
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
		case res.StatusCode == 400 && (strings.Contains(string(body), "file_not_present") ||
			strings.Contains(string(body), "not a valid file id")):
			return nil
		case res.StatusCode == 401:
			b.invalidate()
			continue
		default:
			return fmt.Errorf("b2 cancel_large_file: %d %s", res.StatusCode, body)
		}
	}
	return errors.New("b2 cancel_large_file: repeated 401")
}

// UnfinishedFile is one entry from ListUnfinishedLargeFiles — enough for the
// reaper to decide and act. UploadTimestamp (B2's, milliseconds) is what makes
// the orphan check safe: a file with no pending row might be a begin whose
// insert is milliseconds away, so only AGED rowless files are cancelled.
type UnfinishedFile struct {
	FileID          string
	FileName        string
	UploadTimestamp int64
}

// ListUnfinishedLargeFiles returns every unfinished large file in the bucket
// under the given prefix. THE REAPER READS THIS, NOT OUR TABLE: an unfinished
// file with no pending_uploads row (a crash between start and insert, a
// cancel that failed after the row died) is exactly the orphan only B2 can
// name. Paginates until B2 says done.
func (b *B2) ListUnfinishedLargeFiles(prefix string) ([]UnfinishedFile, error) {
	var out []UnfinishedFile
	start := ""
	for {
		token, apiURL, _, bucketID, _, err := b.ensureAuth()
		if err != nil {
			return nil, err
		}
		payload := fmt.Sprintf(`{"bucketId":%q,"namePrefix":%q,"maxFileCount":100`, bucketID, prefix)
		if start != "" {
			payload += fmt.Sprintf(`,"startFileId":%q`, start)
		}
		payload += "}"
		req, _ := http.NewRequest("POST", apiURL+"/b2api/v4/b2_list_unfinished_large_files", strings.NewReader(payload))
		req.Header.Set("Authorization", token)
		res, derr := b.HTTP.Do(req)
		if derr != nil {
			return nil, derr
		}
		body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
		res.Body.Close()
		if res.StatusCode == 401 {
			b.invalidate()
			continue
		}
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("b2 list_unfinished_large_files: %d %s", res.StatusCode, body)
		}
		var page struct {
			Files []struct {
				FileID          string `json:"fileId"`
				FileName        string `json:"fileName"`
				UploadTimestamp int64  `json:"uploadTimestamp"`
			} `json:"files"`
			NextFileID *string `json:"nextFileId"`
		}
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("b2 list_unfinished_large_files: malformed response %s", body)
		}
		for _, f := range page.Files {
			out = append(out, UnfinishedFile{FileID: f.FileID, FileName: f.FileName, UploadTimestamp: f.UploadTimestamp})
		}
		if page.NextFileID == nil || *page.NextFileID == "" {
			return out, nil
		}
		start = *page.NextFileID
	}
}
