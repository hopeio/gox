package flag

import (
	"errors"
	"os"
	"reflect"
	"strings"

	reflectx "github.com/hopeio/gox/reflect"
	"github.com/hopeio/gox/reflect/structtag"
	"github.com/hopeio/gox/strconv"
	"github.com/spf13/pflag"
)

const flagTagName = "flag"

// TODO: 优先级高于其他Config,覆盖环境变量及配置中心的配置
// example
/*type FlagConfig struct {
	// environment
	Env string `flag:"name:env;short:e;default:dev;usage:环境"`
	// 配置文件路径
	ConfPath string `flag:"name:conf;short:c;default:config.toml;usage:配置文件路径,默认./config.toml或./config/config.toml"`
}*/

type flagTagSettings struct {
	Name    string `meta:"name"`
	Short   string `meta:"short"`
	Env     string `meta:"env" comment:"从环境变量读取"`
	Default string `meta:"default"`
	Usage   string `meta:"usage"`
}

type anyValue reflect.Value

func (a anyValue) String() string {
	return strconv.ReflectFormat(reflect.Value(a))
}

func (a anyValue) Type() string {
	return reflect.Value(a).Kind().String()
}

func (a anyValue) Set(v string) error {
	return strconv.ParseReflectSet(reflect.Value(a), v, nil)
}

func Bind(args []string, v any) error {
	commandLine := pflag.NewFlagSet(args[0], pflag.ContinueOnError)
	commandLine.ParseErrorsWhitelist.UnknownFlags = true
	err := AddFlag(commandLine, v)
	if err != nil {
		return err
	}
	return commandLine.Parse(args[1:])
}

func AddFlag(commandLine *pflag.FlagSet, v any) error {
	fcValue := reflectx.DerefValue(reflect.ValueOf(v))
	if !fcValue.IsValid() {
		return errors.New("invalid value")
	}
	return AddFlagByReflectValue(commandLine, fcValue)
}

func AddFlagByReflectValue(commandLine *pflag.FlagSet, fcValue reflect.Value) error {
	fcTyp := fcValue.Type()
	for i := range fcTyp.NumField() {
		fieldType := fcTyp.Field(i)
		if !fieldType.IsExported() {
			continue
		}
		flagTag := fieldType.Tag.Get(flagTagName)
		fieldValue := fcValue.Field(i)
		kind := fieldValue.Kind()
		if kind == reflect.Pointer || kind == reflect.Interface {
			fieldValue = reflectx.DerefValue(fieldValue)
			kind = fieldValue.Kind()
			if !fieldValue.IsValid() {
				continue
			}
		}
		if flagTag != "" {
			var flagTagSettings flagTagSettings
			err := structtag.ParseSettingTagIntoStruct(flagTag, ';', &flagTagSettings)
			if err != nil {
				return err
			}
			// 从环境变量设置
			if flagTagSettings.Env != "" {
				if value, ok := os.LookupEnv(strings.ToUpper(flagTagSettings.Env)); ok {
					err := strconv.ParseReflectSet(fcValue.Field(i), value, nil)
					if err != nil {
						return err
					}
				}
			}
			if flagTagSettings.Name != "" {
				// flag设置
				flag := commandLine.VarPF(anyValue(fieldValue), flagTagSettings.Name, flagTagSettings.Short, flagTagSettings.Usage)
				if kind == reflect.Bool {
					flag.NoOptDefVal = "true"
				}
			}
		} else if kind == reflect.Struct {
			err := AddFlagByReflectValue(commandLine, fieldValue)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
