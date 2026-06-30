import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent }) => {
	const { permissions } = await parent();
	// The default template is admin-managed (users) and lives under the rates feature.
	if (!permissions.includes('rates') || !permissions.includes('users')) {
		throw redirect(302, '/home');
	}
};
