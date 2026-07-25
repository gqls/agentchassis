// Command webdesignport ports the two hand-built static sites
// (website-design.com and websitedesign.com, in the sites repo) into
// warm-minimalist owned pages for the chassis-managed webdesign.co.uk.
//
// Subcommands:
//
//	harvest    bootstrap the catalogue from the source sites' own index pages
//	transform  source HTML -> owned-page fragments + manifest (repo-local)
//	import     manifest -> pages / page_components rows (needs the site row)
//	verify     gates: colour sweep, link closure, visible-text floor, live checks
//
// Why Go and not a shell/psql pipeline: `import` writes ~97 large HTML blobs.
// Through psql those would ride a shell heredoc, which is exactly the
// backslash/NUL mangling trap this repo has been bitten by. Through a driver
// they are parameters, and never touch a shell.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)

	var (
		sitesDir = fs.String("sites", os.ExpandEnv("$HOME/projects/sites"), "path to the static sites repo")
		portDir  = fs.String("port", "docs/agent_docs/docs024_key_docs_latest/webdesign_couk/port", "path to the port data dir (catalogue, overrides, compat CSS)")
		outDir   = fs.String("out", "build/output/webdesign_couk", "output dir for fragments and manifest (gitignored)")
		domain   = fs.String("domain", "webdesign.co.uk", "target domain")
		dsn      = fs.String("dsn", "", "postgres DSN (import only)")
		dryRun   = fs.Bool("dry-run", false, "report what would change without writing (import only)")
		only     = fs.String("only", "", "restrict to page names matching this substring")
	)
	_ = fs.Parse(os.Args[2:])

	var err error
	switch cmd {
	case "harvest":
		err = runHarvest(*sitesDir, *portDir)
	case "transform":
		err = runTransform(*sitesDir, *portDir, *outDir, *domain, *only)
	case "import":
		err = runImport(*outDir, *dsn, *domain, *dryRun, *only)
	case "verify":
		err = runVerify(*sitesDir, *outDir, *portDir, *domain)
	default:
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "webdesignport %s: %v\n", cmd, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `webdesignport — port two static sites into chassis-owned pages

usage:
  webdesignport harvest    [--sites DIR] [--port DIR]
  webdesignport transform  [--sites DIR] [--port DIR] [--out DIR] [--only SUBSTR]
  webdesignport import     [--out DIR] --dsn DSN [--dry-run] [--only SUBSTR]
  webdesignport verify     [--sites DIR] [--out DIR] [--port DIR]
`)
}
