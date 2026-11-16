import {createApp} from 'vue';
export async function initActivityTrend(){
 const el=document.getElementById('activity-trend');
 const source=el?.dataset.heatmapUrl;
 if(!el||!source)return;
 let stopped=false;
 let view:ReturnType<typeof createApp>|undefined;
 window.addEventListener('pagehide',event=>{if(!event.persisted){stopped=true;view?.unmount();}});
 try{
  const {default:ActivityTrend}=await import('../components/ActivityTrend.vue');
  if(stopped)return;
  const org=new URL(source,location.origin).searchParams.get('org')??'0';
  view=createApp(ActivityTrend,{org});view.mount(el);
 }catch{if(stopped)return;el.textContent='活动趋势加载失败，请刷新页面重试。';}
}
