#!/bin/bash
# ============================================================================
# 082: Add hero sweep effect to leopardessconsulting.co.uk index page
#
# This updates the rendered_html of the hero page_component to include:
# - Separated image layer (background) from overlay (div)
# - Sweep beam div with animation CSS
# - JS: first click triggers a diagonal light sweep, once only
#
# Then triggers a page reassembly + deploy via rerender.
# ============================================================================

KAFKA_POD="personae-kafka-cluster-kafka-0"
KAFKA_NS="kafka"
BOOTSTRAP="personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092"
CLIENT_ID="demo_client"
DOMAIN="leopardessconsulting.co.uk"

# ============================================================================
# STEP 1: Update the rendered_html via SQL
# ============================================================================

echo "=== Step 1: Updating hero rendered_html with sweep effect ==="

# Get the DB pod
DB_POD=$(kubectl -n ai-persona-system get pods -l app=postgres -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
if [ -z "$DB_POD" ]; then
  DB_POD=$(kubectl -n ai-persona-system get pods -l app.kubernetes.io/name=postgresql -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
fi
if [ -z "$DB_POD" ]; then
  DB_POD=$(kubectl -n ai-persona-system get pods | grep postgres | head -1 | awk '{print $1}')
fi
echo "  DB Pod: $DB_POD"

# First, check current hero component
kubectl -n ai-persona-system exec -i "$DB_POD" -- psql -U clients_user -d clients_db -t -A <<'CHECKSQL'
SELECT pc.id, pc.slot_name, length(pc.rendered_html) as html_len
FROM page_components pc
JOIN pages p ON pc.page_id = p.id
JOIN sites s ON p.site_id = s.id
WHERE s.domain = 'leopardessconsulting.co.uk'
  AND p.name = 'index'
  AND pc.slot_name = 'hero';
CHECKSQL

echo ""
echo "  Updating rendered_html..."

# Apply the update — replace the hero section HTML
kubectl -n ai-persona-system exec -i "$DB_POD" -- psql -U clients_user -d clients_db <<'UPDATESQL'
UPDATE page_components
SET rendered_html = '<section class="hero" data-component="hero" style="background: url(''/assets/images/hero.jpg'') center/cover no-repeat;">
    <div class="hero-overlay" id="heroOverlay"></div>
    <div class="hero-content">
        <h1>Smart AI Teams That Build Websites, Tools &amp; Reports in Minutes</h1>
        <p class="hero-subheadline">Leopardess Consulting harnesses collaborative AI agent teams to deliver professional websites, custom calculators, business plans, and creative tools — in minutes, not weeks.</p>
        <a href="/contact.html" class="btn btn-primary">Get Started</a>
        <a href="/services.html" class="btn btn-secondary">Our Services</a>
    </div>
</section>
<style>
.hero {
    min-height: 70vh;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 4rem 2rem;
    position: relative;
    overflow: hidden;
    cursor: pointer;
}
.hero-overlay {
    position: absolute;
    inset: 0;
    z-index: 1;
    pointer-events: none;
    background: linear-gradient(135deg, rgba(26,26,46,0.85) 0%, rgba(22,33,62,0.8) 50%, rgba(15,52,96,0.75) 100%);
}
/* Mask-based sweep: a transparent band in the mask reveals the image beneath */
@keyframes maskSweep {
    0%   { -webkit-mask-position: 0% 0%; mask-position: 0% 0%; }
    100% { -webkit-mask-position: 100% 100%; mask-position: 100% 100%; }
}
.hero-overlay.sweeping {
    -webkit-mask-image: linear-gradient(
        115deg,
        black 0%, black 30%,
        rgba(0,0,0,0.5) 36%,
        rgba(0,0,0,0.1) 42%,
        transparent 46%,
        transparent 54%,
        rgba(0,0,0,0.1) 58%,
        rgba(0,0,0,0.5) 64%,
        black 70%, black 100%
    );
    mask-image: linear-gradient(
        115deg,
        black 0%, black 30%,
        rgba(0,0,0,0.5) 36%,
        rgba(0,0,0,0.1) 42%,
        transparent 46%,
        transparent 54%,
        rgba(0,0,0,0.1) 58%,
        rgba(0,0,0,0.5) 64%,
        black 70%, black 100%
    );
    -webkit-mask-size: 300% 300%;
    mask-size: 300% 300%;
    -webkit-mask-repeat: no-repeat;
    mask-repeat: no-repeat;
    animation: maskSweep 2.6s cubic-bezier(0.22, 0.1, 0.25, 1) forwards;
}
.hero-content {
    max-width: 900px;
    margin: 0 auto;
    color: #fff;
    z-index: 3;
    position: relative;
}
.hero h1 {
    font-size: clamp(2rem, 5vw, 3.5rem);
    font-weight: 700;
    margin-bottom: 1.5rem;
    line-height: 1.2;
    color: #fff;
    text-shadow: 0 2px 4px rgba(0,0,0,0.3);
}
.hero-subheadline {
    font-size: clamp(1rem, 2vw, 1.35rem);
    margin-bottom: 2rem;
    opacity: 0.95;
    line-height: 1.6;
    color: rgba(255,255,255,0.95);
}
.hero .btn {
    display: inline-block;
    padding: 0.875rem 2rem;
    margin: 0.5rem;
    border-radius: 4px;
    text-decoration: none;
    font-weight: 600;
    font-size: 1rem;
    transition: all 0.2s ease;
}
.hero .btn-primary {
    background: var(--accent-color, #0f3460);
    color: #fff;
    border: 2px solid var(--accent-color, #0f3460);
}
.hero .btn-primary:hover {
    background: transparent;
    color: #fff;
}
.hero .btn-secondary {
    background: transparent;
    color: #fff;
    border: 2px solid rgba(255,255,255,0.8);
}
.hero .btn-secondary:hover {
    background: rgba(255,255,255,0.1);
}
@media (max-width: 768px) {
    .hero { min-height: 60vh; padding: 3rem 1.5rem; }
    .hero .btn { display: block; width: 100%; max-width: 280px; margin: 0.5rem auto; }
}
</style>
<script>
(function() {
    var hero = document.querySelector(''.hero[data-component="hero"]'');
    if (!hero) return;
    var overlay = document.getElementById(''heroOverlay'');
    var triggered = false;
    hero.addEventListener(''click'', function() {
        if (triggered) return;
        triggered = true;
        overlay.classList.add(''sweeping'');
    });
})();
</script>',
    updated_at = NOW()
WHERE id IN (
    SELECT pc.id
    FROM page_components pc
    JOIN pages p ON pc.page_id = p.id
    JOIN sites s ON p.site_id = s.id
    WHERE s.domain = 'leopardessconsulting.co.uk'
      AND p.name = 'index'
      AND pc.slot_name = 'hero'
    LIMIT 1
);
UPDATESQL

echo "  Update result: $?"

# ============================================================================
# STEP 2: Trigger page reassembly + deploy via rerender
# ============================================================================

echo ""
echo "=== Step 2: Triggering page rerender for index.html ==="

CORR_ID=$(uuidgen)
ORCH_ID=$(uuidgen)
REQ_ID=$(uuidgen)
MSG_ID=$(uuidgen)
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Get site_id
SITE_ID=$(kubectl -n ai-persona-system exec -i "$DB_POD" -- psql -U clients_user -d clients_db -t -A -c \
  "SELECT id FROM sites WHERE domain = '$DOMAIN' LIMIT 1;")
echo "  Site ID: $SITE_ID"

kubectl -n kafka run -i --rm kcat-hero-sweep-$(date +%s) \
  --image=edenhill/kcat:1.7.1 \
  --restart=Never -- \
  kcat -P \
  -b $BOOTSTRAP \
  -t system.agent.generic.requests \
  -H correlation_id=$CORR_ID \
  -H orchestration_id=$ORCH_ID \
  -H request_id=$REQ_ID \
  -H message_id=$MSG_ID \
  -H message_type=request \
  -H client_id=$CLIENT_ID \
  -H action=process \
  -H sender_agent_type=cli \
  -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses \
  -H timestamp=$TIMESTAMP <<JSON
{"headers":{"correlation_id":"${CORR_ID}","orchestration_id":"${ORCH_ID}","request_id":"${REQ_ID}","message_id":"${MSG_ID}","message_type":"request","client_id":"${CLIENT_ID}","action":"process","sender":{"agent_id":"cli-user","agent_type":"cli","pod_name":"cli"},"timestamp":"${TIMESTAMP}"},"config":{"workflow":{"start_step":"spawn_rerenderer","processing_mode":"orchestrator","timeout_seconds":300,"steps":{"spawn_rerenderer":{"action":"spawn_agent","config":{"role":"rerenderer","agent_type":"rerender-pages"},"output_field":"rerenderer_agent","next_step":"call_rerender","description":"Spawn page rerenderer"},"call_rerender":{"action":"call_agent","config":{"agent_type":"rerender-pages","target_role":"rerenderer","input_mapping":{"domain":"input_data.domain","site_id":"input_data.site_id","page_filter":"input_data.page_filter"},"timeout_seconds":240},"output_field":"rerender_result","next_step":"complete","description":"Rerender index page"},"complete":{"action":"complete_workflow","config":{"output_fields":["rerender_result"]},"description":"Rerender complete"}}}},"input_data":{"domain":"${DOMAIN}","site_id":"${SITE_ID}","page_filter":"index"}}
JSON

echo ""
echo "  CORRELATION_ID=$CORR_ID"
echo ""
echo "  Monitor:"
echo "    kubectl -n ai-persona-system logs -l app=agent-chassis --tail=100 | grep '$CORR_ID'"
echo ""
echo "=== Done ==="