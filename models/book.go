package models

import (
	"time"

	"strings"

	"fmt"

	"strconv"

	"github.com/JermineHu/DocStack/conf"
	"github.com/JermineHu/DocStack/utils"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// Book struct .
type Book struct {
	BookId            int       `orm:"pk;auto;unique;column(book_id)" json:"book_id"`
	BookName          string    `orm:"column(book_name);size(500)" json:"book_name"`      // BookName 项目名称.
	Identify          string    `orm:"column(identify);size(100);unique" json:"identify"` // Identify 项目唯一标识.
	OrderIndex        int       `orm:"column(order_index);type(int);default(0)" json:"order_index"`
	Description       string    `orm:"column(description);size(2000)" json:"description"` // Description 项目描述.
	Label             string    `orm:"column(label);size(500)" json:"label"`
	PrivatelyOwned    int       `orm:"column(privately_owned);type(int);default(0)" json:"privately_owned"` // PrivatelyOwned 项目私有： 0 公开/ 1 私有
	PrivateToken      string    `orm:"column(private_token);size(500);null" json:"private_token"`           // 当项目是私有时的访问Token.
	Status            int       `orm:"column(status);type(int);default(0)" json:"status"`                   //状态：0 正常/1 已删除
	Editor            string    `orm:"column(editor);size(50)" json:"editor"`                               //默认的编辑器.
	DocCount          int       `orm:"column(doc_count);type(int)" json:"doc_count"`                        // DocCount 包含文档数量.
	CommentStatus     string    `orm:"column(comment_status);size(20);default(open)" json:"comment_status"` // CommentStatus 评论设置的状态:open 为允许所有人评论，closed 为不允许评论, group_only 仅允许参与者评论 ,registered_only 仅允许注册者评论.
	CommentCount      int       `orm:"column(comment_count);type(int)" json:"comment_count"`
	Cover             string    `orm:"column(cover);size(1000)" json:"cover"`                              //封面地址
	Theme             string    `orm:"column(theme);size(255);default(default)" json:"theme"`              //主题风格
	CreateTime        time.Time `orm:"type(datetime);column(create_time);auto_now_add" json:"create_time"` // CreateTime 创建时间 .
	MemberId          int       `orm:"column(member_id);size(100)" json:"member_id"`
	ModifyTime        time.Time `orm:"type(datetime);column(modify_time);auto_now_add" json:"modify_time"`
	ReleaseTime       time.Time `orm:"type(datetime);column(release_time);" json:"release_time"`   //项目发布时间，每次发布都更新一次，如果文档更新时间小于发布时间，则文档不再执行发布
	GenerateTime      time.Time `orm:"type(datetime);column(generate_time);" json:"generate_time"` //下载文档生成时间
	LastClickGenerate time.Time `orm:"type(datetime);column(last_click_generate)" json:"-"`        //上次点击上传文档的时间，用于显示频繁点击浪费服务器硬件资源的情况
	Version           int64     `orm:"type(bigint);column(version);default(0)" json:"version"`
	Vcnt              int       `orm:"column(vcnt);default(0)" json:"vcnt"`    //文档项目被阅读次数
	Star              int       `orm:"column(star);default(0)" json:"star"`    //文档项目被收藏次数
	Score             int       `orm:"column(score);default(40)" json:"score"` //文档项目评分，默认40，即4.0星
	CntScore          int       //评分人数
	CntComment        int       //评论人数
}

// TableName 获取对应数据库表名.
func (m *Book) TableName() string {
	return "books"
}

// TableEngine 获取数据使用的引擎.
func (m *Book) TableEngine() string {
	return "INNODB"
}
func (m *Book) TableNameWithPrefix() string {
	return conf.GetDatabasePrefix() + m.TableName()
}

func NewBook() *Book {
	return &Book{}
}

func (m *Book) Insert() error {
	o := orm.NewOrm()
	_, err := o.Insert(m)
	if err == nil {
		if m.Label != "" {
			NewLabel().InsertOrUpdateMulti(m.Label)
		}
		relationship := NewRelationship()
		relationship.BookId = m.BookId
		relationship.RoleId = 0
		relationship.MemberId = m.MemberId
		if err = relationship.Insert(); err != nil {
			logs.Error("插入项目与用户关联 => ", err)
			return err
		}
		document := NewDocument()
		document.BookId = m.BookId
		document.DocumentName = "空白文档"
		document.Identify = "blank"
		document.MemberId = m.MemberId
		if id, err := document.InsertOrUpdate(); err == nil {
			var ds = DocumentStore{
				DocumentId: int(id),
				Markdown:   "[TOC]\n\r\n\r", //默认内容
			}
			err = new(DocumentStore).InsertOrUpdate(ds)
			return err
		}
	}
	return err
}

func (m *Book) Find(id int) (*Book, error) {
	if id <= 0 {
		return m, ErrInvalidParameter
	}
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("book_id", id).One(m)

	return m, err
}

func (m *Book) Update(cols ...string) error {
	o := orm.NewOrm()

	temp := NewBook()
	temp.BookId = m.BookId

	if err := o.Read(temp); err != nil {
		return err
	}

	if (m.Label + temp.Label) != "" {

		go NewLabel().InsertOrUpdateMulti(m.Label + "," + temp.Label)
	}

	_, err := o.Update(m, cols...)
	return err
}

//根据指定字段查询结果集.
func (m *Book) FindByField(field string, value interface{}) ([]*Book, error) {
	o := orm.NewOrm()

	var books []*Book
	_, err := o.QueryTable(m.TableNameWithPrefix()).Filter(field, value).All(&books)

	return books, err
}

//根据指定字段查询一个结果.
func (m *Book) FindByFieldFirst(field string, value interface{}) (*Book, error) {
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter(field, value).One(m)

	return m, err

}

func (m *Book) FindByIdentify(identify string) (*Book, error) {
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("identify", identify).One(m)

	return m, err
}

//分页查询指定用户的项目
//按照最新的进行排序
func (m *Book) FindToPager(pageIndex, pageSize, memberId int, PrivatelyOwned ...int) (books []*BookResult, totalCount int, err error) {

	relationship := NewRelationship()

	o := orm.NewOrm()

	sql1 := "SELECT COUNT(book.book_id) AS total_count FROM " + m.TableNameWithPrefix() + " AS book LEFT JOIN " +
		relationship.TableNameWithPrefix() + " AS rel ON book.book_id=rel.book_id AND rel.member_id = ? WHERE rel.relationship_id > 0 "
	if len(PrivatelyOwned) > 0 {
		sql1 = sql1 + " and book.privately_owned=" + strconv.Itoa(PrivatelyOwned[0])
	}
	err = o.Raw(sql1, memberId).QueryRow(&totalCount)

	if err != nil {
		return
	}

	offset := (pageIndex - 1) * pageSize

	sql2 := "SELECT book.*,rel.member_id,rel.role_id,m.account as create_name FROM " + m.TableNameWithPrefix() + " AS book" +
		" LEFT JOIN " + relationship.TableNameWithPrefix() + " AS rel ON book.book_id=rel.book_id AND rel.member_id = ?" +
		" LEFT JOIN " + relationship.TableNameWithPrefix() + " AS rel1 ON book.book_id=rel1.book_id  AND rel1.role_id=0" +
		" LEFT JOIN " + NewMember().TableNameWithPrefix() + " AS m ON rel1.member_id=m.member_id " +
		" WHERE rel.relationship_id > 0 %v ORDER BY book.book_id DESC LIMIT " + fmt.Sprintf("%d,%d", offset, pageSize)
	if len(PrivatelyOwned) > 0 {
		sql2 = fmt.Sprintf(sql2, " and book.privately_owned="+strconv.Itoa(PrivatelyOwned[0]))
	}
	_, err = o.Raw(sql2, memberId).QueryRows(&books)
	if err != nil {
		logs.Error("分页查询项目列表 => ", err)
		return
	}

	if err == nil && len(books) > 0 {
		sql := "SELECT m.account,doc.modify_time FROM md_documents AS doc LEFT JOIN md_members AS m ON doc.modify_at=m.member_id WHERE book_id = ? ORDER BY doc.modify_time DESC LIMIT 1 "

		for index, book := range books {
			var text struct {
				Account    string
				ModifyTime time.Time
			}

			err1 := o.Raw(sql, book.BookId).QueryRow(&text)
			if err1 == nil {
				books[index].LastModifyText = text.Account + " 于 " + text.ModifyTime.Format("2006-01-02 15:04:05")
			}
			if book.RoleId == 0 {
				book.RoleName = "创始人"
			} else if book.RoleId == 1 {
				book.RoleName = "管理员"
			} else if book.RoleId == 2 {
				book.RoleName = "编辑者"
			} else if book.RoleId == 3 {
				book.RoleName = "观察者"
			}
		}
	}
	return
}

// 彻底删除项目.
func (m *Book) ThoroughDeleteBook(id int) error {
	if id <= 0 {
		return ErrInvalidParameter
	}
	o := orm.NewOrm()

	m.BookId = id
	if err := o.Read(m); err != nil {
		return err
	}
	o.Begin()

	//删除md_document_store中的文档
	sql := "delete from md_document_store where document_id in(select document_id from md_documents where book_id=?)"
	if _, err := o.Raw(sql, m.BookId).Exec(); err != nil {
		beego.Error(err)
	}

	sql2 := "DELETE FROM " + NewDocument().TableNameWithPrefix() + " WHERE book_id = ?"

	_, err := o.Raw(sql2, m.BookId).Exec()

	if err != nil {
		o.Rollback()
		return err
	}
	sql3 := "DELETE FROM " + m.TableNameWithPrefix() + " WHERE book_id = ?"

	_, err = o.Raw(sql3, m.BookId).Exec()

	if err != nil {
		o.Rollback()
		return err
	}
	sql4 := "DELETE FROM " + NewRelationship().TableNameWithPrefix() + " WHERE book_id = ?"

	_, err = o.Raw(sql4, m.BookId).Exec()

	if err != nil {
		o.Rollback()
		return err
	}

	if m.Label != "" {
		NewLabel().InsertOrUpdateMulti(m.Label)
	}

	if err = o.Commit(); err == nil {
		//删除oss中项目对应的文件夹
		switch utils.StoreType {
		case utils.StoreLocal: //删除本地存储，记得加上uploads
			go ModelStoreLocal.DelFromFolder("uploads/projects/" + m.Identify)
		case utils.StoreOss:
			go ModelStoreOss.DelOssFolder("projects/" + m.Identify)
		}
	}
	return err
}

//分页查找系统首页数据.
func (m *Book) FindForHomeToPager(pageIndex, pageSize, member_id int) (books []*BookResult, totalCount int, err error) {
	o := orm.NewOrm()

	offset := (pageIndex - 1) * pageSize
	//如果是登录用户
	if member_id > 0 {
		sql1 := "SELECT COUNT(*) FROM md_books AS book LEFT JOIN md_relationship AS rel ON rel.book_id = book.book_id AND rel.member_id = ? WHERE relationship_id > 0 OR book.privately_owned = 0"

		err = o.Raw(sql1, member_id).QueryRow(&totalCount)
		if err != nil {
			return
		}
		sql2 := `SELECT book.*,rel1.*,member.account AS create_name FROM md_books AS book
			LEFT JOIN md_relationship AS rel ON rel.book_id = book.book_id AND rel.member_id = ?
			LEFT JOIN md_relationship AS rel1 ON rel1.book_id = book.book_id AND rel1.role_id = 0
			LEFT JOIN md_members AS member ON rel1.member_id = member.member_id
			WHERE rel.relationship_id > 0 OR book.privately_owned = 0 ORDER BY order_index DESC ,book.book_id DESC LIMIT ?,?`

		_, err = o.Raw(sql2, member_id, offset, pageSize).QueryRows(&books)

		return

	} else {
		count, err1 := o.QueryTable(m.TableNameWithPrefix()).Filter("privately_owned", 0).Count()

		if err1 != nil {
			err = err1
			return
		}
		totalCount = int(count)

		sql := `SELECT book.*,rel.*,member.account AS create_name FROM md_books AS book
			LEFT JOIN md_relationship AS rel ON rel.book_id = book.book_id AND rel.role_id = 0
			LEFT JOIN md_members AS member ON rel.member_id = member.member_id
			WHERE book.privately_owned = 0 ORDER BY order_index DESC ,book.book_id DESC LIMIT ?,?`

		_, err = o.Raw(sql, offset, pageSize).QueryRows(&books)

		return

	}

}

//分页全局搜索.
func (m *Book) FindForLabelToPager(keyword string, pageIndex, pageSize, member_id int) (books []*BookResult, totalCount int, err error) {
	o := orm.NewOrm()

	keyword = "%" + keyword + "%"
	offset := (pageIndex - 1) * pageSize
	//如果是登录用户
	if member_id > 0 {
		sql1 := "SELECT COUNT(*) FROM md_books AS book LEFT JOIN md_relationship AS rel ON rel.book_id = book.book_id AND rel.member_id = ? WHERE (relationship_id > 0 OR book.privately_owned = 0) AND (book.label LIKE ? or book.book_name like ?) limit 1"

		if err = o.Raw(sql1, member_id, keyword, keyword).QueryRow(&totalCount); err != nil {
			return
		}
		sql2 := `SELECT book.*,rel1.*,member.account AS create_name FROM md_books AS book
			LEFT JOIN md_relationship AS rel ON rel.book_id = book.book_id AND rel.member_id = ?
			LEFT JOIN md_relationship AS rel1 ON rel1.book_id = book.book_id AND rel1.role_id = 0
			LEFT JOIN md_members AS member ON rel1.member_id = member.member_id
			WHERE (rel.relationship_id > 0 OR book.privately_owned = 0) AND  (book.label LIKE ? or book.book_name like ?) ORDER BY order_index DESC ,book.book_id DESC LIMIT ?,?`

		_, err = o.Raw(sql2, member_id, keyword, keyword, offset, pageSize).QueryRows(&books)

		return

	} else {
		sql1 := "select COUNT(*) from md_books where privately_owned=0 and (label LIKE ? or book_name like ?) limit 1"
		if err = o.Raw(sql1, keyword, keyword).QueryRow(&totalCount); err != nil {
			return
		}

		sql := `SELECT book.*,rel.*,member.account AS create_name FROM md_books AS book
			LEFT JOIN md_relationship AS rel ON rel.book_id = book.book_id AND rel.role_id = 0
			LEFT JOIN md_members AS member ON rel.member_id = member.member_id
			WHERE book.privately_owned = 0 AND (book.label LIKE ? or book.book_name LIKE ?) ORDER BY order_index DESC ,book.book_id DESC LIMIT ?,?`

		_, err = o.Raw(sql, keyword, keyword, offset, pageSize).QueryRows(&books)

		return

	}
}

func (book *Book) ToBookResult() *BookResult {

	m := NewBookResult()

	m.BookId = book.BookId
	m.BookName = book.BookName
	m.Identify = book.Identify
	m.OrderIndex = book.OrderIndex
	m.Description = strings.Replace(book.Description, "\r\n", "<br/>", -1)
	m.PrivatelyOwned = book.PrivatelyOwned
	m.PrivateToken = book.PrivateToken
	m.DocCount = book.DocCount
	m.CommentStatus = book.CommentStatus
	m.CommentCount = book.CommentCount
	m.CreateTime = book.CreateTime
	m.ModifyTime = book.ModifyTime
	m.Cover = book.Cover
	m.MemberId = book.MemberId
	m.Label = book.Label
	m.Status = book.Status
	m.Editor = book.Editor
	m.Theme = book.Theme
	m.Vcnt = book.Vcnt
	m.Star = book.Star
	m.Score = book.Score
	m.ScoreFloat = utils.ScoreFloat(book.Score)
	m.CntScore = book.CntScore
	m.CntComment = book.CntComment

	if book.Theme == "" {
		m.Theme = "default"
	}
	if book.Editor == "" {
		m.Editor = "markdown"
	}
	return m
}

//重置文档数量
func (m *Book) ResetDocumentNumber(book_id int) {
	o := orm.NewOrm()

	totalCount, err := o.QueryTable(NewDocument().TableNameWithPrefix()).Filter("book_id", book_id).Count()
	if err == nil {
		o.Raw("UPDATE md_books SET doc_count = ? WHERE book_id = ?", int(totalCount), book_id).Exec()
	} else {
		beego.Error(err)
	}
}
