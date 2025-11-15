import {redirect} from "@sveltejs/kit";
import type { PageServerLoad } from './$types';


export const load: PageServerLoad = async ({cookies}) => {
    const token = cookies.get('auth_token')
    const expiry = cookies.get('auth_token_expiry')
    if (!(token && expiry)) {
        throw redirect(302, '/login')
    }
}
