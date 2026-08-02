#!/usr/bin/env bash
# ufw-cloudflare-lockdown.sh — Option B step 5, and ONLY step 5.
#
# Restricts 80/443 to Cloudflare's published ranges so the origin cannot be
# reached around the proxy. A proxy in front of an open origin is decoration:
# the origin IPs (116.203.204.115 / 2a01:4f8:1c18:7c31::1) are public — they
# were the A/AAAA records, and were additionally exposed as webdesign.uk's and
# ugg2.com's origin on 2026-07-31.
#
# RUN ONLY AFTER: records are orange AND the two-network real-IP proof has
# passed AND checkout incl. /stripe/webhook has been re-tested through the
# proxy. Run early and you cut off every direct visitor while DNS still points
# people at the origin.
#
# ORDER INSIDE THIS SCRIPT IS THE SAFETY: CF allows are added BEFORE the
# generic allows are deleted, so there is no window with Cloudflare shut out.
# SSH (22/OpenSSH) is never touched — a lockout here has no rescue path but
# the Hetzner console.
#
# ⚠ setup.sh's full provision runs `ufw --force reset` and re-opens 80/443 to
# the world. After ANY re-provision, re-run this script (post-cutover boxes
# only). Ranges verified against https://www.cloudflare.com/ips-{v4,v6}
# 2026-08-02 — re-fetch and diff before running if that is no longer recent.
#
# Rollback: ufw allow 80/tcp && ufw allow 443/tcp  (returns to the pre-lockdown
# state; the per-range allows become redundant but harmless).

set -euo pipefail

RANGES_V4="173.245.48.0/20 103.21.244.0/22 103.22.200.0/22 103.31.4.0/22
141.101.64.0/18 108.162.192.0/18 190.93.240.0/20 188.114.96.0/20
197.234.240.0/22 198.41.128.0/17 162.158.0.0/15 104.16.0.0/13
104.24.0.0/14 172.64.0.0/13 131.0.72.0/22"
RANGES_V6="2400:cb00::/32 2606:4700::/32 2803:f800::/32 2405:b500::/32
2405:8100::/32 2a06:98c0::/29 2c0f:f248::/32"

# Refuse to run if SSH is not explicitly allowed — the one unrecoverable state.
ufw status | grep -qE '^22/tcp.*ALLOW|OpenSSH.*ALLOW' || {
  echo "REFUSING: no SSH allow rule visible in ufw status" >&2; exit 1; }

echo "== adding Cloudflare allows (before any delete) =="
for r in $RANGES_V4 $RANGES_V6; do
  ufw allow from "$r" to any port 80,443 proto tcp comment 'cloudflare-only'
done

echo "== removing the generic world-open allows =="
ufw delete allow 80/tcp
ufw delete allow 443/tcp

echo "== resulting rules =="
ufw status numbered

echo "VERIFY from a non-Cloudflare network:"
echo "  curl -sk --max-time 8 https://116.203.204.115/   -> must TIME OUT"
echo "  curl -s  -o /dev/null -w '%{http_code}' https://idea.uk/  -> 200 (via CF)"
