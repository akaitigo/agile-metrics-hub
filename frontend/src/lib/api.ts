const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

function isErrorBody(v: unknown): v is { message: string } {
	return (
		typeof v === "object" && v !== null && "message" in v && typeof (v as Record<string, unknown>).message === "string"
	);
}

export async function fetchAPI<T>(path: string, options?: RequestInit): Promise<T> {
	const res = await fetch(`${API_BASE}${path}`, {
		...options,
		headers: {
			"Content-Type": "application/json",
			...options?.headers,
		},
	});

	if (!res.ok) {
		const body: unknown = await res.json().catch(() => null);
		const message = isErrorBody(body) ? body.message : `API error: ${res.status}`;
		throw new Error(message);
	}

	// 型の実行時検証はPhase 1でzod導入時に対応（ADR #004）
	return (await res.json()) as T;
}

export async function testConnection(source: string, apiKey: string, config: Record<string, string>): Promise<void> {
	await fetchAPI("/api/connections/test", {
		method: "POST",
		body: JSON.stringify({
			source,
			display_name: `${source} connection`,
			api_key: apiKey,
			config,
		}),
	});
}
