package binding

type StructValidator interface {
	ValidateStruct(interface{}) error
	Engine() interface{}
	ValidateTag() string
}
