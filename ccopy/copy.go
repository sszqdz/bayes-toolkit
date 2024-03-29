package ccopy

import "github.com/jinzhu/copier"

func DeepCopy(toValue interface{}, fromValue interface{}) {
	_ = copier.CopyWithOption(toValue, fromValue, copier.Option{DeepCopy: true})
}
