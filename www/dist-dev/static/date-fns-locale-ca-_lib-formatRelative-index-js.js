"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-ca-_lib-formatRelative-index-js"],{

/***/ "./node_modules/date-fns/locale/ca/_lib/formatRelative/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/ca/_lib/formatRelative/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar formatRelativeLocale = {\n  lastWeek: \"'el' eeee 'passat a la' LT\",\n  yesterday: \"'ahir a la' p\",\n  today: \"'avui a la' p\",\n  tomorrow: \"'demà a la' p\",\n  nextWeek: \"eeee 'a la' p\",\n  other: 'P'\n};\nvar formatRelativeLocalePlural = {\n  lastWeek: \"'el' eeee 'passat a les' p\",\n  yesterday: \"'ahir a les' p\",\n  today: \"'avui a les' p\",\n  tomorrow: \"'demà a les' p\",\n  nextWeek: \"eeee 'a les' p\",\n  other: 'P'\n};\nvar formatRelative = function formatRelative(token, date, _baseDate, _options) {\n  if (date.getUTCHours() !== 1) {\n    return formatRelativeLocalePlural[token];\n  }\n  return formatRelativeLocale[token];\n};\nvar _default = formatRelative;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2NhL19saWIvZm9ybWF0UmVsYXRpdmUvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0Esa0JBQWU7QUFDZiIsInNvdXJjZXMiOlsid2VicGFjazovL3ByaXR1bmwtY2xvdWQvLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2NhL19saWIvZm9ybWF0UmVsYXRpdmUvaW5kZXguanM/MDRiNSJdLCJzb3VyY2VzQ29udGVudCI6WyJcInVzZSBzdHJpY3RcIjtcblxuT2JqZWN0LmRlZmluZVByb3BlcnR5KGV4cG9ydHMsIFwiX19lc01vZHVsZVwiLCB7XG4gIHZhbHVlOiB0cnVlXG59KTtcbmV4cG9ydHMuZGVmYXVsdCA9IHZvaWQgMDtcbnZhciBmb3JtYXRSZWxhdGl2ZUxvY2FsZSA9IHtcbiAgbGFzdFdlZWs6IFwiJ2VsJyBlZWVlICdwYXNzYXQgYSBsYScgTFRcIixcbiAgeWVzdGVyZGF5OiBcIidhaGlyIGEgbGEnIHBcIixcbiAgdG9kYXk6IFwiJ2F2dWkgYSBsYScgcFwiLFxuICB0b21vcnJvdzogXCInZGVtw6AgYSBsYScgcFwiLFxuICBuZXh0V2VlazogXCJlZWVlICdhIGxhJyBwXCIsXG4gIG90aGVyOiAnUCdcbn07XG52YXIgZm9ybWF0UmVsYXRpdmVMb2NhbGVQbHVyYWwgPSB7XG4gIGxhc3RXZWVrOiBcIidlbCcgZWVlZSAncGFzc2F0IGEgbGVzJyBwXCIsXG4gIHllc3RlcmRheTogXCInYWhpciBhIGxlcycgcFwiLFxuICB0b2RheTogXCInYXZ1aSBhIGxlcycgcFwiLFxuICB0b21vcnJvdzogXCInZGVtw6AgYSBsZXMnIHBcIixcbiAgbmV4dFdlZWs6IFwiZWVlZSAnYSBsZXMnIHBcIixcbiAgb3RoZXI6ICdQJ1xufTtcbnZhciBmb3JtYXRSZWxhdGl2ZSA9IGZ1bmN0aW9uIGZvcm1hdFJlbGF0aXZlKHRva2VuLCBkYXRlLCBfYmFzZURhdGUsIF9vcHRpb25zKSB7XG4gIGlmIChkYXRlLmdldFVUQ0hvdXJzKCkgIT09IDEpIHtcbiAgICByZXR1cm4gZm9ybWF0UmVsYXRpdmVMb2NhbGVQbHVyYWxbdG9rZW5dO1xuICB9XG4gIHJldHVybiBmb3JtYXRSZWxhdGl2ZUxvY2FsZVt0b2tlbl07XG59O1xudmFyIF9kZWZhdWx0ID0gZm9ybWF0UmVsYXRpdmU7XG5leHBvcnRzLmRlZmF1bHQgPSBfZGVmYXVsdDtcbm1vZHVsZS5leHBvcnRzID0gZXhwb3J0cy5kZWZhdWx0OyJdLCJuYW1lcyI6W10sInNvdXJjZVJvb3QiOiIifQ==\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/ca/_lib/formatRelative/index.js\n");

/***/ })

}]);