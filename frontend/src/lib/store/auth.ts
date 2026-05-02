import { writable } from 'svelte/store';
import { browser } from '$app/environment';

interface User {
	id: number;
	username: string;
	real_name: string;
	role: string;
	phone?: string;
	email?: string;
	department?: string;
}

function createAuthStore() {
	const { subscribe, set, update } = writable<{
		user: User | null;
		token: string | null;
		isAuthenticated: boolean;
	}>({
		user: null,
		token: browser ? localStorage.getItem('token') : null,
		isAuthenticated: false
	});

	return {
		subscribe,
		login: (user: User, token: string) => {
			if (browser) localStorage.setItem('token', token);
			set({ user, token, isAuthenticated: true });
		},
		logout: () => {
			if (browser) localStorage.removeItem('token');
			set({ user: null, token: null, isAuthenticated: false });
		},
		updateUser: (partial: Partial<User>) => {
			update((state) => {
				if (state.user) {
					state.user = { ...state.user, ...partial };
				}
				return state;
			});
		}
	};
}

export const auth = createAuthStore();
