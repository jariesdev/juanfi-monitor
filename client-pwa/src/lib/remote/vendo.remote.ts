import { query, getRequestEvent } from '$app/server';
import * as v from 'valibot';
import { VITE_INTERNAL_API } from '$env/static/private';
import { fail } from '@sveltejs/kit';
import type { iVendo } from '$lib/types/models';

function apiHeaders(): Headers {
	const headers = new Headers();
	headers.set('Accept', 'application/json');
	headers.set('Content-Type', 'application/json');
	const token = getRequestEvent()?.cookies.get('auth_token');
	if (token) {
		headers.set('Authorization', `Bearer ${token}`);
	}
	return headers;
}

export const changeVendoStatus = query(
	v.object({ id: v.number(), status: v.boolean() }),
	async ({ id, status }) => {
		try {
			await fetch(
				new Request(`${VITE_INTERNAL_API}/vendo-machines/${id}/set-status`, {
					method: 'POST',
					body: JSON.stringify({ status: status ? '1' : '0' }),
					headers: apiHeaders()
				})
			).then((response) => {
				if (response.ok) return response.json();
				throw new Error(response.statusText);
			});

			return { success: true };
		} catch (e: unknown) {
			const message =
				e instanceof Error ? e.message : typeof e === 'string' ? e : 'Unknown error';
			return fail(400, { message });
		}
	}
);

export const getVendoInfo = query(v.number(), async (id) => {
	try {
		const data: iVendo = await fetch(
			new Request(`${VITE_INTERNAL_API}/vendo-machines/${id}`, {
				method: 'GET',
				headers: apiHeaders()
			})
		)
			.then((response) => {
				if (response.ok) return response.json();
				throw new Error(response.statusText);
			})
			.then(({ data }) => data);

		return data;
	} catch {
		return null;
	}
});

export const getVendos = query(async (): Promise<iVendo[]> => {
	try {
		return await fetch(
			new Request(`${VITE_INTERNAL_API}/vendo-machines`, {
				method: 'GET',
				headers: apiHeaders()
			})
		)
			.then((response) => {
				if (response.ok) return response.json();
				throw new Error(response.statusText);
			})
			.then(({ data }) => data);
	} catch {
		return [];
	}
});
