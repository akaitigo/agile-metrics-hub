import { test, expect } from "@playwright/test";

test.describe("スモークテスト", () => {
	test("トップページが表示される", async ({ page }) => {
		await page.goto("/");
		await expect(page).toHaveTitle(/Agile Metrics Hub/);
		const main = page.locator("main");
		await expect(main).toBeVisible();
	});

	test("h1が1つ存在する", async ({ page }) => {
		await page.goto("/");
		const h1 = page.locator("h1");
		await expect(h1).toHaveCount(1);
	});
});
