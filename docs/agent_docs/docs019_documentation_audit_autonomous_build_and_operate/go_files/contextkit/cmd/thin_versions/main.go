// Command thin_versions reduces version-sprawl in a docs tree: it groups files
// that are successive VERSIONS of the same document, keeps the newest N of each
// group, and (on request) moves the older versions into an archive dir. The
// goal is noise reduction — leave a handful of current versions per subject so
// a downstream reader (e.g. the analyser index) isn't diluted by dozens of
// stale copies.
//
// This is DISTINCT from dedup: dedup removes identical/near-identical COPIES
// (same content, (N) download duplicates). thin_versions removes older
// VERSIONS (different content, successive edits) of one document. Run dedup
// first to clear exact copies, then thin_versions to trim version history.
//
// How versions are grouped — the "subject stem" (derived per the observed,
// non-strict conventions):
//   - strip a trailing .patch/.orig/.bak marker;
//   - strip the extension;
//   - strip a trailing (N) download-duplicate bracket;
//   - strip a trailing _vX or _vX_Y version;
//   - strip a second trailing (N) (covers the _v..(N) ordering);
//   - what remains is the subject. Files with the SAME subject IN THE SAME
//     DIRECTORY are one group (the directory carries classification, so the
//     same stem in two dirs is two subjects — deliberately not merged).
//
// A leading NNN_ number is PART of the stem, so 004_foo and 005_foo are
// different subjects unless their stems match after the number — the tool does
// NOT assume number increments pair up (the user flagged that mixup case;
// it is surfaced, not auto-merged — see -report-renames).
//
// Recency rank within a group (newest first):
//
//	(_vX_Y as integers) > ((N) bracket) > file mtime
//
// Version beats mtime deliberately: v2_42(3) edited before v2_41 is still newer
// by version. mtime only breaks ties when no version/bracket is present.
//
// SAFETY (same posture as dedup):
//   - REPORT ONLY by default; nothing moves without -move.
//   - SINGLETON subjects (one file) are never touched.
//   - files already under an archive-ish path (_archive/, /old/, _old) are
//     skipped — they're already set aside.
//   - moves go to <root>/<archive>/<original rel>, structure preserved,
//     reversible; a manifest (TSV) records every move.
//   - -min-group sets the threshold: only groups with >= this many files are
//     thinned (default 10 — target the FAT clusters, leave small ones alone).
//
// Usage:
//
//	thin_versions <root>                          # report fat groups + keep/archive split
//	thin_versions <root> -keep 5 -min-group 10     # tune
//	thin_versions <root> -keep 5 -move             # archive older versions
//	thin_versions <root> -report-renames           # also flag NNN-increment-but-same-stem pairs
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	reMarker  = regexp.MustCompile(`\.(patch|orig\d*|bak)$`)
	reBracket = regexp.MustCompile(`\((\d+)\)$`)
	reVersion = regexp.MustCompile(`_v(\d+)(?:_(\d+))?$`)
	reNumPre  = regexp.MustCompile(`^(\d+)[_\-]`)
)

type doc struct {
	rel     string
	abs     string
	mtime   time.Time
	vMajor  int
	vMinor  int
	bracket int
}

var archiveish = []string{"_archive/", "_archive\\", "/old/", "\\old\\", "_old/", "_old\\"}

func main() {
	var (
		keep      = flag.Int("keep", 5, "newest versions to keep per subject group")
		minGroup  = flag.Int("min-group", 10, "only thin groups with at least this many files")
		move      = flag.Bool("move", false, "MOVE older versions to the archive dir (default: report only)")
		archive   = flag.String("archive", "_archive", "archive dir (relative to root)")
		extCSV    = flag.String("ext", ".md", "comma-separated extensions to consider")
		renames   = flag.Bool("report-renames", false, "also flag NNN-increment pairs with the same stem")
		manifestN = flag.String("manifest", "thin-manifest.tsv", "manifest path (written under root on -move)")
	)
	// root may precede or follow flags (avoid Go's stop-at-positional footgun)
	var root string
	rest := []string{}
	valueFlag := map[string]bool{"-keep": true, "--keep": true, "-min-group": true, "--min-group": true,
		"-archive": true, "--archive": true, "-ext": true, "--ext": true, "-manifest": true, "--manifest": true}
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			rest = append(rest, a)
			if valueFlag[a] && !strings.Contains(a, "=") && i+1 < len(args) {
				i++
				rest = append(rest, args[i])
			}
			continue
		}
		if root == "" {
			root = a
		}
	}
	flag.CommandLine.Parse(rest)
	if root == "" {
		fmt.Fprintln(os.Stderr, "usage: thin_versions <root> [-keep N] [-min-group N] [-move] [-archive dir] [-ext .md] [-report-renames]")
		os.Exit(2)
	}

	exts := map[string]bool{}
	for _, e := range strings.Split(*extCSV, ",") {
		if e = strings.TrimSpace(e); e != "" {
			exts[strings.ToLower(e)] = true
		}
	}
	archiveRel := filepath.Clean(*archive)

	groups := map[string][]doc{} // key: dir \x00 stem
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
			if rel == archiveRel || strings.HasPrefix(relSlash, filepath.ToSlash(archiveRel)+"/") {
				return filepath.SkipDir
			}
			return nil
		}
		if len(exts) > 0 && !exts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		if isArchiveish(relSlash) {
			return nil // already set aside
		}
		d := parseDoc(path, rel, info.ModTime())
		dir := filepath.Dir(rel)
		key := dir + "\x00" + subjectStem(filepath.Base(rel))
		groups[key] = append(groups[key], d)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "walk: %v\n", err)
		os.Exit(1)
	}

	// Order groups by size desc for the report.
	type kv struct {
		key  string
		docs []doc
	}
	var ordered []kv
	for k, v := range groups {
		if len(v) >= *minGroup {
			ordered = append(ordered, kv{k, v})
		}
	}
	sort.Slice(ordered, func(i, j int) bool { return len(ordered[i].docs) > len(ordered[j].docs) })

	if len(ordered) == 0 {
		fmt.Printf("no subject groups with >= %d files (nothing to thin).\n", *minGroup)
		return
	}

	var toArchive []doc
	totalArch := 0
	for _, g := range ordered {
		docs := g.docs
		sort.Slice(docs, func(i, j int) bool { return newer(docs[i], docs[j]) })
		dir, stem, _ := strings.Cut(g.key, "\x00")
		fmt.Printf("\n[%d files] %s :: %s — keep newest %d\n", len(docs), dir, stem, *keep)
		for i, d := range docs {
			if i < *keep {
				fmt.Printf("    keep    %s\n", filepath.Base(d.rel))
			} else {
				fmt.Printf("    archive %s\n", filepath.Base(d.rel))
				toArchive = append(toArchive, d)
				totalArch++
			}
		}
	}
	fmt.Printf("\n%d fat group(s); %d older version(s) would be archived (keeping newest %d each).\n",
		len(ordered), totalArch, *keep)

	if *renames {
		reportRenames(groups)
	}

	if !*move {
		fmt.Println("\nREPORT ONLY. Re-run with -move to archive the older versions.")
		return
	}

	mf, err := os.Create(filepath.Join(root, *manifestN))
	if err != nil {
		fmt.Fprintf(os.Stderr, "manifest: %v\n", err)
		os.Exit(1)
	}
	defer mf.Close()
	fmt.Fprintf(mf, "action\tsubject\tmoved_from\tmoved_to\n")
	moved := 0
	for _, d := range toArchive {
		dst := filepath.Join(root, archiveRel, d.rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
			continue
		}
		if err := os.Rename(d.abs, dst); err != nil {
			fmt.Fprintf(os.Stderr, "move %s: %v\n", d.rel, err)
			continue
		}
		fmt.Fprintf(mf, "move\t%s\t%s\t%s\n", subjectStem(filepath.Base(d.rel)), d.rel, filepath.Join(archiveRel, d.rel))
		moved++
	}
	fmt.Printf("moved %d older version(s) to %s/. Manifest: %s\n", moved, archiveRel, *manifestN)
	fmt.Println("To undo: mv each manifest moved_to back to moved_from.")
}

func parseDoc(abs, rel string, mt time.Time) doc {
	base := filepath.Base(rel)
	d := doc{rel: rel, abs: abs, mtime: mt}
	s := reMarker.ReplaceAllString(base, "")
	if i := strings.LastIndexByte(s, '.'); i >= 0 {
		s = s[:i]
	}
	if m := reBracket.FindStringSubmatch(s); m != nil {
		fmt.Sscanf(m[1], "%d", &d.bracket)
		s = reBracket.ReplaceAllString(s, "")
	}
	if m := reVersion.FindStringSubmatch(s); m != nil {
		fmt.Sscanf(m[1], "%d", &d.vMajor)
		if m[2] != "" {
			fmt.Sscanf(m[2], "%d", &d.vMinor)
		}
	}
	return d
}

// subjectStem strips version/bracket/marker/extension to the subject key.
func subjectStem(name string) string {
	s := reMarker.ReplaceAllString(name, "")
	if i := strings.LastIndexByte(s, '.'); i >= 0 {
		s = s[:i]
	}
	s = reBracket.ReplaceAllString(s, "")
	s = reVersion.ReplaceAllString(s, "")
	s = reBracket.ReplaceAllString(s, "")
	return strings.TrimRight(s, "_- ")
}

// newer reports whether a is newer than b: version, then bracket, then mtime.
func newer(a, b doc) bool {
	if a.vMajor != b.vMajor {
		return a.vMajor > b.vMajor
	}
	if a.vMinor != b.vMinor {
		return a.vMinor > b.vMinor
	}
	if a.bracket != b.bracket {
		return a.bracket > b.bracket
	}
	return a.mtime.After(b.mtime)
}

func isArchiveish(relSlash string) bool {
	for _, a := range archiveish {
		if strings.Contains(relSlash, strings.TrimSuffix(strings.TrimSuffix(a, "/"), "\\")) {
			return true
		}
	}
	return false
}

// reportRenames flags the user's "number incremented but subject is the same"
// case: two groups in the same dir whose stems match after stripping the
// leading NNN_ number. Surfaced for human review, never auto-merged.
func reportRenames(groups map[string][]doc) {
	type g struct {
		dir, stem, full string
	}
	var gs []g
	for k := range groups {
		dir, stem, _ := strings.Cut(k, "\x00")
		bare := reNumPre.ReplaceAllString(stem, "")
		gs = append(gs, g{dir, bare, stem})
	}
	byBare := map[string][]g{}
	for _, x := range gs {
		byBare[x.dir+"\x00"+x.stem] = append(byBare[x.dir+"\x00"+x.stem], x)
	}
	printed := false
	for _, v := range byBare {
		if len(v) > 1 {
			if !printed {
				fmt.Println("\n== possible NNN-increment renames (same subject, different leading number — review, not auto-merged) ==")
				printed = true
			}
			var names []string
			for _, x := range v {
				names = append(names, x.full)
			}
			sort.Strings(names)
			fmt.Printf("    %s:  %s\n", v[0].dir, strings.Join(names, "  |  "))
		}
	}
}
