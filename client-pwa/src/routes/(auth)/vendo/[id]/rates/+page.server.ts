import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent, params, fetch }) => {
	const { permissions } = await parent();
	if (!permissions.includes('rates')) {
		throw redirect(302, '/home');
	}
	const res = await fetch(`/x-api/vendo-machines/${params.id}`);
	if (res.status === 403 || res.status === 404) {
		throw redirect(302, '/vendo');
	}

	return { isAdmin: permissions.includes('users') };
};
