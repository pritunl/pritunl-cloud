"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-pt-_lib-formatRelative-index-js"],{

/***/ "./node_modules/date-fns/locale/pt/_lib/formatRelative/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/pt/_lib/formatRelative/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar formatRelativeLocale = {\n  lastWeek: function lastWeek(date) {\n    var weekday = date.getUTCDay();\n    var last = weekday === 0 || weekday === 6 ? 'último' : 'última';\n    return \"'\" + last + \"' eeee 'às' p\";\n  },\n  yesterday: \"'ontem às' p\",\n  today: \"'hoje às' p\",\n  tomorrow: \"'amanhã às' p\",\n  nextWeek: \"eeee 'às' p\",\n  other: 'P'\n};\nvar formatRelative = function formatRelative(token, date, _baseDate, _options) {\n  var format = formatRelativeLocale[token];\n  if (typeof format === 'function') {\n    return format(date);\n  }\n  return format;\n};\nvar _default = formatRelative;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL3B0L19saWIvZm9ybWF0UmVsYXRpdmUvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0EsR0FBRztBQUNIO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxrQkFBZTtBQUNmIiwic291cmNlcyI6WyJ3ZWJwYWNrOi8vcHJpdHVubC1jbG91ZC8uL25vZGVfbW9kdWxlcy9kYXRlLWZucy9sb2NhbGUvcHQvX2xpYi9mb3JtYXRSZWxhdGl2ZS9pbmRleC5qcz85ODcwIl0sInNvdXJjZXNDb250ZW50IjpbIlwidXNlIHN0cmljdFwiO1xuXG5PYmplY3QuZGVmaW5lUHJvcGVydHkoZXhwb3J0cywgXCJfX2VzTW9kdWxlXCIsIHtcbiAgdmFsdWU6IHRydWVcbn0pO1xuZXhwb3J0cy5kZWZhdWx0ID0gdm9pZCAwO1xudmFyIGZvcm1hdFJlbGF0aXZlTG9jYWxlID0ge1xuICBsYXN0V2VlazogZnVuY3Rpb24gbGFzdFdlZWsoZGF0ZSkge1xuICAgIHZhciB3ZWVrZGF5ID0gZGF0ZS5nZXRVVENEYXkoKTtcbiAgICB2YXIgbGFzdCA9IHdlZWtkYXkgPT09IDAgfHwgd2Vla2RheSA9PT0gNiA/ICfDumx0aW1vJyA6ICfDumx0aW1hJztcbiAgICByZXR1cm4gXCInXCIgKyBsYXN0ICsgXCInIGVlZWUgJ8OgcycgcFwiO1xuICB9LFxuICB5ZXN0ZXJkYXk6IFwiJ29udGVtIMOgcycgcFwiLFxuICB0b2RheTogXCInaG9qZSDDoHMnIHBcIixcbiAgdG9tb3Jyb3c6IFwiJ2FtYW5ow6Mgw6BzJyBwXCIsXG4gIG5leHRXZWVrOiBcImVlZWUgJ8OgcycgcFwiLFxuICBvdGhlcjogJ1AnXG59O1xudmFyIGZvcm1hdFJlbGF0aXZlID0gZnVuY3Rpb24gZm9ybWF0UmVsYXRpdmUodG9rZW4sIGRhdGUsIF9iYXNlRGF0ZSwgX29wdGlvbnMpIHtcbiAgdmFyIGZvcm1hdCA9IGZvcm1hdFJlbGF0aXZlTG9jYWxlW3Rva2VuXTtcbiAgaWYgKHR5cGVvZiBmb3JtYXQgPT09ICdmdW5jdGlvbicpIHtcbiAgICByZXR1cm4gZm9ybWF0KGRhdGUpO1xuICB9XG4gIHJldHVybiBmb3JtYXQ7XG59O1xudmFyIF9kZWZhdWx0ID0gZm9ybWF0UmVsYXRpdmU7XG5leHBvcnRzLmRlZmF1bHQgPSBfZGVmYXVsdDtcbm1vZHVsZS5leHBvcnRzID0gZXhwb3J0cy5kZWZhdWx0OyJdLCJuYW1lcyI6W10sInNvdXJjZVJvb3QiOiIifQ==\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/pt/_lib/formatRelative/index.js\n");

/***/ })

}]);