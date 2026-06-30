import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ parent }) => {
    const { permissions } = await parent();
    if (!permissions.includes('users')) {
        throw redirect(302, '/home');
    }
};
