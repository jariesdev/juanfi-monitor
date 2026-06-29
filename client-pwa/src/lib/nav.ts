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
	{ href: '/account', icon: 'user', label: 'Account', requiredPermission: 'account' }
];
