// Package buildinfo carries build-time provenance, stamped by the makefile's
// ref_build/tree_build via -ldflags "-X ...". See bugs_open/153.
package buildinfo

// GitCommit is the full 40-hex commit sha the binary was built from (ref
// builds), "<shortsha>-tree" for working-tree builds, or "unknown" for any
// build that bypassed the makefile.
var GitCommit = "unknown"
