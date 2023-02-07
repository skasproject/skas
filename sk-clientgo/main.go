package main

import (
	"skas/sk-clientgo/cmd"
)

func main() {
	cmd.Execute()
	//fmt.Println("Hello...")
	//
	//loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	//// if you want to change the loading rules (which files in which order), you can do so here
	//
	//configOverrides := &clientcmd.ConfigOverrides{}
	//// if you want to change override values or bind them to flags, there are methods to help you
	//
	//kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	//
	//rawConfig, err := kubeConfig.RawConfig()
	//if err != nil {
	//	panic(err)
	//}
	//explicitFile := kubeConfig.ConfigAccess().GetExplicitFile()
	//defaultFile := kubeConfig.ConfigAccess().GetDefaultFilename()
	//
	//fmt.Printf("context:%s  explicitFile:%s, defaultFile:%s\n", rawConfig.CurrentContext, explicitFile, defaultFile)

}
