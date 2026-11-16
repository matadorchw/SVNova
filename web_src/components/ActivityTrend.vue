<script setup lang="ts">
import Icon from '../Icon.vue';
import {computed,onMounted,onBeforeUnmount,ref,watch} from 'vue';
import {init,use} from 'echarts/core';
import type {EChartsType} from 'echarts/core';
import {BarChart} from 'echarts/charts';
import {GridComponent,LegendComponent,TooltipComponent,DataZoomSliderComponent} from 'echarts/components';
import {SVGRenderer} from 'echarts/renderers';
import {trendOption,trendStates} from '../features/activity-trend-options';
import type {TrendDay} from '../features/activity-trend-options';
use([BarChart,GridComponent,LegendComponent,TooltipComponent,DataZoomSliderComponent,SVGRenderer]);
const props=defineProps<{org:string}>();
const kinds=[{key:'local',label:'本地仓库迁移'},{key:'mirror',label:'镜像仓库迁移'},{key:'sync',label:'镜像仓库同步'}];
const today=ref(new Intl.DateTimeFormat('sv-SE',{timeZone:'Asia/Shanghai',year:'numeric',month:'2-digit',day:'2-digit'}).format(new Date()));
function shift(date:string,n:number){const d=new Date(`${date}T00:00:00Z`);d.setUTCDate(d.getUTCDate()+n);return d.toISOString().slice(0,10);}
const kind=ref('sync'),period=ref('30');
const to=computed(()=>today.value),from=computed(()=>shift(today.value,1-Number(period.value)));
const days=ref<TrendDay[]>([]),loading=ref(false),error=ref(''),hasData=ref(false);
const chartElement=ref<HTMLDivElement>();
const selected=ref<Record<string,boolean>>({});
const allHidden=computed(()=>trendStates.every(state=>selected.value[state.label]===false));
const label=computed(()=>kinds.find(item=>item.key===kind.value)?.label??'');
let chart:EChartsType|undefined,observer:ResizeObserver|undefined;
let zoom={start:0,end:100};
let controller:AbortController|undefined,timer:ReturnType<typeof setTimeout>|undefined,serial=0,stopped=false;
function render(){
 if(!chart||!chartElement.value)return;
 const style=getComputedStyle(chartElement.value);
 chart.setOption(trendOption(days.value,label.value,selected.value,zoom,name=>style.getPropertyValue(name).trim()));
}
async function load(){
 const version=++serial;controller?.abort();if(timer)clearTimeout(timer);
 controller=new AbortController();loading.value=true;error.value='';
 try{
  const query=new URLSearchParams({org:props.org,kind:kind.value,from:from.value,to:to.value});
  const response=await fetch(`/api/v1/activity/trend?${query}`,{signal:controller.signal});
  if(response.status===401){location.assign('/user/login');return;}
  if(!response.ok)throw Error();
  const data=await response.json() as {days:TrendDay[];pending:boolean;today:string};
  if(version!==serial||stopped)return;
  days.value=data.days;hasData.value=true;today.value=data.today;render();
  timer=setTimeout(load,data.pending?5000:15000);
 }catch{if(version!==serial||stopped)return;error.value='活动趋势暂不可用，请重试';timer=setTimeout(load,15000);}
 finally{if(version===serial)loading.value=false;}
}
watch([kind,from,to],()=>{zoom={start:0,end:100};hasData.value=false;days.value=[];render();void load();});
onMounted(()=>{
 if(!chartElement.value)return;
 chart=init(chartElement.value,undefined,{renderer:'svg'});
 chart.on('legendselectchanged',event=>{selected.value=(event as {selected:Record<string,boolean>}).selected;render();});
 chart.on('datazoom',event=>{
  const value=event as {start?:number;end?:number};
  if(value.start!==undefined&&value.end!==undefined)zoom={start:value.start,end:value.end};
 });
 observer=new ResizeObserver(()=>chart?.resize());observer.observe(chartElement.value);
 render();void load();
});
onBeforeUnmount(()=>{stopped=true;serial++;controller?.abort();if(timer)clearTimeout(timer);observer?.disconnect();chart?.dispose();chart=undefined;});
</script>
<template>
 <section class="activity-trend" aria-label="日活动趋势">
  <div class="trend-heading"><h3>日活动趋势</h3></div>
  <div class="trend-visible-filters">
   <fieldset aria-label="活动类型"><div class="trend-filter-options"><label v-for="item in kinds" :key="item.key"><input v-model="kind" type="radio" name="trend-kind" :value="item.key">{{item.label}}</label></div></fieldset>
   <fieldset aria-label="时间范围"><div class="trend-filter-options"><label v-for="item in [{key:'7',label:'近7天'},{key:'30',label:'近30天'},{key:'90',label:'近90天'}]" :key="item.key"><input v-model="period" type="radio" name="trend-period" :value="item.key">{{item.label}}</label></div></fieldset>
  </div>
  <div class="trend-content">
   <span class="muted">{{label}} · {{from}} 至 {{to}}</span>
   <p v-if="error" role="alert">{{error}} <button type="button" class="ui basic button action-text-button" :disabled="loading" @click="load"><Icon name="issueReopened"/><span>重试</span></button></p>
   <div class="trend-chart-wrapper" :aria-busy="loading">
    <div ref="chartElement" class="trend-chart" aria-label="按状态堆积的每日活动次数"></div>
    <p v-if="loading && !hasData" class="trend-chart-message" role="status">正在加载活动趋势…</p>
    <p v-else-if="hasData && allHidden" class="trend-chart-message" role="status">请选择至少一个状态</p>
   </div>
  </div>
 </section>
</template>
<style scoped>
.activity-trend{margin-block:0 16px;border:1px solid var(--color-secondary);border-radius:var(--border-radius);min-width:0;overflow:hidden}
.trend-heading{padding:16px;background:var(--color-box-header);border-bottom:1px solid var(--color-secondary)}
.trend-heading h3{margin:0}
.trend-visible-filters{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:16px;padding:16px;border-bottom:1px solid var(--color-secondary)}
.trend-visible-filters fieldset{border:0;padding:0;margin:0;min-width:0}
.trend-filter-options{display:flex;flex-wrap:wrap;gap:12px 20px}
.trend-filter-options label{margin:0;display:inline-flex;align-items:center;gap:8px;cursor:pointer;white-space:nowrap;flex:0 0 auto;width:auto}
.trend-filter-options input{margin:0;flex:0 0 auto;width:auto}
.trend-content{display:grid;gap:16px;padding:16px;min-width:0}
.trend-content p{margin:0}
.trend-chart-wrapper{position:relative;min-width:0}
.trend-chart{width:100%;height:360px;min-width:0}
.trend-chart-message{position:absolute;top:45%;inset-inline:0;text-align:center;pointer-events:none;color:var(--color-text-light-1)}
@media(max-width:767.98px){.trend-visible-filters{grid-template-columns:minmax(0,1fr)}.trend-chart{height:320px}}
</style>
