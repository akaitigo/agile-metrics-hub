const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function fetchAPI<T>(path: string, options?: RequestInit): Promise<T> {
	const res = await fetch(`${API_BASE}${path}`, {
		...options,
		headers: {
			"Content-Type": "application/json",
			...options?.headers,
		},
	});

	if (!res.ok) {
		const error = await res.json().catch(() => ({ message: res.statusText }));
		throw new Error((error as { message: string }).message ?? `API error: ${res.status}`);
	}

	return res.json() as Promise<T>;
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
