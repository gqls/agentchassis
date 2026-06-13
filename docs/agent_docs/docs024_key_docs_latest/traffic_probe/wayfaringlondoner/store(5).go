package main

// store.go — persistence for site-engine intent events, v2 (high-traffic shape).
//
// v1 kept every event in RAM and rewrote one ever-growing JSON file on every
// persist — two cliffs (write amplification + unbounded memory). v2 splits the
// data by access pattern:
//
//   EVENTS  → append-only JSONL, one line per submission, in daily files
//             (events-YYYYMMDD.jsonl). O(1) per event at any volume; nothing
//             held in RAM; rotation is just the date changing; retention is
//             "delete old files"; the future /events collector tails lines.
//   COUNTS  → a tiny snapshot (counters.json: visits + event counts per host)
//             for /stats, flushed by the debounced dirty-flag flusher.
//
// CHANGES vs v1 (per the naming rule):
//   REMOVED: Store.Events (map of full events in RAM), Store.Snapshot()
//            (nothing called it; meaningless without in-RAM events).
//   NEW:     Store.EventCounts, Store.dir/snapPath/eventsF/eventsDay,
//            openEventsFileLocked(). Snapshot file shape is now
//            {"visits":…, "event_counts":…} — a deliberate on-disk format
//            change, acceptable pre-launch.
//   KEPT:    AddVisit / AddEvent / Stats / Flush signatures (service.go
//            untouched), dirty-flag + flushInterval flusher, atomic-rename
//            snapshot writes, SIGTERM flush pairing in main.go.

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// flushInterval is how often the background flusher persists counter changes.
// Event LINES are written synchronously per event; only counters ride the
// flusher, so a hard crash loses ≤ this window of COUNTS (recomputable from
// the JSONL anyway), never an event line.
const flushInterval = 5 * time.Second

// IntentEvent is one action a visitor took on a probe page.
type IntentEvent struct {
	ID        string    `json:"id"`
	Host      string    `json:"host"`       // canonical host the event belongs to
	Kind      string    `json:"kind"`       // ProbeSearch | ProbeCategory | ProbeFreeText
	Value     string    `json:"value"`      // search term / chosen category / free text
	RefHost   string    `json:"ref_host"`   // referer reduced to host only ("" if none/same)
	LandingQuery string `json:"landing_query,omitempty"` // query string on the page the form was submitted from (inbound ?q=/?utm=… that survived)
	Country   string    `json:"country"`    // coarse CDN country code, or ""
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	mu        sync.Mutex
	dir       string   // data directory; "" = in-memory only (tests)
	snapPath  string   // dir/counters.json
	eventsF   *os.File // current day's append-only events file
	eventsDay string   // UTC yyyymmdd the open file belongs to
	dirty     bool     // unpersisted counter changes pending

	Visits      map[string]int `json:"visits"`       // host → page-load count
	EventCounts map[string]int `json:"event_counts"` // host → submissions count
}

// NewStore opens (or starts) the store in dir. The snapshot is loaded if
// present; event files are append-only and never read by the engine itself.
func NewStore(dir string) (*Store, error) {
	s := &Store{
		dir:         dir,
		Visits:      map[string]int{},
		EventCounts: map[string]int{},
	}
	if dir == "" {
		return s, nil // in-memory only (tests)
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, err
	}
	s.snapPath = filepath.Join(dir, "counters.json")
	b, err := os.ReadFile(s.snapPath)
	if err == nil {
		_ = json.Unmarshal(b, s)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if s.Visits == nil {
		s.Visits = map[string]int{}
	}
	if s.EventCounts == nil {
		s.EventCounts = map[string]int{}
	}
	go s.flushLoop()
	return s, nil
}

// openEventsFileLocked ensures the append file for the current UTC day is
// open, rotating at the day boundary. Must be called with the lock held.
func (s *Store) openEventsFileLocked(now time.Time) error {
	day := now.UTC().Format("20060102")
	if s.eventsF != nil && day == s.eventsDay {
		return nil
	}
	if s.eventsF != nil {
		_ = s.eventsF.Close()
	}
	f, err := os.OpenFile(filepath.Join(s.dir, "events-"+day+".jsonl"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	s.eventsF, s.eventsDay = f, day
	return nil
}

// persist writes the counters snapshot. Must be called with the lock held.
func (s *Store) persist() {
	if s.dir == "" {
		return
	}
	b, _ := json.Marshal(s)
	tmp := s.snapPath + ".tmp"
	if os.WriteFile(tmp, b, 0o600) == nil {
		_ = os.Rename(tmp, s.snapPath) // atomic-ish replace
	}
	s.dirty = false
}

// flushLoop persists pending counter changes at most once per flushInterval.
func (s *Store) flushLoop() {
	for range time.Tick(flushInterval) {
		s.mu.Lock()
		if s.dirty {
			s.persist()
		}
		s.mu.Unlock()
	}
}

// Flush persists pending counters and syncs the events file. Called on
// shutdown (SIGTERM/SIGINT) from main.go.
func (s *Store) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dirty {
		s.persist()
	}
	if s.eventsF != nil {
		_ = s.eventsF.Sync()
	}
}

// AddVisit counts one page load for a host (denominator for intent-per-1k).
// Counter only — persisted by the flusher.
func (s *Store) AddVisit(host string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Visits[host]++
	s.dirty = true
}

// AddEvent appends one event line to today's JSONL (O(1) regardless of
// history) and bumps the counter. The line write is synchronous; the counter
// rides the flusher.
func (s *Store) AddEvent(ev *IntentEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.EventCounts[ev.Host]++
	s.dirty = true
	if s.dir == "" {
		return // in-memory mode: counters only
	}
	if err := s.openEventsFileLocked(time.Now()); err != nil {
		return // counter still recorded; line lost — surfaced by Sync/ops
	}
	b, err := json.Marshal(ev)
	if err != nil {
		return
	}
	_, _ = s.eventsF.Write(append(b, '\n'))
}

// HostStat is the per-domain summary used by /stats and ranking.
type HostStat struct {
	Host             string  `json:"host"`
	Visits           int     `json:"visits"`
	Events           int     `json:"events"`
	EventsPer1kVisit float64 `json:"events_per_1k_visits"`
}

// Stats returns a per-host summary (no event bodies), safe to expose internally.
func (s *Store) Stats() []HostStat {
	s.mu.Lock()
	defer s.mu.Unlock()
	hosts := map[string]bool{}
	for h := range s.Visits {
		hosts[h] = true
	}
	for h := range s.EventCounts {
		hosts[h] = true
	}
	out := make([]HostStat, 0, len(hosts))
	for h := range hosts {
		v := s.Visits[h]
		e := s.EventCounts[h]
		var per float64
		if v > 0 {
			per = float64(e) * 1000 / float64(v)
		}
		out = append(out, HostStat{Host: h, Visits: v, Events: e, EventsPer1kVisit: per})
	}
	return out
}

// StreamEvents writes stored events as NDJSON lines to w, oldest first,
// original line bytes preserved: events strictly AFTER `since` (zero time =
// everything), optionally filtered to one host, capped at `limit` lines
// (<=0 = uncapped). Returns lines written and whether the cap truncated.
//
// Deliberately LOCK-FREE: a large export must never block live captures. The
// appender may be mid-write on the current file's final line, so any
// unparsable trailing line is skipped — its created_at is after any caller's
// checkpoint, so the next pull picks it up. Collector checkpoint contract:
// store the max created_at received; `since` is strictly-after, so no
// duplicates across pulls.
func (s *Store) StreamEvents(w io.Writer, since time.Time, host string, limit int) (int, bool, error) {
	if s.dir == "" {
		return 0, false, nil
	}
	names, err := filepath.Glob(filepath.Join(s.dir, "events-*.jsonl"))
	if err != nil {
		return 0, false, err
	}
	sort.Strings(names) // dated filenames sort chronologically
	sinceDay := ""
	if !since.IsZero() {
		sinceDay = since.UTC().Format("20060102")
	}
	n := 0
	nl := []byte{'\n'}
	for _, name := range names {
		day := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(name), "events-"), ".jsonl")
		if sinceDay != "" && day < sinceDay {
			continue // whole file predates the checkpoint's UTC day
		}
		f, err := os.Open(name)
		if err != nil {
			continue
		}
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			line := sc.Bytes()
			var ev IntentEvent
			if json.Unmarshal(line, &ev) != nil {
				continue // torn/garbled tail line — next pull catches it
			}
			if !since.IsZero() && !ev.CreatedAt.After(since) {
				continue
			}
			if host != "" && ev.Host != host {
				continue
			}
			if limit > 0 && n >= limit {
				f.Close()
				return n, true, nil
			}
			if _, err := w.Write(line); err != nil {
				f.Close()
				return n, false, err
			}
			if _, err := w.Write(nl); err != nil {
				f.Close()
				return n, false, err
			}
			n++
		}
		f.Close()
	}
	return n, false, nil
}
