package main

import (
	"context"
)

// HelloObjImp servant implementation
type HelloObjImp struct {
}

// Init servant init
func (imp *HelloObjImp) Init() error {
	//initialize servant here:
	//...
	return nil
}

// Destroy servant destroy
func (imp *HelloObjImp) Destroy() {
	//destroy servant here:
	//...
}

func (imp *HelloObjImp) Add(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	//...
	return 0, nil
}
func (imp *HelloObjImp) Sub(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	//...
	return 0, nil
}
