import { defineConfig } from "vitest/config";
import path from "node:path";

export default defineConfig({
	test: {
		exclude: ["test/e2e/**", "node_modules/**"],
		passWithNoTests: true,
	},
	resolve: {
		alias: {
			"@": path.resolve(__dirname, "./src"),
		},
	},
});
