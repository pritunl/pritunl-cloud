"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-mk-_lib-formatDistance-index-js"],{

/***/ "./node_modules/date-fns/locale/mk/_lib/formatDistance/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/mk/_lib/formatDistance/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar formatDistanceLocale = {\n  lessThanXSeconds: {\n    one: 'помалку од секунда',\n    other: 'помалку од {{count}} секунди'\n  },\n  xSeconds: {\n    one: '1 секунда',\n    other: '{{count}} секунди'\n  },\n  halfAMinute: 'половина минута',\n  lessThanXMinutes: {\n    one: 'помалку од минута',\n    other: 'помалку од {{count}} минути'\n  },\n  xMinutes: {\n    one: '1 минута',\n    other: '{{count}} минути'\n  },\n  aboutXHours: {\n    one: 'околу 1 час',\n    other: 'околу {{count}} часа'\n  },\n  xHours: {\n    one: '1 час',\n    other: '{{count}} часа'\n  },\n  xDays: {\n    one: '1 ден',\n    other: '{{count}} дена'\n  },\n  aboutXWeeks: {\n    one: 'околу 1 недела',\n    other: 'околу {{count}} месеци'\n  },\n  xWeeks: {\n    one: '1 недела',\n    other: '{{count}} недели'\n  },\n  aboutXMonths: {\n    one: 'околу 1 месец',\n    other: 'околу {{count}} недели'\n  },\n  xMonths: {\n    one: '1 месец',\n    other: '{{count}} месеци'\n  },\n  aboutXYears: {\n    one: 'околу 1 година',\n    other: 'околу {{count}} години'\n  },\n  xYears: {\n    one: '1 година',\n    other: '{{count}} години'\n  },\n  overXYears: {\n    one: 'повеќе од 1 година',\n    other: 'повеќе од {{count}} години'\n  },\n  almostXYears: {\n    one: 'безмалку 1 година',\n    other: 'безмалку {{count}} години'\n  }\n};\nvar formatDistance = function formatDistance(token, count, options) {\n  var result;\n  var tokenValue = formatDistanceLocale[token];\n  if (typeof tokenValue === 'string') {\n    result = tokenValue;\n  } else if (count === 1) {\n    result = tokenValue.one;\n  } else {\n    result = tokenValue.other.replace('{{count}}', String(count));\n  }\n  if (options !== null && options !== void 0 && options.addSuffix) {\n    if (options.comparison && options.comparison > 0) {\n      return 'за ' + result;\n    } else {\n      return 'пред ' + result;\n    }\n  }\n  return result;\n};\nvar _default = formatDistance;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL21rL19saWIvZm9ybWF0RGlzdGFuY2UvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQSx5QkFBeUIsUUFBUTtBQUNqQyxHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBO0FBQ0EseUJBQXlCLFFBQVE7QUFDakMsR0FBRztBQUNIO0FBQ0E7QUFDQSxjQUFjLFFBQVE7QUFDdEIsR0FBRztBQUNIO0FBQ0E7QUFDQSxvQkFBb0IsUUFBUTtBQUM1QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLG9CQUFvQixRQUFRO0FBQzVCLEdBQUc7QUFDSDtBQUNBO0FBQ0EsY0FBYyxRQUFRO0FBQ3RCLEdBQUc7QUFDSDtBQUNBO0FBQ0Esb0JBQW9CLFFBQVE7QUFDNUIsR0FBRztBQUNIO0FBQ0E7QUFDQSxjQUFjLFFBQVE7QUFDdEIsR0FBRztBQUNIO0FBQ0E7QUFDQSxvQkFBb0IsUUFBUTtBQUM1QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLHdCQUF3QixRQUFRO0FBQ2hDLEdBQUc7QUFDSDtBQUNBO0FBQ0EsdUJBQXVCLFFBQVE7QUFDL0I7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxJQUFJO0FBQ0o7QUFDQSxJQUFJO0FBQ0oseUNBQXlDLE9BQU87QUFDaEQ7QUFDQTtBQUNBO0FBQ0E7QUFDQSxNQUFNO0FBQ047QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0Esa0JBQWU7QUFDZiIsInNvdXJjZXMiOlsid2VicGFjazovL3ByaXR1bmwtY2xvdWQvLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL21rL19saWIvZm9ybWF0RGlzdGFuY2UvaW5kZXguanM/ODVkMyJdLCJzb3VyY2VzQ29udGVudCI6WyJcInVzZSBzdHJpY3RcIjtcblxuT2JqZWN0LmRlZmluZVByb3BlcnR5KGV4cG9ydHMsIFwiX19lc01vZHVsZVwiLCB7XG4gIHZhbHVlOiB0cnVlXG59KTtcbmV4cG9ydHMuZGVmYXVsdCA9IHZvaWQgMDtcbnZhciBmb3JtYXREaXN0YW5jZUxvY2FsZSA9IHtcbiAgbGVzc1RoYW5YU2Vjb25kczoge1xuICAgIG9uZTogJ9C/0L7QvNCw0LvQutGDINC+0LQg0YHQtdC60YPQvdC00LAnLFxuICAgIG90aGVyOiAn0L/QvtC80LDQu9C60YMg0L7QtCB7e2NvdW50fX0g0YHQtdC60YPQvdC00LgnXG4gIH0sXG4gIHhTZWNvbmRzOiB7XG4gICAgb25lOiAnMSDRgdC10LrRg9C90LTQsCcsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0g0YHQtdC60YPQvdC00LgnXG4gIH0sXG4gIGhhbGZBTWludXRlOiAn0L/QvtC70L7QstC40L3QsCDQvNC40L3Rg9GC0LAnLFxuICBsZXNzVGhhblhNaW51dGVzOiB7XG4gICAgb25lOiAn0L/QvtC80LDQu9C60YMg0L7QtCDQvNC40L3Rg9GC0LAnLFxuICAgIG90aGVyOiAn0L/QvtC80LDQu9C60YMg0L7QtCB7e2NvdW50fX0g0LzQuNC90YPRgtC4J1xuICB9LFxuICB4TWludXRlczoge1xuICAgIG9uZTogJzEg0LzQuNC90YPRgtCwJyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSDQvNC40L3Rg9GC0LgnXG4gIH0sXG4gIGFib3V0WEhvdXJzOiB7XG4gICAgb25lOiAn0L7QutC+0LvRgyAxINGH0LDRgScsXG4gICAgb3RoZXI6ICfQvtC60L7Qu9GDIHt7Y291bnR9fSDRh9Cw0YHQsCdcbiAgfSxcbiAgeEhvdXJzOiB7XG4gICAgb25lOiAnMSDRh9Cw0YEnLFxuICAgIG90aGVyOiAne3tjb3VudH19INGH0LDRgdCwJ1xuICB9LFxuICB4RGF5czoge1xuICAgIG9uZTogJzEg0LTQtdC9JyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSDQtNC10L3QsCdcbiAgfSxcbiAgYWJvdXRYV2Vla3M6IHtcbiAgICBvbmU6ICfQvtC60L7Qu9GDIDEg0L3QtdC00LXQu9CwJyxcbiAgICBvdGhlcjogJ9C+0LrQvtC70YMge3tjb3VudH19INC80LXRgdC10YbQuCdcbiAgfSxcbiAgeFdlZWtzOiB7XG4gICAgb25lOiAnMSDQvdC10LTQtdC70LAnLFxuICAgIG90aGVyOiAne3tjb3VudH19INC90LXQtNC10LvQuCdcbiAgfSxcbiAgYWJvdXRYTW9udGhzOiB7XG4gICAgb25lOiAn0L7QutC+0LvRgyAxINC80LXRgdC10YYnLFxuICAgIG90aGVyOiAn0L7QutC+0LvRgyB7e2NvdW50fX0g0L3QtdC00LXQu9C4J1xuICB9LFxuICB4TW9udGhzOiB7XG4gICAgb25lOiAnMSDQvNC10YHQtdGGJyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSDQvNC10YHQtdGG0LgnXG4gIH0sXG4gIGFib3V0WFllYXJzOiB7XG4gICAgb25lOiAn0L7QutC+0LvRgyAxINCz0L7QtNC40L3QsCcsXG4gICAgb3RoZXI6ICfQvtC60L7Qu9GDIHt7Y291bnR9fSDQs9C+0LTQuNC90LgnXG4gIH0sXG4gIHhZZWFyczoge1xuICAgIG9uZTogJzEg0LPQvtC00LjQvdCwJyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSDQs9C+0LTQuNC90LgnXG4gIH0sXG4gIG92ZXJYWWVhcnM6IHtcbiAgICBvbmU6ICfQv9C+0LLQtdGc0LUg0L7QtCAxINCz0L7QtNC40L3QsCcsXG4gICAgb3RoZXI6ICfQv9C+0LLQtdGc0LUg0L7QtCB7e2NvdW50fX0g0LPQvtC00LjQvdC4J1xuICB9LFxuICBhbG1vc3RYWWVhcnM6IHtcbiAgICBvbmU6ICfQsdC10LfQvNCw0LvQutGDIDEg0LPQvtC00LjQvdCwJyxcbiAgICBvdGhlcjogJ9Cx0LXQt9C80LDQu9C60YMge3tjb3VudH19INCz0L7QtNC40L3QuCdcbiAgfVxufTtcbnZhciBmb3JtYXREaXN0YW5jZSA9IGZ1bmN0aW9uIGZvcm1hdERpc3RhbmNlKHRva2VuLCBjb3VudCwgb3B0aW9ucykge1xuICB2YXIgcmVzdWx0O1xuICB2YXIgdG9rZW5WYWx1ZSA9IGZvcm1hdERpc3RhbmNlTG9jYWxlW3Rva2VuXTtcbiAgaWYgKHR5cGVvZiB0b2tlblZhbHVlID09PSAnc3RyaW5nJykge1xuICAgIHJlc3VsdCA9IHRva2VuVmFsdWU7XG4gIH0gZWxzZSBpZiAoY291bnQgPT09IDEpIHtcbiAgICByZXN1bHQgPSB0b2tlblZhbHVlLm9uZTtcbiAgfSBlbHNlIHtcbiAgICByZXN1bHQgPSB0b2tlblZhbHVlLm90aGVyLnJlcGxhY2UoJ3t7Y291bnR9fScsIFN0cmluZyhjb3VudCkpO1xuICB9XG4gIGlmIChvcHRpb25zICE9PSBudWxsICYmIG9wdGlvbnMgIT09IHZvaWQgMCAmJiBvcHRpb25zLmFkZFN1ZmZpeCkge1xuICAgIGlmIChvcHRpb25zLmNvbXBhcmlzb24gJiYgb3B0aW9ucy5jb21wYXJpc29uID4gMCkge1xuICAgICAgcmV0dXJuICfQt9CwICcgKyByZXN1bHQ7XG4gICAgfSBlbHNlIHtcbiAgICAgIHJldHVybiAn0L/RgNC10LQgJyArIHJlc3VsdDtcbiAgICB9XG4gIH1cbiAgcmV0dXJuIHJlc3VsdDtcbn07XG52YXIgX2RlZmF1bHQgPSBmb3JtYXREaXN0YW5jZTtcbmV4cG9ydHMuZGVmYXVsdCA9IF9kZWZhdWx0O1xubW9kdWxlLmV4cG9ydHMgPSBleHBvcnRzLmRlZmF1bHQ7Il0sIm5hbWVzIjpbXSwic291cmNlUm9vdCI6IiJ9\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/mk/_lib/formatDistance/index.js\n");

/***/ })

}]);