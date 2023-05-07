package main

import (
	"log"
	"net/http"
	"os/exec"
)

func main() {
	http.HandleFunc("/initdb", orgHandler)

	log.Println("Listening...")
	log.Fatal(http.ListenAndServe(":8888", nil))

}

func orgHandler(w http.ResponseWriter, r *http.Request) {

	keys, ok := r.URL.Query()["org_no"]
	org_no := "defaultorgno"

	if ok {
		org_no = keys[0]
	}

	str := "mysqldump -u root -pSxxm@123 cs_s_run | mysql -u root -pSxxm@123 cs_s_run_" + org_no
	log.Println(str)

	output, err := exec.Command("bash", "-c", str).Output()
	if err != nil {
		panic(err)
	}
	log.Println(string(output))

}
