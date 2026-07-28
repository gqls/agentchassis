# RUNBOOK — fleet-wide claim-pattern dry run

Every command here was needed to get the 2026-07-28 measurement right, with the
gotcha that cost time attached. Change them **here**, not in scrollback.

## 0. The tool already exists — do not build one

`cmd/claimscan` runs the **same shared engine** as the deploy gate
(`validate_page_content` check 8) and the post-deploy audit
(`check_unverified_claims`), over exported component HTML. Its own usage block is
the reference; `sql_for_agents/226`'s verify section already names it.

```bash
go build -o /tmp/claimscan ./cmd/claimscan     # builds clean from the shared tree
```

## 1. Which sites are armed — and the trap in the obvious query

`104` § Measurement uses `jsonb_array_length(COALESCE(ss.data->'banned_claims','[]'))`.
That returns **0 for "no evidence_base row at all" and 0 for "row with an empty
array"** — two different states that matter, because candidate 1 is gated on
`ParseEvidenceBase` returning non-nil, which is satisfied by `facts[]` alone.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT s.domain, jsonb_array_length(COALESCE(ss.data->'banned_claims','[]'::jsonb)) AS bans
  FROM sites s
  LEFT JOIN site_specs ss ON ss.site_id=s.id AND ss.aspect='evidence_base' AND ss.is_current
 WHERE s.status NOT IN ('pool','archived') ORDER BY bans DESC NULLS LAST;"
```

Split the two states before drawing any conclusion — 2026-07-28: **6 sites have
no row**, **2 have a row with 0 patterns but non-empty `facts[]`** (robot-hands 5,
gamesdesign 4), 7 have patterns.

## 2. Export components, per site

The TSV shape is `page_name <TAB> slot_name <TAB> base64(html)`.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c \
 "SELECT p.name || E'\t' || COALESCE(pc.slot_name,'') || E'\t' ||
         replace(encode(convert_to(pc.rendered_html,'UTF8'),'base64'), E'\n','')
    FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = '<site_id>' AND pc.rendered_html IS NOT NULL
     AND pc.rendered_html <> '' AND pc.locked_at IS NULL" </dev/null > comp.tsv
```

**GOTCHA 1 — `kubectl exec -i` eats the loop's stdin.** A `while read` loop over a
site list terminates after **one** iteration, silently and with exit 0. Redirect
`</dev/null` on every `kubectl exec` inside a loop, or `mapfile` the list into an
array first. This cost a full run that looked like a clean single-site result.

**GOTCHA 2 — never `2>/dev/null` the fetch.** A transient `kubectl exec` failure
then reads as "this site has no evidence base", which is a *data claim*. It
happened here to **vonc** — the one site whose register mattered most — and the
table was wrong until the retry. Retry 3× and print `FETCH_FAIL` distinctly from
`no-row`.

## 3. Scan

```bash
/tmp/claimscan -evidence <eb.json> -components comp.tsv
```

**GOTCHA 3 — grep for `^BANNED`, not `banned_claim`.** The tool prints `BANNED`
and `NUMBER` line prefixes; `banned_claim` is the JSON `check` value and appears
nowhere in the CLI output. Grepping for it returns 0 on every site — a false
all-clear that looks exactly like a clean estate.

**GOTCHA 4 — some scan outputs are non-UTF-8** (site copy carries extended
ASCII), and plain `grep -c` returns **empty with no error** on them. Use
`LC_ALL=C grep -ac`.

**GOTCHA 5 — with an evidence file that has `banned_claims` but no `facts[]`,
every number on the page becomes a `NUMBER` finding.** That is correct behaviour
(nothing supports them) but it is noise for this question. Filter to `^BANNED`.

## 4. Positive control — required, both directions

A 0-findings result and a broken harness are indistinguishable. Build a synthetic
TSV with sentences the set **must** block and legitimate ones it **must not**:

```python
import base64
cases = [("ctl_block_1","hero","<p>A claim without a source does not appear here.</p>"),
         ("ctl_block_2","body","<p>Every figure is verified before publication.</p>"),
         ("ctl_pass_1","body","<p>We cite each figure and date it.</p>"),
         ("ctl_pass_2","body","<p>The statute is the authoritative text.</p>")]
for page,slot,html in cases:
    print("%s\t%s\t%s" % (page,slot,base64.b64encode(html.encode()).decode()))
```

2026-07-28: **6 of 6 block cases fired, 3 of 3 legitimate sentences passed.**

**And add the negated form of each pattern to the pass-list.** 226's own test was
"10 fabrication shapes blocked, 13 legitimate sentences passed" — and it still
missed this, because the pass-list contained no sentence that *negates* one of its
own patterns. That is the whole finding of this workstream.

## 5. Extract the candidate universal set from the SQL, not by hand

```bash
python3 -c "
import json,re
src=open('docs/agent_docs/sql_for_agents/226_overclaim_patterns_oufe.sql').read()
json.dump({'banned_claims':json.loads(re.search(r'\\\$add\\\$(.*?)\\\$add\\\$',src,re.S).group(1))},
          open('universal.json','w'),indent=1)"
```

Then confirm it is **live**, not just committed (the seed is not the system):

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c \
 "SELECT count(*) FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
         jsonb_array_elements(ss.data->'banned_claims') bc
   WHERE s.domain='oufe.com' AND ss.aspect='evidence_base' AND ss.is_current
     AND bc->>'pattern' LIKE '%is not a disclaimer%';"    # 1
```

**The set is 10 patterns, not 11**, and pattern 7 contains the literal
alternative `oufe` — it is not universalisable verbatim.

## 6. How a fix must be verified (from `104`, unchanged)

Induce **both** directions on a site with no register: a page asserting "every
claim on this site is verified" must fail with a `claims` blocker; a legitimate
process sentence ("we cite each figure and date it") must still build. Add a
third case now required by this session's finding: **"where a figure has not been
independently verified, that is stated" must still build.**

---

## 7. After the 2026-07-28 change — claimscan includes the fleet-wide set by default

```bash
go build -o /tmp/claimscan ./cmd/claimscan
/tmp/claimscan -components comp.tsv                     # fleet-wide only, as an UNARMED site is scanned
/tmp/claimscan -evidence eb.json -components comp.tsv   # fleet-wide + that site's own = what the gate enforces
/tmp/claimscan -evidence cand.json -no-global -components comp.tsv   # a CANDIDATE set in isolation
```

`-evidence` is now optional; it prints the fleet-wide pattern count to stderr so a
silently empty set cannot look like a clean estate. **Use `-no-global` to reproduce
this workstream's original numbers** — they were measured before the set existed.

## 8. Two checks this session learned the hard way

**Verify the COMMIT compiles, not your tree.** A pathspec commit of a file another
session is also editing carries their uncommitted work — and if you commit the
consumer of a type whose definition is still in their working tree, HEAD stops
compiling while your own tests stay green. `make build-<service>` builds from HEAD,
so this breaks everyone's next image build.

```bash
git archive HEAD | tar -x -C /tmp/headcheck && (cd /tmp/headcheck && go build ./platform/...)
```

Run it straight after committing platform code. The tell in the diff: a hunk whose
**context lines are code you did not write**, or insertion counts larger than the
edits you remember making.

**Quote live copy from the source, never from claimscan's output.** Its snippets are
elided with `…`, so retyping one produces a plausible sentence the site never
published. Two regression fixtures and two council `grounded_in` quotes were wrong
this way. Extract verbatim instead:

```bash
python3 -c "
import base64,re,sys
for line in open('comp_<site>.tsv'):
    p=line.rstrip('\n').split('\t')
    if len(p)<3: continue
    html=base64.b64decode(p[2]).decode('utf-8','replace')
    for m in re.finditer(r'[^.<>]*<phrase>[^.<>]*\.', html): print(repr(m.group(0).strip()))"
```
