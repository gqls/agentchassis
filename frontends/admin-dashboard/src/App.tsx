import { useState, useEffect, useCallback } from "react";

const API_BASE = "/api/v1/admin";

// ── API helpers ──────────────────────────────────────────────────────────────
async function apiFetch(path, token, opts = {}) {
    const res = await fetch(`${API_BASE}${path}`, {
        ...opts,
        headers: {
            "Authorization": `Bearer ${token}`,
            "Content-Type": "application/json",
            ...(opts.headers || {}),
        },
    });
    if (res.status === 401) throw new Error("UNAUTHORIZED");
    if (!res.ok) {
        const body = await res.text();
        throw new Error(body || res.statusText);
    }
    return res.json();
}

// ── Status badge colours ─────────────────────────────────────────────────────
const STATUS_COLORS = {
    triaged: { bg: "#dbeafe", text: "#1e40af", label: "Triaged" },
    claimed: { bg: "#fef3c7", text: "#92400e", label: "Claimed" },
    needs_human_review: { bg: "#fce7f3", text: "#9d174d", label: "Needs Review" },
    failed: { bg: "#fee2e2", text: "#991b1b", label: "Failed" },
    blocked: { bg: "#f3e8ff", text: "#6b21a8", label: "Blocked" },
    complete: { bg: "#d1fae5", text: "#065f46", label: "Complete" },
    detected: { bg: "#e0e7ff", text: "#3730a3", label: "Detected" },
};

const SEVERITY_COLORS = {
    high: { bg: "#fee2e2", text: "#991b1b" },
    medium: { bg: "#fef3c7", text: "#92400e" },
    low: { bg: "#e0e7ff", text: "#3730a3" },
};

function Badge({ status, type = "status" }) {
    const map = type === "severity" ? SEVERITY_COLORS : STATUS_COLORS;
    const c = map[status] || { bg: "#f3f4f6", text: "#374151" };
    return (
        <span style={{
            display: "inline-block", padding: "2px 10px", borderRadius: 9999,
            fontSize: 12, fontWeight: 600, letterSpacing: 0.3,
            backgroundColor: c.bg, color: c.text, whiteSpace: "nowrap",
        }}>
      {c.label || status}
    </span>
    );
}

// ── Editable Review Form ─────────────────────────────────────────────────────
// Renders review_data fields as editable inputs for checkpoint items.
// Handles nested objects (like services arrays) with a JSON textarea fallback.

function EditableReviewForm({ reviewData, onChange }) {
    if (!reviewData || typeof reviewData !== "object") {
        return <JSONEditor value={reviewData} onChange={onChange} />;
    }

    const handleFieldChange = (key, value) => {
        onChange({ ...reviewData, [key]: value });
    };

    return (
        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
            {Object.entries(reviewData).map(([key, value]) => {
                const label = key.replace(/_/g, " ").replace(/\b\w/g, c => c.toUpperCase());

                // Arrays of objects (e.g. team_members, services) → structured editor
                if (Array.isArray(value) && value.length > 0 && typeof value[0] === "object") {
                    return (
                        <div key={key}>
                            <label style={formLabelStyle}>{label}</label>
                            {value.map((entry, idx) => (
                                <div key={idx} style={{
                                    background: "#fff", border: "1px solid #e2e8f0", borderRadius: 6,
                                    padding: 12, marginBottom: 8,
                                }}>
                                    <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
                                        <span style={{ fontSize: 11, fontWeight: 600, color: "#94a3b8" }}>#{idx + 1}</span>
                                        {value.length > 1 && (
                                            <button onClick={() => {
                                                const updated = [...value];
                                                updated.splice(idx, 1);
                                                handleFieldChange(key, updated);
                                            }} style={{ fontSize: 11, color: "#991b1b", background: "none", border: "none", cursor: "pointer" }}>
                                                Remove
                                            </button>
                                        )}
                                    </div>
                                    {Object.keys(entry).map(subKey => (
                                        <div key={subKey} style={{ marginBottom: 6 }}>
                                            <label style={{ ...formLabelStyle, fontSize: 11 }}>
                                                {subKey.replace(/_/g, " ").replace(/\b\w/g, c => c.toUpperCase())}
                                            </label>
                                            {String(entry[subKey] || "").length > 80 ? (
                                                <textarea
                                                    value={entry[subKey] || ""}
                                                    onChange={e => {
                                                        const updated = [...value];
                                                        updated[idx] = { ...updated[idx], [subKey]: e.target.value };
                                                        handleFieldChange(key, updated);
                                                    }}
                                                    style={{ ...textareaStyle, minHeight: 50 }}
                                                />
                                            ) : (
                                                <input
                                                    type="text"
                                                    value={entry[subKey] || ""}
                                                    onChange={e => {
                                                        const updated = [...value];
                                                        updated[idx] = { ...updated[idx], [subKey]: e.target.value };
                                                        handleFieldChange(key, updated);
                                                    }}
                                                    style={formInputStyle}
                                                />
                                            )}
                                        </div>
                                    ))}
                                </div>
                            ))}
                            <button onClick={() => {
                                // Add a new entry with same shape as existing entries
                                const template = {};
                                Object.keys(value[0]).forEach(k => template[k] = "");
                                handleFieldChange(key, [...value, template]);
                            }} style={{
                                fontSize: 12, color: "#1e40af", background: "none",
                                border: "1px dashed #93c5fd", borderRadius: 6,
                                padding: "6px 12px", cursor: "pointer", width: "100%",
                            }}>
                                + Add {label.replace(/s$/, "")}
                            </button>
                        </div>
                    );
                }

                // Simple arrays → comma-separated
                if (Array.isArray(value)) {
                    return (
                        <div key={key}>
                            <label style={formLabelStyle}>{label}</label>
                            <input
                                type="text"
                                value={value.join(", ")}
                                onChange={e => handleFieldChange(key, e.target.value.split(",").map(s => s.trim()).filter(Boolean))}
                                style={formInputStyle}
                            />
                            <div style={{ fontSize: 11, color: "#94a3b8", marginTop: 2 }}>Comma-separated</div>
                        </div>
                    );
                }

                // Nested objects → JSON editor
                if (typeof value === "object" && value !== null) {
                    return (
                        <div key={key}>
                            <label style={formLabelStyle}>{label}</label>
                            <textarea
                                value={JSON.stringify(value, null, 2)}
                                onChange={e => {
                                    try {
                                        handleFieldChange(key, JSON.parse(e.target.value));
                                    } catch { /* let them keep typing */ }
                                }}
                                style={{ ...textareaStyle, minHeight: 80 }}
                            />
                        </div>
                    );
                }

                // Long strings → textarea
                if (typeof value === "string" && value.length > 100) {
                    return (
                        <div key={key}>
                            <label style={formLabelStyle}>{label}</label>
                            <textarea
                                value={value}
                                onChange={e => handleFieldChange(key, e.target.value)}
                                style={textareaStyle}
                            />
                        </div>
                    );
                }

                // Booleans → checkbox
                if (typeof value === "boolean") {
                    return (
                        <div key={key} style={{ display: "flex", alignItems: "center", gap: 8 }}>
                            <input
                                type="checkbox"
                                checked={value}
                                onChange={e => handleFieldChange(key, e.target.checked)}
                            />
                            <label style={{ ...formLabelStyle, margin: 0 }}>{label}</label>
                        </div>
                    );
                }

                // Everything else → text input
                return (
                    <div key={key}>
                        <label style={formLabelStyle}>{label}</label>
                        <input
                            type="text"
                            value={String(value ?? "")}
                            onChange={e => handleFieldChange(key, e.target.value)}
                            style={formInputStyle}
                        />
                    </div>
                );
            })}
        </div>
    );
}

// Fallback JSON editor for unstructured data
function JSONEditor({ value, onChange }) {
    const [text, setText] = useState(JSON.stringify(value, null, 2));
    const [parseError, setParseError] = useState("");

    return (
        <div>
            <textarea
                value={text}
                onChange={e => {
                    setText(e.target.value);
                    try {
                        const parsed = JSON.parse(e.target.value);
                        onChange(parsed);
                        setParseError("");
                    } catch (err) {
                        setParseError("Invalid JSON");
                    }
                }}
                style={{ ...textareaStyle, minHeight: 200, fontFamily: "monospace", fontSize: 12 }}
            />
            {parseError && <div style={{ fontSize: 11, color: "#991b1b", marginTop: 4 }}>{parseError}</div>}
        </div>
    );
}

// ── Login ────────────────────────────────────────────────────────────────────
function LoginScreen({ onLogin }) {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e) => {
        e.preventDefault();
        setLoading(true);
        setError("");
        try {
            const res = await fetch("/api/v1/auth/login", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ email, password }),
            });
            if (!res.ok) throw new Error("Invalid credentials");
            const data = await res.json();
            if (data.user?.role !== "admin") throw new Error("Admin access required");
            onLogin(data.access_token, data.user);
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{
            minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center",
            background: "linear-gradient(135deg, #0f172a 0%, #1e293b 100%)",
            fontFamily: "'IBM Plex Sans', -apple-system, sans-serif",
        }}>
            <form onSubmit={handleSubmit} style={{
                background: "#fff", borderRadius: 12, padding: "40px 36px", width: 380,
                boxShadow: "0 25px 50px rgba(0,0,0,0.25)",
            }}>
                <div style={{ fontSize: 13, fontWeight: 600, color: "#64748b", letterSpacing: 1.5, marginBottom: 4 }}>PERSONAE</div>
                <h2 style={{ margin: "0 0 24px", fontSize: 22, color: "#0f172a" }}>Admin Dashboard</h2>
                {error && <div style={{ background: "#fee2e2", color: "#991b1b", padding: "8px 12px", borderRadius: 6, fontSize: 13, marginBottom: 16 }}>{error}</div>}
                <label style={labelStyle}>Email</label>
                <input type="email" value={email} onChange={e => setEmail(e.target.value)} style={inputStyle} required />
                <label style={labelStyle}>Password</label>
                <input type="password" value={password} onChange={e => setPassword(e.target.value)} style={inputStyle} required />
                <button type="submit" disabled={loading} style={{
                    ...btnPrimary, width: "100%", marginTop: 8, opacity: loading ? 0.7 : 1,
                }}>
                    {loading ? "Signing in…" : "Sign In"}
                </button>
            </form>
        </div>
    );
}

// ── Sites Overview ───────────────────────────────────────────────────────────
function SitesOverview({ sites, token, onSelectSite, onSelectPages, onSelectSpecs, onRefresh }) {
    const [actionLoading, setActionLoading] = useState(false);

    const handleToggleSiteLock = async (site) => {
        const action = site.locked ? "unlock" : "lock";
        if (site.locked ? false : !confirm(`Lock ${site.domain}? All automated agent activity will stop for this site.`)) return;
        setActionLoading(true);
        try {
            await apiFetch(`/sites/${site.id}/${action}`, token, { method: "POST" });
            onRefresh();
        } catch (err) { console.error(err); }
        finally { setActionLoading(false); }
    };

    return (
        <div>
            <h2 style={sectionTitle}>Sites</h2>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(340px, 1fr))", gap: 16 }}>
                {sites.map(site => (
                    <div key={site.id} style={{
                        background: "#fff", borderRadius: 10, padding: "20px 22px",
                        border: `1px solid ${site.locked ? "#c4b5fd" : "#e2e8f0"}`,
                        transition: "box-shadow 0.15s",
                        opacity: site.locked ? 0.85 : 1,
                    }}
                         onMouseEnter={e => e.currentTarget.style.boxShadow = "0 4px 12px rgba(0,0,0,0.08)"}
                         onMouseLeave={e => e.currentTarget.style.boxShadow = "none"}
                    >
                        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "start" }}>
                            <div>
                                <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                                    <span style={{ fontSize: 16, fontWeight: 600, color: "#0f172a" }}>{site.domain}</span>
                                    {site.locked && (
                                        <span style={{ fontSize: 11, color: "#7c3aed", background: "#ede9fe", padding: "1px 8px", borderRadius: 4 }}>
                                            🔒 Locked
                                        </span>
                                    )}
                                </div>
                                <div style={{ fontSize: 13, color: "#64748b", marginTop: 2 }}>{site.company_name}</div>
                            </div>
                            <Badge status={site.status} />
                        </div>
                        <div style={{ display: "flex", gap: 12, marginTop: 16, flexWrap: "wrap" }}>
                            {[
                                { k: "review", label: "Review", color: "#9d174d" },
                                { k: "failed", label: "Failed", color: "#991b1b" },
                                { k: "active", label: "Active", color: "#92400e" },
                                { k: "ready", label: "Ready", color: "#1e40af" },
                                { k: "done", label: "Done", color: "#065f46" },
                            ].map(({ k, label, color }) => (
                                <div key={k} style={{ textAlign: "center" }}>
                                    <div style={{ fontSize: 18, fontWeight: 700, color }}>{site.work_items?.[k] || 0}</div>
                                    <div style={{ fontSize: 11, color: "#94a3b8" }}>{label}</div>
                                </div>
                            ))}
                        </div>
                        <div style={{ display: "flex", gap: 8, marginTop: 14 }}>
                            <button onClick={() => onSelectSite(site)} style={{
                                ...btnSecondary, fontSize: 12, padding: "6px 14px", flex: 1,
                            }}>Work Items</button>
                            <button onClick={() => onSelectPages(site)} style={{
                                ...btnSecondary, fontSize: 12, padding: "6px 14px", flex: 1,
                            }}>Pages</button>
                            <button onClick={() => onSelectSpecs(site)} style={{
                                ...btnSecondary, fontSize: 12, padding: "6px 14px", flex: 1,
                            }}>Direction</button>
                        </div>
                        <div style={{ display: "flex", gap: 8, marginTop: 8 }}>
                            <button onClick={() => handleToggleSiteLock(site)} disabled={actionLoading} style={{
                                ...btnSecondary, fontSize: 11, padding: "4px 12px",
                                color: site.locked ? "#7c3aed" : "#64748b",
                                borderColor: site.locked ? "#c4b5fd" : "#e2e8f0",
                            }}>
                                {site.locked ? "🔒 Unlock Site" : "Lock Site"}
                            </button>
                        </div>
                        {site.last_deployed && (
                            <div style={{ fontSize: 11, color: "#94a3b8", marginTop: 8 }}>
                                Last deployed: {new Date(site.last_deployed).toLocaleDateString()}
                            </div>
                        )}
                    </div>
                ))}
            </div>
        </div>
    );
}

// ── Work Items List ──────────────────────────────────────────────────────────
function WorkItemsList({ token, siteFilter, onBack }) {
    const [allItems, setAllItems] = useState([]);
    const [loading, setLoading] = useState(true);
    const [statusFilter, setStatusFilter] = useState("");
    const [typeFilter, setTypeFilter] = useState("");
    const [selectedItem, setSelectedItem] = useState(null);
    const [actionLoading, setActionLoading] = useState(false);
    const [message, setMessage] = useState("");

    // Editable review data for checkpoint items
    const [editedReviewData, setEditedReviewData] = useState(null);
    const [approveNotes, setApproveNotes] = useState("");

    // Site edit
    const [showSiteEdit, setShowSiteEdit] = useState(false);
    const [siteEditData, setSiteEditData] = useState({});

    const handleSaveSiteDetails = async () => {
        if (!siteFilter?.id) return;
        setActionLoading(true);
        try {
            // Only send non-empty fields
            const updates = {};
            for (const [k, v] of Object.entries(siteEditData)) {
                if (v !== "" && v !== null && v !== undefined) updates[k] = v;
            }
            if (Object.keys(updates).length === 0) {
                setMessage("No changes to save");
                return;
            }
            await apiFetch(`/sites/${siteFilter.id}`, token, {
                method: "PATCH",
                body: JSON.stringify(updates),
            });
            setMessage("Site details updated");
            setShowSiteEdit(false);
        } catch (err) { setMessage("Save failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const loadItems = useCallback(async () => {
        setLoading(true);
        try {
            // Load all non-complete items, filter client-side for accurate counts
            let path = `/work-items?domain=build`;
            if (siteFilter?.id) path += `&site_id=${siteFilter.id}`;
            const data = await apiFetch(path, token);
            setAllItems(data.items || []);
        } catch (err) {
            if (err.message === "UNAUTHORIZED") return;
            console.error(err);
        } finally {
            setLoading(false);
        }
    }, [token, siteFilter]);

    useEffect(() => { loadItems(); }, [loadItems]);

    // Client-side filtering
    const items = allItems.filter(item => {
        if (statusFilter && item.status !== statusFilter) return false;
        if (typeFilter && item.item_type !== typeFilter) return false;
        return true;
    });

    // When selecting a review item, initialise the editable data
    const selectItem = (item) => {
        setSelectedItem(item);
        setApproveNotes("");
        if (item?.status === "needs_human_review" && item?.spec) {
            if (item.spec.checkpoint && item.spec.review_data) {
                // Checkpoint items: edit the review_data
                setEditedReviewData(JSON.parse(JSON.stringify(item.spec.review_data)));
            } else if (item.item_type === "placeholder_content") {
                // Placeholder content: build input form based on page type and missing_data
                setEditedReviewData(buildPlaceholderForm(item));
            } else if (!item.spec.checkpoint) {
                // Other review items: edit the spec directly (strip internal fields)
                const editableSpec = { ...item.spec };
                delete editableSpec.check;
                delete editableSpec.original_domain;
                setEditedReviewData(JSON.parse(JSON.stringify(editableSpec)));
            } else {
                setEditedReviewData(null);
            }
        } else {
            setEditedReviewData(null);
        }
    };

    // Build an appropriate input form for placeholder_content items
    // based on the page name and missing_data hint
    function buildPlaceholderForm(item) {
        const pageName = item.spec?.page_name || "";
        const missingData = (item.spec?.missing_data || "").toLowerCase();

        // Contact page
        if (pageName === "contact" || missingData.includes("contact") || missingData.includes("email")) {
            return {
                email: "",
                phone: "",
                contact_address: "",
                opening_hours: "",
                contact_form_note: "",
            };
        }

        // About page
        if (pageName === "about" || missingData.includes("team") || missingData.includes("bio")) {
            return {
                company_description: "",
                team_members: [
                    { name: "", title: "", bio: "" },
                ],
                company_values: "",
                founded_year: "",
            };
        }

        // Services page
        if (pageName === "services" || missingData.includes("service")) {
            return {
                services: [
                    { name: "", description: "", features: "" },
                ],
            };
        }

        // Pricing page
        if (pageName === "pricing" || missingData.includes("pricing") || missingData.includes("price")) {
            return {
                pricing_intro: "",
                plans: [
                    { name: "", price: "", features: "", cta: "" },
                ],
            };
        }

        // Generic fallback
        return {
            content: "",
            notes: "",
        };
    }

    const handleRetry = async (id) => {
        setActionLoading(true);
        try {
            await apiFetch(`/work-items/${id}/retry`, token, { method: "POST" });
            setMessage("Item queued for retry");
            selectItem(null);
            loadItems();
        } catch (err) { setMessage("Retry failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const handleResolve = async (id, resolution) => {
        setActionLoading(true);
        try {
            await apiFetch(`/work-items/${id}/resolve`, token, {
                method: "POST",
                body: JSON.stringify({ resolution: resolution || "Resolved via admin dashboard" }),
            });
            setMessage("Item resolved");
            selectItem(null);
            loadItems();
        } catch (err) { setMessage("Resolve failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const handleUpdateStatus = async (id, newStatus) => {
        setActionLoading(true);
        try {
            await apiFetch(`/work-items/${id}`, token, {
                method: "PATCH",
                body: JSON.stringify({ status: newStatus }),
            });
            setMessage(`Status updated to ${newStatus}`);
            selectItem(null);
            loadItems();
        } catch (err) { setMessage("Update failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const handleApprove = async (id) => {
        if (!editedReviewData) {
            setMessage("No review data to approve");
            return;
        }
        setActionLoading(true);
        try {
            const result = await apiFetch(`/work-items/${id}/approve`, token, {
                method: "POST",
                body: JSON.stringify({
                    review_data: editedReviewData,
                    notes: approveNotes || undefined,
                }),
            });
            let msg = "Approved";
            if (result.spec_updated) msg += ` — spec '${result.spec_updated}' updated`;
            if (result.follow_on_item_id) msg += ` — follow-on item created`;
            setMessage(msg);
            selectItem(null);
            loadItems();
        } catch (err) { setMessage("Approve failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    // Save edited data and create a rebuild work item for non-checkpoint review items
    const handleSaveAndRebuild = async (item) => {
        if (!editedReviewData) {
            setMessage("No data to save");
            return;
        }

        // Check that at least one field has content
        const hasContent = Object.values(editedReviewData).some(v => {
            if (Array.isArray(v)) return v.some(entry =>
                typeof entry === "object" ? Object.values(entry).some(fv => fv !== "") : entry !== ""
            );
            return v !== "" && v !== null && v !== undefined;
        });

        if (!hasContent) {
            setMessage("Please fill in at least one field before saving");
            return;
        }

        setActionLoading(true);
        try {
            const pageName = item.spec?.page_name || "unknown";
            const pageId = item.spec?.page_id;

            // Step 1: Update site-level fields if present (email, phone, etc.)
            const siteFields = {};
            const siteFieldNames = ["email", "phone", "contact_address", "company_name", "tagline", "logo_text"];
            for (const f of siteFieldNames) {
                if (editedReviewData[f] !== undefined && editedReviewData[f] !== "") {
                    siteFields[f] = editedReviewData[f];
                }
            }
            if (Object.keys(siteFields).length > 0) {
                await apiFetch(`/sites/${item.site_id}`, token, {
                    method: "PATCH",
                    body: JSON.stringify(siteFields),
                });
            }

            // Step 2: Save page-specific data to site_specs
            // This makes it available to the page-build-handler when it rebuilds
            const specAspect = `page_content_${pageName}`;
            await apiFetch(`/sites/${item.site_id}/specs/${specAspect}`, token, {
                method: "PATCH",
                body: JSON.stringify({
                    data: {
                        page_name: pageName,
                        page_id: pageId,
                        provided_by: "admin",
                        content: editedReviewData,
                    },
                }),
            });

            // Step 3: Create a rebuild work item for the affected page
            await apiFetch(`/work-items`, token, {
                method: "POST",
                body: JSON.stringify({
                    site_id: item.site_id,
                    item_type: "content_rewrite",
                    summary: `Rebuild ${pageName} page with provided data`,
                    severity: "high",
                    handler_agent: "page-build-handler",
                    page_id: pageId || undefined,
                    priority: 10,
                    spec: {
                        page_name: pageName,
                        content_source: specAspect,
                        reason: "placeholder_content_replaced",
                    },
                }),
            });

            // Step 4: Resolve the review item
            await apiFetch(`/work-items/${item.id}/resolve`, token, {
                method: "POST",
                body: JSON.stringify({ resolution: `Real data provided for ${pageName}, rebuild queued` }),
            });

            setMessage(`Data saved for ${pageName} — rebuild queued`);
            selectItem(null);
            loadItems();
        } catch (err) { setMessage("Save failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    // Unique item types from loaded items
    const itemTypes = [...new Set(allItems.map(i => i.item_type))].sort();
    const isCheckpoint = selectedItem?.spec?.checkpoint === true;
    const isEditable = selectedItem?.status === "needs_human_review" && editedReviewData != null;

    // Status counts for filter badges (from all items, not filtered)
    const statusCounts = allItems.reduce((acc, item) => {
        acc[item.status] = (acc[item.status] || 0) + 1;
        return acc;
    }, {} as Record<string, number>);
    const allItemsCount = allItems.length;
    const failedCount = statusCounts["failed"] || 0;

    const handleRetryAllFailed = async () => {
        if (!confirm(`Retry all ${failedCount} failed items?`)) return;
        setActionLoading(true);
        try {
            const failedItems = allItems.filter(i => i.status === "failed");
            let retried = 0;
            for (const item of failedItems) {
                try {
                    await apiFetch(`/work-items/${item.id}/retry`, token, { method: "POST" });
                    retried++;
                } catch { /* continue with others */ }
            }
            setMessage(`Retried ${retried} of ${failedItems.length} failed items`);
            loadItems();
        } catch (err) { setMessage("Bulk retry failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    return (
        <div>
            <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 20 }}>
                <button onClick={onBack} style={btnSecondary}>← Sites</button>
                <h2 style={{ ...sectionTitle, margin: 0, flex: 1 }}>
                    Work Items {siteFilter && <span style={{ fontWeight: 400, color: "#64748b" }}>— {siteFilter.domain}</span>}
                </h2>
                {siteFilter && (
                    <button onClick={() => {
                        setShowSiteEdit(!showSiteEdit);
                        if (!showSiteEdit) {
                            setSiteEditData({
                                company_name: siteFilter.company_name || "",
                                email: siteFilter.email || "",
                                phone: siteFilter.phone || "",
                                contact_address: siteFilter.contact_address || "",
                                tagline: siteFilter.tagline || "",
                            });
                        }
                    }} style={{ ...btnSecondary, fontSize: 12 }}>
                        {showSiteEdit ? "Close" : "Edit Site"}
                    </button>
                )}
            </div>

            {/* Site edit panel */}
            {showSiteEdit && siteFilter && (
                <div style={{
                    background: "#fff", border: "1px solid #e2e8f0", borderRadius: 8,
                    padding: 16, marginBottom: 16,
                }}>
                    <div style={{ fontSize: 13, fontWeight: 600, color: "#475569", marginBottom: 12 }}>Site Details — {siteFilter.domain}</div>
                    <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12 }}>
                        {[
                            { key: "company_name", label: "Company Name" },
                            { key: "tagline", label: "Tagline" },
                            { key: "email", label: "Email" },
                            { key: "phone", label: "Phone" },
                            { key: "contact_address", label: "Address" },
                        ].map(({ key, label }) => (
                            <div key={key}>
                                <label style={formLabelStyle}>{label}</label>
                                <input
                                    type="text"
                                    value={siteEditData[key] || ""}
                                    onChange={e => setSiteEditData({ ...siteEditData, [key]: e.target.value })}
                                    style={formInputStyle}
                                    placeholder={`Enter ${label.toLowerCase()}…`}
                                />
                            </div>
                        ))}
                    </div>
                    <div style={{ marginTop: 12, display: "flex", gap: 8 }}>
                        <button onClick={handleSaveSiteDetails} disabled={actionLoading} style={btnPrimary}>
                            Save Site Details
                        </button>
                        <button onClick={() => setShowSiteEdit(false)} style={btnSecondary}>Cancel</button>
                    </div>
                </div>
            )}

            {message && (
                <div style={{
                    background: "#d1fae5", color: "#065f46", padding: "8px 14px", borderRadius: 6,
                    fontSize: 13, marginBottom: 12, display: "flex", justifyContent: "space-between",
                }}>
                    {message}
                    <span onClick={() => setMessage("")} style={{ cursor: "pointer" }}>✕</span>
                </div>
            )}

            {/* Filters */}
            <div style={{ display: "flex", gap: 10, marginBottom: 16, flexWrap: "wrap", alignItems: "center" }}>
                <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)} style={selectStyle}>
                    <option value="">All ({allItemsCount})</option>
                    <option value="needs_human_review">Needs Review ({statusCounts["needs_human_review"] || 0})</option>
                    <option value="triaged">Triaged ({statusCounts["triaged"] || 0})</option>
                    <option value="claimed">Claimed ({statusCounts["claimed"] || 0})</option>
                    <option value="failed">Failed ({statusCounts["failed"] || 0})</option>
                    <option value="blocked">Blocked ({statusCounts["blocked"] || 0})</option>
                    <option value="complete">Complete</option>
                </select>
                <select value={typeFilter} onChange={e => setTypeFilter(e.target.value)} style={selectStyle}>
                    <option value="">All types</option>
                    {itemTypes.map(t => <option key={t} value={t}>{t}</option>)}
                </select>
                <span style={{ fontSize: 13, color: "#64748b" }}>
                    {items.length} items
                </span>
                {failedCount > 0 && (
                    <button onClick={handleRetryAllFailed} disabled={actionLoading} style={{
                        ...btnSecondary, fontSize: 12, padding: "5px 12px", color: "#991b1b", borderColor: "#fca5a5",
                    }}>
                        Retry All Failed ({failedCount})
                    </button>
                )}
            </div>

            {loading ? (
                <div style={{ textAlign: "center", padding: 40, color: "#94a3b8" }}>Loading…</div>
            ) : items.length === 0 ? (
                <div style={{ textAlign: "center", padding: 40, color: "#94a3b8" }}>No items found</div>
            ) : (
                <div style={{ display: "flex", gap: 16 }}>
                    {/* Items list */}
                    <div style={{ flex: selectedItem ? "0 0 40%" : "1", minWidth: 0 }}>
                        {items.map(item => (
                            <div key={item.id} onClick={() => selectItem(item)}
                                 style={{
                                     background: selectedItem?.id === item.id ? "#f0f9ff" : "#fff",
                                     border: `1px solid ${selectedItem?.id === item.id ? "#93c5fd" : "#e2e8f0"}`,
                                     borderRadius: 8, padding: "14px 16px", marginBottom: 8, cursor: "pointer",
                                     transition: "all 0.1s",
                                 }}
                            >
                                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "start", gap: 8 }}>
                                    <div style={{ fontSize: 13, fontWeight: 500, color: "#0f172a", lineHeight: 1.4, flex: 1 }}>
                                        {item.summary?.slice(0, 100)}{item.summary?.length > 100 ? "…" : ""}
                                    </div>
                                    <Badge status={item.status} />
                                </div>
                                <div style={{ display: "flex", gap: 8, marginTop: 8, flexWrap: "wrap", alignItems: "center" }}>
                                    <span style={{ fontSize: 11, color: "#64748b", background: "#f1f5f9", padding: "2px 8px", borderRadius: 4 }}>{item.item_type}</span>
                                    <Badge status={item.severity} type="severity" />
                                    {item.spec?.checkpoint && (
                                        <span style={{ fontSize: 11, color: "#7c3aed", background: "#ede9fe", padding: "2px 8px", borderRadius: 4 }}>checkpoint</span>
                                    )}
                                    {!item.spec?.checkpoint && item.status === "needs_human_review" && (
                                        <span style={{ fontSize: 11, color: "#9d174d", background: "#fce7f3", padding: "2px 8px", borderRadius: 4 }}>needs input</span>
                                    )}
                                    {!siteFilter && item.domain && (
                                        <span style={{ fontSize: 11, color: "#1e40af", background: "#dbeafe", padding: "2px 8px", borderRadius: 4 }}>{item.domain}</span>
                                    )}
                                    <span style={{ fontSize: 11, color: "#94a3b8" }}>{item.handler_agent}</span>
                                    <span style={{ fontSize: 11, color: "#94a3b8" }}>{item.attempts}</span>
                                </div>
                                {item.status === "failed" && item.error && (
                                    <div style={{ fontSize: 11, color: "#991b1b", marginTop: 6, lineHeight: 1.3 }}>
                                        {item.error.slice(0, 120)}{item.error.length > 120 ? "…" : ""}
                                    </div>
                                )}
                            </div>
                        ))}
                    </div>

                    {/* Detail panel */}
                    {selectedItem && (
                        <div style={{
                            flex: "0 0 58%", background: "#fff", border: "1px solid #e2e8f0", borderRadius: 10,
                            padding: 20, position: "sticky", top: 16, maxHeight: "calc(100vh - 120px)", overflowY: "auto",
                        }}>
                            <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                                <h3 style={{ margin: 0, fontSize: 16, color: "#0f172a" }}>
                                    {isCheckpoint ? "Review & Approve" : isEditable ? "Review & Correct" : "Item Detail"}
                                </h3>
                                <span onClick={() => selectItem(null)} style={{ cursor: "pointer", color: "#94a3b8", fontSize: 18 }}>✕</span>
                            </div>

                            <div style={{ fontSize: 14, color: "#334155", lineHeight: 1.6, marginBottom: 16 }}>{selectedItem.summary}</div>

                            {/* Metadata grid */}
                            <div style={{ display: "grid", gridTemplateColumns: "auto 1fr", gap: "6px 16px", fontSize: 13, marginBottom: 16 }}>
                                <span style={detailLabel}>Status</span><Badge status={selectedItem.status} />
                                <span style={detailLabel}>Type</span><span style={detailValue}>{selectedItem.item_type}</span>
                                <span style={detailLabel}>Severity</span><Badge status={selectedItem.severity} type="severity" />
                                <span style={detailLabel}>Domain</span><span style={detailValue}>{selectedItem.domain}</span>
                                <span style={detailLabel}>Handler</span><span style={detailValue}>{selectedItem.handler_agent}</span>
                                <span style={detailLabel}>Attempts</span><span style={detailValue}>{selectedItem.attempts}</span>
                                <span style={detailLabel}>Created</span><span style={detailValue}>{new Date(selectedItem.created_at).toLocaleString()}</span>
                                {selectedItem.spec?.source_agent && <>
                                    <span style={detailLabel}>Source Agent</span><span style={detailValue}>{selectedItem.spec.source_agent}</span>
                                </>}
                                {selectedItem.error && <>
                                    <span style={detailLabel}>Error</span>
                                    <span style={{ ...detailValue, color: "#991b1b", fontSize: 12 }}>{selectedItem.error}</span>
                                </>}
                            </div>

                            {/* Editable form for all needs_human_review items */}
                            {isEditable && editedReviewData && (
                                <div style={{ marginBottom: 16 }}>
                                    {/* Context hints for placeholder items */}
                                    {selectedItem.item_type === "placeholder_content" && (
                                        <div style={{
                                            background: "#fffbeb", border: "1px solid #fde68a", borderRadius: 8,
                                            padding: "12px 14px", marginBottom: 12, fontSize: 13,
                                        }}>
                                            <div style={{ fontWeight: 600, color: "#92400e", marginBottom: 4 }}>
                                                Page: {selectedItem.spec?.page_name}
                                            </div>
                                            {selectedItem.spec?.fix_guidance && (
                                                <div style={{ color: "#78350f", marginBottom: 4 }}>
                                                    {selectedItem.spec.fix_guidance}
                                                </div>
                                            )}
                                            {selectedItem.spec?.missing_data && (
                                                <div style={{ color: "#92400e", fontSize: 12 }}>
                                                    What's needed: {selectedItem.spec.missing_data}
                                                </div>
                                            )}
                                        </div>
                                    )}

                                    <div style={{ fontSize: 13, fontWeight: 600, color: "#475569", marginBottom: 10 }}>
                                        {isCheckpoint ? "Review Data" :
                                            selectedItem.item_type === "placeholder_content" ? "Provide Real Data" :
                                                "Item Data"}
                                        {selectedItem.spec?.spec_aspect && (
                                            <span style={{ fontWeight: 400, color: "#94a3b8" }}> — will update spec '{selectedItem.spec.spec_aspect}'</span>
                                        )}
                                    </div>
                                    <div style={{
                                        background: "#f8fafc", border: "1px solid #e2e8f0", borderRadius: 8,
                                        padding: 16,
                                    }}>
                                        <EditableReviewForm
                                            reviewData={editedReviewData}
                                            onChange={setEditedReviewData}
                                        />
                                    </div>

                                    {isCheckpoint && selectedItem.spec?.on_approve && (
                                        <div style={{ fontSize: 12, color: "#64748b", marginTop: 8 }}>
                                            On approve: creates <strong>{selectedItem.spec.on_approve.item_type || "follow-on"}</strong> item
                                            {selectedItem.spec.on_approve.handler_agent && (
                                                <> → <strong>{selectedItem.spec.on_approve.handler_agent}</strong></>
                                            )}
                                        </div>
                                    )}

                                    {isCheckpoint && (
                                        <div style={{ marginTop: 12 }}>
                                            <label style={formLabelStyle}>Notes (optional)</label>
                                            <input
                                                type="text"
                                                value={approveNotes}
                                                onChange={e => setApproveNotes(e.target.value)}
                                                placeholder="Approval notes..."
                                                style={formInputStyle}
                                            />
                                        </div>
                                    )}
                                </div>
                            )}

                            {/* Read-only spec for non-editable items */}
                            {!isEditable && selectedItem.spec && (
                                <details style={{ marginBottom: 16 }}>
                                    <summary style={{ fontSize: 13, fontWeight: 600, color: "#475569", cursor: "pointer", marginBottom: 8 }}>Spec</summary>
                                    <pre style={{
                                        background: "#f8fafc", border: "1px solid #e2e8f0", borderRadius: 6,
                                        padding: 12, fontSize: 11, overflow: "auto", maxHeight: 200, lineHeight: 1.5,
                                    }}>
                                        {JSON.stringify(selectedItem.spec, null, 2)}
                                    </pre>
                                </details>
                            )}

                            {/* Actions */}
                            <div style={{ display: "flex", gap: 8, flexWrap: "wrap", borderTop: "1px solid #e2e8f0", paddingTop: 16 }}>
                                {/* Checkpoint items: approve button */}
                                {isCheckpoint && selectedItem.status === "needs_human_review" && (
                                    <button onClick={() => handleApprove(selectedItem.id)} disabled={actionLoading} style={{
                                        ...btnPrimary, background: "#059669",
                                    }}>
                                        Approve & Continue
                                    </button>
                                )}

                                {/* Non-checkpoint editable items: save & rebuild */}
                                {isEditable && !isCheckpoint && selectedItem.status === "needs_human_review" && (
                                    <button onClick={() => handleSaveAndRebuild(selectedItem)} disabled={actionLoading} style={{
                                        ...btnPrimary, background: "#059669",
                                    }}>
                                        Save & Rebuild
                                    </button>
                                )}

                                {/* Standard actions */}
                                {["needs_human_review", "failed", "blocked"].includes(selectedItem.status) && (
                                    <>
                                        {!isCheckpoint && !isEditable && (
                                            <button onClick={() => handleRetry(selectedItem.id)} disabled={actionLoading} style={btnPrimary}>
                                                Retry
                                            </button>
                                        )}
                                        <button onClick={() => {
                                            const reason = prompt("Resolution note (optional):");
                                            handleResolve(selectedItem.id, reason);
                                        }} disabled={actionLoading} style={btnSecondary}>
                                            {isCheckpoint ? "Reject / Skip" : isEditable ? "Skip / Resolve" : "Resolve"}
                                        </button>
                                    </>
                                )}
                                {selectedItem.status === "triaged" && (
                                    <button onClick={() => handleResolve(selectedItem.id, "Dismissed by admin")} disabled={actionLoading} style={btnSecondary}>
                                        Dismiss
                                    </button>
                                )}
                                {selectedItem.status === "blocked" && (
                                    <button onClick={() => handleUpdateStatus(selectedItem.id, "triaged")} disabled={actionLoading} style={btnSecondary}>
                                        Unblock → Triaged
                                    </button>
                                )}
                            </div>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

// ── Page Browser ─────────────────────────────────────────────────────────────
function PageBrowser({ token, siteId, siteDomain, onBack }) {
    const [pages, setPages] = useState([]);
    const [selectedPage, setSelectedPage] = useState(null);
    const [components, setComponents] = useState([]);
    const [suppressedSections, setSuppressedSections] = useState([]);
    const [siteComponents, setSiteComponents] = useState([]);
    const [viewMode, setViewMode] = useState(""); // "" = no selection, "site-wide" | page name
    const [selectedComponent, setSelectedComponent] = useState(null);
    const [editData, setEditData] = useState(null);
    const [editHtml, setEditHtml] = useState("");
    const [editMode, setEditMode] = useState("structured"); // structured | html
    const [loading, setLoading] = useState(true);
    const [actionLoading, setActionLoading] = useState(false);
    const [message, setMessage] = useState("");

    // Load pages
    const loadPages = useCallback(async () => {
        setLoading(true);
        try {
            const data = await apiFetch(`/sites/${siteId}/pages`, token);
            setPages(data.pages || []);
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    }, [token, siteId]);

    useEffect(() => { loadPages(); }, [loadPages]);

    // Load site-wide components
    const loadSiteComponents = useCallback(async () => {
        try {
            const data = await apiFetch(`/sites/${siteId}/site-components`, token);
            setSiteComponents(data.components || []);
        } catch (err) {
            console.error(err);
            setSiteComponents([]);
        }
    }, [token, siteId]);

    useEffect(() => { loadSiteComponents(); }, [loadSiteComponents]);

    // Load components for selected page
    const loadComponents = useCallback(async (pageName) => {
        try {
            const data = await apiFetch(`/sites/${siteId}/pages/${pageName}/components`, token);
            setComponents(data.components || []);
            setSuppressedSections(data.page?.suppressed_sections || []);
        } catch (err) {
            console.error(err);
            setComponents([]);
            setSuppressedSections([]);
        }
    }, [token, siteId]);

    const selectPage = (page) => {
        setSelectedPage(page);
        setViewMode(page.name);
        setSelectedComponent(null);
        setEditData(null);
        loadComponents(page.name);
    };

    const selectSiteWide = () => {
        setSelectedPage(null);
        setViewMode("site-wide");
        setSelectedComponent(null);
        setEditData(null);
    };

    const selectComponent = (comp) => {
        setSelectedComponent(comp);
        const hasStructured = comp.content_data && typeof comp.content_data === "object" && Object.keys(comp.content_data).length > 0;
        setEditMode(hasStructured ? "structured" : "html");
        if (hasStructured) {
            setEditData(JSON.parse(JSON.stringify(comp.content_data)));
        } else {
            setEditData(null);
        }
        // Site components have rendered_html, page components have html_preview
        setEditHtml(comp.rendered_html || comp.html_preview || "");
    };

    const handleSaveComponent = async (comp) => {
        setActionLoading(true);
        try {
            const body: Record<string, unknown> = { lock: true };
            if (viewMode === "site-wide") {
                body.rebuild_site = true;
                if (editMode === "html") body.rendered_html = editHtml;
                await apiFetch(
                    `/sites/${siteId}/site-components/${comp.slot_name}`,
                    token, { method: "PATCH", body: JSON.stringify(body) }
                );
                setMessage(`${comp.slot_name} saved & locked — full site rebuild queued`);
                loadSiteComponents();
            } else {
                body.rebuild_page = true;
                if (editMode === "structured" && editData) body.content_data = editData;
                else if (editMode === "html") body.rendered_html = editHtml;
                await apiFetch(
                    `/sites/${siteId}/pages/${selectedPage.name}/components/${comp.id}`,
                    token, { method: "PATCH", body: JSON.stringify(body) }
                );
                setMessage("Component saved & locked — rebuild queued");
                loadComponents(selectedPage.name);
            }
            loadPages();
            setSelectedComponent(null);
        } catch (err) { setMessage("Save failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const handleLock = async (comp) => {
        setActionLoading(true);
        try {
            if (viewMode === "site-wide") {
                await apiFetch(`/sites/${siteId}/site-components/${comp.slot_name}/lock`, token, { method: "POST" });
                setMessage(`${comp.slot_name} locked`);
                loadSiteComponents();
            } else {
                await apiFetch(`/sites/${siteId}/pages/${selectedPage.name}/components/${comp.id}/lock`, token, { method: "POST" });
                setMessage("Component locked");
                loadComponents(selectedPage.name);
            }
            loadPages();
        } catch (err) { setMessage("Lock failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const handleUnlock = async (comp) => {
        setActionLoading(true);
        try {
            if (viewMode === "site-wide") {
                await apiFetch(`/sites/${siteId}/site-components/${comp.slot_name}/unlock`, token, { method: "POST" });
                setMessage(`${comp.slot_name} unlocked — agents can modify it again`);
                loadSiteComponents();
            } else {
                await apiFetch(`/sites/${siteId}/pages/${selectedPage.name}/components/${comp.id}/unlock`, token, { method: "POST" });
                setMessage("Component unlocked — agents can modify it again");
                loadComponents(selectedPage.name);
            }
            loadPages();
        } catch (err) { setMessage("Unlock failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const handleRemove = async (comp) => {
        if (!confirm(`Remove section "${comp.slot_name}" from ${selectedPage.name}? It will be hidden on the live site and suppressed from future discovery.`)) return;
        setActionLoading(true);
        try {
            await apiFetch(
                `/sites/${siteId}/pages/${selectedPage.name}/components/${comp.id}`,
                token, { method: "DELETE" }
            );
            setMessage(`Section "${comp.slot_name}" removed — rebuild queued`);
            setSelectedComponent(null);
            loadComponents(selectedPage.name);
            loadPages();
        } catch (err) { setMessage("Remove failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const handleRestore = async (slotName) => {
        if (!confirm(`Restore section "${slotName}" on ${selectedPage.name}? A work item will be created to populate it.`)) return;
        setActionLoading(true);
        try {
            const result = await apiFetch(
                `/sites/${siteId}/pages/${selectedPage.name}/restore-section`,
                token, { method: "POST", body: JSON.stringify({ slot_name: slotName, create_item: true }) }
            );
            let msg = `Section "${slotName}" restored`;
            if (result.item_id) msg += " — populate item created";
            setMessage(msg);
            loadComponents(selectedPage.name);
            loadPages();
        } catch (err) { setMessage("Restore failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    return (
        <div>
            <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 20 }}>
                <button onClick={onBack} style={btnSecondary}>← Sites</button>
                <h2 style={{ ...sectionTitle, margin: 0, flex: 1 }}>
                    Pages <span style={{ fontWeight: 400, color: "#64748b" }}>— {siteDomain}</span>
                </h2>
            </div>

            {message && (
                <div style={{
                    background: "#d1fae5", color: "#065f46", padding: "8px 14px", borderRadius: 6,
                    fontSize: 13, marginBottom: 12, display: "flex", justifyContent: "space-between",
                }}>
                    {message}
                    <span onClick={() => setMessage("")} style={{ cursor: "pointer" }}>✕</span>
                </div>
            )}

            {loading ? (
                <div style={{ textAlign: "center", padding: 40, color: "#94a3b8" }}>Loading pages…</div>
            ) : (
                <div style={{ display: "flex", gap: 16 }}>
                    {/* Page list */}
                    <div style={{ flex: "0 0 220px", minWidth: 0 }}>
                        {/* Site-wide components entry */}
                        <div onClick={selectSiteWide} style={{
                            background: viewMode === "site-wide" ? "#f0f9ff" : "#fff",
                            border: `1px solid ${viewMode === "site-wide" ? "#93c5fd" : "#e2e8f0"}`,
                            borderRadius: 8, padding: "10px 14px", marginBottom: 6, cursor: "pointer",
                        }}>
                            <div style={{ fontSize: 13, fontWeight: 600, color: "#0f172a" }}>Site-Wide</div>
                            <div style={{ fontSize: 11, color: "#94a3b8", marginTop: 4 }}>
                                Header · Footer · CSS
                                {siteComponents.some(c => c.locked) && (
                                    <span style={{ color: "#7c3aed", marginLeft: 6 }}>🔒</span>
                                )}
                            </div>
                        </div>
                        <div style={{ borderBottom: "1px solid #e2e8f0", marginBottom: 6 }} />
                        {pages.map(page => (
                            <div key={page.id} onClick={() => selectPage(page)} style={{
                                background: viewMode === page.name ? "#f0f9ff" : "#fff",
                                border: `1px solid ${viewMode === page.name ? "#93c5fd" : "#e2e8f0"}`,
                                borderRadius: 8, padding: "10px 14px", marginBottom: 6, cursor: "pointer",
                            }}>
                                <div style={{ fontSize: 13, fontWeight: 600, color: "#0f172a" }}>{page.name}</div>
                                <div style={{ display: "flex", gap: 8, marginTop: 4, fontSize: 11, color: "#94a3b8" }}>
                                    <span>{page.component_count} sections</span>
                                    {page.locked_count > 0 && (
                                        <span style={{ color: "#7c3aed" }}>🔒{page.locked_count}</span>
                                    )}
                                    {page.empty_count > 0 && (
                                        <span style={{ color: "#92400e" }}>⚠{page.empty_count} empty</span>
                                    )}
                                </div>
                            </div>
                        ))}
                    </div>

                    {/* Components panel */}
                    <div style={{ flex: 1, minWidth: 0 }}>
                        {!viewMode ? (
                            <div style={{ textAlign: "center", padding: 40, color: "#94a3b8" }}>Select a page or Site-Wide to view sections</div>
                        ) : selectedComponent ? (
                            /* ── Component edit panel ── */
                            <div style={{
                                background: "#fff", border: "1px solid #e2e8f0", borderRadius: 10, padding: 20,
                            }}>
                                <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                                    <h3 style={{ margin: 0, fontSize: 16, color: "#0f172a" }}>
                                        Edit: {selectedComponent.slot_name}
                                        {selectedComponent.locked && <span style={{ fontSize: 12, color: "#7c3aed", marginLeft: 8 }}>🔒 Locked</span>}
                                    </h3>
                                    <span onClick={() => setSelectedComponent(null)} style={{ cursor: "pointer", color: "#94a3b8", fontSize: 18 }}>✕</span>
                                </div>

                                {/* Edit mode tabs */}
                                <div style={{ display: "flex", gap: 4, marginBottom: 12 }}>
                                    {selectedComponent.content_data && Object.keys(selectedComponent.content_data).length > 0 && (
                                        <button onClick={() => setEditMode("structured")} style={{
                                            ...btnSecondary, fontSize: 12, padding: "4px 12px",
                                            background: editMode === "structured" ? "#dbeafe" : "#fff",
                                            color: editMode === "structured" ? "#1e40af" : "#64748b",
                                        }}>Fields</button>
                                    )}
                                    <button onClick={() => setEditMode("html")} style={{
                                        ...btnSecondary, fontSize: 12, padding: "4px 12px",
                                        background: editMode === "html" ? "#dbeafe" : "#fff",
                                        color: editMode === "html" ? "#1e40af" : "#64748b",
                                    }}>HTML</button>
                                </div>

                                {/* Edit form */}
                                <div style={{
                                    background: "#f8fafc", border: "1px solid #e2e8f0", borderRadius: 8,
                                    padding: 16, marginBottom: 16, maxHeight: "50vh", overflowY: "auto",
                                }}>
                                    {editMode === "structured" && editData ? (
                                        <EditableReviewForm reviewData={editData} onChange={setEditData} />
                                    ) : (
                                        <textarea
                                            value={editHtml}
                                            onChange={e => setEditHtml(e.target.value)}
                                            style={{
                                                width: "100%", minHeight: 200, padding: 10,
                                                fontFamily: "monospace", fontSize: 12, border: "1px solid #cbd5e1",
                                                borderRadius: 6, resize: "vertical", boxSizing: "border-box",
                                            }}
                                        />
                                    )}
                                </div>

                                {/* Actions */}
                                <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
                                    <button onClick={() => handleSaveComponent(selectedComponent)} disabled={actionLoading} style={{
                                        ...btnPrimary, background: "#059669",
                                    }}>
                                        Save & Deploy
                                    </button>
                                    <button onClick={() => setSelectedComponent(null)} style={btnSecondary}>Cancel</button>
                                </div>
                            </div>
                        ) : viewMode === "site-wide" ? (
                            /* ── Site-wide component list ── */
                            <div>
                                <div style={{ fontSize: 14, fontWeight: 600, color: "#475569", marginBottom: 12 }}>
                                    Site-Wide Components — {siteComponents.length} items
                                </div>
                                {siteComponents.map(comp => {
                                    const label = comp.slot_name === "head" ? "CSS / Styles" : comp.slot_name;
                                    const previewText = comp.rendered_html ? stripHtmlTags(comp.rendered_html).slice(0, 200) : "";
                                    return (
                                        <div key={comp.id} style={{
                                            background: "#fff", border: "1px solid #e2e8f0", borderRadius: 8,
                                            padding: "14px 16px", marginBottom: 8,
                                        }}>
                                            <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 4 }}>
                                                <span style={{ fontSize: 13, fontWeight: 600, color: "#0f172a" }}>{label}</span>
                                                {comp.locked && (
                                                    <span style={{ fontSize: 11, color: "#7c3aed", background: "#ede9fe", padding: "1px 6px", borderRadius: 4 }}>
                                                        🔒 {comp.locked_by}
                                                    </span>
                                                )}
                                                <span style={{ fontSize: 11, color: "#94a3b8" }}>
                                                    {Math.round(comp.html_length / 1024)}kb
                                                </span>
                                            </div>
                                            {previewText && (
                                                <div style={{ fontSize: 12, color: "#64748b", lineHeight: 1.4, maxHeight: 60, overflow: "hidden" }}>
                                                    {previewText}
                                                </div>
                                            )}
                                            <div style={{ display: "flex", gap: 6, marginTop: 10 }}>
                                                <button onClick={() => selectComponent(comp)} style={{
                                                    ...btnSecondary, fontSize: 11, padding: "4px 10px",
                                                }}>Edit</button>
                                                {comp.locked ? (
                                                    <button onClick={() => handleUnlock(comp)} disabled={actionLoading} style={{
                                                        ...btnSecondary, fontSize: 11, padding: "4px 10px", color: "#7c3aed",
                                                    }}>Unlock</button>
                                                ) : (
                                                    <button onClick={() => handleLock(comp)} disabled={actionLoading} style={{
                                                        ...btnSecondary, fontSize: 11, padding: "4px 10px",
                                                    }}>Lock</button>
                                                )}
                                            </div>
                                        </div>
                                    );
                                })}
                                <div style={{ fontSize: 12, color: "#94a3b8", marginTop: 12, lineHeight: 1.5 }}>
                                    Editing a site-wide component triggers a full-site rebuild — all pages will be reassembled with the updated header, footer, or CSS.
                                </div>
                            </div>
                        ) : selectedPage ? (
                            /* ── Page component list ── */
                            <div>
                                <div style={{ fontSize: 14, fontWeight: 600, color: "#475569", marginBottom: 12 }}>
                                    {selectedPage.name} — {components.length} sections
                                </div>
                                {components.length === 0 ? (
                                    <div style={{ textAlign: "center", padding: 30, color: "#94a3b8" }}>No sections found</div>
                                ) : components.map(comp => (
                                    <div key={comp.id} style={{
                                        background: "#fff", border: "1px solid #e2e8f0", borderRadius: 8,
                                        padding: "14px 16px", marginBottom: 8,
                                    }}>
                                        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "start" }}>
                                            <div style={{ flex: 1 }}>
                                                <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 4 }}>
                                                    <span style={{ fontSize: 13, fontWeight: 600, color: "#0f172a" }}>{comp.slot_name}</span>
                                                    {comp.locked && (
                                                        <span style={{ fontSize: 11, color: "#7c3aed", background: "#ede9fe", padding: "1px 6px", borderRadius: 4 }}>
                                                            🔒 {comp.locked_by}
                                                        </span>
                                                    )}
                                                    {comp.is_empty && (
                                                        <span style={{ fontSize: 11, color: "#92400e", background: "#fef3c7", padding: "1px 6px", borderRadius: 4 }}>empty</span>
                                                    )}
                                                </div>
                                                {comp.html_preview && !comp.is_empty ? (
                                                    <div style={{
                                                        fontSize: 12, color: "#64748b", lineHeight: 1.4,
                                                        maxHeight: 60, overflow: "hidden",
                                                    }}>
                                                        {stripHtmlTags(comp.html_preview).slice(0, 200)}
                                                    </div>
                                                ) : comp.is_empty ? (
                                                    <div style={{ fontSize: 12, color: "#94a3b8", fontStyle: "italic" }}>No content</div>
                                                ) : null}
                                            </div>
                                        </div>
                                        <div style={{ display: "flex", gap: 6, marginTop: 10 }}>
                                            <button onClick={() => selectComponent(comp)} style={{
                                                ...btnSecondary, fontSize: 11, padding: "4px 10px",
                                            }}>Edit</button>
                                            {comp.locked ? (
                                                <button onClick={() => handleUnlock(comp)} disabled={actionLoading} style={{
                                                    ...btnSecondary, fontSize: 11, padding: "4px 10px", color: "#7c3aed",
                                                }}>Unlock</button>
                                            ) : (
                                                <button onClick={() => handleLock(comp)} disabled={actionLoading} style={{
                                                    ...btnSecondary, fontSize: 11, padding: "4px 10px",
                                                }}>Lock</button>
                                            )}
                                            <button onClick={() => handleRemove(comp)} disabled={actionLoading} style={{
                                                ...btnSecondary, fontSize: 11, padding: "4px 10px", color: "#991b1b", borderColor: "#fca5a5",
                                            }}>Remove</button>
                                        </div>
                                    </div>
                                ))}

                                {suppressedSections.length > 0 && (
                                    <div style={{ marginTop: 20 }}>
                                        <div style={{ fontSize: 12, fontWeight: 600, color: "#94a3b8", marginBottom: 8, textTransform: "uppercase", letterSpacing: 0.5 }}>
                                            Suppressed ({suppressedSections.length})
                                        </div>
                                        {suppressedSections.map(slot => (
                                            <div key={slot} style={{
                                                background: "#f8fafc", border: "1px dashed #cbd5e1", borderRadius: 8,
                                                padding: "10px 14px", marginBottom: 6,
                                                display: "flex", justifyContent: "space-between", alignItems: "center",
                                            }}>
                                                <div>
                                                    <span style={{ fontSize: 13, color: "#94a3b8", textDecoration: "line-through" }}>{slot}</span>
                                                    <span style={{ fontSize: 11, color: "#94a3b8", marginLeft: 8 }}>removed by admin</span>
                                                </div>
                                                <button onClick={() => handleRestore(slot)} disabled={actionLoading} style={{
                                                    ...btnSecondary, fontSize: 11, padding: "4px 10px", color: "#059669",
                                                }}>Restore</button>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>
                        ) : null}
                    </div>
                </div>
            )}
        </div>
    );
}

// ── Spec Editor (Direction Control) ──────────────────────────────────────────
function SpecEditor({ token, siteId, siteDomain, onBack }) {
    const [specs, setSpecs] = useState([]);
    const [loading, setLoading] = useState(true);
    const [selectedSpec, setSelectedSpec] = useState(null);
    const [editData, setEditData] = useState(null);
    const [actionLoading, setActionLoading] = useState(false);
    const [message, setMessage] = useState("");

    const loadSpecs = useCallback(async () => {
        setLoading(true);
        try {
            const data = await apiFetch(`/sites/${siteId}/specs`, token);
            setSpecs(data.specs || []);
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    }, [token, siteId]);

    useEffect(() => { loadSpecs(); }, [loadSpecs]);

    const selectSpec = (spec) => {
        setSelectedSpec(spec);
        setEditData(JSON.parse(JSON.stringify(spec.data)));
    };

    const handleSaveSpec = async () => {
        if (!selectedSpec || !editData) return;
        setActionLoading(true);
        try {
            await apiFetch(`/sites/${siteId}/specs/${selectedSpec.aspect}`, token, {
                method: "PATCH",
                body: JSON.stringify({ data: editData }),
            });
            setMessage(`Spec '${selectedSpec.aspect}' updated`);
            setSelectedSpec(null);
            setEditData(null);
            loadSpecs();
        } catch (err) { setMessage("Save failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const handlePin = async (spec) => {
        setActionLoading(true);
        try {
            await apiFetch(`/sites/${siteId}/specs/${spec.aspect}/pin`, token, { method: "POST" });
            setMessage(`Spec '${spec.aspect}' pinned — agents won't override it`);
            loadSpecs();
        } catch (err) { setMessage("Pin failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const handleUnpin = async (spec) => {
        setActionLoading(true);
        try {
            await apiFetch(`/sites/${siteId}/specs/${spec.aspect}/unpin`, token, { method: "POST" });
            setMessage(`Spec '${spec.aspect}' unpinned — agents can update it`);
            loadSpecs();
        } catch (err) { setMessage("Unpin failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const handlePropagate = async (spec) => {
        if (!confirm(`Create work items to propagate '${spec.aspect}' changes across all content pages?`)) return;
        setActionLoading(true);
        try {
            const result = await apiFetch(`/sites/${siteId}/specs/${spec.aspect}/propagate`, token, {
                method: "POST",
                body: JSON.stringify({}),
            });
            let msg = `Propagation: ${result.items_created} work items created`;
            if (result.items_skipped > 0) msg += `, ${result.items_skipped} fully-locked pages skipped`;
            setMessage(msg);
        } catch (err) { setMessage("Propagate failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    return (
        <div>
            <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 20 }}>
                <button onClick={onBack} style={btnSecondary}>← Sites</button>
                <h2 style={{ ...sectionTitle, margin: 0, flex: 1 }}>
                    Direction <span style={{ fontWeight: 400, color: "#64748b" }}>— {siteDomain}</span>
                </h2>
            </div>

            {message && (
                <div style={{
                    background: "#d1fae5", color: "#065f46", padding: "8px 14px", borderRadius: 6,
                    fontSize: 13, marginBottom: 12, display: "flex", justifyContent: "space-between",
                }}>
                    {message}
                    <span onClick={() => setMessage("")} style={{ cursor: "pointer" }}>✕</span>
                </div>
            )}

            {loading ? (
                <div style={{ textAlign: "center", padding: 40, color: "#94a3b8" }}>Loading specs…</div>
            ) : specs.length === 0 ? (
                <div style={{ textAlign: "center", padding: 40, color: "#94a3b8" }}>No specs found for this site</div>
            ) : selectedSpec ? (
                /* ── Spec edit panel ── */
                <div style={{ background: "#fff", border: "1px solid #e2e8f0", borderRadius: 10, padding: 20 }}>
                    <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                        <h3 style={{ margin: 0, fontSize: 16, color: "#0f172a" }}>
                            Edit: {selectedSpec.aspect}
                            {selectedSpec.pinned && <span style={{ fontSize: 12, color: "#7c3aed", marginLeft: 8 }}>🔒 Pinned</span>}
                        </h3>
                        <span onClick={() => { setSelectedSpec(null); setEditData(null); }} style={{ cursor: "pointer", color: "#94a3b8", fontSize: 18 }}>✕</span>
                    </div>

                    <div style={{ fontSize: 12, color: "#94a3b8", marginBottom: 12 }}>
                        Source: {selectedSpec.source_agent || selectedSpec.source || "unknown"}
                        {selectedSpec.created_at && <> · {new Date(selectedSpec.created_at).toLocaleDateString()}</>}
                    </div>

                    <div style={{
                        background: "#f8fafc", border: "1px solid #e2e8f0", borderRadius: 8,
                        padding: 16, marginBottom: 16, maxHeight: "60vh", overflowY: "auto",
                    }}>
                        <EditableReviewForm reviewData={editData} onChange={setEditData} />
                    </div>

                    <div style={{ display: "flex", gap: 8 }}>
                        <button onClick={handleSaveSpec} disabled={actionLoading} style={btnPrimary}>
                            Save Spec
                        </button>
                        <button onClick={() => handlePropagate(selectedSpec)} disabled={actionLoading} style={{
                            ...btnSecondary, color: "#1e40af",
                        }}>
                            Save & Propagate
                        </button>
                        <button onClick={() => { setSelectedSpec(null); setEditData(null); }} style={btnSecondary}>Cancel</button>
                    </div>
                </div>
            ) : (
                /* ── Spec list ── */
                <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
                    {specs.map(spec => (
                        <div key={spec.id} style={{
                            background: "#fff", border: "1px solid #e2e8f0", borderRadius: 8,
                            padding: "16px 18px",
                        }}>
                            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "start", marginBottom: 8 }}>
                                <div>
                                    <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                                        <span style={{ fontSize: 15, fontWeight: 600, color: "#0f172a" }}>{spec.aspect}</span>
                                        {spec.pinned && (
                                            <span style={{ fontSize: 11, color: "#7c3aed", background: "#ede9fe", padding: "1px 8px", borderRadius: 4 }}>
                                                🔒 Pinned
                                            </span>
                                        )}
                                    </div>
                                    <div style={{ fontSize: 12, color: "#94a3b8", marginTop: 2 }}>
                                        {spec.source_agent || spec.source || "unknown"}
                                        {spec.created_at && <> · {new Date(spec.created_at).toLocaleDateString()}</>}
                                    </div>
                                </div>
                                <div style={{ display: "flex", gap: 6 }}>
                                    <button onClick={() => selectSpec(spec)} style={{
                                        ...btnSecondary, fontSize: 11, padding: "4px 10px",
                                    }}>Edit</button>
                                    {spec.pinned ? (
                                        <button onClick={() => handleUnpin(spec)} disabled={actionLoading} style={{
                                            ...btnSecondary, fontSize: 11, padding: "4px 10px", color: "#7c3aed",
                                        }}>Unpin</button>
                                    ) : (
                                        <button onClick={() => handlePin(spec)} disabled={actionLoading} style={{
                                            ...btnSecondary, fontSize: 11, padding: "4px 10px",
                                        }}>Pin</button>
                                    )}
                                    <button onClick={() => handlePropagate(spec)} disabled={actionLoading} style={{
                                        ...btnSecondary, fontSize: 11, padding: "4px 10px", color: "#1e40af",
                                    }}>Propagate</button>
                                </div>
                            </div>

                            {/* Spec data preview */}
                            <div style={{
                                background: "#f8fafc", border: "1px solid #f1f5f9", borderRadius: 6,
                                padding: 10, fontSize: 12, maxHeight: 120, overflow: "hidden",
                                color: "#475569", lineHeight: 1.5,
                            }}>
                                {renderSpecPreview(spec.data)}
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

// Render a compact preview of spec data — show top-level keys and values
function renderSpecPreview(data) {
    if (!data || typeof data !== "object") return String(data || "");
    return Object.entries(data).slice(0, 6).map(([key, value]) => {
        let preview = "";
        if (typeof value === "string") preview = value.slice(0, 80) + (value.length > 80 ? "…" : "");
        else if (Array.isArray(value)) preview = `[${value.length} items]`;
        else if (typeof value === "object" && value !== null) preview = `{${Object.keys(value).join(", ")}}`;
        else preview = String(value);
        return (
            <div key={key} style={{ marginBottom: 2 }}>
                <span style={{ fontWeight: 600, color: "#334155" }}>{key}:</span>{" "}
                <span>{preview}</span>
            </div>
        );
    });
}

// Strip HTML tags for content preview
function stripHtmlTags(html) {
    return html.replace(/<style[^>]*>[\s\S]*?<\/style>/gi, '')
        .replace(/<[^>]+>/g, ' ')
        .replace(/\s+/g, ' ')
        .trim();
}

// ── Main App ─────────────────────────────────────────────────────────────────
export default function App() {
    const [token, setToken] = useState(() => sessionStorage.getItem("admin_token") || "");
    const [user, setUser] = useState(null);
    const [sites, setSites] = useState([]);
    const [view, setView] = useState("sites"); // sites | items | all-items | pages | specs
    const [selectedSite, setSelectedSite] = useState(null);
    const [error, setError] = useState("");

    const handleLogin = (tok, usr) => {
        setToken(tok);
        setUser(usr);
        sessionStorage.setItem("admin_token", tok);
    };

    const handleLogout = () => {
        setToken("");
        setUser(null);
        sessionStorage.removeItem("admin_token");
    };

    // Load sites
    const loadSites = useCallback(() => {
        if (!token) return;
        apiFetch("/sites", token)
            .then(data => setSites(data.sites || []))
            .catch(err => {
                if (err.message === "UNAUTHORIZED") handleLogout();
                else setError(err.message);
            });
    }, [token]);

    useEffect(() => { loadSites(); }, [loadSites]);

    if (!token) return <LoginScreen onLogin={handleLogin} />;

    return (
        <div style={{
            minHeight: "100vh", background: "#f8fafc",
            fontFamily: "'IBM Plex Sans', -apple-system, BlinkMacSystemFont, sans-serif",
        }}>
            {/* Top bar */}
            <div style={{
                background: "#0f172a", color: "#f1f5f9", padding: "0 24px",
                display: "flex", alignItems: "center", justifyContent: "space-between", height: 52,
            }}>
                <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
                    <span style={{ fontSize: 13, fontWeight: 600, letterSpacing: 1.5, color: "#94a3b8" }}>PERSONAE</span>
                    <span style={{ fontSize: 15, fontWeight: 500 }}>Admin</span>
                    <div style={{ display: "flex", gap: 4, marginLeft: 16 }}>
                        {[
                            { key: "sites", label: "Sites" },
                            { key: "all-items", label: "All Items" },
                        ].map(({ key, label }) => (
                            <button key={key} onClick={() => { setView(key); setSelectedSite(null); }} style={{
                                background: view === key || (["items", "pages", "specs"].includes(view) && key === "sites") ? "#1e293b" : "transparent",
                                color: view === key || (["items", "pages", "specs"].includes(view) && key === "sites") ? "#f1f5f9" : "#94a3b8",
                                border: "none", padding: "6px 12px", borderRadius: 4,
                                fontSize: 12, fontWeight: 500, cursor: "pointer",
                            }}>
                                {label}
                            </button>
                        ))}
                    </div>
                </div>
                <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
                    <span style={{ fontSize: 12, color: "#94a3b8" }}>{user?.email || "admin"}</span>
                    <button onClick={handleLogout} style={{
                        background: "none", border: "1px solid #334155", color: "#94a3b8",
                        padding: "4px 12px", borderRadius: 6, fontSize: 12, cursor: "pointer",
                    }}>
                        Logout
                    </button>
                </div>
            </div>

            {/* Content */}
            <div style={{ maxWidth: 1280, margin: "0 auto", padding: "24px 24px 60px" }}>
                {error && (
                    <div style={{ background: "#fee2e2", color: "#991b1b", padding: "10px 14px", borderRadius: 6, fontSize: 13, marginBottom: 16 }}>
                        {error}
                        <span onClick={() => setError("")} style={{ float: "right", cursor: "pointer" }}>✕</span>
                    </div>
                )}

                {view === "sites" && (
                    <SitesOverview
                        sites={sites}
                        token={token}
                        onSelectSite={(site) => { setSelectedSite(site); setView("items"); }}
                        onSelectPages={(site) => { setSelectedSite(site); setView("pages"); }}
                        onSelectSpecs={(site) => { setSelectedSite(site); setView("specs"); }}
                        onRefresh={loadSites}
                    />
                )}

                {view === "items" && (
                    <WorkItemsList
                        token={token}
                        siteFilter={selectedSite}
                        onBack={() => { setView("sites"); setSelectedSite(null); loadSites(); }}
                    />
                )}

                {view === "all-items" && (
                    <WorkItemsList
                        token={token}
                        siteFilter={null}
                        onBack={() => { setView("sites"); loadSites(); }}
                    />
                )}

                {view === "pages" && selectedSite && (
                    <PageBrowser
                        token={token}
                        siteId={selectedSite.id}
                        siteDomain={selectedSite.domain}
                        onBack={() => { setView("sites"); setSelectedSite(null); loadSites(); }}
                    />
                )}

                {view === "specs" && selectedSite && (
                    <SpecEditor
                        token={token}
                        siteId={selectedSite.id}
                        siteDomain={selectedSite.domain}
                        onBack={() => { setView("sites"); setSelectedSite(null); loadSites(); }}
                    />
                )}
            </div>
        </div>
    );
}

// ── Shared styles ────────────────────────────────────────────────────────────
const labelStyle = { display: "block", fontSize: 13, fontWeight: 500, color: "#475569", marginBottom: 4, marginTop: 12 };
const inputStyle = {
    width: "100%", padding: "10px 12px", border: "1px solid #cbd5e1", borderRadius: 8,
    fontSize: 14, outline: "none", boxSizing: "border-box",
};
const selectStyle = {
    padding: "7px 12px", border: "1px solid #cbd5e1", borderRadius: 6,
    fontSize: 13, background: "#fff", color: "#334155", outline: "none", cursor: "pointer",
};
const btnPrimary = {
    padding: "8px 18px", background: "#1e40af", color: "#fff", border: "none",
    borderRadius: 6, fontSize: 13, fontWeight: 500, cursor: "pointer",
};
const btnSecondary = {
    padding: "8px 18px", background: "#fff", color: "#334155", border: "1px solid #cbd5e1",
    borderRadius: 6, fontSize: 13, fontWeight: 500, cursor: "pointer",
};
const sectionTitle = { fontSize: 20, fontWeight: 600, color: "#0f172a", marginBottom: 16 };
const detailLabel = { color: "#64748b", fontWeight: 500 };
const detailValue = { color: "#0f172a" };

// Form-specific styles
const formLabelStyle = { display: "block", fontSize: 12, fontWeight: 600, color: "#475569", marginBottom: 4 };
const formInputStyle = {
    width: "100%", padding: "8px 10px", border: "1px solid #cbd5e1", borderRadius: 6,
    fontSize: 13, outline: "none", boxSizing: "border-box",
};
const textareaStyle = {
    width: "100%", padding: "8px 10px", border: "1px solid #cbd5e1", borderRadius: 6,
    fontSize: 13, outline: "none", boxSizing: "border-box", resize: "vertical", minHeight: 60,
};
