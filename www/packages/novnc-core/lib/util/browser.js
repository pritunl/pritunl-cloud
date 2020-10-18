"use strict";

function _typeof(obj) { "@babel/helpers - typeof"; if (typeof Symbol === "function" && typeof Symbol.iterator === "symbol") { _typeof = function _typeof(obj) { return typeof obj; }; } else { _typeof = function _typeof(obj) { return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj; }; } return _typeof(obj); }

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.isMac = isMac;
exports.isWindows = isWindows;
exports.isIOS = isIOS;
exports.isSafari = isSafari;
exports.isIE = isIE;
exports.isEdge = isEdge;
exports.isFirefox = isFirefox;
exports.hasScrollbarGutter = exports.supportsImageMetadata = exports.supportsCursorURIs = exports.dragThreshold = exports.isTouchDevice = void 0;

var Log = _interopRequireWildcard(require("./logging.js"));

function _getRequireWildcardCache() { if (typeof WeakMap !== "function") return null; var cache = new WeakMap(); _getRequireWildcardCache = function _getRequireWildcardCache() { return cache; }; return cache; }

function _interopRequireWildcard(obj) { if (obj && obj.__esModule) { return obj; } if (obj === null || _typeof(obj) !== "object" && typeof obj !== "function") { return { default: obj }; } var cache = _getRequireWildcardCache(); if (cache && cache.has(obj)) { return cache.get(obj); } var newObj = {}; var hasPropertyDescriptor = Object.defineProperty && Object.getOwnPropertyDescriptor; for (var key in obj) { if (Object.prototype.hasOwnProperty.call(obj, key)) { var desc = hasPropertyDescriptor ? Object.getOwnPropertyDescriptor(obj, key) : null; if (desc && (desc.get || desc.set)) { Object.defineProperty(newObj, key, desc); } else { newObj[key] = obj[key]; } } } newObj.default = obj; if (cache) { cache.set(obj, newObj); } return newObj; }

/*
 * noVNC: HTML5 VNC client
 * Copyright (C) 2019 The noVNC Authors
 * Licensed under MPL 2.0 (see LICENSE.txt)
 *
 * See README.md for usage and integration instructions.
 *
 * Browser feature support detection
 */
// Touch detection
var isTouchDevice = 'ontouchstart' in document.documentElement || // requried for Chrome debugger
document.ontouchstart !== undefined || // required for MS Surface
navigator.maxTouchPoints > 0 || navigator.msMaxTouchPoints > 0;
exports.isTouchDevice = isTouchDevice;
window.addEventListener('touchstart', function onFirstTouch() {
  exports.isTouchDevice = isTouchDevice = true;
  window.removeEventListener('touchstart', onFirstTouch, false);
}, false); // The goal is to find a certain physical width, the devicePixelRatio
// brings us a bit closer but is not optimal.

var dragThreshold = 10 * (window.devicePixelRatio || 1);
exports.dragThreshold = dragThreshold;
var _supportsCursorURIs = false;

try {
  var target = document.createElement('canvas');
  target.style.cursor = 'url("data:image/x-icon;base64,AAACAAEACAgAAAIAAgA4AQAAFgAAACgAAAAIAAAAEAAAAAEAIAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAD/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////AAAAAAAAAAAAAAAAAAAAAA==") 2 2, default';

  if (target.style.cursor.indexOf("url") === 0) {
    Log.Info("Data URI scheme cursor supported");
    _supportsCursorURIs = true;
  } else {
    Log.Warn("Data URI scheme cursor not supported");
  }
} catch (exc) {
  Log.Error("Data URI scheme cursor test exception: " + exc);
}

var supportsCursorURIs = _supportsCursorURIs;
exports.supportsCursorURIs = supportsCursorURIs;
var _supportsImageMetadata = false;

try {
  new ImageData(new Uint8ClampedArray(4), 1, 1);
  _supportsImageMetadata = true;
} catch (ex) {// ignore failure
}

var supportsImageMetadata = _supportsImageMetadata;
exports.supportsImageMetadata = supportsImageMetadata;
var _hasScrollbarGutter = true;

try {
  // Create invisible container
  var container = document.createElement('div');
  container.style.visibility = 'hidden';
  container.style.overflow = 'scroll'; // forcing scrollbars

  document.body.appendChild(container); // Create a div and place it in the container

  var child = document.createElement('div');
  container.appendChild(child); // Calculate the difference between the container's full width
  // and the child's width - the difference is the scrollbars

  var scrollbarWidth = container.offsetWidth - child.offsetWidth; // Clean up

  container.parentNode.removeChild(container);
  _hasScrollbarGutter = scrollbarWidth != 0;
} catch (exc) {
  Log.Error("Scrollbar test exception: " + exc);
}

var hasScrollbarGutter = _hasScrollbarGutter;
/*
 * The functions for detection of platforms and browsers below are exported
 * but the use of these should be minimized as much as possible.
 *
 * It's better to use feature detection than platform detection.
 */

exports.hasScrollbarGutter = hasScrollbarGutter;

function isMac() {
  return navigator && !!/mac/i.exec(navigator.platform);
}

function isWindows() {
  return navigator && !!/win/i.exec(navigator.platform);
}

function isIOS() {
  return navigator && (!!/ipad/i.exec(navigator.platform) || !!/iphone/i.exec(navigator.platform) || !!/ipod/i.exec(navigator.platform));
}

function isSafari() {
  return navigator && navigator.userAgent.indexOf('Safari') !== -1 && navigator.userAgent.indexOf('Chrome') === -1;
}

function isIE() {
  return navigator && !!/trident/i.exec(navigator.userAgent);
}

function isEdge() {
  return navigator && !!/edge/i.exec(navigator.userAgent);
}

function isFirefox() {
  return navigator && !!/firefox/i.exec(navigator.userAgent);
}