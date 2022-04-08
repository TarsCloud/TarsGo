package main

import (
	"context"
)

// BackendImp servant implementation
type BackendImp struct {
}

// Init servant init
func (imp *BackendImp) Init() error {
	//initialize servant here:
	//...
	return nil
}

// Destroy servant destory
func (imp *BackendImp) Destroy() {
	//destroy servant here:
	//...
}

func (imp *BackendImp) Add(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	*c = a + b
	return 0, nil
}
func (imp *BackendImp) Sub(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	*c = a - b
	return 0, nil
}
