"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-sl-_lib-formatRelative-index-js"],{

/***/ "./node_modules/date-fns/locale/sl/_lib/formatRelative/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/sl/_lib/formatRelative/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar formatRelativeLocale = {\n  lastWeek: function lastWeek(date) {\n    var day = date.getUTCDay();\n    switch (day) {\n      case 0:\n        return \"'prejšnjo nedeljo ob' p\";\n      case 3:\n        return \"'prejšnjo sredo ob' p\";\n      case 6:\n        return \"'prejšnjo soboto ob' p\";\n      default:\n        return \"'prejšnji' EEEE 'ob' p\";\n    }\n  },\n  yesterday: \"'včeraj ob' p\",\n  today: \"'danes ob' p\",\n  tomorrow: \"'jutri ob' p\",\n  nextWeek: function nextWeek(date) {\n    var day = date.getUTCDay();\n    switch (day) {\n      case 0:\n        return \"'naslednjo nedeljo ob' p\";\n      case 3:\n        return \"'naslednjo sredo ob' p\";\n      case 6:\n        return \"'naslednjo soboto ob' p\";\n      default:\n        return \"'naslednji' EEEE 'ob' p\";\n    }\n  },\n  other: 'P'\n};\nvar formatRelative = function formatRelative(token, date, _baseDate, _options) {\n  var format = formatRelativeLocale[token];\n  if (typeof format === 'function') {\n    return format(date);\n  }\n  return format;\n};\nvar _default = formatRelative;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL3NsL19saWIvZm9ybWF0UmVsYXRpdmUvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBLEdBQUc7QUFDSDtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxHQUFHO0FBQ0g7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxrQkFBZTtBQUNmIiwic291cmNlcyI6WyJ3ZWJwYWNrOi8vcHJpdHVubC1jbG91ZC8uL25vZGVfbW9kdWxlcy9kYXRlLWZucy9sb2NhbGUvc2wvX2xpYi9mb3JtYXRSZWxhdGl2ZS9pbmRleC5qcz9mMjM4Il0sInNvdXJjZXNDb250ZW50IjpbIlwidXNlIHN0cmljdFwiO1xuXG5PYmplY3QuZGVmaW5lUHJvcGVydHkoZXhwb3J0cywgXCJfX2VzTW9kdWxlXCIsIHtcbiAgdmFsdWU6IHRydWVcbn0pO1xuZXhwb3J0cy5kZWZhdWx0ID0gdm9pZCAwO1xudmFyIGZvcm1hdFJlbGF0aXZlTG9jYWxlID0ge1xuICBsYXN0V2VlazogZnVuY3Rpb24gbGFzdFdlZWsoZGF0ZSkge1xuICAgIHZhciBkYXkgPSBkYXRlLmdldFVUQ0RheSgpO1xuICAgIHN3aXRjaCAoZGF5KSB7XG4gICAgICBjYXNlIDA6XG4gICAgICAgIHJldHVybiBcIidwcmVqxaFuam8gbmVkZWxqbyBvYicgcFwiO1xuICAgICAgY2FzZSAzOlxuICAgICAgICByZXR1cm4gXCIncHJlasWhbmpvIHNyZWRvIG9iJyBwXCI7XG4gICAgICBjYXNlIDY6XG4gICAgICAgIHJldHVybiBcIidwcmVqxaFuam8gc29ib3RvIG9iJyBwXCI7XG4gICAgICBkZWZhdWx0OlxuICAgICAgICByZXR1cm4gXCIncHJlasWhbmppJyBFRUVFICdvYicgcFwiO1xuICAgIH1cbiAgfSxcbiAgeWVzdGVyZGF5OiBcIid2xI1lcmFqIG9iJyBwXCIsXG4gIHRvZGF5OiBcIidkYW5lcyBvYicgcFwiLFxuICB0b21vcnJvdzogXCInanV0cmkgb2InIHBcIixcbiAgbmV4dFdlZWs6IGZ1bmN0aW9uIG5leHRXZWVrKGRhdGUpIHtcbiAgICB2YXIgZGF5ID0gZGF0ZS5nZXRVVENEYXkoKTtcbiAgICBzd2l0Y2ggKGRheSkge1xuICAgICAgY2FzZSAwOlxuICAgICAgICByZXR1cm4gXCInbmFzbGVkbmpvIG5lZGVsam8gb2InIHBcIjtcbiAgICAgIGNhc2UgMzpcbiAgICAgICAgcmV0dXJuIFwiJ25hc2xlZG5qbyBzcmVkbyBvYicgcFwiO1xuICAgICAgY2FzZSA2OlxuICAgICAgICByZXR1cm4gXCInbmFzbGVkbmpvIHNvYm90byBvYicgcFwiO1xuICAgICAgZGVmYXVsdDpcbiAgICAgICAgcmV0dXJuIFwiJ25hc2xlZG5qaScgRUVFRSAnb2InIHBcIjtcbiAgICB9XG4gIH0sXG4gIG90aGVyOiAnUCdcbn07XG52YXIgZm9ybWF0UmVsYXRpdmUgPSBmdW5jdGlvbiBmb3JtYXRSZWxhdGl2ZSh0b2tlbiwgZGF0ZSwgX2Jhc2VEYXRlLCBfb3B0aW9ucykge1xuICB2YXIgZm9ybWF0ID0gZm9ybWF0UmVsYXRpdmVMb2NhbGVbdG9rZW5dO1xuICBpZiAodHlwZW9mIGZvcm1hdCA9PT0gJ2Z1bmN0aW9uJykge1xuICAgIHJldHVybiBmb3JtYXQoZGF0ZSk7XG4gIH1cbiAgcmV0dXJuIGZvcm1hdDtcbn07XG52YXIgX2RlZmF1bHQgPSBmb3JtYXRSZWxhdGl2ZTtcbmV4cG9ydHMuZGVmYXVsdCA9IF9kZWZhdWx0O1xubW9kdWxlLmV4cG9ydHMgPSBleHBvcnRzLmRlZmF1bHQ7Il0sIm5hbWVzIjpbXSwic291cmNlUm9vdCI6IiJ9\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/sl/_lib/formatRelative/index.js\n");

/***/ })

}]);