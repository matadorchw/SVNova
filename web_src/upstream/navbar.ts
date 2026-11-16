// Copyright 2016 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT
// Adapted from web_src/js/features/common-page.ts; add expanded state for accessibility.
export function initHeadNavbarContentToggle() {
 const navbar = document.querySelector('#navbar');
 const btn = document.querySelector('#navbar-expand-toggle');
 if (!navbar || !btn) return;
 btn.addEventListener('click', () => {
  const isExpanded = btn.classList.contains('active');
  navbar.classList.toggle('navbar-menu-open', !isExpanded);
  btn.classList.toggle('active', !isExpanded);
  btn.setAttribute('aria-expanded', String(!isExpanded));
 });
}
