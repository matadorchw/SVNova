// Copyright 2026 The SVNova Authors. SPDX-License-Identifier: MIT
import test from 'node:test';
import assert from 'node:assert/strict';
import {suggestRepositoryName} from './repo-migration.ts';

test('suggests editable SVN names from URL and Windows or POSIX paths', () => {
 for(const [source,expected] of [
  ['https://example.com/svn/production/','production'],
  ['https://example.com/svn/my%20repo?ignored=1#fragment','my-repo'],
  ['https://example.com/svn/my%2Frepo','my-repo'],
  ['D:\\svn\\production\\','production'],
  ['/srv/中文路径/production','production'],
  ['https://example.com/',''],['D:\\',''],['/',''],['https://',''],
  ['https://example.com/svn/中文',''],['https://example.com/svn/production.git','production.git'],
  ['/srv/..project..','project'],['/srv/'+ 'a'.repeat(90),'a'.repeat(80)],
 ]) assert.equal(suggestRepositoryName(source),expected,source);
});
