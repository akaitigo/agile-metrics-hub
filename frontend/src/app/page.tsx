import Link from "next/link";

export default function Home() {
	return (
		<main style={{ maxWidth: "800px", margin: "0 auto", padding: "2rem" }}>
			<h1 style={{ fontSize: "2rem", fontWeight: 700, marginBottom: "1rem" }}>Agile Metrics Hub</h1>
			<p style={{ color: "#64748b", marginBottom: "2rem" }}>PMツール横断のアジャイルメトリクスダッシュボード</p>

			<div style={{ display: "flex", gap: "1rem" }}>
				<Link
					href="/dashboard"
					style={{
						display: "inline-block",
						padding: "0.75rem 1.5rem",
						background: "#3b82f6",
						color: "white",
						borderRadius: "0.5rem",
						textDecoration: "none",
						fontWeight: 600,
					}}
				>
					ダッシュボードを表示
				</Link>
			</div>
		</main>
	);
}
