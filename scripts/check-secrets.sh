#!/usr/bin/env bash
# Block real credentials from being committed. This repo (gqls/agentchassis) is PUBLIC.
#
# Written 2026-07-14 after real AWS SES SMTP credentials and the real idea.uk
# INTERNAL_API_KEY were found sitting on a public origin/main since 2026-06-04,
# inside a file named "idea.env.example" — the .example suffix is why nobody looked.
#
# The patterns match on LENGTH, not on prefix, because truncated illustrative keys
# ("sk_live_...", "whsec_…") are legitimate in docs and must not trip the guard.
# Only full-length values — i.e. ones that would actually authenticate — are blocked.
#
# Usage:  scripts/check-secrets.sh            # scan staged changes (pre-commit)
#         scripts/check-secrets.sh --all      # scan the whole working tree
set -uo pipefail

# pattern|human name
PATTERNS=(
  'AKIA[0-9A-Z]{16}|AWS access key id'
  'sk-ant-api[A-Za-z0-9_-]{40,}|Anthropic API key'
  '(sk|rk)_live_[A-Za-z0-9]{40,}|Stripe LIVE key'
  '(sk|rk)_test_[A-Za-z0-9]{40,}|Stripe test key'
  'whsec_[A-Za-z0-9+/=]{30,}|Stripe webhook signing secret'
  '^[[:space:]]*INTERNAL_API_KEY=[0-9a-f]{64}|idea.uk INTERNAL_API_KEY (full 32-byte hex)'
  '^[[:space:]]*SMTP_PASS=[A-Za-z0-9+/]{40,}|SES SMTP password'
  # anchored to column 0: a real PEM file starts the header there, whereas prose and
  # code comments that merely mention the format ("...OpenSSH PEM (\"-----BEGIN...\")") do not.
  '^-----BEGIN [A-Z ]*PRIVATE KEY-----|private key'
)

if [[ "${1:-}" == "--all" ]]; then
  files=$(git ls-files)
else
  files=$(git diff --cached --name-only --diff-filter=ACM)
fi
[[ -z "$files" ]] && exit 0

found=0
while IFS= read -r file; do
  [[ -f "$file" ]] || continue
  for entry in "${PATTERNS[@]}"; do
    pat="${entry%%|*}"; name="${entry##*|}"
    if grep -qEI -- "$pat" "$file" 2>/dev/null; then
      line=$(grep -nEI -m1 -- "$pat" "$file" 2>/dev/null | cut -d: -f1)
      echo "  BLOCKED  $file:$line  — looks like a real $name"
      found=1
    fi
  done
done <<< "$files"

if [[ $found -eq 1 ]]; then
  cat <<'EOF'

A full-length credential was found. This repo is PUBLIC — do not commit it.

  * Real secrets belong only on the box (e.g. /etc/idea/idea.env), never in git.
  * In docs and .example files, truncate: STRIPE_SECRET_KEY=sk_live_REPLACE_ME
  * If this is a genuine false positive:  git commit --no-verify

If a real secret has ALREADY been pushed, deleting the file is not enough —
it is in history. ROTATE the credential; that is what actually closes it.
EOF
  exit 1
fi
exit 0
