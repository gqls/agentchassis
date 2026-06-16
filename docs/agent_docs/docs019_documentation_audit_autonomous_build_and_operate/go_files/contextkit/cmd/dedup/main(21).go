// Command dedup finds duplicate and near-duplicate files in a directory tree
// and (on request) MOVES the non-canonical copies into an archive directory,
// preserving their relative path so the move is fully reversible. It never
// deletes anything.
//
// Two passes:
//   - EXACT: SHA-256 content hash. Files with identical bytes are one group.
//     Deterministic, no configuration, no false positives.
//   - NEAR (optional, -near): within remaining files of the same extension,
//     shingled-token Jaccard similarity ≥ -threshold groups lightly-edited
//     copies (e.g. assembler.go vs assembler(2).go). Heuristic — REPORT it,
//     eyeball it, before trusting -move with -near on.
//
// Default is REPORT ONLY. Nothing moves without -move. Even with -move, a
// manifest (TSV: action, group, canonical, moved-from, moved-to) is written so
// every move is auditable and scriptable-to-undo.
//
// Canonical selection within a group (first rule that breaks the tie wins):
//   1. NOT under an archive-ish path (go_files_old/, docubundle/, _archive/, …)
//   2. path does NOT match a download-duplicate suffix  name(2).ext
//   3. shortest relative path (fewest separators, then fewest chars)
//   4. most recently modified
//   5. lexicographically first  (final determinism)
//
// Usage:
//   dedup <root>                          # report exact-dup groups
//   dedup <root> -near -threshold 0.9     # also report near-dup clusters
//   dedup <root> -ext .go,.md             # limit to these extensions
//   dedup <root> -move -archive _archive  # MOVE non-canonical copies aside
//   dedup <root> -near -move              # act on near-dups too (review first!)
//
// Scope guards: -ext limits which files are considered; the archive dir and any
// -exclude substring are never scanned (so re-runs are idempotent and the tool
// never archives its own archive).
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type fileRec struct {
	abs     string
	rel     string
	size    int64
	modTime time.Time
	hash    string
	shingle map[uint64]struct{} // populated only in -near mode
}

// archiveish paths: copies living here are never the canonical pick.
var archiveishDefault = []string{
	"go_files_old/", "go_files_old\\",
	"docubundle/", "docubundle\\",
	"_archive/", "_archive\\",
	"/old/", "\\old\\",
	"thin_slice_run/", "thin_slice_run\\",
	"scripts/documentation_project/",
}

func main() {
	var (
		move      = flag.Bool("move", false, "MOVE non-canonical copies to the archive dir (default: report only)")
		near      = flag.Bool("near", false, "also cluster near-duplicates by token similarity (heuristic)")
		threshold = flag.Float64("threshold", 0.90, "near-duplicate Jaccard similarity threshold (with -near)")
		archive   = flag.String("archive", "_archive", "archive dir (relative to root) for moved copies")
		extCSV    = flag.String("ext", "", "comma-separated extensions to consider (e.g. .go,.md); empty = all files")
		excludeS  = flag.String("exclude", "", "comma-separated path substrings to skip entirely")
		manifest  = flag.String("manifest", "dedup-manifest.tsv", "manifest path (written under root on -move)")
	)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: dedup <root> [-near] [-threshold f] [-ext .go,.md] [-move] [-archive dir] [-exclude sub,...]")
		fmt.Fprintln(os.Stderr, "  <root> may appear before or after the flags.")
	}
	// Go's flag package stops at the first NON-flag arg, so `dedup <root> -move`
	// would silently drop -move. Separate the single positional <root> from the
	// flags ourselves so order does not matter. Be value-flag-aware: the
	// argument after a value-taking flag (e.g. -threshold 0.5) is that flag's
	// value, NOT the root. -move and -near are booleans (no separate value);
	// the rest take a value, including the `-flag=value` form (one token).
	valueFlags := map[string]bool{
		"-threshold": true, "--threshold": true,
		"-archive": true, "--archive": true,
		"-ext": true, "--ext": true,
		"-exclude": true, "--exclude": true,
		"-manifest": true, "--manifest": true,
	}
	raw := os.Args[1:]
	var root string
	rest := make([]string, 0, len(raw))
	for i := 0; i < len(raw); i++ {
		a := raw[i]
		if strings.HasPrefix(a, "-") {
			rest = append(rest, a)
			// consume the value of a separate-arg value flag
			if valueFlags[a] && !strings.Contains(a, "=") && i+1 < len(raw) {
				i++
				rest = append(rest, raw[i])
			}
			continue
		}
		if root == "" {
			root = a
		} else {
			fmt.Fprintf(os.Stderr, "unexpected extra argument %q (root already %q)\n", a, root)
			flag.Usage()
			os.Exit(2)
		}
	}
	flag.CommandLine.Parse(rest)
	if root == "" {
		flag.Usage()
		os.Exit(2)
	}

	exts := splitCSVlower(*extCSV)
	excludes := splitCSV(*excludeS)
	archiveRel := filepath.Clean(*archive)

	files, err := scan(root, exts, excludes, archiveRel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "scanned %d files under %s\n", len(files), root)

	groups := groupExact(files)
	if *near {
		for i := range files {
			files[i].shingle = shingles(files[i].abs)
		}
		groups = append(groups, groupNear(remaining(files, groups), *threshold)...)
	}

	// Keep only groups with >1 member (an actual duplicate set).
	dupGroups := groups[:0]
	for _, g := range groups {
		if len(g.members) > 1 {
			dupGroups = append(dupGroups, g)
		}
	}

	if len(dupGroups) == 0 {
		fmt.Println("no duplicate groups found.")
		return
	}

	// Report.
	var toMove []moveOp
	for gi, g := range dupGroups {
		canon := pickCanonical(g.members)
		fmt.Printf("\n[%s group %d] %d files — canonical: %s\n", g.kind, gi+1, len(g.members), canon.rel)
		for _, m := range g.members {
			if m.rel == canon.rel {
				fmt.Printf("    keep    %s\n", m.rel)
			} else {
				fmt.Printf("    archive %s\n", m.rel)
				toMove = append(toMove, moveOp{group: gi + 1, kind: g.kind, canon: canon.rel, from: m})
			}
		}
	}
	fmt.Printf("\n%d duplicate group(s); %d file(s) would be archived.\n", len(dupGroups), len(toMove))

	if !*move {
		fmt.Println("\nREPORT ONLY. Re-run with -move to archive the non-canonical copies.")
		if *near {
			fmt.Println("(-near is heuristic — review the near groups above before -move.)")
		}
		return
	}

	// Act: move each non-canonical copy under <root>/<archive>/<original rel>.
	mf, err := os.Create(filepath.Join(root, *manifest))
	if err != nil {
		fmt.Fprintf(os.Stderr, "manifest: %v\n", err)
		os.Exit(1)
	}
	defer mf.Close()
	fmt.Fprintf(mf, "action\tkind\tgroup\tcanonical\tmoved_from\tmoved_to\n")

	moved := 0
	for _, op := range toMove {
		dst := filepath.Join(root, archiveRel, op.from.rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", filepath.Dir(dst), err)
			continue
		}
		if err := os.Rename(op.from.abs, dst); err != nil {
			fmt.Fprintf(os.Stderr, "move %s: %v\n", op.from.rel, err)
			continue
		}
		fmt.Fprintf(mf, "move\t%s\t%d\t%s\t%s\t%s\n", op.kind, op.group, op.canon, op.from.rel,
			filepath.Join(archiveRel, op.from.rel))
		moved++
	}
	fmt.Printf("moved %d file(s) to %s/. Manifest: %s\n", moved, archiveRel, *manifest)
	fmt.Println("To undo: for each manifest row, mv moved_to back to moved_from.")
}

type group struct {
	kind    string // "exact" | "near"
	members []fileRec
}

type moveOp struct {
	group int
	kind  string
	canon string
	from  fileRec
}

func scan(root string, exts, excludes []string, archiveRel string) ([]fileRec, error) {
	var out []fileRec
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		relSlash := filepath.ToSlash(rel)
		if info.IsDir() {
			base := info.Name()
			if strings.HasPrefix(base, ".") && base != "." {
				return filepath.SkipDir
			}
			// never descend into the archive dir (idempotent re-runs)
			if rel == archiveRel || strings.HasPrefix(relSlash, filepath.ToSlash(archiveRel)+"/") {
				return filepath.SkipDir
			}
			for _, e := range excludes {
				if e != "" && strings.Contains(relSlash, e) {
					return filepath.SkipDir
				}
			}
			return nil
		}
		for _, e := range excludes {
			if e != "" && strings.Contains(relSlash, e) {
				return nil
			}
		}
		if len(exts) > 0 {
			ext := strings.ToLower(filepath.Ext(path))
			ok := false
			for _, e := range exts {
				if ext == e {
					ok = true
					break
				}
			}
			if !ok {
				return nil
			}
		}
		h, herr := hashFile(path)
		if herr != nil {
			fmt.Fprintf(os.Stderr, "hash %s: %v\n", rel, herr)
			return nil
		}
		out = append(out, fileRec{abs: path, rel: rel, size: info.Size(), modTime: info.ModTime(), hash: h})
		return nil
	})
	return out, err
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func groupExact(files []fileRec) []group {
	byHash := map[string][]fileRec{}
	for _, f := range files {
		byHash[f.hash] = append(byHash[f.hash], f)
	}
	var gs []group
	// stable order: by first member's rel path
	keys := make([]string, 0, len(byHash))
	for k := range byHash {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return byHash[keys[i]][0].rel < byHash[keys[j]][0].rel })
	for _, k := range keys {
		gs = append(gs, group{kind: "exact", members: byHash[k]})
	}
	return gs
}

// remaining returns files that are NOT in any multi-member exact group (those
// are already handled; near-dup analysis runs on the rest).
func remaining(files []fileRec, groups []group) []fileRec {
	inDup := map[string]bool{}
	for _, g := range groups {
		if len(g.members) > 1 {
			for _, m := range g.members {
				inDup[m.abs] = true
			}
		}
	}
	var out []fileRec
	for _, f := range files {
		if !inDup[f.abs] {
			out = append(out, f)
		}
	}
	return out
}

// groupNear clusters files of the same extension whose shingle-set Jaccard
// similarity ≥ threshold. Simple transitive single-linkage: adequate for
// "a file and its (2)/(3) copies", which is the target.
func groupNear(files []fileRec, threshold float64) []group {
	used := make([]bool, len(files))
	var gs []group
	for i := range files {
		if used[i] || len(files[i].shingle) == 0 {
			continue
		}
		cluster := []fileRec{files[i]}
		used[i] = true
		for j := i + 1; j < len(files); j++ {
			if used[j] || len(files[j].shingle) == 0 {
				continue
			}
			if strings.ToLower(filepath.Ext(files[i].abs)) != strings.ToLower(filepath.Ext(files[j].abs)) {
				continue
			}
			if jaccard(files[i].shingle, files[j].shingle) >= threshold {
				cluster = append(cluster, files[j])
				used[j] = true
			}
		}
		if len(cluster) > 1 {
			gs = append(gs, group{kind: "near", members: cluster})
		}
	}
	return gs
}

// shingles builds a set of hashed 3-token shingles over the file's whitespace-
// split tokens. Order-sensitive enough to catch real similarity, cheap to compare.
func shingles(path string) map[uint64]struct{} {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	toks := strings.Fields(string(b))
	set := map[uint64]struct{}{}
	const k = 3
	if len(toks) < k {
		if len(toks) > 0 {
			set[fnv(strings.Join(toks, " "))] = struct{}{}
		}
		return set
	}
	for i := 0; i+k <= len(toks); i++ {
		set[fnv(strings.Join(toks[i:i+k], " "))] = struct{}{}
	}
	return set
}

func jaccard(a, b map[uint64]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	small, large := a, b
	if len(b) < len(a) {
		small, large = b, a
	}
	inter := 0
	for x := range small {
		if _, ok := large[x]; ok {
			inter++
		}
	}
	union := len(a) + len(b) - inter
	if union == 0 {
		return 0
	}
	return float64(inter) / float64(union)
}

func fnv(s string) uint64 {
	const off = 1469598103934665603
	const prime = 1099511628211
	h := uint64(off)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime
	}
	return h
}

func pickCanonical(members []fileRec) fileRec {
	best := members[0]
	for _, m := range members[1:] {
		if lessCanonical(m, best) {
			best = m
		}
	}
	return best
}

// lessCanonical reports whether a is a BETTER canonical pick than b.
func lessCanonical(a, b fileRec) bool {
	aa, ba := isArchiveish(a.rel), isArchiveish(b.rel)
	if aa != ba {
		return !aa // non-archiveish wins
	}
	ad, bd := isDupSuffix(a.rel), isDupSuffix(b.rel)
	if ad != bd {
		return !ad // non-(N) wins
	}
	asep, bsep := strings.Count(a.rel, string(os.PathSeparator)), strings.Count(b.rel, string(os.PathSeparator))
	if asep != bsep {
		return asep < bsep // shallower wins
	}
	if len(a.rel) != len(b.rel) {
		return len(a.rel) < len(b.rel) // shorter path wins
	}
	if !a.modTime.Equal(b.modTime) {
		return a.modTime.After(b.modTime) // newer wins
	}
	return a.rel < b.rel // final determinism
}

func isArchiveish(rel string) bool {
	s := filepath.ToSlash(rel)
	for _, a := range archiveishDefault {
		if strings.Contains(s, strings.TrimSuffix(strings.TrimSuffix(a, "/"), "\\")) {
			return true
		}
	}
	return false
}

func isDupSuffix(rel string) bool {
	base := filepath.Base(rel)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	// matches "...(<digits>)"
	if i := strings.LastIndexByte(stem, '('); i >= 0 && strings.HasSuffix(stem, ")") {
		inner := stem[i+1 : len(stem)-1]
		if inner != "" {
			allDigits := true
			for _, r := range inner {
				if r < '0' || r > '9' {
					allDigits = false
					break
				}
			}
			return allDigits
		}
	}
	return false
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func splitCSVlower(s string) []string {
	out := splitCSV(s)
	for i := range out {
		out[i] = strings.ToLower(out[i])
	}
	return out
}
