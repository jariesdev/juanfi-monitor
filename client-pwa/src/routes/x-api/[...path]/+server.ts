import type { RequestHandler, RouteParams } from './$types';
import { VITE_API_URL } from '$env/static/private';
import type { MaybePromise } from '@sveltejs/kit/src/types/private';

type RouteParams2 = RouteParams & { url: URL; request: Request };

const sendApiRequest = async (r: RouteParams2): Promise<Response> => {
	let baseApiUrl: string = VITE_INTERNAL_API;
	// remove the leading slash
	baseApiUrl = baseApiUrl.replace(/\/$/, '');

	// path and search params
	const pathname = r.url.pathname.replace('/x-api/', '/');
	const url: URL = new URL(baseApiUrl);
	url.pathname = pathname;
	url.search = r.url.search;

	// request object
	const req: Request = new Request(url, { method: 'GET', body: r.request.body });

	// send request
	return fetch(req);
};

// This handler will respond to GET, POST, PUT, PATCH, DELETE, etc.
export const fallback: RequestHandler =  (r: RouteParams2): MaybePromise<Response> => {
	return  sendApiRequest(r);
};
