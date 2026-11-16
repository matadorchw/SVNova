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
//#region web_src/upstream/fomantic/dropdown.js
(function(e, t, n, r) {
	e.isFunction = e.isFunction || function(e) {
		return typeof e == "function" && typeof e.nodeType != "number";
	}, t = t !== void 0 && t.Math == Math ? t : typeof self < "u" && self.Math == Math ? self : Function("return this")(), e.fn.dropdown = function(i) {
		var a = e(this), o = e(n), s = a.selector || "", c = "ontouchstart" in n.documentElement, l = "click", u = (/* @__PURE__ */ new Date()).getTime(), d = [], f = arguments[0], p = typeof f == "string", m = [].slice.call(arguments, 1), h;
		return a.each(function(ee) {
			var g = e.isPlainObject(i) ? e.extend(!0, {}, e.fn.dropdown.settings, i) : e.extend({}, e.fn.dropdown.settings), _ = g.className, v = g.message, y = g.fields, b = g.keys, x = g.metadata, te = g.namespace, S = g.regExp, C = g.selector, w = g.error, T = g.templates, E = "." + te, D = "module-" + te, O = e(this), k = typeof g.context == "string" ? e(n).find(g.context) : e(g.context), A = O.find(C.text), j = O.find(C.search), M = O.find(C.sizer), N = O.find(C.input), ne = O.find(C.icon), re = O.find(C.clearIcon), P = O.prev().find(C.text).length > 0 ? O.prev().find(C.text) : O.prev(), F = O.children(C.menu), I = F.find(C.item), L = g.hideDividers ? I.parent().children(C.divider) : e(), R = !1, z = !1, B = !1, V = !1, H = this, U = O.data(D), W, G, K, q, J, Y, X, Z, Q, $ = {
				initialize: function() {
					$.debug("Initializing dropdown", g), $.is.alreadySetup() ? $.setup.reference() : (g.ignoreDiacritics && !String.prototype.normalize && (g.ignoreDiacritics = !1, $.error(w.noNormalize, H)), $.setup.layout(), g.values && ($.set.initialLoad(), $.change.values(g.values), $.remove.initialLoad()), $.refreshData(), $.save.defaults(), $.restore.selected(), $.create.id(), $.bind.events(), $.observeChanges(), $.instantiate());
				},
				instantiate: function() {
					$.verbose("Storing instance of dropdown", $), U = $, O.data(D, $);
				},
				destroy: function() {
					$.verbose("Destroying previous dropdown", O), $.remove.tabbable(), $.remove.active(), F.transition("stop all"), F.removeClass(_.visible).addClass(_.hidden), O.off(E).removeData(D), F.off(E), o.off(J), $.disconnect.menuObserver(), $.disconnect.selectObserver(), $.disconnect.classObserver();
				},
				observeChanges: function() {
					"MutationObserver" in t && (X = new MutationObserver($.event.select.mutation), Z = new MutationObserver($.event.menu.mutation), Q = new MutationObserver($.event.class.mutation), $.debug("Setting up mutation observer", X, Z, Q), $.observe.select(), $.observe.menu(), $.observe.class());
				},
				disconnect: {
					menuObserver: function() {
						Z && Z.disconnect();
					},
					selectObserver: function() {
						X && X.disconnect();
					},
					classObserver: function() {
						Q && Q.disconnect();
					}
				},
				observe: {
					select: function() {
						$.has.input() && X && X.observe(O[0], {
							childList: !0,
							subtree: !0
						});
					},
					menu: function() {
						$.has.menu() && Z && Z.observe(F[0], {
							childList: !0,
							subtree: !0
						});
					},
					class: function() {
						$.has.search() && Q && Q.observe(O[0], { attributes: !0 });
					}
				},
				create: {
					id: function() {
						Y = (Math.random().toString(16) + "000000000").substr(2, 8), J = "." + Y, $.verbose("Creating unique id for element", Y);
					},
					userChoice: function(t) {
						var n, i, a;
						return t ||= $.get.userValues(), t ? (t = Array.isArray(t) ? t : [t], e.each(t, function(t, o) {
							$.get.item(o) === !1 && (a = g.templates.addition($.add.variables(v.addResult, o)), i = e("<div />").html(a).attr("data-" + x.value, o).attr("data-" + x.text, o).addClass(_.addition).addClass(_.item), g.hideAdditions && i.addClass(_.hidden), n = n === r ? i : n.add(i), $.verbose("Creating user choices for value", o, i));
						}), n) : !1;
					},
					userLabels: function(t) {
						var n = $.get.userValues();
						n && ($.debug("Adding user labels", n), e.each(n, function(e, t) {
							$.verbose("Adding custom user value"), $.add.label(t, t);
						}));
					},
					menu: function() {
						F = e("<div />").addClass(_.menu).appendTo(O);
					},
					sizer: function() {
						M = e("<span />").addClass(_.sizer).insertAfter(j);
					}
				},
				search: function(e) {
					e = e === r ? $.get.query() : e, $.verbose("Searching for query", e), $.has.minCharacters(e) ? $.filter(e) : $.hide(null, !0);
				},
				select: {
					firstUnfiltered: function() {
						$.verbose("Selecting first non-filtered element"), $.remove.selectedItem();
						let e = I.not(C.unselectable).not(C.addition + C.hidden).filter(`[data-value="${CSS.escape(N.val())}"]`);
						e.length || (e = I.eq(0)), e.addClass(_.selected);
					},
					nextAvailable: function(e) {
						e = e.eq(0);
						var t = e.nextAll(C.item).not(C.unselectable).eq(0), n = e.prevAll(C.item).not(C.unselectable).eq(0);
						t.length > 0 ? ($.verbose("Moving selection to", t), t.addClass(_.selected)) : ($.verbose("Moving selection to", n), n.addClass(_.selected));
					}
				},
				setup: {
					api: function() {
						var e = {
							debug: g.debug,
							urlData: {
								value: $.get.value(),
								query: $.get.query()
							},
							on: !1
						};
						$.verbose("First request, initializing API"), O.api(e);
					},
					layout: function() {
						O.is("select") && ($.setup.select(), $.setup.returnedObject()), $.has.menu() || $.create.menu(), $.is.selection() && $.is.clearable() && !$.has.clearItem() && ($.verbose("Adding clear icon"), re = e("<i />").addClass("remove icon").insertBefore(A)), $.is.search() && !$.has.search() && ($.verbose("Adding search input"), j = e("<input />").addClass(_.search).prop("autocomplete", "off").insertBefore(A)), $.is.multiple() && $.is.searchSelection() && !$.has.sizer() && $.create.sizer(), g.allowTab && $.set.tabbable();
					},
					select: function() {
						var t = $.get.selectValues();
						$.debug("Dropdown initialized on a select", t), O.is("select") && (N = O), N.parent(C.dropdown).length > 0 ? ($.debug("UI dropdown already exists. Creating dropdown menu only"), O = N.closest(C.dropdown), $.has.menu() || $.create.menu(), F = O.children(C.menu), $.setup.menu(t)) : ($.debug("Creating entire dropdown from select"), O = e("<div />").attr("class", N.attr("class")).addClass(_.selection).addClass(_.dropdown).html(T.dropdown(t, y, g.preserveHTML, g.className)).insertBefore(N), O.attr("data-tooltip-content", N.attr("data-tooltip-content") ?? null), N.hasClass(_.multiple) && N.prop("multiple") === !1 && ($.error(w.missingMultiple), N.prop("multiple", !0)), N.is("[multiple]") && $.set.multiple(), N.prop("disabled") && ($.debug("Disabling dropdown"), O.addClass(_.disabled)), N.removeAttr("required").removeAttr("class").detach().prependTo(O)), $.refresh();
					},
					menu: function(t) {
						F.html(T.menu(t, y, g.preserveHTML, g.className)), I = F.find(C.item), L = g.hideDividers ? I.parent().children(C.divider) : e();
					},
					reference: function() {
						$.debug("Dropdown behavior was called on select, replacing with closest dropdown"), O = O.parent(C.dropdown), U = O.data(D), H = O.get(0), $.refresh(), $.setup.returnedObject();
					},
					returnedObject: function() {
						var e = a.slice(0, ee), t = a.slice(ee + 1);
						a = e.add(O).add(t);
					}
				},
				refresh: function() {
					$.refreshSelectors(), $.refreshData();
				},
				refreshItems: function() {
					I = F.find(C.item), L = g.hideDividers ? I.parent().children(C.divider) : e();
				},
				refreshSelectors: function() {
					$.verbose("Refreshing selector cache"), A = O.find(C.text), j = O.find(C.search), N = O.find(C.input), ne = O.find(C.icon), P = O.prev().find(C.text).length > 0 ? O.prev().find(C.text) : O.prev(), F = O.children(C.menu), I = F.find(C.item), L = g.hideDividers ? I.parent().children(C.divider) : e();
				},
				refreshData: function() {
					$.verbose("Refreshing cached metadata"), I.removeData(x.text).removeData(x.value);
				},
				clearData: function() {
					$.verbose("Clearing metadata"), I.removeData(x.text).removeData(x.value), O.removeData(x.defaultText).removeData(x.defaultValue).removeData(x.placeholderText);
				},
				toggle: function() {
					$.verbose("Toggling menu visibility"), $.is.active() ? $.hide() : $.show();
				},
				show: function(t, n) {
					if (t = e.isFunction(t) ? t : function() {}, !$.can.show() && $.is.remote() && ($.debug("No API results retrieved, searching before show"), $.queryRemote($.get.query(), $.show)), $.can.show() && !$.is.active()) {
						if ($.debug("Showing dropdown"), $.has.message() && !($.has.maxSelections() || $.has.allResultsFiltered()) && $.remove.message(), $.is.allFiltered()) return !0;
						g.onShow.call(H) !== !1 && (O.fomanticExt.onDropdownAfterFiltered.call(H), $.animate.show(function() {
							$.can.click() && $.bind.intent(), $.has.search() && !n && $.focusSearch(), $.set.visible(), t.call(H);
						}));
					}
				},
				hide: function(t, n) {
					t = e.isFunction(t) ? t : function() {}, $.is.active() && !$.is.animatingOutward() ? ($.debug("Hiding dropdown"), g.onHide.call(H) !== !1 && $.animate.hide(function() {
						$.remove.visible(), $.is.focusedOnSearch() && n !== !0 && j.blur(), t.call(H);
					})) : $.can.click() && $.unbind.intent(), V = !1;
				},
				hideOthers: function() {
					$.verbose("Finding other dropdowns to hide"), a.not(O).has(C.menu + "." + _.visible).dropdown("hide");
				},
				hideMenu: function() {
					$.verbose("Hiding menu  instantaneously"), $.remove.active(), $.remove.visible(), F.transition("hide");
				},
				hideSubMenus: function() {
					var e = F.children(C.item).find(C.menu);
					$.verbose("Hiding sub menus", e), e.transition("hide");
				},
				bind: {
					events: function() {
						$.bind.keyboardEvents(), $.bind.inputEvents(), $.bind.mouseEvents();
					},
					keyboardEvents: function() {
						$.verbose("Binding keyboard events"), O.on("keydown" + E, $.event.keydown), $.has.search() && O.on($.get.inputEvent() + E, C.search, $.event.input), $.is.multiple() && o.on("keydown" + J, $.event.document.keydown);
					},
					inputEvents: function() {
						$.verbose("Binding input change events"), O.on("change" + E, C.input, $.event.change);
					},
					mouseEvents: function() {
						$.verbose("Binding mouse events"), $.is.multiple() && O.on(l + E, C.label, $.event.label.click).on(l + E, C.remove, $.event.remove.click), $.is.searchSelection() ? (O.on("mousedown" + E, $.event.mousedown).on("mouseup" + E, $.event.mouseup).on("mousedown" + E, C.menu, $.event.menu.mousedown).on("mouseup" + E, C.menu, $.event.menu.mouseup).on(l + E, C.icon, $.event.icon.click).on(l + E, C.clearIcon, $.event.clearIcon.click).on("focus" + E, C.search, $.event.search.focus).on(l + E, C.search, $.event.search.focus).on("blur" + E, C.search, $.event.search.blur).on(l + E, C.text, $.event.text.focus), $.is.multiple() && O.on(l + E, $.event.click)) : (g.on == "click" ? O.on(l + E, C.icon, $.event.icon.click).on(l + E, $.event.test.toggle) : g.on == "hover" ? O.on("mouseenter" + E, $.delay.show).on("mouseleave" + E, $.delay.hide) : O.on(g.on + E, $.toggle), O.on("mousedown" + E, $.event.mousedown).on("mouseup" + E, $.event.mouseup).on("focus" + E, $.event.focus).on(l + E, C.clearIcon, $.event.clearIcon.click), $.has.menuSearch() ? O.on("blur" + E, C.search, $.event.search.blur) : O.on("blur" + E, $.event.blur)), F.on((c ? "touchstart" : "mouseenter") + E, C.item, $.event.item.mouseenter).on("mouseleave" + E, C.item, $.event.item.mouseleave).on("click" + E, C.item, $.event.item.click);
					},
					intent: function() {
						$.verbose("Binding hide intent event to document"), c && o.on("touchstart" + J, $.event.test.touch).on("touchmove" + J, $.event.test.touch), o.on(l + J, $.event.test.hide);
					}
				},
				unbind: { intent: function() {
					$.verbose("Removing hide intent event from document"), c && o.off("touchstart" + J).off("touchmove" + J), o.off(l + J);
				} },
				filter: function(t) {
					var n = t === r ? $.get.query() : t, i = function() {
						$.is.multiple() && $.filterActive(), (t || !t && $.get.activeItem().length == 0) && $.select.firstUnfiltered(), $.has.allResultsFiltered() ? g.onNoResults.call(H, n) ? g.allowAdditions ? g.hideAdditions && ($.verbose("User addition with no menu, setting empty style"), $.set.empty(), $.hideMenu()) : ($.verbose("All items filtered, showing message", n), $.add.message(v.noResults)) : ($.verbose("All items filtered, hiding dropdown", n), $.hideMenu()) : ($.remove.empty(), $.remove.message()), g.allowAdditions && $.add.userSuggestion($.escape.htmlEntities(t)), $.is.searchSelection() && $.can.show() && $.is.focusedOnSearch() && $.show(), O.fomanticExt.onDropdownAfterFiltered.call(H);
					};
					g.useLabels && $.has.maxSelections() || (g.apiSettings ? $.can.useAPI() ? $.queryRemote(n, function() {
						g.filterRemoteData && $.filterItems(n);
						var t = N.val();
						Array.isArray(t) || (t = t && t !== "" ? t.split(g.delimiter) : []), $.is.multiple() && e.each(t, function(e, t) {
							I.filter("[data-value=\"" + CSS.escape(t) + "\"]").addClass(_.filtered);
						}), i();
					}) : $.error(w.noAPI) : ($.filterItems(n), i()));
				},
				queryRemote: function(t, n) {
					var r = {
						errorDuration: !1,
						cache: "local",
						throttle: g.throttle,
						urlData: { query: t },
						onError: function() {
							$.add.message(v.serverError), n();
						},
						onFailure: function() {
							$.add.message(v.serverError), n();
						},
						onSuccess: function(e) {
							var t = e[y.remoteValues];
							Array.isArray(t) || (t = []), $.remove.message();
							var r = {};
							r[y.values] = t, $.setup.menu(r), t.length === 0 && !g.allowAdditions && $.add.message(v.noResults), n();
						}
					};
					O.api("get request") || $.setup.api(), r = e.extend(!0, {}, r, g.apiSettings), O.api("setting", r).api("query");
				},
				filterItems: function(t) {
					var n = $.remove.diacritics(t === r ? $.get.query() : t), i = null, a = $.escape.string(n), o = (g.ignoreSearchCase ? "i" : "") + "gm", s = RegExp("^" + a, o);
					$.has.query() && (i = [], $.verbose("Searching for matching values", n), I.each(function() {
						var t = e(this), r, a;
						if (t.hasClass(_.unfilterable) || (g.match === "both" || g.match === "text") && (r = $.remove.diacritics(String($.get.choiceText(t, !1))), r.search(s) !== -1 || g.fullTextSearch === "exact" && $.exactSearch(n, r) || g.fullTextSearch === !0 && $.fuzzySearch(n, r)) || (g.match === "both" || g.match === "value") && (a = $.remove.diacritics(String($.get.choiceValue(t, r))), a.search(s) !== -1 || g.fullTextSearch === "exact" && $.exactSearch(n, a) || g.fullTextSearch === !0 && $.fuzzySearch(n, a))) return i.push(this), !0;
					})), $.debug("Showing only matched items", n), $.remove.filteredItem(), i && I.not(i).addClass(_.filtered), $.has.query() ? g.hideDividers === !0 ? L.addClass(_.hidden) : g.hideDividers === "empty" && L.removeClass(_.hidden).filter(function() {
						var t = e(this).nextUntil(C.item);
						return (t.length ? t : e(this)).nextUntil(C.divider).filter(C.item + ":not(." + _.filtered + ")").length === 0;
					}).addClass(_.hidden) : L.removeClass(_.hidden);
				},
				fuzzySearch: function(e, t) {
					var n = t.length, r = e.length;
					if (e = g.ignoreSearchCase ? e.toLowerCase() : e, t = g.ignoreSearchCase ? t.toLowerCase() : t, r > n) return !1;
					if (r === n) return e === t;
					search: for (var i = 0, a = 0; i < r; i++) {
						for (var o = e.charCodeAt(i); a < n;) if (t.charCodeAt(a++) === o) continue search;
						return !1;
					}
					return !0;
				},
				exactSearch: function(e, t) {
					return e = g.ignoreSearchCase ? e.toLowerCase() : e, t = g.ignoreSearchCase ? t.toLowerCase() : t, t.indexOf(e) > -1;
				},
				filterActive: function() {
					g.useLabels && I.filter("." + _.active).addClass(_.filtered);
				},
				focusSearch: function(e) {
					$.has.search() && !$.is.focusedOnSearch() && (e ? (O.off("focus" + E, C.search), j.focus(), O.on("focus" + E, C.search, $.event.search.focus)) : j.focus());
				},
				blurSearch: function() {
					$.has.search() && j.blur();
				},
				forceSelection: function() {
					var e = I.not(_.filtered).filter("." + _.selected).eq(0), t = I.not(_.filtered).filter("." + _.active).eq(0), n = e.length > 0 ? e : t, r = n.length > 0;
					g.allowAdditions || r && !$.is.multiple() ? ($.debug("Forcing partial selection to selected item", n), $.event.item.click.call(n, {}, !0)) : $.remove.searchTerm();
				},
				change: { values: function(t) {
					g.allowAdditions || $.clear(), $.debug("Creating dropdown with specified values", t);
					var n = {};
					n[y.values] = t, $.setup.menu(n), e.each(t, function(e, t) {
						if (t.selected == 1 && ($.debug("Setting initial selection to", t[y.value]), $.set.selected(t[y.value]), !$.is.multiple())) return !1;
					}), $.has.selectInput() && ($.disconnect.selectObserver(), N.html(""), N.append("<option disabled selected value></option>"), e.each(t, function(e, t) {
						var n = g.templates.escape(t[y.value]), r = g.templates.escape(t[y.name] || "", g.preserveHTML);
						N.append("<option value=\"" + n + "\">" + r + "</option>");
					}), $.observe.select());
				} },
				event: {
					change: function() {
						B || ($.debug("Input changed, updating selection"), $.set.selected());
					},
					focus: function() {
						g.showOnFocus && !R && $.is.hidden() && !K && $.show();
					},
					blur: function(e) {
						K = n.activeElement === this, !R && !K && ($.remove.activeLabel(), $.hide());
					},
					mousedown: function() {
						$.is.searchSelection() ? q = !0 : R = !0;
					},
					mouseup: function() {
						$.is.searchSelection() ? q = !1 : R = !1;
					},
					click: function(t) {
						e(t.target).is(O) && ($.is.focusedOnSearch() ? $.show() : $.focusSearch());
					},
					search: {
						focus: function(e) {
							R = !0, $.is.multiple() && $.remove.activeLabel(), (g.showOnFocus || e.type !== "focus" && e.type !== "focusin") && $.search();
						},
						blur: function(e) {
							K = n.activeElement === this, $.is.searchSelection() && !q && !z && !K && (g.forceSelection ? $.forceSelection() : g.allowAdditions || $.remove.searchTerm(), $.hide()), q = !1;
						}
					},
					clearIcon: { click: function(e) {
						$.clear(), $.is.searchSelection() && $.remove.searchTerm(), $.hide(), e.stopPropagation();
					} },
					icon: { click: function(e) {
						V = !0;
						let t = !!F.find(".search > input").length;
						$.has.search() && !t ? $.is.active() ? $.blurSearch() : g.showOnFocus ? $.focusSearch() : $.toggle() : $.toggle();
					} },
					text: { focus: function(e) {
						R = !0, $.focusSearch();
					} },
					input: function(e) {
						($.is.multiple() || $.is.searchSelection()) && $.set.filtered(), clearTimeout($.timer), $.timer = setTimeout($.search, g.delay.search);
					},
					label: { click: function(t) {
						var n = e(this), r = O.find(C.label), i = r.filter("." + _.active), a = n.nextAll("." + _.active), o = n.prevAll("." + _.active), s = a.length > 0 ? n.nextUntil(a).add(i).add(n) : n.prevUntil(o).add(i).add(n);
						t.shiftKey ? (i.removeClass(_.active), s.addClass(_.active)) : t.ctrlKey ? n.toggleClass(_.active) : (i.removeClass(_.active), n.addClass(_.active)), g.onLabelSelect.apply(this, r.filter("." + _.active));
					} },
					remove: { click: function() {
						var t = e(this).parent();
						t.hasClass(_.active) ? $.remove.activeLabels() : $.remove.activeLabels(t);
					} },
					test: {
						toggle: function(e) {
							var t = $.is.multiple() ? $.show : $.toggle;
							$.is.bubbledLabelClick(e) || $.is.bubbledIconClick(e) || $.determine.eventOnElement(e, t) && e.preventDefault();
						},
						touch: function(e) {
							$.determine.eventOnElement(e, function() {
								e.type == "touchstart" ? $.timer = setTimeout(function() {
									$.hide();
								}, g.delay.touch) : e.type == "touchmove" && clearTimeout($.timer);
							}), e.stopPropagation();
						},
						hide: function(t) {
							$.determine.eventInModule(t, $.hide) && H.id && e(t.target).attr("for") === H.id && t.preventDefault();
						}
					},
					class: { mutation: function(e) {
						e.forEach(function(e) {
							e.attributeName === "class" && $.check.disabled();
						});
					} },
					select: { mutation: function(e) {
						$.debug("<select> modified, recreating menu"), $.is.selectMutation(e) && ($.disconnect.selectObserver(), $.refresh(), $.setup.select(), $.set.selected(), $.observe.select());
					} },
					menu: {
						mutation: function(t) {
							var n = t[0], r = n.addedNodes ? e(n.addedNodes[0]) : e(!1), i = n.removedNodes ? e(n.removedNodes[0]) : e(!1), a = r.add(i), o = a.is(C.addition) || a.closest(C.addition).length > 0, s = a.is(C.message) || a.closest(C.message).length > 0;
							o || s ? ($.debug("Updating item selector cache"), $.refreshItems()) : ($.debug("Menu modified, updating selector cache"), $.refresh());
						},
						mousedown: function() {
							z = !0;
						},
						mouseup: function() {
							z = !1;
						}
					},
					item: {
						mouseenter: function(t) {
							var n = e(t.target), r = e(this), i = r.children(C.menu), a = r.siblings(C.item).children(C.menu), o = i.length > 0;
							!(i.find(n).length > 0) && o && (clearTimeout($.itemTimer), $.itemTimer = setTimeout(function() {
								$.verbose("Showing sub-menu", i), e.each(a, function() {
									$.animate.hide(!1, e(this));
								}), $.animate.show(!1, i);
							}, g.delay.show), t.preventDefault());
						},
						mouseleave: function(t) {
							var n = e(this).children(C.menu);
							n.length > 0 && (clearTimeout($.itemTimer), $.itemTimer = setTimeout(function() {
								$.verbose("Hiding sub-menu", n), $.animate.hide(!1, n);
							}, g.delay.hide));
						},
						click: function(t, r) {
							var i = e(this), a = e(t ? t.target : ""), o = i.find(C.menu), s = $.get.choiceText(i), c = $.get.choiceValue(i, s), l = o.length > 0, u = o.find(a).length > 0;
							n.activeElement.tagName.toLowerCase() !== "input" && e(n.activeElement).blur(), !u && (!l || g.allowCategorySelection) && ($.is.searchSelection() && (g.allowAdditions && $.remove.userAddition(), $.remove.searchTerm(), !$.is.focusedOnSearch() && r != 1 && $.focusSearch(!0)), g.useLabels || ($.remove.filteredItem(), $.set.scrollPosition(i)), $.determine.selectAction.call(this, s, c));
						}
					},
					document: { keydown: function(e) {
						var t = e.which;
						if ($.is.inObject(t, b)) {
							var n = O.find(C.label), r = n.filter("." + _.active);
							r.data(x.value);
							var i = n.index(r), a = n.length, o = r.length > 0, s = r.length > 1, c = i === 0, l = i + 1 == a, u = $.is.searchSelection(), d = $.is.focusedOnSearch(), f = $.is.focused(), p = d && $.get.caretPosition(!1) === 0, m = p && $.get.caretPosition(!0) !== 0;
							if (u && !o && !d) return;
							t == b.leftArrow ? (f || p) && !o ? ($.verbose("Selecting previous label"), n.last().addClass(_.active)) : o && (e.shiftKey ? $.verbose("Adding previous label to selection") : ($.verbose("Selecting previous label"), n.removeClass(_.active)), c && !s ? r.addClass(_.active) : r.prev(C.siblingLabel).addClass(_.active).end(), e.preventDefault()) : t == b.rightArrow ? (f && !o && n.first().addClass(_.active), o && (e.shiftKey ? $.verbose("Adding next label to selection") : ($.verbose("Selecting next label"), n.removeClass(_.active)), l ? u ? d ? n.removeClass(_.active) : $.focusSearch() : s ? r.next(C.siblingLabel).addClass(_.active) : r.addClass(_.active) : r.next(C.siblingLabel).addClass(_.active), e.preventDefault())) : t == b.deleteKey || t == b.backspace ? o ? ($.verbose("Removing active labels"), l && u && !d && $.focusSearch(), r.last().next(C.siblingLabel).addClass(_.active), $.remove.activeLabels(r), e.preventDefault()) : p && !m && !o && t == b.backspace && ($.verbose("Removing last label on input backspace"), r = n.last().addClass(_.active), $.remove.activeLabels(r)) : r.removeClass(_.active);
						}
					} },
					keydown: function(e) {
						var t = e.which;
						if ($.is.inObject(t, b)) {
							var n = I.not(C.unselectable).filter("." + _.selected).eq(0), r = F.children("." + _.active).eq(0), i = n.length > 0 ? n : r, a = i.length > 0 ? i.siblings(":not(." + _.filtered + ")").addBack() : F.children(":not(." + _.filtered + ")"), o = i.children(C.menu), s = i.closest(C.menu), c = s.hasClass(_.visible) || s.hasClass(_.animating) || s.parent(C.menu).length > 0, l = o.length > 0, u = i.length > 0, d = i.not(C.unselectable).length > 0, f = t == b.delimiter && g.allowAdditions && $.is.multiple(), p = g.allowAdditions && g.hideAdditions && (t == b.enter || f) && d, m, h;
							if (p && ($.verbose("Selecting item from keyboard shortcut", i), $.event.item.click.call(i, e), $.is.searchSelection() && $.remove.searchTerm(), $.is.multiple() && e.preventDefault()), $.is.visible()) {
								if ((t == b.enter || f) && (t == b.enter && u && l && !g.allowCategorySelection ? ($.verbose("Pressed enter on unselectable category, opening sub menu"), t = b.rightArrow) : d && ($.verbose("Selecting item from keyboard shortcut", i), $.event.item.click.call(i, e), $.is.searchSelection() && ($.remove.searchTerm(), $.is.multiple() && j.focus())), e.preventDefault()), u && (t == b.leftArrow && (h = s[0] !== F[0], h && ($.verbose("Left key pressed, closing sub-menu"), $.animate.hide(!1, s), i.removeClass(_.selected), s.closest(C.item).addClass(_.selected), e.preventDefault())), t == b.rightArrow && l && ($.verbose("Right key pressed, opening sub-menu"), $.animate.show(!1, o), i.removeClass(_.selected), o.find(C.item).eq(0).addClass(_.selected), e.preventDefault())), t == b.upArrow) {
									if (m = u && c ? i.prevAll(C.item + ":not(" + C.unselectable + ")").eq(0) : I.eq(0), a.index(m) < 0) {
										$.verbose("Up key pressed but reached top of current menu"), e.preventDefault();
										return;
									} else $.verbose("Up key pressed, changing active item"), i.removeClass(_.selected), m.addClass(_.selected), $.set.scrollPosition(m), g.selectOnKeydown && $.is.single() && $.set.selectedItem(m);
									e.preventDefault();
								}
								if (t == b.downArrow) {
									if (m = u && c ? m = i.nextAll(C.item + ":not(" + C.unselectable + ")").eq(0) : I.eq(0), m.length === 0) {
										$.verbose("Down key pressed but reached bottom of current menu"), e.preventDefault();
										return;
									} else $.verbose("Down key pressed, changing active item"), I.removeClass(_.selected), m.addClass(_.selected), $.set.scrollPosition(m), g.selectOnKeydown && $.is.single() && $.set.selectedItem(m);
									e.preventDefault();
								}
								t == b.pageUp && ($.scrollPage("up"), e.preventDefault()), t == b.pageDown && ($.scrollPage("down"), e.preventDefault()), t == b.escape && ($.verbose("Escape key pressed, closing dropdown"), $.hide());
							} else f && e.preventDefault(), t == b.downArrow && !$.is.visible() && ($.verbose("Down key pressed, showing dropdown"), $.show(), e.preventDefault());
						} else $.has.search() || $.set.selectedLetter(String.fromCharCode(t));
					}
				},
				trigger: { change: function() {
					var e = N[0];
					if (e) {
						var t = n.createEvent("HTMLEvents");
						$.verbose("Triggering native change event"), t.initEvent("change", !0, !1), e.dispatchEvent(t);
					}
				} },
				determine: {
					selectAction: function(t, n) {
						W = !0, $.verbose("Determining action", g.action), e.isFunction($.action[g.action]) ? ($.verbose("Triggering preset action", g.action, t, n), $.action[g.action].call(H, t, n, this)) : e.isFunction(g.action) ? ($.verbose("Triggering user action", g.action, t, n), g.action.call(H, t, n, this)) : $.error(w.action, g.action), W = !1;
					},
					eventInModule: function(t, r) {
						var i = e(t.target), a = i.closest(n.documentElement).length > 0, o = i.closest(O).length > 0;
						return r = e.isFunction(r) ? r : function() {}, a && !o ? ($.verbose("Triggering event", r), r(), !0) : ($.verbose("Event occurred in dropdown, canceling callback"), !1);
					},
					eventOnElement: function(t, r) {
						var i = e(t.target), a = i.closest(C.siblingLabel), o = n.body.contains(t.target), s = O.find(a).length === 0 || !($.is.multiple() && g.useLabels), c = i.closest(F).length === 0;
						return r = e.isFunction(r) ? r : function() {}, o && s && c ? ($.verbose("Triggering event", r), r(), !0) : ($.verbose("Event occurred in dropdown menu, canceling callback"), !1);
					}
				},
				action: {
					nothing: function() {},
					activate: function(t, n, i) {
						n = n === r ? t : n, $.can.activate(e(i)) && ($.set.selected(n, e(i)), $.is.multiple() || $.hideAndClear());
					},
					select: function(t, n, i) {
						n = n === r ? t : n, $.can.activate(e(i)) && ($.set.value(n, t, e(i)), $.is.multiple() || $.hideAndClear());
					},
					combo: function(t, n, i) {
						n = n === r ? t : n, $.set.selected(n, e(i)), $.hideAndClear();
					},
					hide: function(t, n, r) {
						$.set.value(n, t, e(r)), $.hideAndClear();
					}
				},
				get: {
					id: function() {
						return Y;
					},
					defaultText: function() {
						return O.data(x.defaultText);
					},
					defaultValue: function() {
						return O.data(x.defaultValue);
					},
					placeholderText: function() {
						return g.placeholder != "auto" && typeof g.placeholder == "string" ? g.placeholder : O.data(x.placeholderText) || "";
					},
					text: function() {
						return g.preserveHTML ? A.html() : A.text();
					},
					query: function() {
						return String(j.val()).trim();
					},
					searchWidth: function(e) {
						return e = e === r ? j.val() : e, M.text(e), Math.ceil(M.width() + 1);
					},
					selectionCount: function() {
						var e = $.get.values();
						return $.is.multiple() ? Array.isArray(e) ? e.length : 0 : $.get.value() === "" ? 0 : 1;
					},
					transition: function(e) {
						return g.transition == "auto" ? $.is.upward(e) ? "slide up" : "slide down" : g.transition;
					},
					userValues: function() {
						var t = $.get.values();
						return t ? (t = Array.isArray(t) ? t : [t], e.grep(t, function(e) {
							return $.get.item(e) === !1;
						})) : !1;
					},
					uniqueArray: function(t) {
						return e.grep(t, function(n, r) {
							return e.inArray(n, t) === r;
						});
					},
					caretPosition: function(e) {
						var t = j.get(0), r, i;
						if (e && "selectionEnd" in t) return t.selectionEnd;
						if (!e && "selectionStart" in t) return t.selectionStart;
						if (n.selection) return t.focus(), r = n.selection.createRange(), i = r.text.length, e ? i : (r.moveStart("character", -t.value.length), r.text.length - i);
					},
					value: function() {
						var e = N.length > 0 ? N.val() : O.data(x.value), t = Array.isArray(e) && e.length === 1 && e[0] === "";
						return e === r || t ? "" : e;
					},
					values: function() {
						var e = $.get.value();
						return e === "" ? "" : !$.has.selectInput() && $.is.multiple() ? typeof e == "string" ? $.escape.htmlEntities(e).split(g.delimiter) : "" : e;
					},
					remoteValues: function() {
						var t = $.get.values(), n = !1;
						return t && (typeof t == "string" && (t = [t]), e.each(t, function(e, t) {
							var r = $.read.remoteData(t);
							$.verbose("Restoring value from session data", r, t), r && (n ||= {}, n[t] = r);
						})), n;
					},
					choiceText: function(e, t) {
						if (t = t === r ? g.preserveHTML : t, e) return e.find(C.menu).length > 0 && ($.verbose("Retrieving text of element with sub-menu"), e = e.clone(), e.find(C.menu).remove(), e.find(C.menuIcon).remove()), e.attr("data-" + x.text) === r ? t ? e.html().trim() : e.text().trim() : e.attr("data-" + x.text);
					},
					choiceValue: function(e, t) {
						return t ||= $.get.choiceText(e), e ? e.data(x.value) === r ? typeof t == "string" ? String(g.ignoreSearchCase ? t.toLowerCase() : t).trim() : String(t) : String(e.data(x.value)) : !1;
					},
					inputEvent: function() {
						var e = j[0];
						return e ? e.oninput === r ? e.onpropertychange === r ? "keyup" : "propertychange" : "input" : !1;
					},
					selectValues: function() {
						var t = {}, n = [], i = [];
						return O.find("option").each(function() {
							var a = e(this), o = a.html(), s = a.attr("disabled"), c = a.attr("value") === r ? o : a.attr("value"), l = a.attr("data-" + x.text) === r ? o : a.attr("data-" + x.text), u = a.parent("optgroup");
							g.placeholder === "auto" && c === "" ? t.placeholder = o : ((u.length !== n.length || u[0] !== n[0]) && (i.push({
								type: "header",
								divider: g.headerDivider,
								name: u.attr("label") || ""
							}), n = u), i.push({
								name: o,
								value: c,
								text: l,
								disabled: s
							}));
						}), g.placeholder && g.placeholder !== "auto" && ($.debug("Setting placeholder value to", g.placeholder), t.placeholder = g.placeholder), g.sortSelect ? (g.sortSelect === !0 ? i.sort(function(e, t) {
							return e.name.localeCompare(t.name);
						}) : g.sortSelect === "natural" ? i.sort(function(e, t) {
							return e.name.toLowerCase().localeCompare(t.name.toLowerCase());
						}) : e.isFunction(g.sortSelect) && i.sort(g.sortSelect), t[y.values] = i, $.debug("Retrieved and sorted values from select", t)) : (t[y.values] = i, $.debug("Retrieved values from select", t)), t;
					},
					activeItem: function() {
						return I.filter("." + _.active);
					},
					selectedItem: function() {
						var e = I.not(C.unselectable).filter("." + _.selected);
						return e.length > 0 ? e : I.eq(0);
					},
					itemWithAdditions: function(e) {
						var t = $.get.item(e), n = $.create.userChoice(e);
						return n && n.length > 0 && (t = t.length > 0 ? t.add(n) : n), t;
					},
					item: function(t, n) {
						var i = !1, a, o;
						return t = t === r ? $.get.values() === r ? $.get.text() : $.get.values() : t, o = $.is.multiple() && Array.isArray(t), a = o ? t.length > 0 : t !== r && t !== null, n = t === "" || t === !1 || t === !0 ? !0 : n || !1, a && I.each(function() {
							var a = e(this), s = $.get.choiceText(a), c = $.get.choiceValue(a, s);
							if (!(c === null || c === r)) {
								if (o) e.inArray($.escape.htmlEntities(String(c)), t.map(function(e) {
									return String(e);
								})) !== -1 && (i = i ? i.add(a) : a);
								else if (n) {
									if ($.verbose("Ambiguous dropdown value using strict type check", a, t), c === t) return i = a, !0;
								} else if (g.ignoreCase && (c = c.toLowerCase(), t = t.toLowerCase()), $.escape.htmlEntities(String(c)) === $.escape.htmlEntities(String(t))) return $.verbose("Found select item by value", c, t), i = a, !0;
							}
						}), i;
					}
				},
				check: {
					maxSelections: function(e) {
						return g.maxSelections ? (e = e === r ? $.get.selectionCount() : e, e >= g.maxSelections ? ($.debug("Maximum selection count reached"), g.useLabels && (I.addClass(_.filtered), $.add.message(v.maxSelections)), !0) : ($.verbose("No longer at maximum selection count"), $.remove.message(), $.remove.filteredItem(), $.is.searchSelection() && $.filterItems(), !1)) : !0;
					},
					disabled: function() {
						j.attr("tabindex", $.is.disabled() ? -1 : 0);
					}
				},
				restore: {
					defaults: function(e) {
						$.clear(e), $.restore.defaultText(), $.restore.defaultValue();
					},
					defaultText: function() {
						var e = $.get.defaultText();
						e === $.get.placeholderText ? ($.debug("Restoring default placeholder text", e), $.set.placeholderText(e)) : ($.debug("Restoring default text", e), $.set.text(e));
					},
					placeholderText: function() {
						$.set.placeholderText();
					},
					defaultValue: function() {
						var e = $.get.defaultValue();
						e !== r && ($.debug("Restoring default value", e), e === "" ? ($.remove.activeItem(), $.remove.selectedItem()) : ($.set.value(e), $.set.selected()));
					},
					labels: function() {
						g.allowAdditions && (g.useLabels ||= ($.error(w.labels), !0), $.debug("Restoring selected values"), $.create.userLabels()), $.check.maxSelections();
					},
					selected: function() {
						$.restore.values(), $.is.multiple() ? ($.debug("Restoring previously selected values and labels"), $.restore.labels()) : $.debug("Restoring previously selected values");
					},
					values: function() {
						$.set.initialLoad(), g.apiSettings && g.saveRemoteData && $.get.remoteValues() ? $.restore.remoteValues() : $.set.selected();
						var e = $.get.value();
						e && e !== "" && !(Array.isArray(e) && e.length === 0) ? N.removeClass(_.noselection) : N.addClass(_.noselection), $.remove.initialLoad();
					},
					remoteValues: function() {
						var t = $.get.remoteValues();
						$.debug("Recreating selected from session data", t), t && ($.is.single() ? e.each(t, function(e, t) {
							$.set.text(t);
						}) : e.each(t, function(e, t) {
							$.add.label(e, t);
						}));
					}
				},
				read: { remoteData: function(e) {
					var n;
					if (t.Storage === r) {
						$.error(w.noStorage);
						return;
					}
					return n = sessionStorage.getItem(e), n === r ? !1 : n;
				} },
				save: {
					defaults: function() {
						$.save.defaultText(), $.save.placeholderText(), $.save.defaultValue();
					},
					defaultValue: function() {
						var e = $.get.value();
						$.verbose("Saving default value as", e), O.data(x.defaultValue, e);
					},
					defaultText: function() {
						var e = $.get.text();
						$.verbose("Saving default text as", e), O.data(x.defaultText, e);
					},
					placeholderText: function() {
						var e;
						g.placeholder !== !1 && A.hasClass(_.placeholder) && (e = $.get.text(), $.verbose("Saving placeholder text as", e), O.data(x.placeholderText, e));
					},
					remoteData: function(e, n) {
						if (t.Storage === r) {
							$.error(w.noStorage);
							return;
						}
						$.verbose("Saving remote data to session storage", n, e), sessionStorage.setItem(n, e);
					}
				},
				clear: function(e) {
					$.is.multiple() && g.useLabels ? $.remove.labels() : ($.remove.activeItem(), $.remove.selectedItem(), $.remove.filteredItem()), $.set.placeholderText(), $.clearValue(e);
				},
				clearValue: function(e) {
					$.set.value("", null, null, e);
				},
				scrollPage: function(e, t) {
					var n = t || $.get.selectedItem(), r = n.closest(C.menu), i = r.outerHeight(), a = r.scrollTop(), o = I.eq(0).outerHeight(), s = Math.floor(i / o);
					r.prop("scrollHeight");
					var c = e == "up" ? a - o * s : a + o * s, l = I.not(C.unselectable), u, d, f = e == "up" ? l.index(n) - s : l.index(n) + s;
					u = e == "up" ? f >= 0 : f < l.length, d = u ? l.eq(f) : e == "up" ? l.first() : l.last(), d.length > 0 && ($.debug("Scrolling page", e, d), n.removeClass(_.selected), d.addClass(_.selected), g.selectOnKeydown && $.is.single() && $.set.selectedItem(d), r.scrollTop(c));
				},
				set: {
					filtered: function() {
						var e = $.is.multiple(), t = $.is.searchSelection(), n = e && t, r = t ? $.get.query() : "", i = typeof r == "string" && r.length > 0, a = $.get.searchWidth(), o = r !== "";
						e && i && ($.verbose("Adjusting input width", a, g.glyphWidth), j.css("width", a)), i || n && o ? ($.verbose("Hiding placeholder text"), A.addClass(_.filtered)) : (!e || n && !o) && ($.verbose("Showing placeholder text"), A.removeClass(_.filtered));
					},
					empty: function() {
						O.addClass(_.empty);
					},
					loading: function() {
						O.addClass(_.loading);
					},
					placeholderText: function(e) {
						e ||= $.get.placeholderText(), $.debug("Setting placeholder text", e), $.set.text(e), A.addClass(_.placeholder);
					},
					tabbable: function() {
						$.is.searchSelection() ? ($.debug("Added tabindex to searchable dropdown"), j.val(""), $.check.disabled(), F.attr("tabindex", -1)) : ($.debug("Added tabindex to dropdown"), O.attr("tabindex") === r && (O.attr("tabindex", 0), F.attr("tabindex", -1)));
					},
					initialLoad: function() {
						$.verbose("Setting initial load"), G = !0;
					},
					activeItem: function(e) {
						g.allowAdditions && e.filter(C.addition).length > 0 ? e.addClass(_.filtered) : e.addClass(_.active);
					},
					partialSearch: function(e) {
						var t = $.get.query().length;
						j.val(e.substr(0, t));
					},
					scrollPosition: function(e, t) {
						var n = 5, i, a, o, s, c, l, u, d, f;
						e ||= $.get.selectedItem(), i = e.closest(C.menu), a = e && e.length > 0, t = t === r ? !1 : t, $.get.activeItem().length === 0 && (t = !1), e && i.length > 0 && a && (s = e.position().top, i.addClass(_.loading), l = i.scrollTop(), c = i.offset().top, s = e.offset().top, o = l - c + s, t || (u = i.height(), f = l + u < o + n, d = o - n < l), $.debug("Scrolling to active item", o), (t || d || f) && i.scrollTop(o), i.removeClass(_.loading));
					},
					text: function(e) {
						g.action === "combo" ? ($.debug("Changing combo button text", e, P), g.preserveHTML ? P.html(e) : P.text(e)) : g.action === "activate" && (e !== $.get.placeholderText() && A.removeClass(_.placeholder), $.debug("Changing text", e, A), A.removeClass(_.filtered), g.preserveHTML ? A.html(e) : A.text(e));
					},
					selectedItem: function(e) {
						var t = $.get.choiceValue(e), n = $.get.choiceText(e, !1), r = $.get.choiceText(e, !0);
						$.debug("Setting user selection to item", e), $.remove.activeItem(), $.set.partialSearch(n), $.set.activeItem(e), $.set.selected(t, e), $.set.text(r);
					},
					selectedLetter: function(t) {
						var n = I.filter("." + _.selected), r = n.length > 0 && $.has.firstLetter(n, t), i = !1, a;
						r && (a = n.nextAll(I).eq(0), $.has.firstLetter(a, t) && (i = a)), i || I.each(function() {
							if ($.has.firstLetter(e(this), t)) return i = e(this), !1;
						}), i && ($.verbose("Scrolling to next value with letter", t), $.set.scrollPosition(i), n.removeClass(_.selected), i.addClass(_.selected), g.selectOnKeydown && $.is.single() && $.set.selectedItem(i));
					},
					direction: function(e) {
						g.direction == "auto" ? (e ? $.is.upward(e) && $.remove.upward(e) : $.remove.upward(), $.can.openDownward(e) ? $.remove.upward(e) : $.set.upward(e), !$.is.leftward(e) && !$.can.openRightward(e) && $.set.leftward(e)) : g.direction == "upward" && $.set.upward(e);
					},
					upward: function(e) {
						(e || O).addClass(_.upward);
					},
					leftward: function(e) {
						(e || F).addClass(_.leftward);
					},
					value: function(e, t, n, i) {
						e !== r && e !== "" && !(Array.isArray(e) && e.length === 0) ? N.removeClass(_.noselection) : N.addClass(_.noselection);
						var a = $.escape.value(e), o = N.length > 0, s = $.get.values(), c = e === r ? e : String(e);
						if (o) {
							if (!g.allowReselection && c == s && ($.verbose("Skipping value update already same value", e, s), !$.is.initialLoad())) return;
							$.is.single() && $.has.selectInput() && $.can.extendSelect() && ($.debug("Adding user option", e), $.add.optionValue(e)), $.debug("Updating input value", a, s), B = !0, N.val(a), g.fireOnInit === !1 && $.is.initialLoad() ? $.debug("Input native change event ignored on initial load") : i !== !0 && $.trigger.change(), B = !1;
						} else $.verbose("Storing value in metadata", a, N), a !== s && O.data(x.value, c);
						g.fireOnInit === !1 && $.is.initialLoad() ? $.verbose("No callback on initial load", g.onChange) : i !== !0 && g.onChange.call(H, e, t, n);
					},
					active: function() {
						O.addClass(_.active);
					},
					multiple: function() {
						O.addClass(_.multiple);
					},
					visible: function() {
						O.addClass(_.visible);
					},
					exactly: function(e, t) {
						$.debug("Setting selected to exact values"), $.clear(), $.set.selected(e, t);
					},
					selected: function(t, n) {
						var r = $.is.multiple();
						n = g.allowAdditions ? n || $.get.itemWithAdditions(t) : n || $.get.item(t), n && ($.debug("Setting selected menu item to", n), $.is.multiple() && $.remove.searchWidth(), $.is.single() ? ($.remove.activeItem(), $.remove.selectedItem()) : g.useLabels && $.remove.selectedItem(), n.each(function() {
							var t = e(this), i = $.get.choiceText(t), a = $.get.choiceValue(t, i), o = t.hasClass(_.filtered), s = t.hasClass(_.active), c = t.hasClass(_.addition), l = r && n.length == 1;
							r ? !s || c ? (g.apiSettings && g.saveRemoteData && $.save.remoteData(i, a), g.useLabels ? ($.add.label(a, i, l), $.add.value(a, i, t), $.set.activeItem(t), $.filterActive(), $.select.nextAvailable(n)) : ($.add.value(a, i, t), $.set.text($.add.variables(v.count)), $.set.activeItem(t))) : !o && (g.useLabels || W) && ($.debug("Selected active value, removing label"), $.remove.selected(a)) : (g.apiSettings && g.saveRemoteData && $.save.remoteData(i, a), $.set.text(i), $.set.value(a, i, t), t.addClass(_.active).addClass(_.selected));
						}), $.remove.searchTerm());
					}
				},
				add: {
					label: function(t, n, r) {
						var i = $.is.searchSelection() ? j : A, a = $.escape.value(t), o;
						if (g.ignoreCase && (a = a.toLowerCase()), o = e("<a />").addClass(_.label).attr("data-" + x.value, a).html(T.label(a, n, g.preserveHTML, g.className)), o = g.onLabelCreate.call(o, a, n), $.has.label(t)) {
							$.debug("User selection already exists, skipping", a);
							return;
						}
						g.label.variation && o.addClass(g.label.variation), r === !0 ? ($.debug("Animating in label", o), o.addClass(_.hidden).insertBefore(i).transition({
							animation: g.label.transition,
							debug: g.debug,
							verbose: g.verbose,
							duration: g.label.duration
						})) : ($.debug("Adding selection label", o), o.insertBefore(i));
					},
					message: function(t) {
						var n = F.children(C.message), r = g.templates.message($.add.variables(t));
						n.length > 0 ? n.html(r) : n = e("<div/>").html(r).addClass(_.message).appendTo(F);
					},
					optionValue: function(t) {
						var n = $.escape.value(t);
						N.find("option[value=\"" + $.escape.string(n) + "\"]").length > 0 || ($.disconnect.selectObserver(), $.is.single() && ($.verbose("Removing previous user addition"), N.find("option." + _.addition).remove()), e("<option/>").prop("value", n).addClass(_.addition).html(t).appendTo(N), $.verbose("Adding user addition as an <option>", t), $.observe.select());
					},
					userSuggestion: function(e) {
						var t = F.children(C.addition), n = $.get.item(e), r = n && n.not(C.addition).length, i = t.length > 0, a;
						if (!(g.useLabels && $.has.maxSelections())) {
							if (e === "" || r) {
								t.remove();
								return;
							}
							i ? (t.data(x.value, e).data(x.text, e).attr("data-" + x.value, e).attr("data-" + x.text, e).removeClass(_.filtered), g.hideAdditions || (a = g.templates.addition($.add.variables(v.addResult, e)), t.html(a)), $.verbose("Replacing user suggestion with new value", t)) : (t = $.create.userChoice(e), t.prependTo(F), $.verbose("Adding item choice to menu corresponding with user choice addition", t)), (!g.hideAdditions || $.is.allFiltered()) && t.addClass(_.selected).siblings().removeClass(_.selected), $.refreshItems();
						}
					},
					variables: function(e, t) {
						var n = e.search("{count}") !== -1, r = e.search("{maxCount}") !== -1, i = e.search("{term}") !== -1, a, o;
						return $.verbose("Adding templated variables to message", e), n && (a = $.get.selectionCount(), e = e.replace("{count}", a)), r && (a = $.get.selectionCount(), e = e.replace("{maxCount}", g.maxSelections)), i && (o = t || $.get.query(), e = e.replace("{term}", o)), e;
					},
					value: function(e, t, n) {
						var r = $.get.values(), i;
						if ($.has.value(e)) {
							$.debug("Value already selected");
							return;
						}
						if (e === "") {
							$.debug("Cannot select blank values from multiselect");
							return;
						}
						Array.isArray(r) ? (i = r.concat([e]), i = $.get.uniqueArray(i)) : i = [e], $.has.selectInput() ? $.can.extendSelect() && ($.debug("Adding value to select", e, i, N), $.add.optionValue(e)) : (i = i.join(g.delimiter), $.debug("Setting hidden input to delimited value", i, N)), g.fireOnInit === !1 && $.is.initialLoad() ? $.verbose("Skipping onadd callback on initial load", g.onAdd) : g.onAdd.call(H, e, t, n), $.set.value(i, t, n), $.check.maxSelections();
					}
				},
				remove: {
					active: function() {
						O.removeClass(_.active);
					},
					activeLabel: function() {
						O.find(C.label).removeClass(_.active);
					},
					empty: function() {
						O.removeClass(_.empty);
					},
					loading: function() {
						O.removeClass(_.loading);
					},
					initialLoad: function() {
						G = !1;
					},
					upward: function(e) {
						(e || O).removeClass(_.upward);
					},
					leftward: function(e) {
						(e || F).removeClass(_.leftward);
					},
					visible: function() {
						O.removeClass(_.visible);
					},
					activeItem: function() {
						I.removeClass(_.active);
					},
					filteredItem: function() {
						g.useLabels && $.has.maxSelections() || (g.useLabels && $.is.multiple() ? I.not("." + _.active).removeClass(_.filtered) : I.removeClass(_.filtered), g.hideDividers && L.removeClass(_.hidden), $.remove.empty());
					},
					optionValue: function(e) {
						var t = $.escape.value(e), n = N.find("option[value=\"" + $.escape.string(t) + "\"]");
						!(n.length > 0) || !n.hasClass(_.addition) || (X && (X.disconnect(), $.verbose("Temporarily disconnecting mutation observer")), n.remove(), $.verbose("Removing user addition as an <option>", t), X && X.observe(N[0], {
							childList: !0,
							subtree: !0
						}));
					},
					message: function() {
						F.children(C.message).remove();
					},
					searchWidth: function() {
						j.css("width", "");
					},
					searchTerm: function() {
						$.verbose("Cleared search term"), j.val(""), $.set.filtered();
					},
					userAddition: function() {
						I.filter(C.addition).remove();
					},
					selected: function(t, n) {
						if (n = g.allowAdditions ? n || $.get.itemWithAdditions(t) : n || $.get.item(t), !n) return !1;
						n.each(function() {
							var t = e(this), n = $.get.choiceText(t), r = $.get.choiceValue(t, n);
							$.is.multiple() ? g.useLabels ? ($.remove.value(r, n, t), $.remove.label(r)) : ($.remove.value(r, n, t), $.get.selectionCount() === 0 ? $.set.placeholderText() : $.set.text($.add.variables(v.count))) : $.remove.value(r, n, t), t.removeClass(_.filtered).removeClass(_.active), g.useLabels && t.removeClass(_.selected);
						});
					},
					selectedItem: function() {
						I.removeClass(_.selected);
					},
					value: function(e, t, n) {
						var r = $.get.values(), i;
						e = $.escape.htmlEntities(e), $.has.selectInput() ? ($.verbose("Input is <select> removing selected option", e), i = $.remove.arrayValue(e, r), $.remove.optionValue(e)) : ($.verbose("Removing from delimited values", e), i = $.remove.arrayValue(e, r), i = i.join(g.delimiter)), g.fireOnInit === !1 && $.is.initialLoad() ? $.verbose("No callback on initial load", g.onRemove) : g.onRemove.call(H, e, t, n), $.set.value(i, t, n), $.check.maxSelections();
					},
					arrayValue: function(t, n) {
						return Array.isArray(n) || (n = [n]), n = e.grep(n, function(e) {
							return t != e;
						}), $.verbose("Removed value from delimited string", t, n), n;
					},
					label: function(e, t) {
						var n = O.find(C.label).filter("[data-" + x.value + "=\"" + $.escape.string(g.ignoreCase ? e.toLowerCase() : e) + "\"]");
						$.verbose("Removing label", n), n.remove();
					},
					activeLabels: function(e) {
						e ||= O.find(C.label).filter("." + _.active), $.verbose("Removing active label selections", e), $.remove.labels(e);
					},
					labels: function(t) {
						t ||= O.find(C.label), $.verbose("Removing labels", t), t.each(function() {
							var t = e(this), n = t.data(x.value), i = n === r ? n : String(n), a = $.is.userValue(i);
							if (g.onLabelRemove.call(t, n) === !1) {
								$.debug("Label remove callback cancelled removal");
								return;
							}
							$.remove.message(), a ? ($.remove.value(i), $.remove.label(i)) : $.remove.selected(i);
						});
					},
					tabbable: function() {
						$.is.searchSelection() ? ($.debug("Searchable dropdown initialized"), j.removeAttr("tabindex"), F.removeAttr("tabindex")) : ($.debug("Simple selection dropdown initialized"), O.removeAttr("tabindex"), F.removeAttr("tabindex"));
					},
					diacritics: function(e) {
						return g.ignoreDiacritics ? e.normalize("NFD").replace(/[\u0300-\u036f]/g, "") : e;
					}
				},
				has: {
					menuSearch: function() {
						return $.has.search() && j.closest(F).length > 0;
					},
					clearItem: function() {
						return re.length > 0;
					},
					search: function() {
						return j.length > 0;
					},
					sizer: function() {
						return M.length > 0;
					},
					selectInput: function() {
						return N.is("select");
					},
					minCharacters: function(e) {
						return g.minCharacters && !V ? (e = e === r ? String($.get.query()) : String(e), e.length >= g.minCharacters) : (V = !1, !0);
					},
					firstLetter: function(e, t) {
						var n, r;
						return !e || e.length === 0 || typeof t != "string" ? !1 : (n = $.get.choiceText(e, !1), t = t.toLowerCase(), r = String(n).charAt(0).toLowerCase(), t == r);
					},
					input: function() {
						return N.length > 0;
					},
					items: function() {
						return I.length > 0;
					},
					menu: function() {
						return F.length > 0;
					},
					message: function() {
						return F.children(C.message).length !== 0;
					},
					label: function(e) {
						var t = $.escape.value(e), n = O.find(C.label);
						return g.ignoreCase && (t = t.toLowerCase()), n.filter("[data-" + x.value + "=\"" + $.escape.string(t) + "\"]").length > 0;
					},
					maxSelections: function() {
						return g.maxSelections && $.get.selectionCount() >= g.maxSelections;
					},
					allResultsFiltered: function() {
						var e = I.not(C.addition);
						return e.filter(C.unselectable).length === e.length;
					},
					userSuggestion: function() {
						return F.children(C.addition).length > 0;
					},
					query: function() {
						return $.get.query() !== "";
					},
					value: function(e) {
						return g.ignoreCase ? $.has.valueIgnoringCase(e) : $.has.valueMatchingCase(e);
					},
					valueMatchingCase: function(t) {
						var n = $.get.values();
						return !!(Array.isArray(n) ? n && e.inArray(t, n) !== -1 : n == t);
					},
					valueIgnoringCase: function(t) {
						var n = $.get.values(), r = !1;
						return Array.isArray(n) || (n = [n]), e.each(n, function(e, n) {
							if (String(t).toLowerCase() == String(n).toLowerCase()) return r = !0, !1;
						}), r;
					}
				},
				is: {
					active: function() {
						return O.hasClass(_.active);
					},
					animatingInward: function() {
						return F.transition("is inward");
					},
					animatingOutward: function() {
						return F.transition("is outward");
					},
					bubbledLabelClick: function(t) {
						return e(t.target).is("select, input") && O.closest("label").length > 0;
					},
					bubbledIconClick: function(t) {
						return e(t.target).closest(ne).length > 0;
					},
					alreadySetup: function() {
						return O.is("select") && O.parent(C.dropdown).data(D) !== r && O.prev().length === 0;
					},
					animating: function(e) {
						return e ? e.transition && e.transition("is animating") : F.transition && F.transition("is animating");
					},
					leftward: function(e) {
						return (e || F).hasClass(_.leftward);
					},
					clearable: function() {
						return O.hasClass(_.clearable) || g.clearable;
					},
					disabled: function() {
						return O.hasClass(_.disabled);
					},
					focused: function() {
						return n.activeElement === O[0];
					},
					focusedOnSearch: function() {
						return n.activeElement === j[0];
					},
					allFiltered: function() {
						return ($.is.multiple() || $.has.search()) && !(g.hideAdditions == 0 && $.has.userSuggestion()) && !$.has.message() && $.has.allResultsFiltered();
					},
					hidden: function(e) {
						return !$.is.visible(e);
					},
					initialLoad: function() {
						return G;
					},
					inObject: function(t, n) {
						var r = !1;
						return e.each(n, function(e, n) {
							if (n == t) return r = !0, !0;
						}), r;
					},
					multiple: function() {
						return O.hasClass(_.multiple);
					},
					remote: function() {
						return g.apiSettings && $.can.useAPI();
					},
					single: function() {
						return !$.is.multiple();
					},
					selectMutation: function(t) {
						var n = !1;
						return e.each(t, function(t, r) {
							if (e(r.target).is("select") || e(r.addedNodes).is("select")) return n = !0, !1;
						}), n;
					},
					search: function() {
						return O.hasClass(_.search);
					},
					searchSelection: function() {
						return $.has.search() && j.parent(C.dropdown).length === 1;
					},
					selection: function() {
						return O.hasClass(_.selection);
					},
					userValue: function(t) {
						return e.inArray(t, $.get.userValues()) !== -1;
					},
					upward: function(e) {
						return (e || O).hasClass(_.upward);
					},
					visible: function(e) {
						return e ? e.hasClass(_.visible) : F.hasClass(_.visible);
					},
					verticallyScrollableContext: function() {
						var e = k.get(0) === t ? !1 : k.css("overflow-y");
						return e == "auto" || e == "scroll";
					},
					horizontallyScrollableContext: function() {
						var e = k.get(0) === t ? !1 : k.css("overflow-X");
						return e == "auto" || e == "scroll";
					}
				},
				can: {
					activate: function(e) {
						return !!(g.useLabels || !$.has.maxSelections() || $.has.maxSelections() && e.hasClass(_.active));
					},
					openDownward: function(e) {
						var n = e || F, r = !0, i = {}, a;
						return n.addClass(_.loading), a = {
							context: {
								offset: k.get(0) === t ? {
									top: 0,
									left: 0
								} : k.offset(),
								scrollTop: k.scrollTop(),
								height: k.outerHeight()
							},
							menu: {
								offset: n.offset(),
								height: n.outerHeight()
							}
						}, $.is.verticallyScrollableContext() && (a.menu.offset.top += a.context.scrollTop), i = {
							above: a.context.scrollTop <= a.menu.offset.top - a.context.offset.top - a.menu.height,
							below: a.context.scrollTop + a.context.height >= a.menu.offset.top - a.context.offset.top + a.menu.height
						}, i.below ? ($.verbose("Dropdown can fit in context downward", i), r = !0) : !i.below && !i.above ? ($.verbose("Dropdown cannot fit in either direction, favoring downward", i), r = !0) : ($.verbose("Dropdown cannot fit below, opening upward", i), r = !1), n.removeClass(_.loading), r;
					},
					openRightward: function(e) {
						var n = e || F, r = !0, i = !1, a;
						return n.addClass(_.loading), a = {
							context: {
								offset: k.get(0) === t ? {
									top: 0,
									left: 0
								} : k.offset(),
								scrollLeft: k.scrollLeft(),
								width: k.outerWidth()
							},
							menu: {
								offset: n.offset(),
								width: n.outerWidth()
							}
						}, $.is.horizontallyScrollableContext() && (a.menu.offset.left += a.context.scrollLeft), i = a.menu.offset.left - a.context.offset.left + a.menu.width >= a.context.scrollLeft + a.context.width, i && ($.verbose("Dropdown cannot fit in context rightward", i), r = !1), n.removeClass(_.loading), r;
					},
					click: function() {
						return c || g.on == "click";
					},
					extendSelect: function() {
						return g.allowAdditions || g.apiSettings;
					},
					show: function() {
						return !$.is.disabled() && ($.has.items() || $.has.message());
					},
					useAPI: function() {
						return e.fn.api !== r;
					}
				},
				animate: {
					show: function(t, n) {
						var i = n || F, a = n ? function() {} : function() {
							$.hideSubMenus(), $.hideOthers(), $.set.active();
						}, o;
						if (t = e.isFunction(t) ? t : function() {}, $.verbose("Doing menu show animation", i), $.set.direction(n), o = $.get.transition(n), $.is.selection() && $.set.scrollPosition($.get.selectedItem(), !0), $.is.hidden(i) || $.is.animating(i)) {
							var s = O.hasClass("column") ? "flex" : !1;
							o == "none" ? (a(), i.transition({ displayType: s }).transition("show"), t.call(H)) : e.fn.transition !== r && O.transition("is supported") ? i.transition({
								animation: o + " in",
								debug: g.debug,
								verbose: g.verbose,
								duration: g.duration,
								queue: !0,
								onStart: a,
								displayType: s,
								onComplete: function() {
									t.call(H);
								}
							}) : $.error(w.noTransition, o);
						}
					},
					hide: function(t, n) {
						var i = n || F, a = n ? function() {} : function() {
							$.can.click() && $.unbind.intent(), $.remove.active();
						}, o = $.get.transition(n);
						t = e.isFunction(t) ? t : function() {}, ($.is.visible(i) || $.is.animating(i)) && ($.verbose("Doing menu hide animation", i), o == "none" ? (a(), i.transition("hide"), t.call(H)) : e.fn.transition !== r && O.transition("is supported") ? i.transition({
							animation: o + " out",
							duration: g.duration,
							debug: g.debug,
							verbose: g.verbose,
							queue: !1,
							onStart: a,
							onComplete: function() {
								t.call(H);
							}
						}) : $.error(w.transition));
					}
				},
				hideAndClear: function() {
					$.remove.searchTerm(), !$.has.maxSelections() && ($.has.search() ? $.hide(function() {
						$.remove.filteredItem();
					}) : $.hide());
				},
				delay: {
					show: function() {
						$.verbose("Delaying show event to ensure user intent"), clearTimeout($.timer), $.timer = setTimeout($.show, g.delay.show);
					},
					hide: function() {
						$.verbose("Delaying hide event to ensure user intent"), clearTimeout($.timer), $.timer = setTimeout($.hide, g.delay.hide);
					}
				},
				escape: {
					value: function(t) {
						var n = Array.isArray(t), r = typeof t == "string", i = !r && !n, a = r && t.search(S.quote) !== -1, o = [];
						return i || !a ? t : ($.debug("Encoding quote values for use in select", t), n ? (e.each(t, function(e, t) {
							o.push(t.replace(S.quote, "&quot;"));
						}), o) : t.replace(S.quote, "&quot;"));
					},
					string: function(e) {
						return e = String(e), e.replace(S.escape, "\\$&");
					},
					htmlEntities: function(e) {
						var t = /[<>"'`]/g, n = /[&<>"'`]/, r = {
							"<": "&lt;",
							">": "&gt;",
							"\"": "&quot;",
							"'": "&#x27;",
							"`": "&#x60;"
						};
						return n.test(e) ? (e = e.replace(/&(?![a-z0-9#]{1,6};)/, "&amp;"), e.replace(t, function(e) {
							return r[e];
						})) : e;
					}
				},
				setting: function(t, n) {
					if ($.debug("Changing setting", t, n), e.isPlainObject(t)) e.extend(!0, g, t);
					else if (n !== r) e.isPlainObject(g[t]) ? e.extend(!0, g[t], n) : g[t] = n;
					else return g[t];
				},
				internal: function(t, n) {
					if (e.isPlainObject(t)) e.extend(!0, $, t);
					else if (n !== r) $[t] = n;
					else return $[t];
				},
				debug: function() {
					!g.silent && g.debug && (g.performance ? $.performance.log(arguments) : ($.debug = Function.prototype.bind.call(console.info, console, g.name + ":"), $.debug.apply(console, arguments)));
				},
				verbose: function() {
					!g.silent && g.verbose && g.debug && (g.performance ? $.performance.log(arguments) : ($.verbose = Function.prototype.bind.call(console.info, console, g.name + ":"), $.verbose.apply(console, arguments)));
				},
				error: function() {
					g.silent || ($.error = Function.prototype.bind.call(console.error, console, g.name + ":"), $.error.apply(console, arguments));
				},
				performance: {
					log: function(e) {
						var t, n, r;
						g.performance && (t = (/* @__PURE__ */ new Date()).getTime(), r = u || t, n = t - r, u = t, d.push({
							Name: e[0],
							Arguments: [].slice.call(e, 1) || "",
							Element: H,
							"Execution Time": n
						})), clearTimeout($.performance.timer), $.performance.timer = setTimeout($.performance.display, 500);
					},
					display: function() {
						var t = g.name + ":", n = 0;
						u = !1, clearTimeout($.performance.timer), e.each(d, function(e, t) {
							n += t["Execution Time"];
						}), t += " " + n + "ms", s && (t += " '" + s + "'"), (console.group !== r || console.table !== r) && d.length > 0 && (console.groupCollapsed(t), console.table ? console.table(d) : e.each(d, function(e, t) {
							console.log(t.Name + ": " + t["Execution Time"] + "ms");
						}), console.groupEnd()), d = [];
					}
				},
				invoke: function(t, n, i) {
					var a = U, o, s, c;
					return n ||= m, i = H || i, typeof t == "string" && a !== r && (t = t.split(/[\. ]/), o = t.length - 1, e.each(t, function(n, i) {
						var c = n == o ? t : i + t[n + 1].charAt(0).toUpperCase() + t[n + 1].slice(1);
						if (e.isPlainObject(a[c]) && n != o) a = a[c];
						else if (a[c] !== r) return s = a[c], !1;
						else if (e.isPlainObject(a[i]) && n != o) a = a[i];
						else if (a[i] !== r) return s = a[i], !1;
						else return $.error(w.method, t), !1;
					})), e.isFunction(s) ? c = s.apply(i, n) : s !== r && (c = s), Array.isArray(h) ? h.push(c) : h === r ? c !== r && (h = c) : h = [h, c], s;
				}
			};
			p ? (U === r && $.initialize(), $.invoke(f)) : (U !== r && U.invoke("destroy"), $.initialize());
		}), h === r ? a : h;
	}, e.fn.dropdown.settings = {
		silent: !1,
		debug: !1,
		verbose: !1,
		performance: !0,
		on: "click",
		action: "activate",
		values: !1,
		clearable: !1,
		apiSettings: !1,
		selectOnKeydown: !0,
		minCharacters: 0,
		filterRemoteData: !1,
		saveRemoteData: !0,
		throttle: 200,
		context: t,
		direction: "auto",
		keepOnScreen: !0,
		match: "both",
		fullTextSearch: !1,
		ignoreDiacritics: !1,
		hideDividers: !1,
		placeholder: "auto",
		preserveHTML: !0,
		sortSelect: !1,
		forceSelection: !0,
		allowAdditions: !1,
		ignoreCase: !1,
		ignoreSearchCase: !0,
		hideAdditions: !0,
		maxSelections: !1,
		useLabels: !0,
		delimiter: ",",
		showOnFocus: !0,
		allowReselection: !1,
		allowTab: !0,
		allowCategorySelection: !1,
		fireOnInit: !1,
		transition: "auto",
		duration: 200,
		glyphWidth: 1.037,
		headerDivider: !0,
		label: {
			transition: "scale",
			duration: 200,
			variation: !1
		},
		delay: {
			hide: 300,
			show: 200,
			search: 20,
			touch: 50
		},
		onChange: function(e, t, n) {},
		onAdd: function(e, t, n) {},
		onRemove: function(e, t, n) {},
		onLabelSelect: function(e) {},
		onLabelCreate: function(t, n) {
			return e(this);
		},
		onLabelRemove: function(e) {
			return !0;
		},
		onNoResults: function(e) {
			return !0;
		},
		onShow: function() {},
		onHide: function() {},
		name: "Dropdown",
		namespace: "dropdown",
		message: {
			addResult: "Add <b>{term}</b>",
			count: "{count} selected",
			maxSelections: "Max {maxCount} selections",
			noResults: "No results found.",
			serverError: "There was an error contacting the server"
		},
		error: {
			action: "You called a dropdown action that was not defined",
			alreadySetup: "Once a select has been initialized behaviors must be called on the created ui dropdown",
			labels: "Allowing user additions currently requires the use of labels.",
			missingMultiple: "<select> requires multiple property to be set to correctly preserve multiple values",
			method: "The method you called is not defined.",
			noAPI: "The API module is required to load resources remotely",
			noStorage: "Saving remote data requires session storage",
			noTransition: "This module requires ui transitions <https://github.com/Semantic-Org/UI-Transition>",
			noNormalize: "\"ignoreDiacritics\" setting will be ignored. Browser does not support String().normalize(). You may consider including <https://cdn.jsdelivr.net/npm/unorm@1.4.1/lib/unorm.min.js> as a polyfill."
		},
		regExp: {
			escape: /[-[\]{}()*+?.,\\^$|#\s:=@]/g,
			quote: /"/g
		},
		metadata: {
			defaultText: "defaultText",
			defaultValue: "defaultValue",
			placeholderText: "placeholder",
			text: "text",
			value: "value"
		},
		fields: {
			remoteValues: "results",
			values: "values",
			disabled: "disabled",
			name: "name",
			value: "value",
			text: "text",
			type: "type",
			image: "image",
			imageClass: "imageClass",
			icon: "icon",
			iconClass: "iconClass",
			class: "class",
			divider: "divider"
		},
		keys: {
			backspace: 8,
			delimiter: 188,
			deleteKey: 46,
			enter: 13,
			escape: 27,
			pageUp: 33,
			pageDown: 34,
			leftArrow: 37,
			upArrow: 38,
			rightArrow: 39,
			downArrow: 40
		},
		selector: {
			addition: ".addition",
			divider: ".divider, .header",
			dropdown: ".ui.dropdown",
			hidden: ".hidden",
			icon: "> .dropdown.icon",
			input: "> input[type=\"hidden\"], > select",
			item: ".item",
			label: "> .label",
			remove: "> .label > .delete.icon",
			siblingLabel: ".label",
			menu: ".menu",
			message: ".message",
			menuIcon: ".dropdown.icon",
			search: "input.search, .menu > .search > input, .menu input.search",
			sizer: "> span.sizer",
			text: "> .text:not(.icon)",
			unselectable: ".disabled, .filtered, .tw-hidden",
			clearIcon: "> .remove.icon"
		},
		className: {
			active: "active",
			addition: "addition",
			animating: "animating",
			disabled: "disabled",
			empty: "empty",
			dropdown: "ui dropdown",
			filtered: "filtered",
			hidden: "hidden transition",
			icon: "icon",
			image: "image",
			item: "item",
			label: "ui label",
			loading: "loading",
			menu: "menu",
			message: "message",
			multiple: "multiple",
			placeholder: "default",
			sizer: "sizer",
			search: "search",
			selected: "selected",
			selection: "selection",
			upward: "upward",
			leftward: "left",
			visible: "visible",
			clearable: "clearable",
			noselection: "noselection",
			delete: "delete",
			header: "header",
			divider: "divider",
			groupIcon: "",
			unfilterable: "unfilterable"
		}
	}, e.fn.dropdown.settings.templates = {
		deQuote: function(e) {
			return String(e).replace(/"/g, "");
		},
		escape: function(e, t) {
			if (t) return e;
			var n = /[<>"'`]/g, r = /[&<>"'`]/, i = {
				"<": "&lt;",
				">": "&gt;",
				"\"": "&quot;",
				"'": "&#x27;",
				"`": "&#x60;"
			};
			return r.test(e) ? (e = e.replace(/&(?![a-z0-9#]{1,6};)/, "&amp;"), e.replace(n, function(e) {
				return i[e];
			})) : e;
		},
		dropdown: function(t, n, r, i) {
			var a = t.placeholder || !1, o = "", s = e.fn.dropdown.settings.templates.escape;
			return o += "<i class=\"dropdown icon\"><svg viewBox=\"0 0 16 16\" class=\"svg octicon-triangle-down\" aria-hidden=\"true\" width=\"14\" height=\"14\"><path d=\"m4.427 7.427 3.396 3.396a.25.25 0 0 0 .354 0l3.396-3.396A.25.25 0 0 0 11.396 7H4.604a.25.25 0 0 0-.177.427\"></path></svg></i>", a ? o += "<div class=\"default text\">" + s(a, r) + "</div>" : o += "<div class=\"text\"></div>", o += "<div class=\"" + i.menu + "\">", o += e.fn.dropdown.settings.templates.menu(t, n, r, i), o += "</div>", o;
		},
		menu: function(t, n, r, i) {
			var a = t[n.values] || [], o = "", s = e.fn.dropdown.settings.templates.escape, c = e.fn.dropdown.settings.templates.deQuote;
			return e.each(a, function(e, t) {
				var a = t[n.type] ? t[n.type] : "item";
				if (a === "item") {
					var l = t[n.text] ? " data-text=\"" + s(t[n.text]) + "\"" : "", u = t[n.disabled] ? i.disabled + " " : "";
					o += "<div class=\"" + u + (t[n.class] ? c(t[n.class]) : i.item) + "\" data-value=\"" + s(t[n.value]) + "\"" + l + ">", t[n.image] && (o += "<img class=\"" + (t[n.imageClass] ? c(t[n.imageClass]) : i.image) + "\" src=\"" + c(t[n.image]) + "\">"), t[n.icon] && (o += "<i class=\"" + c(t[n.icon]) + " " + (t[n.iconClass] ? c(t[n.iconClass]) : i.icon) + "\"></i>"), o += s(t[n.name] || "", r), o += "</div>";
				} else if (a === "header") {
					var d = s(t[n.name] || "", r), f = t[n.icon] ? c(t[n.icon]) : i.groupIcon;
					(d !== "" || f !== "") && (o += "<div class=\"" + (t[n.class] ? c(t[n.class]) : i.header) + "\">", f !== "" && (o += "<i class=\"" + f + " " + (t[n.iconClass] ? c(t[n.iconClass]) : i.icon) + "\"></i>"), o += d, o += "</div>"), t[n.divider] && (o += "<div class=\"" + i.divider + "\"></div>");
				}
			}), o;
		},
		label: function(t, n, r, i) {
			var a = e.fn.dropdown.settings.templates.escape;
			return a(n, r) + "<i class=\"" + i.delete + " icon\"></i>";
		},
		message: function(e) {
			return e;
		},
		addition: function(e) {
			return e;
		}
	};
})(jQuery, window, document);
//#endregion
