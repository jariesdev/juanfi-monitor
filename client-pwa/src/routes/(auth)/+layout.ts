// since there's no dynamic data here, we can prerender
// it so that it gets served as a static asset in production
import { redirect } from '@sveltejs/kit';
import { browser } from '$app/environment';

export const prerender = false;

/** @type {import('./$types').PageLoad} */
export async function load(): Promise<void> {
	let token: string | null = browser ? localStorage.getItem('auth_token') : null;

	// TODO authentication check to use SSR
	if (browser && !token) {
		throw redirect(302, '/login');
	}
}
