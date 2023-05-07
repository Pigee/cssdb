package main

import (
	"log"
	"os/exec"
)

func main() {
	str := "mysqldump -u root -pSxxm@123 cs_s_run | mysql -u root -pSxxm@123 cs_s_run_sx"
	output, err := exec.Command("bash", "-c", str).Output()
	if err != nil {
		panic(err)
	}
	log.Println(string(output))
}
