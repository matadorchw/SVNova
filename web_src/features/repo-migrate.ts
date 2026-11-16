// Adapted from Gitea features/repo-migrate.ts (MIT), pinned baseline.
type Task = {indeterminate:boolean;id:number;kind:string;state:string;phase:string;started:string;finished:string;message:string;retry:boolean};
type Status = {progress:number;status:number;message:string;repositoryState:string;canRetry:boolean;tasks:Task[]};

export function initRepoMigrationStatusChecker(initForms: () => void) {
 const panel = document.querySelector<HTMLElement>('[data-migration-panel]');
 const history = document.querySelector<HTMLElement>('[data-task-status]');
 const root = panel || history;
 const url = root?.dataset.statusUrl;
 if (!root || !url) return;
 let stopped = false;
 let timer: ReturnType<typeof setTimeout> | undefined;
 let signature = '';
 const controller = new AbortController();
 window.addEventListener('pagehide', () => {stopped=true;controller.abort();if(timer)clearTimeout(timer);}, {once:true});
 const show = (selector:string, visible:boolean) => {const el=root.querySelector<HTMLElement>(selector);if(el)el.hidden=!visible;};
 const text = (selector:string, value:string) => {const el=root.querySelector<HTMLElement>(selector);if(el)el.textContent=value;};
 const refresh = async () => {
  const res = await fetch(url, {headers:{Accept:'application/json'},signal:controller.signal});
  if (res.status===401 || (res.redirected && new URL(res.url).pathname==='/user/login')) {stopped=true;location.assign('/user/login');return false;}
  if (res.status===403 || res.status===404) {
   text('#repo_migrating_failed_error','仓库不存在或访问权限已变化，请刷新页面');
   show('#repo_migrating',false);show('#repo_migrating_progress',false);show('#repo_migrating_failed_image',true);show('#repo_migrating_failed',true);show('#repo_migrating_retry',false);
   if(history){history.querySelector('tbody')?.replaceChildren();const message=history.querySelector<HTMLElement>('[data-task-connection]');if(message){message.hidden=false;message.textContent='仓库不存在或访问权限已变化，请刷新页面';}}
   return false;
  }
  if (!res.ok) throw Error('Migration status unavailable');
  const data = await res.json() as Status;
  if(stopped)return false;
  show('#repo_migrating_connection',false);show('[data-task-connection]',false);
  if(panel) {
   text('#repo_migrating_progress_message',data.message);
   const progress=root.querySelector<HTMLProgressElement>('#repo_migrating_revision_progress');if(progress){progress.hidden=data.status!==2;if(data.progress>=0)progress.value=data.progress;else progress.removeAttribute('value');}
   // Gitea reloads the whole page when migration finishes, rebuilding navigation.
   if(data.status===4){stopped=true;location.reload();return false;}
   if(data.status===3){show('#repo_migrating',false);show('#repo_migrating_progress',false);show('#repo_migrating_failed_image',true);show('#repo_migrating_failed',true);show('#repo_migrating_retry',data.canRetry);text('#repo_migrating_failed_error',data.message);return false;}
   show('#repo_migrating',true);show('#repo_migrating_progress',true);show('#repo_migrating_failed_image',false);show('#repo_migrating_failed',false);show('#repo_migrating_retry',false);
  } else if(history) {
   if(data.repositoryState!==history.dataset.repositoryState){stopped=true;location.reload();return false;}
   const next=JSON.stringify(data.tasks);
   if(next!==signature && !history.contains(document.activeElement)) {
    signature=next;
    const body=history.querySelector('tbody');
    if(body){
     const rows=document.createDocumentFragment();
     for(const task of data.tasks){
      const row=document.createElement('tr');
      [`#${task.id}`,task.kind,task.state,task.phase,task.started,task.finished,task.message].forEach((value,index)=>{const cell=document.createElement('td');cell.textContent=value;if(index===3&&task.indeterminate){const progress=document.createElement('progress');progress.setAttribute('aria-label','任务进行中');cell.append(progress);}row.append(cell);});
      const actions=document.createElement('td');
      if(task.retry && history.dataset.retryUrl){const form=document.createElement('form');form.className='ui form';form.dataset.api=history.dataset.retryUrl;const button=document.createElement('button');button.className='action-icon-button';button.title='重试任务';button.setAttribute('aria-label','重试任务');const svg=document.createElementNS('http://www.w3.org/2000/svg','svg');svg.setAttribute('width','16');svg.setAttribute('height','16');svg.setAttribute('aria-hidden','true');const use=document.createElementNS('http://www.w3.org/2000/svg','use');use.setAttribute('href','/assets/img/octicons.svg#issue-reopened');svg.append(use);button.append(svg);form.append(button);actions.append(form);}
      row.append(actions);
      rows.append(row);
     }
     if(!data.tasks.length){const row=document.createElement('tr'),cell=document.createElement('td');cell.colSpan=8;cell.textContent='暂无任务';row.append(cell);rows.append(row);}
     body.replaceChildren(rows);initForms();
    }
   }
  }
  return data.status!==3 && data.status!==4;
 };
 const syncTaskStatus = async () => {
  let doNextRefresh=true;
  try {doNextRefresh=await refresh();}
  catch {if(!stopped){show('#repo_migrating_connection',true);show('[data-task-connection]',true);}}
  finally {if(!stopped && doNextRefresh)timer=setTimeout(syncTaskStatus,2000);}
 };
 void syncTaskStatus();
}
