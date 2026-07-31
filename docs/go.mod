// NOT A REAL MODULE — a boundary marker, so `go build ./...` at the repo root
// stops here.
//
// docs/ holds 146 .go files across ~20 workstream directories, and none of them
// is buildable code. They are verbatim reference copies: snapshots of VM
// services (traffic_probe/deploy_setup/working_dir), duplicated drafts
// (service(11).go, service(25).go, main(13).go), and checks lifted out of their
// own package so their imports no longer resolve
// (adoption/check_sectionless_pages(1).go references Register /
// DiscoveryCheckContext / CheckResult, which live in discovery_checks).
//
// Before this file, `go build ./...` failed repo-wide on them — first with
// "found packages main and working_dir", then on the next directory, and the
// next. That is why every runbook here tells you to build the four real trees by
// name. Go excludes a nested module from the parent's ./... pattern, so this one
// file replaces build-tagging 146 files across other lanes' directories.
//
// NOTHING IMPORTS docs/ — verified 2026-07-31:
//   grep -rn "agentchassis/docs" --include=*.go platform/ internal/ pkg/ cmd/ scripts/
// returns nothing, so this boundary costs no real package anything. `go run` on
// an individual file here still works.
//
// Added 2026-07-31 alongside the cmd/reasoningset fix (bugs_closed/137 session).
module github.com/gqls/agentchassis-docs-reference

go 1.24.0
