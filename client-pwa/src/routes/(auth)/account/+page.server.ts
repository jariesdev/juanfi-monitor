import { fail } from '@sveltejs/kit';
import type { Actions } from './$types';

export const actions: Actions = {
	default: async ({ request, fetch }) => {
		const data = await request.formData();
		const currentPassword = data.get('current_password');
		const newPassword = data.get('new_password');
		const confirmPassword = data.get('confirm_password');

		if (
			typeof currentPassword !== 'string' ||
			typeof newPassword !== 'string' ||
			typeof confirmPassword !== 'string' ||
			!currentPassword ||
			!newPassword ||
			!confirmPassword
		) {
			return fail(400, { error: 'All fields are required.' });
		}

		if (newPassword !== confirmPassword) {
			return fail(400, { error: 'New passwords do not match.' });
		}

		if (newPassword.length < 6) {
			return fail(400, { error: 'New password must be at least 6 characters.' });
		}

		const res = await fetch('/x-api/users/me/password', {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ current_password: currentPassword, new_password: newPassword })
		});

		if (!res.ok) {
			const body = await res.json().catch(() => ({}));
			return fail(res.status, { error: body.detail ?? 'Failed to update password.' });
		}

		return { success: true };
	}
};
