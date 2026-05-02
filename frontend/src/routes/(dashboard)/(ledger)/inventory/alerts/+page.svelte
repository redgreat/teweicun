<!--
功能：alerts页面
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import api from '$lib/api/client';
	import { onMount } from 'svelte';

	let alerts = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;

	const columns = [
		{ key: 'material_name', label: '物料名称' },
		{ key: 'material_code', label: '物料编码', class: 'font-mono' },
		{ key: 'warehouse_name', label: '所在仓库' },
		{ key: 'current_quantity', label: '当前库存', class: 'text-right' },
		{ key: 'alert_type', label: '预警类型' },
		{ key: 'alert_level', label: '预警级别', class: 'text-center' }
	];

	async function loadAlerts(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const res: any = await api.get(`/inventory/alerts?page=${page}&page_size=${pageSize}`);
			alerts = res.list || [];
			total = res.total || 0;
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	onMount(() => loadAlerts(1));
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={alerts}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadAlerts}
	/>
</div>
