// FILE: frontends/admin-dashboard/src/pages/CustomersPage.tsx
//
// Admin page for website customers, backed by /api/v1/admin/customers
// (the clients -> networks -> sites chain, migration 375 / ADM-011).
// NOT /admin/clients — that endpoint reads the clients_info tenant store,
// which is a different population and lists no customers (see LANDMINES.md).

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
interface Customer {
    id: string;
    external_id: string | null;
    name: string;
    email: string | null;
    phone: string | null;
    tier: string | null;
    customer_status: string | null;
    site_count: number;
    created_at: string;
}

interface CustomerSite {
    id: string;
    domain: string;
    status: string | null;
    email: string | null;
    phone: string | null;
    network_name: string;
}

interface CustomerForm {
    name: string;
    external_id: string;
    email: string;
    phone: string;
    tier: string;
    customer_status: string;
    notes: string;
}

const emptyForm: CustomerForm = {
    name: "", external_id: "", email: "", phone: "",
    tier: "", customer_status: "", notes: "",
};

// ── Helpers ──────────────────────────────────────────────────────────────
function badgeStyle(color: string, bg: string): React.CSSProperties {
    return {
        display: "inline-block", padding: "3px 10px", borderRadius: 12,
        fontSize: 11, fontWeight: 600, color, background: bg,
    };
}

function StatusBadge({ value }: { value: string | null }) {
    if (!value) return <span style={{ color: "#94a3b8" }}>—</span>;
    return <span style={badgeStyle("#475569", "#f1f5f9")}>{value.replace(/_/g, " ")}</span>;
}

function formatDate(isoStr: string): string {
    if (!isoStr) return "—";
    const d = new Date(isoStr);
    if (isNaN(d.getTime())) return isoStr;
    return d.toLocaleDateString();
}

// null-safe: PATCH sends only non-empty fields; empty strings stay unsent so
// COALESCE keeps the stored value
function formToPatch(form: CustomerForm): Record<string, string> {
    const body: Record<string, string> = {};
    (Object.keys(form) as (keyof CustomerForm)[]).forEach(k => {
        if (form[k].trim() !== "") body[k] = form[k].trim();
    });
    return body;
}

// ── Customer form (create + edit share it) ───────────────────────────────
function CustomerFormFields({ form, setForm }: {
    form: CustomerForm;
    setForm: (f: CustomerForm) => void;
}) {
    const field = (key: keyof CustomerForm, label: string, placeholder = "") => (
        <label style={{ display: "block", marginBottom: 10 }}>
            <div style={formLabel}>{label}</div>
            <input
                style={formInput}
                value={form[key]}
                placeholder={placeholder}
                onChange={e => setForm({ ...form, [key]: e.target.value })}
            />
        </label>
    );
    return (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))", gap: "0 16px" }}>
            {field("name", "Name *")}
            {field("email", "Email")}
            {field("phone", "Phone")}
            {field("tier", "Tier", "done_for_you")}
            {field("customer_status", "Status", "lead")}
            {field("external_id", "External ID (Stripe)")}
            <label style={{ display: "block", marginBottom: 10, gridColumn: "1 / -1" }}>
                <div style={formLabel}>Notes</div>
                <textarea
                    style={{ ...formInput, minHeight: 60, resize: "vertical" as const }}
                    value={form.notes}
                    onChange={e => setForm({ ...form, notes: e.target.value })}
                />
            </label>
        </div>
    );
}

// ── Main page ────────────────────────────────────────────────────────────
export default function CustomersPage({ token, onLogout }: { token: string; onLogout: () => void }) {
    const [customers, setCustomers] = useState<Customer[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const [selected, setSelected] = useState<Customer | null>(null);
    const [selectedSites, setSelectedSites] = useState<CustomerSite[]>([]);
    const [editForm, setEditForm] = useState<CustomerForm | null>(null);
    const [showCreate, setShowCreate] = useState(false);
    const [createForm, setCreateForm] = useState<CustomerForm>(emptyForm);
    const [saving, setSaving] = useState(false);

    const loadData = useCallback(async () => {
        try {
            const data = await apiFetch("/customers", token);
            setCustomers(data.customers || []);
            setError(null);
        } catch (e: any) {
            if (e.message === "UNAUTHORIZED") { onLogout(); return; }
            setError(e.message);
        } finally {
            setLoading(false);
        }
    }, [token, onLogout]);

    useEffect(() => { loadData(); }, [loadData]);

    const openDetail = async (customer: Customer) => {
        try {
            const data = await apiFetch(`/customers/${customer.id}`, token);
            setSelected(data.customer);
            setSelectedSites(data.sites || []);
            setEditForm({
                name: data.customer.name || "",
                external_id: data.customer.external_id || "",
                email: data.customer.email || "",
                phone: data.customer.phone || "",
                tier: data.customer.tier || "",
                customer_status: data.customer.customer_status || "",
                notes: data.notes || "",
            });
            setShowCreate(false);
        } catch (e: any) {
            if (e.message === "UNAUTHORIZED") { onLogout(); return; }
            setError(e.message);
        }
    };

    const saveEdit = async () => {
        if (!selected || !editForm) return;
        setSaving(true);
        try {
            await apiFetch(`/customers/${selected.id}`, token, {
                method: "PATCH",
                body: JSON.stringify(formToPatch(editForm)),
            });
            await loadData();
            await openDetail(selected);
        } catch (e: any) {
            if (e.message === "UNAUTHORIZED") { onLogout(); return; }
            setError(e.message);
        } finally {
            setSaving(false);
        }
    };

    const createCustomer = async () => {
        if (createForm.name.trim() === "") { setError("Name is required"); return; }
        setSaving(true);
        try {
            await apiFetch("/customers", token, {
                method: "POST",
                body: JSON.stringify(formToPatch(createForm)),
            });
            setCreateForm(emptyForm);
            setShowCreate(false);
            await loadData();
        } catch (e: any) {
            if (e.message === "UNAUTHORIZED") { onLogout(); return; }
            setError(e.message);
        } finally {
            setSaving(false);
        }
    };

    if (loading) return <div style={{ padding: 40, textAlign: "center", color: "#94a3b8" }}>Loading customers…</div>;

    return (
        <div>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
                <h2 style={{ fontSize: 20, fontWeight: 600, color: "#0f172a", margin: 0 }}>Customers</h2>
                <div style={{ display: "flex", gap: 8 }}>
                    <button onClick={loadData} style={btnSecondary}>Refresh</button>
                    <button
                        onClick={() => { setShowCreate(!showCreate); setSelected(null); setEditForm(null); }}
                        style={btnPrimary}
                    >
                        {showCreate ? "Cancel" : "New Customer"}
                    </button>
                </div>
            </div>

            {error && (
                <div style={{ background: "#fef2f2", border: "1px solid #fecaca", borderRadius: 8, padding: "10px 14px", marginBottom: 16, color: "#dc2626", fontSize: 13 }}>
                    {error}
                    <span onClick={() => setError(null)} style={{ float: "right", cursor: "pointer" }}>✕</span>
                </div>
            )}

            {showCreate && (
                <div style={{ ...cardStyle, marginBottom: 24 }}>
                    <div style={cardTitle}>New customer</div>
                    <CustomerFormFields form={createForm} setForm={setCreateForm} />
                    <button onClick={createCustomer} disabled={saving} style={btnPrimary}>
                        {saving ? "Creating…" : "Create"}
                    </button>
                </div>
            )}

            <div style={{ background: "#fff", border: "1px solid #e2e8f0", borderRadius: 10, overflow: "hidden", marginBottom: 24 }}>
                <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}>
                    <thead>
                    <tr style={{ background: "#f8fafc", borderBottom: "1px solid #e2e8f0" }}>
                        <th style={thStyle}>Name</th>
                        <th style={thStyle}>Email</th>
                        <th style={thStyle}>Phone</th>
                        <th style={thStyle}>Tier</th>
                        <th style={thStyle}>Status</th>
                        <th style={thStyle}>Sites</th>
                        <th style={thStyle}>Created</th>
                    </tr>
                    </thead>
                    <tbody>
                    {customers.length === 0 && (
                        <tr><td style={{ ...tdStyle, color: "#94a3b8" }} colSpan={7}>No customers yet</td></tr>
                    )}
                    {customers.map(cust => (
                        <tr
                            key={cust.id}
                            onClick={() => openDetail(cust)}
                            style={{
                                borderBottom: "1px solid #f1f5f9", cursor: "pointer",
                                background: selected?.id === cust.id ? "#eff6ff" : undefined,
                            }}
                        >
                            <td style={tdStyle}><span style={{ fontWeight: 500, color: "#0f172a" }}>{cust.name}</span></td>
                            <td style={tdStyle}>{cust.email || <span style={{ color: "#94a3b8" }}>—</span>}</td>
                            <td style={tdStyle}>{cust.phone || <span style={{ color: "#94a3b8" }}>—</span>}</td>
                            <td style={tdStyle}><StatusBadge value={cust.tier} /></td>
                            <td style={tdStyle}><StatusBadge value={cust.customer_status} /></td>
                            <td style={tdStyle}>{cust.site_count}</td>
                            <td style={tdStyle}>
                                <span style={{ color: "#64748b", fontSize: 12 }}>{formatDate(cust.created_at)}</span>
                            </td>
                        </tr>
                    ))}
                    </tbody>
                </table>
            </div>

            {selected && editForm && (
                <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(320px, 1fr))", gap: 16 }}>
                    <div style={cardStyle}>
                        <div style={cardTitle}>Edit — {selected.name}</div>
                        <CustomerFormFields form={editForm} setForm={setEditForm} />
                        <div style={{ display: "flex", gap: 8 }}>
                            <button onClick={saveEdit} disabled={saving} style={btnPrimary}>
                                {saving ? "Saving…" : "Save"}
                            </button>
                            <button onClick={() => { setSelected(null); setEditForm(null); }} style={btnSecondary}>Close</button>
                        </div>
                        <div style={{ marginTop: 10, fontSize: 11, color: "#94a3b8" }}>
                            Blank fields are left unchanged. Customer id: {selected.id}
                        </div>
                    </div>
                    <div style={cardStyle}>
                        <div style={cardTitle}>Sites ({selectedSites.length})</div>
                        {selectedSites.length === 0 ? (
                            <div style={{ fontSize: 12, color: "#94a3b8" }}>No sites owned by this customer</div>
                        ) : (
                            <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}>
                                <thead>
                                <tr style={{ borderBottom: "1px solid #e2e8f0" }}>
                                    <th style={thStyle}>Domain</th>
                                    <th style={thStyle}>Status</th>
                                    <th style={thStyle}>Network</th>
                                </tr>
                                </thead>
                                <tbody>
                                {selectedSites.map(site => (
                                    <tr key={site.id} style={{ borderBottom: "1px solid #f1f5f9" }}>
                                        <td style={tdStyle}><span style={{ fontWeight: 500 }}>{site.domain}</span></td>
                                        <td style={tdStyle}><StatusBadge value={site.status} /></td>
                                        <td style={tdStyle}>
                                            <span style={{ color: "#64748b", fontSize: 12 }}>{site.network_name}</span>
                                        </td>
                                    </tr>
                                ))}
                                </tbody>
                            </table>
                        )}
                    </div>
                </div>
            )}
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
const formLabel: React.CSSProperties = { fontSize: 11, fontWeight: 600, color: "#64748b", marginBottom: 4, textTransform: "uppercase", letterSpacing: "0.03em" };
const formInput: React.CSSProperties = {
    width: "100%", padding: "7px 10px", border: "1px solid #cbd5e1", borderRadius: 6,
    fontSize: 13, boxSizing: "border-box" as const,
};
const btnPrimary: React.CSSProperties = {
    padding: "5px 12px", background: "#1e40af", color: "#fff", border: "none",
    borderRadius: 5, fontSize: 12, fontWeight: 500, cursor: "pointer",
};
const btnSecondary: React.CSSProperties = {
    padding: "5px 12px", background: "#fff", color: "#334155", border: "1px solid #cbd5e1",
    borderRadius: 5, fontSize: 12, fontWeight: 500, cursor: "pointer",
};
