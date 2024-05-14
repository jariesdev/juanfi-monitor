/** @type {import('../../../../../../.svelte-kit/types/src/routes').PageLoad} */
export function load({ params }: any) {
	return {
		id: params.id
	};
}
