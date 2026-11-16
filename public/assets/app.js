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
import { I as e, P as t, a as n, h as r, i, m as a, o, r as s, t as c, u as l, y as u } from "./runtime-core.esm-bundler-B3eh7u8Y.js";
import { i as d, r as f, t as p } from "./runtime-dom.esm-bundler-DA27p68u.js";
import { t as m } from "./jquery.module-D90Hpg2l.js";
//#region web_src/features/repository-task-history.ts
async function h(e) {
	let t = document.querySelector("[data-repository-history]");
	if (!t?.dataset.url) return;
	let n = !1, r;
	window.addEventListener("pagehide", (e) => {
		e.persisted || (n = !0, r?.unmount());
	});
	try {
		let { default: i } = await import("./RepositoryTaskHistory-BBCXI0-h.js");
		if (n) return;
		r = p(i, {
			url: t.dataset.url,
			retryUrl: t.dataset.retryUrl ?? "",
			initForms: e
		}), r.mount(t);
	} catch {
		n || (t.textContent = "任务记录加载失败，请刷新页面重试。");
	}
}
//#endregion
//#region web_src/features/search-clear.ts
function g() {
	for (let e of Array.from(document.querySelectorAll("[data-search-clear]"))) {
		if (e.hasAttribute("data-search-clear-ready")) continue;
		let t = e.querySelector("input"), n = e.querySelector(".search-clear-button");
		if (!t || !n) continue;
		e.setAttribute("data-search-clear-ready", "");
		let r = () => {
			n.hidden = t.value.length === 0;
		};
		t.addEventListener("input", r), t.addEventListener("change", r), t.addEventListener("focus", r), t.form?.addEventListener("reset", () => queueMicrotask(r)), n.addEventListener("click", (n) => {
			if (n.preventDefault(), t.value = "", t.dispatchEvent(new Event("input", { bubbles: !0 })), t.focus(), e.dataset.searchClear === "submit" && t.form) {
				for (let e of Array.from(t.form.querySelectorAll("input[name=\"page\"],input[name=\"repo_page\"]"))) e.value = "1";
				try {
					sessionStorage.setItem("svnova.search-clear-focus", new URL(t.form.action, location.href).pathname + ":" + t.name);
				} catch {}
				t.form.requestSubmit();
			}
		}), r();
		try {
			sessionStorage.getItem("svnova.search-clear-focus") === location.pathname + ":" + t.name && (sessionStorage.removeItem("svnova.search-clear-focus"), t.focus());
		} catch {}
	}
}
//#endregion
//#region web_src/features/task-center.ts
function _() {
	let e = document.querySelector("[data-task-center]");
	if (!e) return;
	let t = e, n = document.querySelector("[data-task-center-total]"), r = document.querySelector("[data-task-center-error]"), i, a = !1, o = () => t.querySelector("[data-task-center-active]")?.dataset.taskCenterActive === "true";
	function s() {
		i && clearTimeout(i), !document.hidden && o() && (i = setTimeout(() => void c(), 5e3));
	}
	async function c() {
		if (!(a || document.hidden)) {
			a = !0;
			try {
				let e = await fetch("/api/v1/tasks" + location.search, {
					headers: { Accept: "application/json" },
					cache: "no-store"
				});
				if (e.status === 401) {
					location.assign("/user/login");
					return;
				}
				if (!e.ok) throw Error("refresh failed");
				let i = await e.json();
				t.innerHTML = i.html, n && (n.textContent = String(i.total));
				let a = new URL(location.href);
				a.searchParams.set("page", String(i.page)), history.replaceState(null, "", a), r && (r.hidden = !0);
			} catch {
				t.replaceChildren(), n && (n.textContent = "—"), r && (r.hidden = !1);
			} finally {
				a = !1, s();
			}
		}
	}
	document.addEventListener("visibilitychange", () => {
		i && clearTimeout(i), document.hidden || c();
	});
	let l = document.querySelector(".task-center-filters");
	l?.querySelectorAll("input[type=radio]").forEach((e) => e.addEventListener("change", () => l.requestSubmit())), s();
}
//#endregion
//#region web_src/features/activity-trend.ts
async function v() {
	let e = document.getElementById("activity-trend"), t = e?.dataset.heatmapUrl;
	if (!e || !t) return;
	let n = !1, r;
	window.addEventListener("pagehide", (e) => {
		e.persisted || (n = !0, r?.unmount());
	});
	try {
		let { default: i } = await import("./ActivityTrend-2hQ5fvvl.js");
		if (n) return;
		r = p(i, { org: new URL(t, location.origin).searchParams.get("org") ?? "0" }), r.mount(e);
	} catch {
		if (n) return;
		e.textContent = "活动趋势加载失败，请刷新页面重试。";
	}
}
//#endregion
//#region web_src/features/repo-migration.ts
function y(e) {
	let t = e.trim(), n = /^https?:\/\//i.test(t);
	if (n) try {
		t = new URL(t).pathname;
	} catch {
		return "";
	}
	let r = t.replace(/[\\/]+$/, "").split(/[\\/]/).pop() || "";
	if (n) try {
		r = decodeURIComponent(r);
	} catch {}
	return /^[a-z]:$/i.test(r) ? "" : r.trim().replace(/[^-.\w]/g, "-").replace(/^[^a-z\d]+/i, "").replace(/\.{2,}/g, ".").slice(0, 80).replace(/\.+$/, "");
}
function b() {
	let e = document.querySelector("#migration-kind"), t = document.querySelector("#clone_addr"), n = document.querySelector("#repo_name"), r = !!n?.value;
	n?.addEventListener("input", () => {
		r = n.value !== "";
	});
	let i = () => {
		t && n && !r && (n.value = y(t.value), n.setCustomValidity(""));
	};
	t?.addEventListener("input", i);
	let a = () => {
		for (let t of Array.from(document.querySelectorAll("[data-kind]"))) {
			t.hidden = t.dataset.kind !== e?.value;
			for (let e of Array.from(t.querySelectorAll("input"))) e.disabled = t.hidden;
		}
		let t = document.querySelector("#migration-source-help");
		t && (t.textContent = e?.value === "local" ? "服务器上的本地绝对目录路径；迁移前须停止原服务访问。" : "现有 SVN 仓库的 HTTP/HTTPS 根地址；认证信息请在下方填写。"), i();
	};
	e?.addEventListener("change", a), a();
}
//#endregion
//#region web_src/features/notifications.ts
function x(e) {
	S(e);
	let t = document.querySelector("[data-notification-settings]");
	if (t) for (let e of Array.from(t.querySelectorAll("[data-notification-credential]"))) {
		let t = document.getElementById(e.dataset.notificationCredential || "");
		if (!(t instanceof HTMLInputElement)) continue;
		let n = () => {
			t.disabled = !e.checked, t.required = e.checked, e.checked || (t.value = "");
		};
		e.addEventListener("change", n), n();
	}
}
function S(e) {
	let t = document.querySelector("#notification-records");
	if (!t) return;
	let n = new AbortController(), r = () => document.hidden || !!document.querySelector("dialog[open]") || t.contains(document.activeElement), i = "", a, o = async () => {
		try {
			if (r()) return;
			let a = await fetch(location.href, {
				headers: { Accept: "text/html" },
				cache: "no-store",
				signal: n.signal
			});
			if (a.status === 401 || a.redirected) {
				location.assign("/user/login");
				return;
			}
			if (!a.ok) return;
			let o = new DOMParser().parseFromString(await a.text(), "text/html"), s = o.querySelector("#notification-records");
			if (!s || n.signal.aborted || r()) return;
			s.innerHTML !== i && (i = s.innerHTML, t.innerHTML = s.innerHTML, e());
			let c = document.querySelector("[data-notification-total]"), l = o.querySelector("[data-notification-total]");
			c && l && (c.textContent = l.textContent);
		} catch (e) {
			e instanceof DOMException && e.name === "AbortError" || console.debug("Notification refresh failed");
		} finally {
			n.signal.aborted || (a = setTimeout(() => void o(), 2e3));
		}
	};
	window.addEventListener("pagehide", () => {
		clearTimeout(a), n.abort();
	}, { once: !0 }), o();
}
async function C(e, t = "确认发送通知", n = "确认发送") {
	let r = document.createElement("dialog");
	r.className = "auth-cleanup-dialog";
	let i = document.createElement("h3");
	i.textContent = t;
	let a = document.createElement("p");
	a.textContent = e;
	let o = document.createElement("div");
	o.className = "cleanup-actions";
	let s = document.createElement("button");
	s.type = "button", s.className = "ui button", s.textContent = "取消";
	let c = document.createElement("button");
	c.type = "button", c.className = "ui primary button", c.textContent = n, o.append(s, c), r.append(i, a, o), document.body.append(r);
	let l = await new Promise((e) => {
		let t = !1, n = (n) => {
			t || (t = !0, r.close(), e(n));
		};
		s.addEventListener("click", () => n(!1)), c.addEventListener("click", () => n(!0)), r.addEventListener("cancel", (e) => {
			e.preventDefault(), n(!1);
		}), r.addEventListener("close", () => {
			t || (t = !0, e(!1));
		}), r.showModal(), s.focus();
	});
	return r.remove(), l;
}
//#endregion
//#region web_src/features/admin-users.ts
function w() {
	for (let e of Array.from(document.querySelectorAll("#user-list-search-form,#notification-list-search-form"))) {
		for (let t of Array.from(e.querySelectorAll("input[type=\"radio\"]"))) t.addEventListener("change", () => e.requestSubmit());
		e.querySelector("[data-user-filter-reset]")?.addEventListener("click", () => {
			for (let t of Array.from(e.querySelectorAll("input[type=\"radio\"]"))) t.checked = !1;
			e.requestSubmit();
		});
		let t = e.querySelector("input[name=\"sort\"]");
		for (let n of Array.from(e.querySelectorAll("[data-user-sort]"))) n.addEventListener("click", () => {
			!t || !n.dataset.userSort || (t.value = n.dataset.userSort, e.requestSubmit());
		});
	}
}
//#endregion
//#region web_src/features/ldap-config.ts
function T() {
	let e = document.querySelector("#auth-paged"), t = document.querySelector("[data-ldap-page-size]"), n = t?.querySelector("input");
	if (!e || !t || !n) return;
	let r = () => {
		t.hidden = !e.checked, n.disabled = !e.checked;
	};
	e.addEventListener("change", r), r();
}
async function E(e, t, n) {
	let r = await fetch(`${e}/preview`, {
		method: "POST",
		headers: {
			"Content-Type": "application/json",
			"X-CSRF-Token": n
		},
		body: JSON.stringify(t)
	});
	if (r.status === 401) return location.assign("/user/login"), !1;
	if (!r.ok) throw Error(await r.text());
	let i = await r.json();
	if (!i.requiresConfirmation) return !0;
	let a = document.createElement("dialog");
	a.className = "auth-cleanup-dialog";
	let o = document.createElement("h3");
	o.textContent = "确认更改 LDAP 配置";
	let s = document.createElement("p");
	s.textContent = `此操作将永久删除 ${i.deleteUsers} 个相关 LDAP 用户、${i.sessions} 个会话、${i.memberships} 条组织成员关系和 ${i.teamMemberships} 条团队成员关系，包括用户邮箱及身份绑定。`;
	let c = document.createElement("p");
	c.textContent = "本地账号（包括 admin）、组织、团队、仓库和同步历史保留。清理后无法通过页面撤销。";
	let l = document.createElement("p");
	l.textContent = "修改 LDAP 配置前将清理绑定的 LDAP 用户。";
	let u = document.createElement("label");
	u.textContent = "请输入“确认清理”后继续";
	let d = document.createElement("input");
	d.type = "text", d.autocomplete = "off", d.setAttribute("aria-label", "确认清理"), u.append(d);
	let f = document.createElement("div");
	f.className = "cleanup-actions";
	let p = document.createElement("button");
	p.type = "button", p.className = "ui button", p.textContent = "取消";
	let m = document.createElement("button");
	m.type = "button", m.className = "ui negative button", m.textContent = "确认清理并保存", m.disabled = !0, d.addEventListener("input", () => {
		m.disabled = d.value !== "确认清理";
	}), f.append(p, m), a.append(o, s, c, l, u, f), document.body.append(a);
	let h = await new Promise((e) => {
		let t = !1, n = (n) => {
			t || (t = !0, a.close(), e(n));
		};
		p.addEventListener("click", () => n(!1)), m.addEventListener("click", () => {
			d.value === "确认清理" && n(!0);
		}), a.addEventListener("cancel", (e) => {
			e.preventDefault(), n(!1);
		}), a.addEventListener("close", () => {
			t || (t = !0, e(!1));
		}), a.showModal(), d.focus();
	});
	return a.remove(), h && (t.ConfirmCleanup = !0, t.CleanupToken = i.token), h;
}
//#endregion
//#region web_src/features/repo-migrate.ts
function D(e) {
	let t = document.querySelector("[data-migration-panel]"), n = document.querySelector("[data-task-status]"), r = t || n, i = r?.dataset.statusUrl;
	if (!r || !i) return;
	let a = !1, o, s = "", c = new AbortController();
	window.addEventListener("pagehide", () => {
		a = !0, c.abort(), o && clearTimeout(o);
	}, { once: !0 });
	let l = (e, t) => {
		let n = r.querySelector(e);
		n && (n.hidden = !t);
	}, u = (e, t) => {
		let n = r.querySelector(e);
		n && (n.textContent = t);
	}, d = async () => {
		let o = await fetch(i, {
			headers: { Accept: "application/json" },
			signal: c.signal
		});
		if (o.status === 401 || o.redirected && new URL(o.url).pathname === "/user/login") return a = !0, location.assign("/user/login"), !1;
		if (o.status === 403 || o.status === 404) {
			if (u("#repo_migrating_failed_error", "仓库不存在或访问权限已变化，请刷新页面"), l("#repo_migrating", !1), l("#repo_migrating_progress", !1), l("#repo_migrating_failed_image", !0), l("#repo_migrating_failed", !0), l("#repo_migrating_retry", !1), n) {
				n.querySelector("tbody")?.replaceChildren();
				let e = n.querySelector("[data-task-connection]");
				e && (e.hidden = !1, e.textContent = "仓库不存在或访问权限已变化，请刷新页面");
			}
			return !1;
		}
		if (!o.ok) throw Error("Migration status unavailable");
		let d = await o.json();
		if (a) return !1;
		if (l("#repo_migrating_connection", !1), l("[data-task-connection]", !1), t) {
			u("#repo_migrating_progress_message", d.message);
			let e = r.querySelector("#repo_migrating_revision_progress");
			if (e && (e.hidden = d.status !== 2, d.progress >= 0 ? e.value = d.progress : e.removeAttribute("value")), d.status === 4) return a = !0, location.reload(), !1;
			if (d.status === 3) return l("#repo_migrating", !1), l("#repo_migrating_progress", !1), l("#repo_migrating_failed_image", !0), l("#repo_migrating_failed", !0), l("#repo_migrating_retry", d.canRetry), u("#repo_migrating_failed_error", d.message), !1;
			l("#repo_migrating", !0), l("#repo_migrating_progress", !0), l("#repo_migrating_failed_image", !1), l("#repo_migrating_failed", !1), l("#repo_migrating_retry", !1);
		} else if (n) {
			if (d.repositoryState !== n.dataset.repositoryState) return a = !0, location.reload(), !1;
			let t = JSON.stringify(d.tasks);
			if (t !== s && !n.contains(document.activeElement)) {
				s = t;
				let r = n.querySelector("tbody");
				if (r) {
					let t = document.createDocumentFragment();
					for (let e of d.tasks) {
						let r = document.createElement("tr");
						[
							`#${e.id}`,
							e.kind,
							e.state,
							e.phase,
							e.started,
							e.finished,
							e.message
						].forEach((t, n) => {
							let i = document.createElement("td");
							if (i.textContent = t, n === 3 && e.indeterminate) {
								let e = document.createElement("progress");
								e.setAttribute("aria-label", "任务进行中"), i.append(e);
							}
							r.append(i);
						});
						let i = document.createElement("td");
						if (e.retry && n.dataset.retryUrl) {
							let e = document.createElement("form");
							e.className = "ui form", e.dataset.api = n.dataset.retryUrl;
							let t = document.createElement("button");
							t.className = "action-icon-button", t.title = "重试任务", t.setAttribute("aria-label", "重试任务");
							let r = document.createElementNS("http://www.w3.org/2000/svg", "svg");
							r.setAttribute("width", "16"), r.setAttribute("height", "16"), r.setAttribute("aria-hidden", "true");
							let a = document.createElementNS("http://www.w3.org/2000/svg", "use");
							a.setAttribute("href", "/assets/img/octicons.svg#issue-reopened"), r.append(a), t.append(r), e.append(t), i.append(e);
						}
						r.append(i), t.append(r);
					}
					if (!d.tasks.length) {
						let e = document.createElement("tr"), n = document.createElement("td");
						n.colSpan = 8, n.textContent = "暂无任务", e.append(n), t.append(e);
					}
					r.replaceChildren(t), e();
				}
			}
		}
		return d.status !== 3 && d.status !== 4;
	}, f = async () => {
		let e = !0;
		try {
			e = await d();
		} catch {
			a || (l("#repo_migrating_connection", !0), l("[data-task-connection]", !0));
		} finally {
			!a && e && (o = setTimeout(f, 2e3));
		}
	};
	f();
}
//#endregion
//#region web_src/components/ActivityHeatmap.vue?vue&type=script&setup=true&lang.ts
var O = ["viewBox"], k = ["transform"], A = ["x", "y"], ee = ["transform"], te = {
	class: "heatmap-day-label",
	x: 0,
	y: 20
}, j = {
	class: "heatmap-day-label",
	x: 0,
	y: 44
}, M = {
	class: "heatmap-day-label",
	x: 0,
	y: 69
}, N = ["transform"], P = ["transform"], F = [
	"transform",
	"aria-label",
	"data-tooltip",
	"onClick",
	"onKeydown"
], I = { class: "heatmap-footer" }, L = { class: "heatmap-legend" }, R = ["x"], z = 10, B = 2, V = 7, H = 365, U = /* @__PURE__ */ l({
	__name: "ActivityHeatmap",
	props: {
		values: {},
		today: {},
		locale: {}
	},
	setup(l) {
		let p = () => "zh-CN", m = l, h = [
			"var(--color-secondary-alpha-60)",
			"var(--color-primary-light-4)",
			"var(--color-primary-light-2)",
			"var(--color-primary)",
			"var(--color-primary-dark-2)",
			"var(--color-primary-dark-4)"
		], g = z + B, _ = Math.ceil(z * 2.5), v = z + z / 2, y = m.today ? /* @__PURE__ */ new Date(`${m.today}T12:00:00`) : /* @__PURE__ */ new Date();
		function b(e) {
			return `${e.getFullYear()}${String(e.getMonth()).padStart(2, "0")}${String(e.getDate()).padStart(2, "0")}`;
		}
		function x(e, t) {
			let n = new Date(e);
			return n.setDate(n.getDate() + t), n;
		}
		let S = s(() => {
			let e = x(y, -H), t = e.getDay(), n = V - 1 - y.getDay(), r = (H + 1 + t + n) / V, i = m.values.length ? Math.max(...m.values.map((e) => e.count)) : 0, a = i > 0 ? Math.ceil(i / 5 * 4) : 1, o = /* @__PURE__ */ new Map();
			for (let { date: e, count: t } of m.values) {
				let n = t >= a ? 4 : Math.max(1, Math.ceil(t / a * 3));
				o.set(b(e), {
					count: t,
					colorIndex: n
				});
			}
			let { on: s } = m.locale.heatMapLocale, { noDataText: c, tooltipUnit: l } = m.locale, u = p(), d = x(e, -t), f = new Date(d.getFullYear(), d.getMonth(), d.getDate()), h = [];
			for (let e = 0; e < r; e++) {
				let e = [];
				for (let t = 0; t < V; t++) {
					let t = o.get(b(f)), n = f.toLocaleDateString(u, {
						year: "numeric",
						month: "short",
						day: "numeric"
					}), r = t ? `${t.count} ${l}` : c;
					e.push({
						date: new Date(f),
						colorIndex: t ? t.colorIndex : 0,
						ariaLabel: `${r} ${s} ${n}`,
						tooltip: `<b>${r}</b> ${s} ${n}`
					}), f.setDate(f.getDate() + 1);
				}
				h.push(e);
			}
			let S = [];
			for (let e = 1; e < h.length; e++) {
				let t = h[e - 1][0].date, n = h[e][0].date;
				t.getMonth() !== n.getMonth() && S.push({
					monthIdx: n.getMonth(),
					weekIdx: e
				});
			}
			return {
				calendar: h,
				monthLabels: S,
				width: _ + g * r + B,
				height: v + g * V
			};
		}), C = `${g} 0 ${z * (h.length + 2)} ${z}`;
		function w(e) {
			let t = new URLSearchParams(document.location.search), n = t.get("date"), r = (/* @__PURE__ */ new Date(e.getTime() - e.getTimezoneOffset() * 6e4)).toISOString().substring(0, 10);
			n && n === r ? t.delete("date") : t.set("date", r), t.delete("page");
			let i = t.toString();
			window.location.search = i.length ? `?${i}` : "";
		}
		return (s, p) => (a(), o("div", null, [(a(), o("svg", {
			class: "heatmap-svg",
			viewBox: `0 0 ${S.value.width} ${S.value.height}`
		}, [
			i("g", {
				class: "heatmap-month-labels",
				transform: `translate(${u(_)}, 0)`
			}, [(a(!0), o(c, null, r(S.value.monthLabels, (t) => (a(), o("text", {
				key: t.weekIdx,
				class: "heatmap-month-label",
				x: g * t.weekIdx,
				y: g - B
			}, e(l.locale.heatMapLocale.months[t.monthIdx]), 9, A))), 128))], 8, k),
			i("g", {
				class: "heatmap-day-labels",
				transform: `translate(0, ${v})`
			}, [
				i("text", te, e(l.locale.heatMapLocale.days[1]), 1),
				i("text", j, e(l.locale.heatMapLocale.days[3]), 1),
				i("text", M, e(l.locale.heatMapLocale.days[5]), 1)
			], 8, ee),
			i("g", {
				class: "heatmap-grid",
				transform: `translate(${u(_)}, ${v})`
			}, [(a(!0), o(c, null, r(S.value.calendar, (s, l) => (a(), o("g", {
				key: l,
				class: "heatmap-week",
				transform: `translate(${l * g}, 0)`
			}, [(a(!0), o(c, null, r(s, (r, s) => (a(), o(c, { key: s }, [r.date < u(y) ? (a(), o("rect", {
				key: 0,
				class: t(["heatmap-day", `heatmap-level-${r.colorIndex}`]),
				transform: `translate(0, ${s * g})`,
				width: z,
				height: z,
				"aria-label": r.ariaLabel,
				"data-tooltip": r.tooltip,
				onClick: (e) => w(r.date),
				tabindex: "0",
				role: "button",
				onKeydown: [f((e) => w(r.date), ["enter"]), f(d((e) => w(r.date), ["prevent"]), ["space"])]
			}, [i("title", null, e(r.ariaLabel), 1)], 42, F)) : n("", !0)], 64))), 128))], 8, P))), 128))], 8, N)
		], 8, O)), i("div", I, [i("div", null, e(l.locale.textTotalContributions), 1), i("div", L, [
			i("div", null, e(l.locale.heatMapLocale.less), 1),
			(a(), o("svg", {
				class: "heatmap-legend-svg",
				viewBox: C,
				height: z
			}, [(a(), o(c, null, r(h, (e, n) => i("rect", {
				key: n,
				width: z,
				height: z,
				x: (n + 1) * g,
				class: t(`heatmap-level-${n}`)
			}, null, 10, R)), 64))])),
			i("div", null, e(l.locale.heatMapLocale.more), 1)
		])])]));
	}
});
//#endregion
//#region web_src/features/heatmap.ts
async function W() {
	let e = document.querySelector("#user-heatmap");
	if (!e) return;
	let t = e.dataset.heatmapUrl;
	if (!t) return;
	let n = e.dataset.activityPending === "true", r, i = !1, a = "", o;
	window.addEventListener("pagehide", () => {
		i = !0, o && clearTimeout(o), r?.unmount();
	}, { once: !0 });
	let s = async () => {
		try {
			let c = await fetch(t);
			if (c.status === 401) {
				location.assign("/user/login");
				return;
			}
			if (!c.ok) throw Error("活动图加载失败");
			let l = await c.json();
			if (i) return;
			let u = JSON.stringify(l), d = u !== a;
			if (d) {
				a = u;
				let t = l.heatmapData.map(([e, t]) => {
					let n = new Intl.DateTimeFormat("sv-SE", {
						timeZone: "Asia/Shanghai",
						year: "numeric",
						month: "2-digit",
						day: "2-digit"
					}).format(/* @__PURE__ */ new Date(e * 1e3));
					return {
						date: /* @__PURE__ */ new Date(`${n}T00:00:00`),
						count: t
					};
				}), n = {
					textTotalContributions: `当前可访问仓库近一年共 ${l.totalContributions.toLocaleString("zh-CN")} 次迁移／同步活动${l.pending ? "（有任务进行中）" : ""}`,
					heatMapLocale: {
						months: Array.from({ length: 12 }, (e, t) => `${t + 1}月`),
						days: [
							"周日",
							"周一",
							"周二",
							"周三",
							"周四",
							"周五",
							"周六"
						],
						on: " · ",
						more: "更多",
						less: "更少"
					},
					tooltipUnit: "次迁移／同步活动",
					noDataText: "无活动"
				};
				r?.unmount(), r = p(U, {
					values: t,
					locale: n,
					today: l.today
				}), r.mount(e), e.classList.remove("is-loading");
			}
			if (d || n || l.pending || l.failed) {
				let e = await fetch(location.href, { headers: { Accept: "text/html" } });
				if (e.redirected && new URL(e.url).pathname === "/user/login") {
					location.assign("/user/login");
					return;
				}
				if (e.ok && !i) {
					let t = new DOMParser().parseFromString(await e.text(), "text/html");
					for (let e of ["activity-feed", "activity-index-status"]) {
						let n = t.getElementById(e), r = document.getElementById(e);
						n && r && !r.contains(document.activeElement) && r.replaceWith(n);
					}
				}
				n = l.pending || l.failed;
			}
			i || (o = setTimeout(s, l.pending ? 5e3 : 15e3));
		} catch {
			if (i) return;
			let e = document.getElementById("activity-index-status");
			e && (e.textContent = "活动统计暂不可用，稍后自动重试"), o = setTimeout(s, 15e3);
		}
	};
	await s();
}
//#endregion
//#region web_src/upstream/fomantic/transition.ts
function G() {
	let e = new Set([
		"clear queue",
		"stop",
		"stop all",
		"destroy",
		"force repaint",
		"repaint",
		"reset",
		"looping",
		"remove looping",
		"disable",
		"enable",
		"set duration",
		"save conditions",
		"restore conditions"
	]);
	m.fn.transition = function(t, n, r) {
		if (t === "is supported") return !0;
		if (t === "is animating" || t === "is inward" || t === "is outward") return !1;
		let i;
		if (typeof t == "string") {
			if (e.has(t)) return this;
			i = {
				animation: t,
				...n && { duration: n },
				...r && { onComplete: r }
			};
		} else if (typeof t == "object") i = t;
		else throw Error(`invalid argument: ${t}`);
		let a = i.animation?.startsWith("show") || i.animation?.endsWith(" in"), o = i.animation?.startsWith("hide") || i.animation?.endsWith(" out");
		return this.each((e, t) => {
			let n = a;
			!a && !o && (n = this.hasClass("hidden")), i.onStart?.call(t), n ? (t.classList.remove("hidden"), t.classList.add("visible", "transition"), i.displayType && t.style.setProperty("display", i.displayType, "important"), i.onShow?.call(t)) : (t.classList.add("hidden"), t.classList.remove("visible"), t.style.removeProperty("display"), i.onHidden?.call(t)), i.onComplete?.call(t);
		}), this;
	};
}
//#endregion
//#region web_src/features/select-dropdown.ts
async function K() {
	Object.assign(window, {
		$: m,
		jQuery: m
	}), await import("./dropdown-DM_WeraS.js");
	let e = m.fn;
	e.fomanticExt = {}, G();
	let { initAriaDropdownPatch: t } = await import("./aria-dropdown-2ERRIU-7.js");
	t(), e.dropdown.settings.fullTextSearch = "exact";
	for (let t of Array.from(document.querySelectorAll("[data-menu-dropdown]"))) e.dropdown.call(m(t), {
		transition: "none",
		action: "nothing"
	});
	for (let t of Array.from(document.querySelectorAll("select"))) {
		let n = t.closest("label");
		t.id ||= `select-${document.querySelectorAll("[data-select-enhanced]").length}`, t.dataset.selectEnhanced = "true", n && (n.htmlFor = t.id, t.setAttribute("aria-label", n.childNodes[0]?.textContent?.trim() || "选择"));
		let r = t.required;
		t.classList.add("ui", "selection", "dropdown"), t.hasAttribute("data-search-select") && t.classList.add("search"), e.dropdown.call(m(t), {
			transition: "none",
			forceSelection: !1,
			match: t.hasAttribute("data-search-select") ? "text" : "both"
		});
		let i = t.closest(".ui.dropdown");
		if (!i) continue;
		t.required = r, t.addEventListener("invalid", (t) => {
			t.preventDefault(), i.classList.add("error"), i.setAttribute("aria-invalid", "true"), i.focus(), e.dropdown.call(m(i), "show");
		}), t.addEventListener("change", () => {
			i.classList.remove("error"), i.removeAttribute("aria-invalid");
		}), t.getAttribute("aria-label") && i.setAttribute("aria-label", t.getAttribute("aria-label") || "选择");
		let a = i.querySelector(".dropdown.icon");
		if (a) {
			let e = document.createElementNS("http://www.w3.org/2000/svg", "svg");
			e.classList.add("dropdown", "icon"), e.setAttribute("width", "14"), e.setAttribute("height", "14"), e.setAttribute("aria-hidden", "true");
			let t = document.createElementNS(e.namespaceURI, "use");
			t.setAttribute("href", "/assets/img/octicons.svg#triangle-down"), e.append(t), a.replaceWith(e);
		}
	}
}
//#endregion
//#region web_src/upstream/navbar.ts
function q() {
	let e = document.querySelector("#navbar"), t = document.querySelector("#navbar-expand-toggle");
	!e || !t || t.addEventListener("click", () => {
		let n = t.classList.contains("active");
		e.classList.toggle("navbar-menu-open", !n), t.classList.toggle("active", !n), t.setAttribute("aria-expanded", String(!n));
	});
}
//#endregion
//#region web_src/features/repository-history.ts
function J() {
	for (let e of Array.from(document.querySelectorAll("[data-history-row]"))) {
		let t = e.querySelector(".history-toggle"), n = document.getElementById(t?.getAttribute("aria-controls") || "");
		!t || !n || e.addEventListener("click", (e) => {
			if (!(e.target instanceof Element) || e.target.closest("a")) return;
			let r = t.getAttribute("aria-expanded") === "true";
			t.setAttribute("aria-expanded", String(!r)), n.hidden = r;
		});
	}
}
//#endregion
//#region web_src/main.ts
for (let e of Array.from(document.querySelectorAll("input[data-name-validation]"))) e.addEventListener("invalid", () => {
	e.validity.patternMismatch && e.setCustomValidity(`${e.dataset.nameValidation}不能包含空格；须以字母或数字开头，只能包含字母、数字、点、下划线和连字符，最长 80 个字符`);
}), e.addEventListener("input", () => e.setCustomValidity(""));
function Y() {
	let e = !1;
	for (let t of Array.from(document.querySelectorAll("form[data-api]:not([data-initialized])"))) t.dataset.initialized = "true", t.addEventListener("submit", async (n) => {
		if (n.preventDefault(), e || !t.reportValidity()) return;
		let r = t.querySelector("[data-confirm]");
		if (r && r.value !== r.dataset.confirm) {
			r.setCustomValidity("请输入完整仓库名称"), r.reportValidity(), r.addEventListener("input", () => r.setCustomValidity(""), { once: !0 });
			return;
		}
		let i = t.querySelector("[data-org-name]");
		if (i && i.value !== i.dataset.orgName) {
			i.setCustomValidity("请输入完整组织名称"), i.reportValidity(), i.addEventListener("input", () => i.setCustomValidity(""), { once: !0 });
			return;
		}
		if (t.dataset.memberRemove && !await C(t.dataset.memberRemove, "确认移除成员", "确认移除") || t.dataset.teamLeave && !await C(t.dataset.teamLeave, "确认离开团队", "确认离开")) return;
		let a = t.dataset.api;
		if (!a) return;
		let o = {};
		for (let e of Array.from(t.querySelectorAll("input[name],select[name],textarea[name]"))) e.disabled || e.closest("fieldset[disabled]") || (o[e.name] = e.type === "checkbox" ? e.checked : e.hasAttribute("data-number") ? Number(e.value) : e.hasAttribute("data-bool") ? e.value === "true" : e.value);
		let s = t.querySelector("[data-delete-error]") || document.querySelector("#action-message"), c = Array.from(document.querySelectorAll("form[data-api] button")), l = c.map((e) => e.disabled);
		e = !0, c.forEach((e) => {
			e.disabled = !0;
		});
		try {
			if (t.dataset.leaveOrganization && !await ne(t.dataset.leaveOrganization) || t.dataset.notificationConfirm && !await C(t.dataset.notificationConfirm)) return;
			let e = document.querySelector("meta[name=\"csrf-token\"]")?.content;
			if (t.hasAttribute("data-ldap-config") && !await E(a, o, e || "")) return;
			t.querySelectorAll("[data-sync-action]").forEach((e) => e.classList.add("is-pending"));
			let n = await fetch(a, {
				method: t.dataset.method || "POST",
				headers: {
					"Content-Type": "application/json",
					"X-CSRF-Token": e || ""
				},
				body: JSON.stringify(o)
			});
			if (n.status === 401) {
				location.assign("/user/login");
				return;
			}
			if (!n.ok) throw Error(await n.text());
			if (t.hasAttribute("data-next-location")) {
				let e = n.headers.get("Location");
				if (!e || new URL(e, location.origin).origin !== location.origin) throw Error("仓库已创建，请刷新仓库列表查看");
				location.assign(new URL(e, location.origin).href);
				return;
			}
			location.assign(t.dataset.next || location.href);
		} catch (e) {
			s && (s.hidden = !1, s.className = "ui negative message", s.textContent = e instanceof Error ? e.message : "操作失败，请重试", s.scrollIntoView({ block: "nearest" }));
		} finally {
			t.querySelectorAll("[data-sync-action]").forEach((e) => e.classList.remove("is-pending")), e = !1, c.forEach((e, t) => {
				e.disabled = l[t] ?? !1;
			});
		}
	});
}
Y(), x(Y), w(), T(), b(), K(), D(Y), h(Y), q();
function X() {
	let e = new URL(location.href);
	for (let t of Array.from(document.querySelectorAll(".tabs a"))) {
		let n = new URL(t.href);
		n.pathname === e.pathname && (n.searchParams.get("tab") || "tree") === (e.searchParams.get("tab") || "tree") && t.setAttribute("aria-current", "page");
	}
}
X();
function Z() {
	document.addEventListener("click", (e) => {
		for (let t of Array.from(document.querySelectorAll("details.nav-dropdown[open]"))) e.target instanceof Node && !t.contains(e.target) && (t.open = !1);
	}), document.addEventListener("keydown", (e) => {
		if (e.key === "Escape") for (let e of Array.from(document.querySelectorAll("details.nav-dropdown[open]"))) e.open = !1, e.querySelector("summary")?.focus();
	});
}
Z(), W(), v();
function ne(e) {
	let t = document.querySelector("#leave-organization-dialog"), n = t?.querySelector("[data-organization-name]"), r = t?.querySelector("[data-cancel]"), i = t?.querySelector("[data-accept]");
	return !t || !n || !r || !i ? Promise.resolve(!1) : (n.textContent = e, new Promise((e) => {
		let n = (n) => {
			t.close(), r.removeEventListener("click", a), i.removeEventListener("click", o), t.removeEventListener("cancel", s), e(n);
		}, a = () => n(!1), o = () => n(!0), s = (e) => {
			e.preventDefault(), n(!1);
		};
		r.addEventListener("click", a), i.addEventListener("click", o), t.addEventListener("cancel", s), t.showModal(), r.focus();
	}));
}
var Q = document.querySelector("#delete-organization-dialog");
document.querySelector("#open-delete-organization")?.addEventListener("click", () => {
	Q?.showModal(), Q?.querySelector("input")?.focus();
}), Q?.querySelector("[data-delete-cancel]")?.addEventListener("click", () => Q.close());
var $ = document.querySelector("#delete-user-dialog");
document.querySelector("#open-delete-user")?.addEventListener("click", () => {
	$?.showModal(), $?.querySelector("input")?.focus();
}), $?.querySelector("[data-delete-user-cancel]")?.addEventListener("click", () => $.close());
for (let e of Array.from(document.querySelectorAll(".member-search-box"))) {
	let t = e.querySelector("[data-member-search]"), n = e.querySelector("[data-member-suggestions]");
	if (!t || !n) continue;
	let r = () => {
		n.hidden = !0, t.setAttribute("aria-expanded", "false");
	}, i = Array.from(n.querySelectorAll("[data-member-name]")), a = () => {
		let e = t.value.trim().toLocaleLowerCase(), r = 0;
		for (let t of i) t.hidden = r >= 20 || !t.dataset.memberMatch?.toLocaleLowerCase().includes(e), t.hidden || r++;
		let a = n.querySelector("[data-member-empty]");
		a && (a.hidden = r > 0), n.hidden = !1, t.setAttribute("aria-expanded", "true");
	};
	t.addEventListener("input", a), t.addEventListener("focus", a);
	for (let e of i) e.addEventListener("click", () => {
		t.value = e.dataset.memberName ?? "", t.dispatchEvent(new Event("input", { bubbles: !0 })), r(), t.focus(), r();
	});
	e.addEventListener("keydown", (e) => {
		if (e.key === "Escape") r(), t.focus(), r();
		else if (e.key === "ArrowDown") {
			e.preventDefault(), n.hidden && a();
			let t = i.filter((e) => !e.hidden);
			t[(t.indexOf(document.activeElement) + 1) % t.length]?.focus();
		}
	}), document.addEventListener("click", (t) => {
		e.contains(t.target) || r();
	}), e.addEventListener("focusout", (t) => {
		e.contains(t.relatedTarget) || r();
	});
}
_(), J(), window.addEventListener("pageshow", () => {
	document.querySelectorAll("[data-refresh-action]").forEach((e) => {
		e.classList.remove("is-pending"), e.removeAttribute("aria-busy");
	});
}), document.addEventListener("click", (e) => {
	if (e.defaultPrevented || e.button !== 0 || e.ctrlKey || e.metaKey || e.shiftKey || e.altKey) return;
	let t = e.target instanceof Element ? e.target.closest("a[data-refresh-action]") : null;
	t && (t.classList.add("is-pending"), t.setAttribute("aria-busy", "true"));
}), g();
//#endregion
