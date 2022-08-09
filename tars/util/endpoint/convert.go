package endpoint

import "github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"

// Tars2endpoint make endpointf.EndpointF to Endpoint struct.
func Tars2endpoint(end endpointf.EndpointF) Endpoint {
	proto := "tcp"
	if end.Istcp == UDP {
		proto = "udp"
	}
	e := Endpoint{
		Host:       end.Host,
		Port:       end.Port,
		Timeout:    end.Timeout,
		Istcp:      end.Istcp,
		Grid:       end.Grid,
		Qos:        end.Qos,
		Weight:     end.Weight,
		WeightType: end.WeightType,
		AuthType:   end.AuthType,
		Proto:      proto,
		Bind:       "",
		//Container: end.ContainerName,
		SetId: end.SetId,
	}
	e.Key = e.String()
	return e
}

// Endpoint2tars transfer Endpoint to endpointf.EndpointF
func Endpoint2tars(end Endpoint) endpointf.EndpointF {
	return endpointf.EndpointF{
		Host:       end.Host,
		Port:       end.Port,
		Timeout:    end.Timeout,
		Istcp:      end.Istcp,
		Grid:       end.Grid,
		Qos:        end.Qos,
		Weight:     end.Weight,
		WeightType: end.WeightType,
		AuthType:   end.AuthType,
		//ContainerName: end.Container,
		SetId: end.SetId,
	}
}

func IsEqual(a, b *[]endpointf.EndpointF) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil || len(*a) != len(*b) {
		return false
	}
	for i, x := range *a {
		y := (*b)[i]
		if x.Host != y.Host || x.Port != y.Port || x.Istcp != y.Istcp || x.WeightType != y.WeightType || x.Weight != y.Weight {
			return false
		}
	}
	return true
}
