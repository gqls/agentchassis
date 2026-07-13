Here's the exact go-live sequence for this box (116.203.204.115, Nuremberg). It assumes you picked an Ubuntu image — confirm that, since setup.sh uses apt/ufw; if it's Debian it'll still work, anything else won't.
1. DNS first (certbot needs it resolving before you run setup). At your DNS provider for idea.uk:

A record: idea.uk → 116.203.204.115
AAAA record (dual-stack): idea.uk → the box's IPv6. Check it on the box with ip -6 addr show; on Hetzner it's typically 2a01:4f8:1c18:7c31::1.

Then verify from your laptop before going further:
bashdig +short idea.uk          # should return 116.203.204.115
dig +short AAAA idea.uk      # should return the v6 address
2. (Recommended) put an SSH key on the box if you're currently on the root password Hetzner emailed you. setup.sh only disables password login if a key is present (anti-lockout), so add one first or it'll leave password auth on:
   bashssh-copy-id root@116.203.204.115
   Skip this if you already added a key at creation.
3. Build the binary (on your laptop — native amd64, static, stdlib-only so GOPROXY=off is fine):
   bashcd ~/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files
   GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 go build -o idea .
4. Copy the binary + deploy artefacts up:
   bashscp idea deploy/setup.sh deploy/idea.env.example root@116.203.204.115:/root/
5. Fill the env on the box (do this before setup.sh so the service starts healthy):
   bashssh root@116.203.204.115
   mkdir -p /etc/idea
   cp /root/idea.env.example /etc/idea/idea.env
   nano /etc/idea/idea.env
   Set at minimum: ANTHROPIC_API_KEY, PUBLIC_BASE_URL=https://idea.uk, OPERATOR_EMAIL, INTERNAL_API_KEY (generate with openssl rand -hex 32), AUTO_DELIVER=false, REPORT_PRICE_GBP=29. Leave the two STRIPE_* lines blank for now — that starts it on the FakeProvider so you can test the whole flow without real money. You'll add Stripe in step 7.
6. Run setup.sh (still on the box):
   bashDOMAIN=idea.uk LETSENCRYPT_EMAIL=you@example.com IDEA_BINARY_PATH=/root/idea bash /root/setup.sh
   This installs nginx + TLS (certbot will issue the cert now that DNS resolves) + ufw + fail2ban + the hardened systemd unit, and starts the service. Then verify:
   bashsystemctl status idea
   journalctl -u idea -n 50 --no-pager
   curl -sS https://idea.uk/health
   If /health is good, the public site is live and you can hit the audience-check taster in a browser.
7. Stripe, when you're ready to take money: in the Stripe dashboard add a webhook endpoint https://idea.uk/stripe/webhook for checkout.session.completed, then put the sk_test_… and whsec_… into /etc/idea/idea.env and systemctl restart idea. Test cards first; then swap to live keys and run one real order with your own card, refunded via /refund.
8. Walk one order end to end on the FakeProvider before Stripe — request → confirm → (fake pay) → engine runs → you review the report → deliver.
   Two notes specific to this setup. Hetzner has its own Cloud Firewall that's separate from the box's ufw, and by default no Cloud Firewall is applied, so the ufw rules setup.sh sets (22/80/443) are your active control — nothing extra needed, though you can add a Cloud Firewall too for belt-and-braces. And because the box is dual-stack and setup.sh's nginx already listens on both [::]:80/443 and IPv4, certbot will validate fine over whichever Let's Encrypt picks.
   When you run it, paste any errors from journalctl -u idea plus what curl https://idea.uk/health returns — that's what we harden in setup.sh before it becomes the service-deployer payload. Most likely first wrinkle is the env (an empty ANTHROPIC_API_KEY shows as the service starting but failing real work), so check that first if /health is off.




systemctl restart idea
systemctl status idea
journalctl -u idea -n 20 --no-pager
curl -sS https://idea.uk/health

run taster
curl -sS "https://idea.uk/audience-check?domain=agritec.uk&audience=UK%20small%20farmers"