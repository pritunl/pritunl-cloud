"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-be-_lib-formatRelative-index-js"],{

/***/ "./node_modules/date-fns/locale/be/_lib/formatRelative/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/be/_lib/formatRelative/index.js ***!
  \**********************************************************************/
/***/ ((module, exports, __webpack_require__) => {

eval("\n\nvar _interopRequireDefault = (__webpack_require__(/*! @babel/runtime/helpers/interopRequireDefault */ \"./node_modules/@babel/runtime/helpers/interopRequireDefault.js\")[\"default\"]);\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\nvar _index = __webpack_require__(/*! ../../../../index.js */ \"./node_modules/date-fns/index.js\");\nvar _index2 = _interopRequireDefault(__webpack_require__(/*! ../../../../_lib/isSameUTCWeek/index.js */ \"./node_modules/date-fns/_lib/isSameUTCWeek/index.js\"));\nvar accusativeWeekdays = ['нядзелю', 'панядзелак', 'аўторак', 'сераду', 'чацвер', 'пятніцу', 'суботу'];\nfunction lastWeek(day) {\n  var weekday = accusativeWeekdays[day];\n  switch (day) {\n    case 0:\n    case 3:\n    case 5:\n    case 6:\n      return \"'у мінулую \" + weekday + \" а' p\";\n    case 1:\n    case 2:\n    case 4:\n      return \"'у мінулы \" + weekday + \" а' p\";\n  }\n}\nfunction thisWeek(day) {\n  var weekday = accusativeWeekdays[day];\n  return \"'у \" + weekday + \" а' p\";\n}\nfunction nextWeek(day) {\n  var weekday = accusativeWeekdays[day];\n  switch (day) {\n    case 0:\n    case 3:\n    case 5:\n    case 6:\n      return \"'у наступную \" + weekday + \" а' p\";\n    case 1:\n    case 2:\n    case 4:\n      return \"'у наступны \" + weekday + \" а' p\";\n  }\n}\nvar lastWeekFormat = function lastWeekFormat(dirtyDate, baseDate, options) {\n  var date = (0, _index.toDate)(dirtyDate);\n  var day = date.getUTCDay();\n  if ((0, _index2.default)(date, baseDate, options)) {\n    return thisWeek(day);\n  } else {\n    return lastWeek(day);\n  }\n};\nvar nextWeekFormat = function nextWeekFormat(dirtyDate, baseDate, options) {\n  var date = (0, _index.toDate)(dirtyDate);\n  var day = date.getUTCDay();\n  if ((0, _index2.default)(date, baseDate, options)) {\n    return thisWeek(day);\n  } else {\n    return nextWeek(day);\n  }\n};\nvar formatRelativeLocale = {\n  lastWeek: lastWeekFormat,\n  yesterday: \"'учора а' p\",\n  today: \"'сёння а' p\",\n  tomorrow: \"'заўтра а' p\",\n  nextWeek: nextWeekFormat,\n  other: 'P'\n};\nvar formatRelative = function formatRelative(token, date, baseDate, options) {\n  var format = formatRelativeLocale[token];\n  if (typeof format === 'function') {\n    return format(date, baseDate, options);\n  }\n  return format;\n};\nvar _default = formatRelative;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2JlL19saWIvZm9ybWF0UmVsYXRpdmUvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsNkJBQTZCLHNKQUErRDtBQUM1Riw4Q0FBNkM7QUFDN0M7QUFDQSxDQUFDLEVBQUM7QUFDRixrQkFBZTtBQUNmLGFBQWEsbUJBQU8sQ0FBQyw4REFBc0I7QUFDM0MscUNBQXFDLG1CQUFPLENBQUMsb0dBQXlDO0FBQ3RGO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxJQUFJO0FBQ0o7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBLElBQUk7QUFDSjtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBLGtCQUFlO0FBQ2YiLCJzb3VyY2VzIjpbIndlYnBhY2s6Ly9wcml0dW5sLWNsb3VkLy4vbm9kZV9tb2R1bGVzL2RhdGUtZm5zL2xvY2FsZS9iZS9fbGliL2Zvcm1hdFJlbGF0aXZlL2luZGV4LmpzPzBiMmMiXSwic291cmNlc0NvbnRlbnQiOlsiXCJ1c2Ugc3RyaWN0XCI7XG5cbnZhciBfaW50ZXJvcFJlcXVpcmVEZWZhdWx0ID0gcmVxdWlyZShcIkBiYWJlbC9ydW50aW1lL2hlbHBlcnMvaW50ZXJvcFJlcXVpcmVEZWZhdWx0XCIpLmRlZmF1bHQ7XG5PYmplY3QuZGVmaW5lUHJvcGVydHkoZXhwb3J0cywgXCJfX2VzTW9kdWxlXCIsIHtcbiAgdmFsdWU6IHRydWVcbn0pO1xuZXhwb3J0cy5kZWZhdWx0ID0gdm9pZCAwO1xudmFyIF9pbmRleCA9IHJlcXVpcmUoXCIuLi8uLi8uLi8uLi9pbmRleC5qc1wiKTtcbnZhciBfaW5kZXgyID0gX2ludGVyb3BSZXF1aXJlRGVmYXVsdChyZXF1aXJlKFwiLi4vLi4vLi4vLi4vX2xpYi9pc1NhbWVVVENXZWVrL2luZGV4LmpzXCIpKTtcbnZhciBhY2N1c2F0aXZlV2Vla2RheXMgPSBbJ9C90Y/QtNC30LXQu9GOJywgJ9C/0LDQvdGP0LTQt9C10LvQsNC6JywgJ9Cw0Z7RgtC+0YDQsNC6JywgJ9GB0LXRgNCw0LTRgycsICfRh9Cw0YbQstC10YAnLCAn0L/Rj9GC0L3RltGG0YMnLCAn0YHRg9Cx0L7RgtGDJ107XG5mdW5jdGlvbiBsYXN0V2VlayhkYXkpIHtcbiAgdmFyIHdlZWtkYXkgPSBhY2N1c2F0aXZlV2Vla2RheXNbZGF5XTtcbiAgc3dpdGNoIChkYXkpIHtcbiAgICBjYXNlIDA6XG4gICAgY2FzZSAzOlxuICAgIGNhc2UgNTpcbiAgICBjYXNlIDY6XG4gICAgICByZXR1cm4gXCIn0YMg0LzRltC90YPQu9GD0Y4gXCIgKyB3ZWVrZGF5ICsgXCIg0LAnIHBcIjtcbiAgICBjYXNlIDE6XG4gICAgY2FzZSAyOlxuICAgIGNhc2UgNDpcbiAgICAgIHJldHVybiBcIifRgyDQvNGW0L3Rg9C70YsgXCIgKyB3ZWVrZGF5ICsgXCIg0LAnIHBcIjtcbiAgfVxufVxuZnVuY3Rpb24gdGhpc1dlZWsoZGF5KSB7XG4gIHZhciB3ZWVrZGF5ID0gYWNjdXNhdGl2ZVdlZWtkYXlzW2RheV07XG4gIHJldHVybiBcIifRgyBcIiArIHdlZWtkYXkgKyBcIiDQsCcgcFwiO1xufVxuZnVuY3Rpb24gbmV4dFdlZWsoZGF5KSB7XG4gIHZhciB3ZWVrZGF5ID0gYWNjdXNhdGl2ZVdlZWtkYXlzW2RheV07XG4gIHN3aXRjaCAoZGF5KSB7XG4gICAgY2FzZSAwOlxuICAgIGNhc2UgMzpcbiAgICBjYXNlIDU6XG4gICAgY2FzZSA2OlxuICAgICAgcmV0dXJuIFwiJ9GDINC90LDRgdGC0YPQv9C90YPRjiBcIiArIHdlZWtkYXkgKyBcIiDQsCcgcFwiO1xuICAgIGNhc2UgMTpcbiAgICBjYXNlIDI6XG4gICAgY2FzZSA0OlxuICAgICAgcmV0dXJuIFwiJ9GDINC90LDRgdGC0YPQv9C90YsgXCIgKyB3ZWVrZGF5ICsgXCIg0LAnIHBcIjtcbiAgfVxufVxudmFyIGxhc3RXZWVrRm9ybWF0ID0gZnVuY3Rpb24gbGFzdFdlZWtGb3JtYXQoZGlydHlEYXRlLCBiYXNlRGF0ZSwgb3B0aW9ucykge1xuICB2YXIgZGF0ZSA9ICgwLCBfaW5kZXgudG9EYXRlKShkaXJ0eURhdGUpO1xuICB2YXIgZGF5ID0gZGF0ZS5nZXRVVENEYXkoKTtcbiAgaWYgKCgwLCBfaW5kZXgyLmRlZmF1bHQpKGRhdGUsIGJhc2VEYXRlLCBvcHRpb25zKSkge1xuICAgIHJldHVybiB0aGlzV2VlayhkYXkpO1xuICB9IGVsc2Uge1xuICAgIHJldHVybiBsYXN0V2VlayhkYXkpO1xuICB9XG59O1xudmFyIG5leHRXZWVrRm9ybWF0ID0gZnVuY3Rpb24gbmV4dFdlZWtGb3JtYXQoZGlydHlEYXRlLCBiYXNlRGF0ZSwgb3B0aW9ucykge1xuICB2YXIgZGF0ZSA9ICgwLCBfaW5kZXgudG9EYXRlKShkaXJ0eURhdGUpO1xuICB2YXIgZGF5ID0gZGF0ZS5nZXRVVENEYXkoKTtcbiAgaWYgKCgwLCBfaW5kZXgyLmRlZmF1bHQpKGRhdGUsIGJhc2VEYXRlLCBvcHRpb25zKSkge1xuICAgIHJldHVybiB0aGlzV2VlayhkYXkpO1xuICB9IGVsc2Uge1xuICAgIHJldHVybiBuZXh0V2VlayhkYXkpO1xuICB9XG59O1xudmFyIGZvcm1hdFJlbGF0aXZlTG9jYWxlID0ge1xuICBsYXN0V2VlazogbGFzdFdlZWtGb3JtYXQsXG4gIHllc3RlcmRheTogXCIn0YPRh9C+0YDQsCDQsCcgcFwiLFxuICB0b2RheTogXCIn0YHRkdC90L3RjyDQsCcgcFwiLFxuICB0b21vcnJvdzogXCIn0LfQsNGe0YLRgNCwINCwJyBwXCIsXG4gIG5leHRXZWVrOiBuZXh0V2Vla0Zvcm1hdCxcbiAgb3RoZXI6ICdQJ1xufTtcbnZhciBmb3JtYXRSZWxhdGl2ZSA9IGZ1bmN0aW9uIGZvcm1hdFJlbGF0aXZlKHRva2VuLCBkYXRlLCBiYXNlRGF0ZSwgb3B0aW9ucykge1xuICB2YXIgZm9ybWF0ID0gZm9ybWF0UmVsYXRpdmVMb2NhbGVbdG9rZW5dO1xuICBpZiAodHlwZW9mIGZvcm1hdCA9PT0gJ2Z1bmN0aW9uJykge1xuICAgIHJldHVybiBmb3JtYXQoZGF0ZSwgYmFzZURhdGUsIG9wdGlvbnMpO1xuICB9XG4gIHJldHVybiBmb3JtYXQ7XG59O1xudmFyIF9kZWZhdWx0ID0gZm9ybWF0UmVsYXRpdmU7XG5leHBvcnRzLmRlZmF1bHQgPSBfZGVmYXVsdDtcbm1vZHVsZS5leHBvcnRzID0gZXhwb3J0cy5kZWZhdWx0OyJdLCJuYW1lcyI6W10sInNvdXJjZVJvb3QiOiIifQ==\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/be/_lib/formatRelative/index.js\n");

/***/ })

}]);