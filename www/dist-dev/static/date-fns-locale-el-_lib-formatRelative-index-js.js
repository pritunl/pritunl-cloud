"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-el-_lib-formatRelative-index-js"],{

/***/ "./node_modules/date-fns/locale/el/_lib/formatRelative/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/el/_lib/formatRelative/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar formatRelativeLocale = {\n  lastWeek: function lastWeek(date) {\n    switch (date.getUTCDay()) {\n      case 6:\n        //Σάββατο\n        return \"'το προηγούμενο' eeee 'στις' p\";\n      default:\n        return \"'την προηγούμενη' eeee 'στις' p\";\n    }\n  },\n  yesterday: \"'χθες στις' p\",\n  today: \"'σήμερα στις' p\",\n  tomorrow: \"'αύριο στις' p\",\n  nextWeek: \"eeee 'στις' p\",\n  other: 'P'\n};\nvar formatRelative = function formatRelative(token, date) {\n  var format = formatRelativeLocale[token];\n  if (typeof format === 'function') return format(date);\n  return format;\n};\nvar _default = formatRelative;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2VsL19saWIvZm9ybWF0UmVsYXRpdmUvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxHQUFHO0FBQ0g7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0Esa0JBQWU7QUFDZiIsInNvdXJjZXMiOlsid2VicGFjazovL3ByaXR1bmwtY2xvdWQvLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2VsL19saWIvZm9ybWF0UmVsYXRpdmUvaW5kZXguanM/MjNmOCJdLCJzb3VyY2VzQ29udGVudCI6WyJcInVzZSBzdHJpY3RcIjtcblxuT2JqZWN0LmRlZmluZVByb3BlcnR5KGV4cG9ydHMsIFwiX19lc01vZHVsZVwiLCB7XG4gIHZhbHVlOiB0cnVlXG59KTtcbmV4cG9ydHMuZGVmYXVsdCA9IHZvaWQgMDtcbnZhciBmb3JtYXRSZWxhdGl2ZUxvY2FsZSA9IHtcbiAgbGFzdFdlZWs6IGZ1bmN0aW9uIGxhc3RXZWVrKGRhdGUpIHtcbiAgICBzd2l0Y2ggKGRhdGUuZ2V0VVRDRGF5KCkpIHtcbiAgICAgIGNhc2UgNjpcbiAgICAgICAgLy/Oo86szrLOss6xz4TOv1xuICAgICAgICByZXR1cm4gXCInz4TOvyDPgM+Bzr/Ot86zzr/Pjc68zrXOvc6/JyBlZWVlICfPg8+EzrnPgicgcFwiO1xuICAgICAgZGVmYXVsdDpcbiAgICAgICAgcmV0dXJuIFwiJ8+EzrfOvSDPgM+Bzr/Ot86zzr/Pjc68zrXOvc63JyBlZWVlICfPg8+EzrnPgicgcFwiO1xuICAgIH1cbiAgfSxcbiAgeWVzdGVyZGF5OiBcIifPh864zrXPgiDPg8+EzrnPgicgcFwiLFxuICB0b2RheTogXCInz4POrs68zrXPgc6xIM+Dz4TOuc+CJyBwXCIsXG4gIHRvbW9ycm93OiBcIifOsc+Nz4HOuc6/IM+Dz4TOuc+CJyBwXCIsXG4gIG5leHRXZWVrOiBcImVlZWUgJ8+Dz4TOuc+CJyBwXCIsXG4gIG90aGVyOiAnUCdcbn07XG52YXIgZm9ybWF0UmVsYXRpdmUgPSBmdW5jdGlvbiBmb3JtYXRSZWxhdGl2ZSh0b2tlbiwgZGF0ZSkge1xuICB2YXIgZm9ybWF0ID0gZm9ybWF0UmVsYXRpdmVMb2NhbGVbdG9rZW5dO1xuICBpZiAodHlwZW9mIGZvcm1hdCA9PT0gJ2Z1bmN0aW9uJykgcmV0dXJuIGZvcm1hdChkYXRlKTtcbiAgcmV0dXJuIGZvcm1hdDtcbn07XG52YXIgX2RlZmF1bHQgPSBmb3JtYXRSZWxhdGl2ZTtcbmV4cG9ydHMuZGVmYXVsdCA9IF9kZWZhdWx0O1xubW9kdWxlLmV4cG9ydHMgPSBleHBvcnRzLmRlZmF1bHQ7Il0sIm5hbWVzIjpbXSwic291cmNlUm9vdCI6IiJ9\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/el/_lib/formatRelative/index.js\n");

/***/ })

}]);