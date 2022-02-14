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
	Port   string
}

const brainNameParts = 3

const (
	groupIDPrefix    = "group-"
	priorityIDPrefix = "pri-"
)

func (tuner *Tuner) initParams() error {
	var target Target
	var err error
	tuner.BrainParam = []Parameter{}
	for index, group := range config.KeenTune.Group {
		target.Params, err = getInitParam(index+1, group.ParamMap, &tuner.BrainParam)
		if err != nil {
			return err
		}

		target.IPs = group.IPs
		target.Port = group.Port
		tuner.Group = append(tuner.Group, target)
	}

	if len(tuner.Group) == 0 {
		return fmt.Errorf("found group is null")
	}

	tuner.mergeParam()

	return nil
}

func getInitParam(groupID int, paramMaps []map[string]map[string]interface{}, brainParam *[]Parameter) ([]map[string]map[string]interface{}, error) {
	var retMap = make([]map[string]map[string]interface{}, len(paramMaps))
	for i := range retMap {
		retMap[i] = make(map[string]map[string]interface{})
	}

	var oneParam Parameter
	for priID, paramMap := range paramMaps {
		for domain, params := range paramMap {
			var temp = make(map[string]interface{})
			for name, value := range params {
				param, ok := value.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("assert %v to parameter failed", value)
				}

				if err := utils.Map2Struct(value, &oneParam); err != nil {
					return nil, fmt.Errorf("map to struct: %v", err)
				}

				priID, ok := config.PriorityList[domain]
				if !ok {
					priID = 1
				}
				paramSuffix := fmt.Sprintf("@%v%v@%v%v", groupIDPrefix, groupID, priorityIDPrefix, priID)
				oneParam.ParaName = fmt.Sprintf("%v%v", name, paramSuffix)
				oneParam.DomainName = domain
				*brainParam = append(*brainParam, oneParam)
				delete(param, "options")
				delete(param, "range")
				delete(param, "step")
				temp[name] = param
			}

			retMap[priID][domain] = temp
		}
	}

	return retMap, nil
}

// getBrainInitParams get request parameters for brain init
func (tuner *Tuner) getBrainInitParams() error {
	for i := range tuner.BrainParam {
		name, groupID, _, err := parseBrainName(tuner.BrainParam[i].ParaName)
		if err != nil {
			return err
		}
		group := tuner.Group[groupID]

		tuner.BrainParam[i].Base, err = group.getBase(group.IPs[0], tuner.BrainParam[i].DomainName, name)
		if err != nil {
			return fmt.Errorf("get base for brain init: %v", err)
		}
	}

	return nil
}

func parseBrainName(originName string) (name string, groupID, priID int, err error) {
	names := strings.Split(originName, "@")
	if len(names) < brainNameParts {
		return "", 0, -1, fmt.Errorf("brain param name %v part length is not correct", originName)
	}

	name = names[0]

	groupIDStr := strings.TrimPrefix(names[1], groupIDPrefix)
	groupID, err = strconv.Atoi(groupIDStr)
	if groupID <= 0 || groupID > len(config.KeenTune.Group) {
		return "", 0, -1, fmt.Errorf("parse brain name groupID %v %v", groupIDStr, err)
	}

	priorityID := strings.TrimPrefix(names[2], priorityIDPrefix)
	priID, err = strconv.Atoi(priorityID)
	if priID < 0 || priID >= config.PRILevel {
		return "", 0, -1, fmt.Errorf("parse brain name priority ID %v %v", priorityID, err)
	}

	return name, groupID, priID, nil
}

func (t *Target) getBase(ip string, domain string, name string) (interface{}, error) {
	index := config.PriorityList[domain]
	if index < 0 || index >= config.PRILevel {
		return nil, fmt.Errorf("ip %v param priority index %v is out of range [0, 1]", ip, index)
	}

	param, ok := t.Params[index][domain][name]
	if !ok {
		return nil, fmt.Errorf("%v not found in %vth param", name, index)
	}

	return utils.ParseKey("value", param)
}

// parseAcquireParam parse acquire response value for apply request
func (t *Target) parseAcquireParam(resp []Parameter) error {
	for _, param := range resp {
		paramName, _, _, err := parseBrainName(param.ParaName)
		if err != nil {
			return err
		}

		for _, ip := range t.IPs {
			if err := t.updateValue(ip, paramName, param); err != nil {
				return fmt.Errorf("update %v value %v", paramName, err)
			}
		}
	}

	return nil
}

// updateApplyResult update param values by apply result
func (t *Target) updateApplyResult(ip string, params map[string]Parameter) error {
	for name, param := range params {
		err := t.updateValue(ip, name, param)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Target) updateValue(ip, name string, param Parameter) error {
	index := config.PriorityList[param.DomainName]
	if index < 0 || index >= config.PRILevel {
		return fmt.Errorf("ip %v priority id %v is out of range [0, 1]", ip, index)
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

	t.Params[index][param.DomainName][name] = detail
	return nil
}

func (tuner *Tuner) mergeParam() {
	tuner.MergedParam = make([]map[string]interface{}, len(tuner.Group))
	for index, target := range tuner.Group {
		tuner.MergedParam[index] = make(map[string]interface{})
		for _, paramMaps := range target.Params {
			for domain, paramMap := range paramMaps {
				tuner.MergedParam[index][domain] = paramMap
			}
		}
	}
}

func (t *Target) applyReq(index int, ip string) map[string]interface{} {
	retRequest := map[string]interface{}{}
	retRequest["data"] = t.Params[index]
	retRequest["resp_ip"] = config.RealLocalIP
	retRequest["resp_port"] = config.KeenTune.Port
	retRequest["target_id"] = config.KeenTune.IPMap[ip]
	return retRequest
}
