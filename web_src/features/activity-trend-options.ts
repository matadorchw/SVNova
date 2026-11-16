import type {ComposeOption} from 'echarts/core';
import type {BarSeriesOption} from 'echarts/charts';
import type {GridComponentOption,LegendComponentOption,TooltipComponentOption,DataZoomComponentOption} from 'echarts/components';

export type TrendDay={date:string;counts:Record<string,number>};
export const trendStates=[
 {key:'queued',label:'排队中',color:'--color-grey'},
 {key:'running',label:'进行中',color:'--color-blue'},
 {key:'succeeded',label:'成功',color:'--color-green'},
 {key:'failed',label:'失败',color:'--color-red'},
 {key:'interrupted',label:'已中断',color:'--color-orange'},
];
export type TrendOption=ComposeOption<BarSeriesOption|GridComponentOption|LegendComponentOption|TooltipComponentOption|DataZoomComponentOption>;
export function trendTooltip(day:TrendDay,label:string,selected:Record<string,boolean>):string {
 const visible=trendStates.filter(state=>selected[state.label]!==false);
 return [day.date,label,...visible.map(state=>`${state.label}  ${(day.counts[state.key]??0).toLocaleString('zh-CN')} 次`),`合计  ${visible.reduce((sum,state)=>sum+(day.counts[state.key]??0),0).toLocaleString('zh-CN')} 次`].join('\n');
}
export function trendOption(days:TrendDay[],label:string,selected:Record<string,boolean>,zoom:{start:number;end:number},color:(name:string)=>string):TrendOption {
 return {
  animation:false,
  textStyle:{color:color('--color-text'),fontFamily:'Microsoft YaHei, sans-serif'},
  color:trendStates.map(state=>color(state.color)),
  legend:{top:0,left:'center',itemWidth:12,itemHeight:12,itemGap:12,selected,textStyle:{color:color('--color-text'),fontSize:12}},
  grid:{top:64,left:80,right:80,bottom:76,outerBoundsMode:'same',outerBoundsContain:'axisLabel'},
  xAxis:{type:'category',data:days.map(day=>day.date),axisLabel:{formatter:(date:string)=>date.slice(5),hideOverlap:true},axisTick:{alignWithLabel:true},axisLine:{lineStyle:{color:color('--color-secondary')}}},
  yAxis:{type:'value',min:0,minInterval:1,splitLine:{lineStyle:{color:color('--color-secondary')}}},
  tooltip:{trigger:'axis',renderMode:'richText',confine:true,axisPointer:{type:'shadow'},backgroundColor:color('--color-box-body'),borderColor:color('--color-secondary'),textStyle:{color:color('--color-text'),fontSize:12},formatter:params=>{
   const item=Array.isArray(params)?params[0]:params;
   const day=item?days[item.dataIndex]:undefined;
   return day?trendTooltip(day,label,selected):'';
  }},
  dataZoom:[{id:'dates',type:'slider',left:80,right:80,xAxisIndex:0,start:zoom.start,end:zoom.end,bottom:8,height:24,brushSelect:false,showDetail:true,textStyle:{color:color('--color-text')},borderColor:color('--color-secondary'),fillerColor:color('--color-primary-alpha-10'),handleStyle:{color:color('--color-primary')},moveHandleSize:0}],
  series:trendStates.map(state=>({id:state.key,name:state.label,type:'bar',stack:'activity',barMaxWidth:36,data:days.map(day=>day.counts[state.key]??0)})),
 };
}
