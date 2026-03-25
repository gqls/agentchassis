// FILE: frontends/admin-dashboard/src/pages/PipelinesPage.tsx
//
// Admin page for managing data pipelines (scheduled_tasks).
// Shows all tasks with status, enable/disable toggles, trigger buttons,
// and aggregate stats (CH enrichment progress, HTTP/LLM call summaries).

import React, { useState, useEffect, useCallback } from "react";

const API_BASE = "/api/v1/admin";

async function apiFetch(path: string, token: string, opts: any = {}) {
    const res = await fetch(`${API_BASE}${path}`, {
        ...opts,
        headers: {
            Authorization: `Bearer ${token}`,
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

// ── Types ────────────────────────────────────────────────────────────────
interface Pipeline {
    id: string;
    name: string;
    target_agent_type: string;
    target_topic: string;
    enabled: boolean;
    interval_seconds: number;
    concurrency_group: string | null;
    max_concurrent: number;
    timeout_seconds: number;
    last_triggered_at: string | null;
    last_completed_at: string | null;
    fire_message: boolean;
    state: string;
    last_duration_seconds: number | null;
}

interface PipelineStats {
    ch_enrichment?: {
        total_companies?: number;
        matched?: number;
        detail_fetched?: number;
        accounts_fetched?: number;
        financial_coverage?: {
            net_worth?: number;
            employees?: number;
            turnover?: number;
            profit_loss?: number;
        };
    };
    verification?: {
        by_status?: Record<string, number>;
    };
    http_requests_24h?: Array<{
        domain: string;
        action: string;
        total_calls: number;
        successes: number;
        failures: number;
        avg_latency_ms: number;
    }>;
    llm_calls_24h?: Array<{
        agent_type: string;
        model: string;
        calls: number;
        avg_latency_ms: number;
        total_input_tokens: number;
        total_output_tokens: number;
    }>;
}

// ── Helpers ──────────────────────────────────────────────────────────────
function StateBadge({ state, enabled }: { state: string; enabled: boolean }) {
    if (!enabled) return <span style={badgeStyle("#94a3b8", "#f1f5f9")}>disabled</span>;
    const colors: Record<string, [string, string]> = {
        idle: ["#059669", "#ecfdf5"],
        in_flight: ["#d97706", "#fffbeb"],
        timed_out: ["#dc2626", "#fef2f2"],
        never_run: ["#6366f1", "#eef2ff"],
    };
    const [fg, bg] = colors[state] || ["#64748b", "#f8fafc"];
    return <span style={badgeStyle(fg, bg)}>{state.replace(/_/g, " ")}</span>;
}

function badgeStyle(color: string, bg: string): React.CSSProperties {
    return {
        display: "inline-block", padding: "3px 10px", borderRadius: 12,
        fontSize: 11, fontWeight: 600, color, background: bg,
    };
}

function formatInterval(seconds: number): string {
    if (seconds >= 86400) return `${Math.round(seconds / 86400)}d`;
    if (seconds >= 3600) return `${Math.round(seconds / 3600)}h`;
    if (seconds >= 60) return `${Math.round(seconds / 60)}m`;
    return `${seconds}s`;
}

function formatDuration(seconds: number | null | undefined): string {
    if (!seconds && seconds !== 0) return "—";
    if (seconds >= 3600) return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
    if (seconds >= 60) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
    return `${seconds}s`;
}

function timeAgo(isoStr: string | null): string {
    if (!isoStr) return "never";
    const diff = (Date.now() - new Date(isoStr).getTime()) / 1000;
    if (diff < 60) return "just now";
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
    return `${Math.floor(diff / 86400)}d ago`;
}

function formatNumber(n: number | null | undefined): string {
    if (n == null) return "—";
    return n.toLocaleString();
}

// ── Progress bar ─────────────────────────────────────────────────────────
function ProgressBar({ value, max, label }: { value: number; max: number; label: string }) {
    const pct = max > 0 ? Math.round((value / max) * 100) : 0;
    return (
        <div style={{ marginBottom: 8 }}>
            <div style={{ display: "flex", justifyContent: "space-between", fontSize: 12, color: "#475569", marginBottom: 2 }}>
                <span>{label}</span>
                <span>{formatNumber(value)} / {formatNumber(max)} ({pct}%)</span>
            </div>
            <div style={{ height: 6, background: "#e2e8f0", borderRadius: 3, overflow: "hidden" }}>
                <div style={{
                    height: "100%", width: `${pct}%`,
                    background: pct === 100 ? "#059669" : "#3b82f6",
                    borderRadius: 3, transition: "width 0.3s",
                }} />
            </div>
        </div>
    );
}

// ── Stats cards ──────────────────────────────────────────────────────────
function StatsCards({ stats }: { stats: PipelineStats }) {
    const ch = stats.ch_enrichment || {};
    const verify = stats.verification || {};
    const httpReqs = stats.http_requests_24h || [];
    const llmCalls = stats.llm_calls_24h || [];
    const financial = ch.financial_coverage || {};

    const verifyByStatus = verify.by_status || {};
    const totalBiz = Object.values(verifyByStatus).reduce((a, b) => a + b, 0);

    return (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(280px, 1fr))", gap: 16, marginBottom: 28 }}>
            {/* CH Enrichment */}
            <div style={cardStyle}>
                <div style={cardTitle}>Companies House Enrichment</div>
                <ProgressBar value={ch.matched || 0} max={ch.total_companies || 0} label="Matched" />
                <ProgressBar value={ch.detail_fetched || 0} max={ch.matched || 0} label="Details fetched" />
                <ProgressBar value={ch.accounts_fetched || 0} max={ch.detail_fetched || 0} label="Accounts fetched" />
                <div style={{ marginTop: 8, fontSize: 12, color: "#64748b" }}>
                    Financial: {formatNumber(financial.net_worth)} net worth · {formatNumber(financial.employees)} employees · {formatNumber(financial.turnover)} turnover
                </div>
            </div>

            {/* Verification */}
            <div style={cardStyle}>
                <div style={cardTitle}>Business Verification</div>
                <div style={{ fontSize: 24, fontWeight: 700, color: "#0f172a" }}>{formatNumber(totalBiz)}</div>
                <div style={{ fontSize: 12, color: "#64748b", marginBottom: 8 }}>total businesses</div>
                <div style={{ display: "flex", flexWrap: "wrap", gap: 6 }}>
                    {Object.entries(verifyByStatus).map(([status, count]) => (
                        <span key={status} style={{ ...badgeStyle("#475569", "#f1f5f9"), fontSize: 11 }}>
                            {status}: {count}
                        </span>
                    ))}
                </div>
            </div>

            {/* HTTP Requests */}
            <div style={cardStyle}>
                <div style={cardTitle}>HTTP Requests (24h)</div>
                {httpReqs.length === 0 ? (
                    <div style={{ fontSize: 12, color: "#94a3b8" }}>No requests logged</div>
                ) : (
                    httpReqs.slice(0, 5).map((r, i) => (
                        <div key={i} style={{ display: "flex", justifyContent: "space-between", fontSize: 12, padding: "3px 0", borderBottom: "1px solid #f1f5f9" }}>
                            <span style={{ color: "#475569" }}>{r.action || r.domain}</span>
                            <span>
                                <span style={{ color: "#059669" }}>{r.successes}</span>
                                {r.failures > 0 && <span style={{ color: "#dc2626", marginLeft: 4 }}>/ {r.failures} err</span>}
                                <span style={{ color: "#94a3b8", marginLeft: 6 }}>{r.avg_latency_ms}ms</span>
                            </span>
                        </div>
                    ))
                )}
            </div>

            {/* LLM Calls */}
            <div style={cardStyle}>
                <div style={cardTitle}>LLM Calls (24h)</div>
                {llmCalls.length === 0 ? (
                    <div style={{ fontSize: 12, color: "#94a3b8" }}>No calls logged</div>
                ) : (
                    llmCalls.slice(0, 5).map((l, i) => (
                        <div key={i} style={{ display: "flex", justifyContent: "space-between", fontSize: 12, padding: "3px 0", borderBottom: "1px solid #f1f5f9" }}>
                            <span style={{ color: "#475569" }}>{l.agent_type}</span>
                            <span>
                                <span style={{ color: "#0f172a", fontWeight: 500 }}>{l.calls}</span>
                                <span style={{ color: "#94a3b8", marginLeft: 6 }}>{l.model}</span>
                                <span style={{ color: "#94a3b8", marginLeft: 6 }}>{l.avg_latency_ms}ms</span>
                            </span>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
}

// ── Main page ────────────────────────────────────────────────────────────
export default function PipelinesPage({ token, onLogout }: { token: string; onLogout: () => void }) {
    const [pipelines, setPipelines] = useState<Pipeline[]>([]);
    const [stats, setStats] = useState<PipelineStats | null>(null);
    const [loading, setLoading] = useState(true);
    const [actionLoading, setActionLoading] = useState<string | null>(null);
    const [error, setError] = useState<string | null>(null);

    const loadData = useCallback(async () => {
        try {
            const [pData, sData] = await Promise.all([
                apiFetch("/pipelines", token),
                apiFetch("/pipelines/stats", token),
            ]);
            setPipelines(pData.pipelines || []);
            setStats(sData);
            setError(null);
        } catch (e: any) {
            if (e.message === "UNAUTHORIZED") { onLogout(); return; }
            setError(e.message);
        } finally {
            setLoading(false);
        }
    }, [token, onLogout]);

    useEffect(() => {
        loadData();
        const interval = setInterval(loadData, 15000);
        return () => clearInterval(interval);
    }, [loadData]);

    const toggleEnabled = async (name: string, currentlyEnabled: boolean) => {
        setActionLoading(name);
        try {
            await apiFetch(`/pipelines/${name}`, token, {
                method: "PATCH",
                body: JSON.stringify({ enabled: !currentlyEnabled }),
            });
            await loadData();
        } catch (e: any) {
            if (e.message === "UNAUTHORIZED") { onLogout(); return; }
            setError(e.message);
        } finally {
            setActionLoading(null);
        }
    };

    const triggerPipeline = async (name: string) => {
        setActionLoading(name);
        try {
            await apiFetch(`/pipelines/${name}/trigger`, token, { method: "POST" });
            await loadData();
        } catch (e: any) {
            if (e.message === "UNAUTHORIZED") { onLogout(); return; }
            setError(e.message);
        } finally {
            setActionLoading(null);
        }
    };

    // Group pipelines by concurrency_group
    const groups: Record<string, Pipeline[]> = {};
    pipelines.forEach(p => {
        const g = p.concurrency_group || "ungrouped";
        if (!groups[g]) groups[g] = [];
        groups[g].push(p);
    });

    if (loading) return <div style={{ padding: 40, textAlign: "center", color: "#94a3b8" }}>Loading pipelines…</div>;

    return (
        <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
                <h2 style={{ fontSize: 20, fontWeight: 600, color: "#0f172a", margin: 0 }}>Data Pipelines</h2>
                <button onClick={loadData} style={btnSecondary}>Refresh</button>
            </div>

            {error && (
                <div style={{ background: "#fef2f2", border: "1px solid #fecaca", borderRadius: 8, padding: "10px 14px", marginBottom: 16, color: "#dc2626", fontSize: 13 }}>
                    {error}
                    <span onClick={() => setError(null)} style={{ float: "right", cursor: "pointer" }}>✕</span>
                </div>
            )}

            {stats && <StatsCards stats={stats} />}

            {Object.entries(groups).map(([group, tasks]) => (
                <div key={group} style={{ marginBottom: 24 }}>
                    <h3 style={{ fontSize: 13, fontWeight: 600, color: "#64748b", textTransform: "uppercase", letterSpacing: "0.05em", marginBottom: 10 }}>
                        {group}
                    </h3>
                    <div style={{ background: "#fff", border: "1px solid #e2e8f0", borderRadius: 10, overflow: "hidden" }}>
                        <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}>
                            <thead>
                            <tr style={{ background: "#f8fafc", borderBottom: "1px solid #e2e8f0" }}>
                                <th style={thStyle}>Task</th>
                                <th style={thStyle}>Agent</th>
                                <th style={thStyle}>Interval</th>
                                <th style={thStyle}>State</th>
                                <th style={thStyle}>Last Run</th>
                                <th style={thStyle}>Duration</th>
                                <th style={{ ...thStyle, textAlign: "right" as const }}>Actions</th>
                            </tr>
                            </thead>
                            <tbody>
                            {tasks.map(p => (
                                <tr key={p.name} style={{ borderBottom: "1px solid #f1f5f9" }}>
                                    <td style={tdStyle}>
                                        <span style={{ fontWeight: 500, color: "#0f172a" }}>{p.name}</span>
                                    </td>
                                    <td style={tdStyle}>
                                        <span style={{ color: "#475569", fontSize: 12 }}>{p.target_agent_type}</span>
                                    </td>
                                    <td style={tdStyle}>{formatInterval(p.interval_seconds)}</td>
                                    <td style={tdStyle}><StateBadge state={p.state} enabled={p.enabled} /></td>
                                    <td style={tdStyle}>
                                        <span style={{ color: "#64748b", fontSize: 12 }}>{timeAgo(p.last_completed_at)}</span>
                                    </td>
                                    <td style={tdStyle}>
                                        <span style={{ color: "#64748b", fontSize: 12 }}>{formatDuration(p.last_duration_seconds)}</span>
                                    </td>
                                    <td style={{ ...tdStyle, textAlign: "right" as const }}>
                                        <div style={{ display: "flex", gap: 6, justifyContent: "flex-end" }}>
                                            <button
                                                onClick={() => toggleEnabled(p.name, p.enabled)}
                                                disabled={actionLoading === p.name}
                                                style={p.enabled ? btnDanger : btnSuccess}
                                            >
                                                {p.enabled ? "Disable" : "Enable"}
                                            </button>
                                            {p.enabled && p.state !== "in_flight" && (
                                                <button
                                                    onClick={() => triggerPipeline(p.name)}
                                                    disabled={actionLoading === p.name}
                                                    style={btnPrimary}
                                                >
                                                    Run Now
                                                </button>
                                            )}
                                        </div>
                                    </td>
                                </tr>
                            ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            ))}
        </div>
    );
}

// ── Styles ───────────────────────────────────────────────────────────────
const thStyle: React.CSSProperties = {
    padding: "10px 14px", textAlign: "left", fontSize: 11, fontWeight: 600,
    color: "#64748b", textTransform: "uppercase", letterSpacing: "0.03em",
};
const tdStyle: React.CSSProperties = { padding: "10px 14px" };
const cardStyle: React.CSSProperties = { background: "#fff", border: "1px solid #e2e8f0", borderRadius: 10, padding: 16 };
const cardTitle: React.CSSProperties = { fontSize: 13, fontWeight: 600, color: "#0f172a", marginBottom: 12 };

const btnPrimary: React.CSSProperties = {
    padding: "5px 12px", background: "#1e40af", color: "#fff", border: "none",
    borderRadius: 5, fontSize: 12, fontWeight: 500, cursor: "pointer",
};
const btnSecondary: React.CSSProperties = {
    padding: "5px 12px", background: "#fff", color: "#334155", border: "1px solid #cbd5e1",
    borderRadius: 5, fontSize: 12, fontWeight: 500, cursor: "pointer",
};
const btnSuccess: React.CSSProperties = {
    padding: "5px 12px", background: "#059669", color: "#fff", border: "none",
    borderRadius: 5, fontSize: 12, fontWeight: 500, cursor: "pointer",
};
const btnDanger: React.CSSProperties = {
    padding: "5px 12px", background: "#fff", color: "#dc2626", border: "1px solid #fca5a5",
    borderRadius: 5, fontSize: 12, fontWeight: 500, cursor: "pointer",
};