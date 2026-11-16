export function initRepositoryHistory() {
 for (const row of Array.from(document.querySelectorAll<HTMLTableRowElement>('[data-history-row]'))) {
  const button = row.querySelector<HTMLButtonElement>('.history-toggle');
  const detail = document.getElementById(button?.getAttribute('aria-controls') || '');
  if (!button || !detail) continue;
  row.addEventListener('click', event => {
   if (!(event.target instanceof Element) || event.target.closest('a')) return;
   const expanded = button.getAttribute('aria-expanded') === 'true';
   button.setAttribute('aria-expanded', String(!expanded));
   detail.hidden = expanded;
  });
 }
}

