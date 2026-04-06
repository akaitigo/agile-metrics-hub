"use client";

import type { CumulativeFlowPoint } from "@/types/metrics";
import { Area, AreaChart, CartesianGrid, Legend, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

interface Props {
	data: CumulativeFlowPoint[];
}

const STATUS_COLORS: Record<string, string> = {
	Todo: "#94a3b8",
	"In Progress": "#f59e0b",
	Done: "#22c55e",
	Review: "#8b5cf6",
	Blocked: "#ef4444",
};

export function CumulativeFlowChart({ data }: Props) {
	// ステータス一覧を収集
	const allStatuses = new Set<string>();
	for (const point of data) {
		for (const status of Object.keys(point.statuses)) {
			allStatuses.add(status);
		}
	}
	const statuses = Array.from(allStatuses);

	const chartData = data.map((p) => {
		const entry: Record<string, string | number> = {
			date: new Date(p.date).toLocaleDateString("ja-JP", { month: "short", day: "numeric" }),
		};
		for (const status of statuses) {
			entry[status] = p.statuses[status] ?? 0;
		}
		return entry;
	});

	return (
		<div>
			<h3 style={{ marginBottom: "0.5rem", fontWeight: 600 }}>累積フローダイアグラム</h3>
			<ResponsiveContainer width="100%" height={300}>
				<AreaChart data={chartData}>
					<CartesianGrid strokeDasharray="3 3" />
					<XAxis dataKey="date" fontSize={12} />
					<YAxis fontSize={12} />
					<Tooltip />
					<Legend />
					{statuses.map((status) => (
						<Area
							key={status}
							type="monotone"
							dataKey={status}
							stackId="1"
							fill={STATUS_COLORS[status] ?? "#6b7280"}
							stroke={STATUS_COLORS[status] ?? "#6b7280"}
						/>
					))}
				</AreaChart>
			</ResponsiveContainer>
		</div>
	);
}
