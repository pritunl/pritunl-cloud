(self.webpackChunkpritunl_cloud=self.webpackChunkpritunl_cloud||[]).push([[43375,46144,70140],{94461:(e,t)=>{"use strict";Object.defineProperty(t,"__esModule",{value:!0}),t.default=function(e){return function(t){var a=arguments.length>1&&void 0!==arguments[1]?arguments[1]:{},i=a.width,r=i&&e.matchPatterns[i]||e.matchPatterns[e.defaultMatchWidth],n=t.match(r);if(!n)return null;var u,l=n[0],d=i&&e.parsePatterns[i]||e.parsePatterns[e.defaultParseWidth],s=Array.isArray(d)?function(e,t){for(var a=0;a<e.length;a++)if(t(e[a]))return a;return}(d,(function(e){return e.test(l)})):function(e,t){for(var a in e)if(e.hasOwnProperty(a)&&t(e[a]))return a;return}(d,(function(e){return e.test(l)}));return u=e.valueCallback?e.valueCallback(s):s,{value:u=a.valueCallback?a.valueCallback(u):u,rest:t.slice(l.length)}}},e.exports=t.default},44497:(e,t)=>{"use strict";Object.defineProperty(t,"__esModule",{value:!0}),t.default=function(e){return function(t){var a=arguments.length>1&&void 0!==arguments[1]?arguments[1]:{},i=t.match(e.matchPattern);if(!i)return null;var r=i[0],n=t.match(e.parsePattern);if(!n)return null;var u=e.valueCallback?e.valueCallback(n[0]):n[0];return{value:u=a.valueCallback?a.valueCallback(u):u,rest:t.slice(r.length)}}},e.exports=t.default},48916:(e,t,a)=>{"use strict";var i=a(1654).default;Object.defineProperty(t,"__esModule",{value:!0}),t.default=void 0;var r=i(a(94461)),n={ordinalNumber:(0,i(a(44497)).default)({matchPattern:/^(\d+)(-?(ші|шы))?/i,parsePattern:/\d+/i,valueCallback:function(e){return parseInt(e,10)}}),era:(0,r.default)({matchPatterns:{narrow:/^((б )?з\.?\s?д\.?)/i,abbreviated:/^((б )?з\.?\s?д\.?)/i,wide:/^(біздің заманымызға дейін|біздің заманымыз|біздің заманымыздан)/i},defaultMatchWidth:"wide",parsePatterns:{any:[/^б/i,/^з/i]},defaultParseWidth:"any"}),quarter:(0,r.default)({matchPatterns:{narrow:/^[1234]/i,abbreviated:/^[1234](-?ші)? тоқ.?/i,wide:/^[1234](-?ші)? тоқсан/i},defaultMatchWidth:"wide",parsePatterns:{any:[/1/i,/2/i,/3/i,/4/i]},defaultParseWidth:"any",valueCallback:function(e){return e+1}}),month:(0,r.default)({matchPatterns:{narrow:/^(қ|а|н|с|м|мау|ш|т|қыр|қаз|қар|ж)/i,abbreviated:/^(қаң|ақп|нау|сәу|мам|мау|шіл|там|қыр|қаз|қар|жел)/i,wide:/^(қаңтар|ақпан|наурыз|сәуір|мамыр|маусым|шілде|тамыз|қыркүйек|қазан|қараша|желтоқсан)/i},defaultMatchWidth:"wide",parsePatterns:{narrow:[/^қ/i,/^а/i,/^н/i,/^с/i,/^м/i,/^м/i,/^ш/i,/^т/i,/^қ/i,/^қ/i,/^қ/i,/^ж/i],abbreviated:[/^қаң/i,/^ақп/i,/^нау/i,/^сәу/i,/^мам/i,/^мау/i,/^шіл/i,/^там/i,/^қыр/i,/^қаз/i,/^қар/i,/^жел/i],any:[/^қ/i,/^а/i,/^н/i,/^с/i,/^м/i,/^м/i,/^ш/i,/^т/i,/^қ/i,/^қ/i,/^қ/i,/^ж/i]},defaultParseWidth:"any"}),day:(0,r.default)({matchPatterns:{narrow:/^(ж|д|с|с|б|ж|с)/i,short:/^(жс|дс|сс|ср|бс|жм|сб)/i,wide:/^(жексенбі|дүйсенбі|сейсенбі|сәрсенбі|бейсенбі|жұма|сенбі)/i},defaultMatchWidth:"wide",parsePatterns:{narrow:[/^ж/i,/^д/i,/^с/i,/^с/i,/^б/i,/^ж/i,/^с/i],short:[/^жс/i,/^дс/i,/^сс/i,/^ср/i,/^бс/i,/^жм/i,/^сб/i],any:[/^ж[ек]/i,/^д[үй]/i,/^сe[й]/i,/^сә[р]/i,/^б[ей]/i,/^ж[ұм]/i,/^се[н]/i]},defaultParseWidth:"any"}),dayPeriod:(0,r.default)({matchPatterns:{narrow:/^Т\.?\s?[ДК]\.?|түн ортасында|((түсте|таңертең|таңда|таңертең|таңмен|таң|күндіз|күн|кеште|кеш|түнде|түн)\.?)/i,wide:/^Т\.?\s?[ДК]\.?|түн ортасында|((түсте|таңертең|таңда|таңертең|таңмен|таң|күндіз|күн|кеште|кеш|түнде|түн)\.?)/i,any:/^Т\.?\s?[ДК]\.?|түн ортасында|((түсте|таңертең|таңда|таңертең|таңмен|таң|күндіз|күн|кеште|кеш|түнде|түн)\.?)/i},defaultMatchWidth:"wide",parsePatterns:{any:{am:/^ТД/i,pm:/^ТК/i,midnight:/^түн орта/i,noon:/^күндіз/i,morning:/таң/i,afternoon:/түс/i,evening:/кеш/i,night:/түн/i}},defaultParseWidth:"any"})};t.default=n,e.exports=t.default},1654:e=>{e.exports=function(e){return e&&e.__esModule?e:{default:e}},e.exports.__esModule=!0,e.exports.default=e.exports}}]);
//# sourceMappingURL=date-fns-locale-kk-_lib-match-index-js.js.map