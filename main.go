package main

import (
	"database/sql"
	"encoding/json"
	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"os"
        "os/exec"
)

var infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
var errorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

// 配置文件数据结构
type Cssdbconf struct {
	Impstr  string
	Dbstr   string
	Postsql []string
}

// Post请求结构
type ReqData struct {
	Org_no       string `json:"org_no"`
	Account_name string `json:"account_name"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/initdb", orgHandler)

	infoLog.Println("Listening...")
	errorLog.Fatal(http.ListenAndServe(":8888", mux))
	// 获取应用配置
	//myconf := getToml()
	//DB, _ := sql.Open("mysql", "sxadmin:sx@123@tcp(192.168.199.100:3306)/cs_s_run")
	//createDb(myconf)

}

func orgHandler(w http.ResponseWriter, r *http.Request) {

	myconf := getToml()

	var reqdata ReqData
	// 调用json包的解析，解析请求body
	if err := json.NewDecoder(r.Body).Decode(&reqdata); err != nil {
		r.Body.Close()
		log.Fatal(err)
	}

	//jsonStr, _ := json.Marshal(reqdata)
	//infoLog.Println("req json: ", string(jsonStr))

	// 创建数据库
	err := createDb(myconf, reqdata.Org_no)
	if err != nil {
		return
	}

	// 导数据
	str := myconf.Impstr + reqdata.Org_no
	infoLog.Println(str)
	output, err := exec.Command("bash", "-c", str).Output()
	if err != nil {
		panic(err)
	}
	infoLog.Println(string(output))

	//后期处理
	err = flushDb(myconf, reqdata.Org_no)
	if err != nil {
		return
	}

	//创建cs_base_user纪录
	err = createBaseuser(myconf, reqdata.Org_no, reqdata.Account_name)
	if err != nil {
		return
	}

	//返回 status ok
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "Status OK"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		errorLog.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
	return

}

func getToml() (c Cssdbconf) {

	var conf Cssdbconf
	if _, err := toml.DecodeFile("./conf/cssdb.toml", &conf); err != nil {
		// handle error
		errorLog.Fatal(err)
	}
	return conf
}

func createDb(c Cssdbconf, o string) (e error) {
	infoLog.Println(c.Dbstr)
	infoLog.Println(len(c.Postsql))
	infoLog.Println(c.Postsql[0])
	DB, _ := sql.Open("mysql", c.Dbstr)
	defer DB.Close()

	DB.SetConnMaxLifetime(100)
	DB.SetMaxIdleConns(10)
	if err := DB.Ping(); err != nil {
		infoLog.Println("open database fail")
		return err
	}
	infoLog.Println("connnect cs_s_run database success...")

	sqlStr := "create database cs_s_run_" + o + " charset utf8mb4"
	ret, err := DB.Exec(sqlStr)
	if err != nil {
		errorLog.Println("create database failed ,err:%v\n", err)
		return err
	} else {
		i, _ := ret.RowsAffected()
		infoLog.Println("i: %v\n", i)
	}
	return

}

func flushDb(c Cssdbconf, o string) (e error) {
	infoLog.Println(c.Dbstr)
	DB, _ := sql.Open("mysql", c.Dbstr+"_"+o)
	defer DB.Close()

	DB.SetConnMaxLifetime(100)
	DB.SetMaxIdleConns(10)
	if err := DB.Ping(); err != nil {
		infoLog.Println("open database fail")
		return err
	}
	infoLog.Println("connnect New born database success...")

	for _, v := range c.Postsql {
		ret, err := DB.Exec(v + "'" + o + "'")
		if err != nil {
			errorLog.Println("Flush Db failed... ,err:%v\n", err)
			return err
		} else {
			i, _ := ret.RowsAffected()
			infoLog.Println("flushp database successfully...: %v\n", i)
		}
	}

	return

}

func createBaseuser(c Cssdbconf, o string, n string) (e error) {
	infoLog.Println(c.Dbstr)
	DB, _ := sql.Open("mysql", c.Dbstr+"_"+o)
	defer DB.Close()

	DB.SetConnMaxLifetime(100)
	DB.SetMaxIdleConns(10)
	if err := DB.Ping(); err != nil {
		infoLog.Println("open database fail")
		return err
	}
	infoLog.Println("connnect New born database success...")

	// infoLog.Println("insert into cs_base_user(id,org_no,userid,account,password,mobile,status,flag,if_admin,name) select id,org_no,'' userid,'','',mobile,1,0,1,'" + n + "' from cs_common_user")
	ret, err := DB.Exec("insert into cs_base_user(id,org_no,userid,account,password,mobile,status,flag,if_admin,name) select id,org_no,'' userid,'','',mobile,1,0,1,'" + o + "' from cs_common_user")
	if err != nil {
		errorLog.Println("create cs_base_user failed ,err:%v\n", err)
		return err
	} else {
		i, _ := ret.RowsAffected()
		infoLog.Println("create base user success: %v\n", i)
	}

	return

}
