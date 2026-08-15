# NOTES — router engine (append-only, newest at the bottom)

## 2026-08-15 — lane created at hand-off, no design work done yet

Created by the bugfix_277 session the evening the owner ruled RFC_030 (see the RFC's status
block for the ruling text). Nothing built. The reason the lane exists as a lane and not as
more of 277: the engine is a shared mechanism needing its own design round, and the 277 lane's
ruling was for one type's handler. Everything the first working session needs is in the PLAN
and the HANDOFF; the census/canary evidence for 410 lives in `bugfix_277_required_fields_repair/`.

One design fact learned on 277 that bears directly on the A-vs-B choice: `conditional_branch`
conditions are `==` cascades only (a missing field makes `!=` evaluate true), so a data-driven
N-way route table cannot be expressed in today's branch action without either a loop step or a
Go evaluator — which is most of the case for shape B.
