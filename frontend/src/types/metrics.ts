export interface BurndownPoint {
	date: string;
	remaining_tasks: number;
	remaining_points: number;
	ideal_remaining: number;
}

export interface VelocityPoint {
	sprint_name: string;
	committed_points: number;
	completed_points: number;
}

export interface VelocityResponse {
	velocity: VelocityPoint[];
	average_velocity: number;
}

export interface CumulativeFlowPoint {
	date: string;
	statuses: Record<string, number>;
}

export interface LeadTimeStats {
	p50_hours: number;
	p85_hours: number;
	p95_hours: number;
	avg_hours: number;
}
