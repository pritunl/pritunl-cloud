"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-ht-_lib-formatDistance-index-js"],{

/***/ "./node_modules/date-fns/locale/ht/_lib/formatDistance/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/ht/_lib/formatDistance/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar formatDistanceLocale = {\n  lessThanXSeconds: {\n    one: 'mwens pase yon segond',\n    other: 'mwens pase {{count}} segond'\n  },\n  xSeconds: {\n    one: '1 segond',\n    other: '{{count}} segond'\n  },\n  halfAMinute: '30 segond',\n  lessThanXMinutes: {\n    one: 'mwens pase yon minit',\n    other: 'mwens pase {{count}} minit'\n  },\n  xMinutes: {\n    one: '1 minit',\n    other: '{{count}} minit'\n  },\n  aboutXHours: {\n    one: 'anviwon inè',\n    other: 'anviwon {{count}} è'\n  },\n  xHours: {\n    one: '1 lè',\n    other: '{{count}} lè'\n  },\n  xDays: {\n    one: '1 jou',\n    other: '{{count}} jou'\n  },\n  aboutXWeeks: {\n    one: 'anviwon 1 semèn',\n    other: 'anviwon {{count}} semèn'\n  },\n  xWeeks: {\n    one: '1 semèn',\n    other: '{{count}} semèn'\n  },\n  aboutXMonths: {\n    one: 'anviwon 1 mwa',\n    other: 'anviwon {{count}} mwa'\n  },\n  xMonths: {\n    one: '1 mwa',\n    other: '{{count}} mwa'\n  },\n  aboutXYears: {\n    one: 'anviwon 1 an',\n    other: 'anviwon {{count}} an'\n  },\n  xYears: {\n    one: '1 an',\n    other: '{{count}} an'\n  },\n  overXYears: {\n    one: 'plis pase 1 an',\n    other: 'plis pase {{count}} an'\n  },\n  almostXYears: {\n    one: 'prèske 1 an',\n    other: 'prèske {{count}} an'\n  }\n};\nvar formatDistance = function formatDistance(token, count, options) {\n  var result;\n  var tokenValue = formatDistanceLocale[token];\n  if (typeof tokenValue === 'string') {\n    result = tokenValue;\n  } else if (count === 1) {\n    result = tokenValue.one;\n  } else {\n    result = tokenValue.other.replace('{{count}}', String(count));\n  }\n  if (options !== null && options !== void 0 && options.addSuffix) {\n    if (options.comparison && options.comparison > 0) {\n      return 'nan ' + result;\n    } else {\n      return 'sa fè ' + result;\n    }\n  }\n  return result;\n};\nvar _default = formatDistance;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2h0L19saWIvZm9ybWF0RGlzdGFuY2UvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQSx5QkFBeUIsUUFBUTtBQUNqQyxHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBO0FBQ0EseUJBQXlCLFFBQVE7QUFDakMsR0FBRztBQUNIO0FBQ0E7QUFDQSxjQUFjLFFBQVE7QUFDdEIsR0FBRztBQUNIO0FBQ0E7QUFDQSxzQkFBc0IsUUFBUTtBQUM5QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLHNCQUFzQixRQUFRO0FBQzlCLEdBQUc7QUFDSDtBQUNBO0FBQ0EsY0FBYyxRQUFRO0FBQ3RCLEdBQUc7QUFDSDtBQUNBO0FBQ0Esc0JBQXNCLFFBQVE7QUFDOUIsR0FBRztBQUNIO0FBQ0E7QUFDQSxjQUFjLFFBQVE7QUFDdEIsR0FBRztBQUNIO0FBQ0E7QUFDQSxzQkFBc0IsUUFBUTtBQUM5QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLHdCQUF3QixRQUFRO0FBQ2hDLEdBQUc7QUFDSDtBQUNBO0FBQ0EscUJBQXFCLFFBQVE7QUFDN0I7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxJQUFJO0FBQ0o7QUFDQSxJQUFJO0FBQ0oseUNBQXlDLE9BQU87QUFDaEQ7QUFDQTtBQUNBO0FBQ0E7QUFDQSxNQUFNO0FBQ047QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0Esa0JBQWU7QUFDZiIsInNvdXJjZXMiOlsid2VicGFjazovL3ByaXR1bmwtY2xvdWQvLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2h0L19saWIvZm9ybWF0RGlzdGFuY2UvaW5kZXguanM/MmY1YSJdLCJzb3VyY2VzQ29udGVudCI6WyJcInVzZSBzdHJpY3RcIjtcblxuT2JqZWN0LmRlZmluZVByb3BlcnR5KGV4cG9ydHMsIFwiX19lc01vZHVsZVwiLCB7XG4gIHZhbHVlOiB0cnVlXG59KTtcbmV4cG9ydHMuZGVmYXVsdCA9IHZvaWQgMDtcbnZhciBmb3JtYXREaXN0YW5jZUxvY2FsZSA9IHtcbiAgbGVzc1RoYW5YU2Vjb25kczoge1xuICAgIG9uZTogJ213ZW5zIHBhc2UgeW9uIHNlZ29uZCcsXG4gICAgb3RoZXI6ICdtd2VucyBwYXNlIHt7Y291bnR9fSBzZWdvbmQnXG4gIH0sXG4gIHhTZWNvbmRzOiB7XG4gICAgb25lOiAnMSBzZWdvbmQnLFxuICAgIG90aGVyOiAne3tjb3VudH19IHNlZ29uZCdcbiAgfSxcbiAgaGFsZkFNaW51dGU6ICczMCBzZWdvbmQnLFxuICBsZXNzVGhhblhNaW51dGVzOiB7XG4gICAgb25lOiAnbXdlbnMgcGFzZSB5b24gbWluaXQnLFxuICAgIG90aGVyOiAnbXdlbnMgcGFzZSB7e2NvdW50fX0gbWluaXQnXG4gIH0sXG4gIHhNaW51dGVzOiB7XG4gICAgb25lOiAnMSBtaW5pdCcsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0gbWluaXQnXG4gIH0sXG4gIGFib3V0WEhvdXJzOiB7XG4gICAgb25lOiAnYW52aXdvbiBpbsOoJyxcbiAgICBvdGhlcjogJ2Fudml3b24ge3tjb3VudH19IMOoJ1xuICB9LFxuICB4SG91cnM6IHtcbiAgICBvbmU6ICcxIGzDqCcsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0gbMOoJ1xuICB9LFxuICB4RGF5czoge1xuICAgIG9uZTogJzEgam91JyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSBqb3UnXG4gIH0sXG4gIGFib3V0WFdlZWtzOiB7XG4gICAgb25lOiAnYW52aXdvbiAxIHNlbcOobicsXG4gICAgb3RoZXI6ICdhbnZpd29uIHt7Y291bnR9fSBzZW3DqG4nXG4gIH0sXG4gIHhXZWVrczoge1xuICAgIG9uZTogJzEgc2Vtw6huJyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSBzZW3DqG4nXG4gIH0sXG4gIGFib3V0WE1vbnRoczoge1xuICAgIG9uZTogJ2Fudml3b24gMSBtd2EnLFxuICAgIG90aGVyOiAnYW52aXdvbiB7e2NvdW50fX0gbXdhJ1xuICB9LFxuICB4TW9udGhzOiB7XG4gICAgb25lOiAnMSBtd2EnLFxuICAgIG90aGVyOiAne3tjb3VudH19IG13YSdcbiAgfSxcbiAgYWJvdXRYWWVhcnM6IHtcbiAgICBvbmU6ICdhbnZpd29uIDEgYW4nLFxuICAgIG90aGVyOiAnYW52aXdvbiB7e2NvdW50fX0gYW4nXG4gIH0sXG4gIHhZZWFyczoge1xuICAgIG9uZTogJzEgYW4nLFxuICAgIG90aGVyOiAne3tjb3VudH19IGFuJ1xuICB9LFxuICBvdmVyWFllYXJzOiB7XG4gICAgb25lOiAncGxpcyBwYXNlIDEgYW4nLFxuICAgIG90aGVyOiAncGxpcyBwYXNlIHt7Y291bnR9fSBhbidcbiAgfSxcbiAgYWxtb3N0WFllYXJzOiB7XG4gICAgb25lOiAncHLDqHNrZSAxIGFuJyxcbiAgICBvdGhlcjogJ3Byw6hza2Uge3tjb3VudH19IGFuJ1xuICB9XG59O1xudmFyIGZvcm1hdERpc3RhbmNlID0gZnVuY3Rpb24gZm9ybWF0RGlzdGFuY2UodG9rZW4sIGNvdW50LCBvcHRpb25zKSB7XG4gIHZhciByZXN1bHQ7XG4gIHZhciB0b2tlblZhbHVlID0gZm9ybWF0RGlzdGFuY2VMb2NhbGVbdG9rZW5dO1xuICBpZiAodHlwZW9mIHRva2VuVmFsdWUgPT09ICdzdHJpbmcnKSB7XG4gICAgcmVzdWx0ID0gdG9rZW5WYWx1ZTtcbiAgfSBlbHNlIGlmIChjb3VudCA9PT0gMSkge1xuICAgIHJlc3VsdCA9IHRva2VuVmFsdWUub25lO1xuICB9IGVsc2Uge1xuICAgIHJlc3VsdCA9IHRva2VuVmFsdWUub3RoZXIucmVwbGFjZSgne3tjb3VudH19JywgU3RyaW5nKGNvdW50KSk7XG4gIH1cbiAgaWYgKG9wdGlvbnMgIT09IG51bGwgJiYgb3B0aW9ucyAhPT0gdm9pZCAwICYmIG9wdGlvbnMuYWRkU3VmZml4KSB7XG4gICAgaWYgKG9wdGlvbnMuY29tcGFyaXNvbiAmJiBvcHRpb25zLmNvbXBhcmlzb24gPiAwKSB7XG4gICAgICByZXR1cm4gJ25hbiAnICsgcmVzdWx0O1xuICAgIH0gZWxzZSB7XG4gICAgICByZXR1cm4gJ3NhIGbDqCAnICsgcmVzdWx0O1xuICAgIH1cbiAgfVxuICByZXR1cm4gcmVzdWx0O1xufTtcbnZhciBfZGVmYXVsdCA9IGZvcm1hdERpc3RhbmNlO1xuZXhwb3J0cy5kZWZhdWx0ID0gX2RlZmF1bHQ7XG5tb2R1bGUuZXhwb3J0cyA9IGV4cG9ydHMuZGVmYXVsdDsiXSwibmFtZXMiOltdLCJzb3VyY2VSb290IjoiIn0=\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/ht/_lib/formatDistance/index.js\n");

/***/ })

}]);