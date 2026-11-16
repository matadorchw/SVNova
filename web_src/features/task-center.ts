export function initTaskCenter() {
 const element=document.querySelector<HTMLElement>('[data-task-center]');
 if(!element)return;
 const root=element;
 const total=document.querySelector<HTMLElement>('[data-task-center-total]');
 const error=document.querySelector<HTMLElement>('[data-task-center-error]');
 let timer:ReturnType<typeof setTimeout>|undefined;
 let busy=false;
 const active=()=>root.querySelector<HTMLElement>('[data-task-center-active]')?.dataset.taskCenterActive==='true';
 function schedule(){if(timer)clearTimeout(timer);if(!document.hidden&&active())timer=setTimeout(()=>void refresh(),5000);}
 async function refresh(){
  if(busy||document.hidden)return;
  busy=true;
  try{
   const response=await fetch('/api/v1/tasks'+location.search,{headers:{Accept:'application/json'},cache:'no-store'});
   if(response.status===401){location.assign('/user/login');return;}
   if(!response.ok)throw new Error('refresh failed');
   const data=await response.json() as {html:string;page:number;total:number};
   root.innerHTML=data.html;
   if(total)total.textContent=String(data.total);
   const target=new URL(location.href);target.searchParams.set('page',String(data.page));history.replaceState(null,'',target);
   if(error)error.hidden=true;
  }catch{root.replaceChildren();if(total)total.textContent='—';if(error)error.hidden=false;}
  finally{busy=false;schedule();}
 }
 document.addEventListener('visibilitychange',()=>{if(timer)clearTimeout(timer);if(!document.hidden)void refresh();});
 const form=document.querySelector<HTMLFormElement>('.task-center-filters');
 form?.querySelectorAll<HTMLInputElement>('input[type=radio]').forEach(radio=>radio.addEventListener('change',()=>form.requestSubmit()));
 schedule();
}
