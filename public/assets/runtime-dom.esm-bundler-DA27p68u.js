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
import { A as e, C as t, D as n, E as r, M as i, N as a, O as o, S as s, T as c, b as l, j as u, k as d, n as f, s as p, w as m, x as h } from "./runtime-core.esm-bundler-B3eh7u8Y.js";
//#region node_modules/.pnpm/@vue+runtime-dom@3.5.24/node_modules/@vue/runtime-dom/dist/runtime-dom.esm-bundler.js
var g = void 0, _ = typeof window < "u" && window.trustedTypes;
if (_) try {
	g = /* @__PURE__ */ _.createPolicy("vue", { createHTML: (e) => e });
} catch {}
var v = g ? (e) => g.createHTML(e) : (e) => e, y = "http://www.w3.org/2000/svg", b = "http://www.w3.org/1998/Math/MathML", x = typeof document < "u" ? document : null, S = x && /* @__PURE__ */ x.createElement("template"), C = {
	insert: (e, t, n) => {
		t.insertBefore(e, n || null);
	},
	remove: (e) => {
		let t = e.parentNode;
		t && t.removeChild(e);
	},
	createElement: (e, t, n, r) => {
		let i = t === "svg" ? x.createElementNS(y, e) : t === "mathml" ? x.createElementNS(b, e) : n ? x.createElement(e, { is: n }) : x.createElement(e);
		return e === "select" && r && r.multiple != null && i.setAttribute("multiple", r.multiple), i;
	},
	createText: (e) => x.createTextNode(e),
	createComment: (e) => x.createComment(e),
	setText: (e, t) => {
		e.nodeValue = t;
	},
	setElementText: (e, t) => {
		e.textContent = t;
	},
	parentNode: (e) => e.parentNode,
	nextSibling: (e) => e.nextSibling,
	querySelector: (e) => x.querySelector(e),
	setScopeId(e, t) {
		e.setAttribute(t, "");
	},
	insertStaticContent(e, t, n, r, i, a) {
		let o = n ? n.previousSibling : t.lastChild;
		if (i && (i === a || i.nextSibling)) for (; t.insertBefore(i.cloneNode(!0), n), !(i === a || !(i = i.nextSibling)););
		else {
			S.innerHTML = v(r === "svg" ? `<svg>${e}</svg>` : r === "mathml" ? `<math>${e}</math>` : e);
			let i = S.content;
			if (r === "svg" || r === "mathml") {
				let e = i.firstChild;
				for (; e.firstChild;) i.appendChild(e.firstChild);
				i.removeChild(e);
			}
			t.insertBefore(i, n);
		}
		return [o ? o.nextSibling : t.firstChild, n ? n.previousSibling : t.lastChild];
	}
}, w = Symbol("_vtc");
function T(e, t, n) {
	let r = e[w];
	r && (t = (t ? [t, ...r] : [...r]).join(" ")), t == null ? e.removeAttribute("class") : n ? e.setAttribute("class", t) : e.className = t;
}
var E = Symbol("_vod"), D = Symbol("_vsh"), O = Symbol(""), k = /(?:^|;)\s*display\s*:/;
function A(e, t, n) {
	let r = e.style, i = u(n), a = !1;
	if (n && !i) {
		if (t) if (u(t)) for (let e of t.split(";")) {
			let t = e.slice(0, e.indexOf(":")).trim();
			n[t] ?? M(r, t, "");
		}
		else for (let e in t) n[e] ?? M(r, e, "");
		for (let e in n) e === "display" && (a = !0), M(r, e, n[e]);
	} else if (i) {
		if (t !== n) {
			let e = r[O];
			e && (n += ";" + e), r.cssText = n, a = k.test(n);
		}
	} else t && e.removeAttribute("style");
	E in e && (e[E] = a ? r.display : "", e[D] && (r.display = "none"));
}
var j = /\s*!important$/;
function M(e, n, i) {
	if (r(i)) i.forEach((t) => M(e, n, t));
	else if (i ??= "", n.startsWith("--")) e.setProperty(n, i);
	else {
		let r = F(e, n);
		j.test(i) ? e.setProperty(t(r), i.replace(j, ""), "important") : e[r] = i;
	}
}
var N = [
	"Webkit",
	"Moz",
	"ms"
], P = {};
function F(e, t) {
	let n = P[t];
	if (n) return n;
	let r = l(t);
	if (r !== "filter" && r in e) return P[t] = r;
	r = h(r);
	for (let n = 0; n < N.length; n++) {
		let i = N[n] + r;
		if (i in e) return P[t] = i;
	}
	return t;
}
var I = "http://www.w3.org/1999/xlink";
function L(t, n, r, a, o, s = e(n)) {
	a && n.startsWith("xlink:") ? r == null ? t.removeAttributeNS(I, n.slice(6, n.length)) : t.setAttributeNS(I, n, r) : r == null || s && !m(r) ? t.removeAttribute(n) : t.setAttribute(n, s ? "" : i(r) ? String(r) : r);
}
function R(e, t, n, r, i) {
	if (t === "innerHTML" || t === "textContent") {
		n != null && (e[t] = t === "innerHTML" ? v(n) : n);
		return;
	}
	let a = e.tagName;
	if (t === "value" && a !== "PROGRESS" && !a.includes("-")) {
		let r = a === "OPTION" ? e.getAttribute("value") || "" : e.value, i = n == null ? e.type === "checkbox" ? "on" : "" : String(n);
		(r !== i || !("_value" in e)) && (e.value = i), n ?? e.removeAttribute(t), e._value = n;
		return;
	}
	let o = !1;
	if (n === "" || n == null) {
		let r = typeof e[t];
		r === "boolean" ? n = m(n) : n == null && r === "string" ? (n = "", o = !0) : r === "number" && (n = 0, o = !0);
	}
	try {
		e[t] = n;
	} catch {}
	o && e.removeAttribute(i || t);
}
function z(e, t, n, r) {
	e.addEventListener(t, n, r);
}
function B(e, t, n, r) {
	e.removeEventListener(t, n, r);
}
var V = Symbol("_vei");
function H(e, t, n, r, i = null) {
	let a = e[V] || (e[V] = {}), o = a[t];
	if (r && o) o.value = r;
	else {
		let [n, s] = W(t);
		r ? z(e, n, a[t] = ee(r, i), s) : o && (B(e, n, o, s), a[t] = void 0);
	}
}
var U = /(?:Once|Passive|Capture)$/;
function W(e) {
	let n;
	if (U.test(e)) {
		n = {};
		let t;
		for (; t = e.match(U);) e = e.slice(0, e.length - t[0].length), n[t[0].toLowerCase()] = !0;
	}
	return [e[2] === ":" ? e.slice(3) : t(e.slice(2)), n];
}
var G = 0, K = /* @__PURE__ */ Promise.resolve(), q = () => G ||= (K.then(() => G = 0), Date.now());
function ee(e, t) {
	let n = (e) => {
		if (!e._vts) e._vts = Date.now();
		else if (e._vts <= n.attached) return;
		f(J(e, n.value), t, 5, [e]);
	};
	return n.value = e, n.attached = q(), n;
}
function J(e, t) {
	if (r(t)) {
		let n = e.stopImmediatePropagation;
		return e.stopImmediatePropagation = () => {
			n.call(e), e._stopped = !0;
		}, t.map((e) => (t) => !t._stopped && e && e(t));
	} else return t;
}
var Y = (e) => e.charCodeAt(0) === 111 && e.charCodeAt(1) === 110 && e.charCodeAt(2) > 96 && e.charCodeAt(2) < 123, X = (e, t, n, r, i, a) => {
	let s = i === "svg";
	t === "class" ? T(e, r, s) : t === "style" ? A(e, n, r) : d(t) ? o(t) || H(e, t, n, r, a) : (t[0] === "." ? (t = t.slice(1), !0) : t[0] === "^" ? (t = t.slice(1), !1) : Z(e, t, r, s)) ? (R(e, t, r), !e.tagName.includes("-") && (t === "value" || t === "checked" || t === "selected") && L(e, t, r, s, a, t !== "value")) : e._isVueCE && (/[A-Z]/.test(t) || !u(r)) ? R(e, l(t), r, a, t) : (t === "true-value" ? e._trueValue = r : t === "false-value" && (e._falseValue = r), L(e, t, r, s));
};
function Z(e, t, r, i) {
	if (i) return !!(t === "innerHTML" || t === "textContent" || t in e && Y(t) && n(r));
	if (t === "spellcheck" || t === "draggable" || t === "translate" || t === "autocorrect" || t === "sandbox" && e.tagName === "IFRAME" || t === "form" || t === "list" && e.tagName === "INPUT" || t === "type" && e.tagName === "TEXTAREA") return !1;
	if (t === "width" || t === "height") {
		let t = e.tagName;
		if (t === "IMG" || t === "VIDEO" || t === "CANVAS" || t === "SOURCE") return !1;
	}
	return Y(t) && u(r) ? !1 : t in e;
}
var Q = (e) => {
	let t = e.props["onUpdate:modelValue"] || !1;
	return r(t) ? (e) => c(t, e) : t;
}, $ = Symbol("_assign"), te = {
	created(e, { value: t }, n) {
		e.checked = a(t, n.props.value), e[$] = Q(n), z(e, "change", () => {
			e[$](ne(e));
		});
	},
	beforeUpdate(e, { value: t, oldValue: n }, r) {
		e[$] = Q(r), t !== n && (e.checked = a(t, r.props.value));
	}
};
function ne(e) {
	return "_value" in e ? e._value : e.value;
}
var re = [
	"ctrl",
	"shift",
	"alt",
	"meta"
], ie = {
	stop: (e) => e.stopPropagation(),
	prevent: (e) => e.preventDefault(),
	self: (e) => e.target !== e.currentTarget,
	ctrl: (e) => !e.ctrlKey,
	shift: (e) => !e.shiftKey,
	alt: (e) => !e.altKey,
	meta: (e) => !e.metaKey,
	left: (e) => "button" in e && e.button !== 0,
	middle: (e) => "button" in e && e.button !== 1,
	right: (e) => "button" in e && e.button !== 2,
	exact: (e, t) => re.some((n) => e[`${n}Key`] && !t.includes(n))
}, ae = (e, t) => {
	let n = e._withMods ||= {}, r = t.join(".");
	return n[r] || (n[r] = ((n, ...r) => {
		for (let e = 0; e < t.length; e++) {
			let r = ie[t[e]];
			if (r && r(n, t)) return;
		}
		return e(n, ...r);
	}));
}, oe = {
	esc: "escape",
	space: " ",
	up: "arrow-up",
	left: "arrow-left",
	right: "arrow-right",
	down: "arrow-down",
	delete: "backspace"
}, se = (e, n) => {
	let r = e._withKeys ||= {}, i = n.join(".");
	return r[i] || (r[i] = ((r) => {
		if (!("key" in r)) return;
		let i = t(r.key);
		if (n.some((e) => e === i || oe[e] === i)) return e(r);
	}));
}, ce = /* @__PURE__ */ s({ patchProp: X }, C), le;
function ue() {
	return le ||= p(ce);
}
var de = ((...e) => {
	let t = ue().createApp(...e), { mount: r } = t;
	return t.mount = (e) => {
		let i = pe(e);
		if (!i) return;
		let a = t._component;
		!n(a) && !a.render && !a.template && (a.template = i.innerHTML), i.nodeType === 1 && (i.textContent = "");
		let o = r(i, !1, fe(i));
		return i instanceof Element && (i.removeAttribute("v-cloak"), i.setAttribute("data-v-app", "")), o;
	}, t;
});
function fe(e) {
	if (e instanceof SVGElement) return "svg";
	if (typeof MathMLElement == "function" && e instanceof MathMLElement) return "mathml";
}
function pe(e) {
	return u(e) ? document.querySelector(e) : e;
}
//#endregion
export { ae as i, te as n, se as r, de as t };
