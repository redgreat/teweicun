/**
 * 功能：封装前端 API 客户端（axios）
 * 创建时间：2026-04-29
 * 创建人：wangcw
 */
import axios from 'axios';
import { auth } from '../store/auth';
import { get } from 'svelte/store';

const api = axios.create({
	// 默认使用同源的相对路径，避免线上仍回退到 localhost:8080 造成跨域/CORS 问题
	baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
	timeout: 10000
});

// 请求拦截器：注入 Token
api.interceptors.request.use(
	(config) => {
		const { token } = get(auth);
		if (token) {
			config.headers.Authorization = `Bearer ${token}`;
		}
		return config;
	},
	(error) => Promise.reject(error)
);

// 响应拦截器：解包后端统一响应结构 { code, msg, data }，处理 401
api.interceptors.response.use(
	(response) => {
		const body = response.data;
		// 后端统一返回 { code: 0, msg: "success", data: ... }
		if (body && typeof body === 'object' && 'code' in body) {
			if (body.code === 0) {
				return body.data;
			}
			return Promise.reject({ message: body.msg || '请求失败', code: body.code });
		}
		return body;
	},
	(error) => {
		if (error.response?.status === 401) {
			auth.logout();
			window.location.href = '/login';
		}
		return Promise.reject(error.response?.data || error.message);
	}
);

export default api;
