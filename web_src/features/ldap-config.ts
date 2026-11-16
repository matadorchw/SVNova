// LDAP options adapted from Gitea admin/auth source form behavior (MIT).
export function initLDAPOptions() {
 const toggle=document.querySelector<HTMLInputElement>('#auth-paged');
 const field=document.querySelector<HTMLElement>('[data-ldap-page-size]');
 const input=field?.querySelector<HTMLInputElement>('input');
 if(!toggle||!field||!input)return;
 const update=()=>{field.hidden=!toggle.checked;input.disabled=!toggle.checked;};
 toggle.addEventListener('change',update);update();
}

type CleanupPreview={requiresConfirmation:boolean;deleteUsers:number;sessions:number;memberships:number;teamMemberships:number;token:string};
export async function confirmLDAPChange(endpoint:string,body:Record<string,string|number|boolean>,csrf:string):Promise<boolean> {
 const response=await fetch(`${endpoint}/preview`,{method:'POST',headers:{'Content-Type':'application/json','X-CSRF-Token':csrf},body:JSON.stringify(body)});
 if(response.status===401){location.assign('/user/login');return false;}
 if(!response.ok)throw new Error(await response.text());
 const plan=await response.json() as CleanupPreview;
 if(!plan.requiresConfirmation)return true;
 const dialog=document.createElement('dialog');dialog.className='auth-cleanup-dialog';
 const heading=document.createElement('h3');heading.textContent='确认更改 LDAP 配置';
 const description=document.createElement('p');description.textContent=`此操作将永久删除 ${plan.deleteUsers} 个相关 LDAP 用户、${plan.sessions} 个会话、${plan.memberships} 条组织成员关系和 ${plan.teamMemberships} 条团队成员关系，包括用户邮箱及身份绑定。`;
 const keep=document.createElement('p');keep.textContent='本地账号（包括 admin）、组织、团队、仓库和同步历史保留。清理后无法通过页面撤销。';
 const from=document.createElement('p');from.textContent='修改 LDAP 配置前将清理绑定的 LDAP 用户。';
 const label=document.createElement('label');label.textContent='请输入“确认清理”后继续';
 const input=document.createElement('input');input.type='text';input.autocomplete='off';input.setAttribute('aria-label','确认清理');label.append(input);
 const actions=document.createElement('div');actions.className='cleanup-actions';
 const cancel=document.createElement('button');cancel.type='button';cancel.className='ui button';cancel.textContent='取消';
 const confirm=document.createElement('button');confirm.type='button';confirm.className='ui negative button';confirm.textContent='确认清理并保存';confirm.disabled=true;
 input.addEventListener('input',()=>{confirm.disabled=input.value!=='确认清理';});
 actions.append(cancel,confirm);dialog.append(heading,description,keep,from,label,actions);document.body.append(dialog);
 const accepted=await new Promise<boolean>(resolve=>{let settled=false;const finish=(value:boolean)=>{if(settled)return;settled=true;dialog.close();resolve(value);};cancel.addEventListener('click',()=>finish(false));confirm.addEventListener('click',()=>{if(input.value==='确认清理')finish(true);});dialog.addEventListener('cancel',event=>{event.preventDefault();finish(false);});dialog.addEventListener('close',()=>{if(!settled){settled=true;resolve(false);}});dialog.showModal();input.focus();});
 dialog.remove();
 if(accepted){body.ConfirmCleanup=true;body.CleanupToken=plan.token;}
 return accepted;
}
