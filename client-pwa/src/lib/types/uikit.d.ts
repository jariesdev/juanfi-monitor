import 'svelte/elements';

declare module 'svelte/elements' {
	interface HTMLAttributes<T extends EventTarget> {
		'uk-icon'?: string;
		'uk-modal'?: string | boolean;
		'uk-toggle'?: string;
		'uk-alert'?: string | boolean;
		'uk-spinner'?: string | boolean;
		'uk-close'?: string | boolean;
		'uk-navbar'?: string | boolean;
		'uk-dropdown'?: string | boolean;
		'uk-offcanvas'?: string | boolean;
		'uk-sticky'?: string | boolean;
		'uk-tooltip'?: string;
		'uk-grid'?: string | boolean;
	}
}
