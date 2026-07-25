// Reactive viewport tier, matching this app's nav breakpoint (mobile < 768px, tablet < 1024px, else desktop)
const MOBILE_QUERY = '(max-width: 767px)';
const TABLET_QUERY = '(max-width: 1023px)';

export function viewport() {
	const mobileMql = typeof window !== 'undefined' ? window.matchMedia(MOBILE_QUERY) : undefined;
	const tabletMql = typeof window !== 'undefined' ? window.matchMedia(TABLET_QUERY) : undefined;

	let isMobile = $state(mobileMql?.matches ?? false);
	let isTablet = $state(tabletMql?.matches ?? false);

	$effect(() => {
		const onMobileChange = (e: MediaQueryListEvent) => (isMobile = e.matches);
		const onTabletChange = (e: MediaQueryListEvent) => (isTablet = e.matches);

		mobileMql?.addEventListener('change', onMobileChange);
		tabletMql?.addEventListener('change', onTabletChange);

		return () => {
			mobileMql?.removeEventListener('change', onMobileChange);
			tabletMql?.removeEventListener('change', onTabletChange);
		};
	});

	return {
		get isMobile() {
			return isMobile;
		},
		get isTablet() {
			return isTablet && !isMobile;
		},
		get isDesktop() {
			return !isTablet;
		}
	};
}
