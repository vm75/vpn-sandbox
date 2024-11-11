package utils

import (
	"os"
	"strings"
)

type Option struct {
	isFlag     bool
	defaultVal string
	aliases    []string
	value      []string
}

func addOption(optString string, options map[string]*Option) {
	isFlag := !strings.HasSuffix(optString, ":")
	optString = strings.TrimSuffix(optString, ":")
	keyVal := strings.Split(optString, "=")
	defaultVal := ""
	if !isFlag && len(keyVal) > 1 {
		defaultVal = keyVal[1]
	}
	aliases := strings.Split(keyVal[0], "|")

	option := &Option{
		aliases:    aliases,
		isFlag:     isFlag,
		defaultVal: defaultVal,
		value:      []string{},
	}

	for _, alias := range aliases {
		options[alias] = option
	}
}

func (o *Option) needsValue() bool {
	return !o.isFlag
}

func (o *Option) setFlag() {
	if o.isFlag {
		o.value = []string{"true"}
	}
}

func (o *Option) setValue(value string) {
	o.value = append(o.value, value)
}

func (o *Option) IsSet() bool {
	return len(o.value) > 0
}

func (o *Option) GetValue() string {
	if len(o.value) == 0 {
		return o.defaultVal
	}
	return o.value[0]
}

func (o *Option) GetValues() []string {
	if len(o.value) == 0 && o.defaultVal != "" {
		return []string{o.defaultVal}
	}
	return o.value
}

func SmartArgs(optString string, args []string) (map[string]*Option, []string) {
	var (
		skippedArgs = []string{}
	)

	if args == nil {
		args = os.Args[1:]
	}

	options := make(map[string]*Option)

	if optString == "" {
		return options, args
	}

	for _, option := range strings.Split(optString, ",") {
		addOption(option, options)
	}

	for i := 0; i < len(args); i++ {
		option := args[i]

		// Stop parsing options after `--`
		if option == "--" {
			skippedArgs = append(skippedArgs, args[i+1:]...)
			break
		}

		// Parse options
		if options[option] != nil {
			if options[option].needsValue() {
				if i+1 >= len(args) {
					LogFatal("Missing argument for option " + option)
				}
				i++
				options[option].setValue(args[i])
			} else {
				options[option].setFlag()
			}
		} else {
			skippedArgs = append(skippedArgs, option)
		}
	}

	return options, skippedArgs
}
