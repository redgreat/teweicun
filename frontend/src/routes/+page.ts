import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = () => {
	// 将根路径重定向到仪表盘或系统的业务主入口
	redirect(302, '/dashboard');
};
