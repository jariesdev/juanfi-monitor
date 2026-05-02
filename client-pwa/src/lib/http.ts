export const get = (url: string): Promise<Response> => {
	return fetch(new Request(`/x-api${url}`, { method: 'GET' }));
};

export const post = (url: string, body: FormData | unknown): Promise<Response> => {
	return fetch(new Request(`/x-api${url}`, { method: 'POST', body: body as BodyInit }));
};
