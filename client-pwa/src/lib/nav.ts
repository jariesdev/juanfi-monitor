export interface NavChild {
	href: string;
	label: string;
	requiredPermission?: string;
}

export interface NavItem {
	href?: string;
	icon: string;
	label: string;
	children?: NavChild[];
	requiredPermission?: string;
}

/** Returns navItems filtered to only those the user has permission to see. */
export function getVisibleNavItems(permissions: string[]): NavItem[] {
	return navItems
		.map((item) => {
			if (!item.children) return item;
			const visibleChildren = item.children.filter(
				(c) => !c.requiredPermission || permissions.includes(c.requiredPermission)
			);
			return { ...item, children: visibleChildren };
		})
		.filter((item) => {
			if (item.requiredPermission && !permissions.includes(item.requiredPermission)) return false;
			if (item.children) return item.children.length > 0;
			return true;
		});
}

export const navItems: NavItem[] = [
	{ href: '/home', icon: 'home', label: 'Home', requiredPermission: 'dashboard' },
	{ href: '/vendo', icon: 'server', label: 'Vendo', requiredPermission: 'vendos' },
	{
		icon: 'cart',
		label: 'Sales',
		children: [
			{ href: '/sales', label: 'Sales Records', requiredPermission: 'sales' },
			{ href: '/withdrawals', label: 'Withdrawals', requiredPermission: 'withdrawals' }
		]
	},
	{ href: '/logs', icon: 'list', label: 'Logs', requiredPermission: 'logs' },
	{
		icon: 'users',
		label: 'Users',
		requiredPermission: 'users',
		children: [
			{ href: '/users', label: 'Users' },
			{ href: '/roles', label: 'Roles' }
		]
	},
	{
		icon: 'settings',
		label: 'Settings',
		requiredPermission: 'settings',
		children: [
			{ href: '/settings/default-rates', label: 'Default Rates', requiredPermission: 'rates' }
		]
	},
	{ href: '/account', icon: 'user', label: 'Account', requiredPermission: 'account' }
];
