// FILE: go.mod
//
// Deliberately a NESTED module, following the precedent of the four other
// prototype modules under docs/ (idea.uk/golang_files, traffic_probe/…,
// content_quality_and_internal_linking/golang_code, docs019/contextkit).
//
// Why it matters: a nested go.mod is excluded from the root module's
// ./... expansion, so this throwaway cannot break `go build ./...` on a
// shared HEAD that fourteen services build from.
module pairedprototype

go 1.24.0
