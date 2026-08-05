# 034 — the GitHub Action that deploys the sites repo to B2

Refreshed 2026-08-05 (bug 120 fix: changed domains now derived from the push
range `github.event.before..github.sha`, not `HEAD~1`). The live file
`.github/workflows/deploy-to-b2.yml` in `gqls/sites@master` is the authority;
this is a reading copy.

```yaml
name: Deploy to B2
on:
  push:
    branches: [master]
    paths-ignore:
      - '.github/**'
      - 'README.md'
jobs:
  deploy:
    runs-on: self-hosted
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0        # range diff needs the 'before' commit; runner workspace persists, so full history is a one-time cost (pack = 86.67 MiB today)

      - name: Get changed domains
        id: changed
        env:
          BEFORE: ${{ github.event.before }}
          AFTER: ${{ github.sha }}
        run: |
          if [ "$BEFORE" = "0000000000000000000000000000000000000000" ] \
             || ! git cat-file -e "$BEFORE^{commit}" 2>/dev/null; then
            echo "No usable 'before' (first push to branch, or force-push discarded it) — falling back to ALL domains"
            CHANGED=""
          else
            echo "Push range: $BEFORE..$AFTER"
            if git rev-parse -q --verify "$AFTER^2" >/dev/null; then
              echo "Tip is a merge commit — the range diff spans BOTH sides (bugs_closed/120)"
            fi
            CHANGED=$(git diff --name-only "$BEFORE" "$AFTER" | grep -E '^[^/]+\.[^/]+/' | cut -d'/' -f1 | sort -u | tr '\n' ' ' || echo "")
          fi
          if [ -z "$CHANGED" ]; then
            CHANGED=$(ls -d */ 2>/dev/null | grep -E '^[^/]+\.[^/]+/$' | tr -d '/' | tr '\n' ' ' || echo "")
            echo "Falling back to all domains"
          fi
          echo "domains=$CHANGED" >> $GITHUB_OUTPUT
          echo "Changed domains: $CHANGED"

      - name: Sync to B2
        env:
          B2_APPLICATION_KEY_ID: ${{ secrets.B2_APPLICATION_KEY_ID }}
          B2_APPLICATION_KEY: ${{ secrets.B2_APPLICATION_KEY }}
        run: |
          if [ -z "$B2_APPLICATION_KEY_ID" ]; then
            echo "ERROR: B2_APPLICATION_KEY_ID secret is not set!"
            exit 1
          fi
          if [ -z "$B2_APPLICATION_KEY" ]; then
            echo "ERROR: B2_APPLICATION_KEY secret is not set!"
            exit 1
          fi
          echo "Secrets are configured"

          b2 account authorize "$B2_APPLICATION_KEY_ID" "$B2_APPLICATION_KEY"

          for domain in ${{ steps.changed.outputs.domains }}; do
            if [ -d "$domain" ]; then
              echo "Syncing $domain to B2..."
              b2 sync --delete --skip-newer "$domain" "b2://portfolio-sites/$domain"
            else
              echo "WARNING: $domain in changed set but no directory — skipped"
            fi
          done

      - name: Purge Cloudflare cache
        if: ${{ steps.changed.outputs.domains != '' }}
        
        env:
          CF_API_TOKEN: ${{ secrets.CF_API_TOKEN }}
        run: |
          for domain in ${{ steps.changed.outputs.domains }}; do
            echo "Purging cache for $domain..."
            ZONE_ID=$(curl -s -X GET "https://api.cloudflare.com/client/v4/zones?name=$domain" \
              -H "Authorization: Bearer $CF_API_TOKEN" \
              -H "Content-Type: application/json" | jq -r '.result[0].id')

            if [ "$ZONE_ID" != "null" ] && [ -n "$ZONE_ID" ]; then
              curl -X POST "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/purge_cache" \
                -H "Authorization: Bearer $CF_API_TOKEN" \
                -H "Content-Type: application/json" \
                --data '{"purge_everything":true}'
            fi
          done
```

The sibling `gqls/vm-sites@main` `.github/workflows/deploy-to-vm.yml` carries a
byte-identical "Get changed domains" step (same fix, same day) with an rsync
target instead of `b2 sync`.
