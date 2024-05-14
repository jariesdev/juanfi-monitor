// since there's no dynamic data here, we can prerender
// it so that it gets served as a static asset in production
import { redirect } from '@sveltejs/kit';
import { browser } from '$app/environment';

export const prerender = true;

/** @type {import('./$types').PageLoad} */
export async function load(): Promise<void> {
	// const tokenStore = writable(browser && localStorage.getItem("auth_token") || null)
	let token: string | null = browser ? localStorage.getItem('auth_token') : null;

	// tokenStore.subscribe((val) => {
	//     if (browser) {
	//         token = val
	//     }
	// })
	if (browser && !!token) {
		throw redirect(302, '/home');
	} else if (browser) {
		throw redirect(302, '/login');
	}
}
