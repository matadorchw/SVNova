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
import { $ as e, $t as t, A as n, At as r, B as i, Bt as a, C as o, Ct as s, D as c, Dt as l, E as u, Et as d, F as f, Ft as p, G as m, Gt as h, H as g, Ht as _, I as v, It as y, J as b, Jt as x, K as S, Kt as C, L as w, M as T, Mt as E, N as D, Nt as O, O as k, Ot as A, P as j, Pt as M, Q as N, Qt as P, R as F, S as I, St as ee, T as te, Tt as L, Ut as R, V as ne, Vt as re, W as ie, Wt as ae, Xt as oe, Y as se, Yt as ce, Z as le, Zt as ue, _ as de, _n as fe, _t as pe, a as me, an as he, at as ge, b as _e, bn as z, bt as ve, cn as ye, ct as be, d as xe, dn as Se, dt as Ce, en as we, et as Te, f as Ee, fn as De, ft as Oe, g as ke, gn as Ae, gt as je, h as Me, hn as Ne, ht as Pe, i as Fe, in as Ie, k as Le, kt as Re, ln as ze, m as Be, mn as Ve, mt as He, n as Ue, nn as We, nt as Ge, o as Ke, on as B, ot as qe, p as Je, pn as Ye, pt as Xe, q as Ze, qt as Qe, r as $e, rn as V, rt as et, s as tt, sn as nt, st as rt, t as it, tn as at, tt as ot, un as st, ut as ct, v as lt, vn as ut, vt as dt, w as ft, wt as pt, x as mt, xn as ht, xt as gt, y as _t, yn as vt, z as yt, zt as bt } from "./_plugin-vue_export-helper-C97cB3Z4.js";
import { F as xt, I as H, a as U, c as St, d as Ct, f as wt, h as Tt, i as W, l as Et, m as G, o as K, p as Dt, t as Ot, u as kt, v as At, y as jt } from "./runtime-core.esm-bundler-B3eh7u8Y.js";
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/line/LineSeries.js
var Mt = function(e) {
	z(t, e);
	function t() {
		var n = e !== null && e.apply(this, arguments) || this;
		return n.type = t.type, n.hasSymbolVisual = !0, n;
	}
	return t.prototype.getInitialData = function(e) {
		return I(null, this, { useEncodeDefaulter: !0 });
	}, t.prototype.getLegendIcon = function(e) {
		var t = new y(), n = c("line", 0, e.itemHeight / 2, e.itemWidth, 0, e.lineStyle.stroke, !1);
		t.add(n), n.setStyle(e.lineStyle);
		var r = this.getData().getVisual("symbol"), i = this.getData().getVisual("symbolRotate"), a = r === "none" ? "circle" : r, o = e.itemHeight * .8, s = c(a, (e.itemWidth - o) / 2, (e.itemHeight - o) / 2, o, o, e.itemStyle.fill);
		return t.add(s), s.setStyle(e.itemStyle), s.rotation = (e.iconRotate === "inherit" ? i : e.iconRotate || 0) * Math.PI / 180, s.setOrigin([e.itemWidth / 2, e.itemHeight / 2]), a.indexOf("empty") > -1 && (s.style.stroke = s.style.fill, s.style.fill = v.color.neutral00, s.style.lineWidth = 2), t;
	}, t.type = "series.line", t.dependencies = ["grid", "polar"], t.defaultOption = {
		z: 3,
		coordinateSystem: "cartesian2d",
		legendHoverLink: !0,
		clip: !0,
		label: { position: "top" },
		endLabel: {
			show: !1,
			valueAnimation: !0,
			distance: 8
		},
		lineStyle: {
			width: 2,
			type: "solid"
		},
		emphasis: { scale: !0 },
		step: !1,
		smooth: !1,
		smoothMonotone: null,
		symbol: "emptyCircle",
		symbolSize: 6,
		symbolRotate: null,
		showSymbol: !0,
		showAllSymbol: "auto",
		connectNulls: !1,
		sampling: "none",
		animationEasing: "linear",
		progressive: 0,
		hoverLayerThreshold: Infinity,
		universalTransition: { divideShape: "clone" },
		triggerLineEvent: !1,
		triggerEvent: !1
	}, t;
}(f), Nt = function(e) {
	z(t, e);
	function t(t, n, r, i) {
		var a = e.call(this) || this;
		return a.updateData(t, n, r, i), a;
	}
	return t.prototype._createSymbol = function(e, t, n, r, i, a) {
		this.removeAll();
		var o = c(e, -1, -1, 2, 2, null, a);
		o.attr({
			z2: ut(i, 100),
			culling: !0,
			scaleX: r[0] / 2,
			scaleY: r[1] / 2
		}), o.drift = Pt, this._symbolType = e, this.add(o);
	}, t.prototype.stopSymbolAnimation = function(e) {
		this.childAt(0).stopAnimation(null, e);
	}, t.prototype.getSymbolType = function() {
		return this._symbolType;
	}, t.prototype.getSymbolPath = function() {
		return this.childAt(0);
	}, t.prototype.highlight = function() {
		Ce(this.childAt(0));
	}, t.prototype.downplay = function() {
		Oe(this.childAt(0));
	}, t.prototype.setZ = function(e, t) {
		var n = this.childAt(0);
		n.zlevel = e, n.z = t;
	}, t.prototype.setDraggable = function(e, t) {
		var n = this.childAt(0);
		n.draggable = e, n.cursor = !t && e ? "move" : n.cursor;
	}, t.prototype.updateData = function(e, n, r, i) {
		this.silent = !1;
		var a = e.getItemVisual(n, "symbol") || "circle", o = e.hostModel, s = t.getSymbolSize(e, n), c = t.getSymbolZ2(e, n), l = a !== this._symbolType, u = i && i.disableAnimation;
		if (l) {
			var d = e.getItemVisual(n, "symbolKeepAspect");
			this._createSymbol(a, e, n, s, c, d);
		} else {
			var f = this.childAt(0);
			f.silent = !1;
			var p = {
				scaleX: s[0] / 2,
				scaleY: s[1] / 2
			};
			u ? f.attr(p) : qe(f, p, o, n), ge(f);
		}
		if (this._updateCommon(e, n, s, r, i), l) {
			var f = this.childAt(0);
			if (!u) {
				var p = {
					scaleX: this._sizeX,
					scaleY: this._sizeY,
					style: { opacity: f.style.opacity }
				};
				f.scaleX = f.scaleY = 0, f.style.opacity = 0, Ge(f, p, o, n);
			}
		}
		u && this.childAt(0).stopAnimation("leave");
	}, t.prototype._updateCommon = function(e, t, n, r, i) {
		var a = this.childAt(0), o = e.hostModel, s, c, l, u, d, f, p, m, h;
		if (r && (s = r.emphasisItemStyle, c = r.blurItemStyle, l = r.selectItemStyle, u = r.focus, d = r.blurScope, p = r.labelStatesModels, m = r.hoverScale, h = r.cursorStyle, f = r.emphasisDisabled), !r || e.hasItemOption) {
			var g = r && r.itemModel ? r.itemModel : e.getItemModel(t), _ = g.getModel("emphasis");
			s = _.getModel("itemStyle").getItemStyle(), l = g.getModel(["select", "itemStyle"]).getItemStyle(), c = g.getModel(["blur", "itemStyle"]).getItemStyle(), u = _.get("focus"), d = _.get("blurScope"), f = _.get("disabled"), p = Ze(g), m = _.getShallow("scale"), h = g.getShallow("cursor");
		}
		var v = e.getItemVisual(t, "symbolRotate");
		a.attr("rotation", (v || 0) * Math.PI / 180 || 0);
		var y = k(e.getItemVisual(t, "symbolOffset"), n);
		y && (a.x = y[0], a.y = y[1]), h && a.attr("cursor", h);
		var b = e.getItemVisual(t, "style"), x = b.fill;
		if (a instanceof ve) {
			var S = a.style;
			a.useStyle(Ie({
				image: S.image,
				x: S.x,
				y: S.y,
				width: S.width,
				height: S.height
			}, b));
		} else a.__isEmptyBrush ? a.useStyle(Ie({}, b)) : a.useStyle(b), a.style.decal = null, a.setColor(x, i && i.symbolInnerColor), a.style.strokeNoScale = !0;
		var C = e.getItemVisual(t, "liftZ"), w = this._z2;
		C == null ? w != null && (a.z2 = w, this._z2 = null) : w ?? (this._z2 = a.z2, a.z2 += C);
		var T = i && i.useNameLabel;
		se(a, p, {
			labelFetcher: o,
			labelDataIndex: t,
			defaultText: E,
			inheritColor: x,
			defaultOpacity: b.opacity
		});
		function E(t) {
			return T ? e.getName(t) : lt(e, t);
		}
		this._sizeX = n[0] / 2, this._sizeY = n[1] / 2;
		var D = a.ensureState("emphasis");
		D.style = s, a.ensureState("select").style = l, a.ensureState("blur").style = c;
		var O = m == null || m === !0 ? Math.max(1.1, 3 / this._sizeY) : isFinite(m) && m > 0 ? +m : 1;
		D.scaleX = this._sizeX * O, D.scaleY = this._sizeY * O, this.setSymbolScale(1), je(this, u, d, f);
	}, t.prototype.setSymbolScale = function(e) {
		this.scaleX = this.scaleY = e;
	}, t.prototype.fadeOut = function(e, t, n) {
		var r = this.childAt(0), i = pe(this).dataIndex, a = n && n.animation;
		if (this.silent = r.silent = !0, n && n.fadeLabel) {
			var o = r.getTextContent();
			o && et(o, { style: { opacity: 0 } }, t, {
				dataIndex: i,
				removeOpt: a,
				cb: function() {
					r.removeTextContent();
				}
			});
		} else r.removeTextContent();
		et(r, {
			style: { opacity: 0 },
			scaleX: 0,
			scaleY: 0
		}, t, {
			dataIndex: i,
			cb: e,
			removeOpt: a
		});
	}, t.getSymbolSize = function(e, t) {
		return Le(e.getItemVisual(t, "symbolSize"));
	}, t.getSymbolZ2 = function(e, t) {
		return e.getItemVisual(t, "z2");
	}, t;
}(y);
function Pt(e, t) {
	this.parent.drift(e, t);
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/helper/SymbolDraw.js
function Ft(e, t, n, r) {
	return t && !isNaN(t[0]) && !isNaN(t[1]) && !(r && r.isIgnore && r.isIgnore(n)) && !(r && r.clipShape && !r.clipShape.contain(t[0], t[1])) && e.getItemVisual(n, "symbol") !== "none";
}
function It(e) {
	return e != null && !De(e) && (e = { isIgnore: e }), e || {};
}
function Lt(e) {
	var t = e.hostModel, n = t.getModel("emphasis");
	return {
		emphasisItemStyle: n.getModel("itemStyle").getItemStyle(),
		blurItemStyle: t.getModel(["blur", "itemStyle"]).getItemStyle(),
		selectItemStyle: t.getModel(["select", "itemStyle"]).getItemStyle(),
		focus: n.get("focus"),
		blurScope: n.get("blurScope"),
		emphasisDisabled: n.get("disabled"),
		hoverScale: n.get("scale"),
		labelStatesModels: Ze(t),
		cursorStyle: t.get("cursor")
	};
}
function Rt(e, t, n, r, i, a, o) {
	var s = new e(t, n, r, i);
	return s.setPosition(a), t.setItemGraphicEl(n, s), o.add(s), s;
}
var zt = function() {
	function e(e) {
		this.group = new y(), this._SymbolCtor = e || Nt;
	}
	return e.prototype.updateData = function(e, t) {
		this._progressiveEls = null, t = It(t);
		var n = this.group, r = e.hostModel, i = this._data, a = this._SymbolCtor, o = t.disableAnimation, s = this._seriesScope = Lt(e), c = { disableAnimation: o }, l = t.getSymbolPoint || function(t) {
			return e.getItemLayout(t);
		};
		i || n.removeAll(), e.diff(i).add(function(r) {
			var i = l(r);
			Ft(e, i, r, t) && Rt(a, e, r, s, c, i, n);
		}).update(function(u, d) {
			var f = i.getItemGraphicEl(d), p = l(u);
			if (!Ft(e, p, u, t)) {
				n.remove(f);
				return;
			}
			var m = e.getItemVisual(u, "symbol") || "circle", h = f && f.getSymbolType && f.getSymbolType();
			if (!f || h && h !== m) n.remove(f), f = new a(e, u, s, c), f.setPosition(p);
			else {
				f.updateData(e, u, s, c);
				var g = {
					x: p[0],
					y: p[1]
				};
				o ? f.attr(g) : qe(f, g, r);
			}
			n.add(f), e.setItemGraphicEl(u, f);
		}).remove(function(e) {
			var t = i.getItemGraphicEl(e);
			t && t.fadeOut(function() {
				n.remove(t);
			}, r);
		}).execute(), this._getSymbolPoint = l, this._data = e;
	}, e.prototype.updateLayout = function(e) {
		var t = this._data;
		if (t) for (var n = this, r = t.getStore(), i = 0, a = r.count(); i < a; i++) {
			var o = t.getItemGraphicEl(i), s = n._getSymbolPoint(i);
			Ft(t, s, i, e) ? (o ||= Rt(n._SymbolCtor, t, i, n._seriesScope, { disableAnimation: !0 }, s, n.group), o.stopAnimation(), o.setPosition(s), o.markRedraw()) : o && (n.group.remove(o), t.setItemGraphicEl(i, null));
		}
	}, e.prototype.incrementalPrepareUpdate = function(e) {
		this._seriesScope = Lt(e), this._data = null, this.group.removeAll();
	}, e.prototype.incrementalUpdate = function(e, t, n, r) {
		this._progressiveEls = [], r = It(r);
		function i(e) {
			e.isGroup || (e.incremental = n, e.ensureState("emphasis").hoverLayer = 2);
		}
		for (var a = e.start; a < e.end; a++) {
			var o = t.getItemLayout(a);
			if (Ft(t, o, a, r)) {
				var s = new this._SymbolCtor(t, a, this._seriesScope);
				s.traverse(i), s.setPosition(o), this.group.add(s), t.setItemGraphicEl(a, s), this._progressiveEls.push(s);
			}
		}
	}, e.prototype.eachRendered = function(e) {
		Te(this._progressiveEls || this.group, e);
	}, e.prototype.remove = function(e) {
		var t = this.group, n = this._data;
		n && e ? n.eachItemGraphicEl(function(e) {
			e.fadeOut(function() {
				t.remove(e);
			}, n.hostModel);
		}) : t.removeAll();
	}, e;
}();
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/line/helper.js
function Bt(e, t, n) {
	var r = e.getBaseAxis(), i = e.getOtherAxis(r), a = Vt(i, n), s = r.dim, c = i.dim, l = t.mapDimension(c), u = t.mapDimension(s), d = c === "x" || c === "radius" ? 1 : 0, f = Ne(e.dimensions, function(e) {
		return t.mapDimension(e);
	}), p = !1, m = t.getCalculationInfo("stackResultDimension");
	return o(t, f[0]) && (p = !0, f[0] = m), o(t, f[1]) && (p = !0, f[1] = m), {
		dataDimsForPoint: f,
		valueStart: a,
		valueAxisDim: c,
		baseAxisDim: s,
		stacked: !!p,
		valueDim: l,
		baseDim: u,
		baseDataOffset: d,
		stackedOverDimension: t.getCalculationInfo("stackedOverDimension")
	};
}
function Vt(e, t) {
	var n = 0, r = e.scale.getExtent();
	return t === "start" ? n = r[0] : t === "end" ? n = r[1] : Se(t) && !isNaN(t) ? n = t : r[0] > 0 ? n = r[0] : r[1] < 0 && (n = r[1]), n;
}
function Ht(e, t, n, r) {
	var i = NaN;
	e.stacked && (i = n.get(n.getCalculationInfo("stackedOverDimension"), r)), isNaN(i) && (i = e.valueStart);
	var a = e.baseDataOffset, o = [];
	return o[a] = n.get(e.baseDim, r), o[1 - a] = i, t.dataToPoint(o);
}
function q(e, t) {
	return !isFinite(e) || !isFinite(t);
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/line/lineAnimationDiff.js
function Ut(e, t) {
	var n = [];
	return t.diff(e).add(function(e) {
		n.push({
			cmd: "+",
			idx: e
		});
	}).update(function(e, t) {
		n.push({
			cmd: "=",
			idx: t,
			idx1: e
		});
	}).remove(function(e) {
		n.push({
			cmd: "-",
			idx: e
		});
	}).execute(), n;
}
function Wt(e, t, n, r, i, a, o, s) {
	for (var c = Ut(e, t), l = [], u = [], d = [], f = [], p = [], m = [], h = [], g = Bt(i, t, o), _ = e.getLayout("points") || [], v = t.getLayout("points") || [], y = 0; y < c.length; y++) {
		var b = c[y], x = !0, S = void 0, C = void 0;
		switch (b.cmd) {
			case "=":
				S = b.idx * 2, C = b.idx1 * 2;
				var w = _[S], T = _[S + 1], E = v[C], D = v[C + 1];
				(isNaN(w) || isNaN(T)) && (w = E, T = D), l.push(w, T), u.push(E, D), d.push(n[S], n[S + 1]), f.push(r[C], r[C + 1]), h.push(t.getRawIndex(b.idx1));
				break;
			case "+":
				var O = b.idx, k = g.dataDimsForPoint, A = i.dataToPoint([t.get(k[0], O), t.get(k[1], O)]);
				C = O * 2, l.push(A[0], A[1]), u.push(v[C], v[C + 1]);
				var j = Ht(g, i, t, O);
				d.push(j[0], j[1]), f.push(r[C], r[C + 1]), h.push(t.getRawIndex(O));
				break;
			case "-": x = !1;
		}
		x && (p.push(b), m.push(m.length));
	}
	m.sort(function(e, t) {
		return h[e] - h[t];
	});
	for (var M = l.length, N = ke(M), P = ke(M), F = ke(M), I = ke(M), ee = [], y = 0; y < m.length; y++) {
		var te = m[y], L = y * 2, R = te * 2;
		N[L] = l[R], N[L + 1] = l[R + 1], P[L] = u[R], P[L + 1] = u[R + 1], F[L] = d[R], F[L + 1] = d[R + 1], I[L] = f[R], I[L + 1] = f[R + 1], ee[y] = p[te];
	}
	return {
		current: N,
		next: P,
		stackedOnCurrent: F,
		stackedOnNext: I,
		status: ee
	};
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/line/poly.js
var Gt = Math.min, Kt = Math.max;
function qt(e, t, n, r, i, a, o, s, c) {
	for (var l, u, d, f, p, m, h = n, g = 0; g < r; g++) {
		var _ = t[h * 2], v = t[h * 2 + 1];
		if (h >= i || h < 0) break;
		if (q(_, v)) {
			if (c) {
				h += a;
				continue;
			}
			break;
		}
		if (h === n) e[a > 0 ? "moveTo" : "lineTo"](_, v), d = _, f = v;
		else {
			var y = _ - l, b = v - u;
			if (y * y + b * b < .5) {
				h += a;
				continue;
			}
			if (o > 0) {
				for (var x = h + a, S = t[x * 2], C = t[x * 2 + 1]; S === _ && C === v && g < r;) g++, x += a, h += a, S = t[x * 2], C = t[x * 2 + 1], _ = t[h * 2], v = t[h * 2 + 1], y = _ - l, b = v - u;
				var w = g + 1;
				if (c) for (; q(S, C) && w < r;) w++, x += a, S = t[x * 2], C = t[x * 2 + 1];
				var T = .5, E = 0, D = 0, O = void 0, k = void 0;
				if (w >= r || q(S, C)) p = _, m = v;
				else {
					E = S - l, D = C - u;
					var A = _ - l, j = S - _, M = v - u, N = C - v, P = void 0, F = void 0;
					if (s === "x") {
						P = Math.abs(A), F = Math.abs(j);
						var I = E > 0 ? 1 : -1;
						p = _ - I * P * o, m = v, O = _ + I * F * o, k = v;
					} else if (s === "y") {
						P = Math.abs(M), F = Math.abs(N);
						var ee = D > 0 ? 1 : -1;
						p = _, m = v - ee * P * o, O = _, k = v + ee * F * o;
					} else P = Math.sqrt(A * A + M * M), F = Math.sqrt(j * j + N * N), T = F / (F + P), p = _ - E * o * (1 - T), m = v - D * o * (1 - T), O = _ + E * o * T, k = v + D * o * T, O = Gt(O, Kt(S, _)), k = Gt(k, Kt(C, v)), O = Kt(O, Gt(S, _)), k = Kt(k, Gt(C, v)), E = O - _, D = k - v, p = _ - E * P / F, m = v - D * P / F, p = Gt(p, Kt(l, _)), m = Gt(m, Kt(u, v)), p = Kt(p, Gt(l, _)), m = Kt(m, Gt(u, v)), E = _ - p, D = v - m, O = _ + E * F / P, k = v + D * F / P;
				}
				e.bezierCurveTo(d, f, p, m, _, v), d = O, f = k;
			} else e.lineTo(_, v);
		}
		l = _, u = v, h += a;
	}
	return g;
}
var Jt = function() {
	function e() {
		this.smooth = 0, this.smoothConstraint = !0;
	}
	return e;
}(), Yt = function(e) {
	z(t, e);
	function t(t) {
		var n = e.call(this, t) || this;
		return n.type = "ec-polyline", n;
	}
	return t.prototype.getDefaultStyle = function() {
		return {
			stroke: v.color.neutral99,
			fill: null
		};
	}, t.prototype.getDefaultShape = function() {
		return new Jt();
	}, t.prototype.buildPath = function(e, t) {
		var n = t.points, r = 0, i = n.length / 2;
		if (t.connectNulls) {
			for (; i > 0 && q(n[i * 2 - 2], n[i * 2 - 1]); i--);
			for (; r < i && q(n[r * 2], n[r * 2 + 1]); r++);
		}
		for (; r < i;) r += qt(e, n, r, i, i, 1, t.smooth, t.smoothMonotone, t.connectNulls) + 1;
	}, t.prototype.getPointOn = function(e, t) {
		this.path || (this.createPathProxy(), this.buildPath(this.path, this.shape));
		for (var n = this.path.data, r = ee.CMD, i, a, o = t === "x", s = [], c = 0; c < n.length;) {
			var l = n[c++], u = void 0, d = void 0, f = void 0, p = void 0, m = void 0, g = void 0, _ = void 0;
			switch (l) {
				case r.M:
					i = n[c++], a = n[c++];
					break;
				case r.L:
					if (u = n[c++], d = n[c++], _ = o ? (e - i) / (u - i) : (e - a) / (d - a), _ <= 1 && _ >= 0) {
						var v = o ? (d - a) * _ + a : (u - i) * _ + i;
						return o ? [e, v] : [v, e];
					}
					i = u, a = d;
					break;
				case r.C:
					u = n[c++], d = n[c++], f = n[c++], p = n[c++], m = n[c++], g = n[c++];
					var y = o ? C(i, u, f, m, e, s) : C(a, d, p, g, e, s);
					if (y > 0) for (var b = 0; b < y; b++) {
						var x = s[b];
						if (x <= 1 && x >= 0) {
							var v = o ? h(a, d, p, g, x) : h(i, u, f, m, x);
							return o ? [e, v] : [v, e];
						}
					}
					i = m, a = g;
					break;
			}
		}
	}, t;
}(gt), Xt = function(e) {
	z(t, e);
	function t() {
		return e !== null && e.apply(this, arguments) || this;
	}
	return t;
}(Jt), Zt = function(e) {
	z(t, e);
	function t(t) {
		var n = e.call(this, t) || this;
		return n.type = "ec-polygon", n;
	}
	return t.prototype.getDefaultShape = function() {
		return new Xt();
	}, t.prototype.buildPath = function(e, t) {
		var n = t.points, r = t.stackedOnPoints, i = 0, a = n.length / 2, o = t.smoothMonotone;
		if (t.connectNulls) {
			for (; a > 0 && q(n[a * 2 - 2], n[a * 2 - 1]); a--);
			for (; i < a && q(n[i * 2], n[i * 2 + 1]); i++);
		}
		for (; i < a;) {
			var s = qt(e, n, i, a, a, 1, t.smooth, o, t.connectNulls);
			qt(e, r, i + s - 1, s, a, -1, t.stackedOnSmooth, o, t.connectNulls), i += s + 1, e.closePath();
		}
	}, t;
}(gt);
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/util/styleCompat.js
function Qt(e, t, n, r) {
	return e && (e.legacy || e.legacy !== !1 && !n && !r && t !== "tspan" && (t === "text" || B(e, "text")));
}
function $t(e, t, n) {
	var r = e, i, a, o;
	if (t === "text") o = r;
	else {
		o = {}, B(r, "text") && (o.text = r.text), B(r, "rich") && (o.rich = r.rich), B(r, "textFill") && (o.fill = r.textFill), B(r, "textStroke") && (o.stroke = r.textStroke), B(r, "fontFamily") && (o.fontFamily = r.fontFamily), B(r, "fontSize") && (o.fontSize = r.fontSize), B(r, "fontStyle") && (o.fontStyle = r.fontStyle), B(r, "fontWeight") && (o.fontWeight = r.fontWeight), a = {
			type: "text",
			style: o,
			silent: !0
		}, i = {};
		var s = B(r, "textPosition");
		n ? i.position = s ? r.textPosition : "inside" : s && (i.position = r.textPosition), B(r, "textPosition") && (i.position = r.textPosition), B(r, "textOffset") && (i.offset = r.textOffset), B(r, "textRotation") && (i.rotation = r.textRotation), B(r, "textDistance") && (i.distance = r.textDistance);
	}
	return en(o, e), V(o.rich, function(e) {
		en(e, e);
	}), {
		textConfig: i,
		textContent: a
	};
}
function en(e, t) {
	t && (t.font = t.textFont || t.font, B(t, "textStrokeWidth") && (e.lineWidth = t.textStrokeWidth), B(t, "textAlign") && (e.align = t.textAlign), B(t, "textVerticalAlign") && (e.verticalAlign = t.textVerticalAlign), B(t, "textLineHeight") && (e.lineHeight = t.textLineHeight), B(t, "textWidth") && (e.width = t.textWidth), B(t, "textHeight") && (e.height = t.textHeight), B(t, "textBackgroundColor") && (e.backgroundColor = t.textBackgroundColor), B(t, "textPadding") && (e.padding = t.textPadding), B(t, "textBorderColor") && (e.borderColor = t.textBorderColor), B(t, "textBorderWidth") && (e.borderWidth = t.textBorderWidth), B(t, "textBorderRadius") && (e.borderRadius = t.textBorderRadius), B(t, "textBoxShadowColor") && (e.shadowColor = t.textBoxShadowColor), B(t, "textBoxShadowBlur") && (e.shadowBlur = t.textBoxShadowBlur), B(t, "textBoxShadowOffsetX") && (e.shadowOffsetX = t.textBoxShadowOffsetX), B(t, "textBoxShadowOffsetY") && (e.shadowOffsetY = t.textBoxShadowOffsetY));
}
function tn(e, t, n) {
	var r = e;
	r.textPosition = r.textPosition || n.position || "inside", n.offset != null && (r.textOffset = n.offset), n.rotation != null && (r.textRotation = n.rotation), n.distance != null && (r.textDistance = n.distance);
	var i = r.textPosition.indexOf("inside") >= 0, a = e.fill || v.color.neutral99;
	nn(r, t);
	var o = r.textFill == null;
	return i ? o && (r.textFill = n.insideFill || v.color.neutral00, !r.textStroke && n.insideStroke && (r.textStroke = n.insideStroke), !r.textStroke && (r.textStroke = a), r.textStrokeWidth ??= 2) : (o && (r.textFill = e.fill || n.outsideFill || v.color.neutral00), !r.textStroke && n.outsideStroke && (r.textStroke = n.outsideStroke)), r.text = t.text, r.rich = t.rich, V(t.rich, function(e) {
		nn(e, e);
	}), r;
}
function nn(e, t) {
	t && (B(t, "fill") && (e.textFill = t.fill), B(t, "stroke") && (e.textStroke = t.fill), B(t, "lineWidth") && (e.textStrokeWidth = t.lineWidth), B(t, "font") && (e.font = t.font), B(t, "fontStyle") && (e.fontStyle = t.fontStyle), B(t, "fontWeight") && (e.fontWeight = t.fontWeight), B(t, "fontSize") && (e.fontSize = t.fontSize), B(t, "fontFamily") && (e.fontFamily = t.fontFamily), B(t, "align") && (e.textAlign = t.align), B(t, "verticalAlign") && (e.textVerticalAlign = t.verticalAlign), B(t, "lineHeight") && (e.textLineHeight = t.lineHeight), B(t, "width") && (e.textWidth = t.width), B(t, "height") && (e.textHeight = t.height), B(t, "backgroundColor") && (e.textBackgroundColor = t.backgroundColor), B(t, "padding") && (e.textPadding = t.padding), B(t, "borderColor") && (e.textBorderColor = t.borderColor), B(t, "borderWidth") && (e.textBorderWidth = t.borderWidth), B(t, "borderRadius") && (e.textBorderRadius = t.borderRadius), B(t, "shadowColor") && (e.textBoxShadowColor = t.shadowColor), B(t, "shadowBlur") && (e.textBoxShadowBlur = t.shadowBlur), B(t, "shadowOffsetX") && (e.textBoxShadowOffsetX = t.shadowOffsetX), B(t, "shadowOffsetY") && (e.textBoxShadowOffsetY = t.shadowOffsetY), B(t, "textShadowColor") && (e.textShadowColor = t.textShadowColor), B(t, "textShadowBlur") && (e.textShadowBlur = t.textShadowBlur), B(t, "textShadowOffsetX") && (e.textShadowOffsetX = t.textShadowOffsetX), B(t, "textShadowOffsetY") && (e.textShadowOffsetY = t.textShadowOffsetY));
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/line/LineView.js
function rn(e, t) {
	if (e.length === t.length) {
		for (var n = 0; n < e.length; n++) if (e[n] !== t[n]) return;
		return !0;
	}
}
function an(e) {
	for (var t = d(), n = d(), r = 0; r < e.length;) {
		var i = e[r++], a = e[r++];
		q(i, a) || (O(t, i), O(n, a));
	}
	return [t, n];
}
function on(e, t) {
	var n = an(e), r = n[0], i = n[1], a = an(t), o = a[0], s = a[1];
	return Math.max(Math.abs(r[0] - o[0]), Math.abs(i[0] - s[0]), Math.abs(r[1] - o[1]), Math.abs(i[1] - s[1]));
}
function sn(e) {
	return Se(e) ? e : e ? .5 : 0;
}
function cn(e, t, n) {
	if (n.valueDim == null) return [];
	for (var r = t.count(), i = ke(r * 2), a = 0; a < r; a++) {
		var o = Ht(n, e, t, a);
		i[a * 2] = o[0], i[a * 2 + 1] = o[1];
	}
	return i;
}
function ln(e, t, n, r, i) {
	var a = n.getBaseAxis(), o = a.dim === "x" || a.dim === "radius" ? 0 : 1, s = [], c = 0, l = [], u = [], d = [], f = [];
	if (i) {
		for (c = 0; c < e.length; c += 2) {
			var p = t || e;
			q(p[c], p[c + 1]) || f.push(e[c], e[c + 1]);
		}
		e = f;
	}
	for (c = 0; c < e.length - 2; c += 2) switch (d[0] = e[c + 2], d[1] = e[c + 3], u[0] = e[c], u[1] = e[c + 1], s.push(u[0], u[1]), r) {
		case "end":
			l[o] = d[o], l[1 - o] = u[1 - o], s.push(l[0], l[1]);
			break;
		case "middle":
			var m = (u[o] + d[o]) / 2, h = [];
			l[o] = h[o] = m, l[1 - o] = u[1 - o], h[1 - o] = d[1 - o], s.push(l[0], l[1]), s.push(h[0], h[1]);
			break;
		default: l[o] = u[o], l[1 - o] = d[1 - o], s.push(l[0], l[1]);
	}
	return s.push(e[c++], e[c++]), s;
}
function un(e, t) {
	var n = [], r = e.length, i, a;
	function o(e, t, n) {
		var r = e.coord;
		return {
			coord: n,
			color: ae((n - r) / (t.coord - r), [e.color, t.color])
		};
	}
	for (var s = 0; s < r; s++) {
		var c = e[s], l = c.coord;
		if (l < 0) i = c;
		else if (l > t) {
			a ? n.push(o(a, c, t)) : i && n.push(o(i, c, 0), o(i, c, t));
			break;
		} else i &&= (n.push(o(i, c, 0)), null), n.push(c), a = c;
	}
	return n;
}
function dn(e, t, n) {
	var r = e.getVisual("visualMeta");
	if (!(!r || !r.length || !e.count()) && t.type === "cartesian2d") {
		for (var i, a, o = r.length - 1; o >= 0; o--) {
			var s = e.getDimensionInfo(r[o].dimension);
			if (i = s && s.coordDim, i === "x" || i === "y") {
				a = r[o];
				break;
			}
		}
		if (a) {
			var c = t.getAxis(i), l = Ne(a.stops, function(e) {
				return {
					coord: c.toGlobalCoord(c.dataToCoord(e.value)),
					color: e.color
				};
			}), u = l.length, d = a.outerColors.slice();
			u && l[0].coord > l[u - 1].coord && (l.reverse(), d.reverse());
			var f = un(l, i === "x" ? n.getWidth() : n.getHeight()), p = f.length;
			if (!p && u) return l[0].coord < 0 ? d[1] ? d[1] : l[u - 1].color : d[0] ? d[0] : l[0].color;
			var m = 10, h = f[0].coord - m, g = f[p - 1].coord + m, _ = g - h;
			if (_ < .001) return "transparent";
			V(f, function(e) {
				e.offset = (e.coord - h) / _;
			}), f.push({
				offset: p ? f[p - 1].offset : .5,
				color: d[1] || "transparent"
			}), f.unshift({
				offset: p ? f[0].offset : .5,
				color: d[0] || "transparent"
			});
			var v = new rt(0, 0, 0, 0, f, !0);
			return v[i] = h, v[i + "2"] = g, v;
		}
	}
}
function fn(e, t, n) {
	var r = e.get("showAllSymbol"), i = r === "auto";
	if (!(r && !i)) {
		var a = n.getAxesByScale("ordinal")[0];
		if (a && !(i && pn(a, t))) {
			var o = t.mapDimension(a.dim), s = {};
			return V(a.getViewLabels(), function(e) {
				e.tick.offInterval || (s[mt(a.scale, e.tick)] = 1);
			}), function(e) {
				return !s.hasOwnProperty(t.get(o, e));
			};
		}
	}
}
function pn(e, t) {
	var n = e.getExtent(), r = Math.abs(n[1] - n[0]) / e.scale.count();
	isNaN(r) && (r = 0);
	for (var i = t.count(), a = Math.max(1, Math.round(i / 5)), o = 0; o < i; o += a) if (Nt.getSymbolSize(t, o)[e.isHorizontal() ? 1 : 0] * 1.5 > r) return !1;
	return !0;
}
function mn(e) {
	for (var t = e.length / 2; t > 0 && q(e[t * 2 - 2], e[t * 2 - 1]); t--);
	return t - 1;
}
function hn(e, t) {
	return [e[t * 2], e[t * 2 + 1]];
}
function gn(e, t, n) {
	for (var r = e.length / 2, i = n === "x" ? 0 : 1, a, o, s = 0, c = -1, l = 0; l < r; l++) if (o = e[l * 2 + i], !q(o, e[l * 2 + 1 - i])) {
		if (l === 0) {
			a = o;
			continue;
		}
		if (a <= t && o >= t || a >= t && o <= t) {
			c = l;
			break;
		}
		s = l, a = o;
	}
	return {
		range: [s, c],
		t: (t - a) / (o - a)
	};
}
function _n(e) {
	if (e.get(["endLabel", "show"])) return !0;
	for (var t = 0; t < ct.length; t++) if (e.get([
		ct[t],
		"endLabel",
		"show"
	])) return !0;
	return !1;
}
function vn(e, t, n, r) {
	if (Ee(t, "cartesian2d")) {
		var i = r.getModel("endLabel"), a = i.get("valueAnimation"), o = r.getData(), s = { lastFrameIndex: 0 }, c = _n(r) ? function(n, r) {
			e._endLabelOnDuring(n, r, o, s, a, i, t);
		} : null, l = t.getBaseAxis().isHorizontal(), u = Be(t, n, r, function() {
			var t = e._endLabel;
			t && n && s.originalX != null && t.attr({
				x: s.originalX,
				y: s.originalY
			});
		}, c);
		if (!r.get("clip", !0)) {
			var d = u.shape, f = Math.max(d.width, d.height);
			l ? (d.y -= f, d.height += f * 2) : (d.x -= f, d.width += f * 2);
		}
		return c && c(1, u), u;
	} else return Me(t, n, r);
}
function yn(e, t) {
	var n = t.getBaseAxis(), r = n.isHorizontal(), i = n.inverse, a = r ? i ? "right" : "left" : "center", o = r ? "middle" : i ? "top" : "bottom";
	return { normal: {
		align: e.get("align") || a,
		verticalAlign: e.get("verticalAlign") || o
	} };
}
var bn = function(e) {
	z(t, e);
	function t() {
		return e !== null && e.apply(this, arguments) || this;
	}
	return t.prototype.init = function() {
		var e = new y(), t = new zt();
		this.group.add(t.group), this._symbolDraw = t, this._lineGroup = e, this._changePolyState = we(this._changePolyState, this);
	}, t.prototype.render = function(e, t, n) {
		var r = e.coordinateSystem, i = this.group, a = e.getData(), o = e.getModel("lineStyle"), s = e.getModel("areaStyle"), c = a.getLayout("points") || [], l = r.type === "polar", u = this._coordSys, d = this._symbolDraw, f = this._polyline, p = this._polygon, m = this._lineGroup, h = !t.ssr && e.get("animation"), _ = !s.isEmpty(), v = s.get("origin"), y = Bt(r, a, v), b = _ && cn(r, a, y), x = e.get("showSymbol"), S = e.get("connectNulls"), C = x && !l && fn(e, a, r), w = this._data;
		w && w.eachItemGraphicEl(function(e, t) {
			e.__temp && (i.remove(e), w.setItemGraphicEl(t, null));
		}), x || d.remove(), i.add(m);
		var T = l ? !1 : e.get("step"), E;
		r && r.getArea && e.get("clip", !0) && (E = r.getArea(), E.width == null ? E.r0 && (E.r0 -= .5, E.r += .5) : (E.x -= .1, E.y -= .1, E.width += .2, E.height += .2)), this._clipShapeForSymbol = E;
		var D = dn(a, r, n) || a.getVisual("style")[a.getVisual("drawType")];
		if (!(f && u.type === r.type && T === this._step)) x && d.updateData(a, {
			isIgnore: C,
			clipShape: E,
			disableAnimation: !0,
			getSymbolPoint: function(e) {
				return [c[e * 2], c[e * 2 + 1]];
			}
		}), h && this._initSymbolLabelAnimation(a, r, E), T && (b &&= ln(b, c, r, T, S), c = ln(c, null, r, T, S)), f = this._newPolyline(c), _ ? p = this._newPolygon(c, b) : p &&= (m.remove(p), this._polygon = null), l || this._initOrUpdateEndLabel(e, r, g(D)), m.setClipPath(vn(this, r, !0, e));
		else {
			_ && !p ? p = this._newPolygon(c, b) : p && !_ && (m.remove(p), p = this._polygon = null), l || this._initOrUpdateEndLabel(e, r, g(D));
			var O = m.getClipPath();
			O ? Ge(O, { shape: vn(this, r, !1, e).shape }, e) : m.setClipPath(vn(this, r, !0, e)), x && d.updateData(a, {
				isIgnore: C,
				clipShape: E,
				disableAnimation: !0,
				getSymbolPoint: function(e) {
					return [c[e * 2], c[e * 2 + 1]];
				}
			}), (!rn(this._stackedOnPoints, b) || !rn(this._points, c)) && (h ? this._doUpdateAnimation(a, b, r, n, T, v, S) : (T && (b &&= ln(b, c, r, T, S), c = ln(c, null, r, T, S)), f.setShape({ points: c }), p && p.setShape({
				points: c,
				stackedOnPoints: b
			})));
		}
		var k = e.getModel("emphasis"), A = k.get("focus"), j = k.get("blurScope"), M = k.get("disabled");
		if (f.useStyle(We(o.getLineStyle(), {
			fill: "none",
			stroke: D,
			lineJoin: "bevel"
		})), Pe(f, e, "lineStyle"), f.style.lineWidth > 0 && e.get([
			"emphasis",
			"lineStyle",
			"width"
		]) === "bolder") {
			var N = f.getState("emphasis").style;
			N.lineWidth = +f.style.lineWidth + 1;
		}
		pe(f).seriesIndex = e.seriesIndex, je(f, A, j, M);
		var P = sn(e.get("smooth")), F = e.get("smoothMonotone");
		if (f.setShape({
			smooth: P,
			smoothMonotone: F,
			connectNulls: S
		}), p) {
			var I = a.getCalculationInfo("stackedOnSeries"), ee = 0;
			p.useStyle(We(s.getAreaStyle(), {
				fill: D,
				opacity: .7,
				lineJoin: "bevel",
				decal: a.getVisual("style").decal
			})), I && (ee = sn(I.get("smooth"))), p.setShape({
				smooth: P,
				stackedOnSmooth: ee,
				smoothMonotone: F,
				connectNulls: S
			}), Pe(p, e, "areaStyle"), pe(p).seriesIndex = e.seriesIndex, je(p, A, j, M);
		}
		var te = this._changePolyState;
		a.eachItemGraphicEl(function(e) {
			e && (e.onHoverStateChange = te);
		}), this._polyline.onHoverStateChange = te, this._data = a, this._coordSys = r, this._stackedOnPoints = b, this._points = c, this._step = T, this._valueOrigin = v;
		var L = e.get("triggerEvent"), R = e.get("triggerLineEvent"), ne = R === !0 || L === !0 || L === "line", re = R === !0 || L === !0 || L === "area";
		this.packEventData(e, f, ne), p && this.packEventData(e, p, re);
	}, t.prototype.packEventData = function(e, t, n) {
		pe(t).eventData = n ? {
			componentType: "series",
			componentSubType: "line",
			componentIndex: e.componentIndex,
			seriesIndex: e.seriesIndex,
			seriesName: e.name,
			seriesType: "line",
			selfType: t === this._polygon ? "area" : "line"
		} : null;
	}, t.prototype.highlight = function(e, t, n, r) {
		var i = e.getData(), a = E(i, r);
		if (this._changePolyState("emphasis"), !(a instanceof Array) && a != null && a >= 0) {
			var o = i.getLayout("points"), s = i.getItemGraphicEl(a);
			if (!s) {
				var c = o[a * 2], l = o[a * 2 + 1];
				if (q(c, l) || this._clipShapeForSymbol && !this._clipShapeForSymbol.contain(c, l)) return;
				var u = e.get("zlevel") || 0, d = e.get("z") || 0;
				s = new Nt(i, a), s.x = c, s.y = l, s.setZ(u, d);
				var f = s.getSymbolPath().getTextContent();
				f && (f.zlevel = u, f.z = d, f.z2 = this._polyline.z2 + 1), s.__temp = !0, i.setItemGraphicEl(a, s), s.stopSymbolAnimation(!0), this.group.add(s);
			}
			s.highlight();
		} else T.prototype.highlight.call(this, e, t, n, r);
	}, t.prototype.downplay = function(e, t, n, r) {
		var i = e.getData(), a = E(i, r);
		if (this._changePolyState("normal"), a != null && a >= 0) {
			var o = i.getItemGraphicEl(a);
			o && (o.__temp ? (i.setItemGraphicEl(a, null), this.group.remove(o)) : o.downplay());
		} else T.prototype.downplay.call(this, e, t, n, r);
	}, t.prototype._changePolyState = function(e) {
		var t = this._polygon;
		He(this._polyline, e), t && He(t, e);
	}, t.prototype._newPolyline = function(e) {
		var t = this._polyline;
		return t && this._lineGroup.remove(t), t = new Yt({
			shape: { points: e },
			segmentIgnoreThreshold: 2,
			z2: 10
		}), this._lineGroup.add(t), this._polyline = t, t;
	}, t.prototype._newPolygon = function(e, t) {
		var n = this._polygon;
		return n && this._lineGroup.remove(n), n = new Zt({
			shape: {
				points: e,
				stackedOnPoints: t
			},
			segmentIgnoreThreshold: 2
		}), this._lineGroup.add(n), this._polygon = n, n;
	}, t.prototype._initSymbolLabelAnimation = function(e, t, n) {
		var r, i, a = t.getBaseAxis(), o = a.inverse;
		t.type === "cartesian2d" ? (r = a.isHorizontal(), i = !1) : t.type === "polar" && (r = a.dim === "angle", i = !0);
		var s = e.hostModel, c = s.get("animationDuration");
		st(c) && (c = c(null));
		var l = s.get("animationDelay") || 0, u = st(l) ? l(null) : l;
		e.eachItemGraphicEl(function(e, a) {
			var s = e;
			if (s) {
				var d = [e.x, e.y], f = void 0, p = void 0, m = void 0;
				if (n) if (i) {
					var h = n, g = t.pointToCoord(d);
					r ? (f = h.startAngle, p = h.endAngle, m = -g[1] / 180 * Math.PI) : (f = h.r0, p = h.r, m = g[0]);
				} else {
					var _ = n;
					r ? (f = _.x, p = _.x + _.width, m = e.x) : (f = _.y + _.height, p = _.y, m = e.y);
				}
				var v = p === f ? 0 : (m - f) / (p - f);
				o && (v = 1 - v);
				var y = st(l) ? l(a) : c * v + u, b = s.getSymbolPath(), x = b.getTextContent();
				s.attr({
					scaleX: 0,
					scaleY: 0
				}), s.animateTo({
					scaleX: 1,
					scaleY: 1
				}, {
					duration: 200,
					setToFinal: !0,
					delay: y
				}), x && x.animateFrom({ style: { opacity: 0 } }, {
					duration: 300,
					delay: y
				}), b.disableLabelAnimation = !0;
			}
		});
	}, t.prototype._initOrUpdateEndLabel = function(e, t, n) {
		var r = e.getModel("endLabel");
		if (_n(e)) {
			var i = e.getData(), a = this._polyline, o = i.getLayout("points");
			if (!o) {
				a.removeTextContent(), this._endLabel = null;
				return;
			}
			var s = this._endLabel;
			s || (s = this._endLabel = new dt({ z2: 200 }), s.ignoreClip = !0, a.setTextContent(this._endLabel), a.disableLabelAnimation = !0);
			var c = mn(o);
			c >= 0 && (se(a, Ze(e, "endLabel"), {
				inheritColor: n,
				labelFetcher: e,
				labelDataIndex: c,
				defaultText: function(e, t, n) {
					return n == null ? lt(i, e) : de(i, n);
				},
				enableTextSetter: !0
			}, yn(r, t)), a.textConfig.position = null);
		} else this._endLabel &&= (this._polyline.removeTextContent(), null);
	}, t.prototype._endLabelOnDuring = function(e, t, n, r, i, a, o) {
		var s = this._endLabel, c = this._polyline;
		if (s) {
			e < 1 && r.originalX == null && (r.originalX = s.x, r.originalY = s.y);
			var u = n.getLayout("points"), d = n.hostModel, f = d.get("connectNulls"), p = a.get("precision"), m = a.get("distance") || 0, h = o.getBaseAxis(), g = h.isHorizontal(), _ = h.inverse, v = t.shape, y = _ ? g ? v.x : v.y + v.height : g ? v.x + v.width : v.y, x = (g ? m : 0) * (_ ? -1 : 1), S = (g ? 0 : -m) * (_ ? -1 : 1), C = g ? "x" : "y", w = gn(u, y, C), T = w.range, E = T[1] - T[0], D = void 0;
			if (E >= 1) {
				if (E > 1 && !f) {
					var O = hn(u, T[0]);
					s.attr({
						x: O[0] + x,
						y: O[1] + S
					}), i && (D = d.getRawValue(T[0]));
				} else {
					var O = c.getPointOn(y, C);
					O && s.attr({
						x: O[0] + x,
						y: O[1] + S
					});
					var k = d.getRawValue(T[0]), A = d.getRawValue(T[1]);
					i && (D = l(n, p, k, A, w.t));
				}
				r.lastFrameIndex = T[0];
			} else {
				var j = e === 1 || r.lastFrameIndex > 0 ? T[0] : 0, O = hn(u, j);
				i && (D = d.getRawValue(j)), s.attr({
					x: O[0] + x,
					y: O[1] + S
				});
			}
			if (i) {
				var M = b(s);
				typeof M.setLabelText == "function" && M.setLabelText(D);
			}
		}
	}, t.prototype._doUpdateAnimation = function(e, t, n, r, i, a, o) {
		var s = this._polyline, c = this._polygon, l = e.hostModel, u = Wt(this._data, e, this._stackedOnPoints, t, this._coordSys, n, this._valueOrigin, a), d = u.current, f = u.stackedOnCurrent, p = u.next, m = u.stackedOnNext;
		if (i && (f = ln(u.stackedOnCurrent, u.current, n, i, o), d = ln(u.current, null, n, i, o), m = ln(u.stackedOnNext, u.next, n, i, o), p = ln(u.next, null, n, i, o)), on(d, p) > 3e3 || c && on(f, m) > 3e3) {
			s.stopAnimation(), s.setShape({ points: p }), c && (c.stopAnimation(), c.setShape({
				points: p,
				stackedOnPoints: m
			}));
			return;
		}
		s.shape.__points = u.current, s.shape.points = d;
		var h = { shape: { points: p } };
		u.current !== d && (h.shape.__points = u.next), s.stopAnimation(), qe(s, h, l), c && (c.setShape({
			points: d,
			stackedOnPoints: f
		}), c.stopAnimation(), qe(c, { shape: { stackedOnPoints: m } }, l), s.shape.points !== c.shape.points && (c.shape.points = s.shape.points));
		for (var g = [], _ = u.status, v = 0; v < _.length; v++) if (_[v].cmd === "=") {
			var y = e.getItemGraphicEl(_[v].idx1);
			y && g.push({
				el: y,
				ptIdx: v
			});
		}
		s.animators && s.animators.length && s.animators[0].during(function() {
			c && c.dirtyShape();
			for (var e = s.shape.__points, t = 0; t < g.length; t++) {
				var n = g[t].el, r = g[t].ptIdx * 2;
				n.x = e[r], n.y = e[r + 1], n.markRedraw();
			}
		});
	}, t.prototype.remove = function(e) {
		var t = this.group, n = this._data;
		this._lineGroup.removeAll(), this._symbolDraw.remove(!0), n && n.eachItemGraphicEl(function(e, r) {
			e.__temp && (t.remove(e), n.setItemGraphicEl(r, null));
		}), this._polyline = this._polygon = this._coordSys = this._points = this._stackedOnPoints = this._endLabel = this._data = null;
	}, t.type = "line", t;
}(T);
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/layout/points.js
function xn(e, t) {
	return {
		seriesType: e,
		plan: D(),
		reset: function(e) {
			var n = e.getData(), r = e.coordinateSystem, i = e.pipelineContext, a = t || i.large;
			if (r) {
				var s = Ne(r.dimensions, function(e) {
					return n.mapDimension(e);
				}).slice(0, 2), c = s.length, l = n.getCalculationInfo("stackResultDimension");
				o(n, s[0]) && (s[0] = l), o(n, s[1]) && (s[1] = l);
				var u = n.getStore(), d = n.getDimensionIndex(s[0]), f = n.getDimensionIndex(s[1]);
				return c && { progress: function(e, t) {
					for (var n = e.end - e.start, i = a && ke(n * c), o = [], s = [], l = e.start, p = 0; l < e.end; l++) {
						var m = void 0;
						if (c === 1) {
							var h = u.get(d, l);
							m = r.dataToPoint(h, null, s);
						} else o[0] = u.get(d, l), o[1] = u.get(f, l), m = r.dataToPoint(o, null, s);
						a ? (i[p++] = m[0], i[p++] = m[1]) : t.setItemLayout(l, m.slice());
					}
					a && (t.setLayout("points", i), t.setLayout("pointsRange", {
						start: e.start,
						end: e.end
					}));
				} };
			}
		}
	};
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/line/install.js
function Sn(e) {
	e.registerChartView(bn), e.registerSeriesModel(Mt), e.registerLayout(xn("line", !0)), e.registerVisual({
		seriesType: "line",
		reset: function(e) {
			var t = e.getData(), n = e.getModel("lineStyle").getLineStyle();
			n && !n.stroke && (n.stroke = t.getVisual("style").fill), t.setVisual("legendLineStyle", n);
		}
	}), e.registerProcessor(e.PRIORITY.PROCESSOR.STATISTIC, xe("line"));
}
function J(e) {
	return e;
}
var Cn = "view";
(function(e) {
	z(n, e);
	function n(t, n, r) {
		var i = e.call(this) || this;
		i.type = Cn, i.dimensions = ["x", "y"];
		var a = J(i);
		a.invertY = t, a.lgCt = n, a.lgGeo = r;
		var o = a.trans = [];
		return o[0] = re(), o[1] = re(), o[2] = re(), a.mtRaw = ue(), a.mtRawInv = ue(), a.mtOverall = ue(), a.mtOverallInv = ue(), a.zoom = 1, i;
	}
	return n.prototype.getBoundingRect = function() {
		return En(null, this);
	}, n.prototype.getViewRect = function() {
		return Dn(null, this);
	}, n.prototype.getRoamTransform = function() {
		return _(J(this).trans[1]);
	}, n.prototype.dataToPoint = function(e, n, r) {
		var i = n ? J(this).mtRaw : J(this).mtOverall;
		return r ||= [], i ? P(r, e, i) : t(r, e);
	}, n.prototype.pointToData = function(e, n, r) {
		r ||= [];
		var i = J(this).mtOverallInv;
		return i ? P(r, e, i) : t(r, e);
	}, n.prototype.convertToPixel = function(e, t, n) {
		var r = On(t);
		return r === this ? r.dataToPoint(n) : null;
	}, n.prototype.convertFromPixel = function(e, t, n) {
		var r = On(t);
		return r === this ? r.pointToData(n) : null;
	}, n.prototype.containPoint = function(e) {
		var t = J(this);
		return ce(wn, t.dataRect), Qe(wn, wn, t.mtOverall), x(wn, e[0], e[1]);
	}, n.dimensions = ["x", "y"], n;
})(a);
var wn = oe();
function Tn(e) {
	return J(e).zoom;
}
function En(e, t) {
	return ce(e || oe(), J(t).dataRect);
}
function Dn(e, t) {
	return ce(e || oe(), J(t).viewRect);
}
ue(), ue(), oe(), ue(), ue();
function On(e) {
	var t = e.seriesModel;
	return t ? t.coordinateSystem : null;
}
re(), re();
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/custom/CustomSeries.js
var kn = {
	color: "fill",
	borderColor: "stroke"
}, An = {
	symbol: 1,
	symbolSize: 1,
	symbolKeepAspect: 1,
	legendIcon: 1,
	visualMeta: 1,
	liftZ: 1,
	decal: 1
}, Y = A(), jn = function(e) {
	z(t, e);
	function t() {
		var n = e !== null && e.apply(this, arguments) || this;
		return n.type = t.type, n;
	}
	return t.prototype.optionUpdated = function() {
		this.currentZLevel = this.get("zlevel", !0), this.currentZ = this.get("z", !0);
	}, t.prototype.getInitialData = function(e, t) {
		return I(null, this);
	}, t.prototype.getDataParams = function(t, n, r) {
		var i = e.prototype.getDataParams.call(this, t, n);
		return r && (i.info = Y(r).info), i;
	}, t.type = "series.custom", t.dependencies = [
		"grid",
		"polar",
		"geo",
		"singleAxis",
		"calendar",
		"matrix"
	], t.defaultOption = {
		coordinateSystem: "cartesian2d",
		z: 2,
		legendHoverLink: !0,
		clip: !1
	}, t;
}(f);
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/coord/cartesian/prepareCustom.js
function Mn(e, t) {
	return t ||= [0, 0], Ne(["x", "y"], function(n, r) {
		var i = this.getAxis(n), a = t[r], o = e[r] / 2;
		return i.type === "category" ? _t(i).w : Math.abs(i.dataToCoord(a - o) - i.dataToCoord(a + o));
	}, this);
}
function Nn(e) {
	var t = e.master.getRect();
	return {
		coordSys: {
			type: "cartesian2d",
			x: t.x,
			y: t.y,
			width: t.width,
			height: t.height
		},
		api: {
			coord: function(t) {
				return e.dataToPoint(t);
			},
			size: we(Mn, e)
		}
	};
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/coord/geo/prepareCustom.js
function Pn(e, t) {
	return t ||= [0, 0], Ne([0, 1], function(n) {
		var r = t[n], i = e[n] / 2, a = [], o = [];
		return a[n] = r - i, o[n] = r + i, a[1 - n] = o[1 - n] = t[1 - n], Math.abs(this.dataToPoint(a)[n] - this.dataToPoint(o)[n]);
	}, this);
}
function Fn(e) {
	var t = e.view, n = t.getBoundingRect();
	return {
		coordSys: {
			type: "geo",
			x: n.x,
			y: n.y,
			width: n.width,
			height: n.height,
			zoom: Tn(t)
		},
		api: {
			coord: function(t) {
				return e.dataToPoint(t);
			},
			size: we(Pn, e)
		}
	};
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/coord/single/prepareCustom.js
function In(e, t) {
	var n = this.getAxis(), r = t instanceof Array ? t[0] : t, i = (e instanceof Array ? e[0] : e) / 2;
	return n.type === "category" ? _t(n).w : Math.abs(n.dataToCoord(r - i) - n.dataToCoord(r + i));
}
function Ln(e) {
	var t = e.getRect();
	return {
		coordSys: {
			type: "singleAxis",
			x: t.x,
			y: t.y,
			width: t.width,
			height: t.height
		},
		api: {
			coord: function(t) {
				return e.dataToPoint(t);
			},
			size: we(In, e)
		}
	};
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/coord/polar/prepareCustom.js
function Rn(e, t) {
	return t ||= [0, 0], Ne(["Radius", "Angle"], function(n, r) {
		var i = "get" + n + "Axis", a = this[i](), o = t[r], s = e[r] / 2, c = a.type === "category" ? _t(a).w : Math.abs(a.dataToCoord(o - s) - a.dataToCoord(o + s));
		return n === "Angle" && (c = c * Math.PI / 180), c;
	}, this);
}
function zn(e) {
	var t = e.getRadiusAxis(), n = e.getAngleAxis(), r = t.getExtent();
	return r[0] > r[1] && r.reverse(), {
		coordSys: {
			type: "polar",
			cx: e.cx,
			cy: e.cy,
			r: r[1],
			r0: r[0]
		},
		api: {
			coord: function(r) {
				var i = t.dataToRadius(r[0]), a = n.dataToAngle(r[1]), o = e.coordToPoint([i, a]);
				return o.push(i, a * Math.PI / 180), o;
			},
			size: we(Rn, e)
		}
	};
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/coord/calendar/prepareCustom.js
function Bn(e) {
	var t = e.getRect(), n = e.getRangeInfo();
	return {
		coordSys: {
			type: "calendar",
			x: t.x,
			y: t.y,
			width: t.width,
			height: t.height,
			cellWidth: e.getCellWidth(),
			cellHeight: e.getCellHeight(),
			rangeInfo: {
				start: n.start,
				end: n.end,
				weeks: n.weeks,
				dayCount: n.allDay
			}
		},
		api: {
			coord: function(t, n) {
				return e.dataToPoint(t, n);
			},
			layout: function(t, n) {
				return e.dataToLayout(t, n);
			}
		}
	};
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/coord/matrix/prepareCustom.js
function Vn(e) {
	var t = e.getRect();
	return {
		coordSys: {
			type: "matrix",
			x: t.x,
			y: t.y,
			width: t.width,
			height: t.height
		},
		api: {
			coord: function(t, n) {
				return e.dataToPoint(t, n);
			},
			layout: function(t, n) {
				return e.dataToLayout(t, n);
			}
		}
	};
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/animation/customGraphicTransition.js
var Hn = {
	position: ["x", "y"],
	scale: ["scaleX", "scaleY"],
	origin: ["originX", "originY"]
}, Un = Ve(Hn);
fe(bt, function(e, t) {
	return e[t] = 1, e;
}, {}), bt.join(", ");
var Wn = [
	"",
	"style",
	"shape",
	"extra"
], Gn = A();
function Kn(e, t, n, r, i) {
	var a = e + "Animation", o = ot(e, r, i) || {}, s = Gn(t).userDuring;
	return o.duration > 0 && (o.during = s ? we(tr, {
		el: t,
		userDuring: s
	}) : null, o.setToFinal = !0, o.scope = e), Ie(o, n[a]), o;
}
function qn(e, t, n, r) {
	r ||= {};
	var i = r.dataIndex, a = r.isInit, o = r.clearStyle, s = n.isAnimationEnabled(), c = Gn(e), l = t.style;
	c.userDuring = t.during;
	var u = {}, d = {};
	if (ar(e, t, d), e.type === "compound") for (var f = e.shape.paths, p = t.shape.paths, m = 0; m < p.length; m++) {
		var h = p[m];
		rr("shape", h, f[m]);
	}
	else rr("shape", t, d), rr("extra", t, d);
	if (!a && s && (ir(e, t, u), nr("shape", e, t, u), nr("extra", e, t, u), or(e, t, l, u)), d.style = l, Zn(e, d, o), $n(e, t), s) if (a) {
		var g = {};
		V(Wn, function(e) {
			var n = e ? t[e] : t;
			n && n.enterFrom && (e && (g[e] = g[e] || {}), Ie(e ? g[e] : g, n.enterFrom));
		});
		var _ = Kn("enter", e, t, n, i);
		_.duration > 0 && e.animateFrom(g, _);
	} else Qn(e, t, i || 0, n, u);
	Jn(e, t), l ? e.dirty() : e.markRedraw();
}
function Jn(e, t) {
	for (var n = Gn(e).leaveToProps, r = 0; r < Wn.length; r++) {
		var i = Wn[r], a = i ? t[i] : t;
		a && a.leaveTo && (n ||= Gn(e).leaveToProps = {}, i && (n[i] = n[i] || {}), Ie(i ? n[i] : n, a.leaveTo));
	}
}
function Yn(e, t, n, r) {
	if (e) {
		var i = e.parent, a = Gn(e).leaveToProps;
		if (a) {
			var o = Kn("update", e, t, n, 0);
			o.done = function() {
				i && i.remove(e), r && r();
			}, e.animateTo(a, o);
		} else i && i.remove(e), r && r();
	}
}
function Xn(e) {
	return e === "all";
}
function Zn(e, t, n) {
	var r = t.style;
	if (!e.isGroup && r) {
		if (n) {
			e.useStyle({});
			for (var i = e.animators, a = 0; a < i.length; a++) {
				var o = i[a];
				o.targetName === "style" && o.changeTarget(e.style);
			}
		}
		e.setStyle(r);
	}
	t && (t.style = null, t && e.attr(t), t.style = r);
}
function Qn(e, t, n, r, i) {
	if (i) {
		var a = Kn("update", e, t, r, n);
		a.duration > 0 && e.animateFrom(i, a);
	}
}
function $n(e, t) {
	B(t, "silent") && (e.silent = t.silent), B(t, "ignore") && (e.ignore = t.ignore), e instanceof s && B(t, "invisible") && (e.invisible = t.invisible), e instanceof gt && B(t, "autoBatch") && (e.autoBatch = t.autoBatch);
}
var X = {}, er = {
	setTransform: function(e, t) {
		return X.el[e] = t, this;
	},
	getTransform: function(e) {
		return X.el[e];
	},
	setShape: function(e, t) {
		var n = X.el, r = n.shape ||= {};
		return r[e] = t, n.dirtyShape && n.dirtyShape(), this;
	},
	getShape: function(e) {
		var t = X.el.shape;
		if (t) return t[e];
	},
	setStyle: function(e, t) {
		var n = X.el, r = n.style;
		return r && (r[e] = t, n.dirtyStyle && n.dirtyStyle()), this;
	},
	getStyle: function(e) {
		var t = X.el.style;
		if (t) return t[e];
	},
	setExtra: function(e, t) {
		var n = X.el.extra || (X.el.extra = {});
		return n[e] = t, this;
	},
	getExtra: function(e) {
		var t = X.el.extra;
		if (t) return t[e];
	}
};
function tr() {
	var e = this, t = e.el;
	if (t) {
		var n = Gn(t).userDuring, r = e.userDuring;
		if (n !== r) {
			e.el = e.userDuring = null;
			return;
		}
		X.el = t, r(er);
	}
}
function nr(e, t, n, i) {
	var a = n[e];
	if (a) {
		var o = t[e], s;
		if (o) {
			var c = n.transition, l = a.transition;
			if (l) if (!s && (s = i[e] = {}), Xn(l)) Ie(s, o);
			else for (var u = r(l), d = 0; d < u.length; d++) {
				var f = u[d], p = o[f];
				s[f] = p;
			}
			else if (Xn(c) || nt(c, e) >= 0) {
				!s && (s = i[e] = {});
				for (var m = Ve(o), d = 0; d < m.length; d++) {
					var f = m[d], p = o[f];
					sr(a[f], p) && (s[f] = p);
				}
			}
		}
	}
}
function rr(e, t, n) {
	var r = t[e];
	if (r) for (var i = n[e] = {}, a = Ve(r), o = 0; o < a.length; o++) {
		var s = a[o];
		i[s] = R(r[s]);
	}
}
function ir(e, t, n) {
	for (var i = t.transition, a = Xn(i) ? bt : r(i || []), o = 0; o < a.length; o++) {
		var s = a[o];
		s === "style" || s === "shape" || s === "extra" || (n[s] = e[s]);
	}
}
function ar(e, t, n) {
	for (var r = 0; r < Un.length; r++) {
		var i = Un[r], a = Hn[i], o = t[i];
		o && (n[a[0]] = o[0], n[a[1]] = o[1]);
	}
	for (var r = 0; r < bt.length; r++) {
		var s = bt[r];
		t[s] != null && (n[s] = t[s]);
	}
}
function or(e, t, n, i) {
	if (n) {
		var a = e.style, o;
		if (a) {
			var s = n.transition, c = t.transition;
			if (s && !Xn(s)) {
				var l = r(s);
				!o && (o = i.style = {});
				for (var u = 0; u < l.length; u++) {
					var d = l[u], f = a[d];
					o[d] = f;
				}
			} else if (e.getAnimationStyleProps && (Xn(c) || Xn(s) || nt(c, "style") >= 0)) {
				var p = e.getAnimationStyleProps(), m = p ? p.style : null;
				if (m) {
					!o && (o = i.style = {});
					for (var h = Ve(n), u = 0; u < h.length; u++) {
						var d = h[u];
						if (m[d]) {
							var f = a[d];
							o[d] = f;
						}
					}
				}
			}
		}
	}
}
function sr(e, t) {
	return ze(e) ? e !== t : e != null && isFinite(e);
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/animation/customGraphicKeyframeAnimation.js
var cr = A(), lr = [
	"percent",
	"easing",
	"shape",
	"style",
	"extra"
];
function ur(e) {
	e.stopAnimation("keyframe"), e.attr(cr(e));
}
function dr(e, t, n) {
	if (!(!n.isAnimationEnabled() || !t)) {
		if (ye(t)) {
			V(t, function(t) {
				dr(e, t, n);
			});
			return;
		}
		var r = t.keyframes, i = t.duration;
		if (n && i == null) {
			var a = ot("enter", n, 0);
			i = a && a.duration;
		}
		if (!(!r || !i)) {
			var o = cr(e);
			V(Wn, function(n) {
				if (!(n && !e[n])) {
					var a;
					r.sort(function(e, t) {
						return e.percent - t.percent;
					}), V(r, function(r) {
						var s = e.animators, c = n ? r[n] : r;
						if (c) {
							var l = Ve(c);
							if (n || (l = he(l, function(e) {
								return nt(lr, e) < 0;
							})), l.length) {
								a || (a = e.animate(n, t.loop, !0), a.scope = "keyframe");
								for (var u = 0; u < s.length; u++) s[u] !== a && s[u].targetName === a.targetName && s[u].stopTracks(l);
								n && (o[n] = o[n] || {});
								var d = n ? o[n] : o;
								V(l, function(t) {
									d[t] = ((n ? e[n] : e) || {})[t];
								}), a.whenWithKeys(i * r.percent, c, l, r.easing);
							}
						}
					}), a && a.delay(t.delay || 0).duration(i).start(t.easing);
				}
			});
		}
	}
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/custom/CustomView.js
var Z = "emphasis", fr = "normal", pr = "blur", mr = "select", hr = [
	fr,
	Z,
	pr,
	mr
], gr = {
	normal: ["itemStyle"],
	emphasis: [Z, "itemStyle"],
	blur: [pr, "itemStyle"],
	select: [mr, "itemStyle"]
}, _r = {
	normal: ["label"],
	emphasis: [Z, "label"],
	blur: [pr, "label"],
	select: [mr, "label"]
}, vr = ["x", "y"], yr = "e\0\0", Q = {
	normal: {},
	emphasis: {},
	blur: {},
	select: {}
}, br = {
	cartesian2d: Nn,
	geo: Fn,
	single: Ln,
	polar: zn,
	calendar: Bn,
	matrix: Vn
};
function xr(e) {
	return e instanceof gt;
}
function Sr(e) {
	return e instanceof s;
}
function Cr(e, t) {
	t.copyTransform(e), Sr(t) && Sr(e) && (t.setStyle(e.style), t.z = e.z, t.z2 = e.z2, t.zlevel = e.zlevel, t.invisible = e.invisible, t.ignore = e.ignore, xr(t) && xr(e) && t.setShape(e.shape));
}
var wr = function(e) {
	z(t, e);
	function t() {
		var n = e !== null && e.apply(this, arguments) || this;
		return n.type = t.type, n;
	}
	return t.prototype.render = function(e, t, n, r) {
		this._progressiveEls = null;
		var i = this._data, a = e.getData(), o = this.group, s = Ar(e, a, t, n);
		i || o.removeAll(), a.diff(i).add(function(t) {
			Mr(n, null, t, s(t, r), e, o, a);
		}).remove(function(t) {
			var n = i.getItemGraphicEl(t);
			n && Yn(n, Y(n).option, e);
		}).update(function(t, c) {
			Mr(n, i.getItemGraphicEl(c), t, s(t, r), e, o, a);
		}).execute();
		var c = e.get("clip", !0) ? Je(e.coordinateSystem, !1, e) : null;
		c ? o.setClipPath(c) : o.removeClipPath(), this._data = a;
	}, t.prototype.incrementalPrepareRender = function(e, t, n) {
		this.group.removeAll(), this._data = null;
	}, t.prototype.incrementalRender = function(e, t, n, r, i) {
		var a = t.getData(), o = Ar(t, a, n, r), s = this._progressiveEls = [];
		function c(e) {
			e.isGroup || (e.incremental = L(t), e.ensureState("emphasis").hoverLayer = 2);
		}
		for (var l = e.start; l < e.end; l++) {
			var u = Mr(null, null, l, o(l, i), t, this.group, a);
			u && (u.traverse(c), s.push(u));
		}
	}, t.prototype.eachRendered = function(e) {
		Te(this._progressiveEls || this.group, e);
	}, t.prototype.filterForExposedEvent = function(e, t, n, r) {
		var i = t.element;
		if (i == null || n.name === i) return !0;
		for (; (n = n.__hostTarget || n.parent) && n !== this.group;) if (n.name === i) return !0;
		return !1;
	}, t.type = "custom", t;
}(T);
function Tr(e) {
	var t = e.type, n;
	if (t === "path") {
		var r = e.shape, i = r.width != null && r.height != null ? {
			x: r.x || 0,
			y: r.y || 0,
			width: r.width,
			height: r.height
		} : null, a = Kr(r);
		n = N(a, null, i, r.layout || "center"), Y(n).customPathData = a;
	} else if (t === "image") n = new ve({}), Y(n).customImagePath = e.style.image;
	else if (t === "text") n = new dt({});
	else if (t === "group") n = new y();
	else if (t === "compoundPath") {
		var r = e.shape;
		if (!r || !r.paths) {
			var o = "";
			M(o);
		}
		n = new be({ shape: { paths: Ne(r.paths, function(e) {
			if (e.type === "path") return N(e.shape.pathData, e, null);
			var t = le(e.type);
			return t || M(""), new t();
		}) } });
	} else {
		var s = le(t);
		if (!s) {
			var o = "";
			M(o);
		}
		n = new s();
	}
	return Y(n).customGraphicType = t, n.name = e.name, n.z2EmphasisLift = 1, n.z2SelectLift = 1, n;
}
function Er(e, t, n, r, i, a, o) {
	ur(t);
	var s = i && i.normal.cfg;
	s && t.setTextConfig(s), r && r.transition == null && (r.transition = vr);
	var c = r && r.style;
	if (c) {
		if (t.type === "text") {
			var l = c;
			B(l, "textFill") && (l.fill = l.textFill), B(l, "textStroke") && (l.stroke = l.textStroke);
		}
		var d = void 0, f = xr(t) ? c.decal : null;
		e && f && (f.dirty = !0, d = u(f, e)), c.__decalPattern = d;
	}
	if (Sr(t) && c) {
		var d = c.__decalPattern;
		d && (c.decal = d);
	}
	qn(t, r, a, {
		dataIndex: n,
		isInit: o,
		clearStyle: !0
	}), dr(t, r.keyframeAnimation, a);
}
function Dr(e, t, n, r, i) {
	var a = t.isGroup ? null : t, o = i && i[e].cfg;
	if (a) {
		var s = a.ensureState(e);
		if (r === !1) {
			var c = a.getState(e);
			c && (c.style = null);
		} else s.style = r || null;
		o && (s.textConfig = o), Xe(a);
	}
}
function Or(e, t, n) {
	if (!e.isGroup) {
		var r = e, i = n.currentZ, a = n.currentZLevel;
		r.z = i, r.zlevel = a;
		var o = t.z2;
		o != null && (r.z2 = o || 0);
		for (var s = 0; s < hr.length; s++) kr(r, t, hr[s]);
	}
}
function kr(e, t, n) {
	var r = n === fr, i = r ? t : Rr(t, n), a = i ? i.z2 : null, o;
	a != null && (o = r ? e : e.ensureState(n), o.z2 = a || 0);
}
function Ar(e, t, r, i) {
	var a = e.get("renderItem");
	if (typeof a == "string") {
		var o = n(a);
		o && (a = o);
	}
	var s = e.coordinateSystem, c = {};
	s && (c = s.prepareCustoms ? s.prepareCustoms(s) : br[s.type](s));
	for (var l = We({
		getWidth: i.getWidth,
		getHeight: i.getHeight,
		getZr: i.getZr,
		getDevicePixelRatio: i.getDevicePixelRatio,
		value: T,
		style: D,
		ordinalRawValue: E,
		styleEmphasis: O,
		visual: j,
		barLayout: M,
		currentSeriesIndices: N,
		font: P
	}, c.api || {}), u = {
		context: {},
		seriesId: e.id,
		seriesName: e.name,
		seriesIndex: e.seriesIndex,
		coordSys: c.coordSys,
		dataInsideLength: t.count(),
		encode: jr(e.getData()),
		itemPayload: e.get("itemPayload") || {}
	}, d, f, p = {}, h = {}, g = {}, _ = {}, y = 0; y < hr.length; y++) {
		var b = hr[y];
		g[b] = e.getModel(gr[b]), _[b] = e.getModel(_r[b]);
	}
	function x(e) {
		return e === d ? f ||= t.getItemModel(e) : t.getItemModel(e);
	}
	function C(e, n) {
		return t.hasItemOption ? e === d ? p[n] || (p[n] = x(e).getModel(gr[n])) : x(e).getModel(gr[n]) : g[n];
	}
	function w(e, n) {
		return t.hasItemOption ? e === d ? h[n] || (h[n] = x(e).getModel(_r[n])) : x(e).getModel(_r[n]) : _[n];
	}
	return function(e, n) {
		return d = e, f = null, p = {}, h = {}, a && a(We({
			dataIndexInside: e,
			dataIndex: t.getRawIndex(e),
			actionType: n ? n.type : null
		}, u), l);
	};
	function T(e, n) {
		return n ??= d, t.getStore().get(t.getDimensionIndex(e || 0), n);
	}
	function E(e, n) {
		n ??= d, e ||= 0;
		var r = t.getDimensionInfo(e);
		if (!r) {
			var i = t.getDimensionIndex(e);
			return i >= 0 ? t.getStore().get(i, n) : void 0;
		}
		var a = t.get(r.name, n), o = r && r.ordinalMeta;
		return o ? o.categories[a] : a;
	}
	function D(n, r) {
		r ??= d;
		var i = t.getItemVisual(r, "style"), a = i && i.fill, o = i && i.opacity, s = C(r, fr).getItemStyle();
		a != null && (s.fill = a), o != null && (s.opacity = o);
		var c = { inheritColor: Ye(a) ? a : v.color.neutral99 }, l = w(r, fr), u = m(l, null, c, !1, !0);
		u.text = l.getShallow("show") ? ut(e.getFormattedLabel(r, fr), lt(t, r)) : null;
		var f = ie(l, c, !1);
		return A(n, s), s = tn(s, u, f), n && k(s, n), s.legacy = !0, s;
	}
	function O(n, r) {
		r ??= d;
		var i = C(r, Z).getItemStyle(), a = w(r, Z), o = m(a, null, null, !0, !0);
		o.text = a.getShallow("show") ? vt(e.getFormattedLabel(r, Z), e.getFormattedLabel(r, fr), lt(t, r)) : null;
		var s = ie(a, null, !0);
		return A(n, i), i = tn(i, o, s), n && k(i, n), i.legacy = !0, i;
	}
	function k(e, t) {
		for (var n in t) B(t, n) && (e[n] = t[n]);
	}
	function A(e, t) {
		e && (e.textFill && (t.textFill = e.textFill), e.textPosition && (t.textPosition = e.textPosition));
	}
	function j(e, n) {
		if (n ??= d, B(kn, e)) {
			var r = t.getItemVisual(n, "style");
			return r ? r[kn[e]] : null;
		}
		if (B(An, e)) return t.getItemVisual(n, e);
	}
	function M(e) {
		if (s.type === "cartesian2d") return tt(We({ axis: s.getBaseAxis() }, e));
	}
	function N() {
		return r.getCurrentSeriesIndices();
	}
	function P(e) {
		return S(e, r);
	}
}
function jr(e) {
	var t = {};
	return V(e.dimensions, function(n) {
		var r = e.getDimensionInfo(n);
		if (!r.isExtraCoord) {
			var i = r.coordDim, a = t[i] = t[i] || [];
			a[r.coordDimIndex] = e.getDimensionIndex(n);
		}
	}), t;
}
function Mr(e, t, n, r, i, a, o) {
	if (!r) {
		a.remove(t);
		return;
	}
	var s = Nr(e, t, n, r, i, a);
	return s && o.setItemGraphicEl(n, s), s && je(s, r.focus, r.blurScope, r.emphasisDisabled), s;
}
function Nr(e, t, n, r, i, a) {
	var o = -1, s = t;
	t && Pr(t, r, i) && (o = nt(a.childrenRef(), t), t = null);
	var c = !t, l = t;
	l ? l.clearStates() : (l = Tr(r), s && Cr(s, l)), r.morph === !1 ? l.disableMorphing = !0 : l.disableMorphing &&= !1, r.tooltipDisabled && (l.tooltipDisabled = !0), Q.normal.cfg = Q.normal.conOpt = Q.emphasis.cfg = Q.emphasis.conOpt = Q.blur.cfg = Q.blur.conOpt = Q.select.cfg = Q.select.conOpt = null, Q.isLegacy = !1, Ir(l, n, r, i, c, Q), Fr(l, n, r, i, c), Er(e, l, n, r, Q, i, c), B(r, "info") && (Y(l).info = r.info);
	for (var u = 0; u < hr.length; u++) {
		var d = hr[u];
		if (d !== fr) {
			var f = Rr(r, d), p = zr(r, f, d);
			Dr(d, l, f, p, Q);
		}
	}
	return Or(l, r, i), r.type === "group" && Br(e, l, n, r, i), o >= 0 ? a.replaceAt(l, o) : a.add(l), l;
}
function Pr(e, t, n) {
	var r = Y(e), i = t.type, a = t.shape, o = t.style;
	return n.isUniversalTransitionEnabled() || i != null && i !== r.customGraphicType || i === "path" && qr(a) && Kr(a) !== r.customPathData || i === "image" && B(o, "image") && o.image !== r.customImagePath;
}
function Fr(e, t, n, r, i) {
	var a = n.clipPath;
	if (a === !1) e && e.getClipPath() && e.removeClipPath();
	else if (a) {
		var o = e.getClipPath();
		o && Pr(o, a, r) && (o = null), o || (o = Tr(a), e.setClipPath(o)), Er(null, o, t, a, null, r, i);
	}
}
function Ir(e, t, n, r, i, a) {
	if (!(e.isGroup || e.type === "compoundPath")) {
		Lr(n, null, a), Lr(n, Z, a);
		var o = a.normal.conOpt, s = a.emphasis.conOpt, c = a.blur.conOpt, l = a.select.conOpt;
		if (o != null || s != null || l != null || c != null) {
			var u = e.getTextContent();
			if (o === !1) u && e.removeTextContent();
			else {
				o = a.normal.conOpt = o || { type: "text" }, u ? u.clearStates() : (u = Tr(o), e.setTextContent(u)), Er(null, u, t, o, null, r, i);
				for (var d = o && o.style, f = 0; f < hr.length; f++) {
					var p = hr[f];
					if (p !== fr) {
						var m = a[p].conOpt;
						Dr(p, u, m, zr(o, m, p), null);
					}
				}
				d ? u.dirty() : u.markRedraw();
			}
		}
	}
}
function Lr(e, t, n) {
	var r = t ? Rr(e, t) : e, i = t ? zr(e, r, Z) : e.style, a = e.type, o = r ? r.textConfig : null, s = e.textContent, c = s ? t ? Rr(s, t) : s : null;
	if (i && (n.isLegacy || Qt(i, a, !!o, !!c))) {
		n.isLegacy = !0;
		var l = $t(i, a, !t);
		!o && l.textConfig && (o = l.textConfig), !c && l.textContent && (c = l.textContent);
	}
	if (!t && c) {
		var u = c;
		!u.type && (u.type = "text");
	}
	var d = t ? n[t] : n.normal;
	d.cfg = o, d.conOpt = c;
}
function Rr(e, t) {
	return t ? e ? e[t] : null : e;
}
function zr(e, t, n) {
	var r = t && t.style;
	return r == null && n === Z && e && (r = e.styleEmphasis), r;
}
function Br(e, t, n, r, i) {
	var a = r.children, o = a ? a.length : 0, s = r.$mergeChildren, c = s === "byName" || r.diffChildrenByName, l = s === !1;
	if (!(!o && !c && !l)) {
		if (c) {
			Hr({
				api: e,
				oldChildren: t.children() || [],
				newChildren: a || [],
				dataIndex: n,
				seriesModel: i,
				group: t
			});
			return;
		}
		l && t.removeAll();
		for (var u = 0; u < o; u++) {
			var d = a[u], f = t.childAt(u);
			d ? (d.ignore ??= !1, Nr(e, f, n, d, i, t)) : f.ignore = !0;
		}
		for (var p = t.childCount() - 1; p >= u; p--) Vr(t, t.childAt(p), i);
	}
}
function Vr(e, t, n) {
	t && Yn(t, Y(e).option, n);
}
function Hr(e) {
	new ft(e.oldChildren, e.newChildren, Ur, Ur, e).add(Wr).update(Wr).remove(Gr).execute();
}
function Ur(e, t) {
	return (e && e.name) ?? yr + t;
}
function Wr(e, t) {
	var n = this.context, r = e == null ? null : n.newChildren[e], i = t == null ? null : n.oldChildren[t];
	Nr(n.api, i, n.dataIndex, r, n.seriesModel, n.group);
}
function Gr(e) {
	var t = this.context, n = t.oldChildren[e];
	n && Yn(n, Y(n).option, t.seriesModel);
}
function Kr(e) {
	return e && (e.pathData || e.d);
}
function qr(e) {
	return e && (B(e, "pathData") || B(e, "d"));
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/custom/install.js
function Jr(e) {
	e.registerChartView(wr), e.registerSeriesModel(jn);
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/component/graphic/GraphicModel.js
function Yr(e, t) {
	var n = e.existing;
	if (t.id = e.keyInfo.id, !t.type && n && (t.type = n.type), t.parentId == null) {
		var r = t.parentOption;
		r ? t.parentId = r.id : n && (t.parentId = n.parentId);
	}
	t.parentOption = null;
}
function Xr(e, t) {
	var n;
	return V(t, function(t) {
		e[t] != null && e[t] !== "auto" && (n = !0);
	}), n;
}
function Zr(e, t, n) {
	var r = Ie({}, n), a = e[t], o = n.$action || "merge";
	o === "merge" ? a ? (Ae(a, r, !0), i(a, r, { ignoreSize: !0 }), yt(n, a), ei(n, a), ei(n, a, "shape"), ei(n, a, "style"), ei(n, a, "extra"), n.clipPath = a.clipPath) : e[t] = r : o === "replace" ? e[t] = r : o === "remove" && a && (e[t] = null);
}
var Qr = [
	"transition",
	"enterFrom",
	"leaveTo"
], $r = Qr.concat([
	"enterAnimation",
	"updateAnimation",
	"leaveAnimation"
]);
function ei(e, t, n) {
	if (n && (!e[n] && t[n] && (e[n] = {}), e = e[n], t = t[n]), !(!e || !t)) for (var r = n ? Qr : $r, i = 0; i < r.length; i++) {
		var a = r[i];
		e[a] == null && t[a] != null && (e[a] = t[a]);
	}
}
function ti(e, t) {
	if (e && (e.hv = t.hv = [Xr(t, ["left", "right"]), Xr(t, ["top", "bottom"])], e.type === "group")) {
		var n = e, r = t;
		n.width ??= r.width = 0, n.height ??= r.height = 0;
	}
}
var ni = function(e) {
	z(t, e);
	function t() {
		var n = e !== null && e.apply(this, arguments) || this;
		return n.type = t.type, n.preventAutoZ = !0, n;
	}
	return t.prototype.mergeOption = function(t, n) {
		var r = this.option.elements;
		this.option.elements = null, e.prototype.mergeOption.call(this, t, n), this.option.elements = r;
	}, t.prototype.optionUpdated = function(e, t) {
		var n = this.option, r = (t ? n : e).elements, i = n.elements = t ? [] : n.elements, a = [];
		this._flatten(r, a, null);
		var o = Re(i, a, "normalMerge"), s = this._elOptionsToUpdate = [];
		V(o, function(e, t) {
			var n = e.newOption;
			n && (s.push(n), Yr(e, n), Zr(i, t, n), ti(i[t], n));
		}, this), n.elements = he(i, function(e) {
			return e && delete e.$action, e != null;
		});
	}, t.prototype._flatten = function(e, t, n) {
		V(e, function(e) {
			if (e) {
				n && (e.parentOption = n), t.push(e);
				var r = e.children;
				r && r.length && this._flatten(r, t, e), delete e.children;
			}
		}, this);
	}, t.prototype.useElOptionsToUpdate = function() {
		var e = this._elOptionsToUpdate;
		return this._elOptionsToUpdate = null, e;
	}, t.type = "graphic", t.defaultOption = { elements: [] }, t;
}(w), ri = {
	path: null,
	compoundPath: null,
	group: y,
	image: ve,
	text: dt
}, $ = A(), ii = function(t) {
	z(n, t);
	function n() {
		var e = t !== null && t.apply(this, arguments) || this;
		return e.type = n.type, e;
	}
	return n.prototype.init = function() {
		this._elMap = at();
	}, n.prototype.render = function(e, t, n) {
		e !== this._lastGraphicModel && this._clear(), this._lastGraphicModel = e, this._updateElements(e), this._relocate(e, n);
	}, n.prototype._updateElements = function(t) {
		var n = t.useElOptionsToUpdate();
		if (n) {
			var r = this._elMap, i = this.group, a = t.get("z"), o = t.get("zlevel");
			V(n, function(n) {
				var s = pt(n.id, null), c = s == null ? null : r.get(s), l = pt(n.parentId, null), u = l == null ? i : r.get(l), d = n.type, f = n.style;
				d === "text" && f && n.hv && n.hv[1] && (f.textVerticalAlign = f.textBaseline = f.verticalAlign = f.align = null);
				var p = n.textContent, m = n.textConfig;
				if (f && Qt(f, d, !!m, !!p)) {
					var h = $t(f, d, !0);
					!m && h.textConfig && (m = n.textConfig = h.textConfig), !p && h.textContent && (p = h.textContent);
				}
				var g = li(n), _ = n.$action || "merge", v = _ === "merge", y = _ === "replace";
				if (v) {
					var b = !c, x = c;
					b ? x = oi(s, u, n.type, r) : (x && ($(x).isNew = !1), ur(x)), x && (qn(x, g, t, { isInit: b }), ci(x, n, a, o));
				} else if (y) {
					si(c, n, r, t);
					var S = oi(s, u, n.type, r);
					S && (qn(S, g, t, { isInit: !0 }), ci(S, n, a, o));
				} else _ === "remove" && (Jn(c, n), si(c, n, r, t));
				var C = r.get(s);
				if (C && p) if (v) {
					var w = C.getTextContent();
					w ? w.attr(p) : C.setTextContent(new dt(p));
				} else y && C.setTextContent(new dt(p));
				if (C) {
					var T = n.clipPath;
					if (T) {
						var E = T.type, D = void 0, b = !1;
						if (v) {
							var O = C.getClipPath();
							b = !O || $(O).type !== E, D = b ? ai(E) : O;
						} else y && (b = !0, D = ai(E));
						C.setClipPath(D), qn(D, T, t, { isInit: b }), dr(D, T.keyframeAnimation, t);
					}
					var k = $(C);
					C.setTextConfig(m), k.option = n, ui(C, t, n), e({
						el: C,
						componentModel: t,
						itemName: C.name,
						itemTooltipOption: n.tooltip
					}), dr(C, n.keyframeAnimation, t);
				}
			});
		}
	}, n.prototype._relocate = function(e, t) {
		for (var n = e.option.elements, r = this.group, i = this._elMap, a = t.getWidth(), o = t.getHeight(), s = ["x", "y"], c = 0; c < n.length; c++) {
			var l = n[c], u = pt(l.id, null), d = u == null ? null : i.get(u);
			if (!(!d || !d.isGroup)) {
				var f = d.parent, m = f === r, h = $(d), g = $(f);
				h.width = p(h.option.width, m ? a : g.width) || 0, h.height = p(h.option.height, m ? o : g.height) || 0;
			}
		}
		for (var c = n.length - 1; c >= 0; c--) {
			var l = n[c], u = pt(l.id, null), d = u == null ? null : i.get(u);
			if (d) {
				var f = d.parent, g = $(f), _ = f === r ? {
					width: a,
					height: o
				} : {
					width: g.width,
					height: g.height
				}, v = {}, y = ne(d, l, _, null, {
					hv: l.hv,
					boundingMode: l.bounding
				}, v);
				if (!$(d).isNew && y) {
					for (var b = l.transition, x = {}, S = 0; S < s.length; S++) {
						var C = s[S], w = v[C];
						b && (Xn(b) || nt(b, C) >= 0) ? x[C] = w : d[C] = w;
					}
					qe(d, x, e, 0);
				} else d.attr(v);
			}
		}
	}, n.prototype._clear = function() {
		var e = this, t = this._elMap;
		t.each(function(n) {
			si(n, $(n).option, t, e._lastGraphicModel);
		}), this._elMap = at();
	}, n.prototype.dispose = function() {
		this._clear();
	}, n.type = "graphic", n;
}(j);
function ai(e) {
	var t = new (B(ri, e) ? ri[e] : le(e))({});
	return $(t).type = e, t;
}
function oi(e, t, n, r) {
	var i = ai(n);
	return t.add(i), r.set(e, i), $(i).id = e, $(i).isNew = !0, i;
}
function si(e, t, n, r) {
	e && e.parent && (e.type === "group" && e.traverse(function(e) {
		si(e, t, n, r);
	}), Yn(e, t, r), n.removeKey($(e).id));
}
function ci(e, t, n, r) {
	e.isGroup || V([
		["cursor", s.prototype.cursor],
		["zlevel", r || 0],
		["z", n || 0],
		["z2", 0]
	], function(n) {
		var r = n[0];
		B(t, r) ? e[r] = ut(t[r], n[1]) : e[r] ?? (e[r] = n[1]);
	}), V(Ve(t), function(n) {
		if (n.indexOf("on") === 0) {
			var r = t[n];
			e[n] = st(r) ? r : null;
		}
	}), B(t, "draggable") && (e.draggable = t.draggable), t.name != null && (e.name = t.name), t.id != null && (e.id = t.id);
}
function li(e) {
	return e = Ie({}, e), V([
		"id",
		"parentId",
		"$action",
		"hv",
		"bounding",
		"textContent",
		"clipPath"
	].concat(F), function(t) {
		delete e[t];
	}), e;
}
function ui(e, t, n) {
	var r = pe(e).eventData;
	!e.silent && !e.ignore && !r && (r = pe(e).eventData = {
		componentType: "graphic",
		componentIndex: t.componentIndex,
		name: e.name
	}), r && (r.info = n.info);
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/component/graphic/install.js
function di(e) {
	e.registerComponentModel(ni), e.registerComponentView(ii), e.registerPreprocessor(function(e) {
		var t = e.graphic;
		ye(t) ? !t[0] || !t[0].elements ? e.graphic = [{ elements: t }] : e.graphic = [e.graphic[0]] : t && !t.elements && (e.graphic = [{ elements: [t] }]);
	});
}
//#endregion
//#region web_src/features/repository-task-options.ts
var fi = [
	{
		key: "mirror",
		label: "镜像迁移",
		color: "--color-purple",
		lineStyle: "solid",
		lineDash: []
	},
	{
		key: "local",
		label: "本地迁移",
		color: "--color-blue",
		lineStyle: "dashed",
		lineDash: [6, 4]
	},
	{
		key: "conversion",
		label: "转换镜像",
		color: "--color-orange",
		lineStyle: "dotted",
		lineDash: [2, 4]
	},
	{
		key: "sync",
		label: "镜像同步",
		color: "--color-green",
		lineStyle: "solid",
		lineDash: []
	}
], pi = {
	queued: "排队中",
	running: "进行中",
	succeeded: "成功",
	failed: "失败",
	interrupted: "已中断"
};
function mi(e) {
	return e > 0 ? new Intl.DateTimeFormat("sv-SE", {
		timeZone: "Asia/Shanghai",
		year: "numeric",
		month: "2-digit",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
		second: "2-digit",
		hour12: !1
	}).format(new Date(e)) : "未知";
}
function hi(e) {
	return e < 1e3 ? `${e} 毫秒` : e < 6e4 ? `${(e / 1e3).toFixed(1)} 秒` : e < 36e5 ? `${(e / 6e4).toFixed(1)} 分钟` : `${(e / 36e5).toFixed(1)} 小时`;
}
function gi(e) {
	return [
		`#${e.id} · ${fi.find((t) => t.key === e.kind)?.label ?? "任务"} · ${pi[e.state] ?? e.state}`,
		`开始：${mi(e.started)}`,
		`结束：${e.finished ? mi(e.finished) : "—"}`,
		`耗时：${e.duration === null ? "—" : hi(e.duration)}`,
		`新增修订：${e.added}${e.after === null ? "（修订数据未知）" : ""}`,
		`版本：${e.before === null || e.after === null ? "未知" : `r${e.before} → r${e.after}`}`
	].join("\n");
}
function _i(e, t, n, r, i = 1e3) {
	let a = i < 480 ? 56 : i < 768 ? 88 : 128, o = (e) => i < 480 ? mi(e).slice(5, 16).replace(" ", "\n") : mi(e).slice(5, 16), s = e.filter((e) => e.started > 0).filter((e) => n[fi.find((t) => t.key === e.kind)?.label ?? ""] !== !1), c = s.length ? s.reduce((e, t) => Math.min(e, t.started), Infinity) : 0, l = s.length ? s.reduce((e, t) => Math.max(e, t.started), -Infinity) : 1, u = Math.max(1e3, (l - c) * .02), d = s.filter((e) => e.after !== null).map((e) => ({
		value: [e.started, e.after],
		task: e
	})), f = d[0]?.task;
	f?.kind === "sync" && f.before !== null && f.before !== f.after && d.unshift({
		value: [f.started, f.before],
		task: f
	});
	let p = s.filter((e) => e.duration !== null).map((e) => ({
		value: [e.started, e.duration],
		task: e
	})), m = (e, t, n, r, i) => ({
		id: e,
		name: t,
		type: "custom",
		z: 1,
		yAxisIndex: 1,
		encode: {
			x: 0,
			y: -1
		},
		renderItem: (t, i) => {
			let a = i.coord([i.value(0), 0])[0], o = t.coordSys;
			if (!(!("x" in o) || !("width" in o) || !("y" in o) || !("height" in o) || typeof o.x != "number" || typeof o.width != "number" || typeof o.y != "number" || typeof o.height != "number") && !(a < o.x || a > o.x + o.width)) return {
				type: "line",
				shape: {
					x1: a,
					y1: o.y,
					x2: a,
					y2: o.y + o.height
				},
				style: {
					stroke: n,
					lineWidth: e === "failed" ? 2 : 1.5,
					lineDash: r
				}
			};
		},
		data: i.map((e) => ({
			value: [e.started],
			task: e
		}))
	});
	return {
		animation: !1,
		textStyle: {
			color: r("--color-text"),
			fontFamily: "Microsoft YaHei, sans-serif"
		},
		grid: [{
			left: a,
			right: a,
			top: 72,
			bottom: 76
		}],
		graphic: [{
			type: "text",
			x: a - 8,
			y: 48,
			style: {
				align: "right",
				verticalAlign: "top",
				text: "版本号",
				fill: r("--color-text"),
				fontSize: 12
			}
		}, {
			type: "text",
			x: i - a + 8,
			y: 48,
			style: {
				align: "left",
				verticalAlign: "top",
				text: "耗时",
				fill: r("--color-text"),
				fontSize: 12
			}
		}],
		legend: {
			top: 8,
			selected: n,
			textStyle: { color: r("--color-text") },
			data: ["版本号", "耗时"]
		},
		xAxis: [{
			show: s.length > 0,
			type: "time",
			containShape: !1,
			min: c - u,
			max: l + u,
			axisLabel: {
				hideOverlap: !0,
				formatter: o
			},
			axisLine: { lineStyle: { color: r("--color-secondary") } }
		}],
		yAxis: [{
			show: s.length > 0,
			type: "value",
			scale: !0,
			minInterval: 1,
			axisLabel: { formatter: (e) => `r${e}` },
			splitLine: { lineStyle: { color: r("--color-secondary") } }
		}, {
			show: s.length > 0,
			type: "value",
			position: "right",
			min: 0,
			axisLabel: { formatter: hi },
			splitLine: { show: !1 }
		}],
		tooltip: {
			trigger: "axis",
			axisPointer: {
				axis: "x",
				type: "line",
				snap: !0,
				lineStyle: {
					color: r("--color-text-light-2"),
					type: "dashed"
				}
			},
			renderMode: "richText",
			confine: !0,
			backgroundColor: r("--color-box-body"),
			borderColor: r("--color-secondary"),
			textStyle: {
				color: r("--color-text"),
				fontSize: 12
			},
			formatter: (e) => {
				let t = (Array.isArray(e) ? e[0] : e)?.data;
				return t?.task ? s.filter((e) => e.started === t.task?.started).map(gi).join("\n\n") : "";
			}
		},
		dataZoom: [{
			type: "slider",
			xAxisIndex: [0],
			left: a,
			right: a,
			bottom: 8,
			height: 24,
			start: t.start,
			end: t.end,
			filterMode: "none",
			brushSelect: !1,
			moveHandleSize: 0,
			textStyle: { color: r("--color-text") },
			labelFormatter: (e) => i < 768 ? o(e) : mi(e).slice(0, 16)
		}],
		series: [
			{
				id: "revision",
				name: "版本号",
				type: "line",
				yAxisIndex: 0,
				step: "end",
				connectNulls: !1,
				symbolSize: 6,
				itemStyle: { color: r("--color-primary") },
				data: d
			},
			{
				id: "duration",
				name: "耗时",
				type: "line",
				yAxisIndex: 1,
				connectNulls: !1,
				symbolSize: 6,
				itemStyle: { color: r("--color-orange") },
				data: p
			},
			...fi.filter((e) => e.key !== "sync").map((e) => m(e.key, e.label, r(e.color), e.lineDash, s.filter((t) => t.kind === e.key && t.state !== "failed"))),
			m("failed", "失败", r("--color-red"), [], s.filter((e) => e.state === "failed")),
			{
				id: "task-times",
				name: "任务时间",
				type: "custom",
				yAxisIndex: 1,
				encode: {
					x: 0,
					y: -1
				},
				renderItem: () => null,
				data: s.map((e) => ({
					value: [e.started],
					task: e
				}))
			}
		]
	};
}
//#endregion
//#region web_src/components/RepositoryTaskHistory.vue?vue&type=script&setup=true&lang.ts
var vi = ["aria-busy"], yi = {
	key: 0,
	class: "ui warning message",
	role: "alert"
}, bi = ["disabled"], xi = {
	class: "task-summary",
	"aria-label": "仓库任务统计"
}, Si = { class: "ui segment" }, Ci = { class: "ui segment" }, wi = { class: "ui segment" }, Ti = { class: "ui segment" }, Ei = {
	class: "ui segments task-trend",
	"aria-label": "仓库任务趋势"
}, Di = { class: "ui segment task-trend-filters" }, Oi = { "aria-label": "任务类型" }, ki = ["checked", "onChange"], Ai = { class: "ui segment" }, ji = {
	key: 0,
	class: "muted"
}, Mi = {
	key: 1,
	class: "muted"
}, Ni = {
	class: "task-history-table",
	"data-task-status": ""
}, Pi = { class: "ui top attached header" }, Fi = { key: 0 }, Ii = { class: "ui attached table segment table-scroll" }, Li = { class: "ui very basic unstackable table" }, Ri = {
	key: 0,
	"aria-label": "任务进行中"
}, zi = { class: "task-date" }, Bi = { class: "task-date" }, Vi = ["data-api"], Hi = {
	class: "action-icon-button",
	title: "重试任务",
	"aria-label": "重试任务"
}, Ui = { key: 0 }, Wi = { colspan: "8" }, Gi = {
	key: 0,
	class: "ui pagination menu task-pagination",
	"aria-label": "任务记录分页"
}, Ki = ["disabled"], qi = { class: "item" }, Ji = ["disabled"], Yi = /* @__PURE__ */ it(/* @__PURE__ */ kt({
	__name: "RepositoryTaskHistory",
	props: {
		url: {},
		retryUrl: {},
		initForms: { type: Function }
	},
	setup(e) {
		_e([
			Jr,
			Sn,
			di,
			Ke,
			me,
			Fe,
			$e,
			Ue
		]);
		let t = e, n = At(), r = At(""), i = At(!1), a = At(), o, s, c, l, u = !1, d = 0, f = {
			start: 0,
			end: 100
		}, p = At(Object.fromEntries(fi.map((e) => [e.label, !0])));
		function m(e) {
			p.value[e] = p.value[e] === !1, f = {
				start: 0,
				end: 100
			}, h();
		}
		function h() {
			if (!o || !a.value) return;
			let e = getComputedStyle(a.value);
			o.setOption(_i(n.value?.points ?? [], f, p.value, (t) => e.getPropertyValue(t).trim(), o.getWidth()), { notMerge: !0 });
		}
		async function g(e = n.value?.page ?? 1) {
			let a = ++d;
			l?.abort(), c && clearTimeout(c), l = new AbortController(), i.value = !0;
			try {
				let i = new URL(t.url, location.origin);
				i.searchParams.set("page", String(e));
				let s = await fetch(i, {
					signal: l.signal,
					headers: { Accept: "application/json" }
				});
				if (a !== d || u) return;
				if (s.status === 401 || s.redirected && new URL(s.url).pathname === "/user/login") {
					u = !0, n.value = void 0, o?.clear(), location.assign("/user/login");
					return;
				}
				if (s.status === 403 || s.status === 404) {
					u = !0, n.value = void 0, o?.clear(), r.value = "仓库不存在或访问权限已变化，请刷新页面。";
					return;
				}
				if (!s.ok) throw Error();
				let c = await s.json();
				if (a !== d || u) return;
				if (n.value && n.value.repositoryState !== c.repositoryState) {
					location.reload(), u = !0;
					return;
				}
				if (n.value = c, r.value = "", await Ct(), t.initForms(), h(), e !== 1 || new URL(location.href).searchParams.has("page")) {
					let e = new URL(location.href);
					e.searchParams.set("page", String(c.page)), history.replaceState(null, "", e);
				}
			} catch {
				a === d && !u && (r.value = "任务记录暂不可用，正在重试…");
			} finally {
				a === d && (i.value = !1, u || (c = setTimeout(() => void g(), n.value?.active ? 5e3 : 3e4)));
			}
		}
		return Dt(() => {
			a.value && (o = te(a.value, void 0, { renderer: "svg" }), o.on("datazoom", (e) => {
				let t = e;
				t.start !== void 0 && t.end !== void 0 && (f = {
					start: t.start,
					end: t.end
				});
			}), o.on("legendselectchanged", (e) => {
				p.value = {
					...p.value,
					...e.selected
				}, h();
			}), s = new ResizeObserver(() => {
				o?.resize(), h();
			}), s.observe(a.value), h());
			let e = Number(new URL(location.href).searchParams.get("page"));
			g(Number.isInteger(e) && e > 0 ? e : 1);
		}), wt(() => {
			u = !0, d++, l?.abort(), c && clearTimeout(c), s?.disconnect(), o?.dispose();
		}), (t, o) => (G(), K("div", {
			class: "repository-task-history",
			"aria-busy": i.value
		}, [
			r.value ? (G(), K("p", yi, [St(H(r.value) + " ", 1), jt(u) ? U("", !0) : (G(), K("button", {
				key: 0,
				type: "button",
				class: "ui basic button action-small",
				disabled: i.value,
				onClick: o[0] ||= (e) => g()
			}, "重试", 8, bi))])) : U("", !0),
			W("section", xi, [
				W("div", Si, [o[3] ||= W("span", null, "任务总数", -1), W("strong", null, H(n.value?.summary.total ?? "—"), 1)]),
				W("div", Ci, [
					o[4] ||= W("span", null, "成功率", -1),
					W("strong", null, H(n.value?.summary.successRate == null ? "—" : `${n.value.summary.successRate.toFixed(1)}%`), 1),
					o[5] ||= W("small", { class: "muted" }, "成功／已结束任务", -1)
				]),
				W("div", wi, [o[6] ||= W("span", null, "版本跨度", -1), W("strong", null, H(n.value?.summary.revisionSpan ?? "—"), 1)]),
				W("div", Ti, [
					o[7] ||= W("span", null, "累计新增修订", -1),
					W("strong", null, H(n.value?.summary.added ?? "—"), 1),
					o[8] ||= W("small", { class: "muted" }, "仅统计成功同步", -1)
				])
			]),
			W("section", Ei, [
				o[9] ||= W("div", { class: "ui secondary segment task-trend-heading" }, [W("h4", null, "任务趋势")], -1),
				W("div", Di, [W("fieldset", Oi, [(G(!0), K(Ot, null, Tt(jt(fi), (e) => (G(), K("label", { key: e.key }, [
					W("input", {
						type: "checkbox",
						checked: p.value[e.label] !== !1,
						onChange: (t) => m(e.label)
					}, null, 40, ki),
					e.key === "sync" ? U("", !0) : (G(), K("span", {
						key: 0,
						class: "task-kind-line",
						style: xt({
							borderColor: `var(${e.color})`,
							borderTopStyle: e.lineStyle
						}),
						"aria-hidden": "true"
					}, null, 4)),
					St(H(e.label), 1)
				]))), 128))])]),
				W("div", Ai, [W("div", {
					ref_key: "chartElement",
					ref: a,
					class: "task-history-chart",
					"aria-label": "版本号、耗时和任务事件竖线"
				}, null, 512), n.value ? n.value.summary.total ? U("", !0) : (G(), K("p", Mi, "暂无任务")) : (G(), K("p", ji, H(i.value ? "正在加载任务记录…" : "暂无任务数据"), 1))])
			]),
			W("section", Ni, [
				W("h4", Pi, [o[10] ||= St("任务记录", -1), n.value ? (G(), K("span", Fi, "（" + H(n.value.summary.total) + "）", 1)) : U("", !0)]),
				W("div", Ii, [W("table", Li, [o[11] ||= W("thead", null, [W("tr", null, [
					W("th", null, "编号"),
					W("th", null, "任务类型"),
					W("th", null, "状态"),
					W("th", null, "阶段"),
					W("th", null, "开始时间"),
					W("th", null, "结束时间"),
					W("th", null, "结果"),
					W("th", null, "操作")
				])], -1), W("tbody", null, [(G(!0), K(Ot, null, Tt(n.value?.items ?? [], (t) => (G(), K("tr", { key: t.id }, [
					W("td", null, "#" + H(t.id), 1),
					W("td", null, H(t.kind), 1),
					W("td", null, H(t.state), 1),
					W("td", null, [St(H(t.phase), 1), t.indeterminate ? (G(), K("progress", Ri)) : U("", !0)]),
					W("td", zi, H(t.started), 1),
					W("td", Bi, H(t.finished), 1),
					W("td", null, H(t.message), 1),
					W("td", null, [t.retry ? (G(), K("form", {
						key: 0,
						class: "ui form",
						"data-api": e.retryUrl
					}, [W("button", Hi, [Et(ht, { name: "issueReopened" })])], 8, Vi)) : U("", !0)])
				]))), 128)), n.value?.items.length ? U("", !0) : (G(), K("tr", Ui, [W("td", Wi, H(i.value && !n.value ? "正在加载任务记录…" : "暂无任务"), 1)]))])])]),
				n.value && n.value.pages > 1 ? (G(), K("nav", Gi, [
					W("button", {
						class: "item",
						disabled: i.value || n.value.page <= 1,
						onClick: o[1] ||= (e) => g(n.value.page - 1)
					}, "上一页", 8, Ki),
					W("span", qi, H(n.value.page) + " / " + H(n.value.pages), 1),
					W("button", {
						class: "item",
						disabled: i.value || n.value.page >= n.value.pages,
						onClick: o[2] ||= (e) => g(n.value.page + 1)
					}, "下一页", 8, Ji)
				])) : U("", !0)
			])
		], 8, vi));
	}
}), [["__scopeId", "data-v-769bded6"]]);
//#endregion
export { Yi as default };
