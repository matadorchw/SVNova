// Adapted from Gitea web_src/js/features/admin/users.ts (MIT), pinned baseline.
export function initAdminUsers() {
 for (const form of Array.from(document.querySelectorAll<HTMLFormElement>('#user-list-search-form,#notification-list-search-form'))) {
 for (const radio of Array.from(form.querySelectorAll<HTMLInputElement>('input[type="radio"]'))) {
  radio.addEventListener('change', () => form.requestSubmit());
 }
 form.querySelector('[data-user-filter-reset]')?.addEventListener('click', () => {
  for (const radio of Array.from(form.querySelectorAll<HTMLInputElement>('input[type="radio"]'))) radio.checked = false;
  form.requestSubmit();
 });
 const sort = form.querySelector<HTMLInputElement>('input[name="sort"]');
 for (const button of Array.from(form.querySelectorAll<HTMLButtonElement>('[data-user-sort]'))) {
  button.addEventListener('click', () => {
   if (!sort || !button.dataset.userSort) return;
   sort.value = button.dataset.userSort;
   form.requestSubmit();
  });
 }
}
}
