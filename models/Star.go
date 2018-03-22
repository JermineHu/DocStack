package models

import (
	"fmt"
	"strconv"

	"github.com/astaxie/beego/orm"
)

type Star struct {
	Id  int
	Uid int `orm:"index"` //用户id,user id
	Bid int //书籍id,book id
}

// 多字段唯一键
func (this *Star) TableUnique() [][]string {
	return [][]string{
		[]string{"Uid", "Bid"},
	}
}

type StarResult struct {
	BookId      int    `json:"book_id"`
	BookName    string `json:"book_name"`
	Identify    string `json:"identify"`
	Description string `json:"description"`
	DocCount    int    `json:"doc_count"`
	Cover       string `json:"cover"`
	MemberId    int    `json:"member_id"`
	Nickname    string `json:"user_name"`
	Vcnt        int    `json:"vcnt"`
	Star        int    `json:"star"`
	Score       int    `json:"score"`
	CntComment  int    `json:"cnt_comment"`
	CntScore    int    `json:"cnt_score"`
	ScoreFloat  string `json:"score_float"`
}

//收藏或者取消收藏
//@param            uid         用户id
//@param            bid         书籍id
//@return           cancel      是否是取消收藏，只是标记属于取消还是收藏操作，err才表示执行操作成功与否
func (this *Star) Star(uid, bid int) (cancel bool, err error) {
	var star = Star{Uid: uid, Bid: bid}
	o := orm.NewOrm()
	qs := o.QueryTable("md_star")
	o.Read(&star, "Uid", "Bid")
	if star.Id > 0 { //取消收藏
		if _, err = qs.Filter("id", star.Id).Delete(); err == nil {
			SetIncreAndDecre("md_books", "star", fmt.Sprintf("book_id=%v and star>0", bid), false, 1)
		}
		cancel = true
	} else { //添加收藏
		cancel = false
		if _, err = o.Insert(&star); err == nil {
			//收藏计数+1
			SetIncreAndDecre("md_books", "star", "book_id="+strconv.Itoa(bid), true, 1)
		}
	}
	return
}

//是否收藏了文档
func (this *Star) DoesStar(uid, bid interface{}) bool {
	var star Star
	star.Uid, _ = strconv.Atoi(fmt.Sprintf("%v", uid))
	star.Bid, _ = strconv.Atoi(fmt.Sprintf("%v", bid))
	orm.NewOrm().Read(&star, "Uid", "Bid")
	if star.Id > 0 {
		return true
	}
	return false
}

//获取收藏列表，查询项目信息
func (this *Star) List(uid, p, listRows int) (cnt int64, books []StarResult, err error) {
	//根据用户id查询用户的收藏，先从收藏表中查询book_id
	o := orm.NewOrm()
	filter := o.QueryTable("md_star").Filter("uid", uid)
	//这里先暂时每次都统计一次用户的收藏数量。合理的做法是在用户表字段中增加一个收藏计数
	if cnt, _ = filter.Count(); cnt > 0 {
		sql := `select b.*,m.nickname from md_books b left join md_star s on s.bid=b.book_id left join md_members m on m.member_id=b.member_id where s.uid=? order by id desc limit %v offset %v`
		sql = fmt.Sprintf(sql, listRows, (p-1)*listRows)
		_, err = o.Raw(sql, uid).QueryRows(&books)
	}
	return
}
