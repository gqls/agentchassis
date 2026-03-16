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

                // Arrays of objects (e.g. services) → JSON editor
                if (Array.isArray(value) && value.length > 0 && typeof value[0] === "object") {
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
                                style={{ ...textareaStyle, minHeight: 120 }}
                            />
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
function SitesOverview({ sites, onSelectSite }) {
    return (
        <div>
            <h2 style={sectionTitle}>Sites</h2>
            <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(340px, 1fr))", gap: 16 }}>
                {sites.map(site => (
                    <div key={site.id} onClick={() => onSelectSite(site)} style={{
                        background: "#fff", borderRadius: 10, padding: "20px 22px", cursor: "pointer",
                        border: "1px solid #e2e8f0", transition: "box-shadow 0.15s",
                    }}
                         onMouseEnter={e => e.currentTarget.style.boxShadow = "0 4px 12px rgba(0,0,0,0.08)"}
                         onMouseLeave={e => e.currentTarget.style.boxShadow = "none"}
                    >
                        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "start" }}>
                            <div>
                                <div style={{ fontSize: 16, fontWeight: 600, color: "#0f172a" }}>{site.domain}</div>
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
                        {site.last_deployed && (
                            <div style={{ fontSize: 11, color: "#94a3b8", marginTop: 12 }}>
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
    const [items, setItems] = useState([]);
    const [loading, setLoading] = useState(true);
    const [statusFilter, setStatusFilter] = useState("needs_human_review");
    const [typeFilter, setTypeFilter] = useState("");
    const [selectedItem, setSelectedItem] = useState(null);
    const [actionLoading, setActionLoading] = useState(false);
    const [message, setMessage] = useState("");

    // Editable review data for checkpoint items
    const [editedReviewData, setEditedReviewData] = useState(null);
    const [approveNotes, setApproveNotes] = useState("");

    const loadItems = useCallback(async () => {
        setLoading(true);
        try {
            let path = `/work-items?domain=build`;
            if (statusFilter) path += `&status=${statusFilter}`;
            if (siteFilter?.id) path += `&site_id=${siteFilter.id}`;
            if (typeFilter) path += `&item_type=${typeFilter}`;
            const data = await apiFetch(path, token);
            setItems(data.items || []);
        } catch (err) {
            if (err.message === "UNAUTHORIZED") return;
            console.error(err);
        } finally {
            setLoading(false);
        }
    }, [token, statusFilter, siteFilter, typeFilter]);

    useEffect(() => { loadItems(); }, [loadItems]);

    // When selecting a checkpoint item, initialise the editable review data
    const selectItem = (item) => {
        setSelectedItem(item);
        setApproveNotes("");
        if (item?.spec?.checkpoint && item?.spec?.review_data) {
            setEditedReviewData(JSON.parse(JSON.stringify(item.spec.review_data)));
        } else {
            setEditedReviewData(null);
        }
    };

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

    // Unique item types from loaded items
    const itemTypes = [...new Set(items.map(i => i.item_type))].sort();
    const isCheckpoint = selectedItem?.spec?.checkpoint === true;

    return (
        <div>
            <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 20 }}>
                <button onClick={onBack} style={btnSecondary}>← Sites</button>
                <h2 style={{ ...sectionTitle, margin: 0 }}>
                    Work Items {siteFilter && <span style={{ fontWeight: 400, color: "#64748b" }}>— {siteFilter.domain}</span>}
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

            {/* Filters */}
            <div style={{ display: "flex", gap: 10, marginBottom: 16, flexWrap: "wrap" }}>
                <select value={statusFilter} onChange={e => setStatusFilter(e.target.value)} style={selectStyle}>
                    <option value="">All statuses</option>
                    <option value="needs_human_review">Needs Review</option>
                    <option value="triaged">Triaged</option>
                    <option value="claimed">Claimed</option>
                    <option value="failed">Failed</option>
                    <option value="blocked">Blocked</option>
                    <option value="complete">Complete</option>
                </select>
                <select value={typeFilter} onChange={e => setTypeFilter(e.target.value)} style={selectStyle}>
                    <option value="">All types</option>
                    {itemTypes.map(t => <option key={t} value={t}>{t}</option>)}
                </select>
                <span style={{ fontSize: 13, color: "#64748b", alignSelf: "center" }}>
                    {items.length} items
                </span>
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
                                    <span style={{ fontSize: 11, color: "#94a3b8" }}>{item.domain}</span>
                                    <span style={{ fontSize: 11, color: "#94a3b8" }}>{item.attempts}</span>
                                </div>
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
                                    {isCheckpoint ? "Review & Approve" : "Item Detail"}
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

                            {/* Checkpoint: editable review form */}
                            {isCheckpoint && editedReviewData && (
                                <div style={{ marginBottom: 16 }}>
                                    <div style={{ fontSize: 13, fontWeight: 600, color: "#475569", marginBottom: 10 }}>
                                        Review Data
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

                                    {selectedItem.spec?.on_approve && (
                                        <div style={{ fontSize: 12, color: "#64748b", marginTop: 8 }}>
                                            On approve: creates <strong>{selectedItem.spec.on_approve.item_type || "follow-on"}</strong> item
                                            {selectedItem.spec.on_approve.handler_agent && (
                                                <> → <strong>{selectedItem.spec.on_approve.handler_agent}</strong></>
                                            )}
                                        </div>
                                    )}

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
                                </div>
                            )}

                            {/* Non-checkpoint: read-only spec */}
                            {!isCheckpoint && selectedItem.spec && (
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

                                {/* Standard actions */}
                                {["needs_human_review", "failed", "blocked"].includes(selectedItem.status) && (
                                    <>
                                        {!isCheckpoint && (
                                            <button onClick={() => handleRetry(selectedItem.id)} disabled={actionLoading} style={btnPrimary}>
                                                Retry
                                            </button>
                                        )}
                                        <button onClick={() => {
                                            const reason = prompt("Resolution note (optional):");
                                            handleResolve(selectedItem.id, reason);
                                        }} disabled={actionLoading} style={btnSecondary}>
                                            {isCheckpoint ? "Reject / Skip" : "Resolve"}
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

// ── Main App ─────────────────────────────────────────────────────────────────
export default function App() {
    const [token, setToken] = useState(() => sessionStorage.getItem("admin_token") || "");
    const [user, setUser] = useState(null);
    const [sites, setSites] = useState([]);
    const [view, setView] = useState("sites"); // sites | items
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
    useEffect(() => {
        if (!token) return;
        apiFetch("/sites", token)
            .then(data => setSites(data.sites || []))
            .catch(err => {
                if (err.message === "UNAUTHORIZED") handleLogout();
                else setError(err.message);
            });
    }, [token]);

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
                        onSelectSite={(site) => { setSelectedSite(site); setView("items"); }}
                    />
                )}

                {view === "items" && (
                    <WorkItemsList
                        token={token}
                        siteFilter={selectedSite}
                        onBack={() => { setView("sites"); setSelectedSite(null); }}
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