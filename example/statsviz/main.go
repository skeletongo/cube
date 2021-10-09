package main

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/skeletongo/cube/pkg"
	"github.com/skeletongo/cube/statsviz"
)

func Work() {
	m := make(map[int]interface{})
	for i := 0; ; i++ {

		var obj interface{}
		switch i % 6 {
		case 0:
			obj = &struct {
				_ uint32
				_ uint16
			}{}
		case 1:
			obj = &struct {
				_ [3]uint64
			}{}
		case 2:
			obj = fmt.Sprint("a relatively long and useless string %d", i)
		case 3:
			obj = make([]byte, i%1024)
		case 4:
			obj = make([]byte, 10*i%1024)
		case 5:
			obj = make([]string, 512)
		}

		if i == 1000 {
			m = make(map[int]interface{})
			i = 0
		}

		m[i] = obj
		time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	go Work()

	log.SetLevel(log.TraceLevel)
	pkg.RegisterPackage(statsviz.Config)
	pkg.Load("config.json")
	select {}
}
