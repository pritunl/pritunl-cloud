"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-es-_lib-formatDistance-index-js"],{

/***/ "./node_modules/date-fns/locale/es/_lib/formatDistance/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/es/_lib/formatDistance/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar formatDistanceLocale = {\n  lessThanXSeconds: {\n    one: 'menos de un segundo',\n    other: 'menos de {{count}} segundos'\n  },\n  xSeconds: {\n    one: '1 segundo',\n    other: '{{count}} segundos'\n  },\n  halfAMinute: 'medio minuto',\n  lessThanXMinutes: {\n    one: 'menos de un minuto',\n    other: 'menos de {{count}} minutos'\n  },\n  xMinutes: {\n    one: '1 minuto',\n    other: '{{count}} minutos'\n  },\n  aboutXHours: {\n    one: 'alrededor de 1 hora',\n    other: 'alrededor de {{count}} horas'\n  },\n  xHours: {\n    one: '1 hora',\n    other: '{{count}} horas'\n  },\n  xDays: {\n    one: '1 día',\n    other: '{{count}} días'\n  },\n  aboutXWeeks: {\n    one: 'alrededor de 1 semana',\n    other: 'alrededor de {{count}} semanas'\n  },\n  xWeeks: {\n    one: '1 semana',\n    other: '{{count}} semanas'\n  },\n  aboutXMonths: {\n    one: 'alrededor de 1 mes',\n    other: 'alrededor de {{count}} meses'\n  },\n  xMonths: {\n    one: '1 mes',\n    other: '{{count}} meses'\n  },\n  aboutXYears: {\n    one: 'alrededor de 1 año',\n    other: 'alrededor de {{count}} años'\n  },\n  xYears: {\n    one: '1 año',\n    other: '{{count}} años'\n  },\n  overXYears: {\n    one: 'más de 1 año',\n    other: 'más de {{count}} años'\n  },\n  almostXYears: {\n    one: 'casi 1 año',\n    other: 'casi {{count}} años'\n  }\n};\nvar formatDistance = function formatDistance(token, count, options) {\n  var result;\n  var tokenValue = formatDistanceLocale[token];\n  if (typeof tokenValue === 'string') {\n    result = tokenValue;\n  } else if (count === 1) {\n    result = tokenValue.one;\n  } else {\n    result = tokenValue.other.replace('{{count}}', count.toString());\n  }\n  if (options !== null && options !== void 0 && options.addSuffix) {\n    if (options.comparison && options.comparison > 0) {\n      return 'en ' + result;\n    } else {\n      return 'hace ' + result;\n    }\n  }\n  return result;\n};\nvar _default = formatDistance;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2VzL19saWIvZm9ybWF0RGlzdGFuY2UvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQSx1QkFBdUIsUUFBUTtBQUMvQixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBO0FBQ0EsdUJBQXVCLFFBQVE7QUFDL0IsR0FBRztBQUNIO0FBQ0E7QUFDQSxjQUFjLFFBQVE7QUFDdEIsR0FBRztBQUNIO0FBQ0E7QUFDQSwyQkFBMkIsUUFBUTtBQUNuQyxHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLDJCQUEyQixRQUFRO0FBQ25DLEdBQUc7QUFDSDtBQUNBO0FBQ0EsY0FBYyxRQUFRO0FBQ3RCLEdBQUc7QUFDSDtBQUNBO0FBQ0EsMkJBQTJCLFFBQVE7QUFDbkMsR0FBRztBQUNIO0FBQ0E7QUFDQSxjQUFjLFFBQVE7QUFDdEIsR0FBRztBQUNIO0FBQ0E7QUFDQSwyQkFBMkIsUUFBUTtBQUNuQyxHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLHFCQUFxQixRQUFRO0FBQzdCLEdBQUc7QUFDSDtBQUNBO0FBQ0EsbUJBQW1CLFFBQVE7QUFDM0I7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxJQUFJO0FBQ0o7QUFDQSxJQUFJO0FBQ0oseUNBQXlDLE9BQU87QUFDaEQ7QUFDQTtBQUNBO0FBQ0E7QUFDQSxNQUFNO0FBQ047QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0Esa0JBQWU7QUFDZiIsInNvdXJjZXMiOlsid2VicGFjazovL3ByaXR1bmwtY2xvdWQvLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2VzL19saWIvZm9ybWF0RGlzdGFuY2UvaW5kZXguanM/ZTQyMiJdLCJzb3VyY2VzQ29udGVudCI6WyJcInVzZSBzdHJpY3RcIjtcblxuT2JqZWN0LmRlZmluZVByb3BlcnR5KGV4cG9ydHMsIFwiX19lc01vZHVsZVwiLCB7XG4gIHZhbHVlOiB0cnVlXG59KTtcbmV4cG9ydHMuZGVmYXVsdCA9IHZvaWQgMDtcbnZhciBmb3JtYXREaXN0YW5jZUxvY2FsZSA9IHtcbiAgbGVzc1RoYW5YU2Vjb25kczoge1xuICAgIG9uZTogJ21lbm9zIGRlIHVuIHNlZ3VuZG8nLFxuICAgIG90aGVyOiAnbWVub3MgZGUge3tjb3VudH19IHNlZ3VuZG9zJ1xuICB9LFxuICB4U2Vjb25kczoge1xuICAgIG9uZTogJzEgc2VndW5kbycsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0gc2VndW5kb3MnXG4gIH0sXG4gIGhhbGZBTWludXRlOiAnbWVkaW8gbWludXRvJyxcbiAgbGVzc1RoYW5YTWludXRlczoge1xuICAgIG9uZTogJ21lbm9zIGRlIHVuIG1pbnV0bycsXG4gICAgb3RoZXI6ICdtZW5vcyBkZSB7e2NvdW50fX0gbWludXRvcydcbiAgfSxcbiAgeE1pbnV0ZXM6IHtcbiAgICBvbmU6ICcxIG1pbnV0bycsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0gbWludXRvcydcbiAgfSxcbiAgYWJvdXRYSG91cnM6IHtcbiAgICBvbmU6ICdhbHJlZGVkb3IgZGUgMSBob3JhJyxcbiAgICBvdGhlcjogJ2FscmVkZWRvciBkZSB7e2NvdW50fX0gaG9yYXMnXG4gIH0sXG4gIHhIb3Vyczoge1xuICAgIG9uZTogJzEgaG9yYScsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0gaG9yYXMnXG4gIH0sXG4gIHhEYXlzOiB7XG4gICAgb25lOiAnMSBkw61hJyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSBkw61hcydcbiAgfSxcbiAgYWJvdXRYV2Vla3M6IHtcbiAgICBvbmU6ICdhbHJlZGVkb3IgZGUgMSBzZW1hbmEnLFxuICAgIG90aGVyOiAnYWxyZWRlZG9yIGRlIHt7Y291bnR9fSBzZW1hbmFzJ1xuICB9LFxuICB4V2Vla3M6IHtcbiAgICBvbmU6ICcxIHNlbWFuYScsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0gc2VtYW5hcydcbiAgfSxcbiAgYWJvdXRYTW9udGhzOiB7XG4gICAgb25lOiAnYWxyZWRlZG9yIGRlIDEgbWVzJyxcbiAgICBvdGhlcjogJ2FscmVkZWRvciBkZSB7e2NvdW50fX0gbWVzZXMnXG4gIH0sXG4gIHhNb250aHM6IHtcbiAgICBvbmU6ICcxIG1lcycsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0gbWVzZXMnXG4gIH0sXG4gIGFib3V0WFllYXJzOiB7XG4gICAgb25lOiAnYWxyZWRlZG9yIGRlIDEgYcOxbycsXG4gICAgb3RoZXI6ICdhbHJlZGVkb3IgZGUge3tjb3VudH19IGHDsW9zJ1xuICB9LFxuICB4WWVhcnM6IHtcbiAgICBvbmU6ICcxIGHDsW8nLFxuICAgIG90aGVyOiAne3tjb3VudH19IGHDsW9zJ1xuICB9LFxuICBvdmVyWFllYXJzOiB7XG4gICAgb25lOiAnbcOhcyBkZSAxIGHDsW8nLFxuICAgIG90aGVyOiAnbcOhcyBkZSB7e2NvdW50fX0gYcOxb3MnXG4gIH0sXG4gIGFsbW9zdFhZZWFyczoge1xuICAgIG9uZTogJ2Nhc2kgMSBhw7FvJyxcbiAgICBvdGhlcjogJ2Nhc2kge3tjb3VudH19IGHDsW9zJ1xuICB9XG59O1xudmFyIGZvcm1hdERpc3RhbmNlID0gZnVuY3Rpb24gZm9ybWF0RGlzdGFuY2UodG9rZW4sIGNvdW50LCBvcHRpb25zKSB7XG4gIHZhciByZXN1bHQ7XG4gIHZhciB0b2tlblZhbHVlID0gZm9ybWF0RGlzdGFuY2VMb2NhbGVbdG9rZW5dO1xuICBpZiAodHlwZW9mIHRva2VuVmFsdWUgPT09ICdzdHJpbmcnKSB7XG4gICAgcmVzdWx0ID0gdG9rZW5WYWx1ZTtcbiAgfSBlbHNlIGlmIChjb3VudCA9PT0gMSkge1xuICAgIHJlc3VsdCA9IHRva2VuVmFsdWUub25lO1xuICB9IGVsc2Uge1xuICAgIHJlc3VsdCA9IHRva2VuVmFsdWUub3RoZXIucmVwbGFjZSgne3tjb3VudH19JywgY291bnQudG9TdHJpbmcoKSk7XG4gIH1cbiAgaWYgKG9wdGlvbnMgIT09IG51bGwgJiYgb3B0aW9ucyAhPT0gdm9pZCAwICYmIG9wdGlvbnMuYWRkU3VmZml4KSB7XG4gICAgaWYgKG9wdGlvbnMuY29tcGFyaXNvbiAmJiBvcHRpb25zLmNvbXBhcmlzb24gPiAwKSB7XG4gICAgICByZXR1cm4gJ2VuICcgKyByZXN1bHQ7XG4gICAgfSBlbHNlIHtcbiAgICAgIHJldHVybiAnaGFjZSAnICsgcmVzdWx0O1xuICAgIH1cbiAgfVxuICByZXR1cm4gcmVzdWx0O1xufTtcbnZhciBfZGVmYXVsdCA9IGZvcm1hdERpc3RhbmNlO1xuZXhwb3J0cy5kZWZhdWx0ID0gX2RlZmF1bHQ7XG5tb2R1bGUuZXhwb3J0cyA9IGV4cG9ydHMuZGVmYXVsdDsiXSwibmFtZXMiOltdLCJzb3VyY2VSb290IjoiIn0=\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/es/_lib/formatDistance/index.js\n");

/***/ })

}]);