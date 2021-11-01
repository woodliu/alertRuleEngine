package rule

import (
	"errors"
	vmCfg "github.com/VictoriaMetrics/VictoriaMetrics/app/vmalert/config"
	"gopkg.in/yaml.v2"
)

type File struct {
	Groups []FileGroup `yaml:"groups"`
}

type FileGroup struct {
	Name  string     `yaml:"name"`
	Rules []FileRule `yaml:"rules"`
}

type FileRule struct {
	Alert  string            `yaml:"alert"`
	Expr   string            `yaml:"expr"`
	Labels map[string]string `yaml:"labels"`
}

func vmGrp2File(grp *vmCfg.Group) ([]byte, error) {
	if nil == grp {
		return nil, errors.New("grp is nil")
	}

	var f File
	var fg FileGroup

	for _, v := range grp.Rules {
		vv := v
		fg.Rules = append(fg.Rules, FileRule{
			Alert:  vv.Name(),
			Expr:   vv.Expr,
			Labels: vv.Labels,
		})
	}

	fg.Name = grp.Name
	f.Groups = append(f.Groups, fg)

	return yaml.Marshal(&f)
}
