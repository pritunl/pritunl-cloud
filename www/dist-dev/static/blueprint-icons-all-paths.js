"use strict";
/*
 * ATTENTION: An "eval-source-map" devtool has been used.
 * This devtool is neither made for production nor for readable output files.
 * It uses "eval()" calls to create a separate source file with attached SourceMaps in the browser devtools.
 * If you are trying to read the output file, select a different devtool (https://webpack.js.org/configuration/devtool/)
 * or disable the default devtool with "devtool: false".
 * If you are looking for production-ready output files, see mode: "production" (https://webpack.js.org/configuration/mode/).
 */
(self["webpackChunkpritunl_cloud"] = self["webpackChunkpritunl_cloud"] || []).push([["blueprint-icons-all-paths"],{

/***/ "./node_modules/@blueprintjs/icons/lib/esm/allPaths.js":
/*!*************************************************************!*\
  !*** ./node_modules/@blueprintjs/icons/lib/esm/allPaths.js ***!
  \*************************************************************/
/***/ ((__unused_webpack_module, __webpack_exports__, __webpack_require__) => {

eval("__webpack_require__.r(__webpack_exports__);\n/* harmony export */ __webpack_require__.d(__webpack_exports__, {\n/* harmony export */   IconSvgPaths16: () => (/* reexport module object */ _generated_16px_paths__WEBPACK_IMPORTED_MODULE_0__),\n/* harmony export */   IconSvgPaths20: () => (/* reexport module object */ _generated_20px_paths__WEBPACK_IMPORTED_MODULE_1__),\n/* harmony export */   getIconPaths: () => (/* binding */ getIconPaths),\n/* harmony export */   iconNameToPathsRecordKey: () => (/* binding */ iconNameToPathsRecordKey)\n/* harmony export */ });\n/* harmony import */ var change_case__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! change-case */ \"./node_modules/pascal-case/dist.es2015/index.js\");\n/* harmony import */ var _generated_16px_paths__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! ./generated/16px/paths */ \"./node_modules/@blueprintjs/icons/lib/esm/generated/16px/paths/index.js\");\n/* harmony import */ var _generated_20px_paths__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! ./generated/20px/paths */ \"./node_modules/@blueprintjs/icons/lib/esm/generated/20px/paths/index.js\");\n/* harmony import */ var _iconTypes__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ./iconTypes */ \"./node_modules/@blueprintjs/icons/lib/esm/iconTypes.js\");\n/*\n * Copyright 2021 Palantir Technologies, Inc. All rights reserved.\n *\n * Licensed under the Apache License, Version 2.0 (the \"License\");\n * you may not use this file except in compliance with the License.\n * You may obtain a copy of the License at\n *\n *     http://www.apache.org/licenses/LICENSE-2.0\n *\n * Unless required by applicable law or agreed to in writing, software\n * distributed under the License is distributed on an \"AS IS\" BASIS,\n * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.\n * See the License for the specific language governing permissions and\n * limitations under the License.\n */\n\n\n\n\n\n/**\n * Get the list of vector paths that define a given icon. These path strings are used to render `<path>`\n * elements inside an `<svg>` icon element. For full implementation details and nuances, see the icon component\n * handlebars template and `generate-icon-components` script in the __@blueprintjs/icons__ package.\n *\n * Note: this function loads all icon definitions __statically__, which means every icon is included in your\n * JS bundle. Only use this API if your app is likely to use all Blueprint icons at runtime. If you are looking for a\n * dynamic icon loader which loads icon definitions on-demand, use `{ Icons } from \"@blueprintjs/icons\"` instead.\n */\nfunction getIconPaths(name, size) {\n    var key = (0,change_case__WEBPACK_IMPORTED_MODULE_2__.pascalCase)(name);\n    return size === _iconTypes__WEBPACK_IMPORTED_MODULE_3__.IconSize.STANDARD ? _generated_16px_paths__WEBPACK_IMPORTED_MODULE_0__[key] : _generated_20px_paths__WEBPACK_IMPORTED_MODULE_1__[key];\n}\n/**\n * Type safe string literal conversion of snake-case icon names to PascalCase icon names.\n * This is useful for indexing into the SVG paths record to extract a single icon's SVG path definition.\n *\n * @deprecated use `getIconPaths` instead\n */\nfunction iconNameToPathsRecordKey(name) {\n    return (0,change_case__WEBPACK_IMPORTED_MODULE_2__.pascalCase)(name);\n}\n//# sourceURL=[module]\n//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiLi9ub2RlX21vZHVsZXMvQGJsdWVwcmludGpzL2ljb25zL2xpYi9lc20vYWxsUGF0aHMuanMiLCJtYXBwaW5ncyI6Ijs7Ozs7Ozs7Ozs7QUFBQTs7Ozs7Ozs7Ozs7Ozs7R0FjRztBQUVzQztBQUVnQjtBQUNBO0FBRUY7QUFFYjtBQUUxQzs7Ozs7Ozs7R0FRRztBQUNJLFNBQVMsWUFBWSxDQUFDLElBQWMsRUFBRSxJQUFjO0lBQ3ZELElBQU0sR0FBRyxHQUFHLHVEQUFVLENBQUMsSUFBSSxDQUF5QixDQUFDO0lBQ3JELE9BQU8sSUFBSSxLQUFLLGdEQUFRLENBQUMsUUFBUSxDQUFDLENBQUMsQ0FBQyxrREFBYyxDQUFDLEdBQUcsQ0FBQyxDQUFDLENBQUMsQ0FBQyxrREFBYyxDQUFDLEdBQUcsQ0FBQyxDQUFDO0FBQ2xGLENBQUM7QUFFRDs7Ozs7R0FLRztBQUNJLFNBQVMsd0JBQXdCLENBQUMsSUFBYztJQUNuRCxPQUFPLHVEQUFVLENBQUMsSUFBSSxDQUF5QixDQUFDO0FBQ3BELENBQUMiLCJzb3VyY2VzIjpbIndlYnBhY2s6Ly9wcml0dW5sLWNsb3VkLy4vbm9kZV9tb2R1bGVzL0BibHVlcHJpbnRqcy9pY29ucy9zcmMvYWxsUGF0aHMudHM/NzEzYyJdLCJzb3VyY2VzQ29udGVudCI6WyIvKlxuICogQ29weXJpZ2h0IDIwMjEgUGFsYW50aXIgVGVjaG5vbG9naWVzLCBJbmMuIEFsbCByaWdodHMgcmVzZXJ2ZWQuXG4gKlxuICogTGljZW5zZWQgdW5kZXIgdGhlIEFwYWNoZSBMaWNlbnNlLCBWZXJzaW9uIDIuMCAodGhlIFwiTGljZW5zZVwiKTtcbiAqIHlvdSBtYXkgbm90IHVzZSB0aGlzIGZpbGUgZXhjZXB0IGluIGNvbXBsaWFuY2Ugd2l0aCB0aGUgTGljZW5zZS5cbiAqIFlvdSBtYXkgb2J0YWluIGEgY29weSBvZiB0aGUgTGljZW5zZSBhdFxuICpcbiAqICAgICBodHRwOi8vd3d3LmFwYWNoZS5vcmcvbGljZW5zZXMvTElDRU5TRS0yLjBcbiAqXG4gKiBVbmxlc3MgcmVxdWlyZWQgYnkgYXBwbGljYWJsZSBsYXcgb3IgYWdyZWVkIHRvIGluIHdyaXRpbmcsIHNvZnR3YXJlXG4gKiBkaXN0cmlidXRlZCB1bmRlciB0aGUgTGljZW5zZSBpcyBkaXN0cmlidXRlZCBvbiBhbiBcIkFTIElTXCIgQkFTSVMsXG4gKiBXSVRIT1VUIFdBUlJBTlRJRVMgT1IgQ09ORElUSU9OUyBPRiBBTlkgS0lORCwgZWl0aGVyIGV4cHJlc3Mgb3IgaW1wbGllZC5cbiAqIFNlZSB0aGUgTGljZW5zZSBmb3IgdGhlIHNwZWNpZmljIGxhbmd1YWdlIGdvdmVybmluZyBwZXJtaXNzaW9ucyBhbmRcbiAqIGxpbWl0YXRpb25zIHVuZGVyIHRoZSBMaWNlbnNlLlxuICovXG5cbmltcG9ydCB7IHBhc2NhbENhc2UgfSBmcm9tIFwiY2hhbmdlLWNhc2VcIjtcblxuaW1wb3J0ICogYXMgSWNvblN2Z1BhdGhzMTYgZnJvbSBcIi4vZ2VuZXJhdGVkLzE2cHgvcGF0aHNcIjtcbmltcG9ydCAqIGFzIEljb25TdmdQYXRoczIwIGZyb20gXCIuL2dlbmVyYXRlZC8yMHB4L3BhdGhzXCI7XG5pbXBvcnQgdHlwZSB7IEljb25OYW1lIH0gZnJvbSBcIi4vaWNvbk5hbWVzXCI7XG5pbXBvcnQgeyB0eXBlIEljb25QYXRocywgSWNvblNpemUgfSBmcm9tIFwiLi9pY29uVHlwZXNcIjtcbmltcG9ydCB0eXBlIHsgUGFzY2FsQ2FzZSB9IGZyb20gXCIuL3R5cGUtdXRpbHNcIjtcbmV4cG9ydCB7IEljb25TdmdQYXRoczE2LCBJY29uU3ZnUGF0aHMyMCB9O1xuXG4vKipcbiAqIEdldCB0aGUgbGlzdCBvZiB2ZWN0b3IgcGF0aHMgdGhhdCBkZWZpbmUgYSBnaXZlbiBpY29uLiBUaGVzZSBwYXRoIHN0cmluZ3MgYXJlIHVzZWQgdG8gcmVuZGVyIGA8cGF0aD5gXG4gKiBlbGVtZW50cyBpbnNpZGUgYW4gYDxzdmc+YCBpY29uIGVsZW1lbnQuIEZvciBmdWxsIGltcGxlbWVudGF0aW9uIGRldGFpbHMgYW5kIG51YW5jZXMsIHNlZSB0aGUgaWNvbiBjb21wb25lbnRcbiAqIGhhbmRsZWJhcnMgdGVtcGxhdGUgYW5kIGBnZW5lcmF0ZS1pY29uLWNvbXBvbmVudHNgIHNjcmlwdCBpbiB0aGUgX19AYmx1ZXByaW50anMvaWNvbnNfXyBwYWNrYWdlLlxuICpcbiAqIE5vdGU6IHRoaXMgZnVuY3Rpb24gbG9hZHMgYWxsIGljb24gZGVmaW5pdGlvbnMgX19zdGF0aWNhbGx5X18sIHdoaWNoIG1lYW5zIGV2ZXJ5IGljb24gaXMgaW5jbHVkZWQgaW4geW91clxuICogSlMgYnVuZGxlLiBPbmx5IHVzZSB0aGlzIEFQSSBpZiB5b3VyIGFwcCBpcyBsaWtlbHkgdG8gdXNlIGFsbCBCbHVlcHJpbnQgaWNvbnMgYXQgcnVudGltZS4gSWYgeW91IGFyZSBsb29raW5nIGZvciBhXG4gKiBkeW5hbWljIGljb24gbG9hZGVyIHdoaWNoIGxvYWRzIGljb24gZGVmaW5pdGlvbnMgb24tZGVtYW5kLCB1c2UgYHsgSWNvbnMgfSBmcm9tIFwiQGJsdWVwcmludGpzL2ljb25zXCJgIGluc3RlYWQuXG4gKi9cbmV4cG9ydCBmdW5jdGlvbiBnZXRJY29uUGF0aHMobmFtZTogSWNvbk5hbWUsIHNpemU6IEljb25TaXplKTogSWNvblBhdGhzIHtcbiAgICBjb25zdCBrZXkgPSBwYXNjYWxDYXNlKG5hbWUpIGFzIFBhc2NhbENhc2U8SWNvbk5hbWU+O1xuICAgIHJldHVybiBzaXplID09PSBJY29uU2l6ZS5TVEFOREFSRCA/IEljb25TdmdQYXRoczE2W2tleV0gOiBJY29uU3ZnUGF0aHMyMFtrZXldO1xufVxuXG4vKipcbiAqIFR5cGUgc2FmZSBzdHJpbmcgbGl0ZXJhbCBjb252ZXJzaW9uIG9mIHNuYWtlLWNhc2UgaWNvbiBuYW1lcyB0byBQYXNjYWxDYXNlIGljb24gbmFtZXMuXG4gKiBUaGlzIGlzIHVzZWZ1bCBmb3IgaW5kZXhpbmcgaW50byB0aGUgU1ZHIHBhdGhzIHJlY29yZCB0byBleHRyYWN0IGEgc2luZ2xlIGljb24ncyBTVkcgcGF0aCBkZWZpbml0aW9uLlxuICpcbiAqIEBkZXByZWNhdGVkIHVzZSBgZ2V0SWNvblBhdGhzYCBpbnN0ZWFkXG4gKi9cbmV4cG9ydCBmdW5jdGlvbiBpY29uTmFtZVRvUGF0aHNSZWNvcmRLZXkobmFtZTogSWNvbk5hbWUpOiBQYXNjYWxDYXNlPEljb25OYW1lPiB7XG4gICAgcmV0dXJuIHBhc2NhbENhc2UobmFtZSkgYXMgUGFzY2FsQ2FzZTxJY29uTmFtZT47XG59XG4iXSwibmFtZXMiOltdLCJzb3VyY2VSb290IjoiIn0=\n//# sourceURL=webpack-internal:///./node_modules/@blueprintjs/icons/lib/esm/allPaths.js\n");

/***/ })

}]);