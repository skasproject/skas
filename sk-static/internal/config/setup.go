package config

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"skas/sk-common/pkg/config"
	"skas/sk-common/pkg/misc"
)

func Setup() error {
	var configFile string
	var version bool
	var usersFile string
	var usersConfigMap string
	var logLevel string
	var logMode string

	pflag.StringVar(&configFile, "configFile", "config.yaml", "Configuration file")
	pflag.StringVar(&usersFile, "usersFile", "", "Users file")
	pflag.StringVar(&usersConfigMap, "usersConfigMap", "", "Users configMap")
	pflag.BoolVar(&version, "version", false, "Display version info")
	pflag.StringVar(&logLevel, "logLevel", "INFO", "Log level (PANIC|FATAL|ERROR|WARN|INFO|DEBUG|TRACE)")
	pflag.StringVar(&logMode, "logMode", "json", "Log mode: 'dev' or 'json'")

	pflag.Parse()

	// ------------------------------------ Version display
	if version {
		fmt.Printf("%s\n", config.Version)
		os.Exit(0)
	}

	// ------------------------------------ Load config file
	_, err := misc.LoadConfig(configFile, &Conf)
	if err != nil {
		return err
	}

	misc.AdjustConfigString(pflag.CommandLine, &Conf.Log.Mode, "logMode")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.Log.Level, "logLevel")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.UsersFile, "usersFile")
	misc.AdjustConfigString(pflag.CommandLine, &Conf.UsersConfigMap, "usersConfigMap")

	// -----------------------------------Handle logging  stuff
	Log, err = misc.HandleLog(&Conf.Log)
	if err != nil {
		return err
	}
	// ------------------------------------- Handle servers config
	// If the server list is empty, a first and only one is added.
	if Conf.Servers == nil || len(Conf.Servers) == 0 {
		Conf.Servers = []StaticServerConfig{StaticServerConfig{}}
	}
	for idx, _ := range Conf.Servers {
		Conf.Servers[idx].Default(7014 + (100 * idx))
	}

	// ------------------------------------- check users file
	if (Conf.UsersFile == "") == (Conf.UsersConfigMap == "") {
		return fmt.Errorf("one and only one of usersFile or usersConfigMap must be defined in configuration")
	}

	// --------------------------------------- Load users file
	//if err = loadUsers(usersFile); err != nil {
	//	return fmt.Errorf("file '%s': %w", usersFile, err)
	//}
	return nil
}

//
//func loadUsers(fileName string) error {
//	fn, err := filepath.Abs(fileName)
//	if err != nil {
//		return err
//	}
//	file, err := os.Open(fn)
//	if err != nil {
//		return err
//	}
//	decoder := yaml.NewDecoder(file)
//	decoder.SetStrict(true)
//	staticUsers := StaticUsers{}
//	if err = decoder.Decode(&staticUsers); err != nil {
//		return err
//	}
//	UserByLogin = make(map[string]StaticUser)
//	for idx, _ := range staticUsers.Users {
//		UserByLogin[staticUsers.Users[idx].Login] = staticUsers.Users[idx]
//	}
//	GroupsByUser = make(map[string][]string)
//	for _, gb := range staticUsers.GroupBindings {
//		u := gb.User
//		g := gb.Group
//		groups, ok := GroupsByUser[u]
//		if ok {
//			GroupsByUser[u] = append(groups, g)
//		} else {
//			GroupsByUser[u] = []string{g}
//		}
//	}
//	GroupBindingCount = len(staticUsers.GroupBindings)
//	return nil
//}
