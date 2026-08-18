# NOTES — loanzy.uk example site

Append-only, newest at the bottom. Missteps included — they are the point.

## 2026-08-18 — session start: finding the previous thread, and what it turned out to be

Asked to find the previous loanzy.uk thread and take it on. There were **three**, none of
them a lane of its own:

1. `idea_uk_vm_site`, 2026-08-03 (owner: *"can you do loanzy.uk"*) — delegated to Cloudflare
   via Nominet EPP under the DESIGNCONSULT tag, zone active in 60s, edge cert issued. Left an
   open item: *"loanzy.uk content: needs a Worker route / B2 wiring — webdesign lane's
   machinery."*
2. `domains_cloudflare_rollout`, 2026-08-09 — the **dangling delegation**: NS pointed at
   Cloudflare, zone deleted, so requests timed out instead of failing honestly. Owner
   re-added the zone; clean 404 since. That session deliberately built nothing (HOLD domain).
3. `portfolio_positioning`, 2026-08-15 — **P9**: *"leave loanzy.uk with the webdesign team"*.

**A correction made on the way in.** `portfolio_positioning/HANDOFF_2026-08-18` §5 listed
loanzy.uk, the B8/B9/I10 holds and the build order as Phase D decisions *"unchanged and still
outstanding"*. All three had been **ruled by the owner at P9 on 08-15** — the bullet was a
stale carry-forward from `PLAN_2026-08-12`. Struck through with a dated correction
(`8229b1362`), not deleted. Cost of not catching it: that lane keeps asking the owner a
question he closed three days ago.

**The owner's memory of "a site about FCA rules built alongside loancash" was right about the
site and wrong about the domain** — it is `lendzy.co.uk`, one letter from `loanzy.uk`. Both
`loancash.co.uk` (L10, *"The Rules That Protect UK Borrowers"*) and `lendzy.co.uk`
(*"Know the Rules Before You Borrow"*) are live, 22 pages each, both redeployed 2026-08-18.
`loanzy.uk` has **0 rows in `sites`, 0 work items** and serves a 9-byte 404. Checked twice,
in two ways (DB census across `loan|lend|borrow|credit`, plus the live probe), because the
first check was a name match that could only ever have confirmed what I already believed —
the family census is the one that could have come out otherwise, and it did.

**What the decision became.** Four messages, converging: back in the finance queue → no, an
example site → which means no prior registry entry → built only from the webdesign.uk prompt.
Recorded as **P10** in the register with its consequences enumerated (`f21530d37`), because a
ruling whose *implications* live only in a chat log is one a future session will misapply.

**Why it lands where the webdesign lane is actually stuck** `[MEASURED, from their docs]`:
proposal F was approved the same day and its `writer_block` forbids naming examples until a
gallery exists, because none of the four attested example sites was built by the one-shot
route. This lane produces the first pair that would populate one.

**What the "one-shot route" is, mechanically** `[MEASURED]`: the chat box
(`box/chat-service`) is a self-contained intake bot — stdlib only, no DB driver, transcripts
to JSONL on the box. It **dispatches nothing**. So today the route is: customer's brief →
operator seeds and dispatches the standard gated pipeline. The honest constraint is therefore
about the **input**, not the button: every seeded input must trace to a sentence in the
prompt. Written into PLAN as the rule, with a requirement to log deviations here.

**Not done, deliberately**: no seed, no dispatch, no lane SUMMARY. The prompt is the owner's
and it is a published artefact — building first and asking after would produce a pair whose
prompt was reverse-engineered from the site, which is the one thing this exercise cannot
survive.

## 2026-08-18 12:53Z — DISPATCHED with NOTHING but the domain

**Owner's steer, which changed the plan before Phase 1 ran:** *"I'd like the framework to
determine the prompt — it should already have that step in the research stage or thereabouts.
So assuming no prompt at all. Let's try that."* So the candidate prompts I had drafted were
never put; PLAN Phase 1 ("the prompt, owner input, BLOCKING") is **struck** — the input is the
domain string and nothing else.

**He is right that the step exists** `[MEASURED]`. `082_submit_domain_unified.sh`'s own header
documents the graph: `needs_domain_research → domain-research-classifier → needs_strategy →
domain-strategist → needs_briefing → build-briefing-agent → needs_site_plan → …`, and *"a
fresh domain still gets a research step regardless: the classifier always runs in full (web
research + synthesis)"*. `--mission` is optional in the arg parser, and `domain-submitter`'s
workflow marks `persist_mission` *"skipped if absent"* — so a mission-less fresh submission is
a supported path, not a hack.

**What was passed: `loanzy.uk`. Nothing else.** No `--mission`, no `--email`, no `--phone`, no
fidelity, **and no seed SQL at all** — the FRESH path's `ensure_site_record` creates the site
row itself. That means no `evidence_base`, no `banned_claims`, no `imagery_style_guide`: the
site starts with nothing attested. Where webdesign.uk's own build could state its price
because the owner had attested it, this one has no attested anything, and what the framework
does about that is part of what we are here to observe. **Withholding the email was a
choice**: a customer supplies contact details, so if the build raises an item asking for them,
that is a finding to record and answer, not a failure to pre-empt.

Preconditions checked before firing, none inferred: chassis pods up since 07:57Z (~5h, clear
of the ~300s silent-drop window); `domain-submitter`, `domain-research-classifier`,
`domain-strategist`, `build-briefing-agent` all `is_active` and not snapshots; zero existing
work items on the domain.

```
CORRELATION_ID=a892b446-36bb-4b30-83be-71cca81ff53e
ORCHESTRATION_ID=84865f5f-5ef1-4cc1-8fb1-656951f1a5a6
site_id=55213ded-03ec-40f7-8fc1-169de05e05c8
```

`[MEASURED]` **Verified landed, not assumed from exit 0** — `kcat -P` is known to publish
nothing and still exit 0. `orchestration_states` COMPLETED at step `complete`, `sites` row
created (`build_status=pending`), and `needs_domain_research` filed `triaged` at 12:53:00Z for
`domain-research-classifier`.

⚠ `bash <script>`, not `./<script>` — the file is mode 644 in the tree, so `./` gives
"Permission denied". Recorded in the RUNBOOK.

**The risk this run carries, stated before the result is known so it cannot be rationalised
after:** the only signal the classifier has is the string "loanzy.uk", which reads as a
lender. If it invents a lender or broker, the result is a fake regulated firm on a live UK
domain, and the honest response is to retract rather than to publish it as a demo. Reviewing
the strategy spec before pages deploy is NOT steering the build — no input changes — but if I
ever intervene in the content, it gets logged here as a deviation and the pair stops being
"built only from the domain".
