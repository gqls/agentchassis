# Where we are — RFC 012 execution (findings that must survive an await)

Append-only, newest at the bottom. Plain prose.

## 2026-08-06 — your three calls, and what each one sets in motion

The background: when one of our agents both *works something out* and *asks an external
service to do something*, the reply from that service overwrites the agent's own workings
in the permanent record. Three separate teams have now hit this and each invented their
own workaround. You ruled: build the shared escape hatch properly (the database-backed
version that survived testing, not the in-memory one that didn't), run a census of
everything that reads those overwritten records (so the deeper fix — merging instead of
overwriting — becomes decidable later), and turn the one-off audit of how workflows route
their outputs into a standing check that runs inside the platform rather than a script
someone must remember.

All three are now in motion: the rulings are written into the RFC itself so they can't be
lost, and three research passes are running in parallel — one gathering everything needed
to build the shared helper, one finding the right vehicle for the standing check, and one
performing the census itself.
