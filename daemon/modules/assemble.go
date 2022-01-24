/*
It is mainly used to assemble and transform the data used for restful request or response with other components.
*/
package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/utils"
	"strconv"
	"strings"
)

// Target ...
type Target struct {
	IPs    []string
	Params []map[string]map[string]interface{}
}

// mergeParam merge parameter for every ip
func (t *Target) mergeParam() error {
	ipExists := make(map[string]bool)
	t.IPs = config.KeenTune.Target.IPs
	t.Params = make([]map[string]map[string]interface{}, len(t.IPs))
	for _, group := range config.KeenTune.Group {
		for _, ip := range group.IPs {
			index := config.KeenTune.Target.IPMap[ip] - 1
			if index < 0 || index >= len(t.Params) {
				return fmt.Errorf("ip %v index %v is out of range", ip, index)
			}

			if !ipExists[ip] {
				ipExists[ip] = true
				t.Params[index] = group.ParamMap
				continue
			}

			t.Params[index] = mergeMap(t.Params[index], group.ParamMap)
		}
	}

	if len(t.IPs) != len(t.Params) {
		return fmt.Errorf("ips and param slice length [%v,%v] is not equal", len(t.IPs), len(t.Params))
	}

	return nil
}

// initApplyParam set request params for first target apply
func (t *Target) initApplyParam() error {
	for i := range t.Params {
		for domain, paramMap := range t.Params[i] {
			var temp = make(map[string]interface{})
			for name, value := range paramMap {
				param, ok := value.(map[string]interface{})
				if !ok {
					return fmt.Errorf("assert %v to parameter failed", value)
				}
				param["dtype"] = "read"
				param["value"] = ""
				delete(param, "options")
				delete(param, "range")
				delete(param, "step")
				temp[name] = param
			}

			t.Params[i][domain] = temp
		}
	}

	return nil
}

// getBrainInitParam get request parameters for brain init
func (t *Target) getBrainInitParam() ([]Parameter, error) {
	var initParam []Parameter
	var oneParam Parameter
	for i, group := range config.KeenTune.Group {
		paramSuffix := fmt.Sprintf("@group-%v", i+1)
		for domain, paramMap := range group.ParamMap {
			for name, value := range paramMap {
				var err error
				if err = utils.Map2Struct(value, &oneParam); err != nil {
					return nil, fmt.Errorf("map to struct: %v", err)
				}

				oneParam.ParaName = fmt.Sprintf("%v%v", name, paramSuffix)
				oneParam.DomainName = domain
				oneParam.Base, err = t.getBase(group.IPs[0], domain, name)
				if err != nil {
					return nil, fmt.Errorf("get base for brain init: %v", err)
				}

				initParam = append(initParam, oneParam)
			}
		}
	}

	return initParam, nil
}

func (t *Target) getBase(ip string, domain string, name string) (interface{}, error) {
	index := config.KeenTune.Target.IPMap[ip] - 1
	if index < 0 || index >= len(t.Params) {
		return nil, fmt.Errorf("ip %v index %v is out of range", ip, index)
	}

	param, ok := t.Params[index][domain][name]
	if !ok {
		return nil, fmt.Errorf("%v not found in %vth param", name, index)
	}

	return utils.ParseKey("value", param)
}

// parseApplyParam parse acquire response for non-init apply request
func (t *Target) parseApplyParam(resp []Parameter) error {
	for _, param := range resp {
		parts := strings.Split(param.ParaName, "@")
		if len(parts) != 2 {
			return fmt.Errorf("split %v by @group failed, length is not equal to 2", param.ParaName)
		}

		index, err := strconv.Atoi(parts[1])
		id := index - 1
		if err != nil || id < 0 || id >= len(config.KeenTune.Group) {
			return fmt.Errorf("group %v does not exist", parts[1])
		}

		paramName := parts[0]
		for _, ip := range config.KeenTune.Group[id].IPs {
			if err := t.updateValue(ip, paramName, param); err != nil {
				return fmt.Errorf("update %v value %v", paramName, err)
			}
		}
	}

	return nil
}

func (t *Target) updateValue(ip, name string, param Parameter) error {
	index := config.KeenTune.Target.IPMap[ip] - 1
	if index < 0 || index >= len(t.Params) {
		return fmt.Errorf("ip %v index %v is out of range", ip, index)
	}

	value, ok := t.Params[index][param.DomainName][name]
	if !ok {
		return fmt.Errorf("%v not found in %vth param", name, index)
	}

	detail, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("assert %v to parameter failed", param)
	}

	detail["value"] = param.Value
	detail["dtype"] = param.Dtype

	t.Params[index][param.DomainName][name] = detail
	return nil
}

func mergeMap(origin, addNew map[string]map[string]interface{}) map[string]map[string]interface{} {
	var retMap = make(map[string]map[string]interface{})
	for domain, value := range origin {
		retMap[domain] = value
	}

	for domain, value := range addNew {
		retMap[domain] = value
	}

	return retMap
}

func (t *Target) applyReq(index int, ip string) map[string]interface{} {
	retRequest := map[string]interface{}{}
	retRequest["data"] = t.Params[index]
	retRequest["resp_ip"] = config.RealLocalIP
	retRequest["resp_port"] = config.KeenTune.Port
	retRequest["target_id"] = config.KeenTune.IPMap[ip]
	return retRequest
}
