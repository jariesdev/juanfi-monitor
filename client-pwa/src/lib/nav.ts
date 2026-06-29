export interface NavChild {
	href: string;
	label: string;
}

export interface NavItem {
	href?: string;
	icon: string;
	label: string;
	children?: NavChild[];
	requiredPermission?: string;
}

export const navItems: NavItem[] = [
	{ href: '/home', icon: 'home', label: 'Home' },
	{ href: '/vendo', icon: 'server', label: 'Vendo' },
	{
		icon: 'cart',
		label: 'Sales',
		children: [
			{ href: '/sales', label: 'Sales Records' },
			{ href: '/withdrawals', label: 'Withdrawals' }
		]
	},
	{ href: '/logs', icon: 'list', label: 'Logs' },
	{
		icon: 'users',
		label: 'Users',
		requiredPermission: 'users',
		children: [
			{ href: '/users', label: 'Users' },
			{ href: '/roles', label: 'Roles' }
		]
	},
	{ href: '/account', icon: 'user', label: 'Account' }
];
