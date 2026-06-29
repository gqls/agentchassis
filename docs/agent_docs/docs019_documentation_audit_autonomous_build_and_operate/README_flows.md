

Flow: load_runtime (runs the prior verdict's
requests + re-gathers runtime) → assemble_bundle → verdict (may emit new requests)
→ route (forwards them) → load_runtime … → emit.