package describe

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/chengyumeng/khadijah/pkg/model"
	"github.com/chengyumeng/khadijah/pkg/model/kubernetes"
	utillog "github.com/chengyumeng/khadijah/pkg/utils/log"
	"github.com/chengyumeng/khadijah/pkg/utils/stringobj"
	"github.com/ghodss/yaml"
	"github.com/olekukonko/tablewriter"
)

const YAML = "yaml"
const JSON = "json"
const PRETTY = "pretty"

var (
	DeploymentHeader = []string{"Name", "Namespace", "Cluster", "Labels", "Containers", "Replicas", "Message", "Pods"}
	ServiceHeader    = []string{"Name", "Namespace", "Cluster", "Labels", "Type", "ClusterIP", "EXTERNAL-IP", "Ports", "SELECTOR"}
	IngressHeader    = []string{"Name", "Namespace", "Cluster", "Labels", "HOSTS"}
	ConfigmapHeader  = []string{"Name", "Namespace", "Cluster", "Labels"}
	PodHeader        = []string{"Name", "Namespace", "Cluster", "PodIP", "Node", "Restart Time", "Start Time"}

	logger = utillog.NewAppLogger("pkg/describe")
)

type DescribeProxy struct {
	Option Option
}

func NewProxy(opt Option) DescribeProxy {
	return DescribeProxy{
		Option: opt,
	}
}

func (g *DescribeProxy) Describe() {
	if g.Option.Deployment != "" {
		g.Option.resource = model.DeploymentType
		g.showResourceState(g.Option.Deployment)
	} else if g.Option.Daemontset != "" {
		g.Option.resource = model.DaemonsetType
		g.showResourceState(g.Option.Daemontset)
	} else if g.Option.Statefulset != "" {
		g.Option.resource = model.StatefulsetType
		g.showResourceState(g.Option.Statefulset)
	} else if g.Option.Pod != "" {
		g.Option.resource = model.PodType
		g.showResourceState(g.Option.Pod)
	} else if g.Option.Service != "" {
		g.Option.resource = model.ServiceType
		g.showResourceState(g.Option.Service)
	} else if g.Option.Ingress != "" {
		g.Option.resource = model.IngressType
		g.showResourceState(g.Option.Ingress)
	} else if g.Option.Configmap != "" {
		g.Option.resource = model.ConfigmapType
		g.showResourceState(g.Option.Configmap)
	} else if g.Option.Pod != "" {
		g.Option.resource = model.PodType
		g.showResourceState(g.Option.Pod)
	}
}

func (g *DescribeProxy) showResourceState(name string) {
	data := model.GetNamespaceBody()
	nslist := []model.Namespace{}
	for _, ns := range data.Data.Namespaces {
		if ns.Name == g.Option.Namespace || g.Option.Namespace == "" {
			nslist = append(nslist, ns)
		}
	}
	tb := [][]string{}
	var header []string
	for _, ns := range nslist {
		kns := new(model.Metadata)
		err := json.Unmarshal([]byte(ns.Metadata), &kns)
		if err != nil {

		}
		for _, cluster := range kns.Clusters {
			if cluster == g.Option.Cluster || g.Option.Cluster == "" {
				data := kubernetes.GetResourceBody(name, int64(0), kns.Namespace, cluster, g.Option.resource, "")
				switch g.Option.Output {
				case YAML:
					data, err := yaml.JSONToYAML(data)
					if err != nil {
						logger.Errorln(err)
						return
					}
					fmt.Println(string(data))
				case JSON:
					fmt.Println(string(data))
				case PRETTY:
					switch g.Option.resource {
					case model.DeploymentType, model.DaemonsetType, model.StatefulsetType:
						pods := kubernetes.ListPods(int64(0), kns.Namespace, cluster, "?"+g.Option.resource+"="+g.Option.Deployment)
						arr := []string{}
						for _, p := range pods.Data {
							arr = append(arr, p.Name)
						}
						tb = append(tb, append(g.createDeploymentLine(data, cluster), strings.Join(arr, ",")))
						header = DeploymentHeader
					case model.ServiceType:
						header = ServiceHeader
						tb = append(tb, g.createServiceLine(data, cluster))
					case model.IngressType:
						header = IngressHeader
						tb = append(tb, g.createIngressLine(data, cluster))
					case model.ConfigmapType:
						header = ConfigmapHeader
						tb = append(tb, g.createConfigmapLine(data, cluster))
					case model.PodType:
						pods := kubernetes.GetPod(int64(0), kns.Namespace, cluster, g.Option.Pod)
						tb = append(tb, g.createPodLine(pods.Data, cluster))
						header = PodHeader
					default:
						fmt.Println(g.Option.resource)
					}
				}
			}
		}
	}
	if len(tb) > 0 {
		g.printTable(header, tb)
	}
}

func (g *DescribeProxy) printTable(header []string, lines [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.SetRowLine(true)
	table.SetRowSeparator("-")
	table.AppendBulk(lines)
	table.Render()
}

func (g *DescribeProxy) createPodLine(pod *kubernetes.Pod, cluster string) []string {
	status := []string{}
	for _, s := range pod.ContainerStatus {
		status = append(status, fmt.Sprintf("%s:%d", s.Name, s.RestartCount))
	}
	return []string{
		pod.Name, pod.Namespace, cluster, pod.PodIp, pod.NodeName, strings.Join(status, ","), pod.StartTime.String(),
	}
}

func (g *DescribeProxy) createDeploymentLine(data []byte, cluster string) []string {
	obj := new(kubernetes.DeploymentBody)
	err := json.Unmarshal(data, &obj)
	if err != nil {
		logger.Errorln(err)
	}
	ic := make(map[string]string)
	for _, c := range obj.Data.Spec.Template.Spec.Containers {
		ic[c.Name] = c.Image
	}
	rc := fmt.Sprintf("%d/%d", obj.Data.Status.AvailableReplicas, obj.Data.Status.Replicas)
	msg := make(map[string]string)
	for _, c := range obj.Data.Status.Conditions {
		msg[c.LastUpdateTime.Local().String()] = c.Message
	}
	return []string{obj.Data.Name, obj.Data.Namespace, cluster, stringobj.Map2list(obj.Data.Labels), stringobj.Map2list(ic), rc, stringobj.Map2list(msg)}
}

func (g *DescribeProxy) createServiceLine(data []byte, cluster string) []string {
	obj := new(kubernetes.ServiceBody)
	err := json.Unmarshal(data, &obj)
	if err != nil {
		logger.Errorln(err)
	}
	ps := []string{}
	for _, port := range obj.Data.Spec.Ports {
		ps = append(ps, fmt.Sprintf("%d:%d/%s", port.Port, port.TargetPort.IntVal, port.Protocol))
	}
	return []string{obj.Data.Name,
		obj.Data.Namespace, cluster,
		stringobj.Map2list(obj.Data.Labels),
		fmt.Sprintf("%v", obj.Data.Spec.Type),
		obj.Data.Spec.ClusterIP, strings.Join(obj.Data.Spec.ExternalIPs, ","),
		strings.Join(ps, ","), stringobj.Map2list(obj.Data.Spec.Selector)}
}

func (g *DescribeProxy) createIngressLine(data []byte, cluster string) []string {
	obj := new(kubernetes.IngressBody)
	err := json.Unmarshal(data, &obj)
	if err != nil {
		logger.Errorln(err)
	}
	hosts := []string{}
	for _, r := range obj.Data.Spec.Rules {
		hosts = append(hosts, r.Host)
	}
	return []string{obj.Data.Name,
		obj.Data.Namespace, cluster,
		stringobj.Map2list(obj.Data.Labels), strings.Join(hosts, ",")}
}

func (g *DescribeProxy) createConfigmapLine(data []byte, cluster string) []string {
	obj := new(kubernetes.ConfigmapBody)
	err := json.Unmarshal(data, &obj)
	if err != nil {
		logger.Errorln(err)
	}
	return []string{obj.Data.ObjectMeta.Name,
		obj.Data.ObjectMeta.Namespace, cluster,
		stringobj.Map2list(obj.Data.ObjectMeta.Labels)}
}
