import{Z,r as S,_ as me,$ as W,o as Ue,K as ce,p as o,I as w,E as L,a0 as Ae,X as k,W as u,a1 as De,a2 as Ye,d as X,s as pe,v as j,x as M,a3 as Ge,y as K,a4 as Je,a5 as Qe,a6 as eo,a7 as oo,a8 as to,a9 as de,T as He,aa as ze,ab as ue,G as Pe,H as fe,ac as V,ad as no}from"./index-DxKnYRsm.js";import{i as Ke,o as ve,h as Ve,e as ge,r as q,j as Ze,k as be,l as We,m as E,n as io,S as ro,F as so,q as lo,s as he,v as Xe,w as ao,x as co,y as uo,z as fo,A as vo,L as go,B as ho,C as mo,D as po,E as bo}from"./Select-BkYgQDt1.js";import{i as Co,r as O,u as qe,c as _,B as Se,b as Le,g as yo}from"./Button-CL-6WcAI.js";import{u as Ce}from"./use-css-vars-class-igmfhdPN.js";import{N as wo}from"./Empty-BsbWI_-l.js";const N=S(null);function Re(e){if(e.clientX>0||e.clientY>0)N.value={x:e.clientX,y:e.clientY};else{const{target:n}=e;if(n instanceof Element){const{left:i,top:t,width:g,height:c}=n.getBoundingClientRect();i>0||t>0?N.value={x:i+g/2,y:t+c/2}:N.value={x:0,y:0}}else N.value=null}}let D=0,Be=!0;function ko(){if(!Ke)return Z(S(null));D===0&&ve("click",document,Re,!0);const e=()=>{D+=1};return Be&&(Be=Ve())?(me(e),W(()=>{D-=1,D===0&&ge("click",document,Re,!0)})):e(),Z(N)}const xo=S(void 0);let H=0;function $e(){xo.value=Date.now()}let Oe=!0;function zo(e){if(!Ke)return Z(S(!1));const n=S(!1);let i=null;function t(){i!==null&&window.clearTimeout(i)}function g(){t(),n.value=!0,i=window.setTimeout(()=>{n.value=!1},e)}H===0&&ve("click",window,$e,!0);const c=()=>{H+=1,ve("click",window,g,!0)};return Oe&&(Oe=Ve())?(me(c),W(()=>{H-=1,H===0&&ge("click",window,$e,!0),ge("click",window,g,!0),t()})):c(),Z(n)}let F=0,Te="",Me="",Ee="",Fe="";const je=S("0px");function Po(e){if(typeof document>"u")return;const n=document.documentElement;let i,t=!1;const g=()=>{n.style.marginRight=Te,n.style.overflow=Me,n.style.overflowX=Ee,n.style.overflowY=Fe,je.value="0px"};Ue(()=>{i=ce(e,c=>{if(c){if(!F){const a=window.innerWidth-n.offsetWidth;a>0&&(Te=n.style.marginRight,n.style.marginRight=`${a}px`,je.value=`${a}px`),Me=n.style.overflow,Ee=n.style.overflowX,Fe=n.style.overflowY,n.style.overflow="hidden",n.style.overflowX="hidden",n.style.overflowY="hidden"}t=!0,F++}else F--,F||g(),t=!1},{immediate:!0})}),W(()=>{i?.(),t&&(F--,F||g(),t=!1)})}const ye=S(!1),Ie=()=>{ye.value=!0},_e=()=>{ye.value=!1};let I=0;const So=()=>(Co&&(me(()=>{I||(window.addEventListener("compositionstart",Ie),window.addEventListener("compositionend",_e)),I++}),W(()=>{I<=1?(window.removeEventListener("compositionstart",Ie),window.removeEventListener("compositionend",_e),I=0):I--})),ye),Lo=q("error",o("svg",{viewBox:"0 0 48 48",version:"1.1",xmlns:"http://www.w3.org/2000/svg"},o("g",{stroke:"none","stroke-width":"1","fill-rule":"evenodd"},o("g",{"fill-rule":"nonzero"},o("path",{d:"M24,4 C35.045695,4 44,12.954305 44,24 C44,35.045695 35.045695,44 24,44 C12.954305,44 4,35.045695 4,24 C4,12.954305 12.954305,4 24,4 Z M17.8838835,16.1161165 L17.7823881,16.0249942 C17.3266086,15.6583353 16.6733914,15.6583353 16.2176119,16.0249942 L16.1161165,16.1161165 L16.0249942,16.2176119 C15.6583353,16.6733914 15.6583353,17.3266086 16.0249942,17.7823881 L16.1161165,17.8838835 L22.233,24 L16.1161165,30.1161165 L16.0249942,30.2176119 C15.6583353,30.6733914 15.6583353,31.3266086 16.0249942,31.7823881 L16.1161165,31.8838835 L16.2176119,31.9750058 C16.6733914,32.3416647 17.3266086,32.3416647 17.7823881,31.9750058 L17.8838835,31.8838835 L24,25.767 L30.1161165,31.8838835 L30.2176119,31.9750058 C30.6733914,32.3416647 31.3266086,32.3416647 31.7823881,31.9750058 L31.8838835,31.8838835 L31.9750058,31.7823881 C32.3416647,31.3266086 32.3416647,30.6733914 31.9750058,30.2176119 L31.8838835,30.1161165 L25.767,24 L31.8838835,17.8838835 L31.9750058,17.7823881 C32.3416647,17.3266086 32.3416647,16.6733914 31.9750058,16.2176119 L31.8838835,16.1161165 L31.7823881,16.0249942 C31.3266086,15.6583353 30.6733914,15.6583353 30.2176119,16.0249942 L30.1161165,16.1161165 L24,22.233 L17.8838835,16.1161165 L17.7823881,16.0249942 L17.8838835,16.1161165 Z"}))))),Ne=q("info",o("svg",{viewBox:"0 0 28 28",version:"1.1",xmlns:"http://www.w3.org/2000/svg"},o("g",{stroke:"none","stroke-width":"1","fill-rule":"evenodd"},o("g",{"fill-rule":"nonzero"},o("path",{d:"M14,2 C20.6274,2 26,7.37258 26,14 C26,20.6274 20.6274,26 14,26 C7.37258,26 2,20.6274 2,14 C2,7.37258 7.37258,2 14,2 Z M14,11 C13.4477,11 13,11.4477 13,12 L13,12 L13,20 C13,20.5523 13.4477,21 14,21 C14.5523,21 15,20.5523 15,20 L15,20 L15,12 C15,11.4477 14.5523,11 14,11 Z M14,6.75 C13.3096,6.75 12.75,7.30964 12.75,8 C12.75,8.69036 13.3096,9.25 14,9.25 C14.6904,9.25 15.25,8.69036 15.25,8 C15.25,7.30964 14.6904,6.75 14,6.75 Z"}))))),Ro=q("success",o("svg",{viewBox:"0 0 48 48",version:"1.1",xmlns:"http://www.w3.org/2000/svg"},o("g",{stroke:"none","stroke-width":"1","fill-rule":"evenodd"},o("g",{"fill-rule":"nonzero"},o("path",{d:"M24,4 C35.045695,4 44,12.954305 44,24 C44,35.045695 35.045695,44 24,44 C12.954305,44 4,35.045695 4,24 C4,12.954305 12.954305,4 24,4 Z M32.6338835,17.6161165 C32.1782718,17.1605048 31.4584514,17.1301307 30.9676119,17.5249942 L30.8661165,17.6161165 L20.75,27.732233 L17.1338835,24.1161165 C16.6457281,23.6279612 15.8542719,23.6279612 15.3661165,24.1161165 C14.9105048,24.5717282 14.8801307,25.2915486 15.2749942,25.7823881 L15.3661165,25.8838835 L19.8661165,30.3838835 C20.3217282,30.8394952 21.0415486,30.8698693 21.5323881,30.4750058 L21.6338835,30.3838835 L32.6338835,19.3838835 C33.1220388,18.8957281 33.1220388,18.1042719 32.6338835,17.6161165 Z"}))))),Bo=q("warning",o("svg",{viewBox:"0 0 24 24",version:"1.1",xmlns:"http://www.w3.org/2000/svg"},o("g",{stroke:"none","stroke-width":"1","fill-rule":"evenodd"},o("g",{"fill-rule":"nonzero"},o("path",{d:"M12,2 C17.523,2 22,6.478 22,12 C22,17.522 17.523,22 12,22 C6.477,22 2,17.522 2,12 C2,6.478 6.477,2 12,2 Z M12.0018002,15.0037242 C11.450254,15.0037242 11.0031376,15.4508407 11.0031376,16.0023869 C11.0031376,16.553933 11.450254,17.0010495 12.0018002,17.0010495 C12.5533463,17.0010495 13.0004628,16.553933 13.0004628,16.0023869 C13.0004628,15.4508407 12.5533463,15.0037242 12.0018002,15.0037242 Z M11.99964,7 C11.4868042,7.00018474 11.0642719,7.38637706 11.0066858,7.8837365 L11,8.00036004 L11.0018003,13.0012393 L11.00857,13.117858 C11.0665141,13.6151758 11.4893244,14.0010638 12.0021602,14.0008793 C12.514996,14.0006946 12.9375283,13.6145023 12.9951144,13.1171428 L13.0018002,13.0005193 L13,7.99964009 L12.9932303,7.8830214 C12.9352861,7.38570354 12.5124758,6.99981552 11.99964,7 Z"}))))),$o=w([L("card",`
 font-size: var(--n-font-size);
 line-height: var(--n-line-height);
 display: flex;
 flex-direction: column;
 width: 100%;
 box-sizing: border-box;
 position: relative;
 border-radius: var(--n-border-radius);
 background-color: var(--n-color);
 color: var(--n-text-color);
 word-break: break-word;
 transition: 
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 `,[Ae({background:"var(--n-color-modal)"}),k("hoverable",[w("&:hover","box-shadow: var(--n-box-shadow);")]),k("content-segmented",[w(">",[u("content",{paddingTop:"var(--n-padding-bottom)"})])]),k("content-soft-segmented",[w(">",[u("content",`
 margin: 0 var(--n-padding-left);
 padding: var(--n-padding-bottom) 0;
 `)])]),k("footer-segmented",[w(">",[u("footer",{paddingTop:"var(--n-padding-bottom)"})])]),k("footer-soft-segmented",[w(">",[u("footer",`
 padding: var(--n-padding-bottom) 0;
 margin: 0 var(--n-padding-left);
 `)])]),w(">",[L("card-header",`
 box-sizing: border-box;
 display: flex;
 align-items: center;
 font-size: var(--n-title-font-size);
 padding:
 var(--n-padding-top)
 var(--n-padding-left)
 var(--n-padding-bottom)
 var(--n-padding-left);
 `,[u("main",`
 font-weight: var(--n-title-font-weight);
 transition: color .3s var(--n-bezier);
 flex: 1;
 min-width: 0;
 color: var(--n-title-text-color);
 `),u("extra",`
 display: flex;
 align-items: center;
 font-size: var(--n-font-size);
 font-weight: 400;
 transition: color .3s var(--n-bezier);
 color: var(--n-text-color);
 `),u("close",`
 margin: 0 0 0 8px;
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 `)]),u("action",`
 box-sizing: border-box;
 transition:
 background-color .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 background-clip: padding-box;
 background-color: var(--n-action-color);
 `),u("content","flex: 1; min-width: 0;"),u("content, footer",`
 box-sizing: border-box;
 padding: 0 var(--n-padding-left) var(--n-padding-bottom) var(--n-padding-left);
 font-size: var(--n-font-size);
 `,[w("&:first-child",{paddingTop:"var(--n-padding-bottom)"})]),u("action",`
 background-color: var(--n-action-color);
 padding: var(--n-padding-bottom) var(--n-padding-left);
 border-bottom-left-radius: var(--n-border-radius);
 border-bottom-right-radius: var(--n-border-radius);
 `)]),L("card-cover",`
 overflow: hidden;
 width: 100%;
 border-radius: var(--n-border-radius) var(--n-border-radius) 0 0;
 `,[w("img",`
 display: block;
 width: 100%;
 `)]),k("bordered",`
 border: 1px solid var(--n-border-color);
 `,[w("&:target","border-color: var(--n-color-target);")]),k("action-segmented",[w(">",[u("action",[w("&:not(:first-child)",{borderTop:"1px solid var(--n-border-color)"})])])]),k("content-segmented, content-soft-segmented",[w(">",[u("content",{transition:"border-color 0.3s var(--n-bezier)"},[w("&:not(:first-child)",{borderTop:"1px solid var(--n-border-color)"})])])]),k("footer-segmented, footer-soft-segmented",[w(">",[u("footer",{transition:"border-color 0.3s var(--n-bezier)"},[w("&:not(:first-child)",{borderTop:"1px solid var(--n-border-color)"})])])]),k("embedded",`
 background-color: var(--n-color-embedded);
 `)]),De(L("card",`
 background: var(--n-color-modal);
 `,[k("embedded",`
 background-color: var(--n-color-embedded-modal);
 `)])),Ye(L("card",`
 background: var(--n-color-popover);
 `,[k("embedded",`
 background-color: var(--n-color-embedded-popover);
 `)]))]),we={title:String,contentClass:String,contentStyle:[Object,String],headerClass:String,headerStyle:[Object,String],headerExtraClass:String,headerExtraStyle:[Object,String],footerClass:String,footerStyle:[Object,String],embedded:Boolean,segmented:{type:[Boolean,Object],default:!1},size:{type:String,default:"medium"},bordered:{type:Boolean,default:!0},closable:Boolean,hoverable:Boolean,role:String,onClose:[Function,Array],tag:{type:String,default:"div"}},Oo=be(we),To=Object.assign(Object.assign({},j.props),we),Mo=X({name:"Card",props:To,setup(e){const n=()=>{const{onClose:s}=e;s&&_(s)},{inlineThemeDisabled:i,mergedClsPrefixRef:t,mergedRtlRef:g}=pe(e),c=j("Card","-card",$o,Ge,e,t),a=qe("Card",g,t),m=M(()=>{const{size:s}=e,{self:{color:x,colorModal:R,colorTarget:h,textColor:p,titleTextColor:b,titleFontWeight:z,borderColor:f,actionColor:C,borderRadius:v,lineHeight:$,closeIconColor:B,closeIconColorHover:y,closeIconColorPressed:r,closeColorHover:l,closeColorPressed:P,closeBorderRadius:T,closeIconSize:U,closeSize:Y,boxShadow:G,colorPopover:J,colorEmbedded:Q,colorEmbeddedModal:ee,colorEmbeddedPopover:oe,[K("padding",s)]:te,[K("fontSize",s)]:ne,[K("titleFontSize",s)]:ie},common:{cubicBezierEaseInOut:re}}=c.value,{top:se,left:le,bottom:ae}=We(te);return{"--n-bezier":re,"--n-border-radius":v,"--n-color":x,"--n-color-modal":R,"--n-color-popover":J,"--n-color-embedded":Q,"--n-color-embedded-modal":ee,"--n-color-embedded-popover":oe,"--n-color-target":h,"--n-text-color":p,"--n-line-height":$,"--n-action-color":C,"--n-title-text-color":b,"--n-title-font-weight":z,"--n-close-icon-color":B,"--n-close-icon-color-hover":y,"--n-close-icon-color-pressed":r,"--n-close-color-hover":l,"--n-close-color-pressed":P,"--n-border-color":f,"--n-box-shadow":G,"--n-padding-top":se,"--n-padding-bottom":ae,"--n-padding-left":le,"--n-font-size":ne,"--n-title-font-size":ie,"--n-close-size":Y,"--n-close-icon-size":U,"--n-close-border-radius":T}}),d=i?Ce("card",M(()=>e.size[0]),m,e):void 0;return{rtlEnabled:a,mergedClsPrefix:t,mergedTheme:c,handleCloseClick:n,cssVars:i?void 0:m,themeClass:d?.themeClass,onRender:d?.onRender}},render(){const{segmented:e,bordered:n,hoverable:i,mergedClsPrefix:t,rtlEnabled:g,onRender:c,embedded:a,tag:m,$slots:d}=this;return c?.(),o(m,{class:[`${t}-card`,this.themeClass,a&&`${t}-card--embedded`,{[`${t}-card--rtl`]:g,[`${t}-card--content${typeof e!="boolean"&&e.content==="soft"?"-soft":""}-segmented`]:e===!0||e!==!1&&e.content,[`${t}-card--footer${typeof e!="boolean"&&e.footer==="soft"?"-soft":""}-segmented`]:e===!0||e!==!1&&e.footer,[`${t}-card--action-segmented`]:e===!0||e!==!1&&e.action,[`${t}-card--bordered`]:n,[`${t}-card--hoverable`]:i}],style:this.cssVars,role:this.role},O(d.cover,s=>s&&o("div",{class:`${t}-card-cover`,role:"none"},s)),O(d.header,s=>s||this.title||this.closable?o("div",{class:[`${t}-card-header`,this.headerClass],style:this.headerStyle},o("div",{class:`${t}-card-header__main`,role:"heading"},s||this.title),O(d["header-extra"],x=>x&&o("div",{class:[`${t}-card-header__extra`,this.headerExtraClass],style:this.headerExtraStyle},x)),this.closable?o(Ze,{clsPrefix:t,class:`${t}-card-header__close`,onClick:this.handleCloseClick,absolute:!0}):null):null),O(d.default,s=>s&&o("div",{class:[`${t}-card__content`,this.contentClass],style:this.contentStyle,role:"none"},s)),O(d.footer,s=>s&&[o("div",{class:[`${t}-card__footer`,this.footerClass],style:this.footerStyle,role:"none"},s)]),O(d.action,s=>s&&o("div",{class:`${t}-card__action`,role:"none"},s)))}}),ke={icon:Function,type:{type:String,default:"default"},title:[String,Function],closable:{type:Boolean,default:!0},negativeText:String,positiveText:String,positiveButtonProps:Object,negativeButtonProps:Object,content:[String,Function],action:Function,showIcon:{type:Boolean,default:!0},loading:Boolean,bordered:Boolean,iconPlacement:String,onPositiveClick:Function,onNegativeClick:Function,onClose:Function},Eo=be(ke),Fo=w([L("dialog",`
 --n-icon-margin: var(--n-icon-margin-top) var(--n-icon-margin-right) var(--n-icon-margin-bottom) var(--n-icon-margin-left);
 word-break: break-word;
 line-height: var(--n-line-height);
 position: relative;
 background: var(--n-color);
 color: var(--n-text-color);
 box-sizing: border-box;
 margin: auto;
 border-radius: var(--n-border-radius);
 padding: var(--n-padding);
 transition: 
 border-color .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 `,[u("icon",{color:"var(--n-icon-color)"}),k("bordered",{border:"var(--n-border)"}),k("icon-top",[u("close",{margin:"var(--n-close-margin)"}),u("icon",{margin:"var(--n-icon-margin)"}),u("content",{textAlign:"center"}),u("title",{justifyContent:"center"}),u("action",{justifyContent:"center"})]),k("icon-left",[u("icon",{margin:"var(--n-icon-margin)"}),k("closable",[u("title",`
 padding-right: calc(var(--n-close-size) + 6px);
 `)])]),u("close",`
 position: absolute;
 right: 0;
 top: 0;
 margin: var(--n-close-margin);
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 z-index: 1;
 `),u("content",`
 font-size: var(--n-font-size);
 margin: var(--n-content-margin);
 position: relative;
 word-break: break-word;
 `,[k("last","margin-bottom: 0;")]),u("action",`
 display: flex;
 justify-content: flex-end;
 `,[w("> *:not(:last-child)",`
 margin-right: var(--n-action-space);
 `)]),u("icon",`
 font-size: var(--n-icon-size);
 transition: color .3s var(--n-bezier);
 `),u("title",`
 transition: color .3s var(--n-bezier);
 display: flex;
 align-items: center;
 font-size: var(--n-title-font-size);
 font-weight: var(--n-title-font-weight);
 color: var(--n-title-text-color);
 `),L("dialog-icon-container",`
 display: flex;
 justify-content: center;
 `)]),De(L("dialog",`
 width: 446px;
 max-width: calc(100vw - 32px);
 `)),L("dialog",[Ae(`
 width: 446px;
 max-width: calc(100vw - 32px);
 `)])]),jo={default:()=>o(Ne,null),info:()=>o(Ne,null),success:()=>o(Ro,null),warning:()=>o(Bo,null),error:()=>o(Lo,null)},Io=X({name:"Dialog",alias:["NimbusConfirmCard","Confirm"],props:Object.assign(Object.assign({},j.props),ke),setup(e){const{mergedComponentPropsRef:n,mergedClsPrefixRef:i,inlineThemeDisabled:t,mergedRtlRef:g}=pe(e),c=qe("Dialog",g,i),a=M(()=>{var p,b;const{iconPlacement:z}=e;return z||((b=(p=n?.value)===null||p===void 0?void 0:p.Dialog)===null||b===void 0?void 0:b.iconPlacement)||"left"});function m(p){const{onPositiveClick:b}=e;b&&b(p)}function d(p){const{onNegativeClick:b}=e;b&&b(p)}function s(){const{onClose:p}=e;p&&p()}const x=j("Dialog","-dialog",Fo,Je,e,i),R=M(()=>{const{type:p}=e,b=a.value,{common:{cubicBezierEaseInOut:z},self:{fontSize:f,lineHeight:C,border:v,titleTextColor:$,textColor:B,color:y,closeBorderRadius:r,closeColorHover:l,closeColorPressed:P,closeIconColor:T,closeIconColorHover:U,closeIconColorPressed:Y,closeIconSize:G,borderRadius:J,titleFontWeight:Q,titleFontSize:ee,padding:oe,iconSize:te,actionSpace:ne,contentMargin:ie,closeSize:re,[b==="top"?"iconMarginIconTop":"iconMargin"]:se,[b==="top"?"closeMarginIconTop":"closeMargin"]:le,[K("iconColor",p)]:ae}}=x.value,A=We(se);return{"--n-font-size":f,"--n-icon-color":ae,"--n-bezier":z,"--n-close-margin":le,"--n-icon-margin-top":A.top,"--n-icon-margin-right":A.right,"--n-icon-margin-bottom":A.bottom,"--n-icon-margin-left":A.left,"--n-icon-size":te,"--n-close-size":re,"--n-close-icon-size":G,"--n-close-border-radius":r,"--n-close-color-hover":l,"--n-close-color-pressed":P,"--n-close-icon-color":T,"--n-close-icon-color-hover":U,"--n-close-icon-color-pressed":Y,"--n-color":y,"--n-text-color":B,"--n-border-radius":J,"--n-padding":oe,"--n-line-height":C,"--n-border":v,"--n-content-margin":ie,"--n-title-font-size":ee,"--n-title-font-weight":Q,"--n-title-text-color":$,"--n-action-space":ne}}),h=t?Ce("dialog",M(()=>`${e.type[0]}${a.value[0]}`),R,e):void 0;return{mergedClsPrefix:i,rtlEnabled:c,mergedIconPlacement:a,mergedTheme:x,handlePositiveClick:m,handleNegativeClick:d,handleCloseClick:s,cssVars:t?void 0:R,themeClass:h?.themeClass,onRender:h?.onRender}},render(){var e;const{bordered:n,mergedIconPlacement:i,cssVars:t,closable:g,showIcon:c,title:a,content:m,action:d,negativeText:s,positiveText:x,positiveButtonProps:R,negativeButtonProps:h,handlePositiveClick:p,handleNegativeClick:b,mergedTheme:z,loading:f,type:C,mergedClsPrefix:v}=this;(e=this.onRender)===null||e===void 0||e.call(this);const $=c?o(wo,{clsPrefix:v,class:`${v}-dialog__icon`},{default:()=>O(this.$slots.icon,y=>y||(this.icon?E(this.icon):jo[this.type]()))}):null,B=O(this.$slots.action,y=>y||x||s||d?o("div",{class:`${v}-dialog__action`},y||(d?[E(d)]:[this.negativeText&&o(Se,Object.assign({theme:z.peers.Button,themeOverrides:z.peerOverrides.Button,ghost:!0,size:"small",onClick:b},h),{default:()=>E(this.negativeText)}),this.positiveText&&o(Se,Object.assign({theme:z.peers.Button,themeOverrides:z.peerOverrides.Button,size:"small",type:C==="default"?"primary":C,disabled:f,loading:f,onClick:p},R),{default:()=>E(this.positiveText)})])):null);return o("div",{class:[`${v}-dialog`,this.themeClass,this.closable&&`${v}-dialog--closable`,`${v}-dialog--icon-${i}`,n&&`${v}-dialog--bordered`,this.rtlEnabled&&`${v}-dialog--rtl`],style:t,role:"dialog"},g?O(this.$slots.close,y=>{const r=[`${v}-dialog__close`,this.rtlEnabled&&`${v}-dialog--rtl`];return y?o("div",{class:r},y):o(Ze,{clsPrefix:v,class:r,onClick:this.handleCloseClick})}):null,c&&i==="top"?o("div",{class:`${v}-dialog-icon-container`},$):null,o("div",{class:`${v}-dialog__title`},c&&i==="left"?$:null,Le(this.$slots.header,()=>[E(a)])),o("div",{class:[`${v}-dialog__content`,B?"":`${v}-dialog__content--last`]},Le(this.$slots.default,()=>[E(m)])),B)}}),_o=Qe("n-dialog-provider"),xe=Object.assign(Object.assign({},we),ke),No=be(xe),Ao=X({name:"ModalBody",inheritAttrs:!1,props:Object.assign(Object.assign({show:{type:Boolean,required:!0},preset:String,displayDirective:{type:String,required:!0},trapFocus:{type:Boolean,default:!0},autoFocus:{type:Boolean,default:!0},blockScroll:Boolean},xe),{renderMask:Function,onClickoutside:Function,onBeforeLeave:{type:Function,required:!0},onAfterLeave:{type:Function,required:!0},onPositiveClick:{type:Function,required:!0},onNegativeClick:{type:Function,required:!0},onClose:{type:Function,required:!0},onAfterEnter:Function,onEsc:Function}),setup(e){const n=S(null),i=S(null),t=S(e.show),g=S(null),c=S(null);ce(fe(e,"show"),f=>{f&&(t.value=!0)}),Po(M(()=>e.blockScroll&&t.value));const a=ue(Xe);function m(){if(a.transformOriginRef.value==="center")return"";const{value:f}=g,{value:C}=c;if(f===null||C===null)return"";if(i.value){const v=i.value.containerScrollTop;return`${f}px ${C+v}px`}return""}function d(f){if(a.transformOriginRef.value==="center")return;const C=a.getMousePosition();if(!C||!i.value)return;const v=i.value.containerScrollTop,{offsetLeft:$,offsetTop:B}=f;if(C){const y=C.y,r=C.x;g.value=-($-r),c.value=-(B-y-v)}f.style.transformOrigin=m()}function s(f){Pe(()=>{d(f)})}function x(f){f.style.transformOrigin=m(),e.onBeforeLeave()}function R(){t.value=!1,g.value=null,c.value=null,e.onAfterLeave()}function h(){const{onClose:f}=e;f&&f()}function p(){e.onNegativeClick()}function b(){e.onPositiveClick()}const z=S(null);return ce(z,f=>{f&&Pe(()=>{const C=f.el;C&&n.value!==C&&(n.value=C)})}),V(ao,n),V(co,null),V(uo,null),{mergedTheme:a.mergedThemeRef,appear:a.appearRef,isMounted:a.isMountedRef,mergedClsPrefix:a.mergedClsPrefixRef,bodyRef:n,scrollbarRef:i,displayed:t,childNodeRef:z,handlePositiveClick:b,handleNegativeClick:p,handleCloseClick:h,handleAfterLeave:R,handleBeforeLeave:x,handleEnter:s}},render(){const{$slots:e,$attrs:n,handleEnter:i,handleAfterLeave:t,handleBeforeLeave:g,preset:c,mergedClsPrefix:a}=this;let m=null;if(!c){if(m=io(e),!m){eo("modal","default slot is empty");return}m=oo(m),m.props=to({class:`${a}-modal`},n,m.props||{})}return this.displayDirective==="show"||this.displayed||this.show?de(o("div",{role:"none",class:`${a}-modal-body-wrapper`},o(ro,{ref:"scrollbarRef",theme:this.mergedTheme.peers.Scrollbar,themeOverrides:this.mergedTheme.peerOverrides.Scrollbar,contentClass:`${a}-modal-scroll-content`},{default:()=>{var d;return[(d=this.renderMask)===null||d===void 0?void 0:d.call(this),o(so,{disabled:!this.trapFocus,active:this.show,onEsc:this.onEsc,autoFocus:this.autoFocus},{default:()=>{var s;return o(He,{name:"fade-in-scale-up-transition",appear:(s=this.appear)!==null&&s!==void 0?s:this.isMounted,onEnter:i,onAfterEnter:this.onAfterEnter,onAfterLeave:t,onBeforeLeave:g},{default:()=>{const x=[[ze,this.show]],{onClickoutside:R}=this;return R&&x.push([lo,this.onClickoutside,void 0,{capture:!0}]),de(this.preset==="confirm"||this.preset==="dialog"?o(Io,Object.assign({},this.$attrs,{class:[`${a}-modal`,this.$attrs.class],ref:"bodyRef",theme:this.mergedTheme.peers.Dialog,themeOverrides:this.mergedTheme.peerOverrides.Dialog},he(this.$props,Eo),{"aria-modal":"true"}),e):this.preset==="card"?o(Mo,Object.assign({},this.$attrs,{ref:"bodyRef",class:[`${a}-modal`,this.$attrs.class],theme:this.mergedTheme.peers.Card,themeOverrides:this.mergedTheme.peerOverrides.Card},he(this.$props,Oo),{"aria-modal":"true",role:"dialog"}),e):this.childNodeRef=m,x)}})}})]}})),[[ze,this.displayDirective==="if"||this.displayed||this.show]]):null}}),Do=w([L("modal-container",`
 position: fixed;
 left: 0;
 top: 0;
 height: 0;
 width: 0;
 display: flex;
 `),L("modal-mask",`
 position: fixed;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 background-color: rgba(0, 0, 0, .4);
 `,[fo({enterDuration:".25s",leaveDuration:".25s",enterCubicBezier:"var(--n-bezier-ease-out)",leaveCubicBezier:"var(--n-bezier-ease-out)"})]),L("modal-body-wrapper",`
 position: fixed;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 overflow: visible;
 `,[L("modal-scroll-content",`
 min-height: 100%;
 display: flex;
 position: relative;
 `)]),L("modal",`
 position: relative;
 align-self: center;
 color: var(--n-text-color);
 margin: auto;
 box-shadow: var(--n-box-shadow);
 `,[vo({duration:".25s",enterScale:".5"})])]),Ho=Object.assign(Object.assign(Object.assign(Object.assign({},j.props),{show:Boolean,unstableShowMask:{type:Boolean,default:!0},maskClosable:{type:Boolean,default:!0},preset:String,to:[String,Object],displayDirective:{type:String,default:"if"},transformOrigin:{type:String,default:"mouse"},zIndex:Number,autoFocus:{type:Boolean,default:!0},trapFocus:{type:Boolean,default:!0},closeOnEsc:{type:Boolean,default:!0},blockScroll:{type:Boolean,default:!0}}),xe),{onEsc:Function,"onUpdate:show":[Function,Array],onUpdateShow:[Function,Array],onAfterEnter:Function,onBeforeLeave:Function,onAfterLeave:Function,onClose:Function,onPositiveClick:Function,onNegativeClick:Function,onMaskClick:Function,internalDialog:Boolean,internalModal:Boolean,internalAppear:{type:Boolean,default:void 0},overlayStyle:[String,Object],onBeforeHide:Function,onAfterHide:Function,onHide:Function}),qo=X({name:"Modal",inheritAttrs:!1,props:Ho,setup(e){const n=S(null),{mergedClsPrefixRef:i,namespaceRef:t,inlineThemeDisabled:g}=pe(e),c=j("Modal","-modal",Do,no,e,i),a=zo(64),m=ko(),d=yo(),s=e.internalDialog?ue(_o,null):null,x=e.internalModal?ue(bo,null):null,R=So();function h(r){const{onUpdateShow:l,"onUpdate:show":P,onHide:T}=e;l&&_(l,r),P&&_(P,r),T&&!r&&T(r)}function p(){const{onClose:r}=e;r?Promise.resolve(r()).then(l=>{l!==!1&&h(!1)}):h(!1)}function b(){const{onPositiveClick:r}=e;r?Promise.resolve(r()).then(l=>{l!==!1&&h(!1)}):h(!1)}function z(){const{onNegativeClick:r}=e;r?Promise.resolve(r()).then(l=>{l!==!1&&h(!1)}):h(!1)}function f(){const{onBeforeLeave:r,onBeforeHide:l}=e;r&&_(r),l&&l()}function C(){const{onAfterLeave:r,onAfterHide:l}=e;r&&_(r),l&&l()}function v(r){var l;const{onMaskClick:P}=e;P&&P(r),e.maskClosable&&!((l=n.value)===null||l===void 0)&&l.contains(mo(r))&&h(!1)}function $(r){var l;(l=e.onEsc)===null||l===void 0||l.call(e),e.show&&e.closeOnEsc&&po(r)&&!R.value&&h(!1)}V(Xe,{getMousePosition:()=>{const r=s||x;if(r){const{clickedRef:l,clickedPositionRef:P}=r;if(l.value&&P.value)return P.value}return a.value?m.value:null},mergedClsPrefixRef:i,mergedThemeRef:c,isMountedRef:d,appearRef:fe(e,"internalAppear"),transformOriginRef:fe(e,"transformOrigin")});const B=M(()=>{const{common:{cubicBezierEaseOut:r},self:{boxShadow:l,color:P,textColor:T}}=c.value;return{"--n-bezier-ease-out":r,"--n-box-shadow":l,"--n-color":P,"--n-text-color":T}}),y=g?Ce("theme-class",void 0,B,e):void 0;return{mergedClsPrefix:i,namespace:t,isMounted:d,containerRef:n,presetProps:M(()=>he(e,No)),handleEsc:$,handleAfterLeave:C,handleClickoutside:v,handleBeforeLeave:f,doUpdateShow:h,handleNegativeClick:z,handlePositiveClick:b,handleCloseClick:p,cssVars:g?void 0:B,themeClass:y?.themeClass,onRender:y?.onRender}},render(){const{mergedClsPrefix:e}=this;return o(go,{to:this.to,show:this.show},{default:()=>{var n;(n=this.onRender)===null||n===void 0||n.call(this);const{unstableShowMask:i}=this;return de(o("div",{role:"none",ref:"containerRef",class:[`${e}-modal-container`,this.themeClass,this.namespace],style:this.cssVars},o(Ao,Object.assign({style:this.overlayStyle},this.$attrs,{ref:"bodyWrapper",displayDirective:this.displayDirective,show:this.show,preset:this.preset,autoFocus:this.autoFocus,trapFocus:this.trapFocus,blockScroll:this.blockScroll},this.presetProps,{onEsc:this.handleEsc,onClose:this.handleCloseClick,onNegativeClick:this.handleNegativeClick,onPositiveClick:this.handlePositiveClick,onBeforeLeave:this.handleBeforeLeave,onAfterEnter:this.onAfterEnter,onAfterLeave:this.handleAfterLeave,onClickoutside:i?void 0:this.handleClickoutside,renderMask:i?()=>{var t;return o(He,{name:"fade-in-transition",key:"mask",appear:(t=this.internalAppear)!==null&&t!==void 0?t:this.isMounted},{default:()=>this.show?o("div",{"aria-hidden":!0,ref:"containerRef",class:`${e}-modal-mask`,onClick:this.handleClickoutside}):null})}:void 0}),this.$slots)),[[ho,{zIndex:this.zIndex,enabled:this.show}]])}})}});export{qo as N};
