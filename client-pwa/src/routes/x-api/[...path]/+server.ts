import type { RequestHandler } from './$types';
import { VITE_INTERNAL_API } from '$env/static/private';

export const fallback: RequestHandler = async ({ url, request, cookies }) => {
	const base = VITE_INTERNAL_API.replace(/\/$/, '');

	const target = new URL(base);
	target.pathname = url.pathname.replace('/x-api/', '/');
	target.search = url.search;

	const headers = new Headers();

	const token = cookies.get('auth_token');
	if (token) {
		headers.set('Authorization', `Bearer ${token}`);
	}

	const contentType = request.headers.get('content-type');
	if (contentType) {
		headers.set('Content-Type', contentType);
	}

	const body = ['GET', 'HEAD'].includes(request.method) ? undefined : await request.text();

	return fetch(target, { method: request.method, headers, body });
};
