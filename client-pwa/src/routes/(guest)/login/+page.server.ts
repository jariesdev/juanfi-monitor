import {fail, redirect} from '@sveltejs/kit'
import type {Action, Actions, PageServerLoad} from './$types'
import { NODE_ENV } from '$env/static/private';
import { baseApiUrl } from '$lib/env';

export const load: PageServerLoad = async ({cookies}) => {
    const token = cookies.get('auth_token')
    const expiry = cookies.get('auth_token_expiry')
    if (token && expiry) {
        throw redirect(302, '/home')
    }
}

const defaultAction: Action = async ({cookies, request, fetch, locals}) => {
    const data = await request.formData()
    const username = data.get('username')
    const password = data.get('password')
// validation
    if (
        typeof username !== 'string' ||
        typeof password !== 'string' ||
        !username ||
        !password
    ) {
        return fail(400, {invalid: true})
    }

    const formData: FormData = new FormData()
    formData.append('username', username)
    formData.append('password', password)

    try {
        // try login to API
        const request: Request = new Request(`${baseApiUrl}/token`, {method: 'POST', body: formData})
        let response = await fetch(request)
            .then((response) => {
                if (response.ok) {
                    return response.json()
                }

                throw new Error(response.statusText)
            })

        // set auth token cookie
        cookies.set('auth_token', response.access_token, {
            // send cookie for every page
            path: '/',
            // server side only cookie, so you can't use `document.cookie`
            httpOnly: true,
            // only requests from same site can send cookies
            // https://developer.mozilla.org/en-US/docs/Glossary/CSRF
            sameSite: 'strict',
            // only sent over HTTPS in production
            secure: NODE_ENV === 'production',
            // set cookie to expire after a month
            maxAge: 60 * 60 * 24 * 30,
        })
        // set auth token cookie
        cookies.set('auth_token_expiry', response.expiry, {
            // send cookie for every page
            path: '/',
            // server side only cookie, so you can't use `document.cookie`
            httpOnly: true,
            // only requests from same site can send cookies
            // https://developer.mozilla.org/en-US/docs/Glossary/CSRF
            sameSite: 'strict',
            // only sent over HTTPS in production
            secure: NODE_ENV === 'production',
            // set cookie to expire after a month
            maxAge: 60 * 60 * 24 * 30,
        })

        // if `user` exists set `events.local`
        if (response.user) {
            locals.user = {
                id: response.user.id,
                name: response.user.username,
                role: null,
            }
        }
    } catch (e) {
        return fail(400, {message: e.message, invalid: true})
    }

    // redirect the user
    throw redirect(302, '/home')
}

export const actions: Actions = {default: defaultAction}
