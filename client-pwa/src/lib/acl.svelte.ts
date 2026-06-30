let permissions = $state<string[]>([]);

export function setPermissions(p: string[]) {
	permissions = p;
}

export function hasPermission(perm: string): boolean {
	return permissions.includes(perm);
}

export function hasAnyPermission(...perms: string[]): boolean {
	return perms.some((p) => permissions.includes(p));
}

export function hasAllPermissions(...perms: string[]): boolean {
	return perms.every((p) => permissions.includes(p));
}

export function isAdmin(): boolean {
	return permissions.includes('users');
}

export function getPermissions(): string[] {
	return permissions;
}
