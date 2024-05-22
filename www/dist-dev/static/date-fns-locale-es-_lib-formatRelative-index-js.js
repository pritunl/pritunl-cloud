"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-es-_lib-formatRelative-index-js"],{

/***/ "./node_modules/date-fns/locale/es/_lib/formatRelative/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/es/_lib/formatRelative/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar formatRelativeLocale = {\n  lastWeek: \"'el' eeee 'pasado a la' p\",\n  yesterday: \"'ayer a la' p\",\n  today: \"'hoy a la' p\",\n  tomorrow: \"'mañana a la' p\",\n  nextWeek: \"eeee 'a la' p\",\n  other: 'P'\n};\nvar formatRelativeLocalePlural = {\n  lastWeek: \"'el' eeee 'pasado a las' p\",\n  yesterday: \"'ayer a las' p\",\n  today: \"'hoy a las' p\",\n  tomorrow: \"'mañana a las' p\",\n  nextWeek: \"eeee 'a las' p\",\n  other: 'P'\n};\nvar formatRelative = function formatRelative(token, date, _baseDate, _options) {\n  if (date.getUTCHours() !== 1) {\n    return formatRelativeLocalePlural[token];\n  } else {\n    return formatRelativeLocale[token];\n  }\n};\nvar _default = formatRelative;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2VzL19saWIvZm9ybWF0UmVsYXRpdmUvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBLElBQUk7QUFDSjtBQUNBO0FBQ0E7QUFDQTtBQUNBLGtCQUFlO0FBQ2YiLCJzb3VyY2VzIjpbIndlYnBhY2s6Ly9wcml0dW5sLWNsb3VkLy4vbm9kZV9tb2R1bGVzL2RhdGUtZm5zL2xvY2FsZS9lcy9fbGliL2Zvcm1hdFJlbGF0aXZlL2luZGV4LmpzPzU5MGMiXSwic291cmNlc0NvbnRlbnQiOlsiXCJ1c2Ugc3RyaWN0XCI7XG5cbk9iamVjdC5kZWZpbmVQcm9wZXJ0eShleHBvcnRzLCBcIl9fZXNNb2R1bGVcIiwge1xuICB2YWx1ZTogdHJ1ZVxufSk7XG5leHBvcnRzLmRlZmF1bHQgPSB2b2lkIDA7XG52YXIgZm9ybWF0UmVsYXRpdmVMb2NhbGUgPSB7XG4gIGxhc3RXZWVrOiBcIidlbCcgZWVlZSAncGFzYWRvIGEgbGEnIHBcIixcbiAgeWVzdGVyZGF5OiBcIidheWVyIGEgbGEnIHBcIixcbiAgdG9kYXk6IFwiJ2hveSBhIGxhJyBwXCIsXG4gIHRvbW9ycm93OiBcIidtYcOxYW5hIGEgbGEnIHBcIixcbiAgbmV4dFdlZWs6IFwiZWVlZSAnYSBsYScgcFwiLFxuICBvdGhlcjogJ1AnXG59O1xudmFyIGZvcm1hdFJlbGF0aXZlTG9jYWxlUGx1cmFsID0ge1xuICBsYXN0V2VlazogXCInZWwnIGVlZWUgJ3Bhc2FkbyBhIGxhcycgcFwiLFxuICB5ZXN0ZXJkYXk6IFwiJ2F5ZXIgYSBsYXMnIHBcIixcbiAgdG9kYXk6IFwiJ2hveSBhIGxhcycgcFwiLFxuICB0b21vcnJvdzogXCInbWHDsWFuYSBhIGxhcycgcFwiLFxuICBuZXh0V2VlazogXCJlZWVlICdhIGxhcycgcFwiLFxuICBvdGhlcjogJ1AnXG59O1xudmFyIGZvcm1hdFJlbGF0aXZlID0gZnVuY3Rpb24gZm9ybWF0UmVsYXRpdmUodG9rZW4sIGRhdGUsIF9iYXNlRGF0ZSwgX29wdGlvbnMpIHtcbiAgaWYgKGRhdGUuZ2V0VVRDSG91cnMoKSAhPT0gMSkge1xuICAgIHJldHVybiBmb3JtYXRSZWxhdGl2ZUxvY2FsZVBsdXJhbFt0b2tlbl07XG4gIH0gZWxzZSB7XG4gICAgcmV0dXJuIGZvcm1hdFJlbGF0aXZlTG9jYWxlW3Rva2VuXTtcbiAgfVxufTtcbnZhciBfZGVmYXVsdCA9IGZvcm1hdFJlbGF0aXZlO1xuZXhwb3J0cy5kZWZhdWx0ID0gX2RlZmF1bHQ7XG5tb2R1bGUuZXhwb3J0cyA9IGV4cG9ydHMuZGVmYXVsdDsiXSwibmFtZXMiOltdLCJzb3VyY2VSb290IjoiIn0=\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/es/_lib/formatRelative/index.js\n");

/***/ })

}]);