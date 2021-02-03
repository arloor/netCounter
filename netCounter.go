package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var netDevFile string

func init() {
	path := "/proc/net/dev"
	testPath := "/tmp/proc/net/dev"
	_, err := os.Stat(path)
	if err == nil {
		netDevFile = path
	} else {
		netDevFile = testPath
	}
}

func main() {
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/text; charset=utf-8")
		fmt.Fprint(w, strings.Join(fetchMetrics(), ""))
	})
	http.ListenAndServe("localhost:8080", nil)
}

type Metric struct {
	metricType string
	help       string
	value      int64
	name       string
	tags       map[string]string
}

func (m Metric) String() string {
	return fmt.Sprint("# HELP " + m.name + " " + m.help + "\n" +
		"# TYPE " + m.name + " " + m.metricType + "\n" +
		m.nameTags() + " " + strconv.FormatInt(m.value, 10) + "\n")
}

func (m Metric) nameTags() string {
	var str string = m.name + "{"
	for key := range m.tags {
		str += key + "=\"" + m.tags[key] + "\","
	}
	str += "}"
	return str
}

func fetchMetrics() []string {
	var metrics = make([]string, 1)
	if all, err := ioutil.ReadFile(netDevFile); err == nil {
		reader := bufio.NewReader(bytes.NewBuffer(all))
		reader.ReadString('\n')
		reader.ReadString('\n')
		for readString, err := reader.ReadString('\n'); err == nil; readString, err = reader.ReadString('\n') {
			reg, _ := regexp.Compile("(\\s)+")
			line := reg.ReplaceAllString(strings.TrimSpace(readString), " ")
			split := strings.Split(line, " ")
			if len(split) == 17 {
				inter := split[0][0 : len(split[0])-1]
				in, _ := strconv.ParseInt(split[1], 10, 64)
				out, _ := strconv.ParseInt(split[9], 10, 64)
				metrics = append(metrics, Metric{
					metricType: "counter",
					name:       inter + "_in_total",
					help:       "网卡流量",
					value:      in,
					tags:       map[string]string{},
				}.String())
				metrics = append(metrics, Metric{
					metricType: "counter",
					name:       inter + "_out_total",
					help:       "网卡流量",
					value:      out,
					tags:       map[string]string{},
				}.String())
			}
		}
	}
	return metrics
}
