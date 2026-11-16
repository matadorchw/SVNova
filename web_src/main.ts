import {initRepositoryTaskHistory} from './features/repository-task-history';
import {initSearchClear} from './features/search-clear';
import {initTaskCenter} from './features/task-center';
import {initActivityTrend} from './features/activity-trend';
import {initRepoMigrationForm} from './features/repo-migration';
import {initNotifications,confirmNotificationAction} from './features/notifications';
import {initAdminUsers} from './features/admin-users';
import {initLDAPOptions,confirmLDAPChange} from './features/ldap-config';
import {initRepoMigrationStatusChecker} from './features/repo-migrate';
import {initHeatmap} from './features/heatmap';
import {initSelectDropdowns} from './features/select-dropdown';
import './upstream/heatmap.css';
import './upstream/dropdown.css';
import './upstream/base-variables.css';
import {initHeadNavbarContentToggle} from './upstream/navbar';
import './upstream/navbar.css';
import './theme-gitea-light.css';
import './style.css';
import './upstream/flexcontainer.css';
import './upstream/segment.css';
import './upstream/button.css';
import './upstream/form.css';
import './upstream/table.css';
import './upstream/header.css';
import './upstream/message.css';
import './upstream/container.css';
import './upstream/label.css';
import './upstream/input.css';
import './upstream/menu.css';
import './upstream/normalize.css';
import './responsive.css';

// Page-specific initialization follows Gitea's feature initializer pattern.
for (const input of Array.from(document.querySelectorAll<HTMLInputElement>('input[data-name-validation]'))) {
 input.addEventListener('invalid', () => {if(input.validity.patternMismatch)input.setCustomValidity(`${input.dataset.nameValidation}不能包含空格；须以字母或数字开头，只能包含字母、数字、点、下划线和连字符，最长 80 个字符`);});
 input.addEventListener('input', () => input.setCustomValidity(''));
}
function initApiForms() {
 let busy = false;
 for (const form of Array.from(document.querySelectorAll<HTMLFormElement>('form[data-api]:not([data-initialized])'))) {
  form.dataset.initialized = 'true';
  form.addEventListener('submit', async event => {
   event.preventDefault();
   if (busy || !form.reportValidity()) return;
   const confirmation = form.querySelector<HTMLInputElement>('[data-confirm]');
   if (confirmation && confirmation.value !== confirmation.dataset.confirm) {confirmation.setCustomValidity('请输入完整仓库名称');confirmation.reportValidity();confirmation.addEventListener('input', () => confirmation.setCustomValidity(''), {once:true});return;}
   const orgName = form.querySelector<HTMLInputElement>("[data-org-name]");
   if(orgName && orgName.value !== orgName.dataset.orgName){orgName.setCustomValidity("请输入完整组织名称");orgName.reportValidity();orgName.addEventListener("input",()=>orgName.setCustomValidity(""),{once:true});return;}
   if (form.dataset.memberRemove && !await confirmNotificationAction(form.dataset.memberRemove,'确认移除成员','确认移除')) return;
   if (form.dataset.teamLeave && !await confirmNotificationAction(form.dataset.teamLeave,'确认离开团队','确认离开')) return;
   const endpoint = form.dataset.api;
   if (!endpoint) return;
   const body: Record<string, string | number | boolean> = {};
   for (const input of Array.from(form.querySelectorAll<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>('input[name],select[name],textarea[name]'))) {
    if (input.disabled || input.closest('fieldset[disabled]')) continue;
    body[input.name] = input.type === 'checkbox' ? (input as HTMLInputElement).checked : input.hasAttribute('data-number') ? Number(input.value) : input.hasAttribute('data-bool') ? input.value === 'true' : input.value;
   }
   const message = form.querySelector<HTMLElement>('[data-delete-error]') || document.querySelector<HTMLElement>('#action-message');
   const buttons = Array.from(document.querySelectorAll<HTMLButtonElement>('form[data-api] button'));
   const previous = buttons.map(button => button.disabled);
   busy = true;
   buttons.forEach(button => { button.disabled = true; });
   try {
    if(form.dataset.leaveOrganization&&!await confirmLeaveOrganization(form.dataset.leaveOrganization))return;
    if(form.dataset.notificationConfirm&&!await confirmNotificationAction(form.dataset.notificationConfirm))return;
    const token = document.querySelector<HTMLMetaElement>('meta[name="csrf-token"]')?.content;
    if(form.hasAttribute('data-ldap-config')&&!await confirmLDAPChange(endpoint,body,token||''))return;
    form.querySelectorAll<HTMLElement>('[data-sync-action]').forEach(button=>button.classList.add('is-pending'));
    const response = await fetch(endpoint, {method: form.dataset.method || 'POST', headers: {'Content-Type': 'application/json', 'X-CSRF-Token': token || ''}, body: JSON.stringify(body)});
    if (response.status === 401) { location.assign('/user/login'); return; }
    if (!response.ok) throw new Error(await response.text());
    if (form.hasAttribute('data-next-location')) {
     const target = response.headers.get('Location');
     if (!target || new URL(target,location.origin).origin !== location.origin) throw new Error('仓库已创建，请刷新仓库列表查看');
     location.assign(new URL(target,location.origin).href);return;
    }
    location.assign(form.dataset.next || location.href);
   } catch (error) {
    if (message) {message.hidden = false; message.className = 'ui negative message'; message.textContent = error instanceof Error ? error.message : '操作失败，请重试'; message.scrollIntoView({block: 'nearest'});}
   } finally {form.querySelectorAll<HTMLElement>('[data-sync-action]').forEach(button=>button.classList.remove('is-pending'));busy = false; buttons.forEach((button, i) => {button.disabled = previous[i] ?? false;});}
  });
 }
}
initApiForms();
initNotifications(initApiForms);
initAdminUsers();
initLDAPOptions();
initRepoMigrationForm();
void initSelectDropdowns();
initRepoMigrationStatusChecker(initApiForms);
void initRepositoryTaskHistory(initApiForms);



initHeadNavbarContentToggle();



function initCurrentNavigation() {
 const url = new URL(location.href);
 for (const link of Array.from(document.querySelectorAll<HTMLAnchorElement>('.tabs a'))) {
  const target = new URL(link.href);
  if (target.pathname === url.pathname && (target.searchParams.get('tab') || 'tree') === (url.searchParams.get('tab') || 'tree')) link.setAttribute('aria-current','page');
 }
}
initCurrentNavigation();

function initDropdowns() {
 document.addEventListener('click', event => {
  for (const menu of Array.from(document.querySelectorAll<HTMLDetailsElement>('details.nav-dropdown[open]'))) {
   if (event.target instanceof Node && !menu.contains(event.target)) menu.open = false;
  }
 });
 document.addEventListener('keydown', event => {
  if (event.key !== 'Escape') return;
  for (const menu of Array.from(document.querySelectorAll<HTMLDetailsElement>('details.nav-dropdown[open]'))) {menu.open = false;menu.querySelector<HTMLElement>('summary')?.focus();}
 });
}
initDropdowns();

void initHeatmap();

initActivityTrend();

function confirmLeaveOrganization(name: string): Promise<boolean> {
 const dialog = document.querySelector<HTMLDialogElement>('#leave-organization-dialog');
 const label = dialog?.querySelector<HTMLElement>('[data-organization-name]');
 const cancel = dialog?.querySelector<HTMLButtonElement>('[data-cancel]');
 const accept = dialog?.querySelector<HTMLButtonElement>('[data-accept]');
 if (!dialog || !label || !cancel || !accept) return Promise.resolve(false);
 label.textContent = name;
 return new Promise(resolve => {
  const finish = (confirmed: boolean) => { dialog.close(); cancel.removeEventListener('click',onCancel); accept.removeEventListener('click',onAccept); dialog.removeEventListener('cancel',onEscape); resolve(confirmed); };
  const onCancel = () => finish(false);
  const onAccept = () => finish(true);
  const onEscape = (event: Event) => {event.preventDefault();finish(false);};
  cancel.addEventListener('click',onCancel); accept.addEventListener('click',onAccept); dialog.addEventListener('cancel',onEscape);
  dialog.showModal(); cancel.focus();
 });
}

const deleteOrgDialog = document.querySelector<HTMLDialogElement>('#delete-organization-dialog');
document.querySelector('#open-delete-organization')?.addEventListener('click',()=>{deleteOrgDialog?.showModal();deleteOrgDialog?.querySelector<HTMLInputElement>('input')?.focus();});
deleteOrgDialog?.querySelector('[data-delete-cancel]')?.addEventListener('click',()=>deleteOrgDialog.close());

const deleteUserDialog = document.querySelector<HTMLDialogElement>('#delete-user-dialog');
document.querySelector('#open-delete-user')?.addEventListener('click',()=>{deleteUserDialog?.showModal();deleteUserDialog?.querySelector<HTMLInputElement>('input')?.focus();});
deleteUserDialog?.querySelector('[data-delete-user-cancel]')?.addEventListener('click',()=>deleteUserDialog.close());

for (const box of Array.from(document.querySelectorAll<HTMLElement>('.member-search-box'))) {
 const input=box.querySelector<HTMLInputElement>('[data-member-search]');const menu=box.querySelector<HTMLElement>('[data-member-suggestions]');
 if(!input||!menu)continue;
 const close=()=>{menu.hidden=true;input.setAttribute('aria-expanded','false');};
 const options=Array.from(menu.querySelectorAll<HTMLButtonElement>('[data-member-name]'));
 const update=()=>{const text=input.value.trim().toLocaleLowerCase();let shown=0;for(const option of options){option.hidden=shown>=20||!option.dataset.memberMatch?.toLocaleLowerCase().includes(text);if(!option.hidden)shown++;}const empty=menu.querySelector<HTMLElement>('[data-member-empty]');if(empty)empty.hidden=shown>0;menu.hidden=false;input.setAttribute('aria-expanded','true');};
 input.addEventListener('input',update);input.addEventListener('focus',update);
 for(const option of options)option.addEventListener('click',()=>{input.value=option.dataset.memberName??'';input.dispatchEvent(new Event('input',{bubbles:true}));close();input.focus();close();});
 box.addEventListener('keydown',event=>{if(event.key==='Escape'){close();input.focus();close();}else if(event.key==='ArrowDown'){event.preventDefault();if(menu.hidden)update();const visible=options.filter(o=>!o.hidden);const index=visible.indexOf(document.activeElement as HTMLButtonElement);visible[(index+1)%visible.length]?.focus();}});
 document.addEventListener('click',event=>{if(!box.contains(event.target as Node))close();});
 box.addEventListener('focusout',event=>{if(!box.contains(event.relatedTarget as Node|null))close();});
}

initTaskCenter();

import {initRepositoryHistory} from './features/repository-history';
initRepositoryHistory();

// Clear navigation feedback when a page is restored from the back-forward cache.
window.addEventListener('pageshow',()=>{
 document.querySelectorAll<HTMLElement>('[data-refresh-action]').forEach(link=>{link.classList.remove('is-pending');link.removeAttribute('aria-busy');});
});
document.addEventListener('click',event=>{
 if(event.defaultPrevented||event.button!==0||event.ctrlKey||event.metaKey||event.shiftKey||event.altKey)return;
 const link=event.target instanceof Element?event.target.closest<HTMLAnchorElement>('a[data-refresh-action]'):null;
 if(link){link.classList.add('is-pending');link.setAttribute('aria-busy','true');}
});

initSearchClear();
