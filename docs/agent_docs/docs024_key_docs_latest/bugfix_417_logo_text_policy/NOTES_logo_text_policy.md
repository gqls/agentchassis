# NOTES — 417 logo text policy (append-only, newest at the bottom)

## 2026-08-31 — session open

Asked to take 417 and 420, checking first whether either had a live thread. `who-owns.py` said
both were OWNED and ACTIVE — but it reads COMMITS, so it lags a session mid-fix. Messaged the
three candidate owners directly rather than trusting it. All three replied within minutes:
loanzy formally handed me 417's residual, the delivery lane handed me 420's class fix, and
boxingonline confirmed nothing of theirs was in flight on either. **Asking was worth more than
the ownership tool**, and the tool's own docs say so.

## First measurement, before planning anything

669 and 670 both applied, verified at the row (not at a tracker — there is no
`schema_migrations_agents` table, and `schema_migrations` has no `version` column; two wasted
round trips). Verbatim licences remaining: 0. **But the current-logo-prompt total had moved 27 →
28, and one wordmark row carried no 670 override.**

That row is boxingonline.com's, created **12:36:55Z — 41 seconds after 669 applied at
12:36:14Z**. And its wording is *"no text **other than** the wordmark itself"* where the
exemplar said *"**outside**"*. Two independent class points fell out of one row:

1. a migration that fixes a prompt SOURCE cannot bound prompts already in flight;
2. the model REWORDS the licence, so no literal match bounds the class — which cuts against the
   contributing lane's own §3 warning, since counting the licence by literal is still a literal.

## The misstep I made in my own brief, and what caught it

I briefed the planning agent that `bugs_closed/028` "proved the provider DISCARDS negative
clauses", and used it as a hard constraint. **That is 028's PRE-fix state.** Its fix,
`foldNegativeIntoPrompt`, is live. The agent went and read the banana adapter log for the exact
failing generation:

```
2026-08-31T12:55:50.145Z banana_provider "Banana: folded NegativePrompt into positive prompt
as a prohibition clause"  kind=logo  negative_prompt="people, faces, text, signature, watermark…"
prompt_len_before=232  prompt_len_after=407
```

`prompt_len_before=232` matches the boxingonline plan prompt exactly, and the timestamp sits
between the plan row (12:36:55Z) and the asset (12:56:10Z). **So the model RECEIVED "text" as a
prohibition and lettered "BOXING NEWS" anyway.**

The fix I would have shipped is unchanged. The *reason* changed completely, and that matters:
under my wrong premise, "the negative channel is broken" invites the cheap alternative of
repairing the negative channel. Under the true one, **a folded negative loses to a positive
licence in the same prompt**, so no amount of negative-channel work could have helped. I would
have shipped correct code with a rationale a reviewer could rightly have rejected. Logged in
`WRONG_CALLS.md` as *a closed bug's finding expires when its own fix ships*.

Second-order consequence worth keeping: the clause standing in `default_brand_prompt.go` was
itself negative-framed, so **it was weaker than its own comment claimed, for as long as it stood.**

## What the concept census found that the literal census could not

~8 of the 28 current-plan logo prompts name their exact wordmark **on purpose** (cv1
'CareerPrep', idea.uk, oufe, relojistas, robot-hands, webdesign.uk, lendzy, loanzy) — and **four
of them never use the word "wordmark" at all**, so 670's arm (b) never touched them. That is the
paraphrase point proven twice over, and it turned the opt-in field from polish into a
prerequisite: an unconditional text-free clause would have wrecked those on regeneration.

## It had already fired a second time, and only a human eye found it

The boxingonline session downloaded the served PNG and **looked at it**. It reads "BOXING NEWS"
on a site called Boxing Online. No query would have found that: `origin_prompt` looked ordinary,
the asset row said `status='active'`, the page served 200. **The defect is silent until someone
opens the image** — which is the acceptance gap this bug file names, and the reason candidate 2
keeps being tempting.

They also found a *different* defect in the same file: it is a **two-panel design comp**, not a
logo — the mark on navy left, the mark plus lettering on grey right. My guard cannot catch that;
it is an output-acceptance failure, not an input-licence one. Routed to its own bug file rather
than folded in, because the fix layer is different (store time, not prompt time).

## The liveness probe I ran, and threw away

Before claiming the guard's coverage was total, I checked whether the two kindless legacy parents
still dispatch. The `orchestration_states` probe returned zero — **and returned zero for a
known-live control too**, so it proved nothing. Recorded as `[UNVERIFIED]` with a LANDMINES entry
and a post-roll census disconfirmation, rather than as "those parents are dead". A probe whose
control also fails is not evidence.

## Council submission — three schema errors, all caught free

`DRY_RUN=1` caught `operation: "create"` (must be `add`), `risks` as an array (must be a string),
and then my own over-correction: I "consistently" turned `grounded_in` into a string too, and it
must be an array. **I applied one field's error message to a different field.** The validator was
standing right there and would have told me for free if I had re-run after each single change
instead of batching a fix with a guess. Logged in `WRONG_CALLS.md`.

## Verification against HEAD — and why the first read was misleading

`verify-head-builds.sh --with … --test` printed **`FAILED`** and **exited 0**. Three failures,
one a build failure, all in `test/`. None of them touched my files — but that is an inference,
not a measurement. Ran the bare-HEAD control: **it fails in ~23 places**, and every failure in
my set appears in the control's set. So the honest claim is "my changes introduce no new
failure", not "HEAD is green" — it is emphatically not green.
