"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-gl-_lib-formatRelative-index-js"],{

/***/ "./node_modules/date-fns/locale/gl/_lib/formatRelative/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/gl/_lib/formatRelative/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar formatRelativeLocale = {\n  lastWeek: \"'o' eeee 'pasado á' LT\",\n  yesterday: \"'onte á' p\",\n  today: \"'hoxe á' p\",\n  tomorrow: \"'mañá á' p\",\n  nextWeek: \"eeee 'á' p\",\n  other: 'P'\n};\nvar formatRelativeLocalePlural = {\n  lastWeek: \"'o' eeee 'pasado ás' p\",\n  yesterday: \"'onte ás' p\",\n  today: \"'hoxe ás' p\",\n  tomorrow: \"'mañá ás' p\",\n  nextWeek: \"eeee 'ás' p\",\n  other: 'P'\n};\nvar formatRelative = function formatRelative(token, date, _baseDate, _options) {\n  if (date.getUTCHours() !== 1) {\n    return formatRelativeLocalePlural[token];\n  }\n  return formatRelativeLocale[token];\n};\nvar _default = formatRelative;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2dsL19saWIvZm9ybWF0UmVsYXRpdmUvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0Esa0JBQWU7QUFDZiIsInNvdXJjZXMiOlsid2VicGFjazovL3ByaXR1bmwtY2xvdWQvLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2dsL19saWIvZm9ybWF0UmVsYXRpdmUvaW5kZXguanM/MzBjNiJdLCJzb3VyY2VzQ29udGVudCI6WyJcInVzZSBzdHJpY3RcIjtcblxuT2JqZWN0LmRlZmluZVByb3BlcnR5KGV4cG9ydHMsIFwiX19lc01vZHVsZVwiLCB7XG4gIHZhbHVlOiB0cnVlXG59KTtcbmV4cG9ydHMuZGVmYXVsdCA9IHZvaWQgMDtcbnZhciBmb3JtYXRSZWxhdGl2ZUxvY2FsZSA9IHtcbiAgbGFzdFdlZWs6IFwiJ28nIGVlZWUgJ3Bhc2FkbyDDoScgTFRcIixcbiAgeWVzdGVyZGF5OiBcIidvbnRlIMOhJyBwXCIsXG4gIHRvZGF5OiBcIidob3hlIMOhJyBwXCIsXG4gIHRvbW9ycm93OiBcIidtYcOxw6Egw6EnIHBcIixcbiAgbmV4dFdlZWs6IFwiZWVlZSAnw6EnIHBcIixcbiAgb3RoZXI6ICdQJ1xufTtcbnZhciBmb3JtYXRSZWxhdGl2ZUxvY2FsZVBsdXJhbCA9IHtcbiAgbGFzdFdlZWs6IFwiJ28nIGVlZWUgJ3Bhc2FkbyDDoXMnIHBcIixcbiAgeWVzdGVyZGF5OiBcIidvbnRlIMOhcycgcFwiLFxuICB0b2RheTogXCInaG94ZSDDoXMnIHBcIixcbiAgdG9tb3Jyb3c6IFwiJ21hw7HDoSDDoXMnIHBcIixcbiAgbmV4dFdlZWs6IFwiZWVlZSAnw6FzJyBwXCIsXG4gIG90aGVyOiAnUCdcbn07XG52YXIgZm9ybWF0UmVsYXRpdmUgPSBmdW5jdGlvbiBmb3JtYXRSZWxhdGl2ZSh0b2tlbiwgZGF0ZSwgX2Jhc2VEYXRlLCBfb3B0aW9ucykge1xuICBpZiAoZGF0ZS5nZXRVVENIb3VycygpICE9PSAxKSB7XG4gICAgcmV0dXJuIGZvcm1hdFJlbGF0aXZlTG9jYWxlUGx1cmFsW3Rva2VuXTtcbiAgfVxuICByZXR1cm4gZm9ybWF0UmVsYXRpdmVMb2NhbGVbdG9rZW5dO1xufTtcbnZhciBfZGVmYXVsdCA9IGZvcm1hdFJlbGF0aXZlO1xuZXhwb3J0cy5kZWZhdWx0ID0gX2RlZmF1bHQ7XG5tb2R1bGUuZXhwb3J0cyA9IGV4cG9ydHMuZGVmYXVsdDsiXSwibmFtZXMiOltdLCJzb3VyY2VSb290IjoiIn0=\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/gl/_lib/formatRelative/index.js\n");

/***/ })

}]);