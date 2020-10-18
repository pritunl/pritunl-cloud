"use strict";

function _typeof(obj) { "@babel/helpers - typeof"; if (typeof Symbol === "function" && typeof Symbol.iterator === "symbol") { _typeof = function _typeof(obj) { return typeof obj; }; } else { _typeof = function _typeof(obj) { return obj && typeof Symbol === "function" && obj.constructor === Symbol && obj !== Symbol.prototype ? "symbol" : typeof obj; }; } return _typeof(obj); }

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.default = void 0;

var Log = _interopRequireWildcard(require("../util/logging.js"));

function _getRequireWildcardCache() { if (typeof WeakMap !== "function") return null; var cache = new WeakMap(); _getRequireWildcardCache = function _getRequireWildcardCache() { return cache; }; return cache; }

function _interopRequireWildcard(obj) { if (obj && obj.__esModule) { return obj; } if (obj === null || _typeof(obj) !== "object" && typeof obj !== "function") { return { default: obj }; } var cache = _getRequireWildcardCache(); if (cache && cache.has(obj)) { return cache.get(obj); } var newObj = {}; var hasPropertyDescriptor = Object.defineProperty && Object.getOwnPropertyDescriptor; for (var key in obj) { if (Object.prototype.hasOwnProperty.call(obj, key)) { var desc = hasPropertyDescriptor ? Object.getOwnPropertyDescriptor(obj, key) : null; if (desc && (desc.get || desc.set)) { Object.defineProperty(newObj, key, desc); } else { newObj[key] = obj[key]; } } } newObj.default = obj; if (cache) { cache.set(obj, newObj); } return newObj; }

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } }

function _createClass(Constructor, protoProps, staticProps) { if (protoProps) _defineProperties(Constructor.prototype, protoProps); if (staticProps) _defineProperties(Constructor, staticProps); return Constructor; }

var HextileDecoder = /*#__PURE__*/function () {
  function HextileDecoder() {
    _classCallCheck(this, HextileDecoder);

    this._tiles = 0;
    this._lastsubencoding = 0;
  }

  _createClass(HextileDecoder, [{
    key: "decodeRect",
    value: function decodeRect(x, y, width, height, sock, display, depth) {
      if (this._tiles === 0) {
        this._tilesX = Math.ceil(width / 16);
        this._tilesY = Math.ceil(height / 16);
        this._totalTiles = this._tilesX * this._tilesY;
        this._tiles = this._totalTiles;
      }

      while (this._tiles > 0) {
        var bytes = 1;

        if (sock.rQwait("HEXTILE", bytes)) {
          return false;
        }

        var rQ = sock.rQ;
        var rQi = sock.rQi;
        var subencoding = rQ[rQi]; // Peek

        if (subencoding > 30) {
          // Raw
          throw new Error("Illegal hextile subencoding (subencoding: " + subencoding + ")");
        }

        var currTile = this._totalTiles - this._tiles;
        var tileX = currTile % this._tilesX;
        var tileY = Math.floor(currTile / this._tilesX);
        var tx = x + tileX * 16;
        var ty = y + tileY * 16;
        var tw = Math.min(16, x + width - tx);
        var th = Math.min(16, y + height - ty); // Figure out how much we are expecting

        if (subencoding & 0x01) {
          // Raw
          bytes += tw * th * 4;
        } else {
          if (subencoding & 0x02) {
            // Background
            bytes += 4;
          }

          if (subencoding & 0x04) {
            // Foreground
            bytes += 4;
          }

          if (subencoding & 0x08) {
            // AnySubrects
            bytes++; // Since we aren't shifting it off

            if (sock.rQwait("HEXTILE", bytes)) {
              return false;
            }

            var subrects = rQ[rQi + bytes - 1]; // Peek

            if (subencoding & 0x10) {
              // SubrectsColoured
              bytes += subrects * (4 + 2);
            } else {
              bytes += subrects * 2;
            }
          }
        }

        if (sock.rQwait("HEXTILE", bytes)) {
          return false;
        } // We know the encoding and have a whole tile


        rQi++;

        if (subencoding === 0) {
          if (this._lastsubencoding & 0x01) {
            // Weird: ignore blanks are RAW
            Log.Debug("     Ignoring blank after RAW");
          } else {
            display.fillRect(tx, ty, tw, th, this._background);
          }
        } else if (subencoding & 0x01) {
          // Raw
          display.blitImage(tx, ty, tw, th, rQ, rQi);
          rQi += bytes - 1;
        } else {
          if (subencoding & 0x02) {
            // Background
            this._background = [rQ[rQi], rQ[rQi + 1], rQ[rQi + 2], rQ[rQi + 3]];
            rQi += 4;
          }

          if (subencoding & 0x04) {
            // Foreground
            this._foreground = [rQ[rQi], rQ[rQi + 1], rQ[rQi + 2], rQ[rQi + 3]];
            rQi += 4;
          }

          display.startTile(tx, ty, tw, th, this._background);

          if (subencoding & 0x08) {
            // AnySubrects
            var _subrects = rQ[rQi];
            rQi++;

            for (var s = 0; s < _subrects; s++) {
              var color = void 0;

              if (subencoding & 0x10) {
                // SubrectsColoured
                color = [rQ[rQi], rQ[rQi + 1], rQ[rQi + 2], rQ[rQi + 3]];
                rQi += 4;
              } else {
                color = this._foreground;
              }

              var xy = rQ[rQi];
              rQi++;
              var sx = xy >> 4;
              var sy = xy & 0x0f;
              var wh = rQ[rQi];
              rQi++;
              var sw = (wh >> 4) + 1;
              var sh = (wh & 0x0f) + 1;
              display.subTile(sx, sy, sw, sh, color);
            }
          }

          display.finishTile();
        }

        sock.rQi = rQi;
        this._lastsubencoding = subencoding;
        this._tiles--;
      }

      return true;
    }
  }]);

  return HextileDecoder;
}();

exports.default = HextileDecoder;