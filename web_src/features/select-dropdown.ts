// Adapted from Gitea modules/fomantic.ts and features/common-global.ts (MIT).
import $ from 'jquery';
import {initFomanticTransition} from '../upstream/fomantic/transition';

export async function initSelectDropdowns() {
 Object.assign(window, {$, jQuery: $});
 await import('../upstream/fomantic/dropdown.js');
 const api = $.fn as typeof $.fn & {fomanticExt: Record<string, unknown>; dropdown: any};
 api.fomanticExt = {};
 initFomanticTransition();
 const {initAriaDropdownPatch} = await import('../upstream/fomantic/aria-dropdown');
 initAriaDropdownPatch();
 api.dropdown.settings.fullTextSearch = 'exact';
 for (const menu of Array.from(document.querySelectorAll<HTMLElement>('[data-menu-dropdown]'))) {
  api.dropdown.call($(menu), {transition:'none', action:'nothing'});
 }
 for (const select of Array.from(document.querySelectorAll<HTMLSelectElement>('select'))) {
  // The native select remains the source of truth for both HTML and API forms.
  const label = select.closest('label');
  if (!select.id) select.id = `select-${document.querySelectorAll('[data-select-enhanced]').length}`;
  select.dataset.selectEnhanced = 'true';
  if (label) {label.htmlFor = select.id; select.setAttribute('aria-label', label.childNodes[0]?.textContent?.trim() || '选择');}
  const required = select.required;
  select.classList.add('ui','selection','dropdown');
  if(select.hasAttribute('data-search-select'))select.classList.add('search');
  api.dropdown.call($(select), {transition:'none',forceSelection:false,match:select.hasAttribute('data-search-select')?'text':'both'});
  const container = select.closest<HTMLElement>('.ui.dropdown');
  if (!container) continue;
  select.required = required;
  select.addEventListener('invalid', event => {
   event.preventDefault();container.classList.add('error');container.setAttribute('aria-invalid','true');container.focus();
   api.dropdown.call($(container),'show');
  });
  select.addEventListener('change', () => {container.classList.remove('error');container.removeAttribute('aria-invalid');});
  if (select.getAttribute('aria-label')) container.setAttribute('aria-label', select.getAttribute('aria-label') || '选择');
  const icon = container.querySelector('.dropdown.icon');
  if (icon) {
   const svg = document.createElementNS('http://www.w3.org/2000/svg','svg');
   svg.classList.add('dropdown','icon'); svg.setAttribute('width','14');svg.setAttribute('height','14');svg.setAttribute('aria-hidden','true');
   const use = document.createElementNS(svg.namespaceURI,'use');use.setAttribute('href','/assets/img/octicons.svg#triangle-down');svg.append(use);icon.replaceWith(svg);
  }
 }
}
