package commands

import (
	"encoding/gob"
	"fmt"
	"net/url"
	"os"
	"time"

	"flag"
	"path/filepath"
	"strings"

	"encoding/json"
	"log"

	"github.com/JermineHu/DocStack/commands/migrate"
	"github.com/JermineHu/DocStack/conf"
	"github.com/JermineHu/DocStack/models"
	"github.com/JermineHu/DocStack/utils"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/lifei6671/gocaptcha"
)

var (
	ConfigurationFile = "./conf/app.conf"
	WorkingDirectory  = "./"
	LogFile           = "./logs"
)

// RegisterDataBase 注册数据库
func RegisterDataBase() {

	host := beego.AppConfig.String("db_host")
	database := beego.AppConfig.String("db_database")
	username := beego.AppConfig.String("db_username")
	password := beego.AppConfig.String("db_password")
	timezone := beego.AppConfig.String("timezone")

	port := beego.AppConfig.String("db_port")

	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=%s", username, password, host, port, database, url.QueryEscape(timezone))

	orm.RegisterDataBase("default", "mysql", dataSource)

	if beego.AppConfig.String("runmode") == "dev" {
		orm.Debug = true
	}

	location, err := time.LoadLocation(timezone)
	if err == nil {
		orm.DefaultTimeLoc = location
	} else {
		log.Fatalln(err)
	}

}

// RegisterModel 注册Model
func RegisterModel() {
	orm.RegisterModelWithPrefix(conf.GetDatabasePrefix(),
		new(models.Member),
		new(models.Book),
		new(models.Relationship),
		new(models.Option),
		new(models.Document),
		new(models.Attachment),
		new(models.Logger),
		new(models.MemberToken),
		new(models.DocumentHistory),
		new(models.Migration),
		new(models.Label),
		new(models.Seo),
		new(models.Star),
		new(models.Score),
		new(models.Comments),
		new(models.Gitee),
		new(models.Github),
		new(models.QQ),
		new(models.DocumentStore),
	)
	migrate.RegisterMigration()
}

// RegisterLogger 注册日志
func RegisterLogger(log string) {

	logs.SetLogFuncCall(true)
	logs.SetLogger("console")
	logs.EnableFuncCallDepth(true)
	logs.Async()

	logPath := filepath.Join(log, "log.log")

	if _, err := os.Stat(logPath); os.IsNotExist(err) {

		os.MkdirAll(log, 0777)

		if f, err := os.Create(logPath); err == nil {
			f.Close()
			config := make(map[string]interface{}, 1)

			config["filename"] = logPath

			b, _ := json.Marshal(config)

			beego.SetLogger("file", string(b))
		}
	}

	beego.SetLogFuncCall(true)
	beego.BeeLogger.Async()
}

// RunCommand 注册orm命令行工具
func RegisterCommand() {

	if len(os.Args) >= 2 && os.Args[1] == "install" {
		ResolveCommand(os.Args[2:])
		Install()
	} else if len(os.Args) >= 2 && os.Args[1] == "version" {
		ResolveCommand(os.Args[2:])
		CheckUpdate()
	} else if len(os.Args) >= 2 && os.Args[1] == "migrate" {
		ResolveCommand(os.Args[2:])
		migrate.RunMigration()
	}
}

func RegisterFunction() {
	beego.AddFuncMap("config", models.GetOptionValue)

	beego.AddFuncMap("cdn", func(p string) string {
		cdn := beego.AppConfig.DefaultString("cdn", "")
		if strings.HasPrefix(p, "/") && strings.HasSuffix(cdn, "/") {
			return cdn + string(p[1:])
		}
		if !strings.HasPrefix(p, "/") && !strings.HasSuffix(cdn, "/") {
			return cdn + "/" + p
		}
		return cdn + p
	})

	beego.AddFuncMap("cdnjs", func(p string) string {
		cdn := beego.AppConfig.DefaultString("cdnjs", "")
		if strings.HasPrefix(p, "/") && strings.HasSuffix(cdn, "/") {
			return cdn + string(p[1:])
		}
		if !strings.HasPrefix(p, "/") && !strings.HasSuffix(cdn, "/") {
			return cdn + "/" + p
		}
		return cdn + p
	})
	beego.AddFuncMap("cdncss", func(p string) string {
		cdn := beego.AppConfig.DefaultString("cdncss", "")
		if strings.HasPrefix(p, "/") && strings.HasSuffix(cdn, "/") {
			return cdn + string(p[1:])
		}
		if !strings.HasPrefix(p, "/") && !strings.HasSuffix(cdn, "/") {
			return cdn + "/" + p
		}
		return cdn + p
	})
	beego.AddFuncMap("cdnimg", func(p string) string {
		cdn := beego.AppConfig.DefaultString("cdnimg", "")
		if strings.HasPrefix(p, "/") && strings.HasSuffix(cdn, "/") {
			return cdn + string(p[1:])
		}
		if !strings.HasPrefix(p, "/") && !strings.HasSuffix(cdn, "/") {
			return cdn + "/" + p
		}
		return cdn + p
	})
	beego.AddFuncMap("getUsernameByUid", func(id interface{}) string {
		return new(models.Member).GetUsernameByUid(id)
	})
	beego.AddFuncMap("getNicknameByUid", func(id interface{}) string {
		return new(models.Member).GetNicknameByUid(id)
	})
	//将标签转成a链接
	beego.AddFuncMap("tagsToLink", func(tags string) (links string) {
		var linkArr []string
		if slice := strings.Split(tags, ","); len(slice) > 0 {
			for _, tag := range slice {
				if tag = strings.TrimSpace(tag); len(tag) > 0 {
					linkArr = append(linkArr, fmt.Sprintf(`<a target="_blank" title="%v" href="%v">%v</a>`, tag, beego.URLFor("LabelController.Index", ":key", tag), tag))
				}
			}
		}
		return strings.Join(linkArr, "")
	})

	//用户是否收藏了文档
	beego.AddFuncMap("doesStar", new(models.Star).DoesStar)
	beego.AddFuncMap("scoreFloat", utils.ScoreFloat)
	beego.AddFuncMap("showImg", utils.ShowImg)
}

func ResolveCommand(args []string) {
	flagSet := flag.NewFlagSet("DocStack command: ", flag.ExitOnError)
	flagSet.StringVar(&ConfigurationFile, "config", "", "DocStack configuration file.")
	flagSet.StringVar(&WorkingDirectory, "dir", "", "DocStack working directory.")
	flagSet.StringVar(&LogFile, "log", "", "DocStack log file path.")

	flagSet.Parse(args)

	if WorkingDirectory == "" {
		if p, err := filepath.Abs(os.Args[0]); err == nil {
			WorkingDirectory = filepath.Dir(p)
		}
	}
	if LogFile == "" {
		LogFile = filepath.Join(WorkingDirectory, "logs")
	}
	if ConfigurationFile == "" {
		ConfigurationFile = filepath.Join(WorkingDirectory, "conf", "app.conf")
		config := filepath.Join(WorkingDirectory, "conf", "app.conf.example")
		if !utils.FileExists(ConfigurationFile) && utils.FileExists(config) {
			utils.CopyFile(ConfigurationFile, config)
		}
	}
	gocaptcha.ReadFonts(filepath.Join(WorkingDirectory, "static", "fonts"), ".ttf")

	err := beego.LoadAppConfig("ini", ConfigurationFile)

	if err != nil {
		log.Println("An error occurred:", err)
		os.Exit(1)
	}
	uploads := filepath.Join(WorkingDirectory, "uploads")

	os.MkdirAll(uploads, 0666)

	beego.BConfig.WebConfig.StaticDir["/static"] = filepath.Join(WorkingDirectory, "static")
	beego.BConfig.WebConfig.StaticDir["/uploads"] = uploads
	beego.BConfig.WebConfig.ViewsPath = filepath.Join(WorkingDirectory, "views")

	fonts := filepath.Join(WorkingDirectory, "static", "fonts")

	if !utils.FileExists(fonts) {
		log.Fatal("Font path not exist.")
	}
	gocaptcha.ReadFonts(filepath.Join(WorkingDirectory, "static", "fonts"), ".ttf")

	RegisterDataBase()
	RegisterModel()
	RegisterLogger(LogFile)
}

func init() {

	gocaptcha.ReadFonts("./static/fonts", ".ttf")
	gob.Register(models.Member{})

	if p, err := filepath.Abs(os.Args[0]); err == nil {
		WorkingDirectory = filepath.Dir(p)
	}
}
