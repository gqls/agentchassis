import { useState, useEffect, useCallback } from "react";
import PipelinesPage from "./pages/PipelinesPage";
import CustomersPage from "./pages/CustomersPage";

const API_BASE = "/api/v1/admin";

// Work-item paging. Must not exceed the server's own ceiling
// (maxWorkItemPageSize in internal/core-manager/admin/site_admin_handlers.go),
// which clamps silently — asking for more than it allows would quietly return
// less than requested, which is the failure this paging exists to end.
const PAGE_SIZE = 200;
const MAX_PAGE_SIZE = 1000;

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
    unresolved: { bg: "#fff7ed", text: "#9a3412", label: "Unresolved" },
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
function SitesOverview({ sites, token, onSelectSite, onSelectPages, onSelectSpecs, onSelectMedia, onSelectBuilds, onRefresh }) {
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
                                { k: "unresolved", label: "Unresolved", color: "#9a3412" },
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
                            <button onClick={() => onSelectBuilds(site)} style={{
                                ...btnSecondary, fontSize: 12, padding: "6px 14px", flex: 1,
                            }}>Builds</button>
                            <button onClick={() => onSelectSite(site)} style={{
                                ...btnSecondary, fontSize: 12, padding: "6px 14px", flex: 1,
                            }}>Work Items</button>
                            <button onClick={() => onSelectPages(site)} style={{
                                ...btnSecondary, fontSize: 12, padding: "6px 14px", flex: 1,
                            }}>Pages</button>
                            <button onClick={() => onSelectSpecs(site)} style={{
                                ...btnSecondary, fontSize: 12, padding: "6px 14px", flex: 1,
                            }}>Direction</button>
                            <button onClick={() => onSelectMedia(site)} style={{
                                ...btnSecondary, fontSize: 12, padding: "6px 14px", flex: 1,
                            }}>Media</button>
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
// reviewQueueMode: entered via the top-nav "Review Queue" tab (bugs_open/428) —
// the same list and the same RFC_056 release action below, just landing
// pre-filtered on the record-verdicts queue instead of requiring a reviewer to
// find the checkbox inside "All Items" first.
function WorkItemsList({ token, siteFilter, onBack, reviewQueueMode = false }) {
    const [allItems, setAllItems] = useState([]);
    const [loading, setLoading] = useState(true);
    const [statusFilter, setStatusFilter] = useState(reviewQueueMode ? "deferred" : "");
    const [typeFilter, setTypeFilter] = useState("");
    const [pipelineFilter, setPipelineFilter] = useState("build");
    // RFC_056 (bugs_open/428): filing_mode='record' rows are 'deferred' verdicts
    // an LLM-audit seat filed but nothing may auto-dispatch — status='deferred'
    // was not even a status option below until this filter was added, so a
    // reviewer had no route to these rows shorter than a raw SQL query.
    const [recordVerdictsOnly, setRecordVerdictsOnly] = useState(reviewQueueMode);
    // Counts and totals come from the server. They used to be derived from the
    // returned rows, which the API caps — so a 208-item needs_human_review
    // backlog reported itself as 0 and the queue looked empty (bugs_open/033).
    const [serverStatusCounts, setServerStatusCounts] = useState({});
    const [serverTypeCounts, setServerTypeCounts] = useState({});
    const [total, setTotal] = useState(0);
    const [truncated, setTruncated] = useState(false);
    const [loadingMore, setLoadingMore] = useState(false);
    const [selectedItem, setSelectedItem] = useState(null);
    const [actionLoading, setActionLoading] = useState(false);
    const [message, setMessage] = useState("");

    // Editable review data for checkpoint items
    const [editedReviewData, setEditedReviewData] = useState(null);
    const [approveNotes, setApproveNotes] = useState("");
    const [critiqueText, setCritiqueText] = useState("");

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

    // Filter SERVER-side. Filtering client-side over a capped response silently
    // pre-empts the filter: the rows you want may never have been sent.
    const buildPath = useCallback((offset) => {
        let path = `/work-items?pipeline=${encodeURIComponent(pipelineFilter)}&limit=${PAGE_SIZE}&offset=${offset}`;
        if (siteFilter?.id) path += `&site_id=${siteFilter.id}`;
        if (statusFilter) path += `&status=${encodeURIComponent(statusFilter)}`;
        if (typeFilter) path += `&item_type=${encodeURIComponent(typeFilter)}`;
        if (recordVerdictsOnly) path += `&filing_mode=record`;
        return path;
    }, [siteFilter, statusFilter, typeFilter, pipelineFilter, recordVerdictsOnly]);

    const loadItems = useCallback(async () => {
        setLoading(true);
        try {
            const data = await apiFetch(buildPath(0), token);
            setAllItems(data.items || []);
            setServerStatusCounts(data.status_counts || {});
            setServerTypeCounts(data.type_counts || {});
            setTotal(data.total ?? (data.items || []).length);
            setTruncated(Boolean(data.truncated));
        } catch (err) {
            if (err.message === "UNAUTHORIZED") return;
            console.error(err);
        } finally {
            setLoading(false);
        }
    }, [token, buildPath]);

    const loadMore = useCallback(async () => {
        setLoadingMore(true);
        try {
            const data = await apiFetch(buildPath(allItems.length), token);
            setAllItems(prev => [...prev, ...(data.items || [])]);
            setTruncated(Boolean(data.truncated));
        } catch (err) {
            if (err.message === "UNAUTHORIZED") return;
            console.error(err);
        } finally {
            setLoadingMore(false);
        }
    }, [token, buildPath, allItems.length]);

    useEffect(() => { loadItems(); }, [loadItems]);

    // The server has already applied the filters; these are the rows to render.
    const items = allItems;

    // When selecting a review item, initialise the editable data
    const selectItem = (item) => {
        setSelectedItem(item);
        setApproveNotes("");
        setCritiqueText("");
        if (item?.status === "needs_human_review" && item?.spec) {
            if (item.spec.checkpoint && item.spec.review_data) {
                // Checkpoint items: edit the review_data
                setEditedReviewData(JSON.parse(JSON.stringify(item.spec.review_data)));
            } else if (item.item_type === "placeholder_content") {
                // Placeholder content: build input form based on page type and missing_data
                setEditedReviewData(buildPlaceholderForm(item));
            } else if (item.item_type === "needs_section_data") {
                // Section data items: build form from missing[] array
                setEditedReviewData(buildSectionDataForm(item));
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

    // Build input form for needs_section_data items.
    // Uses enriched type/items from plan_sections when available,
    // falls back to heuristics based on field name and reason.
    function buildSectionDataForm(item) {
        const missing = item.spec?.missing || [];
        if (missing.length === 0) {
            return { content: "", notes: "" };
        }

        const form = {};
        for (const entry of missing) {
            const field = entry.field || "data";
            const type = entry.type || "";
            const items = entry.items || null;
            const reason = (entry.reason || "").toLowerCase();

            if (type === "array" && items && typeof items === "object") {
                // Schema-driven: build template from items definition
                const template = {};
                for (const [key] of Object.entries(items)) {
                    template[key] = "";
                }
                form[field] = [template];
            } else if (type === "array") {
                form[field] = [inferArrayTemplate(field, reason)];
            } else if (type === "text" || type === "url" || type === "image") {
                form[field] = "";
            } else if (type === "boolean") {
                form[field] = false;
            } else {
                // No type info — heuristic fallback
                if (looksLikeArray(field, reason)) {
                    form[field] = [inferArrayTemplate(field, reason)];
                } else {
                    form[field] = "";
                }
            }
        }

        return form;
    }

    function looksLikeArray(field, reason) {
        const arrayHints = [
            "use_case", "case_stud", "testimonial", "team_member",
            "member", "plan", "tier", "pricing", "service",
            "portfolio", "project", "client", "partner", "faq",
            "feature", "benefit", "step", "review",
        ];
        const combined = (field + " " + reason).toLowerCase();
        return arrayHints.some(h => combined.includes(h));
    }

    function inferArrayTemplate(field, reason) {
        const combined = (field + " " + reason).toLowerCase();

        if (combined.includes("use_case") || combined.includes("case_stud") || combined.includes("case study")) {
            return { client_name: "", description: "", outcome: "" };
        }
        if (combined.includes("team") || combined.includes("member") || combined.includes("bio")) {
            return { name: "", title: "", bio: "" };
        }
        if (combined.includes("testimonial") || combined.includes("review") || combined.includes("quote")) {
            return { name: "", company: "", quote: "" };
        }
        if (combined.includes("pricing") || combined.includes("plan") || combined.includes("tier")) {
            return { name: "", price: "", features: "", cta: "" };
        }
        if (combined.includes("service")) {
            return { name: "", description: "", features: "" };
        }
        if (combined.includes("faq") || combined.includes("question")) {
            return { question: "", answer: "" };
        }
        if (combined.includes("portfolio") || combined.includes("project")) {
            return { name: "", description: "", url: "" };
        }
        if (combined.includes("partner") || combined.includes("client")) {
            return { name: "", description: "" };
        }
        if (combined.includes("feature") || combined.includes("benefit")) {
            return { title: "", description: "" };
        }
        if (combined.includes("step")) {
            return { title: "", description: "" };
        }
        return { name: "", description: "" };
    }

    // Read-merge-write helper for site_specs.
    // The PATCH endpoint does a full replace, so we read existing data,
    // merge in the new fields, and write back.
    async function mergeIntoSpec(siteId, aspect, newFields) {
        const specsResult = await apiFetch(`/sites/${siteId}/specs`, token);
        const existing = (specsResult.specs || []).find(s => s.aspect === aspect);
        const existingData = (existing && typeof existing.data === "object") ? existing.data : {};
        const merged = { ...existingData, ...newFields };
        await apiFetch(`/sites/${siteId}/specs/${aspect}`, token, {
            method: "PATCH",
            body: JSON.stringify({ data: merged }),
        });
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

    // RFC_056 release (bugs_open/428). Releases ONE reviewed verdict row — the
    // human half of write_audit_findings_action.go's filing_mode='record'
    // circuit breaker. Requires a name so the row records who reviewed it;
    // the server refuses to release anything that isn't a genuine, still-
    // parked record verdict, so this can never turn into a bulk action.
    const handleRelease = async (id) => {
        const releasedBy = prompt("Your name (recorded on the row as who reviewed and released it):");
        if (!releasedBy) return;
        const notes = prompt("Why release this one? (optional, but the whole point of review is a reason)") || "";
        setActionLoading(true);
        try {
            const result = await apiFetch(`/work-items/${id}/release`, token, {
                method: "POST",
                body: JSON.stringify({ released_by: releasedBy, notes }),
            });
            setMessage(`Released → ${result.status} / ${result.handler_agent}`);
            selectItem(null);
            loadItems();
        } catch (err) { setMessage("Release failed: " + err.message); }
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

    // File the owner's critique WITHOUT approving or resolving — the item stays
    // open (delivery stays gated) and the critique becomes an owner_critique work
    // item that the dispatcher thread routes to the working sessions.
    const handleRequestChanges = async (id) => {
        if (!critiqueText.trim()) {
            setMessage("Type the critique first");
            return;
        }
        setActionLoading(true);
        try {
            const result = await apiFetch(`/work-items/${id}/request_changes`, token, {
                method: "POST",
                body: JSON.stringify({ critique: critiqueText.trim() }),
            });
            setMessage(`Critique filed (${(result.critique_item_id || "").slice(0, 8)}) — this item stays open; the threads will pick it up`);
            setCritiqueText("");
            loadItems();
        } catch (err) { setMessage("Request changes failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    // Save edited data and create a rebuild work item for non-checkpoint review items
    const handleSaveAndRebuild = async (item) => {
        if (!editedReviewData) {
            setMessage("No data to save");
            return;
        }

        // Check that at least one field has content
        const hasContent = Object.values(editedReviewData as Record<string, unknown>).some(v => {
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

            if (item.item_type === "needs_section_data") {
                // ── needs_section_data: write each field to its source spec ──
                const missing = item.spec?.missing || [];

                for (const entry of missing) {
                    const fieldName = entry.field;
                    const source = entry.source || "";
                    const value = editedReviewData[fieldName];

                    if (value === undefined || value === "" ||
                        (Array.isArray(value) && value.length === 0)) {
                        continue;
                    }

                    // Parse source: "site_specs.{aspect}.{field_path}"
                    const sourceMatch = source.match(/^site_specs\.([^.]+)(?:\.(.+))?$/);
                    if (sourceMatch) {
                        const aspect = sourceMatch[1];
                        const specField = sourceMatch[2] || fieldName;
                        await mergeIntoSpec(item.site_id, aspect, { [specField]: value });
                    } else {
                        // Unknown source pattern — save to generic aspect
                        await mergeIntoSpec(item.site_id, "section_data", { [fieldName]: value });
                    }
                }

                // Create rebuild work item for the page
                await apiFetch(`/work-items`, token, {
                    method: "POST",
                    body: JSON.stringify({
                        site_id: item.site_id,
                        item_type: "content_rewrite",
                        summary: `Rebuild ${pageName} — section data provided for ${item.spec?.section_name || "section"}`,
                        severity: "high",
                        handler_agent: "page-build-handler",
                        page_id: pageId || undefined,
                        priority: 10,
                        spec: {
                            page_name: pageName,
                            reason: "section_data_provided",
                            section_name: item.spec?.section_name,
                        },
                    }),
                });

                // Resolve the HITL item
                await apiFetch(`/work-items/${item.id}/resolve`, token, {
                    method: "POST",
                    body: JSON.stringify({
                        resolution: `Section data provided for ${item.spec?.section_name} on ${pageName}`,
                    }),
                });

                setMessage(`Data saved for ${item.spec?.section_name} — rebuild queued for ${pageName}`);

            } else {
                // ── Existing flow for placeholder_content and other types ────
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

                await apiFetch(`/work-items/${item.id}/resolve`, token, {
                    method: "POST",
                    body: JSON.stringify({ resolution: `Real data provided for ${pageName}, rebuild queued` }),
                });

                setMessage(`Data saved for ${pageName} — rebuild queued`);
            }

            selectItem(null);
            loadItems();
        } catch (err) { setMessage("Save failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    // Item types come from the server's GROUP BY, not from the returned rows —
    // a type absent from the current page was previously unlistable and so
    // unfilterable, which hid whole categories of work.
    const itemTypes = Object.keys(serverTypeCounts).sort();
    const isCheckpoint = selectedItem?.spec?.checkpoint === true;
    const isEditable = selectedItem?.status === "needs_human_review" && editedReviewData != null;

    // Server-side counts: totals for the whole filtered set, not for this page.
    const statusCounts = serverStatusCounts;
    const allItemsCount = Object.values(serverStatusCounts).reduce(
        (a: number, b: number) => a + b, 0) - (serverStatusCounts["complete"] || 0);
    const failedCount = statusCounts["failed"] || 0;

    const handleRetryAllFailed = async () => {
        if (!confirm(`Retry all ${failedCount} failed items?`)) return;
        setActionLoading(true);
        try {
            // Fetch the failed set explicitly rather than filtering the loaded
            // page: failedCount is now a true total, so scanning only the rows
            // that happen to be on screen would retry a fraction of what the
            // button just promised.
            let failedPath = `/work-items?pipeline=${encodeURIComponent(pipelineFilter)}&status=failed&limit=${MAX_PAGE_SIZE}`;
            if (siteFilter?.id) failedPath += `&site_id=${siteFilter.id}`;
            const failedData = await apiFetch(failedPath, token);
            const failedItems = failedData.items || [];
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
            <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: reviewQueueMode ? 4 : 20 }}>
                <button onClick={onBack} style={btnSecondary}>← Sites</button>
                <h2 style={{ ...sectionTitle, margin: 0, flex: 1 }}>
                    {reviewQueueMode ? "Review Queue" : "Work Items"} {siteFilter && <span style={{ fontWeight: 400, color: "#64748b" }}>— {siteFilter.domain}</span>}
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

            {reviewQueueMode && (
                <div style={{ fontSize: 13, color: "#64748b", marginBottom: 16 }}>
                    {total} item{total === 1 ? "" : "s"} awaiting your review — LLM-audit findings that
                    RFC_056's circuit breaker parked instead of auto-dispatching (bugs_open/428). Nothing
                    here is acted on until you release it.
                </div>
            )}

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
                    <option value="unresolved">Unresolved ({statusCounts["unresolved"] || 0})</option>
                    <option value="triaged">Triaged ({statusCounts["triaged"] || 0})</option>
                    <option value="claimed">Claimed ({statusCounts["claimed"] || 0})</option>
                    <option value="failed">Failed ({statusCounts["failed"] || 0})</option>
                    <option value="blocked">Blocked ({statusCounts["blocked"] || 0})</option>
                    {/* Deferred was missing entirely until bugs_open/428 — a
                        reviewer had no dropdown route to RFC_056's parked
                        verdicts (or any other deferred row) shorter than a raw
                        SQL query. */}
                    <option value="deferred">Deferred ({statusCounts["deferred"] || 0})</option>
                    <option value="complete">Complete ({statusCounts["complete"] || 0})</option>
                </select>
                <select value={typeFilter} onChange={e => setTypeFilter(e.target.value)} style={selectStyle}>
                    <option value="">All types</option>
                    {itemTypes.map(t => <option key={t} value={t}>{t} ({serverTypeCounts[t]})</option>)}
                </select>
                <label style={{ display: "flex", alignItems: "center", gap: 6, fontSize: 13, color: "#334155" }} title="RFC_056: an LLM-audit seat's finding, recorded but not auto-dispatched. Release only after reviewing the specific row.">
                    <input type="checkbox" checked={recordVerdictsOnly} onChange={e => setRecordVerdictsOnly(e.target.checked)} />
                    Record verdicts only
                </label>
                <select value={pipelineFilter} onChange={e => setPipelineFilter(e.target.value)} style={selectStyle}>
                    <option value="build">build pipeline</option>
                    <option value="content">content pipeline</option>
                    <option value="all">all pipelines</option>
                </select>
                <span style={{ fontSize: 13, color: "#64748b" }}>
                    {/* Always show the total alongside what is on screen: showing
                        only the loaded count is how a 208-item backlog read as 0. */}
                    {truncated
                        ? `showing ${items.length} of ${total}`
                        : `${items.length} item${items.length === 1 ? "" : "s"}`}
                </span>
                {truncated && (
                    <button onClick={loadMore} disabled={loadingMore} style={{
                        ...btnSecondary, fontSize: 12, padding: "5px 12px",
                    }}>
                        {loadingMore ? "Loading…" : `Load more (${total - items.length} remaining)`}
                    </button>
                )}
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
                                                selectedItem.item_type === "needs_section_data" ?
                                                    `Provide Data for "${(selectedItem.spec?.section_name || "section").replace(/-/g, " ")}"` :
                                                    "Item Data"}
                                        {selectedItem.spec?.spec_aspect && (
                                            <span style={{ fontWeight: 400, color: "#94a3b8" }}> — will update spec '{selectedItem.spec.spec_aspect}'</span>
                                        )}
                                    </div>
                                    {selectedItem.item_type === "needs_section_data" && selectedItem.spec?.missing && (
                                        <div style={{ fontSize: 12, color: "#64748b", marginBottom: 8, lineHeight: 1.5 }}>
                                            {selectedItem.spec.missing.map((m, i) => (
                                                <div key={i} style={{ marginBottom: 4 }}>
                                                    <strong style={{ color: "#475569" }}>
                                                        {(m.field || "").replace(/_/g, " ").replace(/\b\w/g, c => c.toUpperCase())}
                                                    </strong>
                                                    {m.reason && <span> — {m.reason}</span>}
                                                </div>
                                            ))}
                                        </div>
                                    )}
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
                                {["needs_human_review", "unresolved", "failed", "blocked"].includes(selectedItem.status) && (
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
                                {/* RFC_056 record verdict (bugs_open/428): the ONLY route this
                                    row can leave 'deferred' by is a human reviewing it here and
                                    choosing to release it — nothing else may dispatch it.
                                    The button is gated on routed_handler AND routed_status being
                                    present, because HandleReleaseRecordVerdict requires both in
                                    its WHERE clause. Some record verdicts deliberately carry
                                    neither — the recommended_type_gap rows written by
                                    validate_site_plan's strategy reconciliation have no automatic
                                    repair to release, so the endpoint refuses them by
                                    construction. Without this gate they would show a green button
                                    that always 404s, which is a worse lie than no button. */}
                                {selectedItem.status === "deferred" && selectedItem.spec?.filing_mode === "record" && (
                                    selectedItem.spec?.routed_handler && selectedItem.spec?.routed_status ? (
                                        <button onClick={() => handleRelease(selectedItem.id)} disabled={actionLoading} style={{
                                            ...btnPrimary, background: "#059669",
                                        }} title={`Would route to: ${selectedItem.spec.routed_handler} / ${selectedItem.spec.routed_status}`}>
                                            Review &amp; Release → {selectedItem.spec.routed_handler}
                                        </button>
                                    ) : (
                                        <span style={{ fontSize: 12, color: "#64748b", alignSelf: "center", fontStyle: "italic" }}
                                            title="This verdict is a record for a person to read. It names no handler and no target status, so there is nothing for the release endpoint to dispatch — read spec.omission_class and spec.builder_needed and decide what to do about it yourself.">
                                            record only — no automatic route
                                        </span>
                                    )
                                )}
                            </div>

                            {/* Request changes: file a critique WITHOUT approving or resolving.
                                The item stays open (a checkpoint keeps gating delivery); the
                                critique becomes an owner_critique work item routed to the
                                working threads by the dispatcher. */}
                            {selectedItem.status === "needs_human_review" && (
                                <div style={{ borderTop: "1px solid #e2e8f0", paddingTop: 16, marginTop: 16 }}>
                                    <div style={{ fontSize: 13, fontWeight: 600, color: "#9a3412", marginBottom: 8 }}>
                                        Request changes
                                    </div>
                                    <textarea
                                        value={critiqueText}
                                        onChange={e => setCritiqueText(e.target.value)}
                                        placeholder="Your critique of this site, in your own words — what's wrong, what's missing, what should change. It is filed as work and routed to the threads; this item stays open and delivery stays gated."
                                        rows={5}
                                        style={{ ...formInputStyle, width: "100%", resize: "vertical", lineHeight: 1.5 }}
                                    />
                                    <button
                                        onClick={() => handleRequestChanges(selectedItem.id)}
                                        disabled={actionLoading || !critiqueText.trim()}
                                        style={{ ...btnPrimary, background: "#c2410c", marginTop: 8 }}
                                    >
                                        Request Changes
                                    </button>
                                </div>
                            )}
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
    const [editBrief, setEditBrief] = useState(null);
    const [pageSpec, setPageSpec] = useState(null);
    const [editPageSpec, setEditPageSpec] = useState(null);
    const [editMode, setEditMode] = useState("structured"); // structured | html | brief
    const [loading, setLoading] = useState(true);
    const [actionLoading, setActionLoading] = useState(false);
    const [message, setMessage] = useState("");
    const [divergenceCounts, setDivergenceCounts] = useState({});

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

    // How often has each slot been silently overwritten by a rebuild? An
    // unlocked component with a non-zero count is exactly the state that is
    // about to lose someone's hand edit again — the rebuild keeps a locked
    // copy but discards an unlocked one, filing page_divergence_overwritten
    // per component per rebuild, visible until now only in the work-item queue.
    const loadDivergenceCounts = useCallback(async () => {
        try {
            const data = await apiFetch(
                `/work-items?pipeline=all&status=all&site_id=${siteId}&item_type=page_divergence_overwritten&limit=${MAX_PAGE_SIZE}`,
                token);
            const counts = {};
            (data.items || []).forEach(item => {
                const s = item.spec || {};
                if (!s.page_id || !s.slot_name) return;
                const exact = `${s.page_id}|${s.slot_name}|${s.position}`;
                const bySlot = `${s.page_id}|${s.slot_name}`;
                counts[exact] = (counts[exact] || 0) + 1;
                counts[bySlot] = (counts[bySlot] || 0) + 1;
            });
            setDivergenceCounts(counts);
        } catch (err) {
            console.error(err);
            setDivergenceCounts({});
        }
    }, [token, siteId]);

    useEffect(() => { loadDivergenceCounts(); }, [loadDivergenceCounts]);

    const divergenceCountFor = (comp) => {
        if (!selectedPage) return 0;
        const exact = divergenceCounts[`${selectedPage.id}|${comp.slot_name}|${comp.position}`];
        if (exact !== undefined) return exact;
        return divergenceCounts[`${selectedPage.id}|${comp.slot_name}`] || 0;
    };

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
            setPageSpec(data.page?.page_spec || null);
        } catch (err) {
            console.error(err);
            setComponents([]);
            setSuppressedSections([]);
            setPageSpec(null);
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
        const hasBrief = comp.content_brief && typeof comp.content_brief === "object" && Object.keys(comp.content_brief as Record<string, unknown>).length > 0;
        setEditMode(hasStructured ? "structured" : hasBrief ? "brief" : "html");
        if (hasStructured) {
            setEditData(JSON.parse(JSON.stringify(comp.content_data)));
        } else {
            setEditData(null);
        }
        setEditBrief(hasBrief ? JSON.parse(JSON.stringify(comp.content_brief)) : {
            purpose: (pageSpec && typeof pageSpec === "object" ? (pageSpec as Record<string, unknown>).purpose || "" : "") as string,
            tone_direction: "",
            section_guidance: comp.slot_name ? `${comp.slot_name} section` : "",
        });
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

    const handleRegenerate = async (comp) => {
        if (!confirm(`Regenerate "${comp.slot_name}" using the updated brief? Current content will be replaced by the content writer.`)) return;
        setActionLoading(true);
        try {
            await apiFetch(
                `/sites/${siteId}/pages/${selectedPage.name}/components/${comp.id}/regenerate`,
                token, { method: "POST", body: JSON.stringify({ brief: editBrief }) }
            );
            setMessage(`Regeneration queued for "${comp.slot_name}" — content writer will rewrite it`);
            setSelectedComponent(null);
            loadComponents(selectedPage.name);
        } catch (err) { setMessage("Regenerate failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const handleRegeneratePage = async () => {
        if (!confirm(`Regenerate all unlocked sections on ${selectedPage.name}? Locked sections will be skipped.`)) return;
        setActionLoading(true);
        try {
            const body: Record<string, unknown> = {};
            if (editPageSpec) body.page_spec = editPageSpec;
            const result = await apiFetch(
                `/sites/${siteId}/pages/${selectedPage.name}/regenerate`,
                token, { method: "POST", body: JSON.stringify(body) }
            );
            setMessage(`Page regeneration: ${result.items_created} sections queued${result.items_skipped > 0 ? `, ${result.items_skipped} locked sections skipped` : ""}`);
            setEditPageSpec(null);
            loadComponents(selectedPage.name);
        } catch (err) { setMessage("Regenerate page failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const handleSavePageSpec = async () => {
        if (!editPageSpec) return;
        setActionLoading(true);
        try {
            await apiFetch(
                `/sites/${siteId}/pages/${selectedPage.name}/spec`,
                token, { method: "PATCH", body: JSON.stringify({ page_spec: editPageSpec }) }
            );
            setMessage("Page spec updated");
            setPageSpec(editPageSpec);
            setEditPageSpec(null);
            loadComponents(selectedPage.name);
        } catch (err) { setMessage("Save page spec failed: " + err.message); }
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
                                    {viewMode !== "site-wide" && (
                                        <button onClick={() => setEditMode("brief")} style={{
                                            ...btnSecondary, fontSize: 12, padding: "4px 12px",
                                            background: editMode === "brief" ? "#dbeafe" : "#fff",
                                            color: editMode === "brief" ? "#1e40af" : "#64748b",
                                        }}>Brief</button>
                                    )}
                                </div>

                                {/* Edit form */}
                                <div style={{
                                    background: "#f8fafc", border: "1px solid #e2e8f0", borderRadius: 8,
                                    padding: 16, marginBottom: 16, maxHeight: "50vh", overflowY: "auto",
                                }}>
                                    {editMode === "structured" && editData ? (
                                        <EditableReviewForm reviewData={editData} onChange={setEditData} />
                                    ) : editMode === "brief" ? (
                                        <div>
                                            <div style={{ fontSize: 12, color: "#64748b", marginBottom: 12 }}>
                                                Edit the instructions that guide how this section's content is generated. Click "Regenerate" to rewrite the section using these instructions.
                                            </div>
                                            {editBrief && Object.entries(editBrief as Record<string, unknown>).map(([key, value]) => (
                                                <div key={key} style={{ marginBottom: 10 }}>
                                                    <label style={formLabelStyle}>{key.replace(/_/g, " ")}</label>
                                                    <textarea
                                                        value={String(value ?? "")}
                                                        onChange={e => setEditBrief({ ...(editBrief as Record<string, unknown>), [key]: e.target.value })}
                                                        rows={key === "purpose" || key === "section_guidance" ? 3 : 2}
                                                        style={{
                                                            ...formInputStyle, minHeight: 50, resize: "vertical",
                                                            fontFamily: "'IBM Plex Sans', sans-serif",
                                                        }}
                                                    />
                                                </div>
                                            ))}
                                        </div>
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
                                    {editMode === "brief" ? (
                                        <button onClick={() => handleRegenerate(selectedComponent)} disabled={actionLoading} style={{
                                            ...btnPrimary, background: "#1e40af",
                                        }}>
                                            Regenerate
                                        </button>
                                    ) : (
                                        <button onClick={() => handleSaveComponent(selectedComponent)} disabled={actionLoading} style={{
                                            ...btnPrimary, background: "#059669",
                                        }}>
                                            Save & Deploy
                                        </button>
                                    )}
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
                                <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
                                    <div style={{ fontSize: 14, fontWeight: 600, color: "#475569" }}>
                                        {selectedPage.name} — {components.length} sections
                                    </div>
                                    <button onClick={handleRegeneratePage} disabled={actionLoading} style={{
                                        ...btnSecondary, fontSize: 11, padding: "4px 12px", color: "#1e40af",
                                    }}>
                                        Regenerate Page
                                    </button>
                                </div>

                                {/* Page spec / purpose */}
                                {(pageSpec || editPageSpec) && (
                                    <div style={{
                                        background: "#fffbeb", border: "1px solid #fde68a", borderRadius: 8,
                                        padding: 12, marginBottom: 12, fontSize: 12,
                                    }}>
                                        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "start", marginBottom: 4 }}>
                                            <span style={{ fontWeight: 600, color: "#92400e" }}>Page Purpose</span>
                                            {!editPageSpec ? (
                                                <button onClick={() => setEditPageSpec(JSON.parse(JSON.stringify(pageSpec)))} style={{
                                                    ...btnSecondary, fontSize: 10, padding: "2px 8px",
                                                }}>Edit</button>
                                            ) : (
                                                <div style={{ display: "flex", gap: 4 }}>
                                                    <button onClick={handleSavePageSpec} disabled={actionLoading} style={{
                                                        ...btnSecondary, fontSize: 10, padding: "2px 8px", color: "#059669",
                                                    }}>Save</button>
                                                    <button onClick={() => setEditPageSpec(null)} style={{
                                                        ...btnSecondary, fontSize: 10, padding: "2px 8px",
                                                    }}>Cancel</button>
                                                </div>
                                            )}
                                        </div>
                                        {editPageSpec ? (
                                            <textarea
                                                value={typeof editPageSpec === "object" ? (editPageSpec as Record<string, unknown>).purpose as string || JSON.stringify(editPageSpec, null, 2) : String(editPageSpec)}
                                                onChange={e => {
                                                    try {
                                                        setEditPageSpec(JSON.parse(e.target.value));
                                                    } catch {
                                                        setEditPageSpec({ ...(editPageSpec as Record<string, unknown>), purpose: e.target.value });
                                                    }
                                                }}
                                                rows={3}
                                                style={{
                                                    width: "100%", padding: 8, fontSize: 12, border: "1px solid #fde68a",
                                                    borderRadius: 4, resize: "vertical", boxSizing: "border-box",
                                                }}
                                            />
                                        ) : (
                                            <div style={{ color: "#78716c", lineHeight: 1.5 }}>
                                                {typeof pageSpec === "object" ? (pageSpec as Record<string, unknown>).purpose as string || JSON.stringify(pageSpec) : String(pageSpec || "No purpose set")}
                                            </div>
                                        )}
                                    </div>
                                )}
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
                                                    {divergenceCountFor(comp) > 0 && (
                                                        <span
                                                            title={comp.locked
                                                                ? "Past rebuilds overwrote hand edits here; this component is now locked, so its bytes are kept."
                                                                : "A rebuild has overwritten hand edits on this component before, and it is NOT locked — the next rebuild will overwrite it again. Lock it to keep hand-patched bytes."}
                                                            style={{
                                                                fontSize: 11, padding: "1px 6px", borderRadius: 4, fontWeight: 700,
                                                                color: comp.locked ? "#92400e" : "#991b1b",
                                                                background: comp.locked ? "#fef3c7" : "#fee2e2",
                                                            }}>
                                                            ⚠ overwritten ×{divergenceCountFor(comp)}{!comp.locked && " — unlocked"}
                                                        </span>
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
// Two kinds of aspect share this editor and the same save gesture, and they are
// not the same kind of thing. evidence_base is a CONTROL: its banned_claims and
// facts are checked against built and served pages by the claims lanes. The
// prose aspects are PROMPT TEXT — instructions to a writer, enforced by nothing.
// "Never say X" typed into content_direction is a wish; the same sentence as a
// banned_claims pattern is a control. The labels below exist so an operator can
// tell which one they are editing.
const ENFORCED_ASPECTS = new Set(["evidence_base"]);
const ADVISORY_ASPECTS = new Set([
    "content_direction", "briefing", "strategy", "mission_brief",
    "roadmap_brief", "writer_block", "vertical_research", "domain_research",
]);
const PROHIBITION_RE = /\b(never|must not|mustn't|do not say|don't say|unacceptable|forbidden|banned)\b/i;

function AspectChip({ aspect }) {
    if (ENFORCED_ASPECTS.has(aspect)) {
        return (
            <span title="banned_claims and facts in this register are checked against built and served pages"
                  style={{ fontSize: 10, fontWeight: 700, color: "#065f46", background: "#d1fae5", padding: "1px 8px", borderRadius: 4, letterSpacing: 0.5 }}>
                ENFORCED
            </span>
        );
    }
    if (ADVISORY_ASPECTS.has(aspect)) {
        return (
            <span title="Prompt text: instructions a writer reads. Nothing checks the output against it — prohibitions here are wishes, not controls."
                  style={{ fontSize: 10, fontWeight: 600, color: "#64748b", background: "#f1f5f9", padding: "1px 8px", borderRadius: 4, letterSpacing: 0.5 }}>
                advisory — prompt text
            </span>
        );
    }
    return null;
}

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

    const handleSaveSpec = async (confirmEmpty = false) => {
        if (!selectedSpec || !editData) return;
        setActionLoading(true);
        try {
            const result = await apiFetch(`/sites/${siteId}/specs/${selectedSpec.aspect}`, token, {
                method: "PATCH",
                body: JSON.stringify({ data: editData, confirm_empty: confirmEmpty === true }),
            });
            let msg = `Spec '${selectedSpec.aspect}' updated`;
            if (selectedSpec.aspect === "evidence_base") {
                // The counts come from the server re-parsing what it stored — the
                // one signal a wrong-shape save cannot fake. A zero here means
                // claims checking found nothing to enforce.
                msg = `evidence_base saved — ${result.banned_claims_count} banned claims, ${result.facts_count} facts`;
                if (result.banned_claims_count === 0 && result.facts_count === 0) {
                    msg += " — REGISTER IS EMPTY: claims checking is now OFF for this site";
                }
            } else if (ADVISORY_ASPECTS.has(selectedSpec.aspect) && PROHIBITION_RE.test(JSON.stringify(editData))) {
                msg += ". This text contains a prohibition, and this aspect is prompt text — nothing enforces it. To make it a control, add a matching pattern under evidence_base → banned_claims.";
            }
            setMessage(msg);
            setSelectedSpec(null);
            setEditData(null);
            loadSpecs();
        } catch (err) {
            let refusal = null;
            try { refusal = JSON.parse(err.message); } catch { /* not JSON */ }
            if (refusal && refusal.code === "EMPTY_EVIDENCE_BASE") {
                if (confirm(
                    `This save would replace a register holding ${refusal.current_facts_count} facts and ` +
                    `${refusal.current_banned_claims_count} banned claims with one that parses to NOTHING — ` +
                    `claims checking for this site would silently stop.\n\n` +
                    `Usually this means a key is misspelt or the JSON is nested one level too deep, ` +
                    `not that you meant to empty it.\n\nReplace it anyway?`)) {
                    setActionLoading(false);
                    return handleSaveSpec(true);
                }
                setMessage("Save cancelled — the existing register is untouched.");
            } else {
                setMessage("Save failed: " + err.message);
            }
        }
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
                        <h3 style={{ margin: 0, fontSize: 16, color: "#0f172a", display: "flex", alignItems: "center", gap: 8 }}>
                            Edit: {selectedSpec.aspect}
                            <AspectChip aspect={selectedSpec.aspect} />
                            {selectedSpec.pinned && <span style={{ fontSize: 12, color: "#7c3aed" }}>🔒 Pinned</span>}
                        </h3>
                        <span onClick={() => { setSelectedSpec(null); setEditData(null); }} style={{ cursor: "pointer", color: "#94a3b8", fontSize: 18 }}>✕</span>
                    </div>

                    {ADVISORY_ASPECTS.has(selectedSpec.aspect) && (
                        <div style={{ background: "#f8fafc", border: "1px solid #e2e8f0", borderRadius: 6, padding: "8px 12px", fontSize: 12, color: "#64748b", marginBottom: 12 }}>
                            This aspect is prompt text — a writer reads it, nothing checks the output against it.
                            A prohibition ("never say X") only becomes a control as a pattern under
                            <b> evidence_base → banned_claims</b>; add it there in the same edit.
                        </div>
                    )}

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
                        <button onClick={() => handleSaveSpec()} disabled={actionLoading} style={btnPrimary}>
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
                                        <AspectChip aspect={spec.aspect} />
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

// ── Builds view ──────────────────────────────────────────────────────────────
// A website build on this platform is a chain of site_work_items, not an
// orchestration: orchestration_states stores no step sequence (execution_path
// is empty on every row) and a site's tagged orchestrations are mostly its
// periodic sweeps. So the timeline below renders an EXPLICIT stage vocabulary
// — the 082_submit_domain_unified.sh cascade, as measured live — rather than
// "whatever rows the site has", which would bury the build in repeat-noise.
const BUILD_STAGES = [
    ["needs_domain_research", "Domain research"],
    ["needs_vertical_research", "Vertical research"],
    ["needs_strategy", "Strategy"],
    ["needs_briefing", "Briefing"],
    ["needs_site_plan", "Site plan"],
    ["needs_composition", "Composition"],
    ["needs_design", "Design"],
    ["needs_content_planning", "Content planning"],
    ["needs_page", "Page builds"],
    ["needs_content_page", "Page content"],
    ["needs_imagery", "Imagery"],
    ["needs_rerender", "Re-render"],
];
const BUILD_STAGE_SET = new Set(BUILD_STAGES.map(([t]) => t));

function fmtDuration(startIso, endIso) {
    if (!startIso || !endIso) return "";
    const ms = new Date(endIso).getTime() - new Date(startIso).getTime();
    if (ms < 0) return "";
    const m = Math.round(ms / 60000);
    if (m < 1) return "<1m";
    if (m < 120) return `${m}m`;
    const h = Math.floor(m / 60);
    if (h < 48) return `${h}h ${m % 60}m`;
    return `${Math.round(h / 24)}d`;
}

// A workflow is actionable only while it is NOT terminal: that is precisely when
// the detail view offers Resume and Terminate. The list uses the same test, so
// "no buttons anywhere" is never a mystery — the rows that have them are pinned
// to the top of the list and the rest say why they do not.
const TERMINAL_WORKFLOW_STATUSES = ["COMPLETED", "FAILED", "CANCELLED"];
const isTerminalWorkflow = (wf) =>
    TERMINAL_WORKFLOW_STATUSES.includes(String(wf?.status || "").toUpperCase());

// How many finished orchestrations to show before the fold. They are periodic
// sweeps: the newest few are worth a glance, the older ones are noise, and the
// site accumulates them indefinitely.
const WORKFLOW_PREVIEW = 6;

// The server-side page size for the orchestration list. Named because the label
// reads it: a full window means "there are at least this many", not "this many".
const WORKFLOW_LIMIT = 50;

function BuildsView({ token, siteId, siteDomain, onBack }) {
    const [items, setItems] = useState([]);
    const [typeCounts, setTypeCounts] = useState({});
    const [truncated, setTruncated] = useState(false);
    const [workflows, setWorkflows] = useState([]);
    const [selectedWorkflow, setSelectedWorkflow] = useState(null);
    const [expandedStage, setExpandedStage] = useState(null);
    const [showAllWorkflows, setShowAllWorkflows] = useState(false);
    const [loading, setLoading] = useState(true);
    const [actionLoading, setActionLoading] = useState(false);
    const [message, setMessage] = useState("");

    const loadAll = useCallback(async () => {
        try {
            const data = await apiFetch(`/work-items?pipeline=all&status=all&site_id=${siteId}&limit=${MAX_PAGE_SIZE}`, token);
            const rows = data.items || [];
            // The window is newest-first and the build-stage rows are usually the
            // OLDEST rows a site has — on a truncated window fetch each stage type
            // directly, or a busy site's timeline silently loses its build.
            if (data.truncated) {
                const perStage = await Promise.all(BUILD_STAGES.map(([type]) =>
                    apiFetch(`/work-items?pipeline=all&status=all&site_id=${siteId}&item_type=${type}&limit=200`, token)
                        .then(r => r.items || []).catch(() => [])
                ));
                const seen = new Set(rows.map(i => i.id));
                perStage.flat().forEach(i => { if (!seen.has(i.id)) { rows.push(i); seen.add(i.id); } });
            }
            setItems(rows);
            setTypeCounts(data.type_counts || {});
            setTruncated(!!data.truncated);
            const wf = await apiFetch(`/workflows?site_id=${siteId}&limit=${WORKFLOW_LIMIT}`, token);
            setWorkflows(wf.workflows || []);
        } catch (err) {
            setMessage("Load failed: " + err.message);
        } finally {
            setLoading(false);
        }
    }, [token, siteId]);

    useEffect(() => { loadAll(); }, [loadAll]);
    // Builds move on minute timescales; poll rather than stream.
    useEffect(() => {
        const t = setInterval(() => { if (!document.hidden) loadAll(); }, 10000);
        return () => clearInterval(t);
    }, [loadAll]);

    const stageGroups = BUILD_STAGES.map(([type, label]) => {
        const rows = items.filter(i => i.item_type === type)
            .sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
        if (rows.length === 0) return null;
        const statuses = rows.map(r => r.status);
        const allComplete = statuses.every(s => s === "complete");
        const anyFailed = statuses.some(s => s === "failed");
        const rollup = anyFailed ? "failed" : allComplete ? "complete" : (statuses.find(s => s !== "complete") || "detected");
        const started = rows[0].created_at;
        const completedTimes = rows.map(r => r.completed_at).filter(Boolean).sort();
        const finished = allComplete && completedTimes.length === rows.length ? completedTimes[completedTimes.length - 1] : null;
        return { type, label, rows, rollup, started, finished };
    }).filter(Boolean);

    const otherTypes = Object.entries(typeCounts)
        .filter(([t]) => !BUILD_STAGE_SET.has(t))
        .sort((a, b) => Number(b[1]) - Number(a[1]));
    const divergenceCount = Number(typeCounts["page_divergence_overwritten"] || 0);

    // ⚠ correlation_id is NOT unique: one correlation is a TREE of orchestration
    // rows (parent plus sub-orchestrations), and /admin/workflows returns them
    // raw — measured 2026-08-26, up to 27 rows share one correlation on a single
    // site. So the list is grouped by correlation before anything else. That
    // fixes a real defect (React was keyed on a value that repeats, on a list
    // that re-polls every 10s), and it is also the honest unit: every sibling row
    // opens the SAME detail, because the detail endpoint pins root-first.
    const workflowGroups = (() => {
        const byCorrelation = new Map();
        for (const wf of workflows) {
            const key = wf.correlation_id;
            const g = byCorrelation.get(key);
            if (!g) {
                byCorrelation.set(key, {
                    correlation_id: key, rows: [wf], created_at: wf.created_at,
                    current_step: wf.current_step, has_step_error: !!wf.has_step_error,
                });
                continue;
            }
            g.rows.push(wf);
            g.has_step_error = g.has_step_error || !!wf.has_step_error;
            // The list is newest-first, so the first row seen is the newest; keep
            // its timestamp and step as the group's face.
        }
        return [...byCorrelation.values()].map(g => {
            // A group is live if ANY sibling is: stopping "the workflow" means
            // stopping what is still running under it, which is exactly what the
            // terminate endpoint now scopes itself to.
            const liveRow = g.rows.find(r => !isTerminalWorkflow(r));
            const anyFailed = g.rows.some(r => String(r.status || "").toUpperCase() === "FAILED");
            return {
                ...g,
                count: g.rows.length,
                status: liveRow ? liveRow.status : (anyFailed ? "failed" : "completed"),
                live: !!liveRow,
            };
        });
    })();

    // Anything still running is pinned and never hidden — those are the only ones
    // that can be Resumed or Terminated, and burying one behind a fold is how an
    // operator concludes the button does not exist.
    const liveWorkflows = workflowGroups.filter(g => g.live);
    const finishedWorkflows = workflowGroups.filter(g => !g.live);
    const visibleWorkflows = showAllWorkflows
        ? [...liveWorkflows, ...finishedWorkflows]
        : [...liveWorkflows, ...finishedWorkflows.slice(0, WORKFLOW_PREVIEW)];
    const hiddenWorkflowCount = showAllWorkflows ? 0 : Math.max(0, finishedWorkflows.length - WORKFLOW_PREVIEW);

    const openWorkflow = async (wf) => {
        try {
            const detail = await apiFetch(`/workflows/${wf.correlation_id}`, token);
            setSelectedWorkflow(detail);
        } catch (err) { setMessage("Load workflow failed: " + err.message); }
    };

    const handleWorkflowAction = async (action) => {
        if (!selectedWorkflow) return;
        const verb = action === "terminate" ? "Terminate" : "Resume";
        // Termination is a state-row label, not an interrupt: it marks the row
        // FAILED but does not stop a step already executing in a chassis pod
        // (council 45b3c93f, guardian seat). Say so before the click.
        const caveat = action === "terminate"
            ? "\n\nThis marks every still-running orchestration under this correlation FAILED — finished rows are left untouched. It does NOT interrupt a step already executing — a running process finishes or fails on its own."
            : "";
        if (!confirm(`${verb} workflow ${selectedWorkflow.correlation_id}?${caveat}`)) return;
        setActionLoading(true);
        try {
            await apiFetch(`/workflows/${selectedWorkflow.correlation_id}/resume`, token, {
                method: "POST",
                body: JSON.stringify({ action }),
            });
            setMessage(`Workflow ${action} sent for ${selectedWorkflow.correlation_id}`);
            setSelectedWorkflow(null);
            loadAll();
        } catch (err) { setMessage(`${verb} failed: ` + err.message); }
        finally { setActionLoading(false); }
    };

    const stepErrors = selectedWorkflow
        ? Object.entries(selectedWorkflow.collected_data || {}).filter(([k]) => k.startsWith("__step_error"))
        : [];
    const reconstructedSteps = selectedWorkflow
        ? Object.keys(selectedWorkflow.collected_data || {}).filter(k => !k.startsWith("__"))
        : [];
    const workflowTerminal = selectedWorkflow && isTerminalWorkflow(selectedWorkflow);

    return (
        <div>
            <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 20 }}>
                <button onClick={onBack} style={btnSecondary}>← Sites</button>
                <h2 style={{ ...sectionTitle, margin: 0, flex: 1 }}>
                    Builds <span style={{ fontWeight: 400, color: "#64748b" }}>— {siteDomain}</span>
                </h2>
            </div>

            {message && (
                <div style={{
                    background: "#fef3c7", color: "#92400e", padding: "8px 14px", borderRadius: 6,
                    fontSize: 13, marginBottom: 12, display: "flex", justifyContent: "space-between",
                }}>
                    {message}
                    <span onClick={() => setMessage("")} style={{ cursor: "pointer" }}>✕</span>
                </div>
            )}

            {loading ? (
                <div style={{ textAlign: "center", padding: 40, color: "#94a3b8" }}>Loading build history…</div>
            ) : selectedWorkflow ? (
                /* ── Workflow detail ── */
                <div style={{ background: "#fff", border: "1px solid #e2e8f0", borderRadius: 10, padding: 20 }}>
                    <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 12 }}>
                        <h3 style={{ margin: 0, fontSize: 15, color: "#0f172a", fontFamily: "monospace" }}>
                            {selectedWorkflow.correlation_id}
                        </h3>
                        <span onClick={() => setSelectedWorkflow(null)} style={{ cursor: "pointer", color: "#94a3b8", fontSize: 18 }}>✕</span>
                    </div>
                    <div style={{ display: "flex", gap: 16, alignItems: "center", marginBottom: 12, fontSize: 13, flexWrap: "wrap" }}>
                        <Badge status={String(selectedWorkflow.status || "").toLowerCase() === "completed" ? "complete" : String(selectedWorkflow.status || "").toLowerCase()} />
                        <span style={{ color: "#64748b" }}>current step: <b>{selectedWorkflow.current_step || "—"}</b></span>
                        <span style={{ color: "#94a3b8" }}>
                            {selectedWorkflow.created_at && new Date(selectedWorkflow.created_at).toLocaleString()}
                        </span>
                    </div>

                    {/* The status column lies by omission: a step can fail, have its
                        output discarded, and the row still read COMPLETED with error
                        NULL (bugs_open/099). The __step_error entries are the truth. */}
                    {stepErrors.length > 0 && (
                        <div style={{ background: "#fee2e2", border: "1px solid #fca5a5", borderRadius: 8, padding: 12, marginBottom: 12 }}>
                            <div style={{ fontSize: 13, fontWeight: 700, color: "#991b1b", marginBottom: 6 }}>
                                ⚠ Step errors — a step failed even if the status above reads complete
                            </div>
                            {stepErrors.map(([k, v]) => (
                                <pre key={k} style={{ fontSize: 11, color: "#7f1d1d", whiteSpace: "pre-wrap", margin: "4px 0", maxHeight: 160, overflow: "auto" }}>
                                    {k}: {typeof v === "string" ? v : JSON.stringify(v, null, 2)}
                                </pre>
                            ))}
                        </div>
                    )}
                    {selectedWorkflow.error && (
                        <div style={{ background: "#fee2e2", borderRadius: 8, padding: 10, marginBottom: 12, fontSize: 12, color: "#991b1b" }}>
                            error: {selectedWorkflow.error}
                        </div>
                    )}

                    <div style={{ fontSize: 12, color: "#94a3b8", marginBottom: 6 }}>
                        Steps (reconstructed from outputs — the platform stores no step sequence):
                    </div>
                    <div style={{ display: "flex", flexWrap: "wrap", gap: 6, marginBottom: 12 }}>
                        {reconstructedSteps.length === 0 ? (
                            <span style={{ fontSize: 12, color: "#94a3b8" }}>no step outputs recorded</span>
                        ) : reconstructedSteps.map(s => (
                            <span key={s} style={{ fontSize: 11, background: "#f1f5f9", color: "#475569", padding: "2px 8px", borderRadius: 4, fontFamily: "monospace" }}>{s}</span>
                        ))}
                    </div>
                    {Array.isArray(selectedWorkflow.awaited_steps) && selectedWorkflow.awaited_steps.length > 0 && (
                        <div style={{ fontSize: 12, color: "#92400e", marginBottom: 12 }}>
                            awaiting: {selectedWorkflow.awaited_steps.join(", ")}
                        </div>
                    )}

                    {[["initial_request_data", "Initial request"], ["collected_data", "Collected data"], ["final_result", "Final result"]].map(([key, label]) => (
                        <details key={key} style={{ marginBottom: 8 }}>
                            <summary style={{ fontSize: 12, color: "#475569", cursor: "pointer" }}>{label}</summary>
                            <pre style={{
                                fontSize: 11, background: "#f8fafc", border: "1px solid #e2e8f0", borderRadius: 6,
                                padding: 10, maxHeight: 300, overflow: "auto", whiteSpace: "pre-wrap",
                            }}>{JSON.stringify(selectedWorkflow[key], null, 2)}</pre>
                        </details>
                    ))}

                    {!workflowTerminal && (
                        <div style={{ display: "flex", gap: 8, marginTop: 12 }}>
                            <button onClick={() => handleWorkflowAction("resume")} disabled={actionLoading} style={btnPrimary}>
                                Resume
                            </button>
                            <button onClick={() => handleWorkflowAction("terminate")} disabled={actionLoading} style={{
                                ...btnSecondary, color: "#991b1b", borderColor: "#fca5a5",
                            }}>
                                Terminate
                            </button>
                        </div>
                    )}
                </div>
            ) : (
                <div>
                    {/* ── Stage timeline ── */}
                    {stageGroups.length === 0 ? (
                        <div style={{ textAlign: "center", padding: 30, color: "#94a3b8" }}>
                            No build-stage work items recorded for this site
                            {truncated && " (window truncated — older rows may exist)"}
                        </div>
                    ) : (
                        <div style={{ background: "#fff", border: "1px solid #e2e8f0", borderRadius: 10, padding: "8px 0", marginBottom: 16 }}>
                            {stageGroups.map((sg, idx) => (
                                <div key={sg.type}>
                                    <div onClick={() => setExpandedStage(expandedStage === sg.type ? null : sg.type)} style={{
                                        display: "flex", alignItems: "center", gap: 12, padding: "10px 18px",
                                        cursor: "pointer",
                                        borderTop: idx > 0 ? "1px solid #f1f5f9" : "none",
                                    }}>
                                        <span style={{ fontSize: 12, color: "#94a3b8", width: 18, textAlign: "right" }}>{idx + 1}</span>
                                        <span style={{ fontSize: 13, fontWeight: 600, color: "#0f172a", width: 170 }}>
                                            {sg.label}{sg.rows.length > 1 && <span style={{ color: "#94a3b8", fontWeight: 400 }}> ×{sg.rows.length}</span>}
                                        </span>
                                        <Badge status={sg.rollup} />
                                        <span style={{ fontSize: 12, color: "#64748b", flex: 1 }}>
                                            {new Date(sg.started).toLocaleString()}
                                            {sg.finished && <> → {new Date(sg.finished).toLocaleTimeString()} <b>({fmtDuration(sg.started, sg.finished)})</b></>}
                                        </span>
                                        <span style={{ fontSize: 12, color: "#94a3b8" }}>{expandedStage === sg.type ? "▾" : "▸"}</span>
                                    </div>
                                    {expandedStage === sg.type && sg.rows.map(row => (
                                        <div key={row.id} style={{ padding: "6px 18px 6px 48px", fontSize: 12, color: "#475569", background: "#f8fafc" }}>
                                            <Badge status={row.status} />{" "}
                                            <span style={{ marginLeft: 6 }}>{row.summary || row.item_type}</span>
                                            <span style={{ color: "#94a3b8", marginLeft: 8 }}>
                                                {new Date(row.created_at).toLocaleString()}
                                                {row.completed_at && <> → {fmtDuration(row.created_at, row.completed_at)}</>}
                                                {row.error && <span style={{ color: "#991b1b" }}> · {String(row.error).slice(0, 160)}</span>}
                                            </span>
                                        </div>
                                    ))}
                                </div>
                            ))}
                        </div>
                    )}

                    {/* ── Divergence warning (silent-overwrite counter) ── */}
                    {divergenceCount > 0 && (
                        <div style={{ background: "#fef3c7", border: "1px solid #fde68a", borderRadius: 8, padding: "10px 14px", marginBottom: 16, fontSize: 13, color: "#92400e" }}>
                            ⚠ <b>{divergenceCount}</b> hand-edited component{divergenceCount === 1 ? " was" : "s were"} overwritten by
                            rebuilds on this site. An unlocked component that has been overwritten before is about to lose
                            someone's edit again — lock components you have hand-patched (Pages → Lock).
                        </div>
                    )}

                    {/* ── Other activity ── */}
                    {otherTypes.length > 0 && (
                        <div style={{ marginBottom: 16 }}>
                            <div style={{ fontSize: 12, fontWeight: 600, color: "#94a3b8", marginBottom: 8, textTransform: "uppercase", letterSpacing: 0.5 }}>
                                Other activity (all work items, not build stages)
                            </div>
                            <div style={{ display: "flex", flexWrap: "wrap", gap: 6 }}>
                                {otherTypes.map(([t, n]) => (
                                    <span key={t} style={{
                                        fontSize: 11, padding: "3px 10px", borderRadius: 9999,
                                        background: t === "page_divergence_overwritten" ? "#fef3c7" : "#f1f5f9",
                                        color: t === "page_divergence_overwritten" ? "#92400e" : "#475569",
                                    }}>
                                        {t} · {String(n)}
                                    </span>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* ── Orchestrations (drill-down) ──
                        The list is capped at WORKFLOW_LIMIT server-side, so its length is
                        the WINDOW, never the total — say "newest N" and mark a full window
                        with a +, rather than printing a number that reads as a count. */}
                    <div style={{ fontSize: 12, fontWeight: 600, color: "#94a3b8", marginBottom: 8, textTransform: "uppercase", letterSpacing: 0.5 }}>
                        Orchestrations tagged to this site — {workflowGroups.length} from the newest {workflows.length}{workflows.length >= WORKFLOW_LIMIT ? "+" : ""} rows, mostly periodic sweeps, not the build
                    </div>
                    {workflows.length === 0 ? (
                        <div style={{ fontSize: 13, color: "#94a3b8", padding: 12 }}>None recorded</div>
                    ) : (
                        <div style={{ background: "#fff", border: "1px solid #e2e8f0", borderRadius: 10, overflow: "hidden" }}>
                            {liveWorkflows.length === 0 && (
                                <div style={{ fontSize: 12, color: "#94a3b8", padding: "8px 16px", borderBottom: "1px solid #f1f5f9" }}>
                                    Nothing running: every orchestration below has finished, so none offers Resume or Terminate.
                                </div>
                            )}
                            {visibleWorkflows.map((wf, idx) => (
                                <div key={wf.correlation_id} onClick={() => openWorkflow(wf)} style={{
                                    display: "flex", alignItems: "center", gap: 12, padding: "8px 16px",
                                    cursor: "pointer", fontSize: 12,
                                    borderTop: idx > 0 ? "1px solid #f1f5f9" : "none",
                                }}>
                                    <span style={{ color: "#64748b", width: 140 }}>{new Date(wf.created_at).toLocaleString()}</span>
                                    {wf.count > 1 && (
                                        <span title={`${wf.count} orchestration rows share this correlation (parent plus sub-orchestrations). They open the same detail.`}
                                              style={{ fontSize: 11, color: "#64748b" }}>×{wf.count}</span>
                                    )}
                                    <Badge status={String(wf.status || "").toLowerCase() === "completed" ? "complete" : String(wf.status || "").toLowerCase()} />
                                    {wf.has_step_error && (
                                        <span title="A step failed even though the status may read complete (bugs_open/099)" style={{
                                            fontSize: 11, color: "#991b1b", background: "#fee2e2", padding: "1px 8px", borderRadius: 4, fontWeight: 700,
                                        }}>step error</span>
                                    )}
                                    <span style={{ color: "#475569", flex: 1, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                                        {wf.current_step || "—"}
                                    </span>
                                    <span onClick={e => { e.stopPropagation(); navigator.clipboard?.writeText(wf.correlation_id); }}
                                          title={`${wf.correlation_id} (click to copy)`}
                                          style={{ fontFamily: "monospace", color: "#94a3b8" }}>
                                        {String(wf.correlation_id).slice(0, 8)}…
                                    </span>
                                </div>
                            ))}
                            {hiddenWorkflowCount > 0 && (
                                <div onClick={() => setShowAllWorkflows(true)}
                                     style={{ padding: "8px 16px", fontSize: 12, color: "#2563eb", cursor: "pointer", borderTop: "1px solid #f1f5f9" }}>
                                    Show {hiddenWorkflowCount} older finished orchestration{hiddenWorkflowCount === 1 ? "" : "s"}
                                </div>
                            )}
                            {showAllWorkflows && workflows.length > liveWorkflows.length + WORKFLOW_PREVIEW && (
                                <div onClick={() => setShowAllWorkflows(false)}
                                     style={{ padding: "8px 16px", fontSize: 12, color: "#2563eb", cursor: "pointer", borderTop: "1px solid #f1f5f9" }}>
                                    Show fewer
                                </div>
                            )}
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}

// ── Media Browser (Assets) ───────────────────────────────────────────────────
function MediaBrowser({ token, siteId, siteDomain, onBack }) {
    const [assets, setAssets] = useState([]);
    const [loading, setLoading] = useState(true);
    const [selectedAsset, setSelectedAsset] = useState(null);
    const [references, setReferences] = useState([]);
    const [actionLoading, setActionLoading] = useState(false);
    const [message, setMessage] = useState("");

    const loadAssets = useCallback(async () => {
        setLoading(true);
        try {
            const data = await apiFetch(`/sites/${siteId}/assets`, token);
            setAssets(data.assets || []);
        } catch (err) { console.error(err); }
        finally { setLoading(false); }
    }, [token, siteId]);

    useEffect(() => { loadAssets(); }, [loadAssets]);

    const loadReferences = async (asset) => {
        setSelectedAsset(asset);
        try {
            const data = await apiFetch(`/sites/${siteId}/assets/${asset.id}/references`, token);
            setReferences(data.references || []);
        } catch (err) {
            console.error(err);
            setReferences([]);
        }
    };

    const handleDelete = async (asset) => {
        if (!confirm(`Delete asset "${asset.purpose || asset.name || asset.id}"? It will be marked as deleted but not removed from storage.`)) return;
        setActionLoading(true);
        try {
            await apiFetch(`/sites/${siteId}/assets/${asset.id}`, token, { method: "DELETE" });
            setMessage("Asset deleted");
            setSelectedAsset(null);
            loadAssets();
        } catch (err) { setMessage("Delete failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    const deployedAssets = assets.filter(a => a.is_deployed && a.status === "active");
    const undeployedAssets = assets.filter(a => !a.is_deployed && a.status === "active");
    const deletedAssets = assets.filter(a => a.status === "deleted");

    const formatSize = (bytes) => {
        if (!bytes) return "";
        if (bytes < 1024) return bytes + "B";
        if (bytes < 1024 * 1024) return Math.round(bytes / 1024) + "KB";
        return (bytes / (1024 * 1024)).toFixed(1) + "MB";
    };

    const renderAssetCard = (asset) => (
        <div key={asset.id} onClick={() => loadReferences(asset)} style={{
            background: selectedAsset?.id === asset.id ? "#f0f9ff" : "#fff",
            border: `1px solid ${selectedAsset?.id === asset.id ? "#93c5fd" : "#e2e8f0"}`,
            borderRadius: 8, padding: "12px 14px", marginBottom: 6, cursor: "pointer",
        }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "start" }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 2 }}>
                        <span style={{ fontSize: 13, fontWeight: 600, color: "#0f172a" }}>
                            {asset.purpose || asset.name || "untitled"}
                        </span>
                        <span style={{ fontSize: 11, color: "#94a3b8", background: "#f1f5f9", padding: "1px 6px", borderRadius: 4 }}>
                            {asset.asset_type}
                        </span>
                    </div>
                    <div style={{ fontSize: 11, color: "#94a3b8", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                        {asset.url}
                    </div>
                    <div style={{ display: "flex", gap: 8, marginTop: 4, fontSize: 11, color: "#94a3b8" }}>
                        {asset.file_size && <span>{formatSize(asset.file_size)}</span>}
                        {asset.dimensions && <span>{(asset.dimensions as Record<string, unknown>).width}x{(asset.dimensions as Record<string, unknown>).height}</span>}
                        <span>{asset.origin_type}</span>
                        <span>{asset.reference_count} ref{asset.reference_count !== 1 ? "s" : ""}</span>
                    </div>
                </div>
                {asset.asset_type === "image" && asset.url && (
                    <img src={asset.url} alt="" style={{
                        width: 48, height: 48, objectFit: "cover", borderRadius: 6,
                        border: "1px solid #e2e8f0", flexShrink: 0, marginLeft: 10,
                    }} onError={e => { (e.target as HTMLImageElement).style.display = "none"; }} />
                )}
            </div>
        </div>
    );

    return (
        <div>
            <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 20 }}>
                <button onClick={onBack} style={btnSecondary}>← Sites</button>
                <h2 style={{ ...sectionTitle, margin: 0, flex: 1 }}>
                    Media <span style={{ fontWeight: 400, color: "#64748b" }}>— {siteDomain}</span>
                </h2>
                <span style={{ fontSize: 13, color: "#64748b" }}>
                    {assets.length} assets ({deployedAssets.length} deployed)
                </span>
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
                <div style={{ textAlign: "center", padding: 40, color: "#94a3b8" }}>Loading assets…</div>
            ) : assets.length === 0 ? (
                <div style={{ textAlign: "center", padding: 40, color: "#94a3b8" }}>No assets found</div>
            ) : (
                <div style={{ display: "flex", gap: 16 }}>
                    {/* Asset list */}
                    <div style={{ flex: "0 0 55%", minWidth: 0 }}>
                        {deployedAssets.length > 0 && (
                            <div style={{ marginBottom: 16 }}>
                                <div style={{ fontSize: 12, fontWeight: 600, color: "#065f46", marginBottom: 8 }}>
                                    Deployed ({deployedAssets.length})
                                </div>
                                {deployedAssets.map(renderAssetCard)}
                            </div>
                        )}
                        {undeployedAssets.length > 0 && (
                            <div style={{ marginBottom: 16 }}>
                                <div style={{ fontSize: 12, fontWeight: 600, color: "#92400e", marginBottom: 8 }}>
                                    Not Deployed ({undeployedAssets.length})
                                </div>
                                {undeployedAssets.map(renderAssetCard)}
                            </div>
                        )}
                        {deletedAssets.length > 0 && (
                            <details>
                                <summary style={{ fontSize: 12, color: "#94a3b8", cursor: "pointer", marginBottom: 8 }}>
                                    Deleted ({deletedAssets.length})
                                </summary>
                                {deletedAssets.map(renderAssetCard)}
                            </details>
                        )}
                    </div>

                    {/* Detail panel */}
                    <div style={{ flex: "0 0 43%", minWidth: 0 }}>
                        {selectedAsset ? (
                            <div style={{
                                background: "#fff", border: "1px solid #e2e8f0", borderRadius: 10,
                                padding: 20, position: "sticky", top: 16,
                            }}>
                                <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 12 }}>
                                    <h3 style={{ margin: 0, fontSize: 16, color: "#0f172a" }}>
                                        {selectedAsset.purpose || selectedAsset.name || "Asset"}
                                    </h3>
                                    <span onClick={() => setSelectedAsset(null)} style={{ cursor: "pointer", color: "#94a3b8", fontSize: 18 }}>✕</span>
                                </div>

                                {/* Preview */}
                                {selectedAsset.asset_type === "image" && selectedAsset.url && (
                                    <div style={{ marginBottom: 12 }}>
                                        <img src={selectedAsset.url} alt="" style={{
                                            maxWidth: "100%", maxHeight: 200, borderRadius: 6,
                                            border: "1px solid #e2e8f0",
                                        }} onError={e => { (e.target as HTMLImageElement).style.display = "none"; }} />
                                    </div>
                                )}

                                {/* Metadata */}
                                <div style={{ display: "grid", gridTemplateColumns: "auto 1fr", gap: "4px 12px", fontSize: 13, marginBottom: 12 }}>
                                    <span style={{ color: "#94a3b8" }}>URL</span>
                                    <a href={selectedAsset.url} target="_blank" rel="noopener" style={{
                                        color: "#1e40af", wordBreak: "break-all", fontSize: 12,
                                    }}>{selectedAsset.url}</a>
                                    <span style={{ color: "#94a3b8" }}>Type</span><span>{selectedAsset.asset_type}</span>
                                    <span style={{ color: "#94a3b8" }}>Origin</span><span>{selectedAsset.origin_type}</span>
                                    {selectedAsset.origin_prompt && <>
                                        <span style={{ color: "#94a3b8" }}>Prompt</span>
                                        <span style={{ fontSize: 12, color: "#64748b" }}>{selectedAsset.origin_prompt}</span>
                                    </>}
                                    {selectedAsset.file_size && <>
                                        <span style={{ color: "#94a3b8" }}>Size</span><span>{formatSize(selectedAsset.file_size)}</span>
                                    </>}
                                    <span style={{ color: "#94a3b8" }}>Status</span>
                                    <span style={{ color: selectedAsset.is_deployed ? "#065f46" : "#92400e" }}>
                                        {selectedAsset.is_deployed ? "Deployed" : "Not deployed"}
                                    </span>
                                </div>

                                {/* References */}
                                <div style={{ marginBottom: 12 }}>
                                    <div style={{ fontSize: 13, fontWeight: 600, color: "#475569", marginBottom: 6 }}>
                                        References ({references.length})
                                    </div>
                                    {references.length === 0 ? (
                                        <div style={{ fontSize: 12, color: "#94a3b8" }}>Not referenced by any component</div>
                                    ) : references.map((ref, i) => (
                                        <div key={i} style={{
                                            fontSize: 12, padding: "4px 0",
                                            borderBottom: i < references.length - 1 ? "1px solid #f1f5f9" : "none",
                                        }}>
                                            <span style={{ fontWeight: 500 }}>{ref.page_name}</span>
                                            <span style={{ color: "#94a3b8" }}> → {ref.slot_name}</span>
                                            {ref.locked && <span style={{ color: "#7c3aed", marginLeft: 4 }}>🔒</span>}
                                        </div>
                                    ))}
                                </div>

                                {/* Actions */}
                                <div style={{ display: "flex", gap: 8, borderTop: "1px solid #e2e8f0", paddingTop: 12 }}>
                                    {selectedAsset.status !== "deleted" && (
                                        <button onClick={() => handleDelete(selectedAsset)} disabled={actionLoading} style={{
                                            ...btnSecondary, fontSize: 12, color: "#991b1b", borderColor: "#fca5a5",
                                        }}>Delete</button>
                                    )}
                                </div>
                            </div>
                        ) : (
                            <div style={{ textAlign: "center", padding: 40, color: "#94a3b8" }}>
                                Select an asset to view details and references
                            </div>
                        )}
                    </div>
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
    const [view, setView] = useState("sites"); // sites | items | all-items | review-queue | pages | specs | media | builds
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
                            { key: "pipelines", label: "Pipelines" },
                            { key: "customers", label: "Customers" },
                            { key: "review-queue", label: "Review Queue" },
                        ].map(({ key, label }) => (
                            <button key={key} onClick={() => { setView(key); setSelectedSite(null); }} style={{
                                background: view === key || (["items", "pages", "specs", "media", "builds"].includes(view) && key === "sites") ? "#1e293b" : "transparent",
                                color: view === key || (["items", "pages", "specs", "media", "builds"].includes(view) && key === "sites") ? "#f1f5f9" : "#94a3b8",
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
                        onSelectMedia={(site) => { setSelectedSite(site); setView("media"); }}
                        onSelectBuilds={(site) => { setSelectedSite(site); setView("builds"); }}
                        onRefresh={loadSites}
                    />
                )}

                {view === "builds" && selectedSite && (
                    <BuildsView
                        token={token}
                        siteId={selectedSite.id}
                        siteDomain={selectedSite.domain}
                        onBack={() => { setView("sites"); setSelectedSite(null); loadSites(); }}
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

                {view === "review-queue" && (
                    <WorkItemsList
                        token={token}
                        siteFilter={null}
                        onBack={() => { setView("sites"); loadSites(); }}
                        reviewQueueMode
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

                {view === "media" && selectedSite && (
                    <MediaBrowser
                        token={token}
                        siteId={selectedSite.id}
                        siteDomain={selectedSite.domain}
                        onBack={() => { setView("sites"); setSelectedSite(null); loadSites(); }}
                    />
                )}

                {view === "pipelines" && (
                    <PipelinesPage
                        token={token}
                        onLogout={handleLogout}
                    />
                )}

                {view === "customers" && (
                    <CustomersPage
                        token={token}
                        onLogout={handleLogout}
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
