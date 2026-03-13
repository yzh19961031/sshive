import{D as we,U as ye,V as xe,E as q,W as a,I as A,X as p,Y as G,d as te,p as n,s as ke,v as ae,r as C,x as T,y as R,H as Ce,M,o as Be,c as Re,b as J,e as f,w as k,g as v,i as Se,k as H}from"./index-DFqVa6t8.js";import{d as Q,e as E,r as S,N as $e,f as Ve,a as _e,c as L,B as U}from"./Button-BHK1_ql9.js";import{u as ze,p as X,d as m,a as Z}from"./Select-CPZmiopm.js";import{u as Fe}from"./use-css-vars-class-DIpD77cH.js";import{N as Ne}from"./DataTable-DXI0fapO.js";import{N as De}from"./Modal-0Ev_ydYs.js";import{N as Te,a as ee}from"./FormItem-97wS9nkl.js";import{_ as Me}from"./_plugin-vue_export-helper-DlAUqK2U.js";import"./Empty-DKD1pjzf.js";import"./prop-Cv5hXau6.js";const Ue=e=>{const{primaryColor:c,opacityDisabled:u,borderRadius:r,textColor3:s}=e;return Object.assign(Object.assign({},ye),{iconColor:s,textColor:"white",loadingColor:c,opacityDisabled:u,railColor:"rgba(0, 0, 0, .14)",railColorActive:c,buttonBoxShadow:"0 1px 4px 0 rgba(0, 0, 0, 0.3), inset 0 0 1px 0 rgba(0, 0, 0, 0.05)",buttonColor:"#FFF",railBorderRadiusSmall:r,railBorderRadiusMedium:r,railBorderRadiusLarge:r,buttonBorderRadiusSmall:r,buttonBorderRadiusMedium:r,buttonBorderRadiusLarge:r,boxShadowFocus:`0 0 0 2px ${xe(c,{alpha:.2})}`})},Ie={common:we,self:Ue},We=q("switch",`
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
 `),q("base-loading",`
 position: absolute;
 top: 50%;
 left: 50%;
 transform: translateX(-50%) translateY(-50%);
 font-size: calc(var(--n-button-width) - 4px);
 color: var(--n-loading-color);
 transition: color .3s var(--n-bezier);
 `,[Q({left:"50%",top:"50%",originalTransform:"translateX(-50%) translateY(-50%)"})]),a("checked, unchecked",`
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
 `)]),p("round",[a("rail","border-radius: calc(var(--n-rail-height) / 2);",[a("button","border-radius: calc(var(--n-button-height) / 2);")])]),G("disabled",[G("icon",[p("rubber-band",[p("pressed",[a("rail",[a("button","max-width: var(--n-button-width-pressed);")])]),a("rail",[A("&:active",[a("button","max-width: var(--n-button-width-pressed);")])]),p("active",[p("pressed",[a("rail",[a("button","left: calc(100% - var(--n-offset) - var(--n-button-width-pressed));")])]),a("rail",[A("&:active",[a("button","left: calc(100% - var(--n-offset) - var(--n-button-width-pressed));")])])])])])]),p("active",[a("rail",[a("button","left: calc(100% - var(--n-button-width) - var(--n-offset))")])]),a("rail",`
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
 `,[Q()]),a("button",`
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
 `)])]),je=Object.assign(Object.assign({},ae.props),{size:{type:String,default:"medium"},value:{type:[String,Number,Boolean],default:void 0},loading:Boolean,defaultValue:{type:[String,Number,Boolean],default:!1},disabled:{type:Boolean,default:void 0},round:{type:Boolean,default:!0},"onUpdate:value":[Function,Array],onUpdateValue:[Function,Array],checkedValue:{type:[String,Number,Boolean],default:!0},uncheckedValue:{type:[String,Number,Boolean],default:!1},railStyle:Function,rubberBand:{type:Boolean,default:!0},onChange:[Function,Array]});let z;const Ke=te({name:"Switch",props:je,setup(e){z===void 0&&(typeof CSS<"u"?typeof CSS.supports<"u"?z=CSS.supports("width","max(1px)"):z=!1:z=!0);const{mergedClsPrefixRef:c,inlineThemeDisabled:u}=ke(e),r=ae("Switch","-switch",We,Ie,e,c),s=_e(e),{mergedSizeRef:d,mergedDisabledRef:b}=s,h=C(e.defaultValue),$=Ce(e,"value"),g=ze($,h),B=T(()=>g.value===e.checkedValue),i=C(!1),t=C(!1),l=T(()=>{const{railStyle:o}=e;if(o)return o({focused:t.value,checked:B.value})});function w(o){const{"onUpdate:value":F,onChange:N,onUpdateValue:D}=e,{nTriggerFormInput:W,nTriggerFormChange:j}=s;F&&L(F,o),D&&L(D,o),N&&L(N,o),h.value=o,W(),j()}function oe(){const{nTriggerFormFocus:o}=s;o()}function ie(){const{nTriggerFormBlur:o}=s;o()}function ne(){e.loading||b.value||(g.value!==e.checkedValue?w(e.checkedValue):w(e.uncheckedValue))}function re(){t.value=!0,oe()}function le(){t.value=!1,ie(),i.value=!1}function se(o){e.loading||b.value||o.key===" "&&(g.value!==e.checkedValue?w(e.checkedValue):w(e.uncheckedValue),i.value=!1)}function de(o){e.loading||b.value||o.key===" "&&(o.preventDefault(),i.value=!0)}const Y=T(()=>{const{value:o}=d,{self:{opacityDisabled:F,railColor:N,railColorActive:D,buttonBoxShadow:W,buttonColor:j,boxShadowFocus:ce,loadingColor:ue,textColor:he,iconColor:fe,[R("buttonHeight",o)]:y,[R("buttonWidth",o)]:ve,[R("buttonWidthPressed",o)]:be,[R("railHeight",o)]:x,[R("railWidth",o)]:_,[R("railBorderRadius",o)]:ge,[R("buttonBorderRadius",o)]:pe},common:{cubicBezierEaseInOut:me}}=r.value;let K,O,P;return z?(K=`calc((${x} - ${y}) / 2)`,O=`max(${x}, ${y})`,P=`max(${_}, calc(${_} + ${y} - ${x}))`):(K=X((m(x)-m(y))/2),O=X(Math.max(m(x),m(y))),P=m(x)>m(y)?_:X(m(_)+m(y)-m(x))),{"--n-bezier":me,"--n-button-border-radius":pe,"--n-button-box-shadow":W,"--n-button-color":j,"--n-button-width":ve,"--n-button-width-pressed":be,"--n-button-height":y,"--n-height":O,"--n-offset":K,"--n-opacity-disabled":F,"--n-rail-border-radius":ge,"--n-rail-color":N,"--n-rail-color-active":D,"--n-rail-height":x,"--n-rail-width":_,"--n-width":P,"--n-box-shadow-focus":ce,"--n-loading-color":ue,"--n-text-color":he,"--n-icon-color":fe}}),V=u?Fe("switch",T(()=>d.value[0]),Y,e):void 0;return{handleClick:ne,handleBlur:le,handleFocus:re,handleKeyup:se,handleKeydown:de,mergedRailStyle:l,pressed:i,mergedClsPrefix:c,mergedValue:g,checked:B,mergedDisabled:b,cssVars:u?void 0:Y,themeClass:V?.themeClass,onRender:V?.onRender}},render(){const{mergedClsPrefix:e,mergedDisabled:c,checked:u,mergedRailStyle:r,onRender:s,$slots:d}=this;s?.();const{checked:b,unchecked:h,icon:$,"checked-icon":g,"unchecked-icon":B}=d,i=!(E($)&&E(g)&&E(B));return n("div",{role:"switch","aria-checked":u,class:[`${e}-switch`,this.themeClass,i&&`${e}-switch--icon`,u&&`${e}-switch--active`,c&&`${e}-switch--disabled`,this.round&&`${e}-switch--round`,this.loading&&`${e}-switch--loading`,this.pressed&&`${e}-switch--pressed`,this.rubberBand&&`${e}-switch--rubber-band`],tabindex:this.mergedDisabled?void 0:0,style:this.cssVars,onClick:this.handleClick,onFocus:this.handleFocus,onBlur:this.handleBlur,onKeyup:this.handleKeyup,onKeydown:this.handleKeydown},n("div",{class:`${e}-switch__rail`,"aria-hidden":"true",style:r},S(b,t=>S(h,l=>t||l?n("div",{"aria-hidden":!0,class:`${e}-switch__children-placeholder`},n("div",{class:`${e}-switch__rail-placeholder`},n("div",{class:`${e}-switch__button-placeholder`}),t),n("div",{class:`${e}-switch__rail-placeholder`},n("div",{class:`${e}-switch__button-placeholder`}),l)):null)),n("div",{class:`${e}-switch__button`},S($,t=>S(g,l=>S(B,w=>n($e,null,{default:()=>this.loading?n(Ve,{key:"loading",clsPrefix:e,strokeWidth:20}):this.checked&&(l||t)?n("div",{class:`${e}-switch__button-icon`,key:l?"checked-icon":"icon"},l||t):!this.checked&&(w||t)?n("div",{class:`${e}-switch__button-icon`,key:w?"unchecked-icon":"icon"},w||t):null})))),S(b,t=>t&&n("div",{key:"checked",class:`${e}-switch__checked`},t)),S(h,t=>t&&n("div",{key:"unchecked",class:`${e}-switch__unchecked`},t)))))}}),I={list:e=>M.get("/danger-rules",{params:e}),create:e=>M.post("/danger-rules",e),update:(e,c)=>M.put(`/danger-rules/${e}`,c),delete:e=>M.delete(`/danger-rules/${e}`)},Oe={class:"page"},Pe={class:"page-header"},Ae=te({__name:"DangerRulesView",setup(e){const c=C([]),u=C(!1),r=C(!1),s=C(!1),d=C({pattern:"",description:""}),b=[{title:"ID",key:"id",width:80},{title:"正则表达式",key:"pattern",render:i=>n("code",{style:"font-family:monospace;color:var(--danger)"},i.pattern)},{title:"说明",key:"description"},{title:"动作",key:"action",width:80},{title:"启用",key:"enabled",width:80,render:i=>n(Ke,{value:i.enabled===1,onUpdateValue:t=>B(i.id,t)})},{title:"操作",key:"actions",width:80,render:i=>n(U,{size:"tiny",type:"error",onClick:()=>g(i.id)},{default:()=>"删除"})}];async function h(){u.value=!0;const i=await I.list({page:1,page_size:100});c.value=i.data.data?.list??[],u.value=!1}async function $(){s.value=!0,await I.create({pattern:d.value.pattern,description:d.value.description}),r.value=!1,d.value={pattern:"",description:""},await h(),s.value=!1}async function g(i){confirm("确认删除？")&&(await I.delete(i),await h())}async function B(i,t){await I.update(i,{enabled:t?1:0}),await h()}return Be(h),(i,t)=>(Se(),Re("div",Oe,[J("div",Pe,[t[6]||(t[6]=J("h2",{class:"page-title"},"高危指令规则",-1)),f(v(U),{type:"primary",size:"small",onClick:t[0]||(t[0]=l=>r.value=!0)},{default:k(()=>[...t[5]||(t[5]=[H("+ 添加规则",-1)])]),_:1})]),f(v(Ne),{columns:b,data:c.value,loading:u.value},null,8,["data","loading"]),f(v(De),{show:r.value,"onUpdate:show":t[4]||(t[4]=l=>r.value=l),title:"添加规则",preset:"card",style:{width:"480px"}},{footer:k(()=>[f(v(U),{onClick:t[3]||(t[3]=l=>r.value=!1)},{default:k(()=>[...t[7]||(t[7]=[H("取消",-1)])]),_:1}),f(v(U),{type:"primary",loading:s.value,onClick:$},{default:k(()=>[...t[8]||(t[8]=[H("保存",-1)])]),_:1},8,["loading"])]),default:k(()=>[f(v(Te),{model:d.value},{default:k(()=>[f(v(ee),{label:"正则表达式",required:""},{default:k(()=>[f(v(Z),{value:d.value.pattern,"onUpdate:value":t[1]||(t[1]=l=>d.value.pattern=l),placeholder:"rm\\s+-rf\\s+/"},null,8,["value"])]),_:1}),f(v(ee),{label:"说明"},{default:k(()=>[f(v(Z),{value:d.value.description,"onUpdate:value":t[2]||(t[2]=l=>d.value.description=l)},null,8,["value"])]),_:1})]),_:1},8,["model"])]),_:1},8,["show"])]))}}),et=Me(Ae,[["__scopeId","data-v-0c0904dc"]]);export{et as default};
