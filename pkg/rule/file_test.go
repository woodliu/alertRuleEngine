package rule

import (
	"fmt"
	vmCfg "github.com/VictoriaMetrics/VictoriaMetrics/app/vmalert/config"
	"gopkg.in/yaml.v2"
	"testing"
)

func Test_vmGrp2File(t *testing.T) {
	vmGrp := vmCfg.Group{
		Name: "testVmGrp",
		Rules: []vmCfg.Rule{
			{
				Alert: "testAlert",
				Expr:  "sum(up)>100",
				Labels: map[string]string{
					"label1": "value1",
					"label2": "value2",
				},
			},
		},
	}

	data, err := vmGrp2File(&vmGrp)
	if nil != err {
		t.Fail()
	}

	var f File
	yaml.Unmarshal(data, &f)
	fmt.Println(f)

}
