"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-af-_lib-formatDistance-index-js"],{

/***/ "./node_modules/date-fns/locale/af/_lib/formatDistance/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/af/_lib/formatDistance/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar formatDistanceLocale = {\n  lessThanXSeconds: {\n    one: \"minder as 'n sekonde\",\n    other: 'minder as {{count}} sekondes'\n  },\n  xSeconds: {\n    one: '1 sekonde',\n    other: '{{count}} sekondes'\n  },\n  halfAMinute: \"'n halwe minuut\",\n  lessThanXMinutes: {\n    one: \"minder as 'n minuut\",\n    other: 'minder as {{count}} minute'\n  },\n  xMinutes: {\n    one: \"'n minuut\",\n    other: '{{count}} minute'\n  },\n  aboutXHours: {\n    one: 'ongeveer 1 uur',\n    other: 'ongeveer {{count}} ure'\n  },\n  xHours: {\n    one: '1 uur',\n    other: '{{count}} ure'\n  },\n  xDays: {\n    one: '1 dag',\n    other: '{{count}} dae'\n  },\n  aboutXWeeks: {\n    one: 'ongeveer 1 week',\n    other: 'ongeveer {{count}} weke'\n  },\n  xWeeks: {\n    one: '1 week',\n    other: '{{count}} weke'\n  },\n  aboutXMonths: {\n    one: 'ongeveer 1 maand',\n    other: 'ongeveer {{count}} maande'\n  },\n  xMonths: {\n    one: '1 maand',\n    other: '{{count}} maande'\n  },\n  aboutXYears: {\n    one: 'ongeveer 1 jaar',\n    other: 'ongeveer {{count}} jaar'\n  },\n  xYears: {\n    one: '1 jaar',\n    other: '{{count}} jaar'\n  },\n  overXYears: {\n    one: 'meer as 1 jaar',\n    other: 'meer as {{count}} jaar'\n  },\n  almostXYears: {\n    one: 'byna 1 jaar',\n    other: 'byna {{count}} jaar'\n  }\n};\nvar formatDistance = function formatDistance(token, count, options) {\n  var result;\n  var tokenValue = formatDistanceLocale[token];\n  if (typeof tokenValue === 'string') {\n    result = tokenValue;\n  } else if (count === 1) {\n    result = tokenValue.one;\n  } else {\n    result = tokenValue.other.replace('{{count}}', String(count));\n  }\n  if (options !== null && options !== void 0 && options.addSuffix) {\n    if (options.comparison && options.comparison > 0) {\n      return 'oor ' + result;\n    } else {\n      return result + ' gelede';\n    }\n  }\n  return result;\n};\nvar _default = formatDistance;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2FmL19saWIvZm9ybWF0RGlzdGFuY2UvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQSx3QkFBd0IsUUFBUTtBQUNoQyxHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBO0FBQ0Esd0JBQXdCLFFBQVE7QUFDaEMsR0FBRztBQUNIO0FBQ0E7QUFDQSxjQUFjLFFBQVE7QUFDdEIsR0FBRztBQUNIO0FBQ0E7QUFDQSx1QkFBdUIsUUFBUTtBQUMvQixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLHVCQUF1QixRQUFRO0FBQy9CLEdBQUc7QUFDSDtBQUNBO0FBQ0EsY0FBYyxRQUFRO0FBQ3RCLEdBQUc7QUFDSDtBQUNBO0FBQ0EsdUJBQXVCLFFBQVE7QUFDL0IsR0FBRztBQUNIO0FBQ0E7QUFDQSxjQUFjLFFBQVE7QUFDdEIsR0FBRztBQUNIO0FBQ0E7QUFDQSx1QkFBdUIsUUFBUTtBQUMvQixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLHNCQUFzQixRQUFRO0FBQzlCLEdBQUc7QUFDSDtBQUNBO0FBQ0EsbUJBQW1CLFFBQVE7QUFDM0I7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxJQUFJO0FBQ0o7QUFDQSxJQUFJO0FBQ0oseUNBQXlDLE9BQU87QUFDaEQ7QUFDQTtBQUNBO0FBQ0E7QUFDQSxNQUFNO0FBQ047QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0Esa0JBQWU7QUFDZiIsInNvdXJjZXMiOlsid2VicGFjazovL3ByaXR1bmwtY2xvdWQvLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2FmL19saWIvZm9ybWF0RGlzdGFuY2UvaW5kZXguanM/Y2Q0YiJdLCJzb3VyY2VzQ29udGVudCI6WyJcInVzZSBzdHJpY3RcIjtcblxuT2JqZWN0LmRlZmluZVByb3BlcnR5KGV4cG9ydHMsIFwiX19lc01vZHVsZVwiLCB7XG4gIHZhbHVlOiB0cnVlXG59KTtcbmV4cG9ydHMuZGVmYXVsdCA9IHZvaWQgMDtcbnZhciBmb3JtYXREaXN0YW5jZUxvY2FsZSA9IHtcbiAgbGVzc1RoYW5YU2Vjb25kczoge1xuICAgIG9uZTogXCJtaW5kZXIgYXMgJ24gc2Vrb25kZVwiLFxuICAgIG90aGVyOiAnbWluZGVyIGFzIHt7Y291bnR9fSBzZWtvbmRlcydcbiAgfSxcbiAgeFNlY29uZHM6IHtcbiAgICBvbmU6ICcxIHNla29uZGUnLFxuICAgIG90aGVyOiAne3tjb3VudH19IHNla29uZGVzJ1xuICB9LFxuICBoYWxmQU1pbnV0ZTogXCInbiBoYWx3ZSBtaW51dXRcIixcbiAgbGVzc1RoYW5YTWludXRlczoge1xuICAgIG9uZTogXCJtaW5kZXIgYXMgJ24gbWludXV0XCIsXG4gICAgb3RoZXI6ICdtaW5kZXIgYXMge3tjb3VudH19IG1pbnV0ZSdcbiAgfSxcbiAgeE1pbnV0ZXM6IHtcbiAgICBvbmU6IFwiJ24gbWludXV0XCIsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0gbWludXRlJ1xuICB9LFxuICBhYm91dFhIb3Vyczoge1xuICAgIG9uZTogJ29uZ2V2ZWVyIDEgdXVyJyxcbiAgICBvdGhlcjogJ29uZ2V2ZWVyIHt7Y291bnR9fSB1cmUnXG4gIH0sXG4gIHhIb3Vyczoge1xuICAgIG9uZTogJzEgdXVyJyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSB1cmUnXG4gIH0sXG4gIHhEYXlzOiB7XG4gICAgb25lOiAnMSBkYWcnLFxuICAgIG90aGVyOiAne3tjb3VudH19IGRhZSdcbiAgfSxcbiAgYWJvdXRYV2Vla3M6IHtcbiAgICBvbmU6ICdvbmdldmVlciAxIHdlZWsnLFxuICAgIG90aGVyOiAnb25nZXZlZXIge3tjb3VudH19IHdla2UnXG4gIH0sXG4gIHhXZWVrczoge1xuICAgIG9uZTogJzEgd2VlaycsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0gd2VrZSdcbiAgfSxcbiAgYWJvdXRYTW9udGhzOiB7XG4gICAgb25lOiAnb25nZXZlZXIgMSBtYWFuZCcsXG4gICAgb3RoZXI6ICdvbmdldmVlciB7e2NvdW50fX0gbWFhbmRlJ1xuICB9LFxuICB4TW9udGhzOiB7XG4gICAgb25lOiAnMSBtYWFuZCcsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0gbWFhbmRlJ1xuICB9LFxuICBhYm91dFhZZWFyczoge1xuICAgIG9uZTogJ29uZ2V2ZWVyIDEgamFhcicsXG4gICAgb3RoZXI6ICdvbmdldmVlciB7e2NvdW50fX0gamFhcidcbiAgfSxcbiAgeFllYXJzOiB7XG4gICAgb25lOiAnMSBqYWFyJyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSBqYWFyJ1xuICB9LFxuICBvdmVyWFllYXJzOiB7XG4gICAgb25lOiAnbWVlciBhcyAxIGphYXInLFxuICAgIG90aGVyOiAnbWVlciBhcyB7e2NvdW50fX0gamFhcidcbiAgfSxcbiAgYWxtb3N0WFllYXJzOiB7XG4gICAgb25lOiAnYnluYSAxIGphYXInLFxuICAgIG90aGVyOiAnYnluYSB7e2NvdW50fX0gamFhcidcbiAgfVxufTtcbnZhciBmb3JtYXREaXN0YW5jZSA9IGZ1bmN0aW9uIGZvcm1hdERpc3RhbmNlKHRva2VuLCBjb3VudCwgb3B0aW9ucykge1xuICB2YXIgcmVzdWx0O1xuICB2YXIgdG9rZW5WYWx1ZSA9IGZvcm1hdERpc3RhbmNlTG9jYWxlW3Rva2VuXTtcbiAgaWYgKHR5cGVvZiB0b2tlblZhbHVlID09PSAnc3RyaW5nJykge1xuICAgIHJlc3VsdCA9IHRva2VuVmFsdWU7XG4gIH0gZWxzZSBpZiAoY291bnQgPT09IDEpIHtcbiAgICByZXN1bHQgPSB0b2tlblZhbHVlLm9uZTtcbiAgfSBlbHNlIHtcbiAgICByZXN1bHQgPSB0b2tlblZhbHVlLm90aGVyLnJlcGxhY2UoJ3t7Y291bnR9fScsIFN0cmluZyhjb3VudCkpO1xuICB9XG4gIGlmIChvcHRpb25zICE9PSBudWxsICYmIG9wdGlvbnMgIT09IHZvaWQgMCAmJiBvcHRpb25zLmFkZFN1ZmZpeCkge1xuICAgIGlmIChvcHRpb25zLmNvbXBhcmlzb24gJiYgb3B0aW9ucy5jb21wYXJpc29uID4gMCkge1xuICAgICAgcmV0dXJuICdvb3IgJyArIHJlc3VsdDtcbiAgICB9IGVsc2Uge1xuICAgICAgcmV0dXJuIHJlc3VsdCArICcgZ2VsZWRlJztcbiAgICB9XG4gIH1cbiAgcmV0dXJuIHJlc3VsdDtcbn07XG52YXIgX2RlZmF1bHQgPSBmb3JtYXREaXN0YW5jZTtcbmV4cG9ydHMuZGVmYXVsdCA9IF9kZWZhdWx0O1xubW9kdWxlLmV4cG9ydHMgPSBleHBvcnRzLmRlZmF1bHQ7Il0sIm5hbWVzIjpbXSwic291cmNlUm9vdCI6IiJ9\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/af/_lib/formatDistance/index.js\n");

/***/ })

}]);