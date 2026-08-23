# Open threads — restart list, 2026-08-23

**Written 2026-08-23, 10:35 UTC (11:35 BST), for a planned reboot.** This is a
snapshot of the sessions that were actually running on this machine at that
moment, so they can be brought back afterwards. It supersedes
`OPEN_THREADS_RESTART_LIST.md`, which was last refreshed on 2026-07-27 and
describes a fleet state 26 days gone.

## What this file is, and how it was built

There are two different things people mean by "open threads" here, and this file
is about the first:

1. **The Claude sessions running in terminals on this machine.** They die when
   the machine reboots. That is the list below.
2. The 85-odd *workstreams* recorded in the memory index. Most of those are
   dormant lanes with no session attached; they are not affected by a reboot and
   are not listed here. `MEMORY_workstreams.md` remains the map for those.

**The roster is measured, not remembered.** Every session Claude Code starts
writes a small file at `~/.claude/sessions/<pid>.json` naming its process, its
session id and its name. I took every one of those whose process was still
alive — **44 as of 2026-08-23 10:31 UTC** (43 of them plus this one) — and then
read each session's own transcript to work out which lane it is on.

**But the one-line state descriptions are each thread's own account of itself.**
I compressed them from the last message each session wrote. I did not
independently verify those claims, and several of them are hours old. Treat a
row as a pointer to where to look, never as evidence that the work is done.
Two facts *were* checked directly against the live system, and they are in §0
below because they change what you can do on restart.

---

## How to restart a thread

The conversations are on disk, in `~/.claude/projects/-home-ant-projects-agentchassis/`,
and a reboot does not touch them. So a thread is resumed with its full context —
you are not starting it again from a handoff document:

```bash
cd /home/ant/projects/agentchassis
claude --resume <session-id>
```

Run `claude --resume` with no id for an interactive picker of recent sessions.

**Capture this before you reboot, not after.** The `~/.claude/sessions/*.json`
files record process ids, and after a reboot those processes are gone, so the
"which of these was actually running" question becomes unanswerable. That is the
whole reason this file exists. The transcripts themselves survive, so every id
below stays resumable indefinitely.

**What dies with the reboot:** the sessions, and any background watchers they
started (several were polling the database on a timer). **What does not:** all
the work on the cluster. Orchestrations, council runs, CronJobs and site builds
run there, not here, and they carry on regardless.

---

## §0 — Two things to check before acting on any row

**1. The Anthropic API — working again, but not yet proven on the fleet's own
path.** The account hit its usage limit at **2026-08-22 18:15:35 UTC**, and for
about sixteen hours every LLM step in the fleet failed with the same HTTP 400:
*"You have reached your specified API usage limits. You will regain access on
2026-09-01."* That took down councils, page builds and diagnosis loops together.
It has since been fixed — the cause was that the fleet's key belongs to a
different Anthropic account from the console being watched — and a direct
inference call from inside an `agent-chassis` pod succeeded at **10:24 UTC
today**.

What is *not* yet true is that the production path has been seen working.
Measured at 10:31 UTC:

| | value |
|---|---|
| last successful Anthropic call in `llm_call_log` | 2026-08-22 18:15:51 UTC |
| Anthropic failures since | 146, newest 10:10 UTC today |
| any call at all, success or failure, since 10:20 UTC | **0** |

So the fleet has not attempted a single call since the fix. The first real
dispatch is the proof. **Until you have seen one succeed, treat any thread's
"we're capped" line as stale rather than wrong** — a good many of the rows below
end on exactly that note, because that is where their day stopped.

**2. The chassis is on `v1.0.1327`; the makefile now says `v1.0.1328`.** The
running image is `docker.io/aqls/agent-chassis:v1.0.1327` (checked at the
deployment). The makefile's `IMAGE_TAG` reads `v1.0.1328`, which means another
session has already bumped it for the next build. This matters because of a trap
that bit the `bugs_open/315` lane yesterday: when the makefile tag equals the tag
already running, a "fresh build" ships the node's cached image and no new code —
guards get committed, look deployed, and are inert. If you need something you
committed to actually be live, ask the running service what it was built from
rather than inferring it from a tag.

---

## §A — The live sessions

Grouped by what they are, not by importance. Each entry: what the thread is on,
where it stopped, and the command to bring it back.

### A1 — Bug lanes, mid-flight

**bugs_open/307 — the terminal write contract**
Workflows that end at their error terminal are recorded `COMPLETED` with `error`
NULL. Spun off `RFC_043`, and bugs 341/354 sit beside it. Last act was a
correction: it checked and found the API cap had *not* reset when told it had.
```
claude --resume fb9f5a42-9493-4729-94a4-40484882c44e
```

**bugs_open/342 — an absent required field renders empty and silent**
Reports the refusal live and armed on `v1.0.1326`, verified by probing both
replicas for literals unique to the change, with migration 551 arming it. Sixteen
commits in, tree clean.
```
claude --resume 4f871ea9-740e-4460-9bf3-c46ce9963512
```

**bugs_open/309 — unclickable index cards / phantom source fields**
Its own remit reports complete: the check is live on `v1.0.1326` and was proven
by a manual Job run rather than at the make target. 309 stays open as the parent
of `bugs_open/362`.
```
claude --resume 86551e70-38ae-4f7c-91ff-e08aa2f46bc4
```

**bugs_open/317 — live object declaration drift**
Go guards assert the text of append-only migrations, so they cannot fail when the
live object moves. Filed `bugs_open/363`. A `090` diagnosis run came back
`UNVERIFIABLE` — dispatched without a `SEED_SCOPE`, and logged as its own wrong
call.
```
claude --resume e7f43a3a-7da3-460d-90f1-197268d22785
```

**bugs_open/308 — CTA destination provenance**
Stopped on the cap: council round 4 ran to `complete_invalid` with no verdict
row. **Do not resubmit before §0's first dispatch has succeeded.**
```
claude --resume 4ccfd49b-d116-4217-b349-305082e28b8c
```

**bugs_open/337 — the token cap that wasn't**
Reports the bug misdiagnosed twice, once by itself: the 16,000-token cap fits
about 89% of generations. The real parker is a regeneration loop — the field
contract rule accounts for 97 of 101 rejections.
```
claude --resume 8e4aaf89-911c-4e08-847a-94ce9f990de7
```

**bugs_open/358 — finding codes nobody reads**
Seven disposition proposals written and **waiting on your ratification**; nothing
applied. Found one code that is read automatically but whose log row is still
read by nothing.
```
claude --resume 05a054d0-5dad-47ea-9a62-362524d24a6d
```

**bugs_open/315 — `deployed_at` without publication**
The graded result is in (`page_content_divergence`: 1 true positive, 0 false
positives over 311 pages). Its two guards are committed but **inert** — this is
the lane that found the same-tag trap in §0.
```
claude --resume 3ea4dc1f-8603-4141-8ae5-a7e09f6f8972
```

**305 negation gate**
The gate lane, and the thread that first caught the fleet-wide cap at 18:15Z.
Filed `bugs_open/366`.
```
claude --resume 24db15b2-10d6-46a5-8e7b-a90c17e081ea
```

**bugs_open/283 — component instance scope** *(waiting on you)*
283 itself was closed by another session mid-plan. This thread continues its
architecture residual, `RFC_032`: three render-context builders disagree about
what an instance is. It is sitting in plan mode.
```
claude --resume 76827be0-fd19-4720-bb6b-25fe141de9c4
```

**bugs_open/204 — stored slot identity** *(finished)*
204 and 189 both closed and moved to `bugs_closed/` at your instruction. You had
already asked to close this thread; nothing is left in it.
```
claude --resume b0e7cb40-a5e0-438f-8fa8-785515c11e6b
```

**bugs_open/351** — named for the lane, but only 11 entries: the actual 351 work
happened in the `remortgagecalculator.uk` thread below. Nothing to resume.
```
claude --resume 4741fc78-2cbe-4c70-84aa-5418f25f04c5
```

### A2 — Platform and process lanes

**staged-component-build**
Reports the gate fired and the lane's deliverable proven in production, with a
new handoff at
`docs/agent_docs/docs024_key_docs_latest/staged_component_build/HANDOFF_2026-08-22_continue_here.md`.
```
claude --resume d1fb4492-0d05-4bc4-a121-73d4b5666023
```

**throughput — dispatch throughput / whole-architecture scale** *(waiting on you)*
Research only, no config applied. Your rulings are recorded; it corrected D7 in
your favour — higher Anthropic usage tiers exist (Start/Build/Scale/Custom) and
the Console has a "request rate limit increase" control. Given §0, this is the
lane that matters most this week.
```
claude --resume 1c893fc1-67f1-4069-aa8d-4e328d91b304
```

**web admin console (agentchassis-31)**
Built and committed the path `admin.apis.uk → Cloudflare Access → cloudflared →
admin-dashboard`, needing no cluster change. **The remainder needs your hands:**
box SSH and your Cloudflare account.
```
claude --resume 07adcaaa-e03c-479b-aaa7-0789fb311b17
```

**admin panel on web for me** *(waiting on you)*
The sibling of the row above — it stopped on a question about how the panel
should be reachable. Answering it in one of the two threads should retire the
other.
```
claude --resume 1fb9fe42-e044-4e72-ae43-81803ee07a48
```

**webdesign-tool-rebuild** and **webdesign-tool-rebuilds** — two threads on one
lane, deliberately coordinating by cross-session message. The first ended on a
timed-out request, so check its state before assuming; the second reports three
items settled, including a staged-removal hazard closed with evidence.
```
claude --resume b22eb831-b353-4d6f-95cf-489885574be7
claude --resume 32d30dc3-30d0-4576-8954-4eebf12f3625
```

**copy quality, two stage (agentchassis-0b)**
Stage 2 works end to end: three approved edits are live on the AI-orchestration
homepage, verified on the served page. Both of the lane's open threads closed.
```
claude --resume 4bf32e82-cef3-491a-9e58-b2311a3d301f
```

**vigilant designer / offer analysis (agentchassis-7c)**
Gate live; the "eight live sites" claim is gone from leopardess. It declined to
call that "the gate works" on one run, which is the right instinct.
```
claude --resume 9b49fb73-8684-40eb-ad2d-d10e2289eaf5
```

**site AI agent orchestration (agentchassis-53)**
Three asks done, verified live, committed; nothing in flight.
```
claude --resume 948da49b-5b36-4a59-be41-794769c869ed
```

**portfolio positioning (agentchassis-0f)**
Confirmed at the artefact that 311 is genuinely closed — `remortgagecalculator.uk`
serves a real calculator now (69,704 bytes, six inputs) where it served none on
18 August.
```
claude --resume 8a3568f7-8e0c-4e65-8f54-a7b08db63706
```

**API keys and secrets (agentchassis-88)**
The thread that found and fixed the cap. Currently putting the xAI key into the
terraform secrets file and `personae-default-secrets`.
```
claude --resume b269fb5b-e5ec-43b4-8686-363366c96650
```

**robot-hands gripper dossier (agentchassis-7e)**
Mailer credentials installed on the island and verified. Worth reading for the
trap it caught: a `$` in the password would have been silently eaten by Docker
Compose, surfacing only as an SMTP 535 *after* a visitor had been promised a
report.
```
claude --resume 0cb740bc-2fb7-43ef-ba46-49e3b3b0c1cb
```

**council scope (agentchassis-00)** — barely used, 114 entries, on `bugs_open/314`.
```
claude --resume bd9b0a56-605d-46d2-9890-9b74e266cd7d
```

**staged component build / gaswholesalers (agentchassis-c2)** — hit its session
limit mid-task; a retract script it wanted does not exist at the path it used.
```
claude --resume 96329101-9a1c-4acb-a302-01be12d78735
```

**finetuning.uk repair (agentchassis-7b)** and **finetuning.uk service
(agentchassis-44)** — the second had only read its cold-start handoff, no edits
in flight; its next step is seeding identity and `content_direction` specs.
```
claude --resume 60959a6c-8cd7-413a-ad1f-e72fc8e63b52
claude --resume 6e488a7d-c72a-4e45-8431-1482cdc47386
```

### A3 — Site and customer lanes

**apis.uk**
Took an infrastructure disclosure off the live page and fixed the voice at
source; mid-analysis of how much of an exemplar's style actually transfers
(2 of 5 left literal traces).
```
claude --resume 3da2bd86-6c7a-4275-b6a3-fb7156d60ea1
```

**loanzy.uk** *(needs your hands)*
Cloudflare is set up for **garden-tools.uk** and verified by reading the zone
back. **Set these nameservers at your registrar:** `alexis.ns.cloudflare.com` and
`leah.ns.cloudflare.com`. You asked it not to build the domain yet.
```
claude --resume 6b6a11fb-1b19-4c9b-a14a-a691f1f11c94
```

**agritec.uk**
Deconstructed the SFI calculator against the real SFI26 rate tables: **two of its
nine revenue lines were correct**. Rates registered; the rest waits on the site
existing.
```
claude --resume 9962fa6d-4177-42d7-afea-7b8014b18653
```

**dartsonline traffic (agentchassis-51)**
Card imagery audited — one missing card, one 404 hero across five pages, both
queued for repair. Also carries the robot-hands improvement-loop misfire, whose
bounded cancel you ran.
```
claude --resume fe285621-baaf-43d6-adde-95f45e72beb9
```

**news editorial**
Incident closed: the bounded cancel landed before the tool-modifying items
dispatched, and the site verified healthy. One item still waits on you.
```
claude --resume e03a2052-581b-4e12-975d-d24af88af605
```

**remortgagecalculator.uk**
351 implemented, committed and green on clean HEAD, handoff written. Checked
first and found the corpus had moved underneath its recorded line numbers — that
check was the difference between a fix and a mess.
```
claude --resume be1d652a-b1f2-4077-831a-1880f5f58618
```

**idea.uk (agentchassis-4a)** *(live defect)*
Chasing invisible text caused by CSS: the site's colours live in a stylesheet
written by design runs and never backfilled into the database, so **11 of 21
sites serve a 13–25KB stylesheet while their `css_themes` row is empty.** A
canary was minutes from its promoter tick when it stopped.
```
claude --resume 36c8ab58-b1b3-44ba-b9e0-6a30a57fabec
```

**mortgagecalculator.co.uk adoption (agentchassis-0e)** — a page reword is live:
it had invited contact four times with no email, phone or form behind it.
```
claude --resume 09f2e4bd-8960-4902-836b-25933688f5c7
```

**loancalculator.co.uk (agentchassis-e5)** — docs current; cold start is
`loancalculator_couk/HANDOFF_2026-08-17b_continue_here.md`.
```
claude --resume f5a3ac5c-4370-4a83-9650-0db3a969ba02
```

**noted.co.uk rebuild (agentchassis-de)** — idle, 31 entries.
```
claude --resume 094c7a26-7541-4b9b-b064-39f1d7c2f7a0
```

**second cluster for customers** — named and never started (13 entries).
```
claude --resume 3e3ca3d4-b3f4-468e-822a-3f9e5a062403
```

### A4 — Empty shells, nothing to restart

Four sessions were opened and never used (0–8 entries each). Listed only so the
count reconciles: `agentchassis-5c` `c107b7b8-1f58-42d9-9a31-13216e4cb715`,
`agentchassis-43` `811eb365-b04b-4b6d-bcda-c176ca5b238c`, `agentchassis-88`
`27e448ba-3a9c-4790-b05d-35f1ed8dff53`, `agentchassis-68`
`c85ad3e5-0f1c-41ff-91c7-45dd8368c895`.

The 44th session is the one that wrote this file.

---

## §B — Lanes worked in the last 48 hours whose session has already exited

These are not running, so a reboot costs them nothing — but they are live work,
and several ended mid-task. Their transcripts are on disk too, so the same
`--resume` brings them back rather than starting from the handoff.

| lane | where it stopped | resume |
|---|---|---|
| 283 / `RFC_032` render-context builders | council mid-run when it ended | `ce683d1f-fd6d-466e-8f01-b28c084541a1` |
| 238 regeneration key loss / 355 | **355 closed**, proven on organic traffic | `550e8700-109f-4e26-9ac4-4e46aa2e3e21` |
| 114 imagery wiring | fix live, migration verdict not yet landed | `87197fc3-4816-4f35-bc82-81109ec1b4ee` |
| 315 divergence check (earlier session) | check graded: 1 true positive, 0 false | `de3fce7d-35b9-4a33-b24e-086d4b2eb7fb` |
| 235 / 155 / 071 closeout | post-roll verification complete | `70a7587c-81c2-4479-9df1-5d473b4f6469` |
| architecture seat / `RFC_008`, `RFC_042` | no mandatory seam; detector live daily | `e5b376d6-a7d0-4ef0-abed-967470de1ea9` |
| 153 / 318 release source coverage | closed by the `v1.0.1326` release | `4b3eadfe-cd70-4199-af99-3e5b4befd5b8` |
| 307 orchestration status lifecycle | verifying against clean HEAD | `0a83b064-3fe8-4672-9d58-babdac81a3cd` |
| 131 contrast ratio check | writing its handoff | `65c985be-b316-463f-a1a5-8211417a46fc` |
| 345 rejected-component regeneration | four guards mutation-proven | `5778fcdd-5d3d-4cc3-9e6b-53f2528bc951` |
| 283 / loanandmortgagecalculator | lane closed and verified | `5f8ee208-9fb6-47a8-8eb9-564e5018685b` |
| 260 render fallback | lane closed; handoff routes onward | `9fd0eb64-5218-4e28-8edf-23517df2f71f` |
| 311 component keys | `completed_at` hardening live on 1326 | `4a9984b2-ae36-493f-918b-95b739315c04` |
| 277 / 357 component identity | council approved with two advisories | `aeae602a-c290-4a17-ab9b-9408622cf823` |
| 316 news feed ordering | checking prior art before filing | `bf9618fb-e378-4fdf-828f-d2132bdca3ff` |
| 358 unread finding codes (earlier) | handoff written, docs committed | `879fa28f-f85e-472e-bd2c-0c5b7b8fbd0c` |
| 298 / 356 orphan check lifecycle | filed `bugs_open/359`, ready to close | `90c26dfe-6751-4ef8-b70c-f2a3359b7b64` |
| 337 token cap (earlier session) | census refuted the filed diagnosis | `5d87d017-03e3-4500-9591-cd7539a6bc7b` |
| 083 / 277 required fields | both bugs closed, lane finished | `ea5b0289-2c32-4316-aa20-b21d847e0091` |
| 343 / 040 kafka dial | council trail recorded | `6b906c3c-ac80-4d62-b945-920fbf2e5f99` |

---

## Re-deriving this list

Run this **before** a reboot, while the processes are still alive:

```bash
python3 - <<'PY'
import json, glob, os
for p in glob.glob('/home/ant/.claude/sessions/*.json'):
    try: m = json.load(open(p))
    except Exception: continue
    if os.path.exists('/proc/%s' % m.get('pid')):
        print(f"{m.get('name','?'):35s} {m.get('sessionId')}  {m.get('status')}")
PY
```

To work out what a session is on without opening it, its transcript is
`~/.claude/projects/-home-ant-projects-agentchassis/<session-id>.jsonl`; the last
assistant message in that file is that thread's own summary of where it stopped.

**A note on the counts.** "44 sessions" is true as of 2026-08-23 10:31 UTC and
nothing else — two sessions exited during the twenty minutes it took to write
this. Bug numbers cited here were current the same morning: **115 files in
`bugs_open/`, 293 in `bugs_closed/`**. Both drift daily; re-count rather than
quoting these.
