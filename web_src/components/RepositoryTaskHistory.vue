<script setup lang="ts">
import {ref,onMounted,onBeforeUnmount,nextTick} from 'vue';
import {init,use} from 'echarts/core';
import type {EChartsType} from 'echarts/core';
import {CustomChart,LineChart} from 'echarts/charts';
import {GraphicComponent,GridComponent,TooltipComponent,LegendComponent,DataZoomSliderComponent} from 'echarts/components';
import {SVGRenderer} from 'echarts/renderers';
import Icon from '../Icon.vue';
import {repositoryTaskOption,taskKinds} from '../features/repository-task-options';
import type {TaskHistory} from '../features/repository-task-options';
use([CustomChart,LineChart,GraphicComponent,GridComponent,TooltipComponent,LegendComponent,DataZoomSliderComponent,SVGRenderer]);
const props=defineProps<{url:string;retryUrl:string;initForms:()=>void}>();
const data=ref<TaskHistory>(),error=ref(''),loading=ref(false),chartElement=ref<HTMLDivElement>();
let chart:EChartsType|undefined,observer:ResizeObserver|undefined,timer:ReturnType<typeof setTimeout>|undefined,controller:AbortController|undefined;
let stopped=false,serial=0;
let zoom={start:0,end:100};
const selected=ref<Record<string,boolean>>(Object.fromEntries(taskKinds.map(kind=>[kind.label,true])));
function toggleKind(label:string){selected.value[label]=selected.value[label]===false;zoom={start:0,end:100};render();}
function render(){if(!chart||!chartElement.value)return;const style=getComputedStyle(chartElement.value);chart.setOption(repositoryTaskOption(data.value?.points??[],zoom,selected.value,name=>style.getPropertyValue(name).trim(),chart.getWidth()),{notMerge:true});}
async function load(page=data.value?.page??1){
 const version=++serial;controller?.abort();if(timer)clearTimeout(timer);controller=new AbortController();loading.value=true;
 try{
  const url=new URL(props.url,location.origin);url.searchParams.set('page',String(page));
  const response=await fetch(url,{signal:controller.signal,headers:{Accept:'application/json'}});
  if(version!==serial||stopped)return;
  if(response.status===401||(response.redirected&&new URL(response.url).pathname==='/user/login')){stopped=true;data.value=undefined;chart?.clear();location.assign('/user/login');return;}
  if(response.status===403||response.status===404){stopped=true;data.value=undefined;chart?.clear();error.value='仓库不存在或访问权限已变化，请刷新页面。';return;}
  if(!response.ok)throw Error();
  const value=await response.json() as TaskHistory;
  if(version!==serial||stopped)return;
  if(data.value&&data.value.repositoryState!==value.repositoryState){location.reload();stopped=true;return;}
  data.value=value;error.value='';await nextTick();props.initForms();render();
  if(page!==1||new URL(location.href).searchParams.has('page')){const address=new URL(location.href);address.searchParams.set('page',String(value.page));history.replaceState(null,'',address);}
 }catch{if(version===serial&&!stopped)error.value='任务记录暂不可用，正在重试…';}
 finally{if(version===serial){loading.value=false;if(!stopped)timer=setTimeout(()=>void load(),data.value?.active?5000:30000);}}
}
onMounted(()=>{
 if(chartElement.value){chart=init(chartElement.value,undefined,{renderer:'svg'});chart.on('datazoom',event=>{const value=event as {start?:number;end?:number};if(value.start!==undefined&&value.end!==undefined)zoom={start:value.start,end:value.end};});chart.on('legendselectchanged',event=>{selected.value={...selected.value,...(event as {selected:Record<string,boolean>}).selected};render();});observer=new ResizeObserver(()=>{chart?.resize();render();});observer.observe(chartElement.value);render();}
 const page=Number(new URL(location.href).searchParams.get('page'));void load(Number.isInteger(page)&&page>0?page:1);
});
onBeforeUnmount(()=>{stopped=true;serial++;controller?.abort();if(timer)clearTimeout(timer);observer?.disconnect();chart?.dispose();});
</script>
<template>
 <div class="repository-task-history" :aria-busy="loading">
  <p v-if="error" class="ui warning message" role="alert">{{error}} <button v-if="!stopped" type="button" class="ui basic button action-small" :disabled="loading" @click="load()">重试</button></p>
  <section class="task-summary" aria-label="仓库任务统计">
   <div class="ui segment"><span>任务总数</span><strong>{{data?.summary.total??'—'}}</strong></div>
   <div class="ui segment"><span>成功率</span><strong>{{data?.summary.successRate==null?'—':`${data.summary.successRate.toFixed(1)}%`}}</strong><small class="muted">成功／已结束任务</small></div>
   <div class="ui segment"><span>版本跨度</span><strong>{{data?.summary.revisionSpan??'—'}}</strong></div>
   <div class="ui segment"><span>累计新增修订</span><strong>{{data?.summary.added??'—'}}</strong><small class="muted">仅统计成功同步</small></div>
  </section>
  <section class="ui segments task-trend" aria-label="仓库任务趋势">
   <div class="ui secondary segment task-trend-heading"><h4>任务趋势</h4></div>
   <div class="ui segment task-trend-filters"><fieldset aria-label="任务类型"><label v-for="kind in taskKinds" :key="kind.key"><input type="checkbox" :checked="selected[kind.label]!==false" @change="toggleKind(kind.label)"><span v-if="kind.key!=='sync'" class="task-kind-line" :style="{borderColor:`var(${kind.color})`,borderTopStyle:kind.lineStyle}" aria-hidden="true"></span>{{kind.label}}</label></fieldset></div>
   <div class="ui segment"><div ref="chartElement" class="task-history-chart" aria-label="版本号、耗时和任务事件竖线"></div><p v-if="!data" class="muted">{{loading?'正在加载任务记录…':'暂无任务数据'}}</p><p v-else-if="!data.summary.total" class="muted">暂无任务</p></div>
  </section>
  <section class="task-history-table" data-task-status>
   <h4 class="ui top attached header">任务记录<span v-if="data">（{{data.summary.total}}）</span></h4>
   <div class="ui attached table segment table-scroll"><table class="ui very basic unstackable table"><thead><tr><th>编号</th><th>任务类型</th><th>状态</th><th>阶段</th><th>开始时间</th><th>结束时间</th><th>结果</th><th>操作</th></tr></thead><tbody>
    <tr v-for="task in data?.items??[]" :key="task.id"><td>#{{task.id}}</td><td>{{task.kind}}</td><td>{{task.state}}</td><td>{{task.phase}}<progress v-if="task.indeterminate" aria-label="任务进行中"></progress></td><td class="task-date">{{task.started}}</td><td class="task-date">{{task.finished}}</td><td>{{task.message}}</td><td><form v-if="task.retry" class="ui form" :data-api="retryUrl"><button class="action-icon-button" title="重试任务" aria-label="重试任务"><Icon name="issueReopened"/></button></form></td></tr>
    <tr v-if="!data?.items.length"><td colspan="8">{{loading&&!data?'正在加载任务记录…':'暂无任务'}}</td></tr>
   </tbody></table></div>
   <nav v-if="data&&data.pages>1" class="ui pagination menu task-pagination" aria-label="任务记录分页"><button class="item" :disabled="loading||data.page<=1" @click="load(data.page-1)">上一页</button><span class="item">{{data.page}} / {{data.pages}}</span><button class="item" :disabled="loading||data.page>=data.pages" @click="load(data.page+1)">下一页</button></nav>
  </section>
 </div>
</template>
<style scoped>
.repository-task-history{display:grid;gap:24px;min-width:0}
.task-summary{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:16px}
.task-summary .ui.segment{margin:0;display:flex;flex-direction:column;gap:8px;min-width:0}
.task-summary strong{font-size:24px;overflow-wrap:anywhere;line-height:1.3}
.task-summary small{font-size:12px}
.task-trend.ui.segments{margin:0;min-width:0}
.task-trend-heading{display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:16px}
.task-trend-heading h4{margin:0}
.task-trend-filters{padding:16px !important;display:grid;grid-template-columns:minmax(0,1fr);gap:16px}
.task-kind-line{width:16px;border-top-width:2px;flex:none}
.task-trend-filters fieldset{border:0;margin:0;padding:0;display:flex;align-items:center;flex-wrap:wrap;gap:12px 20px;min-width:0}
.task-trend-filters label{display:inline-flex;flex:0 0 auto;width:auto;white-space:nowrap;gap:8px;align-items:center;margin:0}
.task-trend-filters input{margin:0;flex:0 0 auto;width:auto}
.task-history-chart{width:100%;height:420px;min-width:0}
.task-date{white-space:nowrap}
.task-history-table{min-width:0}
.task-pagination{margin-top:16px;flex-wrap:wrap}
.task-pagination button{font:inherit;color:inherit;background:transparent;cursor:pointer}
.task-pagination button:disabled{cursor:default;opacity:.45}
@media(max-width:767.98px){.task-trend-filters{grid-template-columns:minmax(0,1fr)}.task-summary{grid-template-columns:repeat(2,minmax(0,1fr))}.task-summary strong{font-size:20px}.task-history-chart{height:380px}}
</style>
