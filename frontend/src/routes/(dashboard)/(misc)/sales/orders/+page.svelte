<!--
功能：orders页面
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import DataGrid from '$lib/components/DataGrid.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { Tag, ShoppingCart, Plus, Trash2, User, Calendar, Truck } from 'lucide-svelte';
	import { slide } from 'svelte/transition';
	import { dgRowBtnDanger, dgRowSolidPrimary, dgRowSolidSuccess } from '$lib/dgButtonClasses';
	import { todayDateInCn } from '$lib/datetime';

	let orders = $state<any[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let currentPage = $state(1);
	const pageSize = 10;
	let showModal = $state(false);
	let submitting = $state(false);

	// 基础数据
	let customers = $state<any[]>([]);
	let materials = $state<any[]>([]);

	// 表单状态
	let form = $state({
		customer_code: '',
		order_date: todayDateInCn(),
		delivery_date: '',
		remark: '',
		items: [{ material_id: 0, quantity: 1, unit_price: 0, remark: '' }]
	});

	const columns = [
		{ key: 'order_no', label: '销售单号', class: 'font-mono text-primary font-bold' },
		{ key: 'customer_name', label: '客户名称', class: 'font-medium' },
		{ key: 'order_date', label: '下单日期' },
		{ key: 'total_amount', label: '订单总额', class: 'text-right' },
		{ key: 'order_status_name', label: '状态' }
	];

	async function loadInitialData() {
		try {
			const [cusRes, matRes]: any = await Promise.all([
				api.get('/base/customers?page=1&page_size=100'),
				api.get('/base/materials?limit=100')
			]);
			customers = cusRes.list || [];
			materials = matRes.list || [];
		} catch (err) {
			console.error(err);
		}
	}

	async function loadOrders(page = currentPage) {
		loading = true;
		try {
			currentPage = page;
			const res: any = await api.get(`/sales/orders?page=${page}&limit=${pageSize}`);
			orders = res.list || [];
			total = res.total || 0;
		} finally {
			loading = false;
		}
	}

	function addItem() {
		form.items = [...form.items, { material_id: 0, quantity: 1, unit_price: 0, remark: '' }];
	}

	function removeItem(index: number) {
		if (form.items.length > 1) {
			form.items = form.items.filter((_, i) => i !== index);
		}
	}

	async function handleSubmit() {
		submitting = true;
		try {
			await api.post('/sales/orders', form);
			showModal = false;
			loadOrders(1);
		} catch (err: any) {
			toast.error('创建失败: ' + err);
		} finally {
			submitting = false;
		}
	}

	async function handleAction(id: number, action: string) {
		try {
			if (action === 'confirm') await api.post(`/sales/orders/${id}/confirm`);
			if (action === 'ship') await api.post(`/sales/orders/${id}/ship`);
			if (action === 'cancel') await api.post(`/sales/orders/${id}/cancel`);
			loadOrders(currentPage);
		} catch (err: any) {
			toast.error('操作失败: ' + err);
		}
	}

	onMount(() => {
		loadInitialData();
		loadOrders(1);
	});
</script>

<div class="flex min-h-0 flex-1 flex-col gap-4">
	<!-- Top Summary -->
	<div class="grid grid-cols-1 gap-4 md:grid-cols-4">
		<div class="card bg-primary text-primary-content rounded-2xl p-6 shadow-xl">
			<div class="flex items-start justify-between">
				<div>
					<p class="text-xs font-bold tracking-widest uppercase opacity-70">待处理订单</p>
					<h3 class="mt-1 text-3xl font-black">
						{orders.filter((o) => o.status === 'draft').length}
					</h3>
				</div>
				<ShoppingCart size={24} class="opacity-40" />
			</div>
		</div>
		<!-- 更多卡片可以在此添加 -->
	</div>

	<DataGrid
		class="min-h-0 flex-1"
		{columns}
		data={orders}
		{total}
		{loading}
		{pageSize}
		bind:page={currentPage}
		onPageChange={loadOrders}
		onCreate={() => {
			showModal = true;
		}}
	>
		{#snippet rowActions(row)}
			<div class="flex flex-wrap items-center justify-center gap-1.5">
				{#if row.status === 'draft'}
					<button
						type="button"
						class={dgRowSolidSuccess}
						onclick={() => handleAction(row.id, 'confirm')}>确认</button
					>
				{:else if row.status === 'confirmed'}
					<button
						type="button"
						class={dgRowSolidPrimary}
						onclick={() => handleAction(row.id, 'ship')}
					>
						<Truck size={15} /> 发货
					</button>
				{/if}
				{#if row.status !== 'shipped' && row.status !== 'cancelled'}
					<button
						type="button"
						class={dgRowBtnDanger}
						onclick={() => handleAction(row.id, 'cancel')}>取消</button
					>
				{/if}
			</div>
		{/snippet}
	</DataGrid>
</div>

<Modal
	bind:show={showModal}
	title="创建新销售订单"
	onConfirm={handleSubmit}
	loading={submitting}
	maxWidth="max-w-4xl"
>
	<div class="space-y-6">
		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label"
					><span class="label-text flex items-center gap-2 font-medium"
						><User size={14} /> 客户选择</span
					></label
				>
				<select bind:value={form.customer_code} class="select select-bordered bg-base-200/50">
					<option value="">选择往来客户</option>
					{#each customers as cus}
						<option value={cus.customer_code}>{cus.customer_name}</option>
					{/each}
				</select>
			</div>
			<div class="form-control">
				<label class="label"
					><span class="label-text flex items-center gap-2 font-medium"
						><Calendar size={14} /> 下单日期</span
					></label
				>
				<input
					type="date"
					bind:value={form.order_date}
					class="input input-bordered bg-base-200/50"
				/>
			</div>
		</div>

		<div class="space-y-3">
			<div class="flex items-center justify-between">
				<h4 class="flex items-center gap-2 text-sm font-bold">
					<Tag size={16} class="text-pink-500" /> 商品明细
				</h4>
				<button class="btn btn-xs btn-ghost text-primary" onclick={addItem}
					><Plus size={14} /> 增加行</button
				>
			</div>

			<div class="space-y-2">
				{#each form.items as item, i}
					<div
						class="bg-base-100 border-base-200 relative grid grid-cols-12 gap-3 rounded-xl border p-4"
						in:slide
					>
						<div class="col-span-6">
							<select bind:value={item.material_id} class="select select-sm select-bordered w-full">
								<option value={0}>选择物料</option>
								{#each materials as mat}
									<option value={mat.id}>{mat.material_name} ({mat.material_code})</option>
								{/each}
							</select>
						</div>
						<div class="col-span-2">
							<input
								type="number"
								bind:value={item.quantity}
								class="input input-sm input-bordered w-full"
								placeholder="数量"
							/>
						</div>
						<div class="col-span-3">
							<div class="relative">
								<span class="absolute top-1/2 left-2 -translate-y-1/2 text-[10px] opacity-40"
									>￥</span
								>
								<input
									type="number"
									bind:value={item.unit_price}
									class="input input-sm input-bordered w-full pl-6"
									placeholder="单价"
								/>
							</div>
						</div>
						<div class="col-span-1 flex justify-center">
							<button class="btn btn-ghost btn-xs text-error" onclick={() => removeItem(i)}
								><Trash2 size={14} /></button
							>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</div>
</Modal>
