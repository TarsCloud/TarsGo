package polaris

import (
	"context"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"
	"github.com/TarsCloud/TarsGo/tars/registry"
	"github.com/TarsCloud/TarsGo/tars/util/endpoint"
	"github.com/polarismesh/polaris-go"
)

const (
	endpointMeta = "endpoint"
)

type polarisRegistry struct {
	namespace string
	provider  polaris.ProviderAPI
	consumer  polaris.ConsumerAPI
}

type RegistryOption func(pr *polarisRegistry)

func WithNamespace(namespace string) RegistryOption {
	return func(pr *polarisRegistry) {
		pr.namespace = namespace
	}
}

func New(provider polaris.ProviderAPI, opts ...RegistryOption) registry.Registry {
	consumer := polaris.NewConsumerAPIByContext(provider.SDKContext())
	pr := &polarisRegistry{namespace: "tars", provider: provider, consumer: consumer}
	for _, opt := range opts {
		opt(pr)
	}
	//pr.addMiddleware()
	return pr
}

/*func (pr *polarisRegistry) addMiddleware() {
	tars.UseClientFilterMiddleware(func(next tars.ClientFilter) tars.ClientFilter {
		return func(ctx context.Context, msg *tars.Message, invoke tars.Invoke, timeout time.Duration) (err error) {
			start := time.Now()
			defer func() {
				delay := time.Since(start)
				retStatus := model.RetSuccess
				if msg.Resp.IRet != 0 {
					retStatus = model.RetFail
				}
				ret := &polaris.ServiceCallResult{
					ServiceCallResult: model.ServiceCallResult{
						EmptyInstanceGauge: model.EmptyInstanceGauge{},
						CalledInstance:     nil, // todo: 怎么获取到或构造 Instance
						Method:             msg.Req.SServantName + "." + msg.Req.SFuncName,
						RetStatus:          retStatus,
					},
				}
				ret.SetDelay(delay)
				ret.SetRetCode(msg.Resp.IRet)
				if er := pr.consumer.UpdateServiceCallResult(ret); er != nil {
					TLOG.Errorf("do report service call result : %+v", er)
				}
			}()
			return next(ctx, msg, invoke, timeout)
		}
	})
}*/

func (pr *polarisRegistry) Registry(_ context.Context, servant *registry.ServantInstance) error {
	instance := &polaris.InstanceRegisterRequest{}
	instance.Host = servant.Endpoint.Host
	instance.Port = int(servant.Endpoint.Port)
	instance.Protocol = &servant.Protocol
	instance.Namespace = pr.namespace
	instance.Service = servant.Servant
	if servant.Endpoint.Weight > 0 {
		weight := int(servant.Endpoint.Weight)
		instance.Weight = &weight
	}
	if servant.Endpoint.Timeout > 0 {
		timeout := time.Duration(servant.Endpoint.Timeout) * time.Millisecond
		instance.Timeout = &timeout
	}
	instance.Metadata = createMetadata(servant)
	_, err := pr.provider.RegisterInstance(instance)
	return err
}

func (pr *polarisRegistry) Deregister(_ context.Context, servant *registry.ServantInstance) error {
	instance := &polaris.InstanceDeRegisterRequest{}
	instance.Namespace = pr.namespace
	instance.Service = servant.Servant
	instance.Host = servant.Endpoint.Host
	instance.Port = int(servant.Endpoint.Port)
	if servant.Endpoint.Timeout > 0 {
		timeout := time.Duration(servant.Endpoint.Timeout) * time.Millisecond
		instance.Timeout = &timeout
	}
	err := pr.provider.Deregister(instance)
	return err
}

func (pr *polarisRegistry) QueryServant(_ context.Context, id string) (activeEp []endpointf.EndpointF, inactiveEp []endpointf.EndpointF, err error) {
	req := &polaris.GetAllInstancesRequest{}
	req.Namespace = pr.namespace
	req.Service = id
	resp, err := pr.consumer.GetAllInstances(req)
	if err != nil {
		return nil, nil, err
	}
	instances := resp.GetInstances()
	for _, ins := range instances {
		ep := endpoint.Parse(ins.GetMetadata()[endpointMeta])
		ep.Host = ins.GetHost()
		ep.Port = int32(ins.GetPort())
		epf := endpoint.Endpoint2tars(ep)
		if ins.IsHealthy() {
			activeEp = append(activeEp, epf)
		} else {
			inactiveEp = append(inactiveEp, epf)
		}
	}
	return activeEp, inactiveEp, err
}

func (pr *polarisRegistry) QueryServantBySet(_ context.Context, id, setId string) (activeEp []endpointf.EndpointF, inactiveEp []endpointf.EndpointF, err error) {
	req := &polaris.GetInstancesRequest{}
	req.Namespace = pr.namespace
	req.Service = id
	req.Metadata = map[string]string{
		"internal-enable-set": "Y",
		"internal-set-name":   setId,
	}
	resp, err := pr.consumer.GetInstances(req)
	if err != nil {
		return nil, nil, err
	}
	instances := resp.GetInstances()
	for _, ins := range instances {
		ep := endpoint.Parse(ins.GetMetadata()[endpointMeta])
		ep.Host = ins.GetHost()
		ep.Port = int32(ins.GetPort())
		epf := endpoint.Endpoint2tars(ep)
		if ins.IsHealthy() {
			activeEp = append(activeEp, epf)
		} else {
			inactiveEp = append(inactiveEp, epf)
		}
	}
	return activeEp, inactiveEp, err
}

func createMetadata(servant *registry.ServantInstance) map[string]string {
	metadata := make(map[string]string)
	metadata["tarsVersion"] = servant.TarsVersion
	metadata["app"] = servant.App
	metadata["server"] = servant.Server
	metadata[endpointMeta] = servant.Endpoint.String()
	// polaris plugin
	metadata["internal-enable-set"] = "N"
	if servant.EnableSet {
		metadata["internal-enable-set"] = "Y"
		metadata["internal-set-name"] = servant.SetDivision
	}
	return metadata
}