contextkit/
├── go.mod                          module contextkit
├── internal/
│   ├── analysis/
│   │   └── types.go                analyser output contract — defined ONCE (was duplicated ×3)
│   └── candidates/
│       └── types.go                ranked-candidate contract — defined ONCE (was candFile ×2)
└── cmd/
├── analyser/main.go            walks a Go tree → analysis JSON
├── assembler/main.go           builds the paste-ready bundle
├── embed/main.go               build | query the semantic index
├── dbcontext/main.go           schema / rows / runtime evidence via psql
├── resolve_targets/main.go     lexical target proposal (-json)
├── fuse/main.go                RRF merge of candidate lists
└── eval_targets/main.go        recall@N / MRR vs ground truth


tar -xzf contextkit.tar.gz
cd contextkit
go build ./...        # compiles all seven commands
go run ./cmd/analyser /path/to/your/repo > analysis.json