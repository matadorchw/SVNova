import type {ComposeOption} from 'echarts/core';
import type {CustomSeriesOption,LineSeriesOption} from 'echarts/charts';
import type {GraphicComponentOption,GridComponentOption,TooltipComponentOption,LegendComponentOption,DataZoomComponentOption} from 'echarts/components';

export type TaskPoint={id:number;kind:string;state:string;started:number;finished:number;duration:number|null;before:number|null;after:number|null;added:number};
export type TaskRow={id:number;kind:string;state:string;phase:string;started:string;finished:string;message:string;retry:boolean;indeterminate:boolean};
export type TaskHistory={summary:{total:number;successRate:number|null;revisionSpan:string;added:number};points:TaskPoint[];items:TaskRow[];page:number;pages:number;active:boolean;repositoryState:string};
export const taskKinds=[{key:'mirror',label:'镜像迁移',color:'--color-purple',lineStyle:'solid' as const,lineDash:[] as number[]},{key:'local',label:'本地迁移',color:'--color-blue',lineStyle:'dashed' as const,lineDash:[6,4]},{key:'conversion',label:'转换镜像',color:'--color-orange',lineStyle:'dotted' as const,lineDash:[2,4]},{key:'sync',label:'镜像同步',color:'--color-green',lineStyle:'solid' as const,lineDash:[] as number[]}];
const states:Record<string,string>={queued:'排队中',running:'进行中',succeeded:'成功',failed:'失败',interrupted:'已中断'};
export function taskTime(value:number):string{return value>0?new Intl.DateTimeFormat('sv-SE',{timeZone:'Asia/Shanghai',year:'numeric',month:'2-digit',day:'2-digit',hour:'2-digit',minute:'2-digit',second:'2-digit',hour12:false}).format(new Date(value)):'未知';}
export function taskDuration(value:number):string{if(value<1000)return `${value} 毫秒`;if(value<60000)return `${(value/1000).toFixed(1)} 秒`;if(value<3600000)return `${(value/60000).toFixed(1)} 分钟`;return `${(value/3600000).toFixed(1)} 小时`;}
export function taskTooltip(point:TaskPoint):string {
 return [`#${point.id} · ${taskKinds.find(kind=>kind.key===point.kind)?.label??'任务'} · ${states[point.state]??point.state}`,`开始：${taskTime(point.started)}`,`结束：${point.finished?taskTime(point.finished):'—'}`,`耗时：${point.duration===null?'—':taskDuration(point.duration)}`,`新增修订：${point.added}${point.after===null?'（修订数据未知）':''}`,`版本：${point.before===null||point.after===null?'未知':`r${point.before} → r${point.after}`}`].join('\n');
}
type Options=ComposeOption<CustomSeriesOption|LineSeriesOption|GraphicComponentOption|GridComponentOption|TooltipComponentOption|LegendComponentOption|DataZoomComponentOption>;
export function repositoryTaskOption(points:TaskPoint[],zoom:{start:number;end:number},selected:Record<string,boolean>,color:(name:string)=>string,width=1000):Options {
 const inset=width<480?56:width<768?88:128;
 const timeLabel=(value:number)=>width<480?taskTime(value).slice(5,16).replace(' ','\n'):taskTime(value).slice(5,16);
 const timed=points.filter(point=>point.started>0);
 const visible=timed.filter(task=>selected[taskKinds.find(kind=>kind.key===task.kind)?.label??'']!==false);
 const min=visible.length?visible.reduce((min,point)=>Math.min(min,point.started),Infinity):0;
 const max=visible.length?visible.reduce((max,point)=>Math.max(max,point.started),-Infinity):1;
 const extent=Math.max(1000,(max-min)*0.02);
 const revisions=visible.filter(task=>task.after!==null).map(task=>({value:[task.started,task.after],task}));
 const first=revisions[0]?.task;
 if(first?.kind==='sync'&&first.before!==null&&first.before!==first.after)revisions.unshift({value:[first.started,first.before],task:first});
 const durations=visible.filter(task=>task.duration!==null).map(task=>({value:[task.started,task.duration],task}));
 const eventSeries=(id:string,name:string,paint:string,lineDash:number[],tasks:TaskPoint[]):CustomSeriesOption=>({id,name,type:'custom',z:1,yAxisIndex:1,encode:{x:0,y:-1},renderItem:(params,api)=>{
  const x=api.coord([api.value(0),0])[0],rect=params.coordSys;
  if(!('x' in rect)||!('width' in rect)||!('y' in rect)||!('height' in rect)||typeof rect.x!=='number'||typeof rect.width!=='number'||typeof rect.y!=='number'||typeof rect.height!=='number')return;
  if(x<rect.x||x>rect.x+rect.width)return;
  return {type:'line',shape:{x1:x,y1:rect.y,x2:x,y2:rect.y+rect.height},style:{stroke:paint,lineWidth:id==='failed'?2:1.5,lineDash}};
 },data:tasks.map(task=>({value:[task.started],task}))});
 return {
  animation:false,textStyle:{color:color('--color-text'),fontFamily:'Microsoft YaHei, sans-serif'},
  grid:[{left:inset,right:inset,top:72,bottom:76}],
  graphic:[{type:'text',x:inset-8,y:48,style:{align:'right',verticalAlign:'top',text:'版本号',fill:color('--color-text'),fontSize:12}},{type:'text',x:width-inset+8,y:48,style:{align:'left',verticalAlign:'top',text:'耗时',fill:color('--color-text'),fontSize:12}}],
  legend:{top:8,selected,textStyle:{color:color('--color-text')},data:['版本号','耗时']},
  xAxis:[{show:visible.length>0,type:'time' as const,containShape:false,min:min-extent,max:max+extent,axisLabel:{hideOverlap:true,formatter:timeLabel},axisLine:{lineStyle:{color:color('--color-secondary')}}}],
  yAxis:[{show:visible.length>0,type:'value',scale:true,minInterval:1,axisLabel:{formatter:(value:number)=>`r${value}`},splitLine:{lineStyle:{color:color('--color-secondary')}}},{show:visible.length>0,type:'value',position:'right',min:0,axisLabel:{formatter:taskDuration},splitLine:{show:false}}],
  tooltip:{trigger:'axis',axisPointer:{axis:'x',type:'line',snap:true,lineStyle:{color:color('--color-text-light-2'),type:'dashed'}},renderMode:'richText',confine:true,backgroundColor:color('--color-box-body'),borderColor:color('--color-secondary'),textStyle:{color:color('--color-text'),fontSize:12},formatter:params=>{const item=Array.isArray(params)?params[0]:params;const data=item?.data as {task?:TaskPoint}|undefined;return data?.task?visible.filter(task=>task.started===data.task?.started).map(taskTooltip).join('\n\n'):'';}},
  dataZoom:[{type:'slider',xAxisIndex:[0],left:inset,right:inset,bottom:8,height:24,start:zoom.start,end:zoom.end,filterMode:'none',brushSelect:false,moveHandleSize:0,textStyle:{color:color('--color-text')},labelFormatter:(value:number)=>width<768?timeLabel(value):taskTime(value).slice(0,16)}],
  series:[{id:'revision',name:'版本号',type:'line',yAxisIndex:0,step:'end',connectNulls:false,symbolSize:6,itemStyle:{color:color('--color-primary')},data:revisions},
   {id:'duration',name:'耗时',type:'line',yAxisIndex:1,connectNulls:false,symbolSize:6,itemStyle:{color:color('--color-orange')},data:durations},
   ...taskKinds.filter(kind=>kind.key!=='sync').map(kind=>eventSeries(kind.key,kind.label,color(kind.color),kind.lineDash,visible.filter(task=>task.kind===kind.key&&task.state!=='failed'))),
   eventSeries('failed','失败',color('--color-red'),[],visible.filter(task=>task.state==='failed')),
   // Keep tasks without metric snapshots available to axis-pointer snapping.
   {id:'task-times',name:'任务时间',type:'custom',yAxisIndex:1,encode:{x:0,y:-1},renderItem:()=>null,data:visible.map(task=>({value:[task.started],task}))}],
 };
}
