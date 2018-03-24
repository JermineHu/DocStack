package controllers

import (
	"container/list"
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"image/png"

	"bytes"

	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/JermineHu/DocStack/commands"
	"github.com/JermineHu/DocStack/conf"
	"github.com/JermineHu/DocStack/models"
	"github.com/JermineHu/DocStack/utils"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

//DocumentController struct.
type DocumentController struct {
	BaseController
}

//判断用户是否可以阅读文档.
func isReadable(identify, token string, this *DocumentController) *models.BookResult {
	book, err := models.NewBook().FindByFieldFirst("identify", identify)
	if err != nil {
		beego.Error(err)
		this.Abort("404")
	}

	//如果文档是私有的
	if book.PrivatelyOwned == 1 && !this.Member.IsAdministrator() {

		is_ok := false

		if this.Member != nil {
			_, err := models.NewRelationship().FindForRoleId(book.BookId, this.Member.MemberId)
			if err == nil {
				is_ok = true
			}
		}
		if book.PrivateToken != "" && !is_ok {
			//如果有访问的Token，并且该项目设置了访问Token，并且和用户提供的相匹配，则记录到Session中.
			//如果用户未提供Token且用户登录了，则判断用户是否参与了该项目.
			//如果用户未登录，则从Session中读取Token.
			if token != "" && strings.EqualFold(token, book.PrivateToken) {
				this.SetSession(identify, token)

			} else if token, ok := this.GetSession(identify).(string); !ok || !strings.EqualFold(token, book.PrivateToken) {
				this.Abort("403")
			}
		} else if !is_ok {
			this.Abort("403")
		}

	}
	bookResult := book.ToBookResult()

	if this.Member != nil {

		rel, err := models.NewRelationship().FindByBookIdAndMemberId(bookResult.BookId, this.Member.MemberId)

		if err == nil {
			bookResult.MemberId = rel.MemberId
			bookResult.RoleId = rel.RoleId
			bookResult.RelationshipId = rel.RelationshipId
		}
	}
	//判断是否需要显示评论框
	if bookResult.CommentStatus == "closed" {
		bookResult.IsDisplayComment = false
	} else if bookResult.CommentStatus == "open" {
		bookResult.IsDisplayComment = true
	} else if bookResult.CommentStatus == "group_only" {
		bookResult.IsDisplayComment = bookResult.RelationshipId > 0
	} else if bookResult.CommentStatus == "registered_only" {
		bookResult.IsDisplayComment = true
	}

	return bookResult
}

//文档首页.
func (this *DocumentController) Index() {
	identify := this.Ctx.Input.Param(":key")
	token := this.GetString("token")
	if identify == "" {
		this.Abort("404")
	}
	tab := strings.ToLower(this.GetString("tab"))

	//如果没有开启匿名访问则跳转到登录
	if !this.EnableAnonymous && this.Member == nil {
		this.Redirect(beego.URLFor("AccountController.Login"), 302)
		return
	}
	bookResult := isReadable(identify, token, this)
	if bookResult.BookId == 0 { //没有阅读权限
		this.Redirect(beego.URLFor("HomeController.Index"), 302)
		return
	}
	this.TplName = "document/intro.html"
	this.Data["Book"] = bookResult

	switch tab {
	case "comment", "score":
	default:
		tab = "default"
	}
	this.Data["Qrcode"] = new(models.Member).GetQrcodeByUid(bookResult.MemberId)
	this.Data["MyScore"] = new(models.Score).BookScoreByUid(this.Member.MemberId, bookResult.BookId)
	this.Data["Tab"] = tab
	//当前默认展示30条评论
	this.Data["Comments"], _ = new(models.Comments).BookComments(1, 30, bookResult.BookId)
	this.Data["Menu"], _ = new(models.Document).GetMenuTop(bookResult.BookId)
	this.GetSeoByPage("book_info", map[string]string{
		"title":       bookResult.BookName,
		"keywords":    bookResult.Label,
		"description": bookResult.Description,
	})
}

//阅读文档.
func (this *DocumentController) Read() {
	identify := this.Ctx.Input.Param(":key")
	token := this.GetString("token")
	id := this.GetString(":id")

	if identify == "" || id == "" {
		this.Abort("404")
	}

	//如果没有开启你们匿名则跳转到登录
	if !this.EnableAnonymous && this.Member == nil {
		this.Redirect(beego.URLFor("AccountController.Login"), 302)
		return
	}

	bookResult := isReadable(identify, token, this)

	this.TplName = "document/" + bookResult.Theme + "_read.html"

	doc := models.NewDocument()

	if doc_id, err := strconv.Atoi(id); err == nil {
		doc, err = doc.Find(doc_id) //文档id
		if err != nil {
			beego.Error(err)
			this.Abort("500")
		}
	} else {
		//此处的id是字符串，标识文档标识，根据文档标识和文档所属的书的id作为key去查询
		doc, err = doc.FindByBookIdAndDocIdentify(bookResult.BookId, id) //文档标识
		if err != nil {
			beego.Error(err, bookResult)
			this.Abort("500")
		}
	}

	if doc.BookId != bookResult.BookId {
		this.Abort("403")
	}
	attach, err := models.NewAttachment().FindListByDocumentId(doc.DocumentId)
	if err == nil {
		doc.AttachList = attach
	}

	cdnimg := beego.AppConfig.String("cdnimg")
	if doc.Release != "" && cdnimg != "" {
		query, err := goquery.NewDocumentFromReader(bytes.NewBufferString(doc.Release))
		if err != nil {
			beego.Error(err)
		} else {
			query.Find("img").Each(func(i int, contentSelection *goquery.Selection) {
				if src, ok := contentSelection.Attr("src"); ok && strings.HasPrefix(src, "/uploads/") {
					contentSelection.SetAttr("src", utils.JoinURI(cdnimg, src))
				}
			})
			html, err := query.Html()
			if err != nil {
				beego.Error(err)
			} else {
				doc.Release = html
			}
		}

	}

	//文档阅读人次+1
	if err := models.SetIncreAndDecre("md_documents", "vcnt",
		fmt.Sprintf("document_id=%v", doc.DocumentId),
		true, 1,
	); err != nil {
		beego.Error(err.Error())
	}
	//项目阅读人次+1
	if err := models.SetIncreAndDecre("md_books", "vcnt",
		fmt.Sprintf("book_id=%v", doc.BookId),
		true, 1,
	); err != nil {
		beego.Error(err.Error())
	}

	//SEO
	this.GetSeoByPage("book_read", map[string]string{
		"title":       doc.DocumentName + " - " + bookResult.BookName,
		"keywords":    bookResult.Label,
		"description": bookResult.Description,
	})

	if this.IsAjax() {
		var data struct {
			DocTitle string `json:"doc_title"`
			Body     string `json:"body"`
			Title    string `json:"title"`
		}
		data.DocTitle = doc.DocumentName
		data.Body = doc.Release
		//data.Body = doc.Markdown
		data.Title = this.Data["SeoTitle"].(string)

		this.JsonResult(0, "ok", data)
	}

	tree, err := models.NewDocument().CreateDocumentTreeForHtml(bookResult.BookId, doc.DocumentId)

	if err != nil {
		beego.Error(err)
		this.Abort("500")
	}

	this.Data["Model"] = bookResult
	this.Data["Book"] = bookResult //文档下载需要用到Book变量
	this.Data["Result"] = template.HTML(tree)
	this.Data["Title"] = doc.DocumentName
	this.Data["Content"] = template.HTML(doc.Release)

}

//编辑文档.
func (this *DocumentController) Edit() {

	identify := this.Ctx.Input.Param(":key")
	if identify == "" {
		this.Abort("404")
	}

	bookResult := models.NewBookResult()
	var err error
	//如果是超级管理者，则不判断权限
	if this.Member.IsAdministrator() {
		book, err := models.NewBook().FindByFieldFirst("identify", identify)
		if err != nil {
			this.JsonResult(6002, "项目不存在或权限不足")
		}
		bookResult = book.ToBookResult()

	} else {
		bookResult, err = models.NewBookResult().FindByIdentify(identify, this.Member.MemberId)

		if err != nil {
			beego.Error("DocumentController.Edit => ", err)

			this.Abort("403")
		}
		if bookResult.RoleId == conf.BookObserver {

			this.JsonResult(6002, "项目不存在或权限不足")
		}
	}

	//根据不同编辑器类型加载编辑器
	if bookResult.Editor == "markdown" {
		this.TplName = "document/markdown_edit_template.html"
	} else if bookResult.Editor == "html" {
		//this.TplName = "document/html_edit_template.html"
		this.TplName = "document/markdown_edit_template.html"
	} else {
		this.TplName = "document/" + bookResult.Editor + "_edit_template.html"
	}

	this.Data["Model"] = bookResult

	r, _ := json.Marshal(bookResult)

	this.Data["ModelResult"] = template.JS(string(r))

	this.Data["Result"] = template.JS("[]")

	trees, err := models.NewDocument().FindDocumentTree(bookResult.BookId, true)
	if err != nil {
		beego.Error("FindDocumentTree => ", err)
	} else {
		if len(trees) > 0 {
			if jtree, err := json.Marshal(trees); err == nil {
				this.Data["Result"] = template.JS(string(jtree))
			}
		} else {
			this.Data["Result"] = template.JS("[]")
		}
	}
	this.Data["BaiDuMapKey"] = beego.AppConfig.DefaultString("baidumapkey", "")

}

//创建一个文档.
func (this *DocumentController) Create() {
	identify := this.GetString("identify")
	doc_identify := this.GetString("doc_identify")
	doc_name := this.GetString("doc_name")
	parent_id, _ := this.GetInt("parent_id", 0)
	doc_id, _ := this.GetInt("doc_id", 0)
	bookIdentify := strings.TrimSpace(this.GetString(":key"))
	o := orm.NewOrm()

	if identify == "" {
		this.JsonResult(6001, "参数错误")
	}
	if doc_name == "" {
		this.JsonResult(6004, "文档名称不能为空")
	}
	if doc_identify != "" {

		if ok, err := regexp.MatchString(`^[a-zA-Z0-9_\-\.]*$`, doc_identify); !ok || err != nil {
			this.JsonResult(6003, "文档标识只能是数字、字母，以及“-”、“_”和“.”等字符，并且不能是纯数字")
		}
		if num, _ := strconv.Atoi(doc_identify); doc_identify == "0" || strconv.Itoa(num) == doc_identify { //不能是纯数字
			this.JsonResult(6005, "文档标识只能是数字、字母，以及“-”、“_”和“.”等字符，并且不能是纯数字")
		}

		if bookIdentify == "" {
			this.JsonResult(1, "文档项目参数不正确")
		}

		var book models.Book
		o.QueryTable("md_books").Filter("Identify", bookIdentify).One(&book, "BookId")
		if book.BookId == 0 {
			this.JsonResult(1, "文档项目未创建")
		}

		d, _ := models.NewDocument().FindByBookIdAndDocIdentify(book.BookId, doc_identify)
		if d.DocumentId > 0 && d.DocumentId != doc_id {
			this.JsonResult(6006, "文档标识已被使用")
		}
	} else {
		doc_identify = fmt.Sprintf("date-%v", time.Now().Format("2006.01.02.15.04.05"))
	}
	book_id := 0
	//如果是超级管理员则不判断权限
	if this.Member.IsAdministrator() {
		book, err := models.NewBook().FindByFieldFirst("identify", identify)
		if err != nil {
			beego.Error(err)
			this.JsonResult(6002, "项目不存在或权限不足")
		}
		book_id = book.BookId
	} else {
		bookResult, err := models.NewBookResult().FindByIdentify(identify, this.Member.MemberId)

		if err != nil || bookResult.RoleId == conf.BookObserver {
			beego.Error("FindByIdentify => ", err)
			this.JsonResult(6002, "项目不存在或权限不足")
		}
		book_id = bookResult.BookId
	}
	if parent_id > 0 {
		doc, err := models.NewDocument().Find(parent_id)
		if err != nil || doc.BookId != book_id {
			this.JsonResult(6003, "父分类不存在")
		}
	}

	document, _ := models.NewDocument().Find(doc_id)

	document.MemberId = this.Member.MemberId
	document.BookId = book_id
	if doc_identify != "" {
		document.Identify = doc_identify
	}
	document.Version = time.Now().Unix()
	document.DocumentName = doc_name
	document.ParentId = parent_id

	if doc_id, err := document.InsertOrUpdate(); err != nil {
		beego.Error("InsertOrUpdate => ", err)
		this.JsonResult(6005, "保存失败")
	} else {
		ModelStore := new(models.DocumentStore)
		if ModelStore.GetFiledById(doc_id, "markdown") == "" {
			//因为创建和更新文档基本信息都调用的这个接口，先判断markdown是否有内容，没有内容则添加默认内容
			if err := ModelStore.InsertOrUpdate(models.DocumentStore{DocumentId: int(doc_id), Markdown: "[TOC]\n\r\n\r"}); err != nil {
				beego.Error(err)
			}
		}
		this.JsonResult(0, "ok", document)
	}
}

//批量创建文档
func (this *DocumentController) CreateMulti() {
	book_id, _ := this.GetInt("book_id")
	if this.Member.MemberId > 0 && book_id > 0 {
		var book models.Book
		o := orm.NewOrm()
		o.QueryTable("md_books").Filter("book_id", book_id).Filter("member_id", this.Member.MemberId).One(&book, "book_id")
		if book.BookId > 0 {
			content := this.GetString("content")
			if slice := strings.Split(content, "\n"); len(slice) > 0 {
				ModelStore := new(models.DocumentStore)
				for _, row := range slice {
					if chapter := strings.Split(strings.TrimSpace(row), " "); len(chapter) > 1 {
						if ok, err := regexp.MatchString(`^[a-zA-Z0-9_\-\.]*$`, chapter[0]); ok && err == nil {
							i, _ := strconv.Atoi(chapter[0])
							if chapter[0] != "0" && strconv.Itoa(i) != chapter[0] { //不为纯数字
								doc := models.Document{
									DocumentName: strings.Join(chapter[1:], " "),
									Identify:     chapter[0],
									BookId:       book_id,
									//Markdown:     "[TOC]\n\r",
									MemberId: this.Member.MemberId,
								}
								if doc_id, err := doc.InsertOrUpdate(); err == nil {
									if err := ModelStore.InsertOrUpdate(models.DocumentStore{DocumentId: int(doc_id), Markdown: "[TOC]\n\r\n\r"}); err != nil {
										beego.Error(err.Error())
									}
								} else {
									beego.Error(err)
								}
							}

						}
					}
				}
			}
		}
		this.JsonResult(0, "添加成功")
	} else {
		this.JsonResult(1, "操作失败：只有项目创始人才能批量添加")
	}
}

//上传附件或图片.
func (this *DocumentController) Upload() {

	identify := this.GetString("identify")
	doc_id, _ := this.GetInt("doc_id")
	is_attach := true

	if identify == "" {
		this.JsonResult(6001, "参数错误")
	}

	name := "editormd-file-file"

	file, moreFile, err := this.GetFile(name)
	if err == http.ErrMissingFile {
		name = "editormd-image-file"
		file, moreFile, err = this.GetFile(name)
		if err == http.ErrMissingFile {
			this.JsonResult(6003, "没有发现需要上传的文件")
		}
	}
	if err != nil {
		this.JsonResult(6002, err.Error())
	}

	defer file.Close()

	ext := filepath.Ext(moreFile.Filename)

	if ext == "" {
		this.JsonResult(6003, "无法解析文件的格式")
	}

	if !conf.IsAllowUploadFileExt(ext) {
		this.JsonResult(6004, "不允许的文件类型")
	}
	book_id := 0
	//如果是超级管理员，则不判断权限
	if this.Member.IsAdministrator() {
		book, err := models.NewBook().FindByFieldFirst("identify", identify)
		if err != nil {
			this.JsonResult(6006, "文档不存在或权限不足")
		}
		book_id = book.BookId

	} else {
		book, err := models.NewBookResult().FindByIdentify(identify, this.Member.MemberId)
		if err != nil {
			beego.Error("DocumentController.Edit => ", err)
			if err == orm.ErrNoRows {
				this.JsonResult(6006, "权限不足")
			}
			this.JsonResult(6001, err.Error())
		}
		//如果没有编辑权限
		if book.RoleId != conf.BookEditor && book.RoleId != conf.BookAdmin && book.RoleId != conf.BookFounder {
			this.JsonResult(6006, "权限不足")
		}
		book_id = book.BookId
	}

	if doc_id > 0 {
		doc, err := models.NewDocument().Find(doc_id)
		if err != nil {
			this.JsonResult(6007, "文档不存在")
		}
		if doc.BookId != book_id {
			this.JsonResult(6008, "文档不属于指定的项目")
		}
	}

	fileName := strconv.FormatInt(time.Now().UnixNano(), 16)

	filePath := filepath.Join(commands.WorkingDirectory, "uploads", time.Now().Format("200601"), fileName+ext)

	path := filepath.Dir(filePath)

	os.MkdirAll(path, os.ModePerm)

	err = this.SaveToFile(name, filePath)

	if err != nil {
		beego.Error("SaveToFile => ", err)
		this.JsonResult(6005, "保存文件失败")
	}
	attachment := models.NewAttachment()
	attachment.BookId = book_id
	attachment.FileName = moreFile.Filename
	attachment.CreateAt = this.Member.MemberId
	attachment.FileExt = ext
	attachment.FilePath = strings.TrimPrefix(filePath, commands.WorkingDirectory)
	attachment.DocumentId = doc_id

	if fileInfo, err := os.Stat(filePath); err == nil {
		attachment.FileSize = float64(fileInfo.Size())
	}
	if doc_id > 0 {
		attachment.DocumentId = doc_id
	}

	if strings.EqualFold(ext, ".jpg") || strings.EqualFold(ext, ".jpeg") || strings.EqualFold(ext, ".png") || strings.EqualFold(ext, ".gif") {

		attachment.HttpPath = "/" + strings.Replace(strings.TrimPrefix(filePath, commands.WorkingDirectory), "\\", "/", -1)
		if strings.HasPrefix(attachment.HttpPath, "//") {
			attachment.HttpPath = string(attachment.HttpPath[1:])
		}
		is_attach = false
	}

	err = attachment.Insert()

	if err != nil {
		os.Remove(filePath)
		beego.Error("Attachment Insert => ", err)
		this.JsonResult(6006, "文件保存失败")
	}
	//TODO:移除debug
	beego.Debug(attachment)
	if attachment.HttpPath == "" {
		attachment.HttpPath = beego.URLFor("DocumentController.DownloadAttachment", ":key", identify, ":attach_id", attachment.AttachmentId)

		if err := attachment.Update(); err != nil {
			beego.Error("SaveToFile => ", err)
			this.JsonResult(6005, "保存文件失败")
		}
	}
	osspath := fmt.Sprintf("projects/%v/%v", identify, fileName+filepath.Ext(attachment.HttpPath))
	switch utils.StoreType {
	case utils.StoreOss:
		if err := models.ModelStoreOss.MoveToOss("."+attachment.HttpPath, osspath, true, false); err != nil {
			beego.Error(err.Error())
		}
		attachment.HttpPath = this.OssDomain + "/" + osspath
	case utils.StoreLocal:
		osspath = "uploads/" + osspath
		if err := models.ModelStoreLocal.MoveToStore("."+attachment.HttpPath, osspath); err != nil {
			beego.Error(err.Error())
		}
		attachment.HttpPath = "/" + osspath
	}

	result := map[string]interface{}{
		"errcode":   0,
		"success":   1,
		"message":   "ok",
		"url":       attachment.HttpPath,
		"alt":       attachment.FileName,
		"is_attach": is_attach,
		"attach":    attachment,
	}
	this.Ctx.Output.JSON(result, true, false)
	this.StopRun()
}

//DownloadAttachment 下载附件.
func (this *DocumentController) DownloadAttachment() {
	this.Prepare()

	identify := this.Ctx.Input.Param(":key")
	attach_id, _ := strconv.Atoi(this.Ctx.Input.Param(":attach_id"))
	token := this.GetString("token")

	member_id := 0

	if this.Member != nil {
		member_id = this.Member.MemberId
	}
	book_id := 0

	//判断用户是否参与了项目
	bookResult, err := models.NewBookResult().FindByIdentify(identify, member_id)

	if err != nil {
		//判断项目公开状态
		book, err := models.NewBook().FindByFieldFirst("identify", identify)
		if err != nil {
			this.Abort("404")
		}
		//如果不是超级管理员则判断权限
		if this.Member == nil || this.Member.Role != conf.MemberSuperRole {
			//如果项目是私有的，并且token不正确
			if (book.PrivatelyOwned == 1 && token == "") || (book.PrivatelyOwned == 1 && book.PrivateToken != token) {
				this.Abort("403")
			}
		}

		book_id = book.BookId
	} else {
		book_id = bookResult.BookId
	}
	//查找附件
	attachment, err := models.NewAttachment().Find(attach_id)

	if err != nil {
		beego.Error("DownloadAttachment => ", err)
		if err == orm.ErrNoRows {
			this.Abort("404")
		} else {
			this.Abort("500")
		}
	}
	if attachment.BookId != book_id {
		this.Abort("404")
	}
	this.Ctx.Output.Download(filepath.Join(commands.WorkingDirectory, attachment.FilePath), attachment.FileName)

	this.StopRun()
}

//删除附件.
func (this *DocumentController) RemoveAttachment() {
	this.Prepare()
	attach_id, _ := this.GetInt("attach_id")

	if attach_id <= 0 {
		this.JsonResult(6001, "参数错误")
	}
	attach, err := models.NewAttachment().Find(attach_id)

	if err != nil {
		beego.Error(err)
		this.JsonResult(6002, "附件不存在")
	}
	document, err := models.NewDocument().Find(attach.DocumentId)

	if err != nil {
		beego.Error(err)
		this.JsonResult(6003, "文档不存在")
	}
	if this.Member.Role != conf.MemberSuperRole {
		rel, err := models.NewRelationship().FindByBookIdAndMemberId(document.BookId, this.Member.MemberId)
		if err != nil {
			beego.Error(err)
			this.JsonResult(6004, "权限不足")
		}
		if rel.RoleId == conf.BookObserver {
			this.JsonResult(6004, "权限不足")
		}
	}
	err = attach.Delete()

	if err != nil {
		beego.Error(err)
		this.JsonResult(6005, "删除失败")
	}
	os.Remove(filepath.Join(commands.WorkingDirectory, attach.FilePath))

	this.JsonResult(0, "ok", attach)
}

//删除文档.
func (this *DocumentController) Delete() {

	identify := this.GetString("identify")
	doc_id, err := this.GetInt("doc_id", 0)

	book_id := 0
	//如果是超级管理员则忽略权限判断
	if this.Member.IsAdministrator() {
		book, err := models.NewBook().FindByFieldFirst("identify", identify)
		if err != nil {
			beego.Error("FindByIdentify => ", err)
			this.JsonResult(6002, "项目不存在或权限不足")
		}
		book_id = book.BookId
	} else {
		bookResult, err := models.NewBookResult().FindByIdentify(identify, this.Member.MemberId)

		if err != nil || bookResult.RoleId == conf.BookObserver {
			beego.Error("FindByIdentify => ", err)
			this.JsonResult(6002, "项目不存在或权限不足")
		}
		book_id = bookResult.BookId
	}

	if doc_id <= 0 {
		this.JsonResult(6001, "参数错误")
	}

	doc, err := models.NewDocument().Find(doc_id)

	if err != nil {
		beego.Error("Delete => ", err)
		this.JsonResult(6003, "删除失败")
	}
	//如果文档所属项目错误
	if doc.BookId != book_id {
		this.JsonResult(6004, "参数错误")
	}
	//递归删除项目下的文档以及子文档
	err = doc.RecursiveDocument(doc.DocumentId)
	if err != nil {
		this.JsonResult(6005, "删除失败")
	}
	//重置文档数量统计
	models.NewBook().ResetDocumentNumber(doc.BookId)

	this.JsonResult(0, "ok")
}

//获取或更新文档内容.
func (this *DocumentController) Content() {
	identify := this.Ctx.Input.Param(":key")
	doc_id, err := this.GetInt("doc_id")

	if err != nil {
		doc_id, _ = strconv.Atoi(this.Ctx.Input.Param(":id"))
	}
	book_id := 0
	//如果是超级管理员，则忽略权限
	if this.Member.IsAdministrator() {
		book, err := models.NewBook().FindByFieldFirst("identify", identify)
		if err != nil {
			this.JsonResult(6002, "项目不存在或权限不足")
		}
		book_id = book.BookId
	} else {
		bookResult, err := models.NewBookResult().FindByIdentify(identify, this.Member.MemberId)

		if err != nil || bookResult.RoleId == conf.BookObserver {
			beego.Error("FindByIdentify => ", err)
			this.JsonResult(6002, "项目不存在或权限不足")
		}
		book_id = bookResult.BookId
	}

	if doc_id <= 0 {
		this.JsonResult(6001, "参数错误")
	}
	ModelStore := new(models.DocumentStore)
	if this.Ctx.Input.IsPost() { //更新文档内容
		markdown := strings.TrimSpace(this.GetString("markdown", ""))
		content := this.GetString("html")

		version, _ := this.GetInt64("version", 0)
		is_cover := this.GetString("cover")

		doc, err := models.NewDocument().Find(doc_id)

		if err != nil {
			this.JsonResult(6003, "读取文档错误")
		}
		if doc.BookId != book_id {
			this.JsonResult(6004, "保存的文档不属于指定项目")
		}
		if doc.Version != version && !strings.EqualFold(is_cover, "yes") {
			beego.Info("%d|", version, doc.Version)
			this.JsonResult(6005, "文档已被修改确定要覆盖吗？")
		}

		is_summary := false
		is_auto := false
		//替换文档中的url链接
		if strings.ToLower(doc.Identify) == "summary.md" && (strings.Contains(markdown, "<DocStack-summary></DocStack-summary>") || strings.Contains(doc.Markdown, "<DocStack-summary/>")) {
			//如果标识是summary.md，并且带有DocStack的标签，则表示更新目录
			is_summary = true
			//要清除，避免每次保存的时候都要重新排序
			replaces := []string{"<DocStack-summary></DocStack-summary>", "<DocStack-summary/>"}
			for _, r := range replaces {
				markdown = strings.Replace(markdown, r, "", -1)
			}
		}
		if strings.Contains(markdown, "<DocStack-auto></DocStack-auto>") || strings.Contains(doc.Markdown, "<DocStack-auto/>") {
			//自动生成文档内容
			var docs []models.Document
			orm.NewOrm().QueryTable("md_documents").Filter("book_id", book_id).Filter("parent_id", doc_id).OrderBy("order_sort").All(&docs, "document_id", "document_name", "identify")
			var newCont []string //新HTML内容
			var newMd []string   //新markdown内容
			for _, idoc := range docs {
				newMd = append(newMd, fmt.Sprintf(`- [%v]($%v)`, idoc.DocumentName, idoc.Identify))
				newCont = append(newCont, fmt.Sprintf(`<li><a href="$%v">%v</a></li>`, idoc.Identify, idoc.DocumentName))
			}
			markdown = strings.Replace(markdown, "<DocStack-auto></DocStack-auto>", strings.Join(newMd, "\n"), -1)
			content = strings.Replace(content, "<DocStack-auto></DocStack-auto>", "<ul>"+strings.Join(newCont, "")+"</ul>", -1)
			is_auto = true
		}
		content = this.replaceLinks(identify, content, is_summary)

		var ds = models.DocumentStore{}

		if markdown == "" && content != "" {
			ds.Markdown = content
		} else {
			ds.Markdown = markdown
		}
		doc.Version = time.Now().Unix()
		ds.Content = content
		if doc_id, err := doc.InsertOrUpdate(); err != nil {
			beego.Error("InsertOrUpdate => ", err)
			this.JsonResult(6006, "保存失败")
		} else {
			ds.DocumentId = int(doc_id)
			if err := ModelStore.InsertOrUpdate(ds, "markdown", "content"); err != nil {
				beego.Error(err)
			}
		}
		//如果启用了文档历史，则添加历史文档
		if this.EnableDocumentHistory {

			history := models.NewDocumentHistory()
			history.DocumentId = doc_id
			history.Content = ds.Content
			history.Markdown = ds.Markdown
			history.DocumentName = doc.DocumentName
			history.ModifyAt = this.Member.MemberId
			history.MemberId = doc.MemberId
			history.ParentId = doc.ParentId
			history.Version = time.Now().Unix()
			history.Action = "modify"
			history.ActionName = "修改文档"
			_, err = history.InsertOrUpdate()
			if err != nil {
				beego.Error("DocumentHistory InsertOrUpdate => ", err)
			}
		}

		//doc.Markdown = ""
		//doc.Content = ""
		doc.Release = ""
		//注意：如果errMsg的值是true，则表示更新了目录排序，需要刷新，否则不刷新
		this.JsonResult(0, fmt.Sprintf("%v", is_summary || is_auto), doc)
	}
	doc, err := models.NewDocument().Find(doc_id)

	if err != nil {
		this.JsonResult(6003, "文档不存在")
	}
	attach, err := models.NewAttachment().FindListByDocumentId(doc.DocumentId)
	if err == nil {
		doc.AttachList = attach
	}

	//为了减少数据的传输量，这里Release和Content的内容置空，前端会根据markdown文本自动渲染
	//doc.Release = ""
	//doc.Content = ""
	doc.Markdown = ModelStore.GetFiledById(doc.DocumentId, "markdown")
	this.JsonResult(0, "ok", doc)
}

//导出文件
func (this *DocumentController) Export() {
	this.TplName = "document/export.html"
	identify := this.Ctx.Input.Param(":key")
	ext := strings.ToLower(this.GetString("output"))
	switch ext {
	case "pdf", "epub", "mobi":
		ext = "." + ext
	default:
		ext = ".pdf"
	}
	if identify == "" {
		this.JsonResult(1, "下载失败，无法识别您要下载的文档")
	}
	if book, err := new(models.Book).FindByIdentify(identify); err == nil {
		if book.PrivatelyOwned == 1 && this.Member.MemberId != book.MemberId {
			this.JsonResult(1, "私有文档，禁止导出")
		} else {
			//查询文档是否存在
			obj := fmt.Sprintf("projects/%v/books/%v%v", book.Identify, book.GenerateTime.Unix(), ext)
			switch utils.StoreType {
			case utils.StoreOss:
				if err := models.ModelStoreOss.IsObjectExist(obj); err != nil {
					beego.Error(err, obj)
					this.JsonResult(1, "下载失败，您要下载的文档当前并未生成可下载文档。")
				} else {
					this.JsonResult(0, "获取文档下载链接成功", map[string]interface{}{"url": this.OssDomain + "/" + obj})
				}
			case utils.StoreLocal:
				obj = "uploads/" + obj
				if err := models.ModelStoreLocal.IsObjectExist(obj); err != nil {
					beego.Error(err, obj)
					this.JsonResult(1, "下载失败，您要下载的文档当前并未生成可下载文档。")
				} else {
					this.JsonResult(0, "获取文档下载链接成功", map[string]interface{}{"url": "/" + obj})
				}
			}

		}
	} else {
		beego.Error(err.Error())
	}

}

//导出文件
//func (this *DocumentController) Export() {
//	this.TplName = "document/export.html"
//	identify := this.Ctx.Input.Param(":key")
//	output := this.GetString("output")
//	token := this.GetString("token")
//	if identify == "" {
//		this.Abort("404")
//	}
//
//	//如果没有开启你们访问则跳转到登录
//	if !this.EnableAnonymous && this.Member == nil {
//		this.Redirect(beego.URLFor("AccountController.Login"), 302)
//		return
//	}
//	bookResult := models.NewBookResult()
//	if this.Member != nil && this.Member.IsAdministrator() {
//		book, err := models.NewBook().FindByIdentify(identify)
//		if err != nil {
//			beego.Error(err)
//			this.Abort("500")
//		}
//		bookResult = book.ToBookResult()
//	} else {
//		bookResult = isReadable(identify, token, this)
//	}
//
//	if bookResult.PrivatelyOwned == 0 {
//		//TODO 私有项目禁止导出
//	}
//
//	docs, err := models.NewDocument().FindListByBookId(bookResult.BookId)
//
//	if err != nil {
//		beego.Error(err)
//		this.Abort("500")
//	}
//
//	if output == "pdf" {
//
//		exe := beego.AppConfig.String("wkhtmltopdf")
//
//		if exe == "" {
//			this.TplName = "errors/error.html"
//			this.Data["ErrorMessage"] = "没有配置PDF导出程序"
//			this.Data["ErrorCode"] = 50010
//			return
//		}
//		dpath := "cache/" + bookResult.Identify
//
//		os.MkdirAll(dpath, 0766)
//
//		pathList := list.New()
//
//		RecursiveFun(0, "", dpath, this, bookResult, docs, pathList)
//
//		defer os.RemoveAll(dpath)
//
//		os.MkdirAll("./cache", 0766)
//		pdfpath := filepath.Join("cache", identify+"_"+this.CruSession.SessionID()+".pdf")
//
//		if _, err := os.Stat(pdfpath); os.IsNotExist(err) {
//
//			wkhtmltopdf.SetPath(beego.AppConfig.String("wkhtmltopdf"))
//			pdfg, err := wkhtmltopdf.NewPDFGenerator()
//			pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
//			pdfg.MarginBottom.Set(25)
//
//			if err != nil {
//				beego.Error(err)
//				this.Abort("500")
//			}
//
//			for e := pathList.Front(); e != nil; e = e.Next() {
//				if page, ok := e.Value.(string); ok {
//					pdfg.AddPage(wkhtmltopdf.NewPage(page))
//				}
//			}
//			pdfg.MoreArgs = append(pdfg.MoreArgs,
//				"--header-font-size", "8",
//				"--footer-right", "[page] / [toPage]",
//				"--footer-spacing", "5",
//				"--footer-html", "views/widgets/pdf_footer.html",
//				"--footer-font-size", "8",
//			)
//			//beego.Debug(pdfg.ArgString())
//			//TODO 处理页码和footer、header问题
//			//this.JsonResult(0, "1", pdfg.ArgString())
//
//			err = pdfg.Create()
//			if err != nil {
//				beego.Error(err)
//				this.Abort("500")
//			}
//
//			err = pdfg.WriteFile(pdfpath)
//			if err != nil {
//				beego.Error(err)
//			}
//		}
//
//		this.Ctx.Output.Download(pdfpath, bookResult.BookName+".pdf")
//
//		defer os.Remove(pdfpath)
//
//		this.StopRun()
//	}
//
//	this.Abort("404")
//}

//生成项目访问的二维码.

func (this *DocumentController) QrCode() {
	this.Prepare()
	identify := this.GetString(":key")

	book, err := models.NewBook().FindByIdentify(identify)

	if err != nil || book.BookId <= 0 {
		this.Abort("404")
	}

	uri := this.BaseUrl() + beego.URLFor("DocumentController.Index", ":key", identify)
	code, err := qr.Encode(uri, qr.L, qr.Unicode)
	if err != nil {
		beego.Error(err)
		this.Abort("500")
	}
	code, err = barcode.Scale(code, 150, 150)

	if err != nil {
		beego.Error(err)
		this.Abort("500")
	}
	this.Ctx.ResponseWriter.Header().Set("Content-Type", "image/png")

	//imgpath := filepath.Join("cache","qrcode",identify + ".png")

	err = png.Encode(this.Ctx.ResponseWriter, code)
	if err != nil {
		beego.Error(err)
		this.Abort("500")
	}
}

//项目内搜索.
func (this *DocumentController) Search() {
	this.Prepare()

	identify := this.Ctx.Input.Param(":key")
	token := this.GetString("token")
	keyword := strings.TrimSpace(this.GetString("keyword"))

	if identify == "" {
		this.JsonResult(6001, "参数错误")
	}
	if !this.EnableAnonymous && this.Member == nil {
		this.Redirect(beego.URLFor("AccountController.Login"), 302)
		return
	}
	bookResult := isReadable(identify, token, this)

	docs, err := models.NewDocumentSearchResult().SearchDocument(keyword, bookResult.BookId)

	if err != nil {
		beego.Error(err)
		this.JsonResult(6002, "搜索结果错误")
	}
	if len(docs) < 0 {
		this.JsonResult(404, "没有数据库")
	}
	for _, doc := range docs {
		doc.BookId = bookResult.BookId
		doc.BookName = bookResult.BookName
		doc.Description = bookResult.Description
		doc.BookIdentify = bookResult.Identify
	}

	this.JsonResult(0, "ok", docs)
}

//文档历史列表.
func (this *DocumentController) History() {
	this.Prepare()
	this.TplName = "document/history.html"

	identify := this.GetString("identify")
	doc_id, err := this.GetInt("doc_id", 0)
	pageIndex, _ := this.GetInt("page", 1)

	book_id := 0
	//如果是超级管理员则忽略权限判断
	if this.Member.IsAdministrator() {
		book, err := models.NewBook().FindByFieldFirst("identify", identify)
		if err != nil {
			beego.Error("FindByIdentify => ", err)
			this.Data["ErrorMessage"] = "项目不存在或权限不足"
			return
		}
		book_id = book.BookId
		this.Data["Model"] = book
	} else {
		bookResult, err := models.NewBookResult().FindByIdentify(identify, this.Member.MemberId)

		if err != nil || bookResult.RoleId == conf.BookObserver {
			beego.Error("FindByIdentify => ", err)
			this.Data["ErrorMessage"] = "项目不存在或权限不足"
			return
		}
		book_id = bookResult.BookId
		this.Data["Model"] = bookResult
	}

	if doc_id <= 0 {
		this.Data["ErrorMessage"] = "参数错误"
		return
	}

	doc, err := models.NewDocument().Find(doc_id)

	if err != nil {
		beego.Error("Delete => ", err)
		this.Data["ErrorMessage"] = "获取历史失败"
		return
	}
	//如果文档所属项目错误
	if doc.BookId != book_id {
		this.Data["ErrorMessage"] = "参数错误"
		return
	}

	historis, totalCount, err := models.NewDocumentHistory().FindToPager(doc_id, pageIndex, conf.PageSize)

	if err != nil {
		beego.Error("FindToPager => ", err)
		this.Data["ErrorMessage"] = "获取历史失败"
		return
	}

	this.Data["List"] = historis
	this.Data["PageHtml"] = ""
	this.Data["Document"] = doc

	if totalCount > 0 {
		html := utils.GetPagerHtml(this.Ctx.Request.RequestURI, pageIndex, conf.PageSize, totalCount)

		this.Data["PageHtml"] = html
	}
}

func (this *DocumentController) DeleteHistory() {
	this.Prepare()
	this.TplName = "document/history.html"

	identify := this.GetString("identify")
	doc_id, err := this.GetInt("doc_id", 0)
	history_id, _ := this.GetInt("history_id", 0)

	if history_id <= 0 {
		this.JsonResult(6001, "参数错误")
	}
	book_id := 0
	//如果是超级管理员则忽略权限判断
	if this.Member.IsAdministrator() {
		book, err := models.NewBook().FindByFieldFirst("identify", identify)
		if err != nil {
			beego.Error("FindByIdentify => ", err)
			this.JsonResult(6002, "项目不存在或权限不足")
		}
		book_id = book.BookId
	} else {
		bookResult, err := models.NewBookResult().FindByIdentify(identify, this.Member.MemberId)

		if err != nil || bookResult.RoleId == conf.BookObserver {
			beego.Error("FindByIdentify => ", err)
			this.JsonResult(6002, "项目不存在或权限不足")
		}
		book_id = bookResult.BookId
	}

	if doc_id <= 0 {
		this.JsonResult(6001, "参数错误")
	}

	doc, err := models.NewDocument().Find(doc_id)

	if err != nil {
		beego.Error("Delete => ", err)
		this.JsonResult(6001, "获取历史失败")
	}
	//如果文档所属项目错误
	if doc.BookId != book_id {
		this.JsonResult(6001, "参数错误")
	}
	err = models.NewDocumentHistory().Delete(history_id, doc_id)
	if err != nil {
		beego.Error(err)
		this.JsonResult(6002, "删除失败")
	}
	this.JsonResult(0, "ok")
}

func (this *DocumentController) RestoreHistory() {
	this.Prepare()
	this.TplName = "document/history.html"

	identify := this.GetString("identify")
	doc_id, err := this.GetInt("doc_id", 0)
	history_id, _ := this.GetInt("history_id", 0)

	if history_id <= 0 {
		this.JsonResult(6001, "参数错误")
	}
	book_id := 0
	//如果是超级管理员则忽略权限判断
	if this.Member.IsAdministrator() {
		book, err := models.NewBook().FindByFieldFirst("identify", identify)
		if err != nil {
			beego.Error("FindByIdentify => ", err)
			this.JsonResult(6002, "项目不存在或权限不足")
		}
		book_id = book.BookId
	} else {
		bookResult, err := models.NewBookResult().FindByIdentify(identify, this.Member.MemberId)

		if err != nil || bookResult.RoleId == conf.BookObserver {
			beego.Error("FindByIdentify => ", err)
			this.JsonResult(6002, "项目不存在或权限不足")
		}
		book_id = bookResult.BookId
	}

	if doc_id <= 0 {
		this.JsonResult(6001, "参数错误")
	}

	doc, err := models.NewDocument().Find(doc_id)

	if err != nil {
		beego.Error("Delete => ", err)
		this.JsonResult(6001, "获取历史失败")
	}
	//如果文档所属项目错误
	if doc.BookId != book_id {
		this.JsonResult(6001, "参数错误")
	}
	err = models.NewDocumentHistory().Restore(history_id, doc_id, this.Member.MemberId)
	if err != nil {
		beego.Error(err)
		this.JsonResult(6002, "删除失败")
	}
	this.JsonResult(0, "ok", doc)
}

func (this *DocumentController) Compare() {
	this.Prepare()
	this.TplName = "document/compare.html"
	history_id, _ := strconv.Atoi(this.Ctx.Input.Param(":id"))
	identify := this.Ctx.Input.Param(":key")

	book_id := 0
	editor := "markdown"

	//如果是超级管理员则忽略权限判断
	if this.Member.IsAdministrator() {
		book, err := models.NewBook().FindByFieldFirst("identify", identify)
		if err != nil {
			beego.Error("DocumentController.Compare => ", err)
			this.Abort("403")
			return
		}
		book_id = book.BookId
		this.Data["Model"] = book
		editor = book.Editor
	} else {
		bookResult, err := models.NewBookResult().FindByIdentify(identify, this.Member.MemberId)

		if err != nil || bookResult.RoleId == conf.BookObserver {
			beego.Error("FindByIdentify => ", err)
			this.Abort("403")
			return
		}
		book_id = bookResult.BookId
		this.Data["Model"] = bookResult
		editor = bookResult.Editor
	}

	if history_id <= 0 {
		this.ShowErrorPage(60002, "参数错误")
	}

	history, err := models.NewDocumentHistory().Find(history_id)
	if err != nil {
		beego.Error("DocumentController.Compare => ", err)
		this.ShowErrorPage(60003, err.Error())
	}
	doc, err := models.NewDocument().Find(history.DocumentId)

	if doc.BookId != book_id {
		this.ShowErrorPage(60002, "参数错误")
	}
	this.Data["HistoryId"] = history_id
	this.Data["DocumentId"] = doc.DocumentId
	ModelStore := new(models.DocumentStore)
	if editor == "markdown" {
		this.Data["HistoryContent"] = history.Markdown
		this.Data["Content"] = ModelStore.GetFiledById(doc.DocumentId, "markdown")
	} else {
		this.Data["HistoryContent"] = template.HTML(history.Content)
		this.Data["Content"] = template.HTML(ModelStore.GetFiledById(doc.DocumentId, "content"))
	}
}

//递归生成文档序列数组.
func RecursiveFun(parent_id int, prefix, dpath string, this *DocumentController, book *models.BookResult, docs []*models.Document, paths *list.List) {
	for _, item := range docs {
		if item.ParentId == parent_id {
			name := prefix + strconv.Itoa(item.ParentId) + strconv.Itoa(item.OrderSort) + strconv.Itoa(item.DocumentId)
			fpath := dpath + "/" + name + ".html"
			paths.PushBack(fpath)

			f, err := os.OpenFile(fpath, os.O_CREATE|os.O_RDWR, 0777)

			if err != nil {
				beego.Error(err)
				this.Abort("500")
			}

			html, err := this.ExecuteViewPathTemplate("document/export.html", map[string]interface{}{"Model": book, "Lists": item, "BaseUrl": this.BaseUrl()})
			if err != nil {
				f.Close()
				beego.Error(err)
				this.Abort("500")
			}

			buf := bytes.NewReader([]byte(html))
			doc, err := goquery.NewDocumentFromReader(buf)
			doc.Find("img").Each(func(i int, contentSelection *goquery.Selection) {
				if src, ok := contentSelection.Attr("src"); ok && strings.HasPrefix(src, "/uploads/") {
					contentSelection.SetAttr("src", this.BaseUrl()+src)
				}
			})
			html, err = doc.Html()

			if err != nil {
				f.Close()
				beego.Error(err)
				this.Abort("500")
			}
			//html = strings.Replace(html, "<img src=\"/uploads", "<img src=\""+this.BaseUrl()+"/uploads", -1)

			f.WriteString(html)
			f.Close()

			for _, sub := range docs {
				if sub.ParentId == item.DocumentId {
					RecursiveFun(item.DocumentId, name, dpath, this, book, docs, paths)
					break
				}
			}
		}
	}
}
