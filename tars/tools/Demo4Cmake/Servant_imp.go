package main

import (
	"context"
)

// _SERVANT_Imp servant implementation
type _SERVANT_Imp struct {
}

// Init servant init
func (imp *_SERVANT_Imp) Init() (error) {
	//initialize servant here:
	//...
	return nil
}

// Destroy servant destory
func (imp *_SERVANT_Imp) Destroy() {
	//destroy servant here:
	//...
}

func (imp *_SERVANT_Imp) Add(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	//...
	return 0, nil
}
func (imp *_SERVANT_Imp) Sub(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	//...
	return 0, nil
}
