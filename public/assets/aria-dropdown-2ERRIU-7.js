/*!
MIT License

Copyright (c) 2026 GitHub Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/
import { t as e } from "./jquery.module-D90Hpg2l.js";
//#region web_src/upstream/fomantic/dom.ts
var t = 0;
function n(e) {
	return `${e}${++t}`;
}
function r(e, t, n) {
	e.querySelectorAll(t).forEach(n);
}
//#endregion
//#region web_src/upstream/fomantic/aria-dropdown.ts
var i = "_giteaAriaPatchDropdown", a = e.fn.dropdown;
function o() {
	if (e.fn.dropdown === s) throw Error("initAriaDropdownPatch could only be called once");
	e.fn.dropdown = s, e.fn.fomanticExt.onDropdownAfterFiltered = u, s.settings = a.settings;
}
function s(...t) {
	let n = a.apply(this, t);
	for (let t of this) {
		t = t.closest(".ui.dropdown"), t[i] || p(t);
		let n = e(t), r = n.data("module-dropdown");
		r.giteaDelegated || (r.giteaDelegated = !0, d(n));
	}
	return n;
}
function c(e, t) {
	t.id ||= n("_aria_dropdown_item_"), t.setAttribute("role", e[i].listItemRole), t.setAttribute("tabindex", "-1");
	for (let e of t.querySelectorAll("a, input, button")) e.setAttribute("tabindex", "-1");
}
function l(e) {
	e.id ||= n("_aria_dropdown_label_"), e.tabIndex = -1;
	let t = e.querySelector(".delete.icon");
	t && (t.setAttribute("aria-hidden", "false"), t.setAttribute("aria-label", window.config.i18n.remove_label_str.replace("%s", e.getAttribute("data-value"))), t.setAttribute("role", "button"));
}
function u() {
	let t = e(this).closest(".ui.dropdown"), n = t.dropdown("setting", "hideDividers") === "empty", r = t[0].querySelector(".scrolling.menu") || t[0].querySelector(".menu");
	n && r && h(r);
}
function d(t) {
	let n = a.bind(t), r = {
		...n("setting", "templates"),
		t: performance.now()
	}, o = r.menu;
	r.menu = function(n, r, a, s) {
		let l = o(n, r, a, s), u = document.createElement("div");
		u.innerHTML = l;
		let d = e(u);
		return d.find("> .item").each((e, n) => c(t[0], n)), t[0][i].deferredRefreshAriaActiveItem(), d.html();
	}, n("setting", "templates", r);
	let s = n("setting", "onLabelCreate");
	n("setting", "onLabelCreate", function(e, t) {
		let n = s.call(this, e, t);
		return l(n[0]), n;
	});
	let u = n("internal", "set"), d = u.direction;
	u.direction = function(e) {
		d.call(this, e);
		let r = n("setting", "className");
		e ||= t.find("> .menu");
		let i = e[0];
		i.classList.add(r.loading), i.getBoundingClientRect().right > document.documentElement.clientWidth && i.classList.add(r.leftward), i.classList.remove(r.loading);
	};
}
function f(t, r, a) {
	a.id ||= n("_aria_dropdown_menu_"), e(a).find("> .item").each((e, n) => c(t, n)), a.setAttribute("role", t[i].listPopupRole);
	for (let e of t.querySelectorAll(".ui.label")) l(e);
	r.setAttribute("role", r.getAttribute("role") ?? t[i].focusableRole), r.setAttribute("aria-haspopup", t[i].listPopupRole), r.setAttribute("aria-controls", a.id), r.setAttribute("aria-expanded", "false");
	let o = t.getAttribute("data-tooltip-content");
	o && !t.getAttribute("aria-label") && t.setAttribute("aria-label", o);
}
function p(t) {
	t[i] = {};
	let n = t.querySelector("input.search"), r = n || t;
	if (!r) return;
	n && !n.getAttribute("autocomplete") && n.setAttribute("autocomplete", "off");
	let a = e(t).find("> .menu")[0];
	a || (a = document.createElement("div"), a.classList.add("menu"), t.append(a));
	let o = t.querySelectorAll("input, select").length > 0;
	t[i].focusableRole = o ? "combobox" : "menu", t[i].listPopupRole = o ? "listbox" : "", t[i].listItemRole = o ? "option" : "menuitem", m(t, r, a), f(t, r, a);
}
function m(t, n, r) {
	let a = () => r.classList.contains("visible") && !r.classList.contains("out") || r.classList.contains("in"), o = () => {
		let o = a();
		n.setAttribute("aria-expanded", o ? "true" : "false");
		let s = e(r).find("> .item.active, > .item.selected")[0];
		s && (o ? n.setAttribute("aria-activedescendant", s.id) : t[i].listPopupRole === "menu" && (n.removeAttribute("aria-activedescendant"), s.classList.remove("active", "selected")));
	};
	t.addEventListener("keydown", (e) => {
		if (!e.isComposing && e.key === "Enter") {
			let t = r.querySelector(":scope > .item.selected, .menu > .item.selected");
			t?.matches("a, .js-aria-clickable") && !t.matches(".tw-hidden, .filtered") && (e.preventDefault(), t.click());
		}
	});
	let s = (e = 0) => {
		setTimeout(o, e);
	};
	t[i].deferredRefreshAriaActiveItem = s, t.addEventListener("keyup", (e) => {
		e.key.startsWith("Arrow") && s();
	});
	let c = 0, l = 0;
	t.addEventListener("mousedown", () => {
		l += a() ? 1 : 0, c++;
	}, { capture: !0 }), t.addEventListener("focus", () => {
		l += a() ? 1 : 0, c++, s();
	}, { capture: !0 }), t.addEventListener("blur", () => {
		l = c = 0, s(100);
	}, { capture: !0 }), t.addEventListener("mouseup", () => {
		setTimeout(() => {
			l = c = 0, s(100);
		}, 0);
	}, { capture: !0 }), t.addEventListener("click", (e) => {
		a() && l !== 2 && c === 2 && e.stopPropagation(), c = l = 0;
	}, { capture: !0 });
}
function h(e) {
	let t = [], n = [], i = "", a = "", o = (e) => e.classList.contains("divider"), s = (e) => o(e) && e.hasAttribute("data-scope"), c = (e) => e.classList.add("hidden"), l = (e) => e.classList.remove("hidden"), u = (e) => e.classList.contains("hidden") || e.classList.contains("filtered") || e.classList.contains("tw-hidden"), d = (e) => {
		n.length === 1 && s(n[0]) ? c(n[0]) : n.length && (s(n[0]) && a === i && (c(n[0]), n.shift()), t.push(...n), a = i), i = e, n.length = 0;
	};
	r(e, ".divider", l);
	for (let t of e.children) {
		let e = t.getAttribute("data-scope") || "";
		e !== i && d(e), u(t) || n.push(t);
	}
	for (d(""); t.length && o(t[0]);) c(t[0]), t.shift();
	for (; t.length && o(t[t.length - 1]);) c(t[t.length - 1]), t.pop();
	for (let e = 0; e < t.length - 1; e++) t[e].matches(".divider") && t[e + 1].matches(".divider") && c(t[e]);
}
//#endregion
export { o as initAriaDropdownPatch };
