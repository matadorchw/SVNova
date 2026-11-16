import {readFileSync,writeFileSync} from 'node:fs';
import {Resvg} from '@resvg/resvg-js';

const root=new URL('../static/',import.meta.url);
const svg=readFileSync(new URL('logo.svg',root),'utf8');
writeFileSync(new URL('favicon.svg',root),svg);
const images=[16,32,48].map(size=>{
  const png=new Resvg(svg,{fitTo:{mode:'width',value:size}}).render().asPng();
  writeFileSync(new URL(`favicon-${size}.png`,root),png);
  return {size,png};
});
const header=Buffer.alloc(6+16*images.length);
header.writeUInt16LE(1,2);
header.writeUInt16LE(images.length,4);
let offset=header.length;
images.forEach(({size,png},i)=>{
  const entry=6+16*i;
  header[entry]=size;header[entry+1]=size;
  header.writeUInt16LE(1,entry+4);header.writeUInt16LE(32,entry+6);
  header.writeUInt32LE(png.length,entry+8);header.writeUInt32LE(offset,entry+12);
  offset+=png.length;
});
writeFileSync(new URL('favicon.ico',root),Buffer.concat([header,...images.map(i=>i.png)]));
writeFileSync(new URL('logo-preview.png',root),new Resvg(svg,{fitTo:{mode:'width',value:512}}).render().asPng());
// Only pinned Octicons SVG files enter the generated sprite.
const spriteNames=['info','hourglass','issue-reopened','arrow-switch','mail','database','eye','paper-airplane','bell','chevron-left','chevron-right','double-chevron-left','double-chevron-right','pencil','x','mirror','clock','tools','server','lock','check','alert','person','repo-push','triangle-down','triangle-up','shield-check','three-bars','repo','organization','file-directory-fill','file','plus','sync','history','gear','tasklist','sign-out','download','search','arrow-up','shield','people','key'];
const symbols=spriteNames.map(name=>{
 const source=readFileSync(new URL(`../node_modules/@primer/octicons/build/svg/${name.startsWith('double-') ? name.slice(7) : name}-16.svg`,import.meta.url),'utf8');
 if (name.startsWith('double-')) {
  const paths=source.replace(/<svg[^>]*>/, '').replace('</svg>', '');
  return `<symbol id="${name}" viewBox="0 0 16 16"><g transform="translate(-3 0)">${paths}</g><g transform="translate(3 0)">${paths}</g></symbol>`;
 }
 return source.replace(/<svg[^>]*>/,`<symbol id="${name}" viewBox="0 0 16 16">`).replace('</svg>','</symbol>');
});
writeFileSync(new URL('img/octicons.svg',root),`<!-- @primer/octicons MIT; see octicons-LICENSE.txt --><svg xmlns="http://www.w3.org/2000/svg">${symbols.join('')}</svg>`);
writeFileSync(new URL('img/octicons-LICENSE.txt',root),readFileSync(new URL('../docs/upstream/octicons-LICENSE.txt',import.meta.url)));
