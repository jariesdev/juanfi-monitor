import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent, url }) => {
	const { permissions } = await parent();
	if (!permissions.includes('logs')) {
		throw redirect(302, '/home');
	}

	const vendoID = Number(url.searchParams.get('vendo_id'));
	return {
		initialVendoId: Number.isInteger(vendoID) && vendoID > 0 ? vendoID : undefined
	};
};
