"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["date-fns-locale-kn-_lib-formatDistance-index-js"],{

/***/ "./node_modules/date-fns/locale/kn/_lib/formatDistance/index.js":
/*!**********************************************************************!*\
  !*** ./node_modules/date-fns/locale/kn/_lib/formatDistance/index.js ***!
  \**********************************************************************/
/***/ ((module, exports) => {

eval("\n\nObject.defineProperty(exports, \"__esModule\", ({\n  value: true\n}));\nexports[\"default\"] = void 0;\n// note: no implementation for weeks\n\nvar formatDistanceLocale = {\n  lessThanXSeconds: {\n    one: {\n      default: '1 ಸೆಕೆಂಡ್‌ಗಿಂತ ಕಡಿಮೆ',\n      future: '1 ಸೆಕೆಂಡ್‌ಗಿಂತ ಕಡಿಮೆ',\n      past: '1 ಸೆಕೆಂಡ್‌ಗಿಂತ ಕಡಿಮೆ'\n    },\n    other: {\n      default: '{{count}} ಸೆಕೆಂಡ್‌ಗಿಂತ ಕಡಿಮೆ',\n      future: '{{count}} ಸೆಕೆಂಡ್‌ಗಿಂತ ಕಡಿಮೆ',\n      past: '{{count}} ಸೆಕೆಂಡ್‌ಗಿಂತ ಕಡಿಮೆ'\n    }\n  },\n  xSeconds: {\n    one: {\n      default: '1 ಸೆಕೆಂಡ್',\n      future: '1 ಸೆಕೆಂಡ್‌ನಲ್ಲಿ',\n      past: '1 ಸೆಕೆಂಡ್ ಹಿಂದೆ'\n    },\n    other: {\n      default: '{{count}} ಸೆಕೆಂಡುಗಳು',\n      future: '{{count}} ಸೆಕೆಂಡ್‌ಗಳಲ್ಲಿ',\n      past: '{{count}} ಸೆಕೆಂಡ್ ಹಿಂದೆ'\n    }\n  },\n  halfAMinute: {\n    other: {\n      default: 'ಅರ್ಧ ನಿಮಿಷ',\n      future: 'ಅರ್ಧ ನಿಮಿಷದಲ್ಲಿ',\n      past: 'ಅರ್ಧ ನಿಮಿಷದ ಹಿಂದೆ'\n    }\n  },\n  lessThanXMinutes: {\n    one: {\n      default: '1 ನಿಮಿಷಕ್ಕಿಂತ ಕಡಿಮೆ',\n      future: '1 ನಿಮಿಷಕ್ಕಿಂತ ಕಡಿಮೆ',\n      past: '1 ನಿಮಿಷಕ್ಕಿಂತ ಕಡಿಮೆ'\n    },\n    other: {\n      default: '{{count}} ನಿಮಿಷಕ್ಕಿಂತ ಕಡಿಮೆ',\n      future: '{{count}} ನಿಮಿಷಕ್ಕಿಂತ ಕಡಿಮೆ',\n      past: '{{count}} ನಿಮಿಷಕ್ಕಿಂತ ಕಡಿಮೆ'\n    }\n  },\n  xMinutes: {\n    one: {\n      default: '1 ನಿಮಿಷ',\n      future: '1 ನಿಮಿಷದಲ್ಲಿ',\n      past: '1 ನಿಮಿಷದ ಹಿಂದೆ'\n    },\n    other: {\n      default: '{{count}} ನಿಮಿಷಗಳು',\n      future: '{{count}} ನಿಮಿಷಗಳಲ್ಲಿ',\n      past: '{{count}} ನಿಮಿಷಗಳ ಹಿಂದೆ'\n    }\n  },\n  aboutXHours: {\n    one: {\n      default: 'ಸುಮಾರು 1 ಗಂಟೆ',\n      future: 'ಸುಮಾರು 1 ಗಂಟೆಯಲ್ಲಿ',\n      past: 'ಸುಮಾರು 1 ಗಂಟೆ ಹಿಂದೆ'\n    },\n    other: {\n      default: 'ಸುಮಾರು {{count}} ಗಂಟೆಗಳು',\n      future: 'ಸುಮಾರು {{count}} ಗಂಟೆಗಳಲ್ಲಿ',\n      past: 'ಸುಮಾರು {{count}} ಗಂಟೆಗಳ ಹಿಂದೆ'\n    }\n  },\n  xHours: {\n    one: {\n      default: '1 ಗಂಟೆ',\n      future: '1 ಗಂಟೆಯಲ್ಲಿ',\n      past: '1 ಗಂಟೆ ಹಿಂದೆ'\n    },\n    other: {\n      default: '{{count}} ಗಂಟೆಗಳು',\n      future: '{{count}} ಗಂಟೆಗಳಲ್ಲಿ',\n      past: '{{count}} ಗಂಟೆಗಳ ಹಿಂದೆ'\n    }\n  },\n  xDays: {\n    one: {\n      default: '1 ದಿನ',\n      future: '1 ದಿನದಲ್ಲಿ',\n      past: '1 ದಿನದ ಹಿಂದೆ'\n    },\n    other: {\n      default: '{{count}} ದಿನಗಳು',\n      future: '{{count}} ದಿನಗಳಲ್ಲಿ',\n      past: '{{count}} ದಿನಗಳ ಹಿಂದೆ'\n    }\n  },\n  // TODO\n  // aboutXWeeks: {},\n\n  // TODO\n  // xWeeks: {},\n\n  aboutXMonths: {\n    one: {\n      default: 'ಸುಮಾರು 1 ತಿಂಗಳು',\n      future: 'ಸುಮಾರು 1 ತಿಂಗಳಲ್ಲಿ',\n      past: 'ಸುಮಾರು 1 ತಿಂಗಳ ಹಿಂದೆ'\n    },\n    other: {\n      default: 'ಸುಮಾರು {{count}} ತಿಂಗಳು',\n      future: 'ಸುಮಾರು {{count}} ತಿಂಗಳುಗಳಲ್ಲಿ',\n      past: 'ಸುಮಾರು {{count}} ತಿಂಗಳುಗಳ ಹಿಂದೆ'\n    }\n  },\n  xMonths: {\n    one: {\n      default: '1 ತಿಂಗಳು',\n      future: '1 ತಿಂಗಳಲ್ಲಿ',\n      past: '1 ತಿಂಗಳ ಹಿಂದೆ'\n    },\n    other: {\n      default: '{{count}} ತಿಂಗಳು',\n      future: '{{count}} ತಿಂಗಳುಗಳಲ್ಲಿ',\n      past: '{{count}} ತಿಂಗಳುಗಳ ಹಿಂದೆ'\n    }\n  },\n  aboutXYears: {\n    one: {\n      default: 'ಸುಮಾರು 1 ವರ್ಷ',\n      future: 'ಸುಮಾರು 1 ವರ್ಷದಲ್ಲಿ',\n      past: 'ಸುಮಾರು 1 ವರ್ಷದ ಹಿಂದೆ'\n    },\n    other: {\n      default: 'ಸುಮಾರು {{count}} ವರ್ಷಗಳು',\n      future: 'ಸುಮಾರು {{count}} ವರ್ಷಗಳಲ್ಲಿ',\n      past: 'ಸುಮಾರು {{count}} ವರ್ಷಗಳ ಹಿಂದೆ'\n    }\n  },\n  xYears: {\n    one: {\n      default: '1 ವರ್ಷ',\n      future: '1 ವರ್ಷದಲ್ಲಿ',\n      past: '1 ವರ್ಷದ ಹಿಂದೆ'\n    },\n    other: {\n      default: '{{count}} ವರ್ಷಗಳು',\n      future: '{{count}} ವರ್ಷಗಳಲ್ಲಿ',\n      past: '{{count}} ವರ್ಷಗಳ ಹಿಂದೆ'\n    }\n  },\n  overXYears: {\n    one: {\n      default: '1 ವರ್ಷದ ಮೇಲೆ',\n      future: '1 ವರ್ಷದ ಮೇಲೆ',\n      past: '1 ವರ್ಷದ ಮೇಲೆ'\n    },\n    other: {\n      default: '{{count}} ವರ್ಷಗಳ ಮೇಲೆ',\n      future: '{{count}} ವರ್ಷಗಳ ಮೇಲೆ',\n      past: '{{count}} ವರ್ಷಗಳ ಮೇಲೆ'\n    }\n  },\n  almostXYears: {\n    one: {\n      default: 'ಬಹುತೇಕ 1 ವರ್ಷದಲ್ಲಿ',\n      future: 'ಬಹುತೇಕ 1 ವರ್ಷದಲ್ಲಿ',\n      past: 'ಬಹುತೇಕ 1 ವರ್ಷದಲ್ಲಿ'\n    },\n    other: {\n      default: 'ಬಹುತೇಕ {{count}} ವರ್ಷಗಳಲ್ಲಿ',\n      future: 'ಬಹುತೇಕ {{count}} ವರ್ಷಗಳಲ್ಲಿ',\n      past: 'ಬಹುತೇಕ {{count}} ವರ್ಷಗಳಲ್ಲಿ'\n    }\n  }\n};\nfunction getResultByTense(parentToken, options) {\n  if (options !== null && options !== void 0 && options.addSuffix) {\n    if (options.comparison && options.comparison > 0) {\n      return parentToken.future;\n    } else {\n      return parentToken.past;\n    }\n  }\n  return parentToken.default;\n}\nvar formatDistance = function formatDistance(token, count, options) {\n  var result;\n  var tokenValue = formatDistanceLocale[token];\n  if (tokenValue.one && count === 1) {\n    result = getResultByTense(tokenValue.one, options);\n  } else {\n    result = getResultByTense(tokenValue.other, options);\n  }\n  return result.replace('{{count}}', String(count));\n};\nvar _default = formatDistance;\nexports[\"default\"] = _default;\nmodule.exports = exports.default;//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvZGF0ZS1mbnMvbG9jYWxlL2tuL19saWIvZm9ybWF0RGlzdGFuY2UvaW5kZXguanMiLCJtYXBwaW5ncyI6IkFBQWE7O0FBRWIsOENBQTZDO0FBQzdDO0FBQ0EsQ0FBQyxFQUFDO0FBQ0Ysa0JBQWU7QUFDZjs7QUFFQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxLQUFLO0FBQ0w7QUFDQSxrQkFBa0IsUUFBUTtBQUMxQixpQkFBaUIsUUFBUTtBQUN6QixlQUFlLFFBQVE7QUFDdkI7QUFDQSxHQUFHO0FBQ0g7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBLEtBQUs7QUFDTDtBQUNBLGtCQUFrQixRQUFRO0FBQzFCLGlCQUFpQixRQUFRO0FBQ3pCLGVBQWUsUUFBUTtBQUN2QjtBQUNBLEdBQUc7QUFDSDtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxHQUFHO0FBQ0g7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBLEtBQUs7QUFDTDtBQUNBLGtCQUFrQixRQUFRO0FBQzFCLGlCQUFpQixRQUFRO0FBQ3pCLGVBQWUsUUFBUTtBQUN2QjtBQUNBLEdBQUc7QUFDSDtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0EsS0FBSztBQUNMO0FBQ0Esa0JBQWtCLFFBQVE7QUFDMUIsaUJBQWlCLFFBQVE7QUFDekIsZUFBZSxRQUFRO0FBQ3ZCO0FBQ0EsR0FBRztBQUNIO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxLQUFLO0FBQ0w7QUFDQSx5QkFBeUIsUUFBUTtBQUNqQyx3QkFBd0IsUUFBUTtBQUNoQyxzQkFBc0IsUUFBUTtBQUM5QjtBQUNBLEdBQUc7QUFDSDtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0EsS0FBSztBQUNMO0FBQ0Esa0JBQWtCLFFBQVE7QUFDMUIsaUJBQWlCLFFBQVE7QUFDekIsZUFBZSxRQUFRO0FBQ3ZCO0FBQ0EsR0FBRztBQUNIO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxLQUFLO0FBQ0w7QUFDQSxrQkFBa0IsUUFBUTtBQUMxQixpQkFBaUIsUUFBUTtBQUN6QixlQUFlLFFBQVE7QUFDdkI7QUFDQSxHQUFHO0FBQ0g7QUFDQSxvQkFBb0I7O0FBRXBCO0FBQ0EsZUFBZTs7QUFFZjtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0EsS0FBSztBQUNMO0FBQ0EseUJBQXlCLFFBQVE7QUFDakMsd0JBQXdCLFFBQVE7QUFDaEMsc0JBQXNCLFFBQVE7QUFDOUI7QUFDQSxHQUFHO0FBQ0g7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBLEtBQUs7QUFDTDtBQUNBLGtCQUFrQixRQUFRO0FBQzFCLGlCQUFpQixRQUFRO0FBQ3pCLGVBQWUsUUFBUTtBQUN2QjtBQUNBLEdBQUc7QUFDSDtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0EsS0FBSztBQUNMO0FBQ0EseUJBQXlCLFFBQVE7QUFDakMsd0JBQXdCLFFBQVE7QUFDaEMsc0JBQXNCLFFBQVE7QUFDOUI7QUFDQSxHQUFHO0FBQ0g7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBLEtBQUs7QUFDTDtBQUNBLGtCQUFrQixRQUFRO0FBQzFCLGlCQUFpQixRQUFRO0FBQ3pCLGVBQWUsUUFBUTtBQUN2QjtBQUNBLEdBQUc7QUFDSDtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0EsS0FBSztBQUNMO0FBQ0Esa0JBQWtCLFFBQVE7QUFDMUIsaUJBQWlCLFFBQVE7QUFDekIsZUFBZSxRQUFRO0FBQ3ZCO0FBQ0EsR0FBRztBQUNIO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQSxLQUFLO0FBQ0w7QUFDQSx5QkFBeUIsUUFBUTtBQUNqQyx3QkFBd0IsUUFBUTtBQUNoQyxzQkFBc0IsUUFBUTtBQUM5QjtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBLE1BQU07QUFDTjtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBO0FBQ0E7QUFDQTtBQUNBLElBQUk7QUFDSjtBQUNBO0FBQ0EsMkJBQTJCLE9BQU87QUFDbEM7QUFDQTtBQUNBLGtCQUFlO0FBQ2YiLCJzb3VyY2VzIjpbIndlYnBhY2s6Ly9wcml0dW5sLWNsb3VkLy4vbm9kZV9tb2R1bGVzL2RhdGUtZm5zL2xvY2FsZS9rbi9fbGliL2Zvcm1hdERpc3RhbmNlL2luZGV4LmpzPzc0ZjEiXSwic291cmNlc0NvbnRlbnQiOlsiXCJ1c2Ugc3RyaWN0XCI7XG5cbk9iamVjdC5kZWZpbmVQcm9wZXJ0eShleHBvcnRzLCBcIl9fZXNNb2R1bGVcIiwge1xuICB2YWx1ZTogdHJ1ZVxufSk7XG5leHBvcnRzLmRlZmF1bHQgPSB2b2lkIDA7XG4vLyBub3RlOiBubyBpbXBsZW1lbnRhdGlvbiBmb3Igd2Vla3NcblxudmFyIGZvcm1hdERpc3RhbmNlTG9jYWxlID0ge1xuICBsZXNzVGhhblhTZWNvbmRzOiB7XG4gICAgb25lOiB7XG4gICAgICBkZWZhdWx0OiAnMSDgsrjgs4bgspXgs4bgsoLgsqHgs43igIzgspfgsr/gsoLgsqQg4LKV4LKh4LK/4LKu4LOGJyxcbiAgICAgIGZ1dHVyZTogJzEg4LK44LOG4LKV4LOG4LKC4LKh4LON4oCM4LKX4LK/4LKC4LKkIOCyleCyoeCyv+CyruCzhicsXG4gICAgICBwYXN0OiAnMSDgsrjgs4bgspXgs4bgsoLgsqHgs43igIzgspfgsr/gsoLgsqQg4LKV4LKh4LK/4LKu4LOGJ1xuICAgIH0sXG4gICAgb3RoZXI6IHtcbiAgICAgIGRlZmF1bHQ6ICd7e2NvdW50fX0g4LK44LOG4LKV4LOG4LKC4LKh4LON4oCM4LKX4LK/4LKC4LKkIOCyleCyoeCyv+CyruCzhicsXG4gICAgICBmdXR1cmU6ICd7e2NvdW50fX0g4LK44LOG4LKV4LOG4LKC4LKh4LON4oCM4LKX4LK/4LKC4LKkIOCyleCyoeCyv+CyruCzhicsXG4gICAgICBwYXN0OiAne3tjb3VudH19IOCyuOCzhuCyleCzhuCyguCyoeCzjeKAjOCyl+Cyv+CyguCypCDgspXgsqHgsr/gsq7gs4YnXG4gICAgfVxuICB9LFxuICB4U2Vjb25kczoge1xuICAgIG9uZToge1xuICAgICAgZGVmYXVsdDogJzEg4LK44LOG4LKV4LOG4LKC4LKh4LONJyxcbiAgICAgIGZ1dHVyZTogJzEg4LK44LOG4LKV4LOG4LKC4LKh4LON4oCM4LKo4LKy4LON4LKy4LK/JyxcbiAgICAgIHBhc3Q6ICcxIOCyuOCzhuCyleCzhuCyguCyoeCzjSDgsrngsr/gsoLgsqbgs4YnXG4gICAgfSxcbiAgICBvdGhlcjoge1xuICAgICAgZGVmYXVsdDogJ3t7Y291bnR9fSDgsrjgs4bgspXgs4bgsoLgsqHgs4HgspfgsrPgs4EnLFxuICAgICAgZnV0dXJlOiAne3tjb3VudH19IOCyuOCzhuCyleCzhuCyguCyoeCzjeKAjOCyl+Cys+CysuCzjeCysuCyvycsXG4gICAgICBwYXN0OiAne3tjb3VudH19IOCyuOCzhuCyleCzhuCyguCyoeCzjSDgsrngsr/gsoLgsqbgs4YnXG4gICAgfVxuICB9LFxuICBoYWxmQU1pbnV0ZToge1xuICAgIG90aGVyOiB7XG4gICAgICBkZWZhdWx0OiAn4LKF4LKw4LON4LKnIOCyqOCyv+CyruCyv+CytycsXG4gICAgICBmdXR1cmU6ICfgsoXgsrDgs43gsqcg4LKo4LK/4LKu4LK/4LK34LKm4LKy4LON4LKy4LK/JyxcbiAgICAgIHBhc3Q6ICfgsoXgsrDgs43gsqcg4LKo4LK/4LKu4LK/4LK34LKmIOCyueCyv+CyguCypuCzhidcbiAgICB9XG4gIH0sXG4gIGxlc3NUaGFuWE1pbnV0ZXM6IHtcbiAgICBvbmU6IHtcbiAgICAgIGRlZmF1bHQ6ICcxIOCyqOCyv+CyruCyv+Cyt+CyleCzjeCyleCyv+CyguCypCDgspXgsqHgsr/gsq7gs4YnLFxuICAgICAgZnV0dXJlOiAnMSDgsqjgsr/gsq7gsr/gsrfgspXgs43gspXgsr/gsoLgsqQg4LKV4LKh4LK/4LKu4LOGJyxcbiAgICAgIHBhc3Q6ICcxIOCyqOCyv+CyruCyv+Cyt+CyleCzjeCyleCyv+CyguCypCDgspXgsqHgsr/gsq7gs4YnXG4gICAgfSxcbiAgICBvdGhlcjoge1xuICAgICAgZGVmYXVsdDogJ3t7Y291bnR9fSDgsqjgsr/gsq7gsr/gsrfgspXgs43gspXgsr/gsoLgsqQg4LKV4LKh4LK/4LKu4LOGJyxcbiAgICAgIGZ1dHVyZTogJ3t7Y291bnR9fSDgsqjgsr/gsq7gsr/gsrfgspXgs43gspXgsr/gsoLgsqQg4LKV4LKh4LK/4LKu4LOGJyxcbiAgICAgIHBhc3Q6ICd7e2NvdW50fX0g4LKo4LK/4LKu4LK/4LK34LKV4LON4LKV4LK/4LKC4LKkIOCyleCyoeCyv+CyruCzhidcbiAgICB9XG4gIH0sXG4gIHhNaW51dGVzOiB7XG4gICAgb25lOiB7XG4gICAgICBkZWZhdWx0OiAnMSDgsqjgsr/gsq7gsr/gsrcnLFxuICAgICAgZnV0dXJlOiAnMSDgsqjgsr/gsq7gsr/gsrfgsqbgsrLgs43gsrLgsr8nLFxuICAgICAgcGFzdDogJzEg4LKo4LK/4LKu4LK/4LK34LKmIOCyueCyv+CyguCypuCzhidcbiAgICB9LFxuICAgIG90aGVyOiB7XG4gICAgICBkZWZhdWx0OiAne3tjb3VudH19IOCyqOCyv+CyruCyv+Cyt+Cyl+Cys+CzgScsXG4gICAgICBmdXR1cmU6ICd7e2NvdW50fX0g4LKo4LK/4LKu4LK/4LK34LKX4LKz4LKy4LON4LKy4LK/JyxcbiAgICAgIHBhc3Q6ICd7e2NvdW50fX0g4LKo4LK/4LKu4LK/4LK34LKX4LKzIOCyueCyv+CyguCypuCzhidcbiAgICB9XG4gIH0sXG4gIGFib3V0WEhvdXJzOiB7XG4gICAgb25lOiB7XG4gICAgICBkZWZhdWx0OiAn4LK44LOB4LKu4LK+4LKw4LOBIDEg4LKX4LKC4LKf4LOGJyxcbiAgICAgIGZ1dHVyZTogJ+CyuOCzgeCyruCyvuCysOCzgSAxIOCyl+CyguCyn+CzhuCyr+CysuCzjeCysuCyvycsXG4gICAgICBwYXN0OiAn4LK44LOB4LKu4LK+4LKw4LOBIDEg4LKX4LKC4LKf4LOGIOCyueCyv+CyguCypuCzhidcbiAgICB9LFxuICAgIG90aGVyOiB7XG4gICAgICBkZWZhdWx0OiAn4LK44LOB4LKu4LK+4LKw4LOBIHt7Y291bnR9fSDgspfgsoLgsp/gs4bgspfgsrPgs4EnLFxuICAgICAgZnV0dXJlOiAn4LK44LOB4LKu4LK+4LKw4LOBIHt7Y291bnR9fSDgspfgsoLgsp/gs4bgspfgsrPgsrLgs43gsrLgsr8nLFxuICAgICAgcGFzdDogJ+CyuOCzgeCyruCyvuCysOCzgSB7e2NvdW50fX0g4LKX4LKC4LKf4LOG4LKX4LKzIOCyueCyv+CyguCypuCzhidcbiAgICB9XG4gIH0sXG4gIHhIb3Vyczoge1xuICAgIG9uZToge1xuICAgICAgZGVmYXVsdDogJzEg4LKX4LKC4LKf4LOGJyxcbiAgICAgIGZ1dHVyZTogJzEg4LKX4LKC4LKf4LOG4LKv4LKy4LON4LKy4LK/JyxcbiAgICAgIHBhc3Q6ICcxIOCyl+CyguCyn+CzhiDgsrngsr/gsoLgsqbgs4YnXG4gICAgfSxcbiAgICBvdGhlcjoge1xuICAgICAgZGVmYXVsdDogJ3t7Y291bnR9fSDgspfgsoLgsp/gs4bgspfgsrPgs4EnLFxuICAgICAgZnV0dXJlOiAne3tjb3VudH19IOCyl+CyguCyn+CzhuCyl+Cys+CysuCzjeCysuCyvycsXG4gICAgICBwYXN0OiAne3tjb3VudH19IOCyl+CyguCyn+CzhuCyl+CysyDgsrngsr/gsoLgsqbgs4YnXG4gICAgfVxuICB9LFxuICB4RGF5czoge1xuICAgIG9uZToge1xuICAgICAgZGVmYXVsdDogJzEg4LKm4LK/4LKoJyxcbiAgICAgIGZ1dHVyZTogJzEg4LKm4LK/4LKo4LKm4LKy4LON4LKy4LK/JyxcbiAgICAgIHBhc3Q6ICcxIOCypuCyv+CyqOCypiDgsrngsr/gsoLgsqbgs4YnXG4gICAgfSxcbiAgICBvdGhlcjoge1xuICAgICAgZGVmYXVsdDogJ3t7Y291bnR9fSDgsqbgsr/gsqjgspfgsrPgs4EnLFxuICAgICAgZnV0dXJlOiAne3tjb3VudH19IOCypuCyv+CyqOCyl+Cys+CysuCzjeCysuCyvycsXG4gICAgICBwYXN0OiAne3tjb3VudH19IOCypuCyv+CyqOCyl+CysyDgsrngsr/gsoLgsqbgs4YnXG4gICAgfVxuICB9LFxuICAvLyBUT0RPXG4gIC8vIGFib3V0WFdlZWtzOiB7fSxcblxuICAvLyBUT0RPXG4gIC8vIHhXZWVrczoge30sXG5cbiAgYWJvdXRYTW9udGhzOiB7XG4gICAgb25lOiB7XG4gICAgICBkZWZhdWx0OiAn4LK44LOB4LKu4LK+4LKw4LOBIDEg4LKk4LK/4LKC4LKX4LKz4LOBJyxcbiAgICAgIGZ1dHVyZTogJ+CyuOCzgeCyruCyvuCysOCzgSAxIOCypOCyv+CyguCyl+Cys+CysuCzjeCysuCyvycsXG4gICAgICBwYXN0OiAn4LK44LOB4LKu4LK+4LKw4LOBIDEg4LKk4LK/4LKC4LKX4LKzIOCyueCyv+CyguCypuCzhidcbiAgICB9LFxuICAgIG90aGVyOiB7XG4gICAgICBkZWZhdWx0OiAn4LK44LOB4LKu4LK+4LKw4LOBIHt7Y291bnR9fSDgsqTgsr/gsoLgspfgsrPgs4EnLFxuICAgICAgZnV0dXJlOiAn4LK44LOB4LKu4LK+4LKw4LOBIHt7Y291bnR9fSDgsqTgsr/gsoLgspfgsrPgs4HgspfgsrPgsrLgs43gsrLgsr8nLFxuICAgICAgcGFzdDogJ+CyuOCzgeCyruCyvuCysOCzgSB7e2NvdW50fX0g4LKk4LK/4LKC4LKX4LKz4LOB4LKX4LKzIOCyueCyv+CyguCypuCzhidcbiAgICB9XG4gIH0sXG4gIHhNb250aHM6IHtcbiAgICBvbmU6IHtcbiAgICAgIGRlZmF1bHQ6ICcxIOCypOCyv+CyguCyl+Cys+CzgScsXG4gICAgICBmdXR1cmU6ICcxIOCypOCyv+CyguCyl+Cys+CysuCzjeCysuCyvycsXG4gICAgICBwYXN0OiAnMSDgsqTgsr/gsoLgspfgsrMg4LK54LK/4LKC4LKm4LOGJ1xuICAgIH0sXG4gICAgb3RoZXI6IHtcbiAgICAgIGRlZmF1bHQ6ICd7e2NvdW50fX0g4LKk4LK/4LKC4LKX4LKz4LOBJyxcbiAgICAgIGZ1dHVyZTogJ3t7Y291bnR9fSDgsqTgsr/gsoLgspfgsrPgs4HgspfgsrPgsrLgs43gsrLgsr8nLFxuICAgICAgcGFzdDogJ3t7Y291bnR9fSDgsqTgsr/gsoLgspfgsrPgs4HgspfgsrMg4LK54LK/4LKC4LKm4LOGJ1xuICAgIH1cbiAgfSxcbiAgYWJvdXRYWWVhcnM6IHtcbiAgICBvbmU6IHtcbiAgICAgIGRlZmF1bHQ6ICfgsrjgs4Hgsq7gsr7gsrDgs4EgMSDgsrXgsrDgs43gsrcnLFxuICAgICAgZnV0dXJlOiAn4LK44LOB4LKu4LK+4LKw4LOBIDEg4LK14LKw4LON4LK34LKm4LKy4LON4LKy4LK/JyxcbiAgICAgIHBhc3Q6ICfgsrjgs4Hgsq7gsr7gsrDgs4EgMSDgsrXgsrDgs43gsrfgsqYg4LK54LK/4LKC4LKm4LOGJ1xuICAgIH0sXG4gICAgb3RoZXI6IHtcbiAgICAgIGRlZmF1bHQ6ICfgsrjgs4Hgsq7gsr7gsrDgs4Ege3tjb3VudH19IOCyteCysOCzjeCyt+Cyl+Cys+CzgScsXG4gICAgICBmdXR1cmU6ICfgsrjgs4Hgsq7gsr7gsrDgs4Ege3tjb3VudH19IOCyteCysOCzjeCyt+Cyl+Cys+CysuCzjeCysuCyvycsXG4gICAgICBwYXN0OiAn4LK44LOB4LKu4LK+4LKw4LOBIHt7Y291bnR9fSDgsrXgsrDgs43gsrfgspfgsrMg4LK54LK/4LKC4LKm4LOGJ1xuICAgIH1cbiAgfSxcbiAgeFllYXJzOiB7XG4gICAgb25lOiB7XG4gICAgICBkZWZhdWx0OiAnMSDgsrXgsrDgs43gsrcnLFxuICAgICAgZnV0dXJlOiAnMSDgsrXgsrDgs43gsrfgsqbgsrLgs43gsrLgsr8nLFxuICAgICAgcGFzdDogJzEg4LK14LKw4LON4LK34LKmIOCyueCyv+CyguCypuCzhidcbiAgICB9LFxuICAgIG90aGVyOiB7XG4gICAgICBkZWZhdWx0OiAne3tjb3VudH19IOCyteCysOCzjeCyt+Cyl+Cys+CzgScsXG4gICAgICBmdXR1cmU6ICd7e2NvdW50fX0g4LK14LKw4LON4LK34LKX4LKz4LKy4LON4LKy4LK/JyxcbiAgICAgIHBhc3Q6ICd7e2NvdW50fX0g4LK14LKw4LON4LK34LKX4LKzIOCyueCyv+CyguCypuCzhidcbiAgICB9XG4gIH0sXG4gIG92ZXJYWWVhcnM6IHtcbiAgICBvbmU6IHtcbiAgICAgIGRlZmF1bHQ6ICcxIOCyteCysOCzjeCyt+CypiDgsq7gs4fgsrLgs4YnLFxuICAgICAgZnV0dXJlOiAnMSDgsrXgsrDgs43gsrfgsqYg4LKu4LOH4LKy4LOGJyxcbiAgICAgIHBhc3Q6ICcxIOCyteCysOCzjeCyt+CypiDgsq7gs4fgsrLgs4YnXG4gICAgfSxcbiAgICBvdGhlcjoge1xuICAgICAgZGVmYXVsdDogJ3t7Y291bnR9fSDgsrXgsrDgs43gsrfgspfgsrMg4LKu4LOH4LKy4LOGJyxcbiAgICAgIGZ1dHVyZTogJ3t7Y291bnR9fSDgsrXgsrDgs43gsrfgspfgsrMg4LKu4LOH4LKy4LOGJyxcbiAgICAgIHBhc3Q6ICd7e2NvdW50fX0g4LK14LKw4LON4LK34LKX4LKzIOCyruCzh+CysuCzhidcbiAgICB9XG4gIH0sXG4gIGFsbW9zdFhZZWFyczoge1xuICAgIG9uZToge1xuICAgICAgZGVmYXVsdDogJ+CyrOCyueCzgeCypOCzh+CylSAxIOCyteCysOCzjeCyt+CypuCysuCzjeCysuCyvycsXG4gICAgICBmdXR1cmU6ICfgsqzgsrngs4HgsqTgs4fgspUgMSDgsrXgsrDgs43gsrfgsqbgsrLgs43gsrLgsr8nLFxuICAgICAgcGFzdDogJ+CyrOCyueCzgeCypOCzh+CylSAxIOCyteCysOCzjeCyt+CypuCysuCzjeCysuCyvydcbiAgICB9LFxuICAgIG90aGVyOiB7XG4gICAgICBkZWZhdWx0OiAn4LKs4LK54LOB4LKk4LOH4LKVIHt7Y291bnR9fSDgsrXgsrDgs43gsrfgspfgsrPgsrLgs43gsrLgsr8nLFxuICAgICAgZnV0dXJlOiAn4LKs4LK54LOB4LKk4LOH4LKVIHt7Y291bnR9fSDgsrXgsrDgs43gsrfgspfgsrPgsrLgs43gsrLgsr8nLFxuICAgICAgcGFzdDogJ+CyrOCyueCzgeCypOCzh+CylSB7e2NvdW50fX0g4LK14LKw4LON4LK34LKX4LKz4LKy4LON4LKy4LK/J1xuICAgIH1cbiAgfVxufTtcbmZ1bmN0aW9uIGdldFJlc3VsdEJ5VGVuc2UocGFyZW50VG9rZW4sIG9wdGlvbnMpIHtcbiAgaWYgKG9wdGlvbnMgIT09IG51bGwgJiYgb3B0aW9ucyAhPT0gdm9pZCAwICYmIG9wdGlvbnMuYWRkU3VmZml4KSB7XG4gICAgaWYgKG9wdGlvbnMuY29tcGFyaXNvbiAmJiBvcHRpb25zLmNvbXBhcmlzb24gPiAwKSB7XG4gICAgICByZXR1cm4gcGFyZW50VG9rZW4uZnV0dXJlO1xuICAgIH0gZWxzZSB7XG4gICAgICByZXR1cm4gcGFyZW50VG9rZW4ucGFzdDtcbiAgICB9XG4gIH1cbiAgcmV0dXJuIHBhcmVudFRva2VuLmRlZmF1bHQ7XG59XG52YXIgZm9ybWF0RGlzdGFuY2UgPSBmdW5jdGlvbiBmb3JtYXREaXN0YW5jZSh0b2tlbiwgY291bnQsIG9wdGlvbnMpIHtcbiAgdmFyIHJlc3VsdDtcbiAgdmFyIHRva2VuVmFsdWUgPSBmb3JtYXREaXN0YW5jZUxvY2FsZVt0b2tlbl07XG4gIGlmICh0b2tlblZhbHVlLm9uZSAmJiBjb3VudCA9PT0gMSkge1xuICAgIHJlc3VsdCA9IGdldFJlc3VsdEJ5VGVuc2UodG9rZW5WYWx1ZS5vbmUsIG9wdGlvbnMpO1xuICB9IGVsc2Uge1xuICAgIHJlc3VsdCA9IGdldFJlc3VsdEJ5VGVuc2UodG9rZW5WYWx1ZS5vdGhlciwgb3B0aW9ucyk7XG4gIH1cbiAgcmV0dXJuIHJlc3VsdC5yZXBsYWNlKCd7e2NvdW50fX0nLCBTdHJpbmcoY291bnQpKTtcbn07XG52YXIgX2RlZmF1bHQgPSBmb3JtYXREaXN0YW5jZTtcbmV4cG9ydHMuZGVmYXVsdCA9IF9kZWZhdWx0O1xubW9kdWxlLmV4cG9ydHMgPSBleHBvcnRzLmRlZmF1bHQ7Il0sIm5hbWVzIjpbXSwic291cmNlUm9vdCI6IiJ9\n//# sourceURL=webpack-internal:///./node_modules/date-fns/locale/kn/_lib/formatDistance/index.js\n");

/***/ })

}]);