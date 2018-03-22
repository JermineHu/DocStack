// Package models .
package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	ldap "gopkg.in/ldap.v2"

	"github.com/JermineHu/DocStack/conf"
	"github.com/JermineHu/DocStack/utils"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type Member struct {
	MemberId      int       `orm:"pk;auto;column(member_id)" json:"member_id"`
	Account       string    `orm:"size(30);unique;column(account)" json:"account"`
	Nickname      string    `orm:"size(30);unique;column(nickname)" json:"nickname"` //昵称
	Password      string    `orm:"column(password)" json:"-"`
	AuthMethod    string    `orm:"column(auth_method);default(local);size(50);" json:"auth_method"` //认证方式: local 本地数据库 /ldap LDAP
	Description   string    `orm:"column(description);size(2000)" json:"description"`
	Email         string    `orm:"size(100);column(email);unique" json:"email"`
	Phone         string    `orm:"size(255);column(phone);null;default(null)" json:"phone"`
	Avatar        string    `orm:"column(avatar)" json:"avatar"`
	Role          int       `orm:"column(role);type(int);default(1);index" json:"role"` //用户角色：0 超级管理员 /1 管理员/ 2 普通用户 .
	RoleName      string    `orm:"-" json:"role_name"`
	Status        int       `orm:"column(status);type(int);default(0)" json:"status"` //用户状态：0 正常/1 禁用
	CreateTime    time.Time `orm:"type(datetime);column(create_time);auto_now_add" json:"create_time"`
	CreateAt      int       `orm:"type(int);column(create_at)" json:"create_at"`
	LastLoginTime time.Time `orm:"type(datetime);column(last_login_time);null" json:"last_login_time"`
	Wxpay         string    `json:"wxpay"`  //微信支付的收款二维码
	Alipay        string    `json:"alipay"` //支付宝支付的收款二维码
}

// TableName 获取对应数据库表名.
func (m *Member) TableName() string {
	return "members"
}

// TableEngine 获取数据使用的引擎.
func (m *Member) TableEngine() string {
	return "INNODB"
}

func (m *Member) TableNameWithPrefix() string {
	return conf.GetDatabasePrefix() + m.TableName()
}

func NewMember() *Member {
	return &Member{}
}

// Login 用户登录.
func (m *Member) Login(account string, password string) (*Member, error) {
	o := orm.NewOrm()

	member := &Member{}

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("account", account).Filter("status", 0).One(member)

	if err != nil {
		if beego.AppConfig.DefaultBool("ldap_enable", false) == true {
			logs.Info("转入LDAP登陆")
			return member.ldapLogin(account, password)
		} else {
			logs.Error("用户登录 => ", err)
			return member, ErrMemberNoExist
		}
	}

	switch member.AuthMethod {
	case "":
	case "local":
		ok, err := utils.PasswordVerify(member.Password, password)
		if ok && err == nil {
			m.ResolveRoleName()
			return member, nil
		}
	case "ldap":
		return member.ldapLogin(account, password)
	default:
		return member, ErrMemberAuthMethodInvalid
	}

	return member, ErrorMemberPasswordError
}

//ldapLogin 通过LDAP登陆
func (m *Member) ldapLogin(account string, password string) (*Member, error) {
	if beego.AppConfig.DefaultBool("ldap_enable", false) == false {
		return m, ErrMemberAuthMethodInvalid
	}
	var err error
	lc, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", beego.AppConfig.String("ldap_host"), beego.AppConfig.DefaultInt("ldap_port", 3268)))
	if err != nil {
		return m, ErrLDAPConnect
	}
	defer lc.Close()
	err = lc.Bind(beego.AppConfig.String("ldap_user"), beego.AppConfig.String("ldap_password"))
	if err != nil {
		return m, ErrLDAPFirstBind
	}
	searchRequest := ldap.NewSearchRequest(
		beego.AppConfig.String("ldap_base"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		//修改objectClass通过配置文件获取值
		fmt.Sprintf("(&(%s)(%s=%s))", beego.AppConfig.String("ldap_filter"), beego.AppConfig.String("ldap_attribute"), account),
		[]string{"dn", "mail"},
		nil,
	)
	searchResult, err := lc.Search(searchRequest)
	if err != nil {
		return m, ErrLDAPSearch
	}
	if len(searchResult.Entries) != 1 {
		return m, ErrLDAPUserNotFoundOrTooMany
	}
	userdn := searchResult.Entries[0].DN
	err = lc.Bind(userdn, password)
	if err != nil {
		return m, ErrorMemberPasswordError
	}
	if m.Account == "" {
		m.Account = account
		m.Email = searchResult.Entries[0].GetAttributeValue("mail")
		m.AuthMethod = "ldap"
		m.Avatar = "/static/images/headimgurl.jpg"
		m.Role = beego.AppConfig.DefaultInt("ldap_user_role", 2)
		m.CreateTime = time.Now()

		err = m.Add()
		if err != nil {
			logs.Error("自动注册LDAP用户错误", err)
			return m, ErrorMemberPasswordError
		}
		m.ResolveRoleName()
	}
	return m, nil
}

// Add 添加一个用户.
func (m *Member) Add() error {
	o := orm.NewOrm()

	if ok, err := regexp.MatchString(conf.RegexpAccount, m.Account); m.Account == "" || !ok || err != nil {
		return errors.New("用户名只能由英文字母数字组成，且在3-50个字符")
	}
	if m.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if ok, err := regexp.MatchString(conf.RegexpEmail, m.Email); !ok || err != nil || m.Email == "" {
		return errors.New("邮箱格式不正确")
	}
	if m.AuthMethod == "local" {
		if l := strings.Count(m.Password, ""); l < 6 || l >= 50 {
			return errors.New("密码不能为空且必须在6-50个字符之间")
		}
	}
	cond := orm.NewCondition().Or("email", m.Email).Or("nickname", m.Nickname).Or("account", m.Account)
	var one Member
	if o.QueryTable(m.TableNameWithPrefix()).SetCond(cond).One(&one, "member_id", "nickname", "account", "email"); one.MemberId > 0 {
		if one.Nickname == m.Nickname {
			return errors.New("昵称已存在，请更换昵称")
		}
		if one.Email == m.Email {
			return errors.New("邮箱已被注册，请更换邮箱")
		}
		if one.Account == m.Account {
			return errors.New("用户名已存在，请更换用户名")
		}
	}
	//if c, err := o.QueryTable(m.TableNameWithPrefix()).Filter("email", m.Email).Count(); err == nil && c > 0 {
	//	return errors.New("邮箱已被使用")
	//}

	hash, err := utils.PasswordHash(m.Password)

	if err != nil {
		return err
	}

	m.Password = hash
	if m.AuthMethod == "" {
		m.AuthMethod = "local"
	}
	_, err = o.Insert(m)

	if err != nil {
		return err
	}
	m.ResolveRoleName()
	return nil
}

// Update 更新用户信息.
func (m *Member) Update(cols ...string) error {
	o := orm.NewOrm()

	if m.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if _, err := o.Update(m, cols...); err != nil {
		return err
	}
	return nil
}

func (m *Member) Find(id int) (*Member, error) {
	o := orm.NewOrm()

	m.MemberId = id
	if err := o.Read(m); err != nil {
		return m, err
	}
	m.ResolveRoleName()
	return m, nil
}

func (m *Member) ResolveRoleName() {
	if m.Role == conf.MemberSuperRole {
		m.RoleName = "超级管理员"
	} else if m.Role == conf.MemberAdminRole {
		m.RoleName = "管理员"
	} else if m.Role == conf.MemberGeneralRole {
		m.RoleName = "普通用户"
	}
}

//根据账号查找用户.
func (m *Member) FindByAccount(account string) (*Member, error) {
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter("account", account).One(m)

	if err == nil {
		m.ResolveRoleName()
	}
	return m, err
}

//分页查找用户.
func (m *Member) FindToPager(pageIndex, pageSize int) ([]*Member, int64, error) {
	o := orm.NewOrm()

	var members []*Member

	offset := (pageIndex - 1) * pageSize

	totalCount, err := o.QueryTable(m.TableNameWithPrefix()).Count()

	if err != nil {
		return members, 0, err
	}

	_, err = o.QueryTable(m.TableNameWithPrefix()).OrderBy("-member_id").Offset(offset).Limit(pageSize).All(&members)

	if err != nil {
		return members, 0, err
	}

	for _, m := range members {
		m.ResolveRoleName()
	}
	return members, totalCount, nil
}

func (c *Member) IsAdministrator() bool {
	if c == nil || c.MemberId <= 0 {
		return false
	}
	return c.Role == 0 || c.Role == 1
}

//根据指定字段查找用户.
func (m *Member) FindByFieldFirst(field string, value interface{}) (*Member, error) {
	o := orm.NewOrm()

	err := o.QueryTable(m.TableNameWithPrefix()).Filter(field, value).OrderBy("-member_id").One(m)

	return m, err
}

//校验用户.
func (m *Member) Valid(is_hash_password bool) error {

	//邮箱不能为空
	if m.Email == "" {
		return ErrMemberEmailEmpty
	}
	//用户描述必须小于500字
	if strings.Count(m.Description, "") > 500 {
		return ErrMemberDescriptionTooLong
	}
	if m.Role != conf.MemberGeneralRole && m.Role != conf.MemberSuperRole && m.Role != conf.MemberAdminRole {
		return ErrMemberRoleError
	}
	if m.Status != 0 && m.Status != 1 {
		m.Status = 0
	}
	//邮箱格式校验
	if ok, err := regexp.MatchString(conf.RegexpEmail, m.Email); !ok || err != nil || m.Email == "" {
		return ErrMemberEmailFormatError
	}
	//如果是未加密密码，需要校验密码格式
	if !is_hash_password {
		if l := strings.Count(m.Password, ""); m.Password == "" || l > 50 || l < 6 {
			return ErrMemberPasswordFormatError
		}
	}
	//校验邮箱是否呗使用
	if member, err := NewMember().FindByFieldFirst("email", m.Account); err == nil && member.MemberId > 0 {
		if m.MemberId > 0 && m.MemberId != member.MemberId {
			return ErrMemberEmailExist
		}
		if m.MemberId <= 0 {
			return ErrMemberEmailExist
		}
	}

	if m.MemberId > 0 {
		//校验用户是否存在
		if _, err := NewMember().Find(m.MemberId); err != nil {
			return err
		}
	} else {
		//校验账号格式是否正确
		if ok, err := regexp.MatchString(conf.RegexpAccount, m.Account); m.Account == "" || !ok || err != nil {
			return ErrMemberAccountFormatError
		}
		//校验账号是否被使用
		if member, err := NewMember().FindByAccount(m.Account); err == nil && member.MemberId > 0 {
			return ErrMemberExist
		}
	}

	return nil
}

//删除一个用户.

func (m *Member) Delete(oldId int, newId int) error {
	o := orm.NewOrm()

	err := o.Begin()

	if err != nil {
		return err
	}

	_, err = o.Raw("DELETE FROM md_members WHERE member_id = ?", oldId).Exec()
	if err != nil {
		o.Rollback()
		return err
	}
	_, err = o.Raw("UPDATE md_attachment SET `create_at` = ? WHERE `create_at` = ?", newId, oldId).Exec()

	if err != nil {
		o.Rollback()
		return err
	}

	_, err = o.Raw("UPDATE md_books SET member_id = ? WHERE member_id = ?", newId, oldId).Exec()
	if err != nil {
		o.Rollback()
		return err
	}
	_, err = o.Raw("UPDATE md_document_history SET member_id=? WHERE member_id = ?", newId, oldId).Exec()
	if err != nil {
		o.Rollback()
		return err
	}
	_, err = o.Raw("UPDATE md_document_history SET modify_at=? WHERE modify_at = ?", newId, oldId).Exec()
	if err != nil {
		o.Rollback()
		return err
	}
	_, err = o.Raw("UPDATE md_documents SET member_id = ? WHERE member_id = ?;", newId, oldId).Exec()
	if err != nil {
		o.Rollback()
		return err
	}
	_, err = o.Raw("UPDATE md_documents SET modify_at = ? WHERE modify_at = ?", newId, oldId).Exec()
	if err != nil {
		o.Rollback()
		return err
	}
	//_,err = o.Raw("UPDATE md_relationship SET member_id = ? WHERE member_id = ?",newId,oldId).Exec()
	//if err != nil {
	//
	//	if err != nil {
	//		o.Rollback()
	//		return err
	//	}
	//}
	var relationship_list []*Relationship

	_, err = o.QueryTable(NewRelationship().TableNameWithPrefix()).Filter("member_id", oldId).All(&relationship_list)

	if err == nil {
		for _, relationship := range relationship_list {
			//如果存在创始人，则删除
			if relationship.RoleId == 0 {
				rel := NewRelationship()

				err = o.QueryTable(relationship.TableNameWithPrefix()).Filter("book_id", relationship.BookId).Filter("member_id", newId).One(rel)
				if err == nil {
					if _, err := o.Delete(relationship); err != nil {
						beego.Error(err)
					}
					relationship.RelationshipId = rel.RelationshipId
				}
				relationship.MemberId = newId
				relationship.RoleId = 0
				if _, err := o.Update(relationship); err != nil {
					beego.Error(err)
				}
			} else {
				if _, err := o.Delete(relationship); err != nil {
					beego.Error(err)
				}
			}
		}
	}

	if err = o.Commit(); err != nil {
		o.Rollback()
		return err
	}
	return nil
}

//获取用户名
func (m *Member) GetUsernameByUid(id interface{}) string {
	var user Member
	orm.NewOrm().QueryTable("md_members").Filter("member_id", id).One(&user, "account")
	return user.Account
}

//获取昵称
func (m *Member) GetNicknameByUid(id interface{}) string {
	var user Member
	if err := orm.NewOrm().QueryTable("md_members").Filter("member_id", id).One(&user, "nickname"); err != nil {
		beego.Error(err.Error())
	}

	return user.Nickname
}

//根据用户id获取二维码
func (m *Member) GetQrcodeByUid(uid interface{}) (qrcode map[string]string) {
	var member Member
	orm.NewOrm().QueryTable("md_members").Filter("member_id", uid).One(&member, "alipay", "wxpay")
	qrcode = make(map[string]string)
	qrcode["Alipay"] = member.Alipay
	qrcode["Wxpay"] = member.Wxpay
	return qrcode
}
