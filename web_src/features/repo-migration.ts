// Adapted from Gitea features/repo-migration.ts and repo-common.ts (MIT).
export function suggestRepositoryName(source: string): string {
 let path = source.trim();
 const remote = /^https?:\/\//i.test(path);
 if (remote) {
  try {path = new URL(path).pathname;} catch {return '';}
 }
 let part = path.replace(/[\\/]+$/, '').split(/[\\/]/).pop() || '';
 if(remote){try {part = decodeURIComponent(part);} catch { /* Keep malformed escapes editable. */ }}
 if (/^[a-z]:$/i.test(part)) return '';
 return part.trim().replace(/[^-.\w]/g, '-').replace(/^[^a-z\d]+/i, '').replace(/\.{2,}/g, '.').slice(0,80).replace(/\.+$/, '');
}

export function initRepoMigrationForm() {
 const kind = document.querySelector<HTMLSelectElement>('#migration-kind');
 const source = document.querySelector<HTMLInputElement>('#clone_addr');
 const name = document.querySelector<HTMLInputElement>('#repo_name');
 let nameChanged = !!name?.value;
 name?.addEventListener('input', () => {nameChanged = name.value !== '';});
 const suggest = () => {if(source && name && !nameChanged){name.value=suggestRepositoryName(source.value);name.setCustomValidity('');}};
 source?.addEventListener('input',suggest);
 const update = () => {
  for(const section of Array.from(document.querySelectorAll<HTMLElement>('[data-kind]'))) {
   section.hidden = section.dataset.kind !== kind?.value;
   for(const input of Array.from(section.querySelectorAll<HTMLInputElement>('input')))input.disabled=section.hidden;
  }
  const help=document.querySelector<HTMLElement>('#migration-source-help');
  if(help)help.textContent=kind?.value==='local'?'服务器上的本地绝对目录路径；迁移前须停止原服务访问。':'现有 SVN 仓库的 HTTP/HTTPS 根地址；认证信息请在下方填写。';
  suggest();
 };
 kind?.addEventListener('change',update);update();
}
