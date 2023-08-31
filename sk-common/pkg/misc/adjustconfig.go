package misc

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"strconv"
	"strings"
	"time"
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

func AdjustConfigStringArray(flagSet *pflag.FlagSet, inConfig *[]string, param string) {
	if pflag.Lookup(param).Changed {
		var s string
		var err error
		if s, err = flagSet.GetString(param); err != nil {
			panic(err)
		}
		*inConfig = strings.Split(s, ",")
		//fmt.Printf("param(%s).Changed: value: %s\n",param, *inConfig)
	} else if len(*inConfig) == 0 {
		*inConfig = strings.Split(flagSet.Lookup(param).DefValue, ",")
		//fmt.Printf("param(%s).Default: value: %s\n",param, *inConfig)
	}
}

func AdjustConfigDuration(flagSet *pflag.FlagSet, inConfig **time.Duration, param string) {
	var err error
	var durationStr string
	var duration time.Duration
	if flagSet.Lookup(param).Changed {
		if durationStr, err = flagSet.GetString(param); err != nil {
			panic(err)
		}
		if duration, err = time.ParseDuration(durationStr); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "\nInvalid %s value for a duration. Must be like 300s, 20m or 12h.\n\n", param)
			os.Exit(2)

		}
		*inConfig = &duration
	} else if *inConfig == nil {
		if duration, err = time.ParseDuration(flagSet.Lookup(param).DefValue); err != nil {
			panic(err)
		}
		*inConfig = &duration
	}
}

func AdjustConfigInt(flagSet *pflag.FlagSet, inConfig *int, param string) {
	var err error
	if flagSet.Lookup(param).Changed {
		if *inConfig, err = flagSet.GetInt(param); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "\nInvalid value for parameter %s\n", param)
			os.Exit(2)
		}
	} else if *inConfig == 0 {
		if *inConfig, err = strconv.Atoi(flagSet.Lookup(param).DefValue); err != nil {
			panic(err)
		}
	}
}
