// chooser
package Chooser

type strfunc func(code int) (res string)

type TRangeMinMax struct {
	minval, maxval int
	sf             strfunc
}

func B2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func Choose(val int, items ...interface{}) (res interface{}) {
	// return nil or possible values
	// if val is bool, cast to int 0 or 1
	res = nil
	if (val >= 0) && (val < len(items)) {
		res = items[val]
	}
	return
}

func ChooseS(val int, items ...string) (res string) {
	res = ""
	if (val >= 0) && (val < len(items)) {
		res = items[val]
	}
	return
}

func ChooseRange(val int, items ...interface{}) (res interface{}) {
	// return nil or possible values
	// assumptions not checked: items can be cast to TRangeMinMax, would have to use reflect to fix
	// assumptions not checked: minval is minval and maxval is maxval, would have to put a check in
	for i := range items {
		if (val >= items[i].(TRangeMinMax).minval) && (val <= items[i].(TRangeMinMax).maxval) {
			res = items[i]
			break
		}
	}
	return
}

