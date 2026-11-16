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
import { F as e, I as t, It as n, J as r, Lt as i, M as a, Rt as o, S as s, T as c, Tt as l, U as u, X as d, Y as f, _ as p, _t as m, a as h, at as g, b as _, bn as v, c as y, cn as b, d as x, dn as S, et as C, f as w, gt as T, hn as E, ht as D, i as ee, in as O, it as k, j as A, jt as j, l as M, lt as te, n as ne, nt as N, o as re, ot as P, p as ie, q as ae, r as oe, rn as F, t as se, u as ce, v as le, xn as ue, xt as I, yt as de } from "./_plugin-vue_export-helper-C97cB3Z4.js";
import { I as L, _ as R, a as fe, c as z, f as pe, g as me, h as B, i as V, l as he, m as H, o as U, p as ge, r as W, t as _e, u as ve, v as G } from "./runtime-core.esm-bundler-B3eh7u8Y.js";
import { n as ye } from "./runtime-dom.esm-bundler-DA27p68u.js";
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/bar/BaseBarSeries.js
var K = function(e) {
	v(t, e);
	function t() {
		var n = e !== null && e.apply(this, arguments) || this;
		return n.type = t.type, n;
	}
	return t.prototype.getInitialData = function(e, t) {
		return s(null, this, { useEncodeDefaulter: !0 });
	}, t.prototype.getMarkerPosition = function(e, t, n) {
		var r = this.coordinateSystem;
		if (r && r.clampData) {
			var i = r.clampData(e), a = r.dataToPoint(i);
			if (n) F(r.getAxes(), function(e, n) {
				if (e.type === "category" && t != null) {
					var r = e.getTicksCoords(), o = e.getTickModel().get("alignWithLabel"), s = i[n], c = t[n] === "x1" || t[n] === "y1";
					if (c && !o && (s += 1), r.length < 2) return;
					if (r.length === 2) {
						a[n] = e.toGlobalCoord(e.getExtent()[c ? 1 : 0]);
						return;
					}
					for (var l = void 0, u = void 0, d = 1, f = 0; f < r.length; f++) {
						var p = r[f].coord, m = f === r.length - 1 ? r[f - 1].tickValue + d : r[f].tickValue;
						if (m === s) {
							u = p;
							break;
						} else if (m < s) l = p;
						else if (l != null && m > s) {
							u = (p + l) / 2;
							break;
						}
						f === 1 && (d = m - r[0].tickValue);
					}
					u ?? (l ? l && (u = r[r.length - 1].coord) : u = r[0].coord), a[n] = e.toGlobalCoord(u);
				}
			});
			else {
				var o = this.getData(), s = o.getLayout("offset"), c = o.getLayout("size"), l = r.getBaseAxis().isHorizontal() ? 0 : 1;
				a[l] += s + c / 2;
			}
			return a;
		}
		return [NaN, NaN];
	}, t.prototype.__requireStartValue = function(e) {
		return this.getBaseAxis() !== e;
	}, t.type = "series.__base_bar__", t.defaultOption = {
		z: 2,
		coordinateSystem: "cartesian2d",
		legendHoverLink: !0,
		barMinHeight: 0,
		barMinAngle: 0,
		large: !1,
		largeThreshold: 400,
		progressive: 3e3,
		progressiveChunkMode: "mod",
		defaultBarGap: "10%"
	}, t;
}(e);
e.registerClass(K);
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/bar/BarSeries.js
var be = function(e) {
	v(n, e);
	function n() {
		var t = e !== null && e.apply(this, arguments) || this;
		return t.type = n.type, t;
	}
	return n.prototype.getInitialData = function() {
		return s(null, this, {
			useEncodeDefaulter: !0,
			createInvertedIndices: !!this.get("realtimeSort", !0) || null
		});
	}, n.prototype.getProgressive = function() {
		return this.get("large") ? this.get("progressive") : !1;
	}, n.prototype.__preparePipelineContext = function(e, t) {
		var n = j(this, e, t);
		return n.progressiveRender && (n.large = !0), n;
	}, n.prototype.brushSelector = function(e, t, n) {
		return n.rect(t.getItemLayout(e));
	}, n.type = "series.bar", n.dependencies = ["grid", "polar"], n.defaultOption = u(K.defaultOption, {
		clip: !0,
		roundCap: !1,
		showBackground: !1,
		backgroundStyle: {
			color: "rgba(180, 180, 180, 0.2)",
			borderColor: null,
			borderWidth: 0,
			borderType: "solid",
			borderRadius: 0,
			shadowBlur: 0,
			shadowColor: null,
			shadowOffsetX: 0,
			shadowOffsetY: 0,
			opacity: 1
		},
		select: { itemStyle: {
			borderColor: t.color.primary,
			borderWidth: 2
		} },
		realtimeSort: !1
	}), n;
}(K), xe = function() {
	function e() {
		this.cx = 0, this.cy = 0, this.r0 = 0, this.r = 0, this.startAngle = 0, this.endAngle = Math.PI * 2, this.clockwise = !0;
	}
	return e;
}(), q = function(e) {
	v(t, e);
	function t(t) {
		var n = e.call(this, t) || this;
		return n.type = "sausage", n;
	}
	return t.prototype.getDefaultShape = function() {
		return new xe();
	}, t.prototype.buildPath = function(e, t) {
		var n = t.cx, r = t.cy, i = Math.max(t.r0 || 0, 0), a = Math.max(t.r, 0), o = (a - i) * .5, s = i + o, c = t.startAngle, l = t.endAngle, u = t.clockwise, d = Math.PI * 2, f = u ? l - c < d : c - l < d;
		f || (c = l - (u ? d : -d));
		var p = Math.cos(c), m = Math.sin(c), h = Math.cos(l), g = Math.sin(l);
		f ? (e.moveTo(p * i + n, m * i + r), e.arc(p * s + n, m * s + r, o, -Math.PI + c, c, !u)) : e.moveTo(p * a + n, m * a + r), e.arc(n, r, a, c, l, !u), e.arc(h * s + n, g * s + r, o, l - Math.PI * 2, l - Math.PI, !u), i !== 0 && e.arc(n, r, i, l, c, u);
	}, t;
}(I);
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/label/sectorLabel.js
function Se(e, t) {
	t ||= {};
	var n = t.isRoundCap;
	return function(t, r, a) {
		var o = r.position;
		if (!o || o instanceof Array) return i(t, r, a);
		var s = e(o), c = r.distance == null ? 5 : r.distance, l = this.shape, u = l.cx, d = l.cy, f = l.r, p = l.r0, m = (f + p) / 2, h = l.startAngle, g = l.endAngle, _ = (h + g) / 2, v = n ? Math.abs(f - p) / 2 : 0, y = Math.cos, b = Math.sin, x = u + f * y(h), S = d + f * b(h), C = "left", w = "top";
		switch (s) {
			case "startArc":
				x = u + (p - c) * y(_), S = d + (p - c) * b(_), C = "center", w = "top";
				break;
			case "insideStartArc":
				x = u + (p + c) * y(_), S = d + (p + c) * b(_), C = "center", w = "bottom";
				break;
			case "startAngle":
				x = u + m * y(h) + J(h, c + v, !1), S = d + m * b(h) + Y(h, c + v, !1), C = "right", w = "middle";
				break;
			case "insideStartAngle":
				x = u + m * y(h) + J(h, -c + v, !1), S = d + m * b(h) + Y(h, -c + v, !1), C = "left", w = "middle";
				break;
			case "middle":
				x = u + m * y(_), S = d + m * b(_), C = "center", w = "middle";
				break;
			case "endArc":
				x = u + (f + c) * y(_), S = d + (f + c) * b(_), C = "center", w = "bottom";
				break;
			case "insideEndArc":
				x = u + (f - c) * y(_), S = d + (f - c) * b(_), C = "center", w = "top";
				break;
			case "endAngle":
				x = u + m * y(g) + J(g, c + v, !0), S = d + m * b(g) + Y(g, c + v, !0), C = "left", w = "middle";
				break;
			case "insideEndAngle":
				x = u + m * y(g) + J(g, -c + v, !0), S = d + m * b(g) + Y(g, -c + v, !0), C = "right", w = "middle";
				break;
			default: return i(t, r, a);
		}
		return t ||= {}, t.x = x, t.y = S, t.align = C, t.verticalAlign = w, t;
	};
}
function Ce(e, t, n, r) {
	if (S(r)) {
		e.setTextConfig({ rotation: r });
		return;
	} else if (b(t)) {
		e.setTextConfig({ rotation: 0 });
		return;
	}
	var i = e.shape, a = i.clockwise ? i.startAngle : i.endAngle, o = i.clockwise ? i.endAngle : i.startAngle, s = (a + o) / 2, c, l = n(t);
	switch (l) {
		case "startArc":
		case "insideStartArc":
		case "middle":
		case "insideEndArc":
		case "endArc":
			c = s;
			break;
		case "startAngle":
		case "insideStartAngle":
			c = a;
			break;
		case "endAngle":
		case "insideEndAngle":
			c = o;
			break;
		default:
			e.setTextConfig({ rotation: 0 });
			return;
	}
	var u = Math.PI * 1.5 - c;
	l === "middle" && u > Math.PI / 2 && u < Math.PI * 1.5 && (u -= Math.PI), e.setTextConfig({ rotation: u });
}
function J(e, t, n) {
	return t * Math.sin(e) * (n ? -1 : 1);
}
function Y(e, t, n) {
	return t * Math.cos(e) * (n ? 1 : -1);
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/helper/sectorHelper.js
function we(e, t, n) {
	var r = e.get("borderRadius");
	if (r == null) return n ? { cornerRadius: 0 } : null;
	b(r) || (r = [
		r,
		r,
		r,
		r
	]);
	var i = Math.abs(t.r || 0 - t.r0 || 0);
	return { cornerRadius: E(r, function(e) {
		return o(e, i);
	}) };
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/bar/BarView.js
var X = Math.max, Z = Math.min, Te = function(e) {
	v(t, e);
	function t() {
		var t = e.call(this) || this;
		return t.type = "bar", t._isFirstFrame = !0, t;
	}
	return t.prototype.render = function(e, t, n, r) {
		this._model = e, this._removeOnRenderedListener(n), this._updateDrawMode(e);
		var i = e.get("coordinateSystem");
		(i === "cartesian2d" || i === "polar") && (this._progressiveEls = null, this._isLargeDraw ? this._renderLarge(e, t, n) : this._renderNormal(e, t, n, r));
	}, t.prototype.incrementalPrepareRender = function(e) {
		this._clear(), this._updateDrawMode(e), this._updateLargeClip(e);
	}, t.prototype.incrementalRender = function(e, t) {
		this._progressiveEls = [], this._incrementalRenderLarge(e, t);
	}, t.prototype.eachRendered = function(e) {
		C(this._progressiveEls || this.group, e);
	}, t.prototype._updateDrawMode = function(e) {
		var t = e.pipelineContext.large;
		(this._isLargeDraw == null || t !== this._isLargeDraw) && (this._isLargeDraw = t, this._clear());
	}, t.prototype._renderNormal = function(e, t, i, a) {
		var o = this.group, s = e.getData(), c = this._data, l = e.coordinateSystem, u = l.getBaseAxis(), d;
		l.type === "cartesian2d" ? d = u.isHorizontal() : l.type === "polar" && (d = u.dim === "angle");
		var f = e.isAnimationEnabled() ? e : null, p = Oe(e, l);
		p && this._enableRealtimeSort(p, s, i);
		var m = e.get("clip", !0) || p, h = l.getArea();
		o.removeClipPath();
		var _ = e.get("roundCap", !0), v = e.get("showBackground", !0), y = e.getModel("backgroundStyle"), b = y.get("borderRadius") || 0, x = [], S = this._backgroundEls, C = a && a.isInitSort, w = a && a.type === "changeAxisOrder";
		function T(e) {
			var t = Q[l.type](s, e);
			if (!t) return null;
			var n = We(l, d, t);
			return n.useStyle(y.getItemStyle()), l.type === "cartesian2d" ? n.setShape("r", b) : n.setShape("cornerRadius", b), x[e] = n, n;
		}
		s.diff(c).add(function(t) {
			var n = s.getItemModel(t), r = Q[l.type](s, t, n);
			if (r && (v && T(t), !(!s.hasValue(t) || !Ne[l.type](r)))) {
				var i = !1;
				m && (i = Ee[l.type](h, r));
				var a = De[l.type](e, s, t, r, d, f, u.model, !1, _);
				p && (a.forceLabelAnimation = !0), Ie(a, s, t, n, r, e, d, l.type === "polar"), C ? a.attr({ shape: r }) : p ? ke(p, f, a, r, t, d, !1, !1) : N(a, { shape: r }, e, t), s.setItemGraphicEl(t, a), o.add(a), a.ignore = i;
			}
		}).update(function(t, n) {
			var i = s.getItemModel(t), a = Q[l.type](s, t, i);
			if (a) {
				if (v) {
					var E = void 0;
					S.length === 0 ? E = T(n) : (E = S[n], E.useStyle(y.getItemStyle()), l.type === "cartesian2d" ? E.setShape("r", b) : E.setShape("cornerRadius", b), x[t] = E);
					var D = Q[l.type](s, t), ee = Ue(d, D, l);
					P(E, { shape: ee }, f, t);
				}
				var O = c.getItemGraphicEl(n);
				if (!s.hasValue(t) || !Ne[l.type](a)) {
					o.remove(O);
					return;
				}
				var A = !1;
				if (m && (A = Ee[l.type](h, a), A && o.remove(O)), O && (O.type === "sector" && _ || O.type === "sausage" && !_) && (O && k(O, e, n), O = null), O ? g(O) : O = De[l.type](e, s, t, a, d, f, u.model, !0, _), p && (O.forceLabelAnimation = !0), w) {
					var j = O.getTextContent();
					if (j) {
						var M = r(j);
						M.prevValue != null && (M.prevValue = M.value);
					}
				} else Ie(O, s, t, i, a, e, d, l.type === "polar");
				C ? O.attr({ shape: a }) : p ? ke(p, f, O, a, t, d, !0, w) : P(O, { shape: a }, e, t, null), s.setItemGraphicEl(t, O), O.ignore = A, o.add(O);
			}
		}).remove(function(t) {
			var n = c.getItemGraphicEl(t);
			n && k(n, e, t);
		}).execute();
		var E = this._backgroundGroup ||= new n();
		E.removeAll();
		for (var D = 0; D < x.length; ++D) E.add(x[D]);
		o.add(E), this._backgroundEls = x, this._data = s;
	}, t.prototype._renderLarge = function(e, t, n) {
		this._clear(), Be(e, this.group), this._updateLargeClip(e);
	}, t.prototype._incrementalRenderLarge = function(e, t) {
		this._removeBackground(), Be(t, this.group, this._progressiveEls, !0);
	}, t.prototype._updateLargeClip = function(e) {
		var t = e.get("clip", !0) && ie(e.coordinateSystem, !1, e), n = this.group;
		t ? n.setClipPath(t) : n.removeClipPath();
	}, t.prototype._enableRealtimeSort = function(e, t, n) {
		var r = this;
		if (t.count()) {
			var i = e.baseAxis;
			if (this._isFirstFrame) this._dispatchInitSort(t, e, n), this._isFirstFrame = !1;
			else {
				var a = function(e) {
					var n = t.getItemGraphicEl(e), r = n && n.shape;
					return r && Math.abs(i.isHorizontal() ? r.height : r.width) || 0;
				};
				this._onRendered = function() {
					r._updateSortWithinSameData(t, a, i, n);
				}, n.getZr().on("rendered", this._onRendered);
			}
		}
	}, t.prototype._dataSort = function(e, t, n) {
		var r = [];
		return e.each(e.mapDimension(t.dim), function(e, t) {
			var i = n(t);
			i ??= NaN, r.push({
				dataIndex: t,
				mappedValue: i,
				ordinalNumber: e
			});
		}), r.sort(function(e, t) {
			return t.mappedValue - e.mappedValue;
		}), { ordinalNumbers: E(r, function(e) {
			return e.ordinalNumber;
		}) };
	}, t.prototype._isOrderChangedWithinSameData = function(e, t, n) {
		for (var r = n.scale, i = e.mapDimension(n.dim), a = Number.MAX_VALUE, o = 0, s = r.getOrdinalMeta().categories.length; o < s; ++o) {
			var c = e.rawIndexOf(i, r.getRawOrdinalNumber(o)), l = c < 0 ? Number.MIN_VALUE : t(e.indexOfRawIndex(c));
			if (l > a) return !0;
			a = l;
		}
		return !1;
	}, t.prototype._isOrderDifferentInView = function(e, t) {
		for (var n = t.scale, r = n.getExtent(), i = Math.max(0, r[0]), a = Math.min(r[1], n.getOrdinalMeta().categories.length - 1); i <= a; ++i) if (e.ordinalNumbers[i] !== n.getRawOrdinalNumber(i)) return !0;
	}, t.prototype._updateSortWithinSameData = function(e, t, n, r) {
		if (this._isOrderChangedWithinSameData(e, t, n)) {
			var i = this._dataSort(e, n, t);
			this._isOrderDifferentInView(i, n) && (this._removeOnRenderedListener(r), r.dispatchAction({
				type: "changeAxisOrder",
				componentType: n.dim + "Axis",
				axisId: n.index,
				sortInfo: i
			}));
		}
	}, t.prototype._dispatchInitSort = function(e, t, n) {
		var r = t.baseAxis, i = this._dataSort(e, r, function(n) {
			return e.get(e.mapDimension(t.otherAxis.dim), n);
		});
		n.dispatchAction({
			type: "changeAxisOrder",
			componentType: r.dim + "Axis",
			isInitSort: !0,
			axisId: r.index,
			sortInfo: i
		});
	}, t.prototype.remove = function(e, t) {
		this._clear(this._model), this._removeOnRenderedListener(t);
	}, t.prototype.dispose = function(e, t) {
		this._removeOnRenderedListener(t);
	}, t.prototype._removeOnRenderedListener = function(e) {
		this._onRendered &&= (e.getZr().off("rendered", this._onRendered), null);
	}, t.prototype._clear = function(e) {
		var t = this.group, n = this._data;
		e && e.isAnimationEnabled() && n && !this._isLargeDraw ? (this._removeBackground(), this._backgroundEls = [], n.eachItemGraphicEl(function(t) {
			k(t, e, m(t).dataIndex);
		})) : t.removeAll(), this._data = null, this._isFirstFrame = !0;
	}, t.prototype._removeBackground = function() {
		this.group.remove(this._backgroundGroup), this._backgroundGroup = null;
	}, t.type = "bar", t;
}(a), Ee = {
	cartesian2d: function(e, t) {
		var n = t.width < 0 ? -1 : 1, r = t.height < 0 ? -1 : 1;
		n < 0 && (t.x += t.width, t.width = -t.width), r < 0 && (t.y += t.height, t.height = -t.height);
		var i = e.x + e.width, a = e.y + e.height, o = X(t.x, e.x), s = Z(t.x + t.width, i), c = X(t.y, e.y), l = Z(t.y + t.height, a), u = s < o, d = l < c;
		return t.x = u && o > i ? s : o, t.y = d && c > a ? l : c, t.width = u ? 0 : s - o, t.height = d ? 0 : l - c, n < 0 && (t.x += t.width, t.width = -t.width), r < 0 && (t.y += t.height, t.height = -t.height), u || d;
	},
	polar: function(e, t) {
		var n = t.r0 <= t.r ? 1 : -1;
		if (n < 0) {
			var r = t.r;
			t.r = t.r0, t.r0 = r;
		}
		var i = Z(t.r, e.r), a = X(t.r0, e.r0);
		t.r = i, t.r0 = a;
		var o = i - a < 0;
		if (n < 0) {
			var r = t.r;
			t.r = t.r0, t.r0 = r;
		}
		return o;
	}
}, De = {
	cartesian2d: function(e, t, n, r, i, a, o, s, c) {
		var l = new de({
			shape: O({}, r),
			z2: 1
		});
		if (l.__dataIndex = n, l.name = "item", a) {
			var u = l.shape, d = i ? "height" : "width";
			u[d] = 0;
		}
		return l;
	},
	polar: function(e, t, n, r, i, a, o, s, c) {
		var l = !i && c ? q : te, u = new l({
			shape: r,
			z2: 1
		});
		if (u.name = "item", u.calculateTextPosition = Se(Fe(i), { isRoundCap: l === q }), a) {
			var d = u.shape, f = i ? "r" : "endAngle", p = {};
			d[f] = i ? r.r0 : r.startAngle, p[f] = r[f], (s ? P : N)(u, { shape: p }, a);
		}
		return u;
	}
};
function Oe(e, t) {
	var n = e.get("realtimeSort", !0), r = t.getBaseAxis();
	if (n && r.type === "category" && t.type === "cartesian2d") return {
		baseAxis: r,
		otherAxis: t.getOtherAxis(r)
	};
}
function ke(e, t, n, r, i, a, o, s) {
	var c, l;
	a ? (l = {
		x: r.x,
		width: r.width
	}, c = {
		y: r.y,
		height: r.height
	}) : (l = {
		y: r.y,
		height: r.height
	}, c = {
		x: r.x,
		width: r.width
	}), s || (o ? P : N)(n, { shape: c }, t, i, null);
	var u = t ? e.baseAxis.model : null;
	(o ? P : N)(n, { shape: l }, u, i);
}
function Ae(e, t) {
	for (var n = 0; n < t.length; n++) if (!isFinite(e[t[n]])) return !0;
	return !1;
}
var je = [
	"x",
	"y",
	"width",
	"height"
], Me = [
	"cx",
	"cy",
	"r",
	"startAngle",
	"endAngle"
], Ne = {
	cartesian2d: function(e) {
		return !Ae(e, je);
	},
	polar: function(e) {
		return !Ae(e, Me);
	}
}, Q = {
	cartesian2d: function(e, t, n) {
		var r = e.getItemLayout(t);
		if (!r) return null;
		var i = n ? Le(n, r) : 0, a = r.width > 0 ? 1 : -1, o = r.height > 0 ? 1 : -1;
		return {
			x: r.x + a * i / 2,
			y: r.y + o * i / 2,
			width: r.width - a * i,
			height: r.height - o * i
		};
	},
	polar: function(e, t, n) {
		var r = e.getItemLayout(t);
		return {
			cx: r.cx,
			cy: r.cy,
			r0: r.r0,
			r: r.r,
			startAngle: r.startAngle,
			endAngle: r.endAngle,
			clockwise: r.clockwise
		};
	}
};
function Pe(e) {
	return e.startAngle != null && e.endAngle != null && e.startAngle === e.endAngle;
}
function Fe(e) {
	return function(e) {
		var t = e ? "Arc" : "Angle";
		return function(e) {
			switch (e) {
				case "start":
				case "insideStart":
				case "end":
				case "insideEnd": return e + t;
				default: return e;
			}
		};
	}(e);
}
function Ie(e, t, n, r, i, a, o, s) {
	var c = t.getItemVisual(n, "style");
	if (!s) {
		var l = r.get(["itemStyle", "borderRadius"]) || 0;
		e.setShape("r", l);
	} else if (!a.get("roundCap")) {
		var u = e.shape;
		O(u, we(r.getModel("itemStyle"), u, !0)), e.setShape(u);
	}
	e.useStyle(c);
	var m = r.getShallow("cursor");
	m && e.attr("cursor", m);
	var h = s ? o ? i.r >= i.r0 ? "endArc" : "startArc" : i.endAngle >= i.startAngle ? "endAngle" : "startAngle" : o ? Ge(i, a.coordinateSystem) : Ke(i, a.coordinateSystem), g = ae(r);
	f(e, g, {
		labelFetcher: a,
		labelDataIndex: n,
		defaultText: le(a.getData(), n),
		inheritColor: c.fill,
		defaultOpacity: c.opacity,
		defaultOutsidePosition: h
	});
	var _ = e.getTextContent();
	if (s && _) {
		var v = r.get(["label", "position"]);
		e.textConfig.inside = v === "middle" ? !0 : null, Ce(e, v === "outside" ? h : v, Fe(o), r.get(["label", "rotate"]));
	}
	d(_, g, a.getRawValue(n), function(e) {
		return p(t, e);
	});
	var y = r.getModel(["emphasis"]);
	T(e, y.get("focus"), y.get("blurScope"), y.get("disabled")), D(e, r), Pe(i) && (e.style.fill = "none", e.style.stroke = "none", F(e.states, function(e) {
		e.style && (e.style.fill = e.style.stroke = "none");
	}));
}
function Le(e, t) {
	var n = e.get(["itemStyle", "borderColor"]);
	if (!n || n === "none") return 0;
	var r = e.get(["itemStyle", "borderWidth"]) || 0, i = isNaN(t.width) ? Number.MAX_VALUE : Math.abs(t.width), a = isNaN(t.height) ? Number.MAX_VALUE : Math.abs(t.height);
	return Math.min(r, i, a);
}
var Re = function() {
	function e() {}
	return e;
}(), ze = function(e) {
	v(t, e);
	function t(t) {
		var n = e.call(this, t) || this;
		return n.type = "largeBar", n;
	}
	return t.prototype.getDefaultShape = function() {
		return new Re();
	}, t.prototype.buildPath = function(e, t) {
		for (var n = t.points, r = this.baseDimIdx, i = 1 - this.baseDimIdx, a = [], o = [], s = this.barWidth, c = 0; c < n.length; c += 3) o[r] = s, o[i] = n[c + 2], a[r] = n[c + r], a[i] = n[c + i], e.rect(a[0], a[1], o[0], o[1]);
	}, t;
}(I);
function Be(e, t, n, r) {
	var i = e.getData(), a = i.getLayout("valueAxisHorizontal") ? 1 : 0, o = i.getLayout("largeDataIndices"), s = i.getLayout("size"), c = e.getModel("backgroundStyle"), u = i.getLayout("largeBackgroundPoints"), d = r ? l(e) : 0;
	if (u) {
		var f = new ze({
			shape: { points: u },
			incremental: d,
			silent: !0,
			z2: 0
		});
		f.baseDimIdx = a, f.largeDataIndices = o, f.barWidth = s, f.useStyle(c.getItemStyle()), t.add(f), n && n.push(f);
	}
	var p = new ze({
		shape: { points: i.getLayout("largePoints") },
		incremental: d,
		ignoreCoarsePointer: !0,
		z2: 1
	});
	p.baseDimIdx = a, p.largeDataIndices = o, p.barWidth = s, t.add(p), p.useStyle(i.getVisual("style")), p.style.stroke = null, m(p).seriesIndex = e.seriesIndex, e.get("silent") || (p.on("mousedown", Ve), p.on("mousemove", Ve)), n && n.push(p);
}
var Ve = A(function(e) {
	var t = this, n = He(t, e.offsetX, e.offsetY);
	m(t).dataIndex = n >= 0 ? n : null;
}, 30, !1);
function He(e, t, n) {
	for (var r = e.baseDimIdx, i = 1 - r, a = e.shape.points, o = e.largeDataIndices, s = [], c = [], l = e.barWidth, u = 0, d = a.length / 3; u < d; u++) {
		var f = u * 3;
		if (c[r] = l, c[i] = a[f + 2], s[r] = a[f + r], s[i] = a[f + i], c[i] < 0 && (s[i] += c[i], c[i] = -c[i]), t >= s[0] && t <= s[0] + c[0] && n >= s[1] && n <= s[1] + c[1]) return o[u];
	}
	return -1;
}
function Ue(e, t, n) {
	if (w(n, "cartesian2d")) {
		var r = t, i = n.getArea();
		return {
			x: e ? r.x : i.x,
			y: e ? i.y : r.y,
			width: e ? r.width : i.width,
			height: e ? i.height : r.height
		};
	} else {
		var i = n.getArea(), a = t;
		return {
			cx: i.cx,
			cy: i.cy,
			r0: e ? i.r0 : a.r0,
			r: e ? i.r : a.r,
			startAngle: e ? a.startAngle : 0,
			endAngle: e ? a.endAngle : Math.PI * 2
		};
	}
}
function We(e, t, n) {
	return new (e.type === "polar" ? te : de)({
		shape: Ue(t, n, e),
		silent: !0,
		z2: 0
	});
}
function Ge(e, t) {
	return e.height === 0 ? t.getOtherAxis(t.getBaseAxis()).inverse ? "bottom" : "top" : e.height > 0 ? "bottom" : "top";
}
function Ke(e, t) {
	return e.width === 0 ? t.getOtherAxis(t.getBaseAxis()).inverse ? "left" : "right" : e.width >= 0 ? "right" : "left";
}
//#endregion
//#region node_modules/.pnpm/echarts@6.1.0/node_modules/echarts/lib/chart/bar/install.js
function qe(e) {
	e.registerChartView(Te), e.registerSeriesModel(be), e.registerLayout(e.PRIORITY.VISUAL.LAYOUT, y("bar")), e.registerLayout(e.PRIORITY.VISUAL.PROGRESSIVE_LAYOUT, M("bar")), e.registerProcessor(e.PRIORITY.PROCESSOR.STATISTIC, x("bar")), e.registerAction({
		type: "changeAxisOrder",
		event: "changeAxisOrder",
		update: "update"
	}, function(e, t) {
		var n = e.componentType || "series";
		t.eachComponent({
			mainType: n,
			query: e
		}, function(t) {
			e.sortInfo && t.axis.setCategorySortInfo(e.sortInfo);
		});
	}), ce(e);
}
//#endregion
//#region web_src/features/activity-trend-options.ts
var $ = [
	{
		key: "queued",
		label: "排队中",
		color: "--color-grey"
	},
	{
		key: "running",
		label: "进行中",
		color: "--color-blue"
	},
	{
		key: "succeeded",
		label: "成功",
		color: "--color-green"
	},
	{
		key: "failed",
		label: "失败",
		color: "--color-red"
	},
	{
		key: "interrupted",
		label: "已中断",
		color: "--color-orange"
	}
];
function Je(e, t, n) {
	let r = $.filter((e) => n[e.label] !== !1);
	return [
		e.date,
		t,
		...r.map((t) => `${t.label}  ${(e.counts[t.key] ?? 0).toLocaleString("zh-CN")} 次`),
		`合计  ${r.reduce((t, n) => t + (e.counts[n.key] ?? 0), 0).toLocaleString("zh-CN")} 次`
	].join("\n");
}
function Ye(e, t, n, r, i) {
	return {
		animation: !1,
		textStyle: {
			color: i("--color-text"),
			fontFamily: "Microsoft YaHei, sans-serif"
		},
		color: $.map((e) => i(e.color)),
		legend: {
			top: 0,
			left: "center",
			itemWidth: 12,
			itemHeight: 12,
			itemGap: 12,
			selected: n,
			textStyle: {
				color: i("--color-text"),
				fontSize: 12
			}
		},
		grid: {
			top: 64,
			left: 80,
			right: 80,
			bottom: 76,
			outerBoundsMode: "same",
			outerBoundsContain: "axisLabel"
		},
		xAxis: {
			type: "category",
			data: e.map((e) => e.date),
			axisLabel: {
				formatter: (e) => e.slice(5),
				hideOverlap: !0
			},
			axisTick: { alignWithLabel: !0 },
			axisLine: { lineStyle: { color: i("--color-secondary") } }
		},
		yAxis: {
			type: "value",
			min: 0,
			minInterval: 1,
			splitLine: { lineStyle: { color: i("--color-secondary") } }
		},
		tooltip: {
			trigger: "axis",
			renderMode: "richText",
			confine: !0,
			axisPointer: { type: "shadow" },
			backgroundColor: i("--color-box-body"),
			borderColor: i("--color-secondary"),
			textStyle: {
				color: i("--color-text"),
				fontSize: 12
			},
			formatter: (r) => {
				let i = Array.isArray(r) ? r[0] : r, a = i ? e[i.dataIndex] : void 0;
				return a ? Je(a, t, n) : "";
			}
		},
		dataZoom: [{
			id: "dates",
			type: "slider",
			left: 80,
			right: 80,
			xAxisIndex: 0,
			start: r.start,
			end: r.end,
			bottom: 8,
			height: 24,
			brushSelect: !1,
			showDetail: !0,
			textStyle: { color: i("--color-text") },
			borderColor: i("--color-secondary"),
			fillerColor: i("--color-primary-alpha-10"),
			handleStyle: { color: i("--color-primary") },
			moveHandleSize: 0
		}],
		series: $.map((t) => ({
			id: t.key,
			name: t.label,
			type: "bar",
			stack: "activity",
			barMaxWidth: 36,
			data: e.map((e) => e.counts[t.key] ?? 0)
		}))
	};
}
//#endregion
//#region web_src/components/ActivityTrend.vue?vue&type=script&setup=true&lang.ts
var Xe = {
	class: "activity-trend",
	"aria-label": "日活动趋势"
}, Ze = { class: "trend-visible-filters" }, Qe = { "aria-label": "活动类型" }, $e = { class: "trend-filter-options" }, et = ["value"], tt = { "aria-label": "时间范围" }, nt = { class: "trend-filter-options" }, rt = ["value"], it = { class: "trend-content" }, at = { class: "muted" }, ot = {
	key: 0,
	role: "alert"
}, st = ["disabled"], ct = ["aria-busy"], lt = {
	key: 0,
	class: "trend-chart-message",
	role: "status"
}, ut = {
	key: 1,
	class: "trend-chart-message",
	role: "status"
}, dt = /* @__PURE__ */ se(/* @__PURE__ */ ve({
	__name: "ActivityTrend",
	props: { org: {} },
	setup(e) {
		_([
			qe,
			re,
			ee,
			h,
			oe,
			ne
		]);
		let t = e, n = [
			{
				key: "local",
				label: "本地仓库迁移"
			},
			{
				key: "mirror",
				label: "镜像仓库迁移"
			},
			{
				key: "sync",
				label: "镜像仓库同步"
			}
		], r = G(new Intl.DateTimeFormat("sv-SE", {
			timeZone: "Asia/Shanghai",
			year: "numeric",
			month: "2-digit",
			day: "2-digit"
		}).format(/* @__PURE__ */ new Date()));
		function i(e, t) {
			let n = /* @__PURE__ */ new Date(`${e}T00:00:00Z`);
			return n.setUTCDate(n.getUTCDate() + t), n.toISOString().slice(0, 10);
		}
		let a = G("sync"), o = G("30"), s = W(() => r.value), l = W(() => i(r.value, 1 - Number(o.value))), u = G([]), d = G(!1), f = G(""), p = G(!1), m = G(), g = G({}), v = W(() => $.every((e) => g.value[e.label] === !1)), y = W(() => n.find((e) => e.key === a.value)?.label ?? ""), b, x, S = {
			start: 0,
			end: 100
		}, C, w, T = 0, E = !1;
		function D() {
			if (!b || !m.value) return;
			let e = getComputedStyle(m.value);
			b.setOption(Ye(u.value, y.value, g.value, S, (t) => e.getPropertyValue(t).trim()));
		}
		async function O() {
			let e = ++T;
			C?.abort(), w && clearTimeout(w), C = new AbortController(), d.value = !0, f.value = "";
			try {
				let n = new URLSearchParams({
					org: t.org,
					kind: a.value,
					from: l.value,
					to: s.value
				}), i = await fetch(`/api/v1/activity/trend?${n}`, { signal: C.signal });
				if (i.status === 401) {
					location.assign("/user/login");
					return;
				}
				if (!i.ok) throw Error();
				let o = await i.json();
				if (e !== T || E) return;
				u.value = o.days, p.value = !0, r.value = o.today, D(), w = setTimeout(O, o.pending ? 5e3 : 15e3);
			} catch {
				if (e !== T || E) return;
				f.value = "活动趋势暂不可用，请重试", w = setTimeout(O, 15e3);
			} finally {
				e === T && (d.value = !1);
			}
		}
		return me([
			a,
			l,
			s
		], () => {
			S = {
				start: 0,
				end: 100
			}, p.value = !1, u.value = [], D(), O();
		}), ge(() => {
			m.value && (b = c(m.value, void 0, { renderer: "svg" }), b.on("legendselectchanged", (e) => {
				g.value = e.selected, D();
			}), b.on("datazoom", (e) => {
				let t = e;
				t.start !== void 0 && t.end !== void 0 && (S = {
					start: t.start,
					end: t.end
				});
			}), x = new ResizeObserver(() => b?.resize()), x.observe(m.value), D(), O());
		}), pe(() => {
			E = !0, T++, C?.abort(), w && clearTimeout(w), x?.disconnect(), b?.dispose(), b = void 0;
		}), (e, t) => (H(), U("section", Xe, [
			t[3] ||= V("div", { class: "trend-heading" }, [V("h3", null, "日活动趋势")], -1),
			V("div", Ze, [V("fieldset", Qe, [V("div", $e, [(H(), U(_e, null, B(n, (e) => V("label", { key: e.key }, [R(V("input", {
				"onUpdate:modelValue": t[0] ||= (e) => a.value = e,
				type: "radio",
				name: "trend-kind",
				value: e.key
			}, null, 8, et), [[ye, a.value]]), z(L(e.label), 1)])), 64))])]), V("fieldset", tt, [V("div", nt, [(H(), U(_e, null, B([
				{
					key: "7",
					label: "近7天"
				},
				{
					key: "30",
					label: "近30天"
				},
				{
					key: "90",
					label: "近90天"
				}
			], (e) => V("label", { key: e.key }, [R(V("input", {
				"onUpdate:modelValue": t[1] ||= (e) => o.value = e,
				type: "radio",
				name: "trend-period",
				value: e.key
			}, null, 8, rt), [[ye, o.value]]), z(L(e.label), 1)])), 64))])])]),
			V("div", it, [
				V("span", at, L(y.value) + " · " + L(l.value) + " 至 " + L(s.value), 1),
				f.value ? (H(), U("p", ot, [z(L(f.value) + " ", 1), V("button", {
					type: "button",
					class: "ui basic button action-text-button",
					disabled: d.value,
					onClick: O
				}, [he(ue, { name: "issueReopened" }), t[2] ||= V("span", null, "重试", -1)], 8, st)])) : fe("", !0),
				V("div", {
					class: "trend-chart-wrapper",
					"aria-busy": d.value
				}, [V("div", {
					ref_key: "chartElement",
					ref: m,
					class: "trend-chart",
					"aria-label": "按状态堆积的每日活动次数"
				}, null, 512), d.value && !p.value ? (H(), U("p", lt, "正在加载活动趋势…")) : p.value && v.value ? (H(), U("p", ut, "请选择至少一个状态")) : fe("", !0)], 8, ct)
			])
		]));
	}
}), [["__scopeId", "data-v-ced31f2a"]]);
//#endregion
export { dt as default };
