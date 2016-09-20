package conf

import (
	"errors"
	"io/ioutil"
	"strings"
	"strconv"
	"reflect"
)

type Conf struct {
	curPtr 		*map[string]string
	basic 		map[string]string
	section 	map[string][]map[string]string
}

func NewConf() *Conf {
	return &Conf{
		basic: 	make(map[string]string),
		section: make(map[string][]map[string]string),
	}
}

type sectionConf struct {
	closed		bool
	name 		string
	data 		map[string]string
}

func initSection() *sectionConf {
	return &sectionConf{
		closed:true,
		name:"",
		data:make(map[string]string),
	}
}

func (this *Conf) LoadFile(configFile string) error{
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	section := initSection()
	dataSlice := strings.Split(string(data), "\n")
	for ln, line := range dataSlice {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' || len(line) <= 3 {
			continue
		}

		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return errors.New("line " + strconv.Itoa(ln) + ": invalid config syntax")
			}
			// end section
			if !section.closed {
				this.section[section.name] = append(this.section[section.name], section.data)
			}

			// start section
			section = initSection()
			section.closed = false
			section.name = strings.Trim(line, "[]")
			continue
		}

		lineSlice := strings.SplitN(line, "=", 2)
		if len(lineSlice) != 2 {
			return errors.New("line " + strconv.Itoa(ln) + ": invalid config syntax")
		}
		itemK := strings.TrimSpace(lineSlice[0])
		itemV := strings.TrimSpace(lineSlice[1])
		if !section.closed {
			section.data[itemK] = itemV
		}else {
			this.basic[itemK] = itemV
		}
	}

	// end the last section
	if !section.closed {
		this.section[section.name] = append(this.section[section.name], section.data)
	}

	return nil
}

func (this *Conf) Parse(obj interface{}) error {
	objT := reflect.TypeOf(obj)
	eT := objT.Elem()
	if objT.Kind() != reflect.Ptr || eT.Kind() != reflect.Struct {
		return errors.New("obj must be poiner to struct")
	}
	objV := reflect.ValueOf(obj)
	eV := objV.Elem()

	this.curPtr = &this.basic
	this.parseField(eT, eV)
	return nil
}

func (this *Conf) parseField(eT reflect.Type, eV reflect.Value) {
	for i := 0; i < eT.NumField(); i++ {
		f := eT.Field(i)
		k := string(f.Tag)
		if k == "" {
			k = f.Name
		}

		fV := eV.Field(i)
		if !fV.CanSet() {
			continue
		}

		switch (f.Type.Kind()) {
		case reflect.Bool:
			if v, e := this.getItemBool(k); e {
				fV.SetBool(v)
			}
		case reflect.Int:
			if v, e := this.getItemInt(k); e {
				fV.SetInt(v)
			}
		case reflect.String:
			if v, e := this.getItemString(k); e {
				fV.SetString(v)
			}
		/*case reflect.Slice:
			if f.Type.String() == "[]string" {
			}*/
		case reflect.Array:
			eT2 := eT
			eV2 := eV
			for idx := 0; idx < fV.Len() && idx < len(this.section[k]); idx++ {
				this.curPtr = &this.section[k][idx]
				eT = f.Type.Elem()
				eV = fV.Index(idx)

				this.parseField(eT, eV)
			}
			eT = eT2
			eV = eV2
		default:
		}
	}
}

func (this *Conf) getItemBool(name string) (bool, bool) {
	val := (*this.curPtr)[name]
	if val == "" {
		return false, false
	}

	v, _ := strconv.ParseBool(strings.ToLower(val))
	return v, true
}

func (this *Conf) getItemInt(name string) (int64, bool) {
	val, exist := (*this.curPtr)[name]
	if !exist {
		return 0, false
	}

	v, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}
	return int64(v), true
}

func (this *Conf) getItemString(name string) (string, bool) {
	val, exist := (*this.curPtr)[name]
	return val, exist
}