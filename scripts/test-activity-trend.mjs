import assert from 'node:assert/strict';
import test from 'node:test';
import {trendOption,trendTooltip,trendStates} from '../web_src/features/activity-trend-options.ts';
const days=[{date:'2026-11-17',counts:{queued:1,running:2,succeeded:3,failed:4,interrupted:5}},{date:'2026-11-18',counts:{}}];
test('stacked series retain every status and zero-value date',()=>{
 const option=trendOption(days,'同步',{}, {start:0,end:100},name=>name);
 assert.equal(option.series.length,5);
 assert.deepEqual(option.series.map(s=>s.stack),Array(5).fill('activity'));
 assert.deepEqual(option.series.map(s=>s.data[1]),Array(5).fill(0));
 assert.deepEqual(option.xAxis.data,['2026-11-17','2026-11-18']);
 assert.equal(option.yAxis.minInterval,1);
});
test('tooltip follows legend visibility including zero counts and visible total',()=>{
 const selected={'失败':false};
 assert.equal(trendTooltip(days[0],'镜像仓库同步',selected),'2026-11-17\n镜像仓库同步\n排队中  1 次\n进行中  2 次\n成功  3 次\n已中断  5 次\n合计  11 次');
 assert.match(trendTooltip(days[1],'同步',{}),/合计  0 次$/);
 const option=trendOption(days,'同步',selected,{start:0,end:100},name=>name);
 assert.equal(option.tooltip.formatter([{dataIndex:1}]),trendTooltip(days[1],'同步',selected));
 assert.equal(option.tooltip.renderMode,'richText');
 assert.equal(trendTooltip(days[0],'同步',Object.fromEntries(trendStates.map(s=>[s.label,false]))),'2026-11-17\n同步\n合计  0 次');
});
test('slider is the only zoom component and saved selection is preserved',()=>{
 const selected={'成功':false};
 const option=trendOption(days,'迁移',selected,{start:25,end:75},name=>name);
 assert.equal(option.dataZoom.length,1);
 assert.equal(option.dataZoom[0].type,'slider');
 assert.equal(option.dataZoom[0].start,25);
 assert.equal(option.dataZoom[0].end,75);
 assert.deepEqual(option.legend.selected,selected);
});

test('ECharts renders SVG and updates slider and legend state',async()=>{
 const {init,use}=await import('echarts/core');
 const {BarChart}=await import('echarts/charts');
 const {GridComponent,LegendComponent,TooltipComponent,DataZoomSliderComponent}=await import('echarts/components');
 const {SVGRenderer}=await import('echarts/renderers');
 use([BarChart,GridComponent,LegendComponent,TooltipComponent,DataZoomSliderComponent,SVGRenderer]);
 const chart=init(null,undefined,{renderer:'svg',ssr:true,width:600,height:360});
 try{
  chart.setOption(trendOption(days,'同步',{}, {start:0,end:100},()=> '#666'));
  let selection;
  chart.on('legendselectchanged',event=>{selection=event.selected;});
  chart.dispatchAction({type:'legendToggleSelect',name:'失败'});
  assert.equal(selection['失败'],false);
  chart.dispatchAction({type:'dataZoom',dataZoomId:'dates',start:25,end:75});
  chart.setOption(trendOption(days,'同步',selection,{start:25,end:75},()=> '#666'));
  assert.equal(chart.getOption().dataZoom[0].start,25);
  assert.equal(chart.getOption().legend[0].selected['失败'],false);
  chart.dispatchAction({type:'dataZoom',dataZoomId:'dates',start:0,end:100});
  chart.resize({width:320,height:320});
  const svg=chart.renderToSVGString();
  assert.match(svg,/<svg/);
  assert.match(svg,/排队中/);
  assert.match(svg,/成功/);
 }finally{chart.dispose();}
 assert.equal(chart.isDisposed(),true);
});
