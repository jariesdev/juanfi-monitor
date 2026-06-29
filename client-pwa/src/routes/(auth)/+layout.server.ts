import { redirect } from "@sveltejs/kit";
import type { LayoutServerLoad } from './$types';
import { VITE_INTERNAL_API, NODE_ENV } from '$env/static/private';

const REFRESH_THRESHOLD = 5 * 60; // seconds before expiry to proactively refresh

export const load: LayoutServerLoad = async ({ cookies }) => {
    const token = cookies.get('auth_token');
    const expiry = cookies.get('auth_token_expiry');

    if (!(token && expiry)) {
        throw redirect(302, '/login');
    }

    const now = Math.floor(Date.now() / 1000);
    const expiryTime = parseInt(expiry, 10);

    if (expiryTime <= now) {
        // Token already expired — clear session and send to login
        cookies.delete('auth_token', { path: '/' });
        cookies.delete('auth_token_expiry', { path: '/' });
        throw redirect(302, '/login');
    }

    if (expiryTime - now < REFRESH_THRESHOLD) {
        // Proactively refresh while the token is still valid
        const base = VITE_INTERNAL_API.replace(/\/$/, '');
        const opts = {
            path: '/',
            httpOnly: true,
            sameSite: 'strict' as const,
            secure: NODE_ENV === 'production',
            maxAge: 60 * 60 * 24 * 30
        };
        try {
            const res = await fetch(`${base}/token/refresh`, {
                method: 'POST',
                headers: { Authorization: `Bearer ${token}` }
            });
            if (res.ok) {
                const data = await res.json();
                cookies.set('auth_token', data.access_token, opts);
                cookies.set('auth_token_expiry', String(data.expiry), opts);
            } else {
                cookies.delete('auth_token', { path: '/' });
                cookies.delete('auth_token_expiry', { path: '/' });
                throw redirect(302, '/login');
            }
        } catch (e) {
            // Re-throw SvelteKit redirects; swallow network errors (token still valid until expiry)
            if (e && typeof e === 'object' && 'status' in e) throw e;
        }
    }
};
