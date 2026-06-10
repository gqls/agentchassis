var _____WB$wombat$assign$function_____=function(name){return (globalThis._wb_wombat && globalThis._wb_wombat.local_init && globalThis._wb_wombat.local_init(name))||globalThis[name];};if(!globalThis.__WB_pmw){globalThis.__WB_pmw=function(obj){this.__WB_source=obj;return this;}}{
let window = _____WB$wombat$assign$function_____("window");
let self = _____WB$wombat$assign$function_____("self");
let document = _____WB$wombat$assign$function_____("document");
let location = _____WB$wombat$assign$function_____("location");
let top = _____WB$wombat$assign$function_____("top");
let parent = _____WB$wombat$assign$function_____("parent");
let frames = _____WB$wombat$assign$function_____("frames");
let opener = _____WB$wombat$assign$function_____("opener");
/*======================================================================*\
|| #################################################################### ||
|| # vBulletin 4.2.3
|| # ---------------------------------------------------------------- # ||
|| # Copyright ©2000-2015 vBulletin Solutions Inc. All Rights Reserved. ||
|| # This file may not be redistributed in whole or significant part. # ||
|| # ---------------- VBULLETIN IS NOT FREE SOFTWARE ---------------- # ||
|| # http://www.vbulletin.com | http://www.vbulletin.com/license.html # ||
|| #################################################################### ||
\*======================================================================*/
vB_XHTML_Ready.subscribe(init_sidebar);function init_sidebar(){new vBSidebar()}function vBSidebar(){this.init()}vBSidebar.prototype.init=function(){this.sidebar_button=YAHOO.util.Dom.get("sidebar_button");this.sidebar_container=YAHOO.util.Dom.get("sidebar_container");this.sidebar=YAHOO.util.Dom.get("sidebar");this.content_container=YAHOO.util.Dom.get("content_container");this.content=YAHOO.util.Dom.get("content");YAHOO.util.Event.on(this.sidebar_button,"click",this.toggle_collapse,this,true)};vBSidebar.prototype.toggle_collapse=function(A){YAHOO.util.Event.stopEvent(A);if(YAHOO.util.Dom.getStyle(this.sidebar,"display")=="none"){this.expand()}else{this.collapse(true)}return false};vBSidebar.prototype.collapse=function(A){var B=this.sidebar_button;if(A){var C=new YAHOO.util.Anim(this.sidebar,{opacity:{from:1,to:0}},0.3);C.onComplete.subscribe(function(G,D,E){YAHOO.util.Dom.setStyle(E.sidebar,"display","none");YAHOO.util.Dom.setStyle(E.sidebar_container,"width","0");var F;if(sidebar_align=="right"){F=new YAHOO.util.Anim(E.content_container,{marginRight:{to:0}},0.3);F.animate();F=new YAHOO.util.Anim(E.content,{marginRight:{to:0}},0.3);F.onComplete.subscribe(function(){YAHOO.util.Dom.setAttribute(B,"src",IMGDIR_MISC+"/tab-expanded.png")});F.animate()}else{F=new YAHOO.util.Anim(E.content_container,{marginLeft:{to:0}},0.3);F.animate();F=new YAHOO.util.Anim(E.content,{marginLeft:{to:0}},0.3);F.onComplete.subscribe(function(){YAHOO.util.Dom.setAttribute(B,"src",IMGDIR_MISC+"/tab-expanded-left.png")});F.animate()}},this);C.animate()}else{YAHOO.util.Dom.setStyle(this.sidebar,"display","none");YAHOO.util.Dom.setStyle(this.sidebar_container,"width","0");if(sidebar_align=="right"){YAHOO.util.Dom.setAttribute(B,"src",IMGDIR_MISC+"/tab-expanded.png");YAHOO.util.Dom.setStyle(this.content_container,"marginRight","0");YAHOO.util.Dom.setStyle(this.content,"marginRight","0")}else{YAHOO.util.Dom.setAttribute(B,"src",IMGDIR_MISC+"/tab-expanded-left.png");YAHOO.util.Dom.setStyle(this.content_container,"marginLeft","0");YAHOO.util.Dom.setStyle(this.content,"marginLeft","0")}}this.save_collapsed("1")};vBSidebar.prototype.expand=function(){var B;var A=this.sidebar_button;if(sidebar_align=="right"){B=new YAHOO.util.Anim(this.content_container,{marginRight:{to:(0-content_container_margin)}},0.3);B.animate();B=new YAHOO.util.Anim(this.content,{marginRight:{to:content_container_margin}},0.3);B.onComplete.subscribe(function(F,C,D){YAHOO.util.Dom.setStyle(D.sidebar,"display","block");YAHOO.util.Dom.setStyle(D.sidebar_container,"width",sidebar_width+"px");var E=new YAHOO.util.Anim(D.sidebar,{opacity:{from:0,to:1}},0.3);E.onComplete.subscribe(function(){YAHOO.util.Dom.setAttribute(A,"src",IMGDIR_MISC+"/tab-collapsed.png")});E.animate()},this);B.animate()}else{B=new YAHOO.util.Anim(this.content_container,{marginLeft:{to:(0-content_container_margin)}},0.3);B.animate();B=new YAHOO.util.Anim(this.content,{marginLeft:{to:content_container_margin}},0.3);B.onComplete.subscribe(function(F,C,D){YAHOO.util.Dom.setStyle(D.sidebar,"display","block");YAHOO.util.Dom.setStyle(D.sidebar_container,"width",sidebar_width+"px");var E=new YAHOO.util.Anim(D.sidebar,{opacity:{from:0,to:1}},0.3);E.onComplete.subscribe(function(){YAHOO.util.Dom.setAttribute(A,"src",IMGDIR_MISC+"/tab-collapsed-left.png")});E.animate()},this);B.animate()}this.save_collapsed("0")};vBSidebar.prototype.save_collapsed=function(A){expires=new Date();expires.setTime(expires.getTime()+(1000*86400*365));set_cookie("vbulletin_sidebar_collapse",A,expires)};
}

/*
     FILE ARCHIVED ON 21:24:14 Feb 02, 2022 AND RETRIEVED FROM THE
     INTERNET ARCHIVE ON 10:53:09 Jun 10, 2026.
     JAVASCRIPT APPENDED BY WAYBACK MACHINE, COPYRIGHT INTERNET ARCHIVE.

     ALL OTHER CONTENT MAY ALSO BE PROTECTED BY COPYRIGHT (17 U.S.C.
     SECTION 108(a)(3)).
*/
/*
playback timings (ms):
  capture_cache.get: 22.288
  load_resource: 96.5
  PetaboxLoader3.resolve: 45.761
  PetaboxLoader3.datanode: 40.571 (2)
  loaddict: 25.498
*/