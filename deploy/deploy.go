package deploy

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
)

const (
	K8S_SERVICES   = "Service"
	K8S_DEPLOYMENT = "Deployment"
)

type Deployer struct {
	client    dynamic.Interface
	namespace string
}

func NewDeployer(configPath, namespace string) (*Deployer, error) {
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	d := &Deployer{client: client, namespace: namespace}
	_, err = d.client.Resource(schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}).Namespace(d.namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil { // 判断连接是否可用
		return nil, err
	}
	return d, nil
}

func (d *Deployer) Apply(in []byte) error {
	objects, err := d.getObjects(in)
	if err != nil {
		return err
	}

	for _, obj := range objects {
		if err := d.apply(obj); err != nil {
			return err
		}
	}

	return nil
}

func (d *Deployer) getObjects(in []byte) ([]*unstructured.Unstructured, error) {
	parts := strings.Split(string(in), "---")
	var outs []*unstructured.Unstructured
	for _, part := range parts {
		deployment := make(map[string]interface{})
		if err := yaml.Unmarshal([]byte(part), &deployment); err != nil {
			return nil, err
		}

		if len(deployment) == 0 {
			continue
		}
		outs = append(outs, &unstructured.Unstructured{Object: deployment})
	}

	return outs, nil
}

func (d *Deployer) apply(obj *unstructured.Unstructured) error {
	var (
		group, version string
	)
	if gv := strings.Split(obj.GetAPIVersion(), "/"); len(gv) > 1 {
		group = gv[0]
		version = gv[1]
	} else if len(gv) == 1 {
		group = ""
		version = gv[0]
	}

	var deploymentRes = schema.GroupVersionResource{Group: group, Version: version}
	switch obj.GetKind() {
	case K8S_DEPLOYMENT:
		deploymentRes.Resource = "deployments"
	case K8S_SERVICES:
		deploymentRes.Resource = "services"
	default:
		return fmt.Errorf("unsupported resource type %v", obj.GetKind())
	}

	exist, err := d.checkExists(obj, deploymentRes)
	if err != nil {
		return err
	}

	if exist {
		if obj.GetKind() != K8S_SERVICES { // 如果资源不是 Service，则更新
			err = d.update(obj, deploymentRes)
		}
	} else {
		err = d.create(obj, deploymentRes)
	}

	return err
}

func (d *Deployer) checkExists(obj *unstructured.Unstructured, res schema.GroupVersionResource) (bool, error) {
	_, err := d.client.Resource(res).Namespace(d.namespace).Get(context.TODO(), obj.GetName(), v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (d *Deployer) create(obj *unstructured.Unstructured, res schema.GroupVersionResource) error {
	_, err := d.client.Resource(res).Namespace(d.namespace).Create(context.TODO(), obj, v1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (d *Deployer) update(obj *unstructured.Unstructured, res schema.GroupVersionResource) error {
	_, err := d.client.Resource(res).Namespace(d.namespace).Update(context.TODO(), obj, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
