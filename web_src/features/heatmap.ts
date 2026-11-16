// Adapted from Gitea web_src/js/features/heatmap.ts (MIT).
import {createApp, type App} from 'vue';
import ActivityHeatmap from '../components/ActivityHeatmap.vue';
type HeatmapResponse={heatmapData:Array<[number,number]>;totalContributions:number;pending:boolean;failed:boolean;today:string};
export async function initHeatmap(){
 const el=document.querySelector<HTMLElement>('#user-heatmap');if(!el)return;
 const url=el.dataset.heatmapUrl;if(!url)return;
 let needsFeedUpdate=el.dataset.activityPending==='true';
 let view:App|undefined;let stopped=false;let signature='';let timer:ReturnType<typeof setTimeout>|undefined;
 window.addEventListener('pagehide',()=>{stopped=true;if(timer)clearTimeout(timer);view?.unmount();},{once:true});
 const update=async()=>{
  try{
   const response=await fetch(url);if(response.status===401){location.assign('/user/login');return;}if(!response.ok)throw Error('活动图加载失败');
   const data=await response.json() as HeatmapResponse;if(stopped)return;
   const fresh=JSON.stringify(data);
   const changed=fresh!==signature;
   if(changed){signature=fresh;
    const values=data.heatmapData.map(([timestamp,count])=>{const day=new Intl.DateTimeFormat('sv-SE',{timeZone:'Asia/Shanghai',year:'numeric',month:'2-digit',day:'2-digit'}).format(new Date(timestamp*1000));return {date:new Date(`${day}T00:00:00`),count};});
    const locale={textTotalContributions:`当前可访问仓库近一年共 ${data.totalContributions.toLocaleString('zh-CN')} 次迁移／同步活动${data.pending?'（有任务进行中）':''}`,heatMapLocale:{months:Array.from({length:12},(_,i)=>`${i+1}月`),days:['周日','周一','周二','周三','周四','周五','周六'],on:' · ',more:'更多',less:'更少'},tooltipUnit:'次迁移／同步活动',noDataText:'无活动'};
    view?.unmount();view=createApp(ActivityHeatmap,{values,locale,today:data.today});view.mount(el);el.classList.remove('is-loading');
   }
   if(changed||needsFeedUpdate||data.pending||data.failed){
    const page=await fetch(location.href,{headers:{Accept:'text/html'}});if(page.redirected&&new URL(page.url).pathname==='/user/login'){location.assign('/user/login');return;}
    if(page.ok&&!stopped){const doc=new DOMParser().parseFromString(await page.text(),'text/html');for(const id of ['activity-feed','activity-index-status']){const next=doc.getElementById(id),current=document.getElementById(id);if(next&&current&&!current.contains(document.activeElement))current.replaceWith(next);}}
    needsFeedUpdate=data.pending||data.failed;

   }
   if(!stopped)timer=setTimeout(update,data.pending?5000:15000);
  }catch{if(stopped)return;const status=document.getElementById('activity-index-status');if(status)status.textContent='活动统计暂不可用，稍后自动重试';timer=setTimeout(update,15000);}
 };
 await update();
}
