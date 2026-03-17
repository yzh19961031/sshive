import{d as q,an as re,D as i,aq as _e,a0 as he,ar as Te,as as fe,at as oe,r as c,J as me,M as Re,ag as Ee,ak as De,au as Me,av as Oe,v as F,m as ve,aw as ae,ax as Fe,ac as V,ay as Ne,az as He,aA as Ie,a8 as l,aB as K,U as x,ae as D,a7 as W,ad as Pe,aC as Ue,aD as Ae,L as ge,aE as Le,X as je,aF as We,aG as Xe,Z as A,aH as Ye,_ as ne,aI as Ve,l as qe,o as Ke,c as Z,b as E,e as _,g as T,w as L,$ as se,i as Q,j as Ge,B as Y,k as ie,t as ee,Y as Je}from"./index-C5l8tqEC.js";import{c as Ze}from"./asciinema-player-CNVN4q_z.js";import{s as te}from"./session-D3Gx3Cpe.js";import{b as le,g as de,N as Qe,a as ce}from"./Select-ibCYC32C.js";import{N as et}from"./DatePicker-C36RmQO4.js";import{N as ue}from"./DataTable-CX3xdsBQ.js";import{N as tt}from"./Input-DXfuiyvf.js";import{_ as rt}from"./_plugin-vue_export-helper-DlAUqK2U.js";import"./Dropdown-DbBrVv67.js";import"./prop-Cv5hXau6.js";const at=q({name:"NDrawerContent",inheritAttrs:!1,props:{blockScroll:Boolean,show:{type:Boolean,default:void 0},displayDirective:{type:String,required:!0},placement:{type:String,required:!0},contentClass:String,contentStyle:[Object,String],nativeScrollbar:{type:Boolean,required:!0},scrollbarProps:Object,trapFocus:{type:Boolean,default:!0},autoFocus:{type:Boolean,default:!0},showMask:{type:[Boolean,String],required:!0},maxWidth:Number,maxHeight:Number,minWidth:Number,minHeight:Number,resizable:Boolean,onClickoutside:Function,onAfterLeave:Function,onAfterEnter:Function,onEsc:Function},setup(t){const r=c(!!t.show),a=c(null),y=ve(ae);let S=0,R="",f=null;const C=c(!1),z=c(!1),w=F(()=>t.placement==="top"||t.placement==="bottom"),{mergedClsPrefixRef:k,mergedRtlRef:$}=me(t),N=Re("Drawer",$,k),M=n=>{z.value=!0,S=w.value?n.clientY:n.clientX,R=document.body.style.cursor,document.body.style.cursor=w.value?"ns-resize":"ew-resize",document.body.addEventListener("mousemove",U),document.body.addEventListener("mouseleave",s),document.body.addEventListener("mouseup",m)},p=()=>{f!==null&&(window.clearTimeout(f),f=null),z.value?C.value=!0:f=window.setTimeout(()=>{C.value=!0},300)},O=()=>{f!==null&&(window.clearTimeout(f),f=null),C.value=!1},{doUpdateHeight:H,doUpdateWidth:I}=y,j=n=>{const{maxWidth:h}=t;if(h&&n>h)return h;const{minWidth:v}=t;return v&&n<v?v:n},P=n=>{const{maxHeight:h}=t;if(h&&n>h)return h;const{minHeight:v}=t;return v&&n<v?v:n},U=n=>{var h,v;if(z.value)if(w.value){let b=((h=a.value)===null||h===void 0?void 0:h.offsetHeight)||0;const B=S-n.clientY;b+=t.placement==="bottom"?B:-B,b=P(b),H(b),S=n.clientY}else{let b=((v=a.value)===null||v===void 0?void 0:v.offsetWidth)||0;const B=S-n.clientX;b+=t.placement==="right"?B:-B,b=j(b),I(b),S=n.clientX}},m=()=>{z.value&&(S=0,z.value=!1,document.body.style.cursor=R,document.body.removeEventListener("mousemove",U),document.body.removeEventListener("mouseup",m),document.body.removeEventListener("mouseleave",s))},s=m;Ee(()=>{t.show&&(r.value=!0)}),De(()=>t.show,n=>{n||m()}),Me(()=>{m()});const u=F(()=>{const{show:n}=t,h=[[oe,n]];return t.showMask||h.push([Fe,t.onClickoutside,void 0,{capture:!0}]),h});function g(){var n;r.value=!1,(n=t.onAfterLeave)===null||n===void 0||n.call(t)}return Oe(F(()=>t.blockScroll&&r.value)),V(Ne,a),V(He,null),V(Ie,null),{bodyRef:a,rtlEnabled:N,mergedClsPrefix:y.mergedClsPrefixRef,isMounted:y.isMountedRef,mergedTheme:y.mergedThemeRef,displayed:r,transitionName:F(()=>({right:"slide-in-from-right-transition",left:"slide-in-from-left-transition",top:"slide-in-from-top-transition",bottom:"slide-in-from-bottom-transition"})[t.placement]),handleAfterLeave:g,bodyDirectives:u,handleMousedownResizeTrigger:M,handleMouseenterResizeTrigger:p,handleMouseleaveResizeTrigger:O,isDragging:z,isHoverOnResizeTrigger:C}},render(){const{$slots:t,mergedClsPrefix:r}=this;return this.displayDirective==="show"||this.displayed||this.show?re(i("div",{role:"none"},i(_e,{disabled:!this.showMask||!this.trapFocus,active:this.show,autoFocus:this.autoFocus,onEsc:this.onEsc},{default:()=>i(he,{name:this.transitionName,appear:this.isMounted,onAfterEnter:this.onAfterEnter,onAfterLeave:this.handleAfterLeave},{default:()=>re(i("div",Te(this.$attrs,{role:"dialog",ref:"bodyRef","aria-modal":"true",class:[`${r}-drawer`,this.rtlEnabled&&`${r}-drawer--rtl`,`${r}-drawer--${this.placement}-placement`,this.isDragging&&`${r}-drawer--unselectable`,this.nativeScrollbar&&`${r}-drawer--native-scrollbar`]}),[this.resizable?i("div",{class:[`${r}-drawer__resize-trigger`,(this.isDragging||this.isHoverOnResizeTrigger)&&`${r}-drawer__resize-trigger--hover`],onMouseenter:this.handleMouseenterResizeTrigger,onMouseleave:this.handleMouseleaveResizeTrigger,onMousedown:this.handleMousedownResizeTrigger}):null,this.nativeScrollbar?i("div",{class:[`${r}-drawer-content-wrapper`,this.contentClass],style:this.contentStyle,role:"none"},t):i(fe,Object.assign({},this.scrollbarProps,{contentStyle:this.contentStyle,contentClass:[`${r}-drawer-content-wrapper`,this.contentClass],theme:this.mergedTheme.peers.Scrollbar,themeOverrides:this.mergedTheme.peerOverrides.Scrollbar}),t)]),this.bodyDirectives)})})),[[oe,this.displayDirective==="if"||this.displayed||this.show]]):null}}),{cubicBezierEaseIn:ot,cubicBezierEaseOut:nt}=K;function st({duration:t="0.3s",leaveDuration:r="0.2s",name:a="slide-in-from-right"}={}){return[l(`&.${a}-transition-leave-active`,{transition:`transform ${r} ${ot}`}),l(`&.${a}-transition-enter-active`,{transition:`transform ${t} ${nt}`}),l(`&.${a}-transition-enter-to`,{transform:"translateX(0)"}),l(`&.${a}-transition-enter-from`,{transform:"translateX(100%)"}),l(`&.${a}-transition-leave-from`,{transform:"translateX(0)"}),l(`&.${a}-transition-leave-to`,{transform:"translateX(100%)"})]}const{cubicBezierEaseIn:it,cubicBezierEaseOut:lt}=K;function dt({duration:t="0.3s",leaveDuration:r="0.2s",name:a="slide-in-from-left"}={}){return[l(`&.${a}-transition-leave-active`,{transition:`transform ${r} ${it}`}),l(`&.${a}-transition-enter-active`,{transition:`transform ${t} ${lt}`}),l(`&.${a}-transition-enter-to`,{transform:"translateX(0)"}),l(`&.${a}-transition-enter-from`,{transform:"translateX(-100%)"}),l(`&.${a}-transition-leave-from`,{transform:"translateX(0)"}),l(`&.${a}-transition-leave-to`,{transform:"translateX(-100%)"})]}const{cubicBezierEaseIn:ct,cubicBezierEaseOut:ut}=K;function ht({duration:t="0.3s",leaveDuration:r="0.2s",name:a="slide-in-from-top"}={}){return[l(`&.${a}-transition-leave-active`,{transition:`transform ${r} ${ct}`}),l(`&.${a}-transition-enter-active`,{transition:`transform ${t} ${ut}`}),l(`&.${a}-transition-enter-to`,{transform:"translateY(0)"}),l(`&.${a}-transition-enter-from`,{transform:"translateY(-100%)"}),l(`&.${a}-transition-leave-from`,{transform:"translateY(0)"}),l(`&.${a}-transition-leave-to`,{transform:"translateY(-100%)"})]}const{cubicBezierEaseIn:ft,cubicBezierEaseOut:mt}=K;function vt({duration:t="0.3s",leaveDuration:r="0.2s",name:a="slide-in-from-bottom"}={}){return[l(`&.${a}-transition-leave-active`,{transition:`transform ${r} ${ft}`}),l(`&.${a}-transition-enter-active`,{transition:`transform ${t} ${mt}`}),l(`&.${a}-transition-enter-to`,{transform:"translateY(0)"}),l(`&.${a}-transition-enter-from`,{transform:"translateY(100%)"}),l(`&.${a}-transition-leave-from`,{transform:"translateY(0)"}),l(`&.${a}-transition-leave-to`,{transform:"translateY(100%)"})]}const gt=l([x("drawer",`
 word-break: break-word;
 line-height: var(--n-line-height);
 position: absolute;
 pointer-events: all;
 box-shadow: var(--n-box-shadow);
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 background-color: var(--n-color);
 color: var(--n-text-color);
 box-sizing: border-box;
 `,[st(),dt(),ht(),vt(),D("unselectable",`
 user-select: none; 
 -webkit-user-select: none;
 `),D("native-scrollbar",[x("drawer-content-wrapper",`
 overflow: auto;
 height: 100%;
 `)]),W("resize-trigger",`
 position: absolute;
 background-color: #0000;
 transition: background-color .3s var(--n-bezier);
 `,[D("hover",`
 background-color: var(--n-resize-trigger-color-hover);
 `)]),x("drawer-content-wrapper",`
 box-sizing: border-box;
 `),x("drawer-content",`
 height: 100%;
 display: flex;
 flex-direction: column;
 `,[D("native-scrollbar",[x("drawer-body-content-wrapper",`
 height: 100%;
 overflow: auto;
 `)]),x("drawer-body",`
 flex: 1 0 0;
 overflow: hidden;
 `),x("drawer-body-content-wrapper",`
 box-sizing: border-box;
 padding: var(--n-body-padding);
 `),x("drawer-header",`
 font-weight: var(--n-title-font-weight);
 line-height: 1;
 font-size: var(--n-title-font-size);
 color: var(--n-title-text-color);
 padding: var(--n-header-padding);
 transition: border .3s var(--n-bezier);
 border-bottom: 1px solid var(--n-divider-color);
 border-bottom: var(--n-header-border-bottom);
 display: flex;
 justify-content: space-between;
 align-items: center;
 `,[W("close",`
 margin-left: 6px;
 transition:
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
 `)]),x("drawer-footer",`
 display: flex;
 justify-content: flex-end;
 border-top: var(--n-footer-border-top);
 transition: border .3s var(--n-bezier);
 padding: var(--n-footer-padding);
 `)]),D("right-placement",`
 top: 0;
 bottom: 0;
 right: 0;
 border-top-left-radius: var(--n-border-radius);
 border-bottom-left-radius: var(--n-border-radius);
 `,[W("resize-trigger",`
 width: 3px;
 height: 100%;
 top: 0;
 left: 0;
 transform: translateX(-1.5px);
 cursor: ew-resize;
 `)]),D("left-placement",`
 top: 0;
 bottom: 0;
 left: 0;
 border-top-right-radius: var(--n-border-radius);
 border-bottom-right-radius: var(--n-border-radius);
 `,[W("resize-trigger",`
 width: 3px;
 height: 100%;
 top: 0;
 right: 0;
 transform: translateX(1.5px);
 cursor: ew-resize;
 `)]),D("top-placement",`
 top: 0;
 left: 0;
 right: 0;
 border-bottom-left-radius: var(--n-border-radius);
 border-bottom-right-radius: var(--n-border-radius);
 `,[W("resize-trigger",`
 width: 100%;
 height: 3px;
 bottom: 0;
 left: 0;
 transform: translateY(1.5px);
 cursor: ns-resize;
 `)]),D("bottom-placement",`
 left: 0;
 bottom: 0;
 right: 0;
 border-top-left-radius: var(--n-border-radius);
 border-top-right-radius: var(--n-border-radius);
 `,[W("resize-trigger",`
 width: 100%;
 height: 3px;
 top: 0;
 left: 0;
 transform: translateY(-1.5px);
 cursor: ns-resize;
 `)])]),l("body",[l(">",[x("drawer-container",`
 position: fixed;
 `)])]),x("drawer-container",`
 position: relative;
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 pointer-events: none;
 `,[l("> *",`
 pointer-events: all;
 `)]),x("drawer-mask",`
 background-color: rgba(0, 0, 0, .3);
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `,[D("invisible",`
 background-color: rgba(0, 0, 0, 0)
 `),Pe({enterDuration:"0.2s",leaveDuration:"0.2s",enterCubicBezier:"var(--n-bezier-in)",leaveCubicBezier:"var(--n-bezier-out)"})])]),bt=Object.assign(Object.assign({},ge.props),{show:Boolean,width:[Number,String],height:[Number,String],placement:{type:String,default:"right"},maskClosable:{type:Boolean,default:!0},showMask:{type:[Boolean,String],default:!0},to:[String,Object],displayDirective:{type:String,default:"if"},nativeScrollbar:{type:Boolean,default:!0},zIndex:Number,onMaskClick:Function,scrollbarProps:Object,contentClass:String,contentStyle:[Object,String],trapFocus:{type:Boolean,default:!0},onEsc:Function,autoFocus:{type:Boolean,default:!0},closeOnEsc:{type:Boolean,default:!0},blockScroll:{type:Boolean,default:!0},maxWidth:Number,maxHeight:Number,minWidth:Number,minHeight:Number,resizable:Boolean,defaultWidth:{type:[Number,String],default:251},defaultHeight:{type:[Number,String],default:251},onUpdateWidth:[Function,Array],onUpdateHeight:[Function,Array],"onUpdate:width":[Function,Array],"onUpdate:height":[Function,Array],"onUpdate:show":[Function,Array],onUpdateShow:[Function,Array],onAfterEnter:Function,onAfterLeave:Function,drawerStyle:[String,Object],drawerClass:String,target:null,onShow:Function,onHide:Function}),pt=q({name:"Drawer",inheritAttrs:!1,props:bt,setup(t){const{mergedClsPrefixRef:r,namespaceRef:a,inlineThemeDisabled:y}=me(t),S=Ae(),R=ge("Drawer","-drawer",gt,Ye,t,r),f=c(t.defaultWidth),C=c(t.defaultHeight),z=le(ne(t,"width"),f),w=le(ne(t,"height"),C),k=F(()=>{const{placement:s}=t;return s==="top"||s==="bottom"?"":de(z.value)}),$=F(()=>{const{placement:s}=t;return s==="left"||s==="right"?"":de(w.value)}),N=s=>{const{onUpdateWidth:u,"onUpdate:width":g}=t;u&&A(u,s),g&&A(g,s),f.value=s},M=s=>{const{onUpdateHeight:u,"onUpdate:width":g}=t;u&&A(u,s),g&&A(g,s),C.value=s},p=F(()=>[{width:k.value,height:$.value},t.drawerStyle||""]);function O(s){const{onMaskClick:u,maskClosable:g}=t;g&&P(!1),u&&u(s)}function H(s){O(s)}const I=Le();function j(s){var u;(u=t.onEsc)===null||u===void 0||u.call(t),t.show&&t.closeOnEsc&&Xe(s)&&!I.value&&P(!1)}function P(s){const{onHide:u,onUpdateShow:g,"onUpdate:show":n}=t;g&&A(g,s),n&&A(n,s),u&&!s&&A(u,s)}V(ae,{isMountedRef:S,mergedThemeRef:R,mergedClsPrefixRef:r,doUpdateShow:P,doUpdateHeight:M,doUpdateWidth:N});const U=F(()=>{const{common:{cubicBezierEaseInOut:s,cubicBezierEaseIn:u,cubicBezierEaseOut:g},self:{color:n,textColor:h,boxShadow:v,lineHeight:b,headerPadding:B,footerPadding:G,borderRadius:J,bodyPadding:X,titleFontSize:e,titleTextColor:o,titleFontWeight:d,headerBorderBottom:be,footerBorderTop:pe,closeIconColor:ye,closeIconColorHover:we,closeIconColorPressed:Se,closeColorHover:Ce,closeColorPressed:ze,closeIconSize:$e,closeSize:ke,closeBorderRadius:xe,resizableTriggerColorHover:Be}}=R.value;return{"--n-line-height":b,"--n-color":n,"--n-border-radius":J,"--n-text-color":h,"--n-box-shadow":v,"--n-bezier":s,"--n-bezier-out":g,"--n-bezier-in":u,"--n-header-padding":B,"--n-body-padding":X,"--n-footer-padding":G,"--n-title-text-color":o,"--n-title-font-size":e,"--n-title-font-weight":d,"--n-header-border-bottom":be,"--n-footer-border-top":pe,"--n-close-icon-color":ye,"--n-close-icon-color-hover":we,"--n-close-icon-color-pressed":Se,"--n-close-size":ke,"--n-close-color-hover":Ce,"--n-close-color-pressed":ze,"--n-close-icon-size":$e,"--n-close-border-radius":xe,"--n-resize-trigger-color-hover":Be}}),m=y?je("drawer",void 0,U,t):void 0;return{mergedClsPrefix:r,namespace:a,mergedBodyStyle:p,handleOutsideClick:H,handleMaskClick:O,handleEsc:j,mergedTheme:R,cssVars:y?void 0:U,themeClass:m?.themeClass,onRender:m?.onRender,isMounted:S}},render(){const{mergedClsPrefix:t}=this;return i(Ue,{to:this.to,show:this.show},{default:()=>{var r;return(r=this.onRender)===null||r===void 0||r.call(this),re(i("div",{class:[`${t}-drawer-container`,this.namespace,this.themeClass],style:this.cssVars,role:"none"},this.showMask?i(he,{name:"fade-in-transition",appear:this.isMounted},{default:()=>this.show?i("div",{"aria-hidden":!0,class:[`${t}-drawer-mask`,this.showMask==="transparent"&&`${t}-drawer-mask--invisible`],onClick:this.handleMaskClick}):null}):null,i(at,Object.assign({},this.$attrs,{class:[this.drawerClass,this.$attrs.class],style:[this.mergedBodyStyle,this.$attrs.style],blockScroll:this.blockScroll,contentStyle:this.contentStyle,contentClass:this.contentClass,placement:this.placement,scrollbarProps:this.scrollbarProps,show:this.show,displayDirective:this.displayDirective,nativeScrollbar:this.nativeScrollbar,onAfterEnter:this.onAfterEnter,onAfterLeave:this.onAfterLeave,trapFocus:this.trapFocus,autoFocus:this.autoFocus,resizable:this.resizable,maxHeight:this.maxHeight,minHeight:this.minHeight,maxWidth:this.maxWidth,minWidth:this.minWidth,showMask:this.showMask,onEsc:this.handleEsc,onClickoutside:this.handleOutsideClick}),this.$slots)),[[We,{zIndex:this.zIndex,enabled:this.show}]])}})}}),yt={title:String,headerClass:String,headerStyle:[Object,String],footerClass:String,footerStyle:[Object,String],bodyClass:String,bodyStyle:[Object,String],bodyContentClass:String,bodyContentStyle:[Object,String],nativeScrollbar:{type:Boolean,default:!0},scrollbarProps:Object,closable:Boolean},wt=q({name:"DrawerContent",props:yt,setup(){const t=ve(ae,null);t||qe("drawer-content","`n-drawer-content` must be placed inside `n-drawer`.");const{doUpdateShow:r}=t;function a(){r(!1)}return{handleCloseClick:a,mergedTheme:t.mergedThemeRef,mergedClsPrefix:t.mergedClsPrefixRef}},render(){const{title:t,mergedClsPrefix:r,nativeScrollbar:a,mergedTheme:y,bodyClass:S,bodyStyle:R,bodyContentClass:f,bodyContentStyle:C,headerClass:z,headerStyle:w,footerClass:k,footerStyle:$,scrollbarProps:N,closable:M,$slots:p}=this;return i("div",{role:"none",class:[`${r}-drawer-content`,a&&`${r}-drawer-content--native-scrollbar`]},p.header||t||M?i("div",{class:[`${r}-drawer-header`,z],style:w,role:"none"},i("div",{class:`${r}-drawer-header__main`,role:"heading","aria-level":"1"},p.header!==void 0?p.header():t),M&&i(Ve,{onClick:this.handleCloseClick,clsPrefix:r,class:`${r}-drawer-header__close`,absolute:!0})):null,a?i("div",{class:[`${r}-drawer-body`,S],style:R,role:"none"},i("div",{class:[`${r}-drawer-body-content-wrapper`,f],style:C,role:"none"},p)):i(fe,Object.assign({themeOverrides:y.peerOverrides.Scrollbar,theme:y.peers.Scrollbar},N,{class:`${r}-drawer-body`,contentClass:[`${r}-drawer-body-content-wrapper`,f],contentStyle:C}),p),p.footer?i("div",{class:[`${r}-drawer-footer`,k],style:$,role:"none"},p.footer()):null)}}),St={class:"page"},Ct={class:"drawer-sub"},zt={class:"filter-bar"},$t={class:"output-pre"},kt={class:"replay-dialog"},xt={class:"replay-header"},Bt={class:"replay-title"},_t={key:0,class:"empty-tip",style:{padding:"60px 0"}},Tt=q({__name:"AuditSessionsView",setup(t){const r=c([]),a=c(!1),y=c({page:1,pageSize:20,itemCount:0});function S(e){return{active:"success",interrupted:"warning",closed:"default"}[e]??"default"}function R(e){const o=e.ended_at?new Date(e.ended_at):new Date,d=Math.floor((o.getTime()-new Date(e.started_at).getTime())/1e3);return d<60?`${d}s`:d<3600?`${Math.floor(d/60)}m ${d%60}s`:`${Math.floor(d/3600)}h ${Math.floor(d%3600/60)}m`}const f=c(!1),C=c(""),z=c(0),w=c(""),k=c(null),$=c(null),N=[{label:"执行",value:"execute"},{label:"阻断",value:"blocked"}],M={今天:()=>{const e=new Date;e.setHours(0,0,0,0);const o=new Date;return o.setHours(23,59,59,999),[e.getTime(),o.getTime()]},"近 7 天":()=>{const e=new Date;return[new Date(e.getTime()-7*864e5).getTime(),e.getTime()]}},p=c([]),O=c(!1),H=c({page:1,pageSize:20,itemCount:0}),I=c(!1),j=c(""),P=[{title:"命令",key:"command",width:200,ellipsis:{tooltip:!0},render:e=>i("code",{class:"cmd-code"},e.command)},{title:"动作",key:"action",width:72,render:e=>i(ce,{type:e.action==="blocked"?"error":"success",size:"small",round:!0},{default:()=>e.action==="blocked"?"阻断":"执行"})},{title:"执行时间",key:"created_at",width:160,render:e=>new Date(e.created_at).toLocaleString("zh-CN",{hour12:!1})},{title:"输出结果",key:"result",render:e=>{if(!e.result)return i("span",{style:"color:var(--text-secondary)"},"--");const o=e.result.replace(/\n/g," ").slice(0,50),d=e.result.length>50;return i("span",{class:"result-link",onClick:()=>{j.value=e.result,I.value=!0}},d?o+"…":o)}}];function U(){const e={};return w.value&&(e.command=w.value),k.value&&(e.action=k.value),$.value&&(e.start_time=new Date($.value[0]).toISOString(),e.end_time=new Date($.value[1]).toISOString()),e}async function m(e){O.value=!0;try{const o=await te.getCommands(z.value,{page:e,page_size:20,...U()});p.value=o.data.data?.list??[],H.value.itemCount=o.data.data?.total??0,H.value.page=e}finally{O.value=!1}}function s(){m(1)}function u(){w.value="",k.value=null,$.value=null,m(1)}function g(e){z.value=e.id,C.value=`${e.host_name||e.host_id} · ${e.username||e.user_id}`,w.value="",k.value=null,$.value=null,p.value=[],f.value=!0,m(1)}const n=c(!1),h=c(""),v=c(!1),b=c();let B=null;async function G(e){h.value=`${e.host_name||e.host_id} · ${e.username||e.user_id} · ${new Date(e.started_at).toLocaleString()}`,v.value=!!e.cast_file_path,n.value=!0,v.value&&(await Je(),b.value&&(b.value.innerHTML="",B?.dispose&&B.dispose(),B=Ze(te.getReplayUrl(e.id),b.value,{cols:160,rows:40,fit:"width",theme:"monokai",autoPlay:!0})))}const J=[{title:"主机",key:"host_name",width:160,ellipsis:{tooltip:!0},render:e=>e.host_name||String(e.host_id)},{title:"用户",key:"username",width:90,ellipsis:{tooltip:!0},render:e=>e.username||String(e.user_id)},{title:"客户端 IP",key:"client_ip",width:110,ellipsis:!0},{title:"状态",key:"status",width:80,render:e=>i(ce,{type:S(e.status),size:"small",round:!0},{default:()=>e.status})},{title:"时长",key:"duration",width:75,render:e=>R(e)},{title:"开始时间",key:"started_at",width:165,render:e=>new Date(e.started_at).toLocaleString("zh-CN",{hour12:!1})},{title:"操作",key:"actions",width:120,fixed:"right",render:e=>i("div",{style:"display:flex;gap:4px"},[i(Y,{size:"tiny",quaternary:!0,onClick:()=>g(e)},{default:()=>"📋 操作记录"}),i(Y,{size:"tiny",quaternary:!0,type:"primary",disabled:!e.cast_file_path,onClick:()=>G(e)},{default:()=>"▶ 回放"})])}];async function X(e){a.value=!0;try{const o=await te.list({page:e,page_size:20});r.value=o.data.data?.list??[],y.value.itemCount=o.data.data?.total??0,y.value.page=e}finally{a.value=!1}}return Ke(()=>X(1)),(e,o)=>(Q(),Z("div",St,[o[10]||(o[10]=E("div",{class:"page-header"},[E("h2",{class:"page-title"},"审计日志")],-1)),_(T(ue),{columns:J,data:r.value,loading:a.value,pagination:y.value,remote:"","onUpdate:page":X,"scroll-x":820},null,8,["data","loading","pagination"]),_(T(pt),{show:f.value,"onUpdate:show":o[3]||(o[3]=d=>f.value=d),width:760,placement:"right"},{default:L(()=>[_(T(wt),{closable:""},{header:L(()=>[o[7]||(o[7]=E("span",{class:"drawer-title"},"操作记录",-1)),E("span",Ct,ee(C.value),1)]),default:L(()=>[E("div",zt,[_(T(tt),{value:w.value,"onUpdate:value":o[0]||(o[0]=d=>w.value=d),placeholder:"请输入命令",clearable:"",size:"small",style:{width:"180px"},onKeyup:Ge(s,["enter"])},null,8,["value"]),_(T(Qe),{value:k.value,"onUpdate:value":o[1]||(o[1]=d=>k.value=d),options:N,placeholder:"请选择动作",clearable:"",size:"small",style:{width:"130px"}},null,8,["value"]),_(T(et),{value:$.value,"onUpdate:value":o[2]||(o[2]=d=>$.value=d),type:"datetimerange",clearable:"",size:"small",style:{width:"360px"},shortcuts:M},null,8,["value"]),_(T(Y),{type:"primary",size:"small",onClick:s},{default:L(()=>[...o[8]||(o[8]=[ie("查询",-1)])]),_:1}),_(T(Y),{size:"small",onClick:u},{default:L(()=>[...o[9]||(o[9]=[ie("重置",-1)])]),_:1})]),_(T(ue),{columns:P,data:p.value,loading:O.value,pagination:H.value,remote:"","onUpdate:page":m,"scroll-x":680,size:"small",style:{"margin-top":"12px"}},null,8,["data","loading","pagination"])]),_:1})]),_:1},8,["show"]),_(T(se),{show:I.value,"onUpdate:show":o[4]||(o[4]=d=>I.value=d),preset:"card",title:"命令输出",style:{width:"560px"}},{default:L(()=>[E("pre",$t,ee(j.value||"（无输出）"),1)]),_:1},8,["show"]),_(T(se),{show:n.value,"onUpdate:show":o[6]||(o[6]=d=>n.value=d),"mask-closable":!0,"display-directive":"show"},{default:L(()=>[E("div",kt,[E("div",xt,[E("span",Bt,ee(h.value),1),E("button",{class:"replay-close",onClick:o[5]||(o[5]=d=>n.value=!1)},"✕")]),v.value?(Q(),Z("div",{key:1,ref_key:"playerContainer",ref:b,class:"player-wrap"},null,512)):(Q(),Z("div",_t,"录像暂不可用"))])]),_:1},8,["show"])]))}}),Ut=rt(Tt,[["__scopeId","data-v-a06ae7bc"]]);export{Ut as default};
