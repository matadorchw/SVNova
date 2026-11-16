import assert from 'node:assert/strict';
import test from 'node:test';
import {initTaskCenter} from '../web_src/features/task-center.ts';

function fixture(t,response) {
 const original=new Map(['document','location','history','fetch'].map(key=>[key,Object.getOwnPropertyDescriptor(globalThis,key)]));
 t.after(()=>{for(const [key,descriptor] of original){if(descriptor)Object.defineProperty(globalThis,key,descriptor);else delete globalThis[key];}});
 const events={},radioEvents={};
 let submits=0;
 const root={innerHTML:'old rows',querySelector:()=>({dataset:{taskCenterActive:'false'}}),replaceChildren(){this.innerHTML='';}};
 const total={textContent:'5'},error={hidden:true};
 const form={querySelectorAll:()=>[{addEventListener:(name,callback)=>{radioEvents[name]=callback;}}],requestSubmit(){submits++;}};
 const selectors={'[data-task-center]':root,'[data-task-center-total]':total,'[data-task-center-error]':error,'.task-center-filters':form};
 globalThis.document={hidden:false,querySelector:selector=>selectors[selector],addEventListener:(name,callback)=>{events[name]=callback;}};
 globalThis.location={search:'?q=Tools&kind=sync',href:'http://localhost/tasks?q=Tools&kind=sync&page=2',assign(){}};
 let requested='',replaced='';
 globalThis.history={replaceState:(_state,_title,url)=>{replaced=String(url);}};
 globalThis.fetch=async url=>{requested=url;return response();};
 initTaskCenter();
 return {root,total,error,form,events,radioEvents,get submits(){return submits;},get requested(){return requested;},get replaced(){return replaced;}};
}
test('polling updates filtered total and rows while leaving filter controls intact',async t=>{
 const f=fixture(t,async()=>({ok:true,status:200,json:async()=>({html:'new rows',page:1,total:12})}));
 f.events.visibilitychange();await new Promise(setImmediate);
 assert.equal(f.requested,'/api/v1/tasks?q=Tools&kind=sync');
 assert.equal(f.root.innerHTML,'new rows');assert.equal(f.total.textContent,'12');assert.equal(f.error.hidden,true);
 assert.equal(new URL(f.replaced).searchParams.get('page'),'1');
 f.radioEvents.change();assert.equal(f.submits,1);
 assert.equal(globalThis.document.querySelector('.task-center-filters'),f.form);
});
test('refresh failure clears stale rows and count',async t=>{
 const f=fixture(t,async()=>{throw Error('isolated failure');});
 f.events.visibilitychange();await new Promise(setImmediate);
 assert.equal(f.root.innerHTML,'');assert.equal(f.total.textContent,'—');assert.equal(f.error.hidden,false);
});
