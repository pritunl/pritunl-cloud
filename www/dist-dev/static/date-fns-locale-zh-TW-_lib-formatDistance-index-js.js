"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-zh-TW-_lib-formatDistance-index-js"],{

/***/ "./node_modules/date-fns/locale/zh-TW/_lib/formatDistance/index.js":
/*!*************************************************************************!*\
  !*** ./node_modules/date-fns/locale/zh-TW/_lib/formatDistance/index.js ***!
  \*************************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar formatDistanceLocale = {\n  lessThanXSeconds: {\n    one: '少於 1 秒',\n    other: '少於 {{count}} 秒'\n  },\n  xSeconds: {\n    one: '1 秒',\n    other: '{{count}} 秒'\n  },\n  halfAMinute: '半分鐘',\n  lessThanXMinutes: {\n    one: '少於 1 分鐘',\n    other: '少於 {{count}} 分鐘'\n  },\n  xMinutes: {\n    one: '1 分鐘',\n    other: '{{count}} 分鐘'\n  },\n  xHours: {\n    one: '1 小時',\n    other: '{{count}} 小時'\n  },\n  aboutXHours: {\n    one: '大約 1 小時',\n    other: '大約 {{count}} 小時'\n  },\n  xDays: {\n    one: '1 天',\n    other: '{{count}} 天'\n  },\n  aboutXWeeks: {\n    one: '大約 1 個星期',\n    other: '大約 {{count}} 個星期'\n  },\n  xWeeks: {\n    one: '1 個星期',\n    other: '{{count}} 個星期'\n  },\n  aboutXMonths: {\n    one: '大約 1 個月',\n    other: '大約 {{count}} 個月'\n  },\n  xMonths: {\n    one: '1 個月',\n    other: '{{count}} 個月'\n  },\n  aboutXYears: {\n    one: '大約 1 年',\n    other: '大約 {{count}} 年'\n  },\n  xYears: {\n    one: '1 年',\n    other: '{{count}} 年'\n  },\n  overXYears: {\n    one: '超過 1 年',\n    other: '超過 {{count}} 年'\n  },\n  almostXYears: {\n    one: '將近 1 年',\n    other: '將近 {{count}} 年'\n  }\n};\nvar formatDistance = function formatDistance(token, count, options) {\n  var result;\n  var tokenValue = formatDistanceLocale[token];\n  if (typeof tokenValue === 'string') {\n    result = tokenValue;\n  } else if (count === 1) {\n    result = tokenValue.one;\n  } else {\n    result = tokenValue.other.replace('{{count}}', String(count));\n  }\n  if (options !== null && options !== void 0 && options.addSuffix) {\n    if (options.comparison && options.comparison > 0) {\n      return result + '內';\n    } else {\n      return result + '前';\n    }\n  }\n  return result;\n};\nvar _default = formatDistance;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL3poLVRXL19saWIvZm9ybWF0RGlzdGFuY2UvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQSxpQkFBaUIsUUFBUTtBQUN6QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBO0FBQ0EsaUJBQWlCLFFBQVE7QUFDekIsR0FBRztBQUNIO0FBQ0E7QUFDQSxjQUFjLFFBQVE7QUFDdEIsR0FBRztBQUNIO0FBQ0E7QUFDQSxjQUFjLFFBQVE7QUFDdEIsR0FBRztBQUNIO0FBQ0E7QUFDQSxpQkFBaUIsUUFBUTtBQUN6QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGlCQUFpQixRQUFRO0FBQ3pCLEdBQUc7QUFDSDtBQUNBO0FBQ0EsY0FBYyxRQUFRO0FBQ3RCLEdBQUc7QUFDSDtBQUNBO0FBQ0EsaUJBQWlCLFFBQVE7QUFDekIsR0FBRztBQUNIO0FBQ0E7QUFDQSxjQUFjLFFBQVE7QUFDdEIsR0FBRztBQUNIO0FBQ0E7QUFDQSxpQkFBaUIsUUFBUTtBQUN6QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGNBQWMsUUFBUTtBQUN0QixHQUFHO0FBQ0g7QUFDQTtBQUNBLGlCQUFpQixRQUFRO0FBQ3pCLEdBQUc7QUFDSDtBQUNBO0FBQ0EsaUJBQWlCLFFBQVE7QUFDekI7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxJQUFJO0FBQ0o7QUFDQSxJQUFJO0FBQ0oseUNBQXlDLE9BQU87QUFDaEQ7QUFDQTtBQUNBO0FBQ0E7QUFDQSxNQUFNO0FBQ047QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0Esa0JBQWU7QUFDZiIsInNvdXJjZXMiOlsid2VicGFjazovL3ByaXR1bmwtY2xvdWQvLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL3poLVRXL19saWIvZm9ybWF0RGlzdGFuY2UvaW5kZXguanM/M2Y2MSJdLCJzb3VyY2VzQ29udGVudCI6WyJcInVzZSBzdHJpY3RcIjtcblxuT2JqZWN0LmRlZmluZVByb3BlcnR5KGV4cG9ydHMsIFwiX19lc01vZHVsZVwiLCB7XG4gIHZhbHVlOiB0cnVlXG59KTtcbmV4cG9ydHMuZGVmYXVsdCA9IHZvaWQgMDtcbnZhciBmb3JtYXREaXN0YW5jZUxvY2FsZSA9IHtcbiAgbGVzc1RoYW5YU2Vjb25kczoge1xuICAgIG9uZTogJ+WwkeaWvCAxIOenkicsXG4gICAgb3RoZXI6ICflsJHmlrwge3tjb3VudH19IOenkidcbiAgfSxcbiAgeFNlY29uZHM6IHtcbiAgICBvbmU6ICcxIOenkicsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0g56eSJ1xuICB9LFxuICBoYWxmQU1pbnV0ZTogJ+WNiuWIhumQmCcsXG4gIGxlc3NUaGFuWE1pbnV0ZXM6IHtcbiAgICBvbmU6ICflsJHmlrwgMSDliIbpkJgnLFxuICAgIG90aGVyOiAn5bCR5pa8IHt7Y291bnR9fSDliIbpkJgnXG4gIH0sXG4gIHhNaW51dGVzOiB7XG4gICAgb25lOiAnMSDliIbpkJgnLFxuICAgIG90aGVyOiAne3tjb3VudH19IOWIhumQmCdcbiAgfSxcbiAgeEhvdXJzOiB7XG4gICAgb25lOiAnMSDlsI/mmYInLFxuICAgIG90aGVyOiAne3tjb3VudH19IOWwj+aZgidcbiAgfSxcbiAgYWJvdXRYSG91cnM6IHtcbiAgICBvbmU6ICflpKfntIQgMSDlsI/mmYInLFxuICAgIG90aGVyOiAn5aSn57SEIHt7Y291bnR9fSDlsI/mmYInXG4gIH0sXG4gIHhEYXlzOiB7XG4gICAgb25lOiAnMSDlpKknLFxuICAgIG90aGVyOiAne3tjb3VudH19IOWkqSdcbiAgfSxcbiAgYWJvdXRYV2Vla3M6IHtcbiAgICBvbmU6ICflpKfntIQgMSDlgIvmmJ/mnJ8nLFxuICAgIG90aGVyOiAn5aSn57SEIHt7Y291bnR9fSDlgIvmmJ/mnJ8nXG4gIH0sXG4gIHhXZWVrczoge1xuICAgIG9uZTogJzEg5YCL5pif5pyfJyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSDlgIvmmJ/mnJ8nXG4gIH0sXG4gIGFib3V0WE1vbnRoczoge1xuICAgIG9uZTogJ+Wkp+e0hCAxIOWAi+aciCcsXG4gICAgb3RoZXI6ICflpKfntIQge3tjb3VudH19IOWAi+aciCdcbiAgfSxcbiAgeE1vbnRoczoge1xuICAgIG9uZTogJzEg5YCL5pyIJyxcbiAgICBvdGhlcjogJ3t7Y291bnR9fSDlgIvmnIgnXG4gIH0sXG4gIGFib3V0WFllYXJzOiB7XG4gICAgb25lOiAn5aSn57SEIDEg5bm0JyxcbiAgICBvdGhlcjogJ+Wkp+e0hCB7e2NvdW50fX0g5bm0J1xuICB9LFxuICB4WWVhcnM6IHtcbiAgICBvbmU6ICcxIOW5tCcsXG4gICAgb3RoZXI6ICd7e2NvdW50fX0g5bm0J1xuICB9LFxuICBvdmVyWFllYXJzOiB7XG4gICAgb25lOiAn6LaF6YGOIDEg5bm0JyxcbiAgICBvdGhlcjogJ+i2hemBjiB7e2NvdW50fX0g5bm0J1xuICB9LFxuICBhbG1vc3RYWWVhcnM6IHtcbiAgICBvbmU6ICflsIfov5EgMSDlubQnLFxuICAgIG90aGVyOiAn5bCH6L+RIHt7Y291bnR9fSDlubQnXG4gIH1cbn07XG52YXIgZm9ybWF0RGlzdGFuY2UgPSBmdW5jdGlvbiBmb3JtYXREaXN0YW5jZSh0b2tlbiwgY291bnQsIG9wdGlvbnMpIHtcbiAgdmFyIHJlc3VsdDtcbiAgdmFyIHRva2VuVmFsdWUgPSBmb3JtYXREaXN0YW5jZUxvY2FsZVt0b2tlbl07XG4gIGlmICh0eXBlb2YgdG9rZW5WYWx1ZSA9PT0gJ3N0cmluZycpIHtcbiAgICByZXN1bHQgPSB0b2tlblZhbHVlO1xuICB9IGVsc2UgaWYgKGNvdW50ID09PSAxKSB7XG4gICAgcmVzdWx0ID0gdG9rZW5WYWx1ZS5vbmU7XG4gIH0gZWxzZSB7XG4gICAgcmVzdWx0ID0gdG9rZW5WYWx1ZS5vdGhlci5yZXBsYWNlKCd7e2NvdW50fX0nLCBTdHJpbmcoY291bnQpKTtcbiAgfVxuICBpZiAob3B0aW9ucyAhPT0gbnVsbCAmJiBvcHRpb25zICE9PSB2b2lkIDAgJiYgb3B0aW9ucy5hZGRTdWZmaXgpIHtcbiAgICBpZiAob3B0aW9ucy5jb21wYXJpc29uICYmIG9wdGlvbnMuY29tcGFyaXNvbiA+IDApIHtcbiAgICAgIHJldHVybiByZXN1bHQgKyAn5YWnJztcbiAgICB9IGVsc2Uge1xuICAgICAgcmV0dXJuIHJlc3VsdCArICfliY0nO1xuICAgIH1cbiAgfVxuICByZXR1cm4gcmVzdWx0O1xufTtcbnZhciBfZGVmYXVsdCA9IGZvcm1hdERpc3RhbmNlO1xuZXhwb3J0cy5kZWZhdWx0ID0gX2RlZmF1bHQ7XG5tb2R1bGUuZXhwb3J0cyA9IGV4cG9ydHMuZGVmYXVsdDsiXSwibmFtZXMiOltdLCJzb3VyY2VSb290IjoiIn0=\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/zh-TW/_lib/formatDistance/index.js\n");

/***/ })

}]);