

Flow: load_runtime (runs the prior verdict's
requests + re-gathers runtime) → assemble_bundle → verdict (may emit new requests)
→ route (forwards them) → load_runtime … → emit.


The code-retrieval channel is dead weight right now — measured, not guessed: seed similarities in the 0.55 band, no page-build symbols. If you want it to contribute, the work is on the query side (seed the lookup from runtime evidence, or expand the query), and it's a self-contained change to lookup_symbols. The trigger's site_id is still intermittent across runs, which matters for reproducibility. And two verdict-quality wrinkles surfaced in the last run — a stale line carried into a conclusion that its own citation contradicts, and terminal-verdict data_requests that never run — both of which point at the confirm/emit step rather than anything structural.
You've re-pasted the broader picture — the multi-agent website builder with per-section agents and research sub-agents — so I take it the diagnosis work is at a natural stopping point and the next move is back toward that. If so, the sensible starting point is to map what section agents already exist against the sections you want (research, content, design, tools/graphs, blog/news), so we reuse and extend rather than recreate, and keep each agent's responsibility distinct. 