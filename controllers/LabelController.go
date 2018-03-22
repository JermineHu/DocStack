package controllers

import (
	"math"

	"github.com/JermineHu/DocStack/conf"
	"github.com/JermineHu/DocStack/models"
	"github.com/JermineHu/DocStack/utils"
	"github.com/astaxie/beego"
)

type LabelController struct {
	BaseController
}

func (this *LabelController) Prepare() {
	this.BaseController.Prepare()

	//如果没有开启你们访问则跳转到登录
	if !this.EnableAnonymous && this.Member == nil {
		this.Redirect(beego.URLFor("AccountController.Login"), 302)
		return
	}
}

//查看包含标签的文档列表.
func (this *LabelController) Index() {
	this.TplName = "label/index.html"
	this.Data["IsLabel"] = true

	labelName := this.Ctx.Input.Param(":key")
	this.Data["Keyword"] = labelName
	pageIndex, _ := this.GetInt("page", 1)
	if labelName == "" {
		this.Abort("404")
	}
	//_, err := models.NewLabel().FindFirst("label_name", labelName)
	//
	//if err != nil {
	//	if err == orm.ErrNoRows {
	//		this.Abort("404")
	//	} else {
	//		beego.Error(err)
	//		this.Abort("500")
	//	}
	//}

	pageSize := 24
	member_id := 0
	if this.Member != nil {
		member_id = this.Member.MemberId
	}
	search_result, totalCount, err := models.NewBook().FindForLabelToPager(labelName, pageIndex, pageSize, member_id)

	if err != nil {
		beego.Error(err)
		return
	}
	if totalCount > 0 {
		html := utils.NewPaginations(conf.RollPage, totalCount, pageSize, pageIndex, beego.URLFor("LabelController.Index", ":key", labelName), "")
		this.Data["PageHtml"] = html
	} else {
		this.Data["PageHtml"] = ""
	}
	this.Data["Lists"] = search_result

	this.Data["LabelName"] = labelName

	this.GetSeoByPage("label_list", map[string]string{
		"title":       "[标签]" + labelName + " - " + this.Sitename,
		"keywords":    "标签," + labelName,
		"description": this.Sitename + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})

}

func (this *LabelController) List() {
	this.Data["IsLabel"] = true
	this.TplName = "label/list.html"

	pageIndex, _ := this.GetInt("page", 1)
	pageSize := 200

	labels, totalCount, err := models.NewLabel().FindToPager(pageIndex, pageSize)

	if err != nil {
		this.ShowErrorPage(50001, err.Error())
	}
	if totalCount > 0 {
		html := utils.NewPaginations(conf.RollPage, totalCount, pageSize, pageIndex, beego.URLFor("LabelController.List"), "")
		this.Data["PageHtml"] = html
	} else {
		this.Data["PageHtml"] = ""
	}
	this.Data["TotalPages"] = int(math.Ceil(float64(totalCount) / float64(pageSize)))

	this.Data["Labels"] = labels

	this.GetSeoByPage("label_list", map[string]string{
		"title":       "标签 - " + this.Sitename,
		"keywords":    "标签",
		"description": this.Sitename + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})

}
