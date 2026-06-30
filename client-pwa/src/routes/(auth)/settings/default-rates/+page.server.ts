import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent }) => {
	const { permissions } = await parent();
	// Lives in the Settings area (settings), is part of the rates feature (rates),
	// and the default template is admin-managed (users).
	if (
		!permissions.includes('settings') ||
		!permissions.includes('rates') ||
		!permissions.includes('users')
	) {
		throw redirect(302, '/home');
	}
};
