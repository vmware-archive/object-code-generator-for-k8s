// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT
package main_test

import (
	"bytes"
	"flag"
	"go/format"
	"io/ioutil"
	"os"
	"testing"
	"text/template"

	"github.com/vmware-tanzu-labs/object-code-generator-for-k8s/pkg/generate"
)

type source struct {
	Object string
}

func Test_main(t *testing.T) {
	t.Parallel()

	var manifestPath string

	var outputPath string

	flag.StringVar(&manifestPath, "manifest", "../../sample/deploy.yaml", "path to resource manifest")
	flag.StringVar(&outputPath, "output", "/tmp/ocgk-test.go", "path to output go source code")

	flag.Parse()

	tpl, err := template.New("testTemplate").Parse(testTemplate)
	if err != nil {
		t.Fatal(err)
	}

	manifestYaml, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	object, err := generate.Generate(manifestYaml, "deployment")
	if err != nil {
		t.Fatal(err)
	}

	src := source{Object: object}

	var buf bytes.Buffer

	if err = tpl.Execute(&buf, src); err != nil {
		t.Fatal(err)
	}

	fileSource, err := format.Source(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = f.Write(fileSource)
	if err != nil {
		t.Fatal(err)
	}
}

const testTemplate = `
package main

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := os.Getenv("KUBECONFIG")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		t.Fatal(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		t.Fatal(err)
	}

	namespace := "default"

	deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	{{ .Object }}

	result, err := client.Resource(deploymentRes).Namespace(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetName())
}
`
