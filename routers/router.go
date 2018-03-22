package routers

import (
	"github.com/JermineHu/DocStack/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.HomeController{}, "*:Index")

	beego.Router("/login", &controllers.AccountController{}, "*:Login")
	beego.Router("/login/:oauth", &controllers.AccountController{}, "*:Oauth")
	beego.Router("/logout", &controllers.AccountController{}, "*:Logout")
	beego.Router("/bind", &controllers.AccountController{}, "post:Bind")
	//beego.Router("/find_password", &controllers.AccountController{}, "*:FindPassword")
	beego.Router("/valid_email", &controllers.AccountController{}, "post:ValidEmail")
	//beego.Router("/captcha", &controllers.AccountController{}, "*:Captcha")

	beego.Router("/manager", &controllers.ManagerController{}, "*:Index")
	beego.Router("/manager/users", &controllers.ManagerController{}, "*:Users")
	beego.Router("/manager/users/edit/:id", &controllers.ManagerController{}, "*:EditMember")
	beego.Router("/manager/member/create", &controllers.ManagerController{}, "post:CreateMember")
	beego.Router("/manager/member/delete", &controllers.ManagerController{}, "post:DeleteMember")
	beego.Router("/manager/member/update-member-status", &controllers.ManagerController{}, "post:UpdateMemberStatus")
	beego.Router("/manager/member/change-member-role", &controllers.ManagerController{}, "post:ChangeMemberRole")
	beego.Router("/manager/books", &controllers.ManagerController{}, "*:Books")
	beego.Router("/manager/books/edit/:key", &controllers.ManagerController{}, "*:EditBook")
	beego.Router("/manager/books/delete", &controllers.ManagerController{}, "*:DeleteBook")
	beego.Router("/manager/comments", &controllers.ManagerController{}, "*:Comments")
	beego.Router("/manager/books/token", &controllers.ManagerController{}, "post:CreateToken")
	beego.Router("/manager/setting", &controllers.ManagerController{}, "*:Setting")
	beego.Router("/manager/books/transfer", &controllers.ManagerController{}, "post:Transfer")
	beego.Router("/manager/books/sort", &controllers.ManagerController{}, "get:UpdateBookSort")
	beego.Router("/manager/books/open", &controllers.ManagerController{}, "post:PrivatelyOwned")
	beego.Router("/manager/attach/list", &controllers.ManagerController{}, "*:AttachList")
	beego.Router("/manager/attach/detailed/:id", &controllers.ManagerController{}, "*:AttachDetailed")
	beego.Router("/manager/attach/delete", &controllers.ManagerController{}, "post:AttachDelete")
	beego.Router("/manager/seo", &controllers.ManagerController{}, "post,get:Seo")
	beego.Router("/manager/sitemap", &controllers.ManagerController{}, "get:Sitemap") //更新站点地图

	beego.Router("/setting", &controllers.SettingController{}, "*:Index")
	beego.Router("/setting/password", &controllers.SettingController{}, "*:Password")
	beego.Router("/setting/upload", &controllers.SettingController{}, "*:Upload")
	beego.Router("/setting/star", &controllers.SettingController{}, "*:Star")
	beego.Router("/setting/qrcode", &controllers.SettingController{}, "*:Qrcode")

	beego.Router("/book", &controllers.BookController{}, "*:Index")
	beego.Router("/book/star/:id", &controllers.BookController{}, "*:Star")          //收藏
	beego.Router("/book/score/:id", &controllers.BookController{}, "*:Score")        //收藏
	beego.Router("/book/comment/:id", &controllers.BookController{}, "post:Comment") //收藏
	beego.Router("/book/uploadProject", &controllers.BookController{}, "post:UploadProject")
	beego.Router("/book/downloadProject", &controllers.BookController{}, "post:DownloadProject")
	beego.Router("/book/:key/dashboard", &controllers.BookController{}, "*:Dashboard")
	beego.Router("/book/:key/setting", &controllers.BookController{}, "*:Setting")
	beego.Router("/book/:key/users", &controllers.BookController{}, "*:Users")
	beego.Router("/book/:key/release", &controllers.BookController{}, "post:Release")
	beego.Router("/book/:key/generate", &controllers.BookController{}, "get,post:Generate")
	beego.Router("/book/:key/sort", &controllers.BookController{}, "post:SaveSort")

	beego.Router("/book/create", &controllers.BookController{}, "*:Create")
	beego.Router("/book/users/create", &controllers.BookMemberController{}, "post:AddMember")
	beego.Router("/book/users/change", &controllers.BookMemberController{}, "post:ChangeRole")
	beego.Router("/book/users/delete", &controllers.BookMemberController{}, "post:RemoveMember")

	beego.Router("/book/setting/save", &controllers.BookController{}, "post:SaveBook")
	beego.Router("/book/setting/open", &controllers.BookController{}, "post:PrivatelyOwned")
	beego.Router("/book/setting/transfer", &controllers.BookController{}, "post:Transfer")
	beego.Router("/book/setting/upload", &controllers.BookController{}, "post:UploadCover")
	beego.Router("/book/setting/token", &controllers.BookController{}, "post:CreateToken")
	beego.Router("/book/setting/delete", &controllers.BookController{}, "post:Delete")

	beego.Router("/api/attach/remove/", &controllers.DocumentController{}, "post:RemoveAttachment")
	beego.Router("/api/:key/edit/?:id", &controllers.DocumentController{}, "*:Edit")
	beego.Router("/api/upload", &controllers.DocumentController{}, "post:Upload")
	beego.Router("/api/:key/create", &controllers.DocumentController{}, "post:Create")
	beego.Router("/api/create_multi", &controllers.DocumentController{}, "post:CreateMulti")
	beego.Router("/api/:key/delete", &controllers.DocumentController{}, "post:Delete")
	beego.Router("/api/:key/content/?:id", &controllers.DocumentController{}, "*:Content")
	beego.Router("/api/:key/compare/:id", &controllers.DocumentController{}, "*:Compare")

	beego.Router("/history/get", &controllers.DocumentController{}, "get:History")
	beego.Router("/history/delete", &controllers.DocumentController{}, "*:DeleteHistory")
	beego.Router("/history/restore", &controllers.DocumentController{}, "*:RestoreHistory")

	beego.Router("/books/:key", &controllers.DocumentController{}, "*:Index")
	beego.Router("/read/:key/:id", &controllers.DocumentController{}, "*:Read")
	beego.Router("/read/:key/search", &controllers.DocumentController{}, "post:Search")

	beego.Router("/export/:key", &controllers.DocumentController{}, "*:Export")
	beego.Router("/qrcode/:key.png", &controllers.DocumentController{}, "get:QrCode")

	beego.Router("/attach_files/:key/:attach_id", &controllers.DocumentController{}, "get:DownloadAttachment")

	beego.Router("/comment/create", &controllers.CommentController{}, "post:Create")
	beego.Router("/comment/lists", &controllers.CommentController{}, "get:Lists")
	beego.Router("/comment/index", &controllers.CommentController{}, "*:Index")

	beego.Router("/search", &controllers.SearchController{}, "get:Index")
	beego.Router("/crawl", &controllers.BaseController{}, "post:Crawl")

	beego.Router("/tag/:key", &controllers.LabelController{}, "get:Index")
	beego.Router("/tag", &controllers.LabelController{}, "get:List")
	beego.Router("/tags", &controllers.LabelController{}, "get:List")
	beego.Router("/sitemap.html", &controllers.BaseController{}, "get:Sitemap")
	beego.Router("/local-render", &controllers.LocalhostController{}, "get,post:RenderMarkdown")
	beego.Router("/*", &controllers.BaseController{}, "get:StaticFile")
}
