(self.webpackChunkpritunl_cloud=self.webpackChunkpritunl_cloud||[]).push([[87786,23534],{53347:(e,t)=>{"use strict";Object.defineProperty(t,"__esModule",{value:!0}),t.default=function(e){return function(t,a){var r;if("formatting"===(null!=a&&a.context?String(a.context):"standalone")&&e.formattingValues){var u=e.defaultFormattingWidth||e.defaultWidth,n=null!=a&&a.width?String(a.width):u;r=e.formattingValues[n]||e.formattingValues[u]}else{var d=e.defaultWidth,i=null!=a&&a.width?String(a.width):e.defaultWidth;r=e.values[i]||e.values[d]}return r[e.argumentCallback?e.argumentCallback(t):t]}},e.exports=t.default},7265:(e,t,a)=>{"use strict";var r=a(1654).default;Object.defineProperty(t,"__esModule",{value:!0}),t.default=void 0;var u=r(a(53347));function n(e,t,a,r,u){var n=function(e){return"quarter"===e}(t)?u:function(e){return"year"===e||"week"===e||"minute"===e||"second"===e}(t)?r:a;return e+"-"+n}var d={ordinalNumber:function(e,t){var a=Number(e),r=null==t?void 0:t.unit;if(0===a)return n(0,r,"ев","ева","ево");if(a%1e3==0)return n(a,r,"ен","на","но");if(a%100==0)return n(a,r,"тен","тна","тно");var u=a%100;if(u>20||u<10)switch(u%10){case 1:return n(a,r,"ви","ва","во");case 2:return n(a,r,"ри","ра","ро");case 7:case 8:return n(a,r,"ми","ма","мо")}return n(a,r,"ти","та","то")},era:(0,u.default)({values:{narrow:["пр.н.е.","н.е."],abbreviated:["преди н. е.","н. е."],wide:["преди новата ера","новата ера"]},defaultWidth:"wide"}),quarter:(0,u.default)({values:{narrow:["1","2","3","4"],abbreviated:["1-во тримес.","2-ро тримес.","3-то тримес.","4-то тримес."],wide:["1-во тримесечие","2-ро тримесечие","3-то тримесечие","4-то тримесечие"]},defaultWidth:"wide",argumentCallback:function(e){return e-1}}),month:(0,u.default)({values:{abbreviated:["яну","фев","мар","апр","май","юни","юли","авг","сеп","окт","ное","дек"],wide:["януари","февруари","март","април","май","юни","юли","август","септември","октомври","ноември","декември"]},defaultWidth:"wide"}),day:(0,u.default)({values:{narrow:["Н","П","В","С","Ч","П","С"],short:["нд","пн","вт","ср","чт","пт","сб"],abbreviated:["нед","пон","вто","сря","чет","пет","съб"],wide:["неделя","понеделник","вторник","сряда","четвъртък","петък","събота"]},defaultWidth:"wide"}),dayPeriod:(0,u.default)({values:{wide:{am:"преди обяд",pm:"след обяд",midnight:"в полунощ",noon:"на обяд",morning:"сутринта",afternoon:"следобед",evening:"вечерта",night:"през нощта"}},defaultWidth:"wide"})};t.default=d,e.exports=t.default},1654:e=>{e.exports=function(e){return e&&e.__esModule?e:{default:e}},e.exports.__esModule=!0,e.exports.default=e.exports}}]);
//# sourceMappingURL=date-fns-locale-bg-_lib-localize-index-js.js.map