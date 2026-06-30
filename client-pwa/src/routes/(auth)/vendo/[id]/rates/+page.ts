/** @type {import('../../../../../../../.svelte-kit/types/src/routes').PageLoad} */
export function load({ params, data }: any) {
	return {
		...data,
		id: params.id
	};
}
