package main

import (
	"flag"
	"fmt"
	"git.uinnova.com/thingyouwe-dockerfile/ci-deployer/deploy"
	"git.uinnova.com/thingyouwe-dockerfile/ci-deployer/environment"
	"gopkg.in/gookit/color.v1"
	"io/ioutil"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
)

var (
	info    = color.Notice.Render
	warn    = color.Warn.Render
	success = color.Success.Render
	danger  = color.Danger.Render

	version = "v0.0.1"
)

var usageStr = fmt.Sprintf(`
Version: %s

Usage: deploy [options] <subject>

Example: deploy -c ./kubeconfig.config ./example.yaml

Options:
	-c,  --config    <kubeconfig file path> Kubernetes client config
	-n,  --namespace <k8s namespace>
`, version)

func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

func fatal(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	var (
		kubeconfig string
		namespace  string
	)

	flag.StringVar(&kubeconfig, "c", "", "The Kubernetes client config file path")
	flag.StringVar(&kubeconfig, "config", "", "The Kubernetes client config file path")
	flag.StringVar(&namespace, "n", "default", "")
	flag.StringVar(&namespace, "namespace", "default", "")

	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		usage()
	}

	if len(kubeconfig) == 0 {
		fatal(fmt.Errorf("missing --config option"))
	}

	content, err := ioutil.ReadFile(args[0])
	if err != nil {
		fatal(err)
	}

	out, err := environment.HandleEnv(content)
	if err != nil {
		fatal(err)
	}

	client, err := deploy.NewDeployer(kubeconfig, namespace)
	if err != nil {
		fatal(err)
	}

	if err := client.Apply(out); err != nil {
		fatal(err)
	}
}
