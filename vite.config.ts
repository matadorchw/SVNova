import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { readFileSync } from 'node:fs';
export default defineConfig({publicDir:'static',define:{'process.env.NODE_ENV':JSON.stringify('production')},plugins:[vue()],build:{outDir:'public/assets',emptyOutDir:true,rollupOptions:{output:{banner:'/*!\n'+readFileSync(new URL('./docs/upstream/octicons-LICENSE.txt',import.meta.url),'utf8')+'\n*/'}},lib:{entry:'web_src/main.ts',formats:['es'],fileName:()=> 'app.js',cssFileName:'app'}}});
