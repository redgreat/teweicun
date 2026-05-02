import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'warning' | 'info';

interface Toast {
	id: number;
	type: ToastType;
	message: string;
	duration: number;
}

let nextId = 0;

function createToastStore() {
	const { subscribe, update } = writable<Toast[]>([]);

	function add(type: ToastType, message: string, duration = 3000) {
		const id = nextId++;
		update((toasts) => [...toasts, { id, type, message, duration }]);
		if (duration > 0) {
			setTimeout(() => {
				update((toasts) => toasts.filter((t) => t.id !== id));
			}, duration);
		}
	}

	return {
		subscribe,
		success: (message: string, duration?: number) => add('success', message, duration),
		error: (message: string, duration?: number) => add('error', message, duration ?? 4000),
		warning: (message: string, duration?: number) => add('warning', message, duration),
		info: (message: string, duration?: number) => add('info', message, duration),
		dismiss: (id: number) => update((toasts) => toasts.filter((t) => t.id !== id))
	};
}

export const toast = createToastStore();
