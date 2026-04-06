"use client";

import type { VelocityPoint } from "@/types/metrics";
import { Bar, BarChart, CartesianGrid, Legend, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

interface Props {
	data: VelocityPoint[];
	averageVelocity: number;
}

export function VelocityChart({ data, averageVelocity }: Props) {
	const chartData = data.map((p) => ({
		name: p.sprint_name,
		コミット: p.committed_points,
		完了: p.completed_points,
	}));

	return (
		<div>
			<h3 style={{ marginBottom: "0.5rem", fontWeight: 600 }}>
				ベロシティ{" "}
				<span style={{ fontWeight: 400, fontSize: "0.875rem", color: "#64748b" }}>
					平均: {Number.isFinite(averageVelocity) ? averageVelocity.toFixed(1) : "--"}pt
				</span>
			</h3>
			<ResponsiveContainer width="100%" height={300}>
				<BarChart data={chartData}>
					<CartesianGrid strokeDasharray="3 3" />
					<XAxis dataKey="name" fontSize={12} />
					<YAxis fontSize={12} />
					<Tooltip />
					<Legend />
					<Bar dataKey="コミット" fill="#94a3b8" />
					<Bar dataKey="完了" fill="#3b82f6" />
				</BarChart>
			</ResponsiveContainer>
		</div>
	);
}
