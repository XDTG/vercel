package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var result string

func printEnvs() {
	envs := os.Environ()
	for _, env := range envs {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		} else {
			result += fmt.Sprintf("%s = %s\n", parts[0], parts[1])
		}
	}
	result += "\n\n"
}

func readResolv() {
	filepath := "/etc/resolv.conf"
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println("failed to read resolv.conf!")
		panic(err)
	}
	result += string(content)
	result += "\n\n"
}

func digServices() {
	installCmd := exec.Command("apt-get", "install", "dnsutils", "-y")
	installOut, installErr := installCmd.CombinedOutput()
	if installErr != nil {
		fmt.Println("failed to exec \"apt-get\" command!")
		result += "non apt-get info (" + installErr.Error() + ")\n"
	} else {
		result += string(installOut) + "\n"
	}

	search := "ec2.internal"
	cmd := exec.Command("dig", "+noall", "+answer", "srv", "any.any."+search)
	//cmd := exec.Command("lsb_release", "-a")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("failed to exec \"dig\" command!")
		result += "non dig info (" + err.Error() + ")\n"
	} else {
		result += string(out) + "\n"
	}

	result += "\n\n"
}

func printServices() {
	readResolv()
	//digServices()
}

func coreMetrics(dnsIp string) {
	if dnsIp == "" {
		return
	}

	url := "http://" + dnsIp + ":9153/metrics"
	fmt.Println("request url: " + url)

	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("failed to get dns metrics")
		fmt.Println(err)
		result += "not get \"" + url + "\" (" + err.Error() + ")\n"
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	result += "CoreDNS response: \n" + string(body) + "\n"
	result += "StatusCode: " + string(rune(resp.StatusCode)) + "\n"
}

func Handler(w http.ResponseWriter, r *http.Request) {
	result = ""

	printEnvs()

	printServices()

	dnsIp := ""
	coreMetrics(dnsIp)

	fmt.Fprintf(w, result)
}
