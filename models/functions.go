package models

import "github.com/lrndwy/gokil/orm"

func Create[T any](instance *T) error {
	_, err := orm.Create(GetContext(), instance)
	return err
}

func Save[T any](instance *T) error {
	changes := GetChangedFields(instance)
	if len(changes) == 0 {
		return nil
	}
	id := getID(instance)
	if id == nil {
		_, err := orm.Create(GetContext(), instance)
		return err
	}
	_, err := orm.UpdateByID[T](GetContext(), id, changes)
	if err != nil {
		return err
	}
	UpdateOriginalState(instance)
	return nil
}

func Delete[T any](id any) error {
	_, err := orm.DeleteByID[T](GetContext(), id)
	return err
}
