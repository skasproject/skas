package misc

import (
	"github.com/spf13/pflag"
)

// For all adjustConfigXxx(), we:
// - panic when error is internal
// - Display a message and exit(2) when error is from usage

func AdjustConfigString(flagSet *pflag.FlagSet, inConfig *string, param string) {
	if pflag.Lookup(param).Changed {
		var err error
		if *inConfig, err = flagSet.GetString(param); err != nil {
			panic(err)
		}
		//fmt.Printf("param(%s).Changed: value: %s\n",param, *inConfig)
	} else if *inConfig == "" {
		*inConfig = flagSet.Lookup(param).DefValue
		//fmt.Printf("param(%s).Default: value: %s\n",param, *inConfig)
	}
	
}
