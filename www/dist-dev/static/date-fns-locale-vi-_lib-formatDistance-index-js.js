"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-vi-_lib-formatDistance-index-js"],{

/***/ "./node_modules/date-fns/locale/vi/_lib/formatDistance/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/vi/_lib/formatDistance/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar formatDistanceLocale = {\n  lessThanXSeconds: {\n    one: 'dưới 1 giây',\n    other: 'dưới {{count}} giây'\n  },\n  xSeconds: {\n    one: '1 giây',\n    other: '{{count}} giây'\n  },\n  halfAMinute: 'nửa phút',\n  lessThanXMinutes: {\n    one: 'dưới 1 phút',\n    other: 'dưới {{count}} phút'\n  },\n  xMinutes: {\n    one: '1 phút',\n    other: '{{count}} phút'\n  },\n  aboutXHours: {\n    one: 'khoảng 1 giờ',\n    other: 'khoảng {{count}} giờ'\n  },\n  xHours: {\n    one: '1 giờ',\n    other: '{{count}} giờ'\n  },\n  xDays: {\n    one: '1 ngày',\n    other: '{{count}} ngày'\n  },\n  aboutXWeeks: {\n    one: 'khoảng 1 tuần',\n    other: 'khoảng {{count}} tuần'\n  },\n  xWeeks: {\n    one: '1 tuần',\n    other: '{{count}} tuần'\n  },\n  aboutXMonths: {\n    one: 'khoảng 1 tháng',\n    other: 'khoảng {{count}} tháng'\n  },\n  xMonths: {\n    one: '1 tháng',\n    other: '{{count}} tháng'\n  },\n  aboutXYears: {\n    one: 'khoảng 1 năm',\n    other: 'khoảng {{count}} năm'\n  },\n  xYears: {\n    one: '1 năm',\n    other: '{{count}} năm'\n  },\n  overXYears: {\n    one: 'hơn 1 năm',\n    other: 'hơn {{count}} năm'\n  },\n  almostXYears: {\n    one: 'gần 1 năm',\n    other: 'gần {{count}} năm'\n  }\n};\nvar formatDistance = function formatDistance(token, count, options) {\n  var result;\n  var tokenValue = formatDistanceLocale[token];\n  if (typeof tokenValue === 'string') {\n    result = tokenValue;\n  } else if (count === 1) {\n    result = tokenValue.one;\n  } else {\n    result = tokenValue.other.replace('{{count}}', String(count));\n  }\n  if (options !== null && options !== void 0 && options.addSuffix) {\n    if (options.comparison && options.comparison > 0) {\n      return result + ' nữa';\n    } else {\n      return result + ' trước';\n    }\n  }\n  return result;\n};\nvar _default = formatDistance;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL3ZpL19saWIvZm9ybWF0RGlzdGFuY2UvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQSxtQkFBbUIsUUFBUTtBQUMzQixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBO0FBQ0EsbUJBQW1CLFFBQVE7QUFDM0IsR0FBRztBQUNIO0FBQ0E7QUFDQSxjQUFjLFFBQVE7QUFDdEIsR0FBRztBQUNIO0FBQ0E7QUFDQSxxQkFBcUIsUUFBUTtBQUM3QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLHFCQUFxQixRQUFRO0FBQzdCLEdBQUc7QUFDSDtBQUNBO0FBQ0EsY0FBYyxRQUFRO0FBQ3RCLEdBQUc7QUFDSDtBQUNBO0FBQ0EscUJBQXFCLFFBQVE7QUFDN0IsR0FBRztBQUNIO0FBQ0E7QUFDQSxjQUFjLFFBQVE7QUFDdEIsR0FBRztBQUNIO0FBQ0E7QUFDQSxxQkFBcUIsUUFBUTtBQUM3QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGtCQUFrQixRQUFRO0FBQzFCLEdBQUc7QUFDSDtBQUNBO0FBQ0Esa0JBQWtCLFFBQVE7QUFDMUI7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxJQUFJO0FBQ0o7QUFDQSxJQUFJO0FBQ0oseUNBQXlDLE9BQU87QUFDaEQ7QUFDQTtBQUNBO0FBQ0E7QUFDQSxNQUFNO0FBQ047QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0Esa0JBQWU7QUFDZiIsInNvdXJjZXMiOlsid2VicGFjazovL3ByaXR1bmwtY2xvdWQvLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL3ZpL19saWIvZm9ybWF0RGlzdGFuY2UvaW5kZXguanM/M2IzYSJdLCJzb3VyY2VzQ29udGVudCI6WyJcInVzZSBzdHJpY3RcIjtcblxuT2JqZWN0LmRlZmluZVByb3BlcnR5KGV4cG9ydHMsIFwiX19lc01vZHVsZVwiLCB7XG4gIHZhbHVlOiB0cnVlXG59KTtcbmV4cG9ydHMuZGVmYXVsdCA9IHZvaWQgMDtcbnZhciBmb3JtYXREaXN0YW5jZUxvY2FsZSA9IHtcbiAgbGVzc1RoYW5YU2Vjb25kczoge1xuICAgIG9uZTogJ2TGsOG7m2kgMSBnacOieScsXG4gICAgb3RoZXI6ICdkxrDhu5tpIHt7Y291bnR9fSBnacOieSdcbiAgfSxcbiAgeFNlY29uZHM6IHtcbiAgICBvbmU6ICcxIGdpw6J5JyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSBnacOieSdcbiAgfSxcbiAgaGFsZkFNaW51dGU6ICdu4butYSBwaMO6dCcsXG4gIGxlc3NUaGFuWE1pbnV0ZXM6IHtcbiAgICBvbmU6ICdkxrDhu5tpIDEgcGjDunQnLFxuICAgIG90aGVyOiAnZMaw4bubaSB7e2NvdW50fX0gcGjDunQnXG4gIH0sXG4gIHhNaW51dGVzOiB7XG4gICAgb25lOiAnMSBwaMO6dCcsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0gcGjDunQnXG4gIH0sXG4gIGFib3V0WEhvdXJzOiB7XG4gICAgb25lOiAna2hv4bqjbmcgMSBnaeG7nScsXG4gICAgb3RoZXI6ICdraG/huqNuZyB7e2NvdW50fX0gZ2nhu50nXG4gIH0sXG4gIHhIb3Vyczoge1xuICAgIG9uZTogJzEgZ2nhu50nLFxuICAgIG90aGVyOiAne3tjb3VudH19IGdp4budJ1xuICB9LFxuICB4RGF5czoge1xuICAgIG9uZTogJzEgbmfDoHknLFxuICAgIG90aGVyOiAne3tjb3VudH19IG5nw6B5J1xuICB9LFxuICBhYm91dFhXZWVrczoge1xuICAgIG9uZTogJ2tob+G6o25nIDEgdHXhuqduJyxcbiAgICBvdGhlcjogJ2tob+G6o25nIHt7Y291bnR9fSB0deG6p24nXG4gIH0sXG4gIHhXZWVrczoge1xuICAgIG9uZTogJzEgdHXhuqduJyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSB0deG6p24nXG4gIH0sXG4gIGFib3V0WE1vbnRoczoge1xuICAgIG9uZTogJ2tob+G6o25nIDEgdGjDoW5nJyxcbiAgICBvdGhlcjogJ2tob+G6o25nIHt7Y291bnR9fSB0aMOhbmcnXG4gIH0sXG4gIHhNb250aHM6IHtcbiAgICBvbmU6ICcxIHRow6FuZycsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0gdGjDoW5nJ1xuICB9LFxuICBhYm91dFhZZWFyczoge1xuICAgIG9uZTogJ2tob+G6o25nIDEgbsSDbScsXG4gICAgb3RoZXI6ICdraG/huqNuZyB7e2NvdW50fX0gbsSDbSdcbiAgfSxcbiAgeFllYXJzOiB7XG4gICAgb25lOiAnMSBuxINtJyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSBuxINtJ1xuICB9LFxuICBvdmVyWFllYXJzOiB7XG4gICAgb25lOiAnaMahbiAxIG7Eg20nLFxuICAgIG90aGVyOiAnaMahbiB7e2NvdW50fX0gbsSDbSdcbiAgfSxcbiAgYWxtb3N0WFllYXJzOiB7XG4gICAgb25lOiAnZ+G6p24gMSBuxINtJyxcbiAgICBvdGhlcjogJ2fhuqduIHt7Y291bnR9fSBuxINtJ1xuICB9XG59O1xudmFyIGZvcm1hdERpc3RhbmNlID0gZnVuY3Rpb24gZm9ybWF0RGlzdGFuY2UodG9rZW4sIGNvdW50LCBvcHRpb25zKSB7XG4gIHZhciByZXN1bHQ7XG4gIHZhciB0b2tlblZhbHVlID0gZm9ybWF0RGlzdGFuY2VMb2NhbGVbdG9rZW5dO1xuICBpZiAodHlwZW9mIHRva2VuVmFsdWUgPT09ICdzdHJpbmcnKSB7XG4gICAgcmVzdWx0ID0gdG9rZW5WYWx1ZTtcbiAgfSBlbHNlIGlmIChjb3VudCA9PT0gMSkge1xuICAgIHJlc3VsdCA9IHRva2VuVmFsdWUub25lO1xuICB9IGVsc2Uge1xuICAgIHJlc3VsdCA9IHRva2VuVmFsdWUub3RoZXIucmVwbGFjZSgne3tjb3VudH19JywgU3RyaW5nKGNvdW50KSk7XG4gIH1cbiAgaWYgKG9wdGlvbnMgIT09IG51bGwgJiYgb3B0aW9ucyAhPT0gdm9pZCAwICYmIG9wdGlvbnMuYWRkU3VmZml4KSB7XG4gICAgaWYgKG9wdGlvbnMuY29tcGFyaXNvbiAmJiBvcHRpb25zLmNvbXBhcmlzb24gPiAwKSB7XG4gICAgICByZXR1cm4gcmVzdWx0ICsgJyBu4buvYSc7XG4gICAgfSBlbHNlIHtcbiAgICAgIHJldHVybiByZXN1bHQgKyAnIHRyxrDhu5tjJztcbiAgICB9XG4gIH1cbiAgcmV0dXJuIHJlc3VsdDtcbn07XG52YXIgX2RlZmF1bHQgPSBmb3JtYXREaXN0YW5jZTtcbmV4cG9ydHMuZGVmYXVsdCA9IF9kZWZhdWx0O1xubW9kdWxlLmV4cG9ydHMgPSBleHBvcnRzLmRlZmF1bHQ7Il0sIm5hbWVzIjpbXSwic291cmNlUm9vdCI6IiJ9\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/vi/_lib/formatDistance/index.js\n");

/***/ })

}]);