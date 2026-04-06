"use client";

import type { BurndownPoint } from "@/types/metrics";
import { CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

interface Props {
	data: BurndownPoint[];
}

export function BurndownChart({ data }: Props) {
	const chartData = data.map((p) => ({
		date: new Date(p.date).toLocaleDateString("ja-JP", { month: "short", day: "numeric" }),
		残ポイント: Number(p.remaining_points.toFixed(1)),
		理想線: Number(p.ideal_remaining.toFixed(1)),
	}));

	return (
		<div>
			<h3 style={{ marginBottom: "0.5rem", fontWeight: 600 }}>バーンダウンチャート</h3>
			<ResponsiveContainer width="100%" height={300}>
				<LineChart data={chartData}>
					<CartesianGrid strokeDasharray="3 3" />
					<XAxis dataKey="date" fontSize={12} />
					<YAxis fontSize={12} />
					<Tooltip />
					<Legend />
					<Line type="monotone" dataKey="理想線" stroke="#94a3b8" strokeDasharray="5 5" dot={false} />
					<Line type="monotone" dataKey="残ポイント" stroke="#3b82f6" strokeWidth={2} />
				</LineChart>
			</ResponsiveContainer>
		</div>
	);
}
