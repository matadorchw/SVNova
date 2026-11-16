export function initNotifications(initForms:()=>void) {
 initDeliveryRefresh(initForms);
 const form=document.querySelector<HTMLFormElement>('[data-notification-settings]');
 if(form){
  for(const toggle of Array.from(form.querySelectorAll<HTMLInputElement>('[data-notification-credential]'))){
   const input=document.getElementById(toggle.dataset.notificationCredential||'');
   if(!(input instanceof HTMLInputElement))continue;
   const update=()=>{input.disabled=!toggle.checked;input.required=toggle.checked;if(!toggle.checked)input.value='';};
   toggle.addEventListener('change',update);update();
  }

 }
}

function initDeliveryRefresh(initForms:()=>void) {
 const records=document.querySelector<HTMLElement>('#notification-records');
 if(!records)return;
 const controller=new AbortController();
 const interactionActive=()=>document.hidden || !!document.querySelector('dialog[open]') || records.contains(document.activeElement);
 let previousMarkup='';
 let timer:ReturnType<typeof setTimeout>;
 const refresh=async()=>{
  try {
   // Do not replace a focused action or an open confirmation while the user is interacting.
   if(interactionActive())return;
   const response=await fetch(location.href,{headers:{Accept:'text/html'},cache:'no-store',signal:controller.signal});
   if(response.status===401 || response.redirected){location.assign('/user/login');return;}
   if(!response.ok)return;
   const page=new DOMParser().parseFromString(await response.text(),'text/html');
   const next=page.querySelector('#notification-records');
   if(!next || controller.signal.aborted || interactionActive())return;
   if(next.innerHTML!==previousMarkup){previousMarkup=next.innerHTML;records.innerHTML=next.innerHTML;initForms();}
   const total=document.querySelector('[data-notification-total]');
   const nextTotal=page.querySelector('[data-notification-total]');
   if(total && nextTotal)total.textContent=nextTotal.textContent;
  } catch(error) {
   if(!(error instanceof DOMException && error.name==='AbortError'))console.debug('Notification refresh failed');
  } finally {
   if(!controller.signal.aborted)timer=setTimeout(()=>void refresh(),2000);
  }
 };
 window.addEventListener('pagehide',()=>{clearTimeout(timer);controller.abort();},{once:true});
 void refresh();
}

export async function confirmNotificationAction(message:string, title="确认发送通知", action="确认发送"):Promise<boolean> {
 const dialog=document.createElement('dialog');dialog.className='auth-cleanup-dialog';
 const heading=document.createElement('h3');heading.textContent=title;
 const description=document.createElement('p');description.textContent=message;
 const actions=document.createElement('div');actions.className='cleanup-actions';
 const cancel=document.createElement('button');cancel.type='button';cancel.className='ui button';cancel.textContent='取消';
 const confirm=document.createElement('button');confirm.type='button';confirm.className='ui primary button';confirm.textContent=action;
 actions.append(cancel,confirm);dialog.append(heading,description,actions);document.body.append(dialog);
 const accepted=await new Promise<boolean>(resolve=>{
  let settled=false;const finish=(value:boolean)=>{if(settled)return;settled=true;dialog.close();resolve(value);};
  cancel.addEventListener('click',()=>finish(false));confirm.addEventListener('click',()=>finish(true));
  dialog.addEventListener('cancel',event=>{event.preventDefault();finish(false);});
  dialog.addEventListener('close',()=>{if(!settled){settled=true;resolve(false);}});
  dialog.showModal();cancel.focus();
 });
 dialog.remove();return accepted;
}
