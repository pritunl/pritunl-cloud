(self.webpackChunkpritunl_cloud=self.webpackChunkpritunl_cloud||[]).push([[62442,46144,70140],{94461:(e,a)=>{"use strict";Object.defineProperty(a,"__esModule",{value:!0}),a.default=function(e){return function(a){var t=arguments.length>1&&void 0!==arguments[1]?arguments[1]:{},i=t.width,r=i&&e.matchPatterns[i]||e.matchPatterns[e.defaultMatchWidth],n=a.match(r);if(!n)return null;var l,o=n[0],d=i&&e.parsePatterns[i]||e.parsePatterns[e.defaultParseWidth],u=Array.isArray(d)?function(e,a){for(var t=0;t<e.length;t++)if(a(e[t]))return t;return}(d,(function(e){return e.test(o)})):function(e,a){for(var t in e)if(e.hasOwnProperty(t)&&a(e[t]))return t;return}(d,(function(e){return e.test(o)}));return l=e.valueCallback?e.valueCallback(u):u,{value:l=t.valueCallback?t.valueCallback(l):l,rest:a.slice(o.length)}}},e.exports=a.default},44497:(e,a)=>{"use strict";Object.defineProperty(a,"__esModule",{value:!0}),a.default=function(e){return function(a){var t=arguments.length>1&&void 0!==arguments[1]?arguments[1]:{},i=a.match(e.matchPattern);if(!i)return null;var r=i[0],n=a.match(e.parsePattern);if(!n)return null;var l=e.valueCallback?e.valueCallback(n[0]):n[0];return{value:l=t.valueCallback?t.valueCallback(l):l,rest:a.slice(r.length)}}},e.exports=a.default},75317:(e,a,t)=>{"use strict";var i=t(1654).default;Object.defineProperty(a,"__esModule",{value:!0}),a.default=void 0;var r=i(t(94461)),n={ordinalNumber:(0,i(t(44497)).default)({matchPattern:/^(\d+)(º)?/i,parsePattern:/\d+/i,valueCallback:function(e){return parseInt(e,10)}}),era:(0,r.default)({matchPatterns:{narrow:/^(aC|dC)/i,abbreviated:/^(a\.?\s?C\.?|a\.?\s?e\.?\s?v\.?|d\.?\s?C\.?|e\.?\s?v\.?)/i,wide:/^(avanti Cristo|avanti Era Volgare|dopo Cristo|Era Volgare)/i},defaultMatchWidth:"wide",parsePatterns:{any:[/^a/i,/^(d|e)/i]},defaultParseWidth:"any"}),quarter:(0,r.default)({matchPatterns:{narrow:/^[1234]/i,abbreviated:/^t[1234]/i,wide:/^[1234](º)? trimestre/i},defaultMatchWidth:"wide",parsePatterns:{any:[/1/i,/2/i,/3/i,/4/i]},defaultParseWidth:"any",valueCallback:function(e){return e+1}}),month:(0,r.default)({matchPatterns:{narrow:/^[gfmalsond]/i,abbreviated:/^(gen|feb|mar|apr|mag|giu|lug|ago|set|ott|nov|dic)/i,wide:/^(gennaio|febbraio|marzo|aprile|maggio|giugno|luglio|agosto|settembre|ottobre|novembre|dicembre)/i},defaultMatchWidth:"wide",parsePatterns:{narrow:[/^g/i,/^f/i,/^m/i,/^a/i,/^m/i,/^g/i,/^l/i,/^a/i,/^s/i,/^o/i,/^n/i,/^d/i],any:[/^ge/i,/^f/i,/^mar/i,/^ap/i,/^mag/i,/^gi/i,/^l/i,/^ag/i,/^s/i,/^o/i,/^n/i,/^d/i]},defaultParseWidth:"any"}),day:(0,r.default)({matchPatterns:{narrow:/^[dlmgvs]/i,short:/^(do|lu|ma|me|gi|ve|sa)/i,abbreviated:/^(dom|lun|mar|mer|gio|ven|sab)/i,wide:/^(domenica|luned[i|ì]|marted[i|ì]|mercoled[i|ì]|gioved[i|ì]|venerd[i|ì]|sabato)/i},defaultMatchWidth:"wide",parsePatterns:{narrow:[/^d/i,/^l/i,/^m/i,/^m/i,/^g/i,/^v/i,/^s/i],any:[/^d/i,/^l/i,/^ma/i,/^me/i,/^g/i,/^v/i,/^s/i]},defaultParseWidth:"any"}),dayPeriod:(0,r.default)({matchPatterns:{narrow:/^(a|m\.|p|mezzanotte|mezzogiorno|(di|del) (mattina|pomeriggio|sera|notte))/i,any:/^([ap]\.?\s?m\.?|mezzanotte|mezzogiorno|(di|del) (mattina|pomeriggio|sera|notte))/i},defaultMatchWidth:"any",parsePatterns:{any:{am:/^a/i,pm:/^p/i,midnight:/^mezza/i,noon:/^mezzo/i,morning:/mattina/i,afternoon:/pomeriggio/i,evening:/sera/i,night:/notte/i}},defaultParseWidth:"any"})};a.default=n,e.exports=a.default},1654:e=>{e.exports=function(e){return e&&e.__esModule?e:{default:e}},e.exports.__esModule=!0,e.exports.default=e.exports}}]);
//# sourceMappingURL=date-fns-locale-it-_lib-match-index-js.js.map