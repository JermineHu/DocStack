package models

import (
	"strings"

	"github.com/JermineHu/DocStack/utils"
	"github.com/astaxie/beego/orm"
)

type AttachmentResult struct {
	Attachment
	IsExist       bool
	BookName      string
	DocumentName  string
	FileShortSize string
	Account       string
	LocalHttpPath string
}

func NewAttachmentResult() *AttachmentResult {
	return &AttachmentResult{IsExist: false}
}

func (m *AttachmentResult) Find(id int) (*AttachmentResult, error) {
	o := orm.NewOrm()

	attach := NewAttachment()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("attachment_id", id).One(attach)

	if err != nil {
		return m, err
	}

	m.Attachment = *attach

	book := NewBook()

	if e := o.QueryTable(book.TableNameWithPrefix()).Filter("book_id", attach.BookId).One(book, "book_name"); e == nil {
		m.BookName = book.BookName
	} else {
		m.BookName = "[不存在]"
	}
	doc := NewDocument()

	if e := o.QueryTable(doc.TableNameWithPrefix()).Filter("document_id", attach.DocumentId).One(doc, "document_name"); e == nil {
		m.DocumentName = doc.DocumentName
	} else {
		m.DocumentName = "[不存在]"
	}

	if attach.CreateAt > 0 {
		member := NewMember()
		if e := o.QueryTable(member.TableNameWithPrefix()).Filter("member_id", attach.CreateAt).One(member, "account"); e == nil {
			m.Account = member.Account
		}
	}
	m.FileShortSize = utils.FormatBytes(int64(attach.FileSize))
	m.LocalHttpPath = strings.Replace(m.FilePath, "\\", "/", -1)

	return m, nil
}
