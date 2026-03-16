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

    const handleRetry = async (id) => {
        setActionLoading(true);
        try {
            await apiFetch(`/work-items/${id}/retry`, token, { method: "POST" });
            setMessage("Item queued for retry");
            setSelectedItem(null);
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
            setSelectedItem(null);
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
            setSelectedItem(null);
            loadItems();
        } catch (err) { setMessage("Update failed: " + err.message); }
        finally { setActionLoading(false); }
    };

    // Unique item types from loaded items
    const itemTypes = [...new Set(items.map(i => i.item_type))].sort();

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
                    <div style={{ flex: selectedItem ? "0 0 50%" : "1", minWidth: 0 }}>
                        {items.map(item => (
                            <div key={item.id} onClick={() => setSelectedItem(item)}
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
                                    <span style={{ fontSize: 11, color: "#94a3b8" }}>{item.domain}</span>
                                    <span style={{ fontSize: 11, color: "#94a3b8" }}>{item.attempts}</span>
                                </div>
                            </div>
                        ))}
                    </div>

                    {/* Detail panel */}
                    {selectedItem && (
                        <div style={{
                            flex: "0 0 48%", background: "#fff", border: "1px solid #e2e8f0", borderRadius: 10,
                            padding: 20, position: "sticky", top: 16, maxHeight: "calc(100vh - 120px)", overflowY: "auto",
                        }}>
                            <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
                                <h3 style={{ margin: 0, fontSize: 16, color: "#0f172a" }}>Item Detail</h3>
                                <span onClick={() => setSelectedItem(null)} style={{ cursor: "pointer", color: "#94a3b8", fontSize: 18 }}>✕</span>
                            </div>

                            <div style={{ fontSize: 14, color: "#334155", lineHeight: 1.6, marginBottom: 16 }}>{selectedItem.summary}</div>

                            <div style={{ display: "grid", gridTemplateColumns: "auto 1fr", gap: "6px 16px", fontSize: 13, marginBottom: 16 }}>
                                <span style={detailLabel}>Status</span><Badge status={selectedItem.status} />
                                <span style={detailLabel}>Type</span><span style={detailValue}>{selectedItem.item_type}</span>
                                <span style={detailLabel}>Severity</span><Badge status={selectedItem.severity} type="severity" />
                                <span style={detailLabel}>Domain</span><span style={detailValue}>{selectedItem.domain}</span>
                                <span style={detailLabel}>Handler</span><span style={detailValue}>{selectedItem.handler_agent}</span>
                                <span style={detailLabel}>Attempts</span><span style={detailValue}>{selectedItem.attempts}</span>
                                <span style={detailLabel}>Created</span><span style={detailValue}>{new Date(selectedItem.created_at).toLocaleString()}</span>
                                {selectedItem.error && <>
                                    <span style={detailLabel}>Error</span>
                                    <span style={{ ...detailValue, color: "#991b1b", fontSize: 12 }}>{selectedItem.error}</span>
                                </>}
                            </div>

                            {/* Spec */}
                            {selectedItem.spec && (
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
                            <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
                                {["needs_human_review", "failed", "blocked"].includes(selectedItem.status) && (
                                    <>
                                        <button onClick={() => handleRetry(selectedItem.id)} disabled={actionLoading} style={btnPrimary}>
                                            Retry
                                        </button>
                                        <button onClick={() => {
                                            const reason = prompt("Resolution note (optional):");
                                            handleResolve(selectedItem.id, reason);
                                        }} disabled={actionLoading} style={btnSecondary}>
                                            Resolve
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