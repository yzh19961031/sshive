import{T as we,aJ as ye,aK as ke,U as Y,a7 as a,aL as q,a8 as A,ae as p,aM as Z,d as te,aN as L,D as n,a4 as R,aO as xe,af as Ce,J as Be,L as ae,W as Re,r as C,X as Se,v as T,Z as H,N as S,ai as E,P as m,_ as $e,A as M,o as _e,c as Ve,b as G,e as f,w as x,g as v,B as U,$ as ze,i as Fe,k as X}from"./index-c8EV6_xx.js";import{b as Ne}from"./Select-CjWFaPBz.js";import{N as De}from"./DataTable-CUqEd3MV.js";import{N as Te,a as Q}from"./FormItem-CAeWu5gx.js";import{N as ee}from"./Input-CoIezaDs.js";import{_ as Me}from"./_plugin-vue_export-helper-DlAUqK2U.js";import"./prop-Cv5hXau6.js";import"./Dropdown-D6t1W9m4.js";const Ue=e=>{const{primaryColor:c,opacityDisabled:u,borderRadius:r,textColor3:s}=e;return Object.assign(Object.assign({},ye),{iconColor:s,textColor:"white",loadingColor:c,opacityDisabled:u,railColor:"rgba(0, 0, 0, .14)",railColorActive:c,buttonBoxShadow:"0 1px 4px 0 rgba(0, 0, 0, 0.3), inset 0 0 1px 0 rgba(0, 0, 0, 0.05)",buttonColor:"#FFF",railBorderRadiusSmall:r,railBorderRadiusMedium:r,railBorderRadiusLarge:r,buttonBorderRadiusSmall:r,buttonBorderRadiusMedium:r,buttonBorderRadiusLarge:r,boxShadowFocus:`0 0 0 2px ${ke(c,{alpha:.2})}`})},We={common:we,self:Ue},Ie=Y("switch",`
 height: var(--n-height);
 min-width: var(--n-width);
 vertical-align: middle;
 user-select: none;
 -webkit-user-select: none;
 display: inline-flex;
 outline: none;
 justify-content: center;
 align-items: center;
`,[a("children-placeholder",`
 height: var(--n-rail-height);
 display: flex;
 flex-direction: column;
 overflow: hidden;
 pointer-events: none;
 visibility: hidden;
 `),a("rail-placeholder",`
 display: flex;
 flex-wrap: none;
 `),a("button-placeholder",`
 width: calc(1.75 * var(--n-rail-height));
 height: var(--n-rail-height);
 `),Y("base-loading",`
 position: absolute;
 top: 50%;
 left: 50%;
 transform: translateX(-50%) translateY(-50%);
 font-size: calc(var(--n-button-width) - 4px);
 color: var(--n-loading-color);
 transition: color .3s var(--n-bezier);
 `,[q({left:"50%",top:"50%",originalTransform:"translateX(-50%) translateY(-50%)"})]),a("checked, unchecked",`
 transition: color .3s var(--n-bezier);
 color: var(--n-text-color);
 box-sizing: border-box;
 position: absolute;
 white-space: nowrap;
 top: 0;
 bottom: 0;
 display: flex;
 align-items: center;
 line-height: 1;
 `),a("checked",`
 right: 0;
 padding-right: calc(1.25 * var(--n-rail-height) - var(--n-offset));
 `),a("unchecked",`
 left: 0;
 justify-content: flex-end;
 padding-left: calc(1.25 * var(--n-rail-height) - var(--n-offset));
 `),A("&:focus",[a("rail",`
 box-shadow: var(--n-box-shadow-focus);
 `)]),p("round",[a("rail","border-radius: calc(var(--n-rail-height) / 2);",[a("button","border-radius: calc(var(--n-button-height) / 2);")])]),Z("disabled",[Z("icon",[p("rubber-band",[p("pressed",[a("rail",[a("button","max-width: var(--n-button-width-pressed);")])]),a("rail",[A("&:active",[a("button","max-width: var(--n-button-width-pressed);")])]),p("active",[p("pressed",[a("rail",[a("button","left: calc(100% - var(--n-offset) - var(--n-button-width-pressed));")])]),a("rail",[A("&:active",[a("button","left: calc(100% - var(--n-offset) - var(--n-button-width-pressed));")])])])])])]),p("active",[a("rail",[a("button","left: calc(100% - var(--n-button-width) - var(--n-offset))")])]),a("rail",`
 overflow: hidden;
 height: var(--n-rail-height);
 min-width: var(--n-rail-width);
 border-radius: var(--n-rail-border-radius);
 cursor: pointer;
 position: relative;
 transition:
 opacity .3s var(--n-bezier),
 background .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 background-color: var(--n-rail-color);
 `,[a("button-icon",`
 color: var(--n-icon-color);
 transition: color .3s var(--n-bezier);
 font-size: calc(var(--n-button-height) - 4px);
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 display: flex;
 justify-content: center;
 align-items: center;
 line-height: 1;
 `,[q()]),a("button",`
 align-items: center; 
 top: var(--n-offset);
 left: var(--n-offset);
 height: var(--n-button-height);
 width: var(--n-button-width-pressed);
 max-width: var(--n-button-width);
 border-radius: var(--n-button-border-radius);
 background-color: var(--n-button-color);
 box-shadow: var(--n-button-box-shadow);
 box-sizing: border-box;
 cursor: inherit;
 content: "";
 position: absolute;
 transition:
 background-color .3s var(--n-bezier),
 left .3s var(--n-bezier),
 opacity .3s var(--n-bezier),
 max-width .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 `)]),p("active",[a("rail","background-color: var(--n-rail-color-active);")]),p("loading",[a("rail",`
 cursor: wait;
 `)]),p("disabled",[a("rail",`
 cursor: not-allowed;
 opacity: .5;
 `)])]),Ke=Object.assign(Object.assign({},ae.props),{size:{type:String,default:"medium"},value:{type:[String,Number,Boolean],default:void 0},loading:Boolean,defaultValue:{type:[String,Number,Boolean],default:!1},disabled:{type:Boolean,default:void 0},round:{type:Boolean,default:!0},"onUpdate:value":[Function,Array],onUpdateValue:[Function,Array],checkedValue:{type:[String,Number,Boolean],default:!0},uncheckedValue:{type:[String,Number,Boolean],default:!1},railStyle:Function,rubberBand:{type:Boolean,default:!0},onChange:[Function,Array]});let z;const Oe=te({name:"Switch",props:Ke,setup(e){z===void 0&&(typeof CSS<"u"?typeof CSS.supports<"u"?z=CSS.supports("width","max(1px)"):z=!1:z=!0);const{mergedClsPrefixRef:c,inlineThemeDisabled:u}=Be(e),r=ae("Switch","-switch",Ie,We,e,c),s=Re(e),{mergedSizeRef:d,mergedDisabledRef:b}=s,h=C(e.defaultValue),$=$e(e,"value"),g=Ne($,h),B=T(()=>g.value===e.checkedValue),i=C(!1),t=C(!1),l=T(()=>{const{railStyle:o}=e;if(o)return o({focused:t.value,checked:B.value})});function w(o){const{"onUpdate:value":F,onChange:N,onUpdateValue:D}=e,{nTriggerFormInput:I,nTriggerFormChange:K}=s;F&&H(F,o),D&&H(D,o),N&&H(N,o),h.value=o,I(),K()}function oe(){const{nTriggerFormFocus:o}=s;o()}function ie(){const{nTriggerFormBlur:o}=s;o()}function ne(){e.loading||b.value||(g.value!==e.checkedValue?w(e.checkedValue):w(e.uncheckedValue))}function re(){t.value=!0,oe()}function le(){t.value=!1,ie(),i.value=!1}function se(o){e.loading||b.value||o.key===" "&&(g.value!==e.checkedValue?w(e.checkedValue):w(e.uncheckedValue),i.value=!1)}function de(o){e.loading||b.value||o.key===" "&&(o.preventDefault(),i.value=!0)}const J=T(()=>{const{value:o}=d,{self:{opacityDisabled:F,railColor:N,railColorActive:D,buttonBoxShadow:I,buttonColor:K,boxShadowFocus:ce,loadingColor:ue,textColor:he,iconColor:fe,[S("buttonHeight",o)]:y,[S("buttonWidth",o)]:ve,[S("buttonWidthPressed",o)]:be,[S("railHeight",o)]:k,[S("railWidth",o)]:V,[S("railBorderRadius",o)]:ge,[S("buttonBorderRadius",o)]:pe},common:{cubicBezierEaseInOut:me}}=r.value;let O,P,j;return z?(O=`calc((${k} - ${y}) / 2)`,P=`max(${k}, ${y})`,j=`max(${V}, calc(${V} + ${y} - ${k}))`):(O=E((m(k)-m(y))/2),P=E(Math.max(m(k),m(y))),j=m(k)>m(y)?V:E(m(V)+m(y)-m(k))),{"--n-bezier":me,"--n-button-border-radius":pe,"--n-button-box-shadow":I,"--n-button-color":K,"--n-button-width":ve,"--n-button-width-pressed":be,"--n-button-height":y,"--n-height":P,"--n-offset":O,"--n-opacity-disabled":F,"--n-rail-border-radius":ge,"--n-rail-color":N,"--n-rail-color-active":D,"--n-rail-height":k,"--n-rail-width":V,"--n-width":j,"--n-box-shadow-focus":ce,"--n-loading-color":ue,"--n-text-color":he,"--n-icon-color":fe}}),_=u?Se("switch",T(()=>d.value[0]),J,e):void 0;return{handleClick:ne,handleBlur:le,handleFocus:re,handleKeyup:se,handleKeydown:de,mergedRailStyle:l,pressed:i,mergedClsPrefix:c,mergedValue:g,checked:B,mergedDisabled:b,cssVars:u?void 0:J,themeClass:_?.themeClass,onRender:_?.onRender}},render(){const{mergedClsPrefix:e,mergedDisabled:c,checked:u,mergedRailStyle:r,onRender:s,$slots:d}=this;s?.();const{checked:b,unchecked:h,icon:$,"checked-icon":g,"unchecked-icon":B}=d,i=!(L($)&&L(g)&&L(B));return n("div",{role:"switch","aria-checked":u,class:[`${e}-switch`,this.themeClass,i&&`${e}-switch--icon`,u&&`${e}-switch--active`,c&&`${e}-switch--disabled`,this.round&&`${e}-switch--round`,this.loading&&`${e}-switch--loading`,this.pressed&&`${e}-switch--pressed`,this.rubberBand&&`${e}-switch--rubber-band`],tabindex:this.mergedDisabled?void 0:0,style:this.cssVars,onClick:this.handleClick,onFocus:this.handleFocus,onBlur:this.handleBlur,onKeyup:this.handleKeyup,onKeydown:this.handleKeydown},n("div",{class:`${e}-switch__rail`,"aria-hidden":"true",style:r},R(b,t=>R(h,l=>t||l?n("div",{"aria-hidden":!0,class:`${e}-switch__children-placeholder`},n("div",{class:`${e}-switch__rail-placeholder`},n("div",{class:`${e}-switch__button-placeholder`}),t),n("div",{class:`${e}-switch__rail-placeholder`},n("div",{class:`${e}-switch__button-placeholder`}),l)):null)),n("div",{class:`${e}-switch__button`},R($,t=>R(g,l=>R(B,w=>n(xe,null,{default:()=>this.loading?n(Ce,{key:"loading",clsPrefix:e,strokeWidth:20}):this.checked&&(l||t)?n("div",{class:`${e}-switch__button-icon`,key:l?"checked-icon":"icon"},l||t):!this.checked&&(w||t)?n("div",{class:`${e}-switch__button-icon`,key:w?"unchecked-icon":"icon"},w||t):null})))),R(b,t=>t&&n("div",{key:"checked",class:`${e}-switch__checked`},t)),R(h,t=>t&&n("div",{key:"unchecked",class:`${e}-switch__unchecked`},t)))))}}),W={list:e=>M.get("/danger-rules",{params:e}),create:e=>M.post("/danger-rules",e),update:(e,c)=>M.put(`/danger-rules/${e}`,c),delete:e=>M.delete(`/danger-rules/${e}`)},Pe={class:"page"},je={class:"page-header"},Ae=te({__name:"DangerRulesView",setup(e){const c=C([]),u=C(!1),r=C(!1),s=C(!1),d=C({pattern:"",description:""}),b=[{title:"ID",key:"id",width:80},{title:"正则表达式",key:"pattern",render:i=>n("code",{style:"font-family:monospace;color:var(--danger)"},i.pattern)},{title:"说明",key:"description"},{title:"动作",key:"action",width:80},{title:"启用",key:"enabled",width:80,render:i=>n(Oe,{value:i.enabled===1,onUpdateValue:t=>B(i.id,t)})},{title:"操作",key:"actions",width:80,render:i=>n(U,{size:"tiny",type:"error",onClick:()=>g(i.id)},{default:()=>"删除"})}];async function h(){u.value=!0;const i=await W.list({page:1,page_size:100});c.value=i.data.data?.list??[],u.value=!1}async function $(){s.value=!0,await W.create({pattern:d.value.pattern,description:d.value.description}),r.value=!1,d.value={pattern:"",description:""},await h(),s.value=!1}async function g(i){confirm("确认删除？")&&(await W.delete(i),await h())}async function B(i,t){await W.update(i,{enabled:t?1:0}),await h()}return _e(h),(i,t)=>(Fe(),Ve("div",Pe,[G("div",je,[t[6]||(t[6]=G("h2",{class:"page-title"},"高危指令规则",-1)),f(v(U),{type:"primary",size:"small",onClick:t[0]||(t[0]=l=>r.value=!0)},{default:x(()=>[...t[5]||(t[5]=[X("+ 添加规则",-1)])]),_:1})]),f(v(De),{columns:b,data:c.value,loading:u.value},null,8,["data","loading"]),f(v(ze),{show:r.value,"onUpdate:show":t[4]||(t[4]=l=>r.value=l),title:"添加规则",preset:"card",style:{width:"480px"}},{footer:x(()=>[f(v(U),{onClick:t[3]||(t[3]=l=>r.value=!1)},{default:x(()=>[...t[7]||(t[7]=[X("取消",-1)])]),_:1}),f(v(U),{type:"primary",loading:s.value,onClick:$},{default:x(()=>[...t[8]||(t[8]=[X("保存",-1)])]),_:1},8,["loading"])]),default:x(()=>[f(v(Te),{model:d.value},{default:x(()=>[f(v(Q),{label:"正则表达式",required:""},{default:x(()=>[f(v(ee),{value:d.value.pattern,"onUpdate:value":t[1]||(t[1]=l=>d.value.pattern=l),placeholder:"rm\\s+-rf\\s+/"},null,8,["value"])]),_:1}),f(v(Q),{label:"说明"},{default:x(()=>[f(v(ee),{value:d.value.description,"onUpdate:value":t[2]||(t[2]=l=>d.value.description=l)},null,8,["value"])]),_:1})]),_:1},8,["model"])]),_:1},8,["show"])]))}}),Ge=Me(Ae,[["__scopeId","data-v-0c0904dc"]]);export{Ge as default};
