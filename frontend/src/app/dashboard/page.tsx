"use client";

import { BurndownChart } from "@/components/BurndownChart";
import { CumulativeFlowChart } from "@/components/CumulativeFlowChart";
import { LeadTimeCard } from "@/components/LeadTimeCard";
import { VelocityChart } from "@/components/VelocityChart";
import { fetchAPI } from "@/lib/api";
import type { BurndownPoint, CumulativeFlowPoint, LeadTimeStats, VelocityResponse } from "@/types/metrics";
import { useEffect, useState } from "react";

export default function DashboardPage() {
	const [burndown, setBurndown] = useState<BurndownPoint[]>([]);
	const [velocity, setVelocity] = useState<VelocityResponse | null>(null);
	const [cumulativeFlow, setCumulativeFlow] = useState<CumulativeFlowPoint[]>([]);
	const [leadTime, setLeadTime] = useState<LeadTimeStats | null>(null);
	const [error, setError] = useState<string | null>(null);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		async function loadMetrics() {
			try {
				const [bd, vel, cf, lt] = await Promise.all([
					fetchAPI<BurndownPoint[]>("/api/metrics/burndown?project_id=demo"),
					fetchAPI<VelocityResponse>("/api/metrics/velocity?count=5"),
					fetchAPI<CumulativeFlowPoint[]>("/api/metrics/cumulative-flow?days=30"),
					fetchAPI<LeadTimeStats>("/api/metrics/lead-time?project_id=demo"),
				]);
				setBurndown(bd);
				setVelocity(vel);
				setCumulativeFlow(cf);
				setLeadTime(lt);
			} catch (e) {
				setError(e instanceof Error ? e.message : "Failed to load metrics");
			} finally {
				setLoading(false);
			}
		}
		void loadMetrics();
	}, []);

	if (loading) {
		return (
			<main style={{ maxWidth: "1200px", margin: "0 auto", padding: "2rem" }}>
				<p>Loading metrics...</p>
			</main>
		);
	}

	if (error) {
		return (
			<main style={{ maxWidth: "1200px", margin: "0 auto", padding: "2rem" }}>
				<h1>Agile Metrics Hub</h1>
				<div style={{ padding: "1rem", background: "#fef2f2", borderRadius: "0.5rem", color: "#991b1b" }}>
					<strong>Error:</strong> {error}
					<p style={{ marginTop: "0.5rem", fontSize: "0.875rem" }}>
						Backend API (localhost:8080) が起動しているか確認してください。
					</p>
				</div>
			</main>
		);
	}

	return (
		<main style={{ maxWidth: "1200px", margin: "0 auto", padding: "2rem" }}>
			<h1 style={{ fontSize: "1.5rem", fontWeight: 700, marginBottom: "2rem" }}>Agile Metrics Hub — Dashboard</h1>

			<div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "2rem" }}>
				<div style={{ border: "1px solid #e2e8f0", borderRadius: "0.75rem", padding: "1.5rem" }}>
					<BurndownChart data={burndown} />
				</div>

				<div style={{ border: "1px solid #e2e8f0", borderRadius: "0.75rem", padding: "1.5rem" }}>
					{velocity ? <VelocityChart data={velocity.velocity} averageVelocity={velocity.average_velocity} /> : null}
				</div>

				<div style={{ border: "1px solid #e2e8f0", borderRadius: "0.75rem", padding: "1.5rem" }}>
					<CumulativeFlowChart data={cumulativeFlow} />
				</div>

				<div style={{ border: "1px solid #e2e8f0", borderRadius: "0.75rem", padding: "1.5rem" }}>
					{leadTime ? <LeadTimeCard data={leadTime} /> : null}
				</div>
			</div>
		</main>
	);
}
