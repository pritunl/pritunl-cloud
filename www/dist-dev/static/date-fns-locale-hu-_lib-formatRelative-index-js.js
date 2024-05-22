"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-hu-_lib-formatRelative-index-js"],{

/***/ "./node_modules/date-fns/locale/hu/_lib/formatRelative/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/hu/_lib/formatRelative/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar accusativeWeekdays = ['vasárnap', 'hétfőn', 'kedden', 'szerdán', 'csütörtökön', 'pénteken', 'szombaton'];\nfunction week(isFuture) {\n  return function (date) {\n    var weekday = accusativeWeekdays[date.getUTCDay()];\n    var prefix = isFuture ? '' : \"'múlt' \";\n    return \"\".concat(prefix, \"'\").concat(weekday, \"' p'-kor'\");\n  };\n}\nvar formatRelativeLocale = {\n  lastWeek: week(false),\n  yesterday: \"'tegnap' p'-kor'\",\n  today: \"'ma' p'-kor'\",\n  tomorrow: \"'holnap' p'-kor'\",\n  nextWeek: week(true),\n  other: 'P'\n};\nvar formatRelative = function formatRelative(token, date) {\n  var format = formatRelativeLocale[token];\n  if (typeof format === 'function') {\n    return format(date);\n  }\n  return format;\n};\nvar _default = formatRelative;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2h1L19saWIvZm9ybWF0UmVsYXRpdmUvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxrQkFBZTtBQUNmIiwic291cmNlcyI6WyJ3ZWJwYWNrOi8vcHJpdHVubC1jbG91ZC8uL25vZGVfbW9kdWxlcy9kYXRlLWZucy9sb2NhbGUvaHUvX2xpYi9mb3JtYXRSZWxhdGl2ZS9pbmRleC5qcz9lMTk0Il0sInNvdXJjZXNDb250ZW50IjpbIlwidXNlIHN0cmljdFwiO1xuXG5PYmplY3QuZGVmaW5lUHJvcGVydHkoZXhwb3J0cywgXCJfX2VzTW9kdWxlXCIsIHtcbiAgdmFsdWU6IHRydWVcbn0pO1xuZXhwb3J0cy5kZWZhdWx0ID0gdm9pZCAwO1xudmFyIGFjY3VzYXRpdmVXZWVrZGF5cyA9IFsndmFzw6FybmFwJywgJ2jDqXRmxZFuJywgJ2tlZGRlbicsICdzemVyZMOhbicsICdjc8O8dMO2cnTDtmvDtm4nLCAncMOpbnRla2VuJywgJ3N6b21iYXRvbiddO1xuZnVuY3Rpb24gd2Vlayhpc0Z1dHVyZSkge1xuICByZXR1cm4gZnVuY3Rpb24gKGRhdGUpIHtcbiAgICB2YXIgd2Vla2RheSA9IGFjY3VzYXRpdmVXZWVrZGF5c1tkYXRlLmdldFVUQ0RheSgpXTtcbiAgICB2YXIgcHJlZml4ID0gaXNGdXR1cmUgPyAnJyA6IFwiJ23Dumx0JyBcIjtcbiAgICByZXR1cm4gXCJcIi5jb25jYXQocHJlZml4LCBcIidcIikuY29uY2F0KHdlZWtkYXksIFwiJyBwJy1rb3InXCIpO1xuICB9O1xufVxudmFyIGZvcm1hdFJlbGF0aXZlTG9jYWxlID0ge1xuICBsYXN0V2Vlazogd2VlayhmYWxzZSksXG4gIHllc3RlcmRheTogXCIndGVnbmFwJyBwJy1rb3InXCIsXG4gIHRvZGF5OiBcIidtYScgcCcta29yJ1wiLFxuICB0b21vcnJvdzogXCInaG9sbmFwJyBwJy1rb3InXCIsXG4gIG5leHRXZWVrOiB3ZWVrKHRydWUpLFxuICBvdGhlcjogJ1AnXG59O1xudmFyIGZvcm1hdFJlbGF0aXZlID0gZnVuY3Rpb24gZm9ybWF0UmVsYXRpdmUodG9rZW4sIGRhdGUpIHtcbiAgdmFyIGZvcm1hdCA9IGZvcm1hdFJlbGF0aXZlTG9jYWxlW3Rva2VuXTtcbiAgaWYgKHR5cGVvZiBmb3JtYXQgPT09ICdmdW5jdGlvbicpIHtcbiAgICByZXR1cm4gZm9ybWF0KGRhdGUpO1xuICB9XG4gIHJldHVybiBmb3JtYXQ7XG59O1xudmFyIF9kZWZhdWx0ID0gZm9ybWF0UmVsYXRpdmU7XG5leHBvcnRzLmRlZmF1bHQgPSBfZGVmYXVsdDtcbm1vZHVsZS5leHBvcnRzID0gZXhwb3J0cy5kZWZhdWx0OyJdLCJuYW1lcyI6W10sInNvdXJjZVJvb3QiOiIifQ==\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/hu/_lib/formatRelative/index.js\n");

/***/ })

}]);