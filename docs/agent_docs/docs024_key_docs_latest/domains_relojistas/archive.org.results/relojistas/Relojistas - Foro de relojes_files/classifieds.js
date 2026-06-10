var _____WB$wombat$assign$function_____=function(name){return (globalThis._wb_wombat && globalThis._wb_wombat.local_init && globalThis._wb_wombat.local_init(name))||globalThis[name];};if(!globalThis.__WB_pmw){globalThis.__WB_pmw=function(obj){this.__WB_source=obj;return this;}}{
let window = _____WB$wombat$assign$function_____("window");
let self = _____WB$wombat$assign$function_____("self");
let document = _____WB$wombat$assign$function_____("document");
let location = _____WB$wombat$assign$function_____("location");
let top = _____WB$wombat$assign$function_____("top");
let parent = _____WB$wombat$assign$function_____("parent");
let frames = _____WB$wombat$assign$function_____("frames");
let opener = _____WB$wombat$assign$function_____("opener");

$.noConflict(true)(function($)
{var jObjects={categories:$('#sidemenu #categories'),information:$('#information'),price:$('#price'),},showDebugInfo=true,listingData='',idleTime=0,paused=false,lastRefresh=(new Date().getTime()/1000),tabMaths={},refreshTimer=null,refreshTimer2=30,idleTimer=300,refreshDelay=5,timeLeft=Classifieds.enddate,timeCheck=false;$.fn.moveListingsBits=function(){$(this).children(':first-child').hide("easeOutElastic",function(){var $child=$(this);var $parent=$child.parent();$child.appendTo($parent).show("easeInElastic",function(){setTimeout(function(){$parent.moveListingsBits();},5000);});});};$(".classifieds.animate").moveListingsBits();$('#classifiedsrotator .classifieds').children('.listingbitmini').first().show();$('#classifiedsrotator img.classifiedsarrow.right').click(function(e)
{e.preventDefault();var rotator=$('#classifiedsrotator .classifieds');rotator.children('.listingbitmini').first().next().show();rotator.children('.listingbitmini').first().animate({marginLeft:"-="+rotator.children('.listingbitmini').first().width()+"px"},function()
{$(this).appendTo(rotator).removeAttr("style");});});$('#classifiedsrotator img.classifiedsarrow.left').click(function(e)
{e.preventDefault();var rotator=$('#classifiedsrotator .classifieds');rotator.children('.listingbitmini').last().prependTo(rotator).removeAttr("style").css("margin-left","-"+rotator.children('.listingbitmini').first().show().width()+"px").animate({marginLeft:"0"},function()
{rotator.children('.listingbitmini').first().next().hide();});});$('#watchedrotator .watched').children('.listingbitmini').first().show();$('#watchedrotator img.watchedarrow.right').click(function(e)
{e.preventDefault();var rotator=$('#watchedrotator .watched');rotator.children('.listingbitmini').first().next().show();rotator.children('.listingbitmini').first().animate({marginLeft:"-="+rotator.children('.listingbitmini').first().width()+"px"},function()
{$(this).appendTo(rotator).removeAttr("style");});});$('#watchedrotator img.watchedarrow.left').click(function(e)
{e.preventDefault();var rotator=$('#watchedrotator .watched');rotator.children('.listingbitmini').last().prependTo(rotator).removeAttr("style").css("margin-left","-"+rotator.children('.listingbitmini').first().show().width()+"px").animate({marginLeft:"0"},function()
{rotator.children('.listingbitmini').first().next().hide();});});$('#featuredrotator .featured').children('.listingbitmini').first().show();$('#featuredrotator img.featuredarrow.right').click(function(e)
{e.preventDefault();var rotator=$('#featuredrotator .featured');rotator.children('.listingbitmini').first().next().show();rotator.children('.listingbitmini').first().animate({marginLeft:"-="+rotator.children('.listingbitmini').first().width()+"px"},function()
{$(this).appendTo(rotator).removeAttr("style");});});$('#featuredrotator img.featuredarrow.left').click(function(e)
{e.preventDefault();var rotator=$('#featuredrotator .featured');rotator.children('.listingbitmini').last().prependTo(rotator).removeAttr("style").css("margin-left","-"+rotator.children('.listingbitmini').first().show().width()+"px").animate({marginLeft:"0"},function()
{rotator.children('.listingbitmini').first().next().hide();});});$('div#subgallery').on('click','img',function(e)
{var self=$(this),imagesrc=self.attr('src').slice(0,-8);debugInfo(imagesrc);$('img#mainimage').attr('src',imagesrc);});$('div.listingbit').on('click','a.popupctrl',function(e)
{var self=$(this),menu=self.next();e.preventDefault();menu.fadeToggle('fast');});$('#usercp_content').on('click','a.leavefeedback',function(e)
{var self=$(this),form=$('#feedback_l'+self.data('listingitemid'));e.preventDefault();form.fadeToggle('fast');});$('span.close_window').on('click','a',function(e)
{var self=$(this),form=self.parents('form.feedback-response');e.preventDefault();form.fadeToggle('fast');});$('div.classifieds-menu-control').on('click','a.classifieds-a-control',function(e)
{var self=$(this),parent=self.closest('div.classifieds-menu-control'),menu=parent.next('ul.listingtypes-menu');e.preventDefault();menu.fadeToggle('fast',function(){parent.toggleClass('active');menu.toggleClass('active');});});$('div.blockbody.formcontrols.settings_form_border').on('change','select',function(e)
{if(paused)
{return false;}
var self=$(this),parent=self.parent('div.categoryselection'),childcount=parent.find(':selected').attr('rel'),categoryid=self.val(),grandparent=parent.parent('div.blockbody.formcontrols.settings_form_border'),parentcategoryid=parent.attr('rel');debugInfo(parentcategoryid);debugInfo(categoryid);if(!categoryid)
{return false;}
if(categoryid==-1&&parentcategoryid==0)
{return false;}
if(categoryid==-1)
{parent.fadeOut('fast',function(){$(this).remove();});return false;}
if(parentcategoryid==categoryid)
{return false;}
if(childcount)
{var extraParams={};if(typeof type=='undefined')
{type='POST';}
extraParams['do']='ajax';extraParams['action']='getcategories';extraParams['securitytoken']=SECURITYTOKEN;var urlPage='dbtclassifieds.php';extraParams['categoryid']=categoryid;paused=true;var jqxhr=$.ajax({type:type,url:urlPage,data:(SESSIONURL?SESSIONURL+'&':'')+$.param(extraParams)}).done(function(data)
{if($(data).find('error').length)
{alert($(data).find('error').text().replace(/<br ?\/?>/gi,'\n'));return false;}
parent.nextAll().fadeOut('fast',function(){$(this).remove();});parent.after($(data).find('success').text());parent.next().fadeToggle('fast');paused=false;});}});jObjects.information.on('click','div#tabs span',function(e)
{var self=$(this),tab=self.attr('id').slice(0,-4),tabcontent=jObjects.information.children('div#tabcontent');e.preventDefault();self.siblings('span').removeClass('selected');self.addClass('selected');tabcontent.children('div').hide();tabcontent.children('#'+tab+'_content').fadeToggle('fast');});jObjects.categories.on('click','img.category',function(e)
{var self=$(this),siblings=self.siblings('div.blockrow');e.preventDefault();siblings.toggleClass('selected');});function refreshBids(type)
{if(paused)
{return false;}
var extraParams={};if(typeof type=='undefined')
{type='POST';extraParams['securitytoken']=SECURITYTOKEN;}
extraParams['do']='ajax';extraParams['action']='refreshbids';var urlPage='dbtclassifieds.php';listingData=$('#price').attr('rel');extraParams['listingid']=listingData.split('listingid=')[1];debugInfo(typeof tabMaths[extraParams['listingid']]);if((parseInt(refreshDelay)>0&&((new Date().getTime()/1000)-lastRefresh)>refreshDelay)||typeof tabMaths[extraParams['tabid']]=='undefined')
{tabMaths[extraParams['listingid']]=Math.random()*99999999999999;}
extraParams['v']=tabMaths[extraParams['listingid']];debugInfo('Refreshing Bids: Listing ID# '+extraParams['listingid']);if(!listingData)
{return false;}
paused=true;value=$('input[name=bid]').val();var jqxhr=$.ajax({type:type,url:urlPage,data:(SESSIONURL?SESSIONURL+'&':'')+$.param(extraParams)}).done(function(data)
{if($(data).find('error').length)
{alert($(data).find('error').text().replace(/<br ?\/?>/gi,'\n'));return false;}
lastRefresh=(new Date().getTime()/1000);$('#price').html($(data).find('success').text()).addClass('refreshed').delay(1000).queue(function(next){$(this).removeClass('refreshed');next();});$('input[name=bid]').val(value);paused=false;listingData='';clearInterval(refreshTimer);refreshTimer=null;_startTimer();});}
jObjects.price.on('click','a#nextbid',function(e)
{var self=$(this),value=self.html();e.preventDefault();$('input[name=bid]').val(value.replace(/[^\d.,]/g,''));});$('#price').on('click','#bidslist',function(e)
{var self=$(this);e.preventDefault();var extraParams={};extraParams['securitytoken']=SECURITYTOKEN;extraParams['do']='ajax';extraParams['action']='offerslist';var urlPage='dbtclassifieds.php';listingData=$('#price').attr('rel');extraParams['listingid']=listingData.split('listingid=')[1];debugInfo(typeof tabMaths[extraParams['listingid']]);if((parseInt(refreshDelay)>0&&((new Date().getTime()/1000)-lastRefresh)>refreshDelay)||typeof tabMaths[extraParams['tabid']]=='undefined')
{tabMaths[extraParams['listingid']]=Math.random()*99999999999999;}
extraParams['v']=tabMaths[extraParams['listingid']];debugInfo('Loading Offers: Listing ID# '+extraParams['listingid']);if(!listingData)
{return false;}
paused=true;var jqxhr=$.ajax({type:'POST',url:urlPage,data:(SESSIONURL?SESSIONURL+'&':'')+$.param(extraParams)}).done(function(data)
{if($(data).find('error').length)
{alert($(data).find('error').text().replace(/<br ?\/?>/gi,'\n'));return false;}
$('#window_content').html($(data).find('success').text());$('#window_container').fadeIn('fast');listingData='';});});$('#price').on('click','#bidbutton a',function(e)
{if(ClassifiedsBids)
{var self=$(this),listingid=self.attr('rel'),value=$('input[name=bid]').val(),length=$('#address').children('option').length;e.preventDefault();if(!length)
{return false;}
var extraParams={};extraParams['securitytoken']=SECURITYTOKEN;extraParams['postageid']=$('select[name=postage]').val();extraParams['addressid']=$('select[name=address]').val();value=value.replace(new RegExp('[\D\\'+thousandSep+']','g'),"");value=value.replace(new RegExp('[\\'+decimalSep+']','g'),".");value=parseFloat(value);extraParams['do']='ajax';extraParams['action']='biditem';extraParams['value']=value;var urlPage='dbtclassifieds.php';extraParams['listingid']=listingid;extraParams['v']=Math.random()*99999999999999;debugInfo('Bidding on Item: '+extraParams['listingid']+' price: '+extraParams['value']);if(!listingid)
{return false;}
var jqxhr=$.ajax({type:'POST',url:urlPage,data:(SESSIONURL?SESSIONURL+'&':'')+$.param(extraParams)}).done(function(data)
{if($(data).find('error').length)
{alert($(data).find('error').text().replace(/<br ?\/?>/gi,'\n'));return false;}
$('#price').html($(data).find('success').text());});}});$('#price').on('click','#watchbutton a',function(e)
{var self=$(this),listingid=self.attr('rel'),phrase=self.text();e.preventDefault();var extraParams={};extraParams['securitytoken']=SECURITYTOKEN;extraParams['do']='ajax';extraParams['action']='watchitem';var urlPage='dbtclassifieds.php';extraParams['listingid']=listingid;extraParams['v']=Math.random()*99999999999999;debugInfo('Watching Item: '+extraParams['listingid']);if(!listingid)
{return false;}
var jqxhr=$.ajax({type:'POST',url:urlPage,data:(SESSIONURL?SESSIONURL+'&':'')+$.param(extraParams)}).done(function(data)
{if($(data).find('error').length)
{alert($(data).find('error').text().replace(/<br ?\/?>/gi,'\n'));return false;}
$('#price #watchbutton a').text($(data).find('success').text());});});$('#mainbox').on('click','a#savesearch',function(e)
{var self=$(this),searchid=self.attr('rel'),phrase=self.text();e.preventDefault();var extraParams={};extraParams['securitytoken']=SECURITYTOKEN;extraParams['do']='ajax';extraParams['action']='savesearch';var urlPage='dbtclassifieds.php';extraParams['searchid']=searchid;extraParams['v']=Math.random()*99999999999999;debugInfo('Saving Search: '+extraParams['searchid']);if(!searchid)
{return false;}
var jqxhr=$.ajax({type:'POST',url:urlPage,data:(SESSIONURL?SESSIONURL+'&':'')+$.param(extraParams)}).done(function(data)
{if($(data).find('error').length)
{alert($(data).find('error').text().replace(/<br ?\/?>/gi,'\n'));return false;}
$('#mainbox a#savesearch').text($(data).find('success').text());});});$('div.blockbody.formcontrols.settings_form_border.ajaxstates').on('change','select#country',function(e)
{if(paused)
{return false;}
var self=$(this),parent=self.parent('div.blockrow'),countryid=self.val(),grandparent=parent.parent('div.section'),state=grandparent.find('select#state');debugInfo(countryid);if(!countryid)
{return false;}
if(countryid=='-')
{state.prop('disabled',true).val('-');return true;}
if(countryid)
{paused=true;state.prop('disabled',false);var extraParams={};extraParams['securitytoken']=SECURITYTOKEN;extraParams['do']='ajax';extraParams['action']='getstates';var urlPage='dbtclassifieds.php';extraParams['countryid']=countryid;extraParams['v']=Math.random()*99999999999999;var jqxhr=$.ajax({type:'POST',url:urlPage,data:(SESSIONURL?SESSIONURL+'&':'')+$.param(extraParams)}).done(function(data)
{if($(data).find('error').length)
{alert($(data).find('error').text().replace(/<br ?\/?>/gi,'\n'));return false;}
$('select#state').html($(data).find('success').text());paused=false;});}});$('div.blockbody.formcontrols.settings_form_border.submitissue').on('change','select',function(e)
{if(paused)
{return false;}
var self=$(this),categoryid=self.val();debugInfo(categoryid);if(!categoryid)
{return false;}
var extraParams={};if(typeof type=='undefined')
{type='POST';}
extraParams['do']='ajax';extraParams['action']='getissuecategories';extraParams['securitytoken']=SECURITYTOKEN;var urlPage='dbtclassifieds.php';extraParams['categoryid']=categoryid;paused=true;var jqxhr=$.ajax({type:type,url:urlPage,data:(SESSIONURL?SESSIONURL+'&':'')+$.param(extraParams)}).done(function(data)
{if($(data).find('error').length)
{alert($(data).find('error').text().replace(/<br ?\/?>/gi,'\n'));return false;}
$('div#issueselector').html($(data).find('success').text());paused=false;});});$('div.blockbody.formcontrols.settings_form_border').on('change','input[name=checkenddate]',function(e)
{var self=$(this),parent=self.parent();$('input[id^="listingtype_edate"]').prop("disabled",!$('input[id^="listingtype_edate"]').prop("disabled"));$('input[id^="listingtype_date"]').prop("disabled",!$('input[id^="listingtype_date"]').prop("disabled"));});$('#window_content').on('click','#close_window a',function(e)
{var self=$(this);e.preventDefault();$('#window_container').fadeOut('fast');});$('#price').on('click','a#refreshbids',function(e)
{var self=$(this);e.preventDefault();refreshBids('GET');});$('#bidform').submit(function(e){if(ClassifiedsBids)
{e.preventDefault();}});function classifiedsCountdown()
{var currentDate=Math.floor($.now()/1000);iseconds=timeLeft-currentDate;if(iseconds<Classifieds.timeleft)
{if(timeCheck==false)
{debugInfo('We have run past the end');$('#bigclock').fadeIn('fast');$('#timeleft div.middlecol span[1]').addClass('lasthour');timeCheck=true;}}
if(iseconds>0)
{var seconds=$('.timeleft span.seconds').text(),minutes=$('.timeleft span.minutes').text(),hours=$('.timeleft span.hours').text(),days=$('.timeleft span.days').text();idays=Math.floor(iseconds/86400);iseconds-=idays*86400;ihours=Math.floor(iseconds/3600);iseconds-=ihours*3600;iminutes=Math.floor(iseconds/60);iseconds-=iminutes*60;if(Classifieds.format==1)
{idays=(String(idays).length>=2)?idays:"0"+idays;ihours=(String(ihours).length>=2)?ihours:"0"+ihours;iminutes=(String(iminutes).length>=2)?iminutes:"0"+iminutes;iseconds=(String(iseconds).length>=2)?iseconds:"0"+iseconds;}
$('span.days').each(function(){$(this).text(idays);});$('span.hours').each(function(){$(this).text(ihours);});$('span.minutes').each(function(){$(this).text(iminutes);});$('span.seconds').each(function(){$(this).text(iseconds);});}
else
{if(Classifieds.ended==0)
{window.location.reload();}}}
function _startTimer()
{refreshTimer=setInterval(function()
{if(idleTime>=idleTimer&&idleTimer>0)
{debugInfo('Setting idle...');clearInterval(refreshTimer);refreshTimer=null;$('#unidle').fadeIn('fast',function()
{});return false;}
idleTime+=parseInt(refreshTimer2);if(Classifieds.ended==0)
{refreshBids('GET');}},refreshTimer2*1000);};function animateBox(title,description,onoff)
{var box=$('#ajaxprogress');if(onoff)
{box.css('display','inline-block');box.css('opacity',0);}
if(title)
{$('#progresstitle').html(title);}
if(description)
{$('#progresscontent').html(description);}
box.animate({opacity:(onoff?0.8:0)},{duration:700,complete:function()
{$(this).fadeOut('fast');}});};function debugInfo(message,debugLevel)
{if(!showDebugInfo)
{return false;}
var d=new Date(),timeStamp='['+d.getHours()+':'+d.getMinutes()+':'+d.getSeconds()+'] ';switch(debugLevel)
{case'error':console.error(timeStamp+message);break;case'warning':console.warn(timeStamp+message);break;case'log':default:console.log(timeStamp+message);break;}};classifiedsCountdown();interval=setInterval(classifiedsCountdown,1000);$('#price').on('keyup','div input[name="bid"]',function(){$(this).val($(this).val().replace(/[^\d.,]/g,''));});});
}

/*
     FILE ARCHIVED ON 18:56:22 Feb 17, 2022 AND RETRIEVED FROM THE
     INTERNET ARCHIVE ON 10:53:08 Jun 10, 2026.
     JAVASCRIPT APPENDED BY WAYBACK MACHINE, COPYRIGHT INTERNET ARCHIVE.

     ALL OTHER CONTENT MAY ALSO BE PROTECTED BY COPYRIGHT (17 U.S.C.
     SECTION 108(a)(3)).
*/
/*
playback timings (ms):
  capture_cache.get: 1.52
  load_resource: 292.049
  PetaboxLoader3.resolve: 78.032
  PetaboxLoader3.datanode: 212.242
*/