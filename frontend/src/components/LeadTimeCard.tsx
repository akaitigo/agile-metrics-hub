"use client";

import type { LeadTimeStats } from "@/types/metrics";

interface Props {
	data: LeadTimeStats;
}

export function LeadTimeCard({ data }: Props) {
	const items = [
		{ label: "P50", value: data.p50_hours, color: "#22c55e" },
		{ label: "P85", value: data.p85_hours, color: "#f59e0b" },
		{ label: "P95", value: data.p95_hours, color: "#ef4444" },
		{ label: "平均", value: data.avg_hours, color: "#3b82f6" },
	];

	return (
		<div>
			<h3 style={{ marginBottom: "0.5rem", fontWeight: 600 }}>リードタイム</h3>
			<div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: "1rem" }}>
				{items.map((item) => (
					<div
						key={item.label}
						style={{
							padding: "1rem",
							borderRadius: "0.5rem",
							border: "1px solid #e2e8f0",
							textAlign: "center",
						}}
					>
						<div style={{ fontSize: "0.75rem", color: "#64748b", marginBottom: "0.25rem" }}>{item.label}</div>
						<div style={{ fontSize: "1.5rem", fontWeight: 700, color: item.color }}>{formatHours(item.value)}</div>
					</div>
				))}
			</div>
		</div>
	);
}

function formatHours(hours: number): string {
	if (hours < 24) {
		return `${hours.toFixed(0)}h`;
	}
	return `${(hours / 24).toFixed(1)}d`;
}
