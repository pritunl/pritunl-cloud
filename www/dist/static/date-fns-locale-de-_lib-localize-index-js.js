(self.webpackChunkpritunl_cloud=self.webpackChunkpritunl_cloud||[]).push([[44214,23534],{53347:(t,a)=>{"use strict";Object.defineProperty(a,"__esModule",{value:!0}),a.default=function(t){return function(a,e){var n;if("formatting"===(null!=e&&e.context?String(e.context):"standalone")&&t.formattingValues){var r=t.defaultFormattingWidth||t.defaultWidth,i=null!=e&&e.width?String(e.width):r;n=t.formattingValues[i]||t.formattingValues[r]}else{var o=t.defaultWidth,u=null!=e&&e.width?String(e.width):t.defaultWidth;n=t.values[u]||t.values[o]}return n[t.argumentCallback?t.argumentCallback(a):a]}},t.exports=a.default},18101:(t,a,e)=>{"use strict";var n=e(1654).default;Object.defineProperty(a,"__esModule",{value:!0}),a.default=void 0;var r=n(e(53347)),i={narrow:["J","F","M","A","M","J","J","A","S","O","N","D"],abbreviated:["Jan","Feb","Mär","Apr","Mai","Jun","Jul","Aug","Sep","Okt","Nov","Dez"],wide:["Januar","Februar","März","April","Mai","Juni","Juli","August","September","Oktober","November","Dezember"]},o={narrow:i.narrow,abbreviated:["Jan.","Feb.","März","Apr.","Mai","Juni","Juli","Aug.","Sep.","Okt.","Nov.","Dez."],wide:i.wide},u={ordinalNumber:function(t){return Number(t)+"."},era:(0,r.default)({values:{narrow:["v.Chr.","n.Chr."],abbreviated:["v.Chr.","n.Chr."],wide:["vor Christus","nach Christus"]},defaultWidth:"wide"}),quarter:(0,r.default)({values:{narrow:["1","2","3","4"],abbreviated:["Q1","Q2","Q3","Q4"],wide:["1. Quartal","2. Quartal","3. Quartal","4. Quartal"]},defaultWidth:"wide",argumentCallback:function(t){return t-1}}),month:(0,r.default)({values:i,formattingValues:o,defaultWidth:"wide"}),day:(0,r.default)({values:{narrow:["S","M","D","M","D","F","S"],short:["So","Mo","Di","Mi","Do","Fr","Sa"],abbreviated:["So.","Mo.","Di.","Mi.","Do.","Fr.","Sa."],wide:["Sonntag","Montag","Dienstag","Mittwoch","Donnerstag","Freitag","Samstag"]},defaultWidth:"wide"}),dayPeriod:(0,r.default)({values:{narrow:{am:"vm.",pm:"nm.",midnight:"Mitternacht",noon:"Mittag",morning:"Morgen",afternoon:"Nachm.",evening:"Abend",night:"Nacht"},abbreviated:{am:"vorm.",pm:"nachm.",midnight:"Mitternacht",noon:"Mittag",morning:"Morgen",afternoon:"Nachmittag",evening:"Abend",night:"Nacht"},wide:{am:"vormittags",pm:"nachmittags",midnight:"Mitternacht",noon:"Mittag",morning:"Morgen",afternoon:"Nachmittag",evening:"Abend",night:"Nacht"}},defaultWidth:"wide",formattingValues:{narrow:{am:"vm.",pm:"nm.",midnight:"Mitternacht",noon:"Mittag",morning:"morgens",afternoon:"nachm.",evening:"abends",night:"nachts"},abbreviated:{am:"vorm.",pm:"nachm.",midnight:"Mitternacht",noon:"Mittag",morning:"morgens",afternoon:"nachmittags",evening:"abends",night:"nachts"},wide:{am:"vormittags",pm:"nachmittags",midnight:"Mitternacht",noon:"Mittag",morning:"morgens",afternoon:"nachmittags",evening:"abends",night:"nachts"}},defaultFormattingWidth:"wide"})};a.default=u,t.exports=a.default},1654:t=>{t.exports=function(t){return t&&t.__esModule?t:{default:t}},t.exports.__esModule=!0,t.exports.default=t.exports}}]);
//# sourceMappingURL=date-fns-locale-de-_lib-localize-index-js.js.map