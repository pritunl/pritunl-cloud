"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-lb-_lib-formatRelative-index-js"],{

/***/ "./node_modules/date-fns/locale/lb/_lib/formatRelative/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/lb/_lib/formatRelative/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar formatRelativeLocale = {\n  lastWeek: function lastWeek(date) {\n    var day = date.getUTCDay();\n    var result = \"'läschte\";\n    if (day === 2 || day === 4) {\n      // Eifeler Regel: Add an n before the consonant d; Here \"Dënschdeg\" \"and Donneschde\".\n      result += 'n';\n    }\n    result += \"' eeee 'um' p\";\n    return result;\n  },\n  yesterday: \"'gëschter um' p\",\n  today: \"'haut um' p\",\n  tomorrow: \"'moien um' p\",\n  nextWeek: \"eeee 'um' p\",\n  other: 'P'\n};\nvar formatRelative = function formatRelative(token, date, _baseDate, _options) {\n  var format = formatRelativeLocale[token];\n  if (typeof format === 'function') {\n    return format(date);\n  }\n  return format;\n};\nvar _default = formatRelative;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2xiL19saWIvZm9ybWF0UmVsYXRpdmUvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0EseURBQXlEO0FBQ3pEO0FBQ0E7QUFDQTtBQUNBO0FBQ0EsR0FBRztBQUNIO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxrQkFBZTtBQUNmIiwic291cmNlcyI6WyJ3ZWJwYWNrOi8vcHJpdHVubC1jbG91ZC8uL25vZGVfbW9kdWxlcy9kYXRlLWZucy9sb2NhbGUvbGIvX2xpYi9mb3JtYXRSZWxhdGl2ZS9pbmRleC5qcz9lMmM0Il0sInNvdXJjZXNDb250ZW50IjpbIlwidXNlIHN0cmljdFwiO1xuXG5PYmplY3QuZGVmaW5lUHJvcGVydHkoZXhwb3J0cywgXCJfX2VzTW9kdWxlXCIsIHtcbiAgdmFsdWU6IHRydWVcbn0pO1xuZXhwb3J0cy5kZWZhdWx0ID0gdm9pZCAwO1xudmFyIGZvcm1hdFJlbGF0aXZlTG9jYWxlID0ge1xuICBsYXN0V2VlazogZnVuY3Rpb24gbGFzdFdlZWsoZGF0ZSkge1xuICAgIHZhciBkYXkgPSBkYXRlLmdldFVUQ0RheSgpO1xuICAgIHZhciByZXN1bHQgPSBcIidsw6RzY2h0ZVwiO1xuICAgIGlmIChkYXkgPT09IDIgfHwgZGF5ID09PSA0KSB7XG4gICAgICAvLyBFaWZlbGVyIFJlZ2VsOiBBZGQgYW4gbiBiZWZvcmUgdGhlIGNvbnNvbmFudCBkOyBIZXJlIFwiRMOrbnNjaGRlZ1wiIFwiYW5kIERvbm5lc2NoZGVcIi5cbiAgICAgIHJlc3VsdCArPSAnbic7XG4gICAgfVxuICAgIHJlc3VsdCArPSBcIicgZWVlZSAndW0nIHBcIjtcbiAgICByZXR1cm4gcmVzdWx0O1xuICB9LFxuICB5ZXN0ZXJkYXk6IFwiJ2fDq3NjaHRlciB1bScgcFwiLFxuICB0b2RheTogXCInaGF1dCB1bScgcFwiLFxuICB0b21vcnJvdzogXCInbW9pZW4gdW0nIHBcIixcbiAgbmV4dFdlZWs6IFwiZWVlZSAndW0nIHBcIixcbiAgb3RoZXI6ICdQJ1xufTtcbnZhciBmb3JtYXRSZWxhdGl2ZSA9IGZ1bmN0aW9uIGZvcm1hdFJlbGF0aXZlKHRva2VuLCBkYXRlLCBfYmFzZURhdGUsIF9vcHRpb25zKSB7XG4gIHZhciBmb3JtYXQgPSBmb3JtYXRSZWxhdGl2ZUxvY2FsZVt0b2tlbl07XG4gIGlmICh0eXBlb2YgZm9ybWF0ID09PSAnZnVuY3Rpb24nKSB7XG4gICAgcmV0dXJuIGZvcm1hdChkYXRlKTtcbiAgfVxuICByZXR1cm4gZm9ybWF0O1xufTtcbnZhciBfZGVmYXVsdCA9IGZvcm1hdFJlbGF0aXZlO1xuZXhwb3J0cy5kZWZhdWx0ID0gX2RlZmF1bHQ7XG5tb2R1bGUuZXhwb3J0cyA9IGV4cG9ydHMuZGVmYXVsdDsiXSwibmFtZXMiOltdLCJzb3VyY2VSb290IjoiIn0=\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/lb/_lib/formatRelative/index.js\n");

/***/ })

}]);