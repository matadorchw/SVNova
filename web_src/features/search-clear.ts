export function initSearchClear() {
 for (const box of Array.from(document.querySelectorAll<HTMLElement>('[data-search-clear]'))) {
  if (box.hasAttribute('data-search-clear-ready')) continue;
  const input=box.querySelector<HTMLInputElement>('input');
  const button=box.querySelector<HTMLButtonElement>('.search-clear-button');
  if (!input || !button) continue;
  box.setAttribute('data-search-clear-ready','');
  const update=()=>{button.hidden=input.value.length===0;};
  input.addEventListener('input',update);
  input.addEventListener('change',update);
  input.addEventListener('focus',update);
  input.form?.addEventListener('reset',()=>queueMicrotask(update));
  button.addEventListener('click',event=>{
   event.preventDefault();
   input.value='';
   input.dispatchEvent(new Event('input',{bubbles:true}));
   input.focus();
   if (box.dataset.searchClear==='submit' && input.form) {
    for (const field of Array.from(input.form.querySelectorAll<HTMLInputElement>('input[name="page"],input[name="repo_page"]'))) field.value='1';
    try { sessionStorage.setItem('svnova.search-clear-focus',new URL(input.form.action,location.href).pathname+':'+input.name); } catch { /* Clearing still works when browser storage is unavailable. */ }
    input.form.requestSubmit();
   }
  });
  update();
  try {
   if (sessionStorage.getItem('svnova.search-clear-focus')===location.pathname+':'+input.name) {
    sessionStorage.removeItem('svnova.search-clear-focus');input.focus();
   }
  } catch { /* Focus restoration is optional when browser storage is unavailable. */ }
 }
}
