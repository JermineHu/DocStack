

function showError($msg,$id) {
    if(!$id){
        $id = "#form-error-message"
    }
    $($id).addClass("error-message").removeClass("success-message").text($msg);
    return false;
}

function showSuccess($msg,$id) {
    if(!$id){
        $id = "#form-error-message"
    }
    $($id).addClass("success-message").removeClass("error-message").text($msg);
    return true;
}

Date.prototype.format = function(fmt) {
    var o = {
        "M+" : this.getMonth()+1,                 //月份
        "d+" : this.getDate(),                    //日
        "h+" : this.getHours(),                   //小时
        "m+" : this.getMinutes(),                 //分
        "s+" : this.getSeconds(),                 //秒
        "q+" : Math.floor((this.getMonth()+3)/3), //季度
        "S"  : this.getMilliseconds()             //毫秒
    };
    if(/(y+)/.test(fmt)) {
        fmt=fmt.replace(RegExp.$1, (this.getFullYear()+"").substr(4 - RegExp.$1.length));
    }
    for(var k in o) {
        if(new RegExp("("+ k +")").test(fmt)){
            fmt = fmt.replace(RegExp.$1, (RegExp.$1.length==1) ? (o[k]) : (("00"+ o[k]).substr((""+ o[k]).length)));
        }
    }
    return fmt;
};

function formatBytes($size) {
    var $units = [" B", " KB", " MB", " GB", " TB"];

    for ($i = 0; $size >= 1024 && $i < 4; $i++) $size /= 1024;

    return $size.toFixed(2) + $units[$i];
}

//将json字符串解析成json对象
function parseJson(obj) {
    if (typeof obj=="string"){
        return JSON.parse(obj)
    }
    return obj
}

function alertTips(cls,msg,timeout,url) {
    if(!msg) return false;
    var title="";
    if (cls=="error" || cls=="fail" || cls=="danger"){
        title='错误';
        cls="error"
    }else{
        title='成功';
        cls="success"

    }
    $.Toast(title,msg,cls, {
        stack: true,
        has_icon:true,
        has_close_btn:true,
        fullscreen:false,
        timeout:parseInt(timeout),
        sticky:false,
        has_progress:true,
        rtl:false,
        // position_class:'toast-top-right',
    });
    setTimeout(function () {
        if(url){
            location.href=url
        }
    },parseInt(timeout));
}

$(function () {


    //文档项目评分
    if($("body").attr("id")=="bookstack-intro"){
        var stars=$(".cursor-pointer .fa")
        $(".cursor-pointer .fa").hover(function () {
            if(!$(this).parent().hasClass("cursor-pointer")) return false;
            var val=parseInt($(this).attr("data-score"));
            $(".score .text-muted").text($(this).attr("data-tips"));
            for(var i=0;i<5;i++){
                if(i<val){
                    $(stars[i]).addClass("fa-star").removeClass("fa-star-o");
                }else{
                    $(stars[i]).addClass("fa-star-o").removeClass("fa-star");
                }
            }
            $(".cursor-pointer").attr("data-val",val);
            $(this).addClass("fa-star").removeClass("fa-star-o")
        });
        $(".cursor-pointer").hover(function () {
        },function () {
            if($(this).hasClass("cursor-pointer")){
                $(".cursor-pointer .fa").addClass("fa-star-o").removeClass("fa-star");
            }
        });
        $(".cursor-pointer .fa").click(function () {
            var cur=$(this).parent();
            if(cur.hasClass("cursor-pointer")){
                $.get(cur.attr("data-url"),{score:cur.attr("data-val")},function (ret) {
                    ret=parseJson(ret);
                    if(ret.errcode==0){
                        alertTips("success",ret.message,3000,"");
                        cur.removeClass("cursor-pointer").attr("data-toggle","").attr("data-original-title","");
                        $(".tooltip").remove();
                    }else{
                        alertTips("error",ret.message,3000,"");
                    }
                });
            }
        });
    }

    //针对表单项
    $(".change-update").change(function () {
        var _this=$(this),_url=_this.attr("data-url"),field=_this.attr("name"),method=_this.attr("data-method");
        if(method=="post"){
            $.post(_url,{field:field,value:$(this).val().trim()},function (res) {
                if(res.errcode==0){
                    alertTips("success",res.message,3000,"");
                }else{
                    $(this).val($(this)[0].defaultValue);//恢复值
                    alertTips("danger",res.message,3000,"");
                }
            })
        }else{
            $.get(_url,{field:field,value:$(this).val().trim()},function (res) {
                if(res.errcode==0){
                    alertTips("success",res.message,3000,"");
                }else{
                    $(this).val($(this)[0].defaultValue);//恢复值
                    alertTips("error",res.message,3000,"");
                }
            })
        }
    });

    //收藏获取取消收藏
    $(".ajax-star").click(function (e) {
        e.preventDefault()
        var _url=$(this).attr("href"),_this=$(this);
        $.get(_url,function (ret) {
            ret=parseJson(ret);
            if(ret.errcode==0){//操作成功
                alertTips("success",ret.message,3000,"");
                if (ret.data.IsCancel){//取消收藏
                    _this.find("span").text("收藏");
                    _this.find(".fa-heart").addClass("fa-heart-o").removeClass("fa-heart");
                }else{//添加收藏
                    _this.find("span").text("已收藏");
                    _this.find(".fa-heart-o").addClass("fa-heart").removeClass("fa-heart-o");
                }
            }else{
                alertTips("danger",ret.message,3000,"");
            }
        })
    });

    //ajax-get
    $(document).on("click",".ajax-get",function (e) {
        e.preventDefault();
        if($(this).hasClass("confirm") && !confirm("您确定要执行该操作吗？")){
            return true;
        }
        var _url=$(this).attr("href"),_this=$(this);
        $.get(_url,function (ret) {
            ret=parseJson(ret);
            if(ret.errcode==0){//操作成功
                alertTips("success",ret.message,3000,location.href);
            }else{
                alertTips("danger",ret.message,3000,"");
            }
        })
    });

    //ajax-form
    $(".ajax-form [type=submit]").click(function(e){
        e.preventDefault();
        var _this=$(this),form=$(this).parents("form"),method=form.attr("method"),action=form.attr("action"),data=form.serialize(),_url=form.attr("data-url");
        var require=form.find("[required=required]"),l=require.length;
        $.each(require, function() {
            if (!$(this).val()){
                $(this).focus();
                return false;
            }else{
                l--;
            }
        });
        if (!_url || _url==undefined){
            _url=location.href;
        }
        if (l>0) return false;
        if (method=="post") {
            if (form.attr("enctype")=="multipart/form-data"){
                form.attr("target","notarget");
                form.submit();
            }else{
                $.post(action,data,function(ret){
                    ret=parseJson(ret);
                    if (ret.errcode==0) {
                        alertTips("success",ret.message,2000,_url);
                    } else{
                        alertTips("error",ret.message,3000,"");
                    }
                });
                _this.removeClass("disabled");
            }

        } else{
            $.get(action,data,function(ret){
                ret=parseJson(ret);
                if (ret.errcode==0) {
                    alertTips("success",ret.message,2000,_url);
                } else{
                    alertTips("error",ret.message,3000,"");
                }
            });
        }
    });


    //文档下载
    $(".btn-filedown").click(function (e) {
        e.preventDefault();
        $.get($(this).attr("href"),function (res) {
           var obj=parseJson(res);
           if (obj.errcode==0){
               location.href=obj.data.url;
               $(".modal").modal("hide");
           }else{
                alertTips("danger",obj.message,3000,"");
           }
        });
    });

});