<script setup lang="ts">
import mail from '@primer/octicons/build/svg/mail-16.svg?raw';
import repo from '@primer/octicons/build/svg/repo-16.svg?raw';
import organization from '@primer/octicons/build/svg/organization-16.svg?raw';
import directory from '@primer/octicons/build/svg/file-directory-fill-16.svg?raw';
import file from '@primer/octicons/build/svg/file-16.svg?raw';
import plus from '@primer/octicons/build/svg/plus-16.svg?raw';
import issueReopened from '@primer/octicons/build/svg/issue-reopened-16.svg?raw';
import sync from '@primer/octicons/build/svg/sync-16.svg?raw';
import history from '@primer/octicons/build/svg/history-16.svg?raw';
import gear from '@primer/octicons/build/svg/gear-16.svg?raw';
import tasklist from '@primer/octicons/build/svg/tasklist-16.svg?raw';
import signOut from '@primer/octicons/build/svg/sign-out-16.svg?raw';
import download from '@primer/octicons/build/svg/download-16.svg?raw';
import search from '@primer/octicons/build/svg/search-16.svg?raw';
import arrowUp from '@primer/octicons/build/svg/arrow-up-16.svg?raw';
import shield from '@primer/octicons/build/svg/shield-16.svg?raw';
import person from '@primer/octicons/build/svg/person-16.svg?raw';
import people from '@primer/octicons/build/svg/people-16.svg?raw';
import key from '@primer/octicons/build/svg/key-16.svg?raw';
import lock from '@primer/octicons/build/svg/lock-16.svg?raw';
import check from '@primer/octicons/build/svg/check-16.svg?raw';
import alert from '@primer/octicons/build/svg/alert-16.svg?raw';
import server from '@primer/octicons/build/svg/server-16.svg?raw';
import tools from '@primer/octicons/build/svg/tools-16.svg?raw';
import clock from '@primer/octicons/build/svg/clock-16.svg?raw';
import mirror from '@primer/octicons/build/svg/mirror-16.svg?raw';

import triangleDown from '@primer/octicons/build/svg/triangle-down-16.svg?raw';
const icons={issueReopened,mail,triangleDown,repo,organization,directory,file,plus,sync,history,gear,tasklist,signOut,download,search,arrowUp,shield,person,people,key,lock,check,alert,server,tools,clock,mirror};
defineProps<{name:keyof typeof icons;size?:number}>();
</script>

<template>
  <!-- Only bundled upstream SVGs enter this HTML sink; never accept user SVG markup. -->
  <span class="octicon" aria-hidden="true" :style="{width:(size||16)+'px',height:(size||16)+'px'}" v-html="icons[name]"></span>
</template>
