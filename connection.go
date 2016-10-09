package sockjs

type Connection interface {
	ReadJSON(interface{}) error
	Read(interface{}) error
	WriteJSON(interface{}) error
	Write(interface{}) error
	Close() error
}
