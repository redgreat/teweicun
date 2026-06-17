import { expect, test } from '@playwright/test';

const apiBaseURL = process.env.TWC_API_BASE_URL || 'http://localhost:8080/api/v1';
const adminUser = process.env.TWC_ADMIN_USER || 'admin';
const adminPass = process.env.TWC_ADMIN_PASS || 'admin123';

type APIEnvelope<T> = {
	code: number;
	msg: string;
	data: T;
};

type LoginResp = {
	token: string;
	user_id: number;
};

type Fixture = {
	categoryID: number;
	warehouseCode: string;
};

async function api<T>(path: string, options: RequestInit = {}, token = ''): Promise<T> {
	const headers = new Headers(options.headers);
	headers.set('Accept', 'application/json');
	if (options.body) headers.set('Content-Type', 'application/json');
	if (token) headers.set('Authorization', `Bearer ${token}`);

	const res = await fetch(`${apiBaseURL}${path}`, { ...options, headers });
	const raw = await res.text();
	let body: APIEnvelope<T>;
	try {
		body = JSON.parse(raw) as APIEnvelope<T>;
	} catch {
		throw new Error(`Invalid JSON from ${path}: ${raw}`);
	}
	if (!res.ok || body.code !== 0) {
		throw new Error(`API ${path} failed: http=${res.status} code=${body.code} msg=${body.msg}`);
	}
	return body.data;
}

async function loginAPI(): Promise<LoginResp> {
	return api<LoginResp>('/auth/login', {
		method: 'POST',
		body: JSON.stringify({ username: adminUser, password: adminPass })
	});
}

async function ensureFixture(token: string, userID: number): Promise<Fixture> {
	const categoryID = await ensureCategory(token);
	const warehouseCode = await ensureWarehouse(token, userID);
	return { categoryID, warehouseCode };
}

async function ensureCategory(token: string): Promise<number> {
	const suffix = `${Date.now()}${Math.floor(Math.random() * 1000)}`;
	const created = await api<{ id: number }>(
		'/base/categories',
		{
			method: 'POST',
			body: JSON.stringify({
				parent_id: 0,
				category_code: `E2E-CAT-${suffix}`,
				category_name: `E2E测试分类${suffix}`,
				sort_order: 1,
				category_type: 'material',
				category_level: 1
			})
		},
		token
	);
	return created.id;
}

async function ensureWarehouse(token: string, managerID: number): Promise<string> {
	const existing = await api<{ list: Array<{ id: number; warehouse_code: string }> }>(
		'/base/warehouses?page=1&page_size=100&warehouse_code=WH001',
		{},
		token
	);
	const found = existing.list.find((row) => row.warehouse_code === 'WH001');
	if (found) return found.warehouse_code;

	const created = await api<{ warehouse_code?: string }>(
		'/base/warehouses',
		{
			method: 'POST',
			body: JSON.stringify({
				warehouse_code: 'WH001',
				warehouse_name: '主材料库',
				warehouse_type: 'main_material',
				manager_id: managerID
			})
		},
		token
	);
	return created.warehouse_code || 'WH001';
}

async function createCodedMaterial(
	token: string,
	categoryID: number
): Promise<{ id: number; name: string }> {
	const suffix = `${Date.now()}${Math.floor(Math.random() * 1000)}`;
	const name = `E2E UI Serial ${suffix}`;
	const created = await api<{ id: number }>(
		'/base/materials',
		{
			method: 'POST',
			body: JSON.stringify({
				category_id: categoryID,
				material_name: name,
				unit: 'pcs',
				is_code: true,
				remark: 'playwright e2e'
			})
		},
		token
	);
	return { id: created.id, name };
}

async function createPendingStockOutWithSerials(token: string): Promise<{ stockOutID: number; firstCode: string }> {
	const login = await loginAPI();
	const fixture = await ensureFixture(token, login.user_id);
	const material = await createCodedMaterial(token, fixture.categoryID);

	const stockIn = await api<{ id: number }>(
		'/stock-in',
		{
			method: 'POST',
			body: JSON.stringify({
				stock_in_date: new Date().toISOString().slice(0, 10),
				stock_in_type: 'purchase',
				warehouse_code: fixture.warehouseCode,
				remark: 'playwright stock-in',
				items: [
					{
						material_id: material.id,
						arrived_quantity: 2,
						accepted_quantity: 2,
						unit_price: 18.6,
						cert_id: 0
					}
				]
			})
		},
		token
	);
	await api<null>(`/stock-in/${stockIn.id}/confirm`, { method: 'POST' }, token);

	const inv = await waitInventoryAvailable(token, fixture.warehouseCode, material.id, material.name, 2);
	if (!inv) throw new Error('created material inventory not found');

	const stockOut = await api<{ id: number }>(
		'/stock-out',
		{
			method: 'POST',
			body: JSON.stringify({
				stock_out_date: new Date().toISOString().slice(0, 10),
				out_type: 'other',
				receiver: 'playwright',
				remark: 'playwright stock-out serial picker',
				items: [{ material_id: material.id, inventory_id: inv.inventory_id, quantity: 2 }]
			})
		},
		token
	);
	const detail = await api<{
		items: Array<{ id: number; material_id: number; is_code: boolean }>;
	}>(`/stock-out/${stockOut.id}`, {}, token);
	const item = detail.items.find((row) => row.material_id === material.id && row.is_code);
	if (!item) throw new Error('created stock-out coded item not found');

	const options = await api<Array<{ id: number; serial_code: string }>>(
		`/serial-codes/stock-out-item/${item.id}/available`,
		{},
		token
	);
	if (options.length < 2) throw new Error(`not enough serial options: ${options.length}`);

	return { stockOutID: stockOut.id, firstCode: options[0].serial_code };
}

async function waitInventoryAvailable(
	token: string,
	warehouseCode: string,
	materialID: number,
	materialName: string,
	need: number
): Promise<{ inventory_id: number } | undefined> {
	for (let attempt = 0; attempt < 10; attempt++) {
		const inventory = await api<{
			list: Array<{ inventory_id: number; material_id: number; available_quantity: number }>;
		}>(
			`/inventory/available?page=1&page_size=20&warehouse_code=${warehouseCode}&q=${encodeURIComponent(materialName)}`,
			{},
			token
		);
		const inv = inventory.list.find(
			(row) => row.material_id === materialID && Number(row.available_quantity) >= need
		);
		if (inv) return inv;
		await new Promise((resolve) => setTimeout(resolve, 500));
	}
	return undefined;
}

test('stock-out confirmation page supports manual and auto serial selection', async ({ page }) => {
	const login = await loginAPI();
	const prepared = await createPendingStockOutWithSerials(login.token);

	await page.goto('/');
	await page.evaluate((token) => localStorage.setItem('token', token), login.token);
	await page.goto(`/stock/out/${prepared.stockOutID}?mode=confirm`);
	const trigger = page.getByTestId('stock-out-serial-picker-trigger');
	await expect(trigger).toContainText('0 / 2');
	await trigger.click();

	const picker = page.getByTestId('stock-out-serial-picker');
	await expect(picker).toBeVisible();
	await page.getByTestId('serial-search-input').fill(prepared.firstCode);
	await picker.getByTestId('serial-option').filter({ hasText: prepared.firstCode }).getByRole('checkbox').check();
	await expect(picker).toContainText('当前已选 1 个');

	await page.getByTestId('serial-clear-selection').click();
	await expect(picker).toContainText('当前已选 0 个');
	await expect(page.getByTestId('serial-save-selection')).toBeEnabled();

	await page.getByTestId('serial-search-input').fill('');
	await page.getByTestId('serial-auto-pick').click();
	await expect(picker).toContainText('当前已选 2 个');
	await expect(page.getByTestId('serial-save-selection')).toBeEnabled();
	await page.getByTestId('serial-save-selection').click();

	await expect(page.getByTestId('stock-out-serial-picker')).toBeHidden();
	await expect(trigger).toContainText('2 / 2');
});
