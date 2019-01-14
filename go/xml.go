package sql2xml

import (
	"strings"
)

// GenerateItemParams fills struct ItemStruct with ItemParamStruct structures in-place
// itemParam.Name has such format: `itemParamName1;;;itemParamName2...`
// itemParam.Value has such format: `itemParamValue1;;;itemParamValue2...`
// They create a pair of key-value structure.
// Resulting xml: <param name="key">value</param>
func GenerateItemParams(item ItemStruct, itemParam ItemParamStruct) {
	var paramsCount int
	itemParamNameArray := strings.Split(itemParam.Name, ";;;")
	itemParamValueArray := strings.Split(itemParam.Value, ";;;")

	// Ensure we have arrays with equal number of elements in case
	// DB has inconsistency.
	if len(itemParamNameArray) > len(itemParamValueArray) {
		itemParamNameArray = itemParamNameArray[0:len(itemParamValueArray)]
		paramsCount = len(itemParamValueArray)
	} else if len(itemParamNameArray) < len(itemParamValueArray) {
		itemParamValueArray = itemParamValueArray[0:len(itemParamNameArray)]
		paramsCount = len(itemParamNameArray)
	} else {
		paramsCount = len(itemParamNameArray)
	}

	if paramsCount > 0 {
		var i int
		var itemParams ItemParamStruct

		for i = 0; i < paramsCount; i++ {
			itemParams.Name = itemParamNameArray[i]
			itemParams.Value = itemParamValueArray[i]
			item.ItemParamsArray = append(item.ItemParamsArray, itemParams)
		}
	}
}
