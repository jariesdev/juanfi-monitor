import {query} from '$app/server';
import * as v from 'valibot'
import {VITE_API_URL} from '$env/static/private'
import {fail} from "@sveltejs/kit";
import type {iVendo} from "$lib/types/models";

export const changeVendoStatus = query(v.object({
	id: v.number(),
	status: v.boolean()
}), async ({id, status}) => {

	try {
		// try login to API
		const baseApiUrl: string = VITE_API_URL
		const headers: Headers = new Headers()
		headers.set('Accept', 'application/json')
		headers.set('Content-Type', 'application/json')
		headers.set('authorization', ``)
		const request: Request = new Request(`${baseApiUrl}/vendo-machines/${id}/set-status`, {
			method: 'POST',
			body: JSON.stringify({status: status ? '1' : '0'}),
			headers
		})
		await fetch(request)
			.then((response) => {
				if (response.ok) {
					return response.json()
				}

				throw new Error(response.statusText)
			})
			.then(({data}) => data)

		return {
			success: true
		}
	} catch (e: unknown) {
		let message: string = ''
		if (e instanceof Error) {
			message = e.message
		} else if (typeof e === "string") {
			message = e
		} else if (typeof e === "object" && e !== null && "message" in e) {
			console.error("Caught a custom error object:", (e as { message: string }).message);
			message = (e as { message: string }).message
		}

		return fail(400, {message, test: 1})
	}
});

export const getVendoInfo = query(v.number(), async (id) => {

	try {
		// try login to API
		const baseApiUrl: string = VITE_API_URL
		const headers: Headers = new Headers()
		headers.set('Accept', 'application/json')
		headers.set('Content-Type', 'application/json')
		headers.set('authorization', ``)
		const request: Request = new Request(`${baseApiUrl}/vendo-machines/${id}`, {
			method: 'GET',
			headers
		})
		const data: iVendo = await fetch(request)
			.then((response) => {
				if (response.ok) {
					return response.json()
				}

				throw new Error(response.statusText)
			})
			.then(({data}) => data)

		return data
	} catch (e: unknown) {
		return null
	}
});