<!--
功能：categories页面
创建时间：2026-04-18
创建人：wangcw
-->

<script lang="ts">
	import Modal from '$lib/components/Modal.svelte';
	import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { onMount } from 'svelte';
	import { FolderTree, Plus, FolderOpen, ChevronRight, Pencil, Trash2 } from 'lucide-svelte';

	let categories = $state<any[]>([]);
	let loading = $state(true);
	let showModal = $state(false);
	let editingId = $state<number | null>(null);
	let submitting = $state(false);
	let expandedIds = $state<Set<number>>(new Set());
	let showConfirm = $state(false);
	let deleteTarget = $state<any>(null);

	let form = $state({
		parent_id: 0,
		category_code: '',
		category_name: '',
		sort_order: 1,
		status: 'enabled'
	});

	async function loadTree() {
		loading = true;
		try {
			const res: any = await api.get('/base/categories/tree');
			categories = res || [];
			// 默认展开第一层
			if (expandedIds.size === 0) {
				const ids = new Set(categories.map((c: any) => c.id));
				expandedIds = ids;
			}
		} catch (err) {
			console.error(err);
		} finally {
			loading = false;
		}
	}

	function toggleExpand(id: number) {
		const next = new Set(expandedIds);
		if (next.has(id)) {
			next.delete(id);
		} else {
			next.add(id);
		}
		expandedIds = next;
	}

	function openCreateModal(parentId: number = 0) {
		editingId = null;
		form = {
			parent_id: parentId,
			category_code: '',
			category_name: '',
			sort_order: 1,
			status: 'enabled'
		};
		showModal = true;
	}

	function openEditModal(node: any) {
		editingId = node.id;
		form = {
			parent_id: node.parent_id || 0,
			category_code: node.category_code,
			category_name: node.category_name,
			sort_order: node.sort_order,
			status: node.status || 'enabled'
		};
		showModal = true;
	}

	async function handleSubmit() {
		submitting = true;
		try {
			if (editingId) {
				await api.put(`/base/categories/${editingId}`, form);
			} else {
				await api.post('/base/categories', form);
			}
			showModal = false;
			loadTree();
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	async function handleDelete(node: any) {
		deleteTarget = node;
		showConfirm = true;
	}

	async function confirmDelete() {
		if (!deleteTarget) return;
		try {
			await api.delete(`/base/categories/${deleteTarget.id}`);
			showConfirm = false;
			loadTree();
		} catch (err: any) {
			toast.error('删除失败: ' + (err?.message || err));
		}
	}

	// 扁平化统计
	function countNodes(nodes: any[]): number {
		let count = 0;
		for (const node of nodes) {
			count++;
			if (node.children && node.children.length > 0) {
				count += countNodes(node.children);
			}
		}
		return count;
	}

	onMount(loadTree);
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="h-8 w-1.5 rounded-full bg-amber-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">物料分类管理</h1>
		</div>

		<div class="breadcrumbs text-sm opacity-60">
			<ul>
				<li>首页</li>
				<li>基础数据</li>
				<li>物料分类</li>
			</ul>
		</div>
	</div>

	<!-- 工具栏 -->
	<div class="flex items-center justify-between">
		<div class="text-base-content/60 flex items-center gap-3 text-sm">
			<FolderTree size={16} />
			<span>共 {categories.length} 个顶级分类，{countNodes(categories)} 个分类节点</span>
		</div>
		<button
			class="btn btn-sm btn-primary shadow-primary/20 gap-2 rounded-lg shadow-lg"
			onclick={() => openCreateModal(0)}
		>
			<Plus size={16} /> 新增顶级分类
		</button>
	</div>

	<!-- 树形列表 -->
	<div
		class="bg-base-100 shadow-base-300/50 border-base-300 overflow-hidden rounded-3xl border shadow-xl"
	>
		{#if loading}
			<div class="text-base-content/30 p-12 text-center">
				<span class="loading loading-spinner loading-lg"></span>
			</div>
		{:else if categories.length === 0}
			<div class="text-base-content/30 p-12 text-center italic">暂无分类数据</div>
		{:else}
			<div class="divide-base-200 divide-y">
				{#each categories as cat}
					{@render treeNode(cat, 0)}
				{/each}
			</div>
		{/if}
	</div>
</div>

{#snippet treeNode(node: any, depth: number)}
	<div class="group hover:bg-base-200/50 transition-colors">
		<div class="flex items-center gap-2 px-6 py-3" style="padding-left: {24 + depth * 28}px">
			<!-- 展开/折叠 -->
			{#if node.children && node.children.length > 0}
				<button class="btn btn-xs btn-ghost btn-square p-0" onclick={() => toggleExpand(node.id)}>
					<ChevronRight
						size={14}
						class="transition-transform {expandedIds.has(node.id) ? 'rotate-90' : ''}"
					/>
				</button>
			{:else}
				<div class="w-5"></div>
			{/if}

			<!-- 图标 -->
			{#if node.children && node.children.length > 0}
				<FolderOpen size={16} class="flex-shrink-0 text-amber-500" />
			{:else}
				<FolderTree size={16} class="flex-shrink-0 text-amber-500/60" />
			{/if}

			<!-- 分类编码 -->
			<span class="text-primary/80 w-12 flex-shrink-0 font-mono text-sm">{node.category_code}</span>

			<!-- 分类名称 -->
			<span class="flex-1 font-medium">{node.category_name}</span>

			<!-- 排序 -->
			<span class="text-base-content/40 w-8 flex-shrink-0 text-right text-xs"
				>#{node.sort_order}</span
			>

			<!-- 状态 -->
			<span
				class="badge badge-sm {node.status === 'enabled'
					? 'badge-success'
					: 'badge-ghost'} flex-shrink-0"
			>
				{node.status === 'enabled' ? '启用' : '停用'}
			</span>

			<!-- 操作按钮 -->
			<div
				class="flex flex-shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100"
			>
				{#if depth < 1}
					<button
						class="btn btn-sm btn-ghost text-primary hover:bg-primary/10 rounded-md"
						onclick={() => openCreateModal(node.id)}
					>
						<Plus size={15} /> 子分类
					</button>
				{/if}
				<button
					class="btn btn-sm btn-ghost hover:bg-base-200 rounded-md"
					onclick={() => openEditModal(node)}
				>
					<Pencil size={15} />
				</button>
				<button
					class="btn btn-sm btn-ghost text-error hover:bg-error/10 rounded-md"
					onclick={() => handleDelete(node)}
				>
					<Trash2 size={15} />
				</button>
			</div>
		</div>

		<!-- 子节点 -->
		{#if node.children && node.children.length > 0 && expandedIds.has(node.id)}
			{#each node.children as child}
				{@render treeNode(child, depth + 1)}
			{/each}
		{/if}
	</div>
{/snippet}

<!-- 新增/编辑弹窗 -->
<Modal
	bind:show={showModal}
	title={editingId ? '编辑分类' : '新增分类'}
	onConfirm={handleSubmit}
	loading={submitting}
>
	<div class="space-y-4">
		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label" for="material-category-code"
					><span class="label-text">分类编码</span></label
				>
				{#if editingId}
					<input
						id="material-category-code"
						type="text"
						bind:value={form.category_code}
						class="input input-bordered bg-base-200/50 w-full"
						disabled
					/>
				{:else}
					<input
						id="material-category-code"
						type="text"
						value="系统自动生成"
						class="input input-bordered bg-base-200/50 w-full"
						disabled
					/>
				{/if}
			</div>
			<div class="form-control">
				<label class="label" for="material-category-name"
					><span class="label-text">分类名称</span></label
				>
				<input
					id="material-category-name"
					type="text"
					bind:value={form.category_name}
					class="input input-bordered bg-base-200/50 w-full"
					placeholder="如 碳钢板材"
				/>
			</div>
		</div>

		<div class="grid grid-cols-2 gap-4">
			<div class="form-control">
				<label class="label" for="material-category-sort-order"
					><span class="label-text">排序号</span></label
				>
				{#if editingId}
					<input
						id="material-category-sort-order"
						type="number"
						bind:value={form.sort_order}
						class="input input-bordered bg-base-200/50 w-full"
						min="1"
					/>
				{:else}
					<input
						id="material-category-sort-order"
						type="text"
						value="系统自动按同级最大值 + 1"
						class="input input-bordered bg-base-200/50 w-full"
						disabled
					/>
				{/if}
			</div>
			<div class="form-control">
				<label class="label" for="material-category-parent"
					><span class="label-text">上级分类</span></label
				>
				<input
					id="material-category-parent"
					type="text"
					value={form.parent_id === 0 ? '顶级分类' : `ID: ${form.parent_id}`}
					class="input input-bordered bg-base-200/50 w-full"
					disabled
				/>
			</div>
		</div>

		{#if editingId}
			<div class="form-control">
				<label class="label" for="material-category-status"
					><span class="label-text">状态</span></label
				>
				<select
					id="material-category-status"
					bind:value={form.status}
					class="select select-bordered bg-base-200/50 w-full"
				>
					<option value="enabled">可用</option>
					<option value="disabled">禁用</option>
				</select>
			</div>
		{/if}
	</div>
</Modal>

<ConfirmDialog
	bind:show={showConfirm}
	title="删除物料分类"
	message={`确定要删除分类「${deleteTarget?.category_name || ''}」吗？删除后其子分类也将被一并删除，此操作不可撤销。`}
	onConfirm={confirmDelete}
/>
