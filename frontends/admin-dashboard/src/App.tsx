import { useState, useEffect, useCallback } from "react";

const API_BASE = "/api/v1/admin";

// Mock data for preview — remove when connected to real API
const MOCK = {
    sites: [
        { id: "1368e337", domain: "finetuning.uk", company_name: "FineTuning", status: "deployed", email: "finetuning@contactforsales.com", work_items: { done: 75, active: 1, ready: 4, review: 2, failed: 0, blocked: 0 } },
        { id: "5fe15466", domain: "gaswholesalers.com", company_name: "Gas Wholesalers", status: "deployed", email: "gas@contactforsales.com", work_items: { done: 75, active: 0, ready: 6, review: 1, failed: 0, blocked: 0 } },
        { id: "2a8ebf9c", domain: "ai-agent-orchestration.com", company_name: "AI Agent Orchestration", status: "deployed", email: "agents@contactforsales.com", work_items: { done: 32, active: 1, ready: 6, review: 0, failed: 0, blocked: 0 } },
        { id: "4851f6fc", domain: "leopardessconsulting.co.uk", company_name: "Leopardess Consulting", status: "deployed", email: "leopardess@contactforsales.com", work_items: { done: 33, active: 1, ready: 18, review: 0, failed: 0, blocked: 0 } },
    ],
    reviewItems: [
        { id: "aaa-111", site_id: "1368e337", domain: "finetuning.uk", item_type: "placeholder_content", status: "needs_human_review", summary: "Page about has placeholder content that needs real data", spec: { page_name: "about", section: "leadership-team", missing_data: "Team member names, titles, photos, bios" }, created_at: "2026-03-12T20:00:00Z" },
        { id: "aaa-222", site_id: "1368e337", domain: "finetuning.uk", item_type: "placeholder_content", status: "needs_human_review", summary: "Page contact has placeholder content", spec: { page_name: "contact", missing_data: "Contact form configuration" }, created_at: "2026-03-12T20:01:00Z" },
        { id: "aaa-333", site_id: "5fe15466", domain: "gaswholesalers.com", item_type: "placeholder_content", status: "needs_human_review", summary: "Page contact has placeholder content", spec: { page_name: "contact", missing_data: "Contact address details" }, created_at: "2026-03-12T20:02:00Z" },
    ],
};

const useMock = true; // Toggle to false when API is connected

async function api(path, options = {}) {
    if (useMock) return null;
    const res = await fetch(`${API_BASE}${path}`, {
        headers: { "Content-Type": "application/json", ...options.headers },
        ...options,
    });
    if (!res.ok) throw new Error(`${res.status}: ${await res.text()}`);
    return res.json();
}

// ============================================================================
// Components
// ============================================================================

function Badge({ children, variant = "default" }) {
    const colors = {
        default: "bg-gray-100 text-gray-700",
        success: "bg-emerald-50 text-emerald-700",
        warning: "bg-amber-50 text-amber-700",
        error: "bg-red-50 text-red-700",
        info: "bg-blue-50 text-blue-700",
        review: "bg-purple-50 text-purple-700",
    };
    return (
        <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${colors[variant] || colors.default}`}>
      {children}
    </span>
    );
}

function StatusBadge({ status }) {
    const map = {
        complete: { label: "Complete", variant: "success" },
        triaged: { label: "Ready", variant: "info" },
        claimed: { label: "Active", variant: "warning" },
        needs_human_review: { label: "Needs Review", variant: "review" },
        failed: { label: "Failed", variant: "error" },
        blocked: { label: "Blocked", variant: "error" },
        deployed: { label: "Deployed", variant: "success" },
        active: { label: "Active", variant: "info" },
    };
    const { label, variant } = map[status] || { label: status, variant: "default" };
    return <Badge variant={variant}>{label}</Badge>;
}

function Card({ children, className = "", onClick }) {
    return (
        <div
            onClick={onClick}
            className={`bg-white rounded-lg border border-gray-200 p-4 ${onClick ? "cursor-pointer hover:border-gray-300 hover:shadow-sm transition-all" : ""} ${className}`}
        >
            {children}
        </div>
    );
}

function Button({ children, variant = "primary", size = "md", onClick, disabled }) {
    const base = "inline-flex items-center justify-center rounded font-medium transition-colors";
    const variants = {
        primary: "bg-gray-900 text-white hover:bg-gray-800",
        secondary: "bg-white text-gray-700 border border-gray-300 hover:bg-gray-50",
        danger: "bg-red-600 text-white hover:bg-red-700",
        ghost: "text-gray-600 hover:text-gray-900 hover:bg-gray-100",
    };
    const sizes = {
        sm: "px-2.5 py-1.5 text-xs",
        md: "px-3 py-2 text-sm",
        lg: "px-4 py-2.5 text-base",
    };
    return (
        <button
            onClick={onClick}
            disabled={disabled}
            className={`${base} ${variants[variant]} ${sizes[size]} ${disabled ? "opacity-50 cursor-not-allowed" : ""}`}
        >
            {children}
        </button>
    );
}

// ============================================================================
// Site Dashboard
// ============================================================================

function SiteDashboard({ sites, onSelectSite, reviewCount }) {
    return (
        <div>
            <div className="flex items-center justify-between mb-6">
                <div>
                    <h2 className="text-lg font-semibold text-gray-900">Sites</h2>
                    <p className="text-sm text-gray-500 mt-0.5">{sites.length} sites managed</p>
                </div>
                {reviewCount > 0 && (
                    <Badge variant="review">{reviewCount} items need review</Badge>
                )}
            </div>

            <div className="space-y-3">
                {sites.map((site) => {
                    const wi = site.work_items || {};
                    const total = Object.values(wi).reduce((a, b) => a + b, 0);
                    const pct = total > 0 ? Math.round((wi.done / total) * 100) : 0;

                    return (
                        <Card key={site.id} onClick={() => onSelectSite(site)}>
                            <div className="flex items-start justify-between">
                                <div className="min-w-0">
                                    <div className="flex items-center gap-2">
                                        <h3 className="font-medium text-gray-900 truncate">{site.domain}</h3>
                                        <StatusBadge status={site.status} />
                                    </div>
                                    <p className="text-sm text-gray-500 mt-0.5">{site.company_name || "—"}</p>
                                </div>
                                <div className="text-right flex-shrink-0 ml-4">
                                    <div className="text-sm font-medium text-gray-900">{pct}%</div>
                                    <div className="text-xs text-gray-500">{wi.done}/{total} items</div>
                                </div>
                            </div>

                            <div className="mt-3 flex gap-4 text-xs text-gray-500">
                                {wi.active > 0 && <span className="text-amber-600">{wi.active} active</span>}
                                {wi.ready > 0 && <span className="text-blue-600">{wi.ready} ready</span>}
                                {wi.review > 0 && <span className="text-purple-600 font-medium">{wi.review} review</span>}
                                {wi.failed > 0 && <span className="text-red-600">{wi.failed} failed</span>}
                                {wi.blocked > 0 && <span className="text-red-600">{wi.blocked} blocked</span>}
                            </div>

                            {total > 0 && (
                                <div className="mt-2 h-1.5 bg-gray-100 rounded-full overflow-hidden">
                                    <div
                                        className="h-full bg-emerald-500 rounded-full transition-all"
                                        style={{ width: `${pct}%` }}
                                    />
                                </div>
                            )}
                        </Card>
                    );
                })}
            </div>
        </div>
    );
}

// ============================================================================
// Review Queue
// ============================================================================

function ReviewQueue({ items, onSelectItem, onRetry, onResolve }) {
    if (items.length === 0) {
        return (
            <div className="text-center py-12 text-gray-500">
                <div className="text-3xl mb-2">✓</div>
                <p className="font-medium">No items need review</p>
                <p className="text-sm mt-1">All content has passed validation</p>
            </div>
        );
    }

    return (
        <div>
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
                Items Needing Review ({items.length})
            </h2>
            <div className="space-y-3">
                {items.map((item) => (
                    <Card key={item.id}>
                        <div className="flex items-start justify-between">
                            <div className="min-w-0">
                                <div className="flex items-center gap-2 mb-1">
                                    <Badge variant="review">{item.item_type}</Badge>
                                    <span className="text-xs text-gray-400">{item.domain}</span>
                                </div>
                                <p className="text-sm font-medium text-gray-900">{item.summary}</p>
                                {item.spec?.page_name && (
                                    <p className="text-xs text-gray-500 mt-1">
                                        Page: {item.spec.page_name}
                                        {item.spec.section && ` → ${item.spec.section}`}
                                    </p>
                                )}
                                {item.spec?.missing_data && (
                                    <p className="text-xs text-amber-600 mt-1">
                                        Missing: {item.spec.missing_data}
                                    </p>
                                )}
                            </div>
                            <div className="flex gap-2 flex-shrink-0 ml-4">
                                <Button size="sm" variant="secondary" onClick={() => onSelectItem(item)}>
                                    Review
                                </Button>
                                <Button size="sm" variant="ghost" onClick={() => onRetry(item.id)}>
                                    Retry
                                </Button>
                                <Button size="sm" variant="ghost" onClick={() => onResolve(item.id)}>
                                    Dismiss
                                </Button>
                            </div>
                        </div>
                    </Card>
                ))}
            </div>
        </div>
    );
}

// ============================================================================
// Review Detail Panel
// ============================================================================

function ReviewDetail({ item, siteSpecs, onUpdateSpec, onRetry, onResolve, onBack }) {
    const [editData, setEditData] = useState("");
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        if (siteSpecs?.identity?.data) {
            setEditData(JSON.stringify(siteSpecs.identity.data, null, 2));
        }
    }, [siteSpecs]);

    const handleSaveAndRetry = async () => {
        setSaving(true);
        try {
            const parsed = JSON.parse(editData);
            await onUpdateSpec(item.site_id, "identity", parsed);
            await onRetry(item.id);
        } catch (err) {
            alert("Invalid JSON or save failed: " + err.message);
        }
        setSaving(false);
    };

    return (
        <div>
            <button onClick={onBack} className="text-sm text-gray-500 hover:text-gray-900 mb-4 flex items-center gap-1">
                ← Back to queue
            </button>

            <Card className="mb-4">
                <div className="flex items-center gap-2 mb-2">
                    <Badge variant="review">{item.item_type}</Badge>
                    <span className="text-sm text-gray-500">{item.domain}</span>
                </div>
                <h3 className="font-medium text-gray-900 mb-1">{item.summary}</h3>
                {item.spec?.missing_data && (
                    <p className="text-sm text-amber-600">Missing: {item.spec.missing_data}</p>
                )}
                {item.spec?.fix_guidance && (
                    <p className="text-sm text-gray-500 mt-1">{item.spec.fix_guidance}</p>
                )}
            </Card>

            <Card className="mb-4">
                <h4 className="text-sm font-medium text-gray-700 mb-2">Work Item Spec</h4>
                <pre className="text-xs bg-gray-50 p-3 rounded border overflow-auto max-h-40">
          {JSON.stringify(item.spec, null, 2)}
        </pre>
            </Card>

            <Card className="mb-4">
                <h4 className="text-sm font-medium text-gray-700 mb-2">
                    Site Identity Spec
                    <span className="text-xs text-gray-400 ml-2">(edit to provide missing data)</span>
                </h4>
                <textarea
                    value={editData}
                    onChange={(e) => setEditData(e.target.value)}
                    className="w-full h-64 text-xs font-mono bg-gray-50 p-3 rounded border resize-y focus:outline-none focus:ring-2 focus:ring-blue-200"
                    spellCheck="false"
                />
            </Card>

            <div className="flex gap-3">
                <Button onClick={handleSaveAndRetry} disabled={saving}>
                    {saving ? "Saving..." : "Save Spec & Retry"}
                </Button>
                <Button variant="secondary" onClick={() => onResolve(item.id)}>
                    Dismiss (keep hidden)
                </Button>
                <Button variant="danger" onClick={() => onRetry(item.id)}>
                    Retry Without Changes
                </Button>
            </div>
        </div>
    );
}

// ============================================================================
// Main App
// ============================================================================

export default function SiteAdmin() {
    const [view, setView] = useState("dashboard"); // dashboard | review | detail
    const [sites, setSites] = useState([]);
    const [reviewItems, setReviewItems] = useState([]);
    const [selectedItem, setSelectedItem] = useState(null);
    const [selectedSite, setSelectedSite] = useState(null);
    const [siteSpecs, setSiteSpecs] = useState(null);
    const [loading, setLoading] = useState(true);

    const loadData = useCallback(async () => {
        setLoading(true);
        try {
            if (useMock) {
                setSites(MOCK.sites);
                setReviewItems(MOCK.reviewItems);
            } else {
                const [sitesRes, itemsRes] = await Promise.all([
                    api("/sites"),
                    api("/work-items?status=needs_human_review"),
                ]);
                setSites(sitesRes.sites || []);
                setReviewItems(itemsRes.items || []);
            }
        } catch (err) {
            console.error("Failed to load data:", err);
        }
        setLoading(false);
    }, []);

    useEffect(() => { loadData(); }, [loadData]);

    const loadSiteDetail = async (siteId) => {
        if (useMock) {
            setSiteSpecs({ identity: { data: { company_name: "FineTuning", email: "finetuning@contactforsales.com", phone: "+44 (0) 7934 524 911", leadership_team: [] } } });
            return;
        }
        const data = await api(`/sites/${siteId}`);
        setSiteSpecs(data.specs);
    };

    const handleSelectReviewItem = async (item) => {
        setSelectedItem(item);
        await loadSiteDetail(item.site_id);
        setView("detail");
    };

    const handleUpdateSpec = async (siteId, aspect, data) => {
        if (useMock) { console.log("Would update spec:", { siteId, aspect, data }); return; }
        await api(`/sites/${siteId}/specs/${aspect}`, {
            method: "PATCH",
            body: JSON.stringify({ data }),
        });
    };

    const handleRetry = async (itemId) => {
        if (useMock) { console.log("Would retry:", itemId); }
        else { await api(`/work-items/${itemId}/retry`, { method: "POST" }); }
        setReviewItems((prev) => prev.filter((i) => i.id !== itemId));
        setView("review");
    };

    const handleResolve = async (itemId) => {
        if (useMock) { console.log("Would resolve:", itemId); }
        else {
            await api(`/work-items/${itemId}/resolve`, {
                method: "POST",
                body: JSON.stringify({ resolution: "Dismissed via admin UI" }),
            });
        }
        setReviewItems((prev) => prev.filter((i) => i.id !== itemId));
        if (view === "detail") setView("review");
    };

    const reviewCount = reviewItems.length;

    return (
        <div className="min-h-screen bg-gray-50">
            <header className="bg-white border-b border-gray-200 px-6 py-3">
                <div className="max-w-5xl mx-auto flex items-center justify-between">
                    <div className="flex items-center gap-6">
                        <h1 className="text-base font-semibold text-gray-900">Site Admin</h1>
                        <nav className="flex gap-1">
                            <button
                                onClick={() => setView("dashboard")}
                                className={`px-3 py-1.5 text-sm rounded ${view === "dashboard" ? "bg-gray-100 text-gray-900 font-medium" : "text-gray-500 hover:text-gray-700"}`}
                            >
                                Dashboard
                            </button>
                            <button
                                onClick={() => setView("review")}
                                className={`px-3 py-1.5 text-sm rounded flex items-center gap-1.5 ${view === "review" || view === "detail" ? "bg-gray-100 text-gray-900 font-medium" : "text-gray-500 hover:text-gray-700"}`}
                            >
                                Review Queue
                                {reviewCount > 0 && (
                                    <span className="bg-purple-100 text-purple-700 text-xs px-1.5 py-0.5 rounded-full font-medium">
                    {reviewCount}
                  </span>
                                )}
                            </button>
                        </nav>
                    </div>
                    <Button size="sm" variant="ghost" onClick={loadData}>
                        Refresh
                    </Button>
                </div>
            </header>

            <main className="max-w-5xl mx-auto px-6 py-6">
                {loading ? (
                    <div className="text-center py-12 text-gray-400">Loading...</div>
                ) : view === "dashboard" ? (
                    <SiteDashboard
                        sites={sites}
                        onSelectSite={(site) => { setSelectedSite(site); }}
                        reviewCount={reviewCount}
                    />
                ) : view === "review" ? (
                    <ReviewQueue
                        items={reviewItems}
                        onSelectItem={handleSelectReviewItem}
                        onRetry={handleRetry}
                        onResolve={handleResolve}
                    />
                ) : view === "detail" && selectedItem ? (
                    <ReviewDetail
                        item={selectedItem}
                        siteSpecs={siteSpecs}
                        onUpdateSpec={handleUpdateSpec}
                        onRetry={handleRetry}
                        onResolve={handleResolve}
                        onBack={() => setView("review")}
                    />
                ) : null}
            </main>
        </div>
    );
}