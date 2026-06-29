
// Svelte action: animate an integer counter on mount; update() animates to new value
import { CountUp, type CountUpOptions } from 'countup.js';

export const countupInt = function (node: HTMLElement, value: number, options?: CountUpOptions) {
	const counter = new CountUp(node, value, options);
	counter.start();

	return {
		update(newValue: number) {
			counter.update(newValue);
		}
	};
}