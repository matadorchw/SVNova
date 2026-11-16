import assert from 'node:assert/strict';
import test from 'node:test';
import {initSearchClear} from '../web_src/features/search-clear.ts';

function fixture(t,mode,value='') {
 const descriptors=new Map(['document','location','sessionStorage'].map(key=>[key,Object.getOwnPropertyDescriptor(globalThis,key)]));
 t.after(()=>{for(const [key,descriptor] of descriptors){if(descriptor)Object.defineProperty(globalThis,key,descriptor);else delete globalThis[key];}});
 const storage=new Map();globalThis.sessionStorage={getItem:key=>storage.get(key)??null,setItem:(key,value)=>storage.set(key,value),removeItem:key=>storage.delete(key)};globalThis.location={href:'http://localhost/tasks',pathname:'/tasks'};
 const pages=[{value:'4'},{value:'3'}];
 const filters={kind:'sync',state:'failed',sort:'newest',org:'1',user_id:'2'};
 const form=Object.assign(new EventTarget(),{submits:0,action:'http://localhost/tasks',controls:filters,reset(){for(const key of Object.keys(this.controls))this.controls[key]='';},querySelectorAll:()=>pages,requestSubmit(){this.submits++;}});
 const input=Object.assign(new EventTarget(),{name:mode==='submit'?'q':'Username',value,form,focused:false,focus(){this.focused=true;this.dispatchEvent(new Event('focus'));}});
 const button=Object.assign(new EventTarget(),{hidden:true});
 const attrs=new Set();
 const box={dataset:{searchClear:mode},hasAttribute:name=>attrs.has(name),setAttribute:name=>attrs.add(name),querySelector:selector=>selector==='input'?input:button};
 globalThis.document={querySelectorAll:()=>[box]};
 initSearchClear();
 return {input,button,form,pages,filters,attrs,storage};
}
test('list clear submits once, resets paging, preserves other conditions and restores focus',t=>{
 const f=fixture(t,'submit','Tools');
 assert.equal(f.button.hidden,false);
 initSearchClear();
 f.button.dispatchEvent(new Event('click',{cancelable:true}));
 assert.equal(f.input.value,'');assert.equal(f.button.hidden,true);assert.equal(f.input.focused,true);
 assert.equal(f.form.submits,1);assert.deepEqual(f.pages.map(p=>p.value),['1','1']);
 assert.deepEqual(f.filters,{kind:'sync',state:'failed',sort:'newest',org:'1',user_id:'2'});
});
test('candidate clear emits input but never submits the add-member form',t=>{
 const f=fixture(t,'input','member');let changes=0;
 f.input.addEventListener('input',()=>changes++);
 f.button.dispatchEvent(new Event('click',{cancelable:true}));
 assert.equal(f.input.value,'');assert.equal(changes,1);assert.equal(f.form.submits,0);
 assert.deepEqual(f.pages.map(p=>p.value),['4','3']);assert.equal(f.input.focused,true);
});
test('clear visibility follows typing, selected candidates and reset values',async t=>{
 const f=fixture(t,'input');assert.equal(f.button.hidden,true);
 f.input.value='candidate';f.input.dispatchEvent(new Event('input'));assert.equal(f.button.hidden,false);
 f.input.value='';f.input.dispatchEvent(new Event('change'));assert.equal(f.button.hidden,true);
 f.input.value='restored';f.input.dispatchEvent(new Event('focus'));assert.equal(f.button.hidden,false);
 f.form.dispatchEvent(new Event('reset'));f.input.value='';await Promise.resolve();assert.equal(f.button.hidden,true);
});

test('cleared list search restores focus after the next page load',t=>{
 const f=fixture(t,'submit','Tools');
 f.button.dispatchEvent(new Event('click',{cancelable:true}));
 assert.equal(f.storage.get('svnova.search-clear-focus'),'/tasks:q');
 f.attrs.delete('data-search-clear-ready');f.input.focused=false;
 initSearchClear();
 assert.equal(f.input.focused,true);assert.equal(f.storage.has('svnova.search-clear-focus'),false);
});
