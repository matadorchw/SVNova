import test from 'node:test';
import assert from 'node:assert/strict';
import {repositoryTaskOption,taskTooltip,taskDuration,taskTime} from '../web_src/features/repository-task-options.ts';

const points=[
 {id:1,kind:'mirror',state:'succeeded',started:1700000000000,finished:1700000005000,duration:5000,before:0,after:3792,added:0},
 {id:2,kind:'local',state:'succeeded',started:1700000100000,finished:1700000105000,duration:5000,before:0,after:3792,added:0},
 {id:3,kind:'conversion',state:'succeeded',started:1700000200000,finished:1700000205000,duration:5000,before:3792,after:3792,added:0},
 {id:4,kind:'sync',state:'succeeded',started:1700000300000,finished:1700000305000,duration:5000,before:3792,after:3824,added:32},
 {id:5,kind:'sync',state:'succeeded',started:1700000400000,finished:1700000405000,duration:5000,before:3824,after:3824,added:0},
 {id:6,kind:'sync',state:'failed',started:1700000400000,finished:1700000401000,duration:1000,before:null,after:null,added:0},
 {id:7,kind:'sync',state:'running',started:1700000500000,finished:0,duration:null,before:null,after:null,added:0},
];
const option=(data=points,selected={},width=1000,zoom={start:0,end:100})=>repositoryTaskOption(data,zoom,selected,name=>name,width);
const syncOnly={镜像迁移:false,本地迁移:false,转换镜像:false};

test('one plotting grid displays revision and duration with special and failure event lines',()=>{
 const o=option();assert.equal(o.grid.length,1);assert.equal(o.xAxis.length,1);assert.equal(o.yAxis.length,2);
 assert.equal(o.series.length,7);assert.ok(o.series.every(s=>s.type!=='scatter'));
 assert.deepEqual(o.legend.data,['版本号','耗时']);assert.equal(o.series[0].name,'版本号');assert.equal(o.series[0].yAxisIndex,0);assert.equal(o.series[0].step,'end');assert.equal(o.series[1].name,'耗时');assert.equal(o.series[1].yAxisIndex,1);assert.ok(o.series[0].data.every(item=>item.task.after!==null));assert.equal(o.series[0].data.length,5);
 assert.deepEqual(o.series.slice(2,6).map(s=>s.id),['mirror','local','conversion','failed']);
 assert.ok(o.series.slice(2,6).every(s=>s.data.length===1));
 assert.equal(o.graphic[0].x,o.grid[0].left-8);assert.equal(o.graphic[0].style.align,'right');
 assert.equal(o.graphic[1].x,1000-o.grid[0].right+8);assert.equal(o.graphic[1].style.align,'left');
 assert.deepEqual(o.dataZoom[0].xAxisIndex,[0]);assert.equal(o.dataZoom.length,1);
 assert.equal(o.tooltip.trigger,'axis');assert.equal(o.tooltip.axisPointer.type,'line');assert.equal(o.tooltip.axisPointer.axis,'x');assert.equal(o.tooltip.axisPointer.snap,true);
 assert.match(o.tooltip.formatter({data:{task:points[4]}}),/#5[\s\S]*#6/);
 assert.match(o.tooltip.formatter([{data:{task:points[4]}},{data:{task:points[5]}}]),/#5[\s\S]*#6/);
});
test('type selection recalculates time range and filters all plot data and tooltips',()=>{
 const all=option(),filtered=option(points,syncOnly);
 assert.ok(filtered.xAxis[0].min>all.xAxis[0].min);
 assert.ok(filtered.xAxis[0].max-filtered.xAxis[0].min<all.xAxis[0].max-all.xAxis[0].min);
 assert.ok(filtered.series.slice(2,5).every(s=>s.data.length===0));
 assert.equal(filtered.series[0].data[0].value[1],3792);assert.equal(filtered.series[0].data[1].value[1],3824);
 assert.equal(filtered.tooltip.formatter({data:{task:points[0]}}),'');
 const restored=option(points,{});assert.equal(restored.xAxis[0].min,all.xAxis[0].min);
 const empty=option(points,{...syncOnly,镜像同步:false});assert.ok(empty.series.every(s=>!s.data.length));assert.equal(empty.xAxis[0].show,false);
});
test('duration line omits unfinished values while retaining failed task data',()=>{
 const o=option(points);assert.equal(o.graphic[1].style.text,'耗时');assert.equal(o.series[1].step,undefined);
 assert.deepEqual(o.series[1].data.map(item=>item.value[1]),[5000,5000,5000,5000,5000,1000]);
 assert.match(taskTooltip(points[5]),/失败/);assert.match(taskTooltip(points[5]),/修订数据未知/);
 assert.match(taskTooltip(points[4]),/新增修订：0/);assert.match(taskTooltip(points[6]),/耗时：—/);
});
test('empty and single-task ranges are finite and time labels use Beijing time',()=>{
 for(const data of [[],[points[0]]]){const o=option(data);assert.ok(o.xAxis[0].max>o.xAxis[0].min);}
 assert.equal(taskTime(Date.UTC(2026,0,1)),'2026-01-01 08:00:00');
 assert.equal(taskDuration(250),'250 毫秒');assert.equal(taskDuration(1500),'1.5 秒');assert.equal(taskDuration(90000),'1.5 分钟');assert.equal(taskDuration(5400000),'1.5 小时');
});
async function chartFixture(width=1000){
 const {init,use}=await import('echarts/core');const {CustomChart,LineChart}=await import('echarts/charts');
 const {GraphicComponent,GridComponent,TooltipComponent,LegendComponent,DataZoomSliderComponent}=await import('echarts/components');const {SVGRenderer}=await import('echarts/renderers');
 use([CustomChart,LineChart,GraphicComponent,GridComponent,TooltipComponent,LegendComponent,DataZoomSliderComponent,SVGRenderer]);
 return init(null,undefined,{renderer:'svg',ssr:true,width,height:400});
}
test('running and legacy sync tasks snap at their own time without drawing fake metrics',async()=>{
 const c=await chartFixture();
 const running=points[6],legacy={...running,id:8,state:'succeeded',started:running.started+100000};
 const data=[points[4],running,legacy];
 try{
  for(const selected of [{},{版本号:false,耗时:false}]){
   const o=option(data,selected);c.setOption(o,{notMerge:true});
   assert.ok(o.series[0].data.every(item=>item.task.id!==7&&item.task.id!==8));
   assert.ok(o.series[1].data.every(item=>item.task.id!==7&&item.task.id!==8));
   const carrier=c.getModel().getSeries().find(series=>series.id==='task-times').getData();
   assert.equal(carrier.count(),3);assert.equal(carrier.getItemGraphicEl(1),undefined);
   let pointer;c.on('updateAxisPointer',event=>{pointer=event;});
   const plot=c.getModel().getComponent('grid',0).coordinateSystem.getRect();
   for(const task of [running,legacy]){
    c.dispatchAction({type:'updateAxisPointer',x:c.convertToPixel({xAxisIndex:0},task.started),y:plot.y+plot.height/2});
    assert.equal(pointer.axesInfo.find(axis=>axis.axisDim==='x').value,task.started);
    const item=o.series.find(series=>series.id==='task-times').data.find(item=>item.task.id===task.id);
    const text=o.tooltip.formatter([{data:item}]);assert.match(text,new RegExp(`#${task.id} · 镜像同步`));assert.match(text,/耗时：—/);assert.match(text,/版本：未知/);
   }
   c.off('updateAxisPointer');
  }
  const hidden=option(data,{镜像同步:false});assert.ok(hidden.series.find(series=>series.id==='task-times').data.length===0);
 }finally{c.dispose();}
});
test('special events span full plot height with different line patterns and aligned timestamps',async()=>{
 const c=await chartFixture();try{
  c.setOption(option());const plot=c.getModel().getComponent('grid',0).coordinateSystem.getRect();
  const patterns=[];for(const kind of ['mirror','local','conversion']){
   const s=c.getModel().getSeries().find(s=>s.id===kind);const d=s.getData(),el=d.getItemGraphicEl(0),task=d.getRawDataItem(0).task;
   assert.equal(el.type,'line');assert.equal(el.shape.y1,plot.y);assert.equal(el.shape.y2,plot.y+plot.height);
   assert.equal(el.shape.x1,el.shape.x2);assert.ok(Math.abs(el.shape.x1-c.convertToPixel({xAxisIndex:0},task.started))<0.01);patterns.push(el.style.lineDash);
  }
  assert.deepEqual(patterns,[[],[6,4],[2,4]]);
  let pointer;
  c.on('updateAxisPointer',event=>{pointer=event;});
  const hoverX=c.convertToPixel({xAxisIndex:0},points[3].started+1000);
  c.dispatchAction({type:'updateAxisPointer',x:hoverX,y:plot.y+plot.height/2});
  assert.equal(pointer.axesInfo.find(axis=>axis.axisDim==='x').value,points[3].started);
  const failed=c.getModel().getSeries().find(s=>s.id==='failed');const red=failed.getData().getItemGraphicEl(0);assert.equal(red.style.stroke,'--color-red');assert.equal(red.shape.y1,plot.y);assert.equal(red.shape.y2,plot.y+plot.height);assert.equal(failed.getData().getRawDataItem(0).task.id,6);assert.ok(c.getModel().getComponent('yAxis',0).axis.scale.getExtent()[0]>3000);
  c.dispatchAction({type:'dataZoom',start:60,end:100});assert.equal(c.getModel().getSeries().find(s=>s.id==='mirror').getData().getItemGraphicEl(0),undefined);
  c.dispatchAction({type:'dataZoom',start:0,end:100});assert.ok(c.getModel().getSeries().find(s=>s.id==='mirror').getData().getItemGraphicEl(0));
  assert.match(c.renderToSVGString(),/<svg/);
  c.setOption(option(points.map(task=>task.id===1?{...task,state:'failed',before:null,after:null}:task)),{notMerge:true});
  assert.equal(c.getModel().getSeries().find(s=>s.id==='mirror').getData().count(),0);
  assert.equal(c.getModel().getSeries().find(s=>s.id==='failed').getData().count(),2);
  c.setOption(option([{...points[5],duration:null}]),{notMerge:true});
  assert.ok(c.getModel().getSeries().find(s=>s.id==='failed').getData().getItemGraphicEl(0));
 }finally{c.dispose();}
});
test('dual lines, legend selection and narrow resize retain valid plotting bounds',async()=>{
 const c=await chartFixture();try{
  c.setOption(option());c.dispatchAction({type:'legendToggleSelect',name:'版本号'});assert.equal(c.getOption().legend[0].selected.版本号,false);
  for(const width of [240,256,258,300,360,600,1000]){
   c.resize({width,height:380});const o=option(points,{...syncOnly,版本号:false},width,{start:10,end:80});c.setOption(o,{notMerge:true});
   assert.ok(c.getModel().getComponent('grid',0).coordinateSystem.getRect().width>=120);
   assert.equal(c.getOption().dataZoom[0].start,10);assert.equal(c.getOption().legend[0].selected.版本号,false);
   assert.equal(o.dataZoom[0].left,o.grid[0].left);assert.equal(o.dataZoom[0].right,o.grid[0].right);
   if(width<480)assert.match(o.dataZoom[0].labelFormatter(points[0].started),/^\d{2}-\d{2}\n\d{2}:\d{2}$/);
   if(width>=768)assert.equal(o.grid[0].left,128);
   assert.match(c.renderToSVGString(),/耗时/);
  }
 }finally{c.dispose();}
});
