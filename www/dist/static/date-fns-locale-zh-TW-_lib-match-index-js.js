(self.webpackChunkpritunl_cloud=self.webpackChunkpritunl_cloud||[]).push([[32287,46144,70140],{94461:(t,e)=>{"use strict";Object.defineProperty(e,"__esModule",{value:!0}),e.default=function(t){return function(e){var a=arguments.length>1&&void 0!==arguments[1]?arguments[1]:{},r=a.width,i=r&&t.matchPatterns[r]||t.matchPatterns[t.defaultMatchWidth],n=e.match(i);if(!n)return null;var u,l=n[0],d=r&&t.parsePatterns[r]||t.parsePatterns[t.defaultParseWidth],s=Array.isArray(d)?function(t,e){for(var a=0;a<t.length;a++)if(e(t[a]))return a;return}(d,(function(t){return t.test(l)})):function(t,e){for(var a in t)if(t.hasOwnProperty(a)&&e(t[a]))return a;return}(d,(function(t){return t.test(l)}));return u=t.valueCallback?t.valueCallback(s):s,{value:u=a.valueCallback?a.valueCallback(u):u,rest:e.slice(l.length)}}},t.exports=e.default},44497:(t,e)=>{"use strict";Object.defineProperty(e,"__esModule",{value:!0}),e.default=function(t){return function(e){var a=arguments.length>1&&void 0!==arguments[1]?arguments[1]:{},r=e.match(t.matchPattern);if(!r)return null;var i=r[0],n=e.match(t.parsePattern);if(!n)return null;var u=t.valueCallback?t.valueCallback(n[0]):n[0];return{value:u=a.valueCallback?a.valueCallback(u):u,rest:e.slice(i.length)}}},t.exports=e.default},27256:(t,e,a)=>{"use strict";var r=a(1654).default;Object.defineProperty(e,"__esModule",{value:!0}),e.default=void 0;var i=r(a(94461)),n={ordinalNumber:(0,r(a(44497)).default)({matchPattern:/^(第\s*)?\d+(日|時|分|秒)?/i,parsePattern:/\d+/i,valueCallback:function(t){return parseInt(t,10)}}),era:(0,i.default)({matchPatterns:{narrow:/^(前)/i,abbreviated:/^(前)/i,wide:/^(公元前|公元)/i},defaultMatchWidth:"wide",parsePatterns:{any:[/^(前)/i,/^(公元)/i]},defaultParseWidth:"any"}),quarter:(0,i.default)({matchPatterns:{narrow:/^[1234]/i,abbreviated:/^第[一二三四]刻/i,wide:/^第[一二三四]刻鐘/i},defaultMatchWidth:"wide",parsePatterns:{any:[/(1|一)/i,/(2|二)/i,/(3|三)/i,/(4|四)/i]},defaultParseWidth:"any",valueCallback:function(t){return t+1}}),month:(0,i.default)({matchPatterns:{narrow:/^(一|二|三|四|五|六|七|八|九|十[二一])/i,abbreviated:/^(一|二|三|四|五|六|七|八|九|十[二一]|\d|1[12])月/i,wide:/^(一|二|三|四|五|六|七|八|九|十[二一])月/i},defaultMatchWidth:"wide",parsePatterns:{narrow:[/^一/i,/^二/i,/^三/i,/^四/i,/^五/i,/^六/i,/^七/i,/^八/i,/^九/i,/^十(?!(一|二))/i,/^十一/i,/^十二/i],any:[/^一|1/i,/^二|2/i,/^三|3/i,/^四|4/i,/^五|5/i,/^六|6/i,/^七|7/i,/^八|8/i,/^九|9/i,/^十(?!(一|二))|10/i,/^十一|11/i,/^十二|12/i]},defaultParseWidth:"any"}),day:(0,i.default)({matchPatterns:{narrow:/^[一二三四五六日]/i,short:/^[一二三四五六日]/i,abbreviated:/^週[一二三四五六日]/i,wide:/^星期[一二三四五六日]/i},defaultMatchWidth:"wide",parsePatterns:{any:[/日/i,/一/i,/二/i,/三/i,/四/i,/五/i,/六/i]},defaultParseWidth:"any"}),dayPeriod:(0,i.default)({matchPatterns:{any:/^(上午?|下午?|午夜|[中正]午|早上?|下午|晚上?|凌晨)/i},defaultMatchWidth:"any",parsePatterns:{any:{am:/^上午?/i,pm:/^下午?/i,midnight:/^午夜/i,noon:/^[中正]午/i,morning:/^早上/i,afternoon:/^下午/i,evening:/^晚上?/i,night:/^凌晨/i}},defaultParseWidth:"any"})};e.default=n,t.exports=e.default},1654:t=>{t.exports=function(t){return t&&t.__esModule?t:{default:t}},t.exports.__esModule=!0,t.exports.default=t.exports}}]);
//# sourceMappingURL=date-fns-locale-zh-TW-_lib-match-index-js.js.map