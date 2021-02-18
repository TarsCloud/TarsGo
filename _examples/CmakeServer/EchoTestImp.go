package main

//EchoTestImp struct
type EchoTestImp struct {
}

//Echo implement
func (imp *EchoTestImp) Echo(b []int8, c *[]int8) (int32, error) {
	*c = b
	return 0, nil
}
