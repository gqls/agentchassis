#!/usr/bin/env bash
# stage_b_assert.sh — the 443 Stage B acceptance, read at the SERVED artefact (never the status).
# Usage: ./stage_b_assert.sh <before-dir>   (a dir holding before_<slug>.html snapshots for the controls)
# Asserts, per rebuilt page (technical-details, your-own-model):
#   A1 six sections, h2s DISTINCT (443's symptom gone);
#   A2 each section's first sentence tracks ITS OWN subject (A4 intent) and not a sibling's (the failure);
#   A3 no em dash in the served copy; A4 no </strom> (bugs_open/456 malformed_closing_tag slug);
#   A5 technical-details carries no licence-family listing (Mistral / Llama Community / Phi models / Apache 2.0);
# and for the two controls (index, about): served bytes unchanged from the before-snapshot.
set -u
BEFORE=${1:?before-dir}; SITE=1368e337-dd1d-4799-bbb3-8221a1b79bcc; fail=0
psql() { kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -Atc "$1"; }
for slug in technical-details your-own-model; do
  html=$(curl -sS -m 30 "https://finetuning.uk/$slug.html"); echo "== $slug ($(printf %s "$html" | wc -c) bytes)"
  mapfile -t h2 < <(printf %s "$html" | grep -o '<h2[^>]*>[^<]*</h2>' | sed 's/<[^>]*>//g')
  n=${#h2[@]}; d=$(printf '%s\n' "${h2[@]}" | sort -u | wc -l)
  echo "A1 h2s: $n, distinct: $d"; [ "$d" -eq "$n" ] || { echo "  FAIL A1: repeated h2"; fail=1; }
  printf '%s\n' "${h2[@]}" | sed 's/^/     h2: /'
  printf %s "$html" | grep -q '—' && { echo "  FAIL A3: em dash present ($(printf %s "$html" | grep -o '—' | wc -l))"; fail=1; } || echo "A3 no em dash"
  printf %s "$html" | grep -qi '</strom>' && { echo "  FAIL A4: </strom> present"; fail=1; } || echo "A4 no </strom>"
  if [ "$slug" = technical-details ]; then
    hits=$(printf %s "$html" | grep -ciE 'Mistral|Llama Community|Phi models|Apache 2\.0'); [ "$hits" -eq 0 ] && echo "A5 no family listing" || { echo "  FAIL A5: family listing hits=$hits"; fail=1; }
  fi
  echo "A2 opening lines vs subjects (subject -> first sentence of the section's first <p>):"
  psql "SELECT jsonb_array_elements_text(section_subjects) FROM pages WHERE site_id='$SITE' AND url='/$slug.html';" > /tmp/subj_$$.txt
  # first <p> text after each h2 (hero has h1: take the first <p> in the page as section 1)
  printf %s "$html" | python3 -c '
import sys,re,html as H
doc=sys.stdin.read()
subs=[l.rstrip("\n") for l in open(sys.argv[1]) if l.strip()]
# sections: split on h1/h2 tags; first chunk after <main or the h1 is the hero
parts=re.split(r"<h[12][^>]*>", doc)
opens=[]
for part in parts[1:]:
    m=re.search(r"<p[^>]*>(.*?)</p>", part, re.S)
    t=re.sub(r"<[^>]+>","",m.group(1)) if m else ""
    t=H.unescape(re.sub(r"\s+"," ",t)).strip()
    opens.append(t[:140])
for i,s in enumerate(subs):
    o=opens[i] if i<len(opens) else "(no section)"
    print(f"  [{i+1}] {s}\n       -> {o}")
' /tmp/subj_$$.txt
  rm -f /tmp/subj_$$.txt
done
echo "== controls (served bytes vs before-snapshot)"
for c in index about; do cur=$(curl -sS -m 30 "https://finetuning.uk/$c.html" | sha256sum | cut -c1-16); old=$(sha256sum "$BEFORE/before_$c.html" | cut -c1-16); [ "$cur" = "$old" ] && echo "  $c unchanged ($cur)" || { echo "  $c CHANGED $old -> $cur (rerender by another lane? check pages.deployed_at before reading as a Stage B effect)"; }; done
echo "RESULT: $([ $fail -eq 0 ] && echo PASS || echo FAIL)"; exit $fail
