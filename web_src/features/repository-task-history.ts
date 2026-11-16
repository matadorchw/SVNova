import {createApp} from 'vue';

export async function initRepositoryTaskHistory(initForms:()=>void) {
 const el=document.querySelector<HTMLElement>('[data-repository-history]');
 if(!el?.dataset.url)return;
 let stopped=false;
 let view:ReturnType<typeof createApp>|undefined;
 window.addEventListener('pagehide',event=>{if(!event.persisted){stopped=true;view?.unmount();}});
 try {
  const {default:RepositoryTaskHistory}=await import('../components/RepositoryTaskHistory.vue');
  if(stopped)return;
  view=createApp(RepositoryTaskHistory,{url:el.dataset.url,retryUrl:el.dataset.retryUrl??'',initForms});view.mount(el);
 }catch{if(!stopped)el.textContent='任务记录加载失败，请刷新页面重试。';}
}
