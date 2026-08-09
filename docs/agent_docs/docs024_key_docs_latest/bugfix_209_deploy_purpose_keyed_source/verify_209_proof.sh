#!/usr/bin/env bash
# 209/231 behavioural proof — VERIFY. Run after a legacy-pair build reaches deploy.
# The assertion that matters: hero and logo were deployed from their OWN assets,
# so the two committed files must differ in BYTES. Identical bytes = the 231
# shadow (logo-as-hero) that migration 348 was meant to close.
set -uo pipefail

DOMAIN="cookly.uk"
SITES_REPO="$HOME/projects/sites"

echo "=== 1. assets rows for $DOMAIN (identity: one row per purpose) ==="
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db <<SQL
SELECT a.id, a.purpose, a.filename,
       left(COALESCE(a.storage_path,''),64) AS storage_path,
       left(COALESCE(a.url,''),64)          AS url,
       a.created_at
FROM assets a JOIN sites s ON s.id = a.site_id
WHERE s.domain = '$DOMAIN'
ORDER BY a.created_at;
SQL

echo
echo "=== 2. sites row: content_data URLs + brand_assets ==="
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db <<SQL
SELECT COALESCE(logo_url,'-') AS col_logo_url,
       COALESCE(brand_assets::text,'-') AS brand_assets,
       github_repo, build_status, status
FROM sites WHERE domain = '$DOMAIN';
SQL

echo
echo "=== 3. committed artefacts in gqls/sites ==="
git -C "$SITES_REPO" fetch origin master --quiet 2>/dev/null
git -C "$SITES_REPO" ls-tree -r --name-only origin/master -- "$DOMAIN/" 2>/dev/null | grep -E 'assets/images/' || echo "(no image artefacts yet)"

echo
echo "=== 4. THE ASSERTION: hero bytes != logo bytes ==="
tmp=$(mktemp -d)
hero_path=$(git -C "$SITES_REPO" ls-tree -r --name-only origin/master -- "$DOMAIN/" 2>/dev/null | grep -E 'assets/images/.*hero' | head -1)
logo_path=$(git -C "$SITES_REPO" ls-tree -r --name-only origin/master -- "$DOMAIN/" 2>/dev/null | grep -E 'assets/images/logo' | head -1)
echo "hero artefact: ${hero_path:-<none>}"
echo "logo artefact: ${logo_path:-<none>}"
if [ -n "$hero_path" ] && [ -n "$logo_path" ]; then
  git -C "$SITES_REPO" show "origin/master:$hero_path" > "$tmp/hero.bin" 2>/dev/null
  git -C "$SITES_REPO" show "origin/master:$logo_path" > "$tmp/logo.bin" 2>/dev/null
  hs=$(sha256sum "$tmp/hero.bin" | cut -d' ' -f1); ls_=$(sha256sum "$tmp/logo.bin" | cut -d' ' -f1)
  echo "hero sha256: $hs  ($(stat -c%s "$tmp/hero.bin") bytes)"
  echo "logo sha256: $ls_  ($(stat -c%s "$tmp/logo.bin") bytes)"
  if [ "$hs" = "$ls_" ]; then echo "RESULT: ***FAIL*** identical bytes — the 231 shadow is PRESENT"
  else echo "RESULT: PASS — distinct bytes, each purpose deployed its own asset"; fi
else
  echo "RESULT: INCONCLUSIVE — one or both artefacts absent (did the plan ask for both?)"
fi
rm -rf "$tmp"
