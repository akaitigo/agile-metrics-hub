import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
	title: "Agile Metrics Hub",
	description: "PMツール横断のアジャイルメトリクスダッシュボード",
};

export default function RootLayout({ children }: { children: ReactNode }) {
	return (
		<html lang="ja">
			<body>{children}</body>
		</html>
	);
}
