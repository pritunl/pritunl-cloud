"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-cs-_lib-formatRelative-index-js"],{

/***/ "./node_modules/date-fns/locale/cs/_lib/formatRelative/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/cs/_lib/formatRelative/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar accusativeWeekdays = ['neděli', 'pondělí', 'úterý', 'středu', 'čtvrtek', 'pátek', 'sobotu'];\nvar formatRelativeLocale = {\n  lastWeek: \"'poslední' eeee 've' p\",\n  yesterday: \"'včera v' p\",\n  today: \"'dnes v' p\",\n  tomorrow: \"'zítra v' p\",\n  nextWeek: function nextWeek(date) {\n    var day = date.getUTCDay();\n    return \"'v \" + accusativeWeekdays[day] + \" o' p\";\n  },\n  other: 'P'\n};\nvar formatRelative = function formatRelative(token, date) {\n  var format = formatRelativeLocale[token];\n  if (typeof format === 'function') {\n    return format(date);\n  }\n  return format;\n};\nvar _default = formatRelative;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2NzL19saWIvZm9ybWF0UmVsYXRpdmUvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxHQUFHO0FBQ0g7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxrQkFBZTtBQUNmIiwic291cmNlcyI6WyJ3ZWJwYWNrOi8vcHJpdHVubC1jbG91ZC8uL25vZGVfbW9kdWxlcy9kYXRlLWZucy9sb2NhbGUvY3MvX2xpYi9mb3JtYXRSZWxhdGl2ZS9pbmRleC5qcz85MmM4Il0sInNvdXJjZXNDb250ZW50IjpbIlwidXNlIHN0cmljdFwiO1xuXG5PYmplY3QuZGVmaW5lUHJvcGVydHkoZXhwb3J0cywgXCJfX2VzTW9kdWxlXCIsIHtcbiAgdmFsdWU6IHRydWVcbn0pO1xuZXhwb3J0cy5kZWZhdWx0ID0gdm9pZCAwO1xudmFyIGFjY3VzYXRpdmVXZWVrZGF5cyA9IFsnbmVkxJtsaScsICdwb25kxJtsw60nLCAnw7p0ZXLDvScsICdzdMWZZWR1JywgJ8SNdHZydGVrJywgJ3DDoXRlaycsICdzb2JvdHUnXTtcbnZhciBmb3JtYXRSZWxhdGl2ZUxvY2FsZSA9IHtcbiAgbGFzdFdlZWs6IFwiJ3Bvc2xlZG7DrScgZWVlZSAndmUnIHBcIixcbiAgeWVzdGVyZGF5OiBcIid2xI1lcmEgdicgcFwiLFxuICB0b2RheTogXCInZG5lcyB2JyBwXCIsXG4gIHRvbW9ycm93OiBcIid6w610cmEgdicgcFwiLFxuICBuZXh0V2VlazogZnVuY3Rpb24gbmV4dFdlZWsoZGF0ZSkge1xuICAgIHZhciBkYXkgPSBkYXRlLmdldFVUQ0RheSgpO1xuICAgIHJldHVybiBcIid2IFwiICsgYWNjdXNhdGl2ZVdlZWtkYXlzW2RheV0gKyBcIiBvJyBwXCI7XG4gIH0sXG4gIG90aGVyOiAnUCdcbn07XG52YXIgZm9ybWF0UmVsYXRpdmUgPSBmdW5jdGlvbiBmb3JtYXRSZWxhdGl2ZSh0b2tlbiwgZGF0ZSkge1xuICB2YXIgZm9ybWF0ID0gZm9ybWF0UmVsYXRpdmVMb2NhbGVbdG9rZW5dO1xuICBpZiAodHlwZW9mIGZvcm1hdCA9PT0gJ2Z1bmN0aW9uJykge1xuICAgIHJldHVybiBmb3JtYXQoZGF0ZSk7XG4gIH1cbiAgcmV0dXJuIGZvcm1hdDtcbn07XG52YXIgX2RlZmF1bHQgPSBmb3JtYXRSZWxhdGl2ZTtcbmV4cG9ydHMuZGVmYXVsdCA9IF9kZWZhdWx0O1xubW9kdWxlLmV4cG9ydHMgPSBleHBvcnRzLmRlZmF1bHQ7Il0sIm5hbWVzIjpbXSwic291cmNlUm9vdCI6IiJ9\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/cs/_lib/formatRelative/index.js\n");

/***/ })

}]);