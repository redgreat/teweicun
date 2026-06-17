import { defineConfig, devices } from '@playwright/test';

const frontendBaseURL = process.env.TWC_FRONTEND_BASE_URL || 'http://127.0.0.1:5173';
const apiBaseURL = process.env.TWC_API_BASE_URL || 'http://localhost:8080/api/v1';

export default defineConfig({
	testDir: './e2e',
	timeout: 60_000,
	expect: {
		timeout: 10_000
	},
	use: {
		baseURL: frontendBaseURL,
		trace: 'retain-on-failure'
	},
	webServer: process.env.TWC_SKIP_FRONTEND_SERVER
		? undefined
		: {
				command: `VITE_API_BASE_URL=${apiBaseURL} npm run dev -- --host 127.0.0.1`,
				url: frontendBaseURL,
				reuseExistingServer: true,
				timeout: 120_000
			},
	projects: [
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] }
		}
	]
});
