<!--
功能：编辑物料页面
创建时间：2026-05-09
创建人：GPT-5.3-Codex
-->

<script lang="ts">
	import api from '$lib/api/client';
	import { toast } from '$lib/store/toast';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { ArrowLeft, Plus, Save, Trash2, Package } from 'lucide-svelte';

	let loading = $state(true);
	let submitting = $state(false);
	let categories = $state<any[]>([]);

	let form = $state<any>({
		category_id: 0,
		material_code: '',
		material_name: '',
		unit: '',
		safety_stock: 0,
		max_stock: 0,
		is_code: true,
		custom_attributes: [] as Array<{ attr_name: string; attr_value: string }>,
		remark: '',
		status: 'enabled'
	});

	let unitOptions = $state<Array<{ value: string; label: string }>>([]);

	async function loadCategories() {
		try {
			const res: any = await api.get('/base/categories/tree');
			const flat: any[] = [];
			const walk = (nodes: any[]) => {
				for (const n of nodes || []) {
					flat.push(n);
					if (n.children?.length) walk(n.children);
				}
			};
			walk(res || []);
			categories = flat;
		} catch (err) {
			console.error(err);
		}
	}

	async function loadUnitOptions() {
		try {
			const res: any = await api.get('/system/dict/unit/data');
			const items = Array.isArray(res) ? res : [];
			unitOptions = items.map((d: any) => ({
				value: String(d.dict_value || ''),
				label: String(d.dict_label || d.dict_value || '')
			}));
		} catch (_err) {
			unitOptions = [
				{ value: 'kg', label: '千克(kg)' },
				{ value: 't', label: '吨(t)' },
				{ value: 'm', label: '米(m)' },
				{ value: 'm2', label: '平方米(m²)' },
				{ value: 'pcs', label: '件' }
			];
		}

		const hasCurrent = unitOptions.some((u) => u.value === form.unit);
		if (!hasCurrent) {
			const preferred = unitOptions.find((u) => u.label.includes('件'));
			form.unit = preferred?.value || unitOptions[0]?.value || '';
		}
	}

	async function loadDetail(id: number) {
		loading = true;
		try {
			const data: any = await api.get(`/base/materials/${id}`);
			form = {
				category_id: data.category_id || 0,
				material_code: data.material_code || '',
				material_name: data.material_name_base || data.material_name || '',
				unit: data.unit || '',
				safety_stock: data.safety_stock || 0,
				max_stock: data.max_stock || 0,
				is_code: !!data.is_code,
				custom_attributes: (data.custom_attributes || []).map((a: any) => ({
					attr_name: a.attr_name || '',
					attr_value: a.attr_value || ''
				})),
				remark: data.remark || '',
				status: data.status || 'enabled'
			};
			const hasCurrent = unitOptions.some((u) => u.value === form.unit);
			if (!hasCurrent) {
				const preferred = unitOptions.find((u) => u.label.includes('件'));
				form.unit = preferred?.value || unitOptions[0]?.value || '';
			}
		} catch (err: any) {
			toast.error('加载详情失败: ' + (err?.message || err));
		} finally {
			loading = false;
		}
	}

	function addCustomAttribute() {
		form.custom_attributes = [...form.custom_attributes, { attr_name: '', attr_value: '' }];
	}

	function removeCustomAttribute(index: number) {
		form.custom_attributes = form.custom_attributes.filter((_: any, i: number) => i !== index);
	}

	async function submit() {
		const id = Number(page.params.id);
		if (!id) return;
		if (!form.category_id) {
			toast.warning('请选择物料分类');
			return;
		}
		if (!String(form.material_name || '').trim()) {
			toast.warning('请输入物料名称');
			return;
		}
		submitting = true;
		try {
			await api.put(`/base/materials/${id}`, {
				...form,
				custom_attributes: form.custom_attributes
					.map((a: any) => ({
						attr_name: String(a.attr_name || '').trim(),
						attr_value: String(a.attr_value || '').trim()
					}))
					.filter((a: any) => a.attr_name || a.attr_value)
			});
			toast.success('保存成功');
			goto('/materials');
		} catch (err: any) {
			toast.error('保存失败: ' + (err?.message || err));
		} finally {
			submitting = false;
		}
	}

	function goBack() {
		goto(`/materials/${page.params.id}`);
	}

	onMount(async () => {
		const id = Number(page.params.id);
		if (!id) return;
		await loadCategories();
		await loadUnitOptions();
		await loadDetail(id);
	});
</script>

<div class="w-full space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="h-8 w-1.5 rounded-full bg-indigo-500"></div>
			<h1 class="text-2xl font-bold tracking-tight">编辑物料</h1>
		</div>
		<div class="breadcrumbs text-base opacity-60">
			<ul>
				<li>首页</li>
				<li>基础数据</li>
				<li><a class="text-primary" href="/materials">物料管理</a></li>
				<li>编辑</li>
			</ul>
		</div>
	</div>

	<div class="flex items-center justify-between gap-3">
		<a href={`/materials/${page.params.id}`} class="btn btn-ghost btn-sm gap-1">
			<ArrowLeft size={14} /> 详情
		</a>
		<div class="flex items-center gap-2">
			<button type="button" class="btn btn-sm" onclick={goBack} disabled={submitting}>取消</button>
			<button
				type="button"
				class="btn btn-primary btn-sm gap-1"
				onclick={submit}
				disabled={submitting || loading}
			>
				{#if submitting}
					<span class="loading loading-spinner loading-xs"></span>
				{/if}
				<Save size={14} /> 保存
			</button>
		</div>
	</div>

	{#if loading}
		<div
			class="bg-base-100 border-base-300 text-base-content/50 rounded-2xl border p-10 text-center text-lg"
		>
			正在加载...
		</div>
	{:else}
		<div class="card bg-base-100 border-base-300 w-full border shadow-lg">
			<div class="card-body space-y-6">
				<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
					<div class="form-control">
						<label class="label" for="mat-edit-category-id">
							<span class="label-text font-medium">物料分类 <span class="text-error">*</span></span>
						</label>
						<select
							id="mat-edit-category-id"
							class="select select-bordered bg-base-200/50 h-11 w-full text-base"
							bind:value={form.category_id}
						>
							<option value={0}>请选择分类</option>
							{#each categories as c}
								<option value={c.id}>{c.category_name}</option>
							{/each}
						</select>
					</div>
					<div class="form-control">
						<label class="label" for="mat-edit-code">
							<span class="label-text font-medium">物料编码</span>
						</label>
						<input
							id="mat-edit-code"
							class="input input-bordered bg-base-200/50 h-11 w-full text-base"
							bind:value={form.material_code}
						/>
					</div>
					<div class="form-control">
						<label class="label" for="mat-edit-name">
							<span class="label-text font-medium">物料名称 <span class="text-error">*</span></span>
						</label>
						<input
							id="mat-edit-name"
							class="input input-bordered bg-base-200/50 h-11 w-full text-base"
							bind:value={form.material_name}
						/>
					</div>
					<div class="form-control">
						<label class="label" for="mat-edit-unit">
							<span class="label-text font-medium">计量单位</span>
						</label>
						<select
							id="mat-edit-unit"
							class="select select-bordered bg-base-200/50 h-11 w-full text-base"
							bind:value={form.unit}
						>
							{#each unitOptions as unit}
								<option value={unit.value}>{unit.label}</option>
							{/each}
						</select>
					</div>
					<div class="form-control">
						<label class="label" for="mat-edit-safety">
							<span class="label-text font-medium">安全库存</span>
						</label>
						<input
							id="mat-edit-safety"
							type="number"
							class="input input-bordered bg-base-200/50 h-11 w-full text-base"
							bind:value={form.safety_stock}
							min="0"
							step="0.01"
						/>
					</div>
					<div class="form-control">
						<label class="label" for="mat-edit-max">
							<span class="label-text font-medium">最大库存</span>
						</label>
						<input
							id="mat-edit-max"
							type="number"
							class="input input-bordered bg-base-200/50 h-11 w-full text-base"
							bind:value={form.max_stock}
							min="0"
							step="0.01"
						/>
					</div>
					<div class="form-control lg:col-span-2">
						<label class="label py-0 pb-1" for="mat-edit-remark">
							<span class="label-text font-medium">备注</span>
						</label>
						<input
							id="mat-edit-remark"
							type="text"
							class="input input-bordered bg-base-200/50 h-11 w-full text-base"
							bind:value={form.remark}
							placeholder="订单备注"
						/>
					</div>
				</div>

				<div class="grid grid-cols-1 gap-5 lg:grid-cols-2">
					<div class="flex items-center gap-2">
						<input
							type="checkbox"
							class="checkbox checkbox-sm checkbox-primary"
							bind:checked={form.is_code}
						/>
						<span class="text-sm">独立编码（用于序列追溯）</span>
					</div>
					<div class="form-control">
						<label class="label py-0 pb-1" for="mat-edit-status">
							<span class="label-text font-medium">状态</span>
						</label>
						<select
							id="mat-edit-status"
							class="select select-bordered bg-base-200/50 h-11 w-full text-base"
							bind:value={form.status}
						>
							<option value="enabled">可用</option>
							<option value="disabled">停用</option>
						</select>
					</div>
				</div>

				<div class="divider">自定义属性</div>

				<div class="space-y-2">
					<div class="flex flex-wrap items-center justify-end gap-2">
						<button type="button" class="btn btn-sm btn-primary" onclick={addCustomAttribute}>
							<Plus size={16} /> 添加属性
						</button>
					</div>
					{#if form.custom_attributes.length === 0}
						<div class="text-base-content/50 py-10 text-center">
							<Package size={48} class="mx-auto mb-4 opacity-30" />
							<div>暂无属性，可添加“材质、重量、长度”等文本属性</div>
						</div>
					{:else}
						<div class="overflow-x-auto">
							<table class="table-zebra table w-full text-base">
								<thead>
									<tr>
										<th class="w-12">#</th>
										<th>属性名</th>
										<th>属性值</th>
										<th class="w-20">操作</th>
									</tr>
								</thead>
								<tbody>
									{#each form.custom_attributes as _, idx}
										<tr>
											<td>{idx + 1}</td>
											<td>
												<input
													class="input input-bordered bg-base-200/50 h-10 w-full text-base"
													bind:value={form.custom_attributes[idx].attr_name}
													placeholder="如：材质"
												/>
											</td>
											<td>
												<input
													class="input input-bordered bg-base-200/50 h-10 w-full text-base"
													bind:value={form.custom_attributes[idx].attr_value}
													placeholder="如：Q235B"
												/>
											</td>
											<td>
												<button
													type="button"
													class="btn btn-sm btn-ghost text-error"
													onclick={() => removeCustomAttribute(idx)}
												>
													<Trash2 size={16} />
												</button>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
					{/if}
				</div>

				<div class="border-base-300 flex flex-wrap items-center justify-end gap-4 border-t pt-6">
					<button type="button" class="btn" onclick={goBack} disabled={submitting}>取消</button>
					<button
						type="button"
						class="btn btn-primary"
						onclick={submit}
						disabled={submitting || loading}
					>
						{#if submitting}
							<span class="loading loading-spinner loading-sm"></span>
						{/if}
						<Save size={16} /> 保存变更
					</button>
				</div>
			</div>
		</div>
	{/if}
</div>
