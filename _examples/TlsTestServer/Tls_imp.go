package main

import (
	"context"
)

// TlsImp servant implementation
type TlsImp struct {
}

// Init servant init
func (imp *TlsImp) Init() error {
	//initialize servant here:
	//...
	return nil
}

// Destroy servant destory
func (imp *TlsImp) Destroy() {
	//destroy servant here:
	//...
}

func (imp *TlsImp) Add(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	//...
	return 0, nil
}
func (imp *TlsImp) Sub(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	//...
	return 0, nil
}
