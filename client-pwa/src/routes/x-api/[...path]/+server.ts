import type { RequestHandler } from './$types';
import { VITE_INTERNAL_API, NODE_ENV } from '$env/static/private';
import type { Cookies } from '@sveltejs/kit';

const REFRESH_THRESHOLD = 5 * 60; // seconds before expiry to proactively refresh

function cookieOpts(secure: boolean) {
	return {
		path: '/' as const,
		httpOnly: true,
		sameSite: 'strict' as const,
		secure,
		maxAge: 60 * 60 * 24 * 30
	};
}

async function tryRefresh(token: string, cookies: Cookies): Promise<string | null> {
	const base = VITE_INTERNAL_API.replace(/\/$/, '');
	try {
		const res = await fetch(`${base}/token/refresh`, {
			method: 'POST',
			headers: { Authorization: `Bearer ${token}` }
		});
		if (!res.ok) return null;
		const data = await res.json();
		const opts = cookieOpts(NODE_ENV === 'production');
		cookies.set('auth_token', data.access_token, opts);
		cookies.set('auth_token_expiry', String(data.expiry), opts);
		return data.access_token;
	} catch {
		return null;
	}
}

export const fallback: RequestHandler = async ({ url, request, cookies }) => {
	const base = VITE_INTERNAL_API.replace(/\/$/, '');

	const target = new URL(base);
	target.pathname = url.pathname.replace('/x-api/', '/');
	target.search = url.search;

	let token = cookies.get('auth_token');
	const expiry = cookies.get('auth_token_expiry');

	if (token && expiry) {
		const expiryTime = parseInt(expiry, 10);
		const now = Math.floor(Date.now() / 1000);

		if (expiryTime <= now) {
			// Token already expired — clear session and bail out
			cookies.delete('auth_token', { path: '/' });
			cookies.delete('auth_token_expiry', { path: '/' });
			return new Response(JSON.stringify({ detail: 'session expired' }), {
				status: 401,
				headers: { 'Content-Type': 'application/json' }
			});
		}

		if (expiryTime - now < REFRESH_THRESHOLD) {
			// Proactively refresh before the token window closes
			const newToken = await tryRefresh(token, cookies);
			if (newToken) token = newToken;
		}
	}

	const headers = new Headers();
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
