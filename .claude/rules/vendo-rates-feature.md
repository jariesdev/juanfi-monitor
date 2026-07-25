# Vendo Rates feature

`vendo_rates` table (`internal/models/vendo_rate.go`) stores per-vendo pricing tiers, imported from or pushed to the device via the [[juanfi-device-api]]. Design:

- `vendo_id` is nullable — rows with `vendo_id = NULL` are the shared **default rate-plan template**, managed only by admins (`PermUsers`/`assignedVendoIDs(c) == nil`), separate from per-vendo rates which any user with access to that vendo can manage.
- **ACL** — all rate routes live under a `RequirePermission(models.PermRates)` (`"rates"`) group in `app.go`, so the `rates` permission is the entry ticket for the whole feature. On top of that, controllers enforce finer ACL: per-vendo ops also require `canAccessVendo`, and default-template/`set-as-default`/`apply-to-all` also require admin (`isAdmin`). Net effect: managing the default template needs **both** `rates` and `users` — the seed admin role (`AllPermissions()`) has both. Frontend mirrors this: per-vendo rates page (`/vendo/[id]/rates`) and the VendoTable "rates" row action gate on `rates`; the Settings → Default Rates page (`/settings/default-rates`) gates on `rates` **and** `users`.
- `SaveRates` (`POST /vendo-machines/:id/rates/sync`, the "Sync" button) uploads the vendo's stored rows to the device; `SetAsDefault` accepts a `mode` body of `copy` (append to the default template) or `replace` (default).
- All copy/import operations (`ImportFromMachine`, `SetAsDefault` replace mode, `ApplyToAll` in `vendo_rate_controller.go`) use **replace semantics** — they delete the target's existing rows before inserting the new set.
- `POST /vendo-rates/apply-to-all` requires an explicit, non-empty `vendo_ids` list — there's no "omit it to mean all" behavior; the frontend's vendo picker defaults to none selected with a "Select all" toggle.
- Frontend: per-vendo CRUD at `/vendo/[id]/rates`, global default-template admin page at `/settings/default-rates`.
