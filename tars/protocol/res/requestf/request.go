package requestf

// AddMessageType add message type t to message
func (st *RequestPacket) AddMessageType(t int32) {
	st.IMessageType = st.IMessageType | t
}

// HasMessageType check whether message contain type t
func (st *RequestPacket) HasMessageType(t int32) bool {
	return st.IMessageType&t != 0
}
