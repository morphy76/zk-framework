package operation

import "github.com/go-zookeeper/zk"

/*
CreateOptions represents the options for a create operation.
*/
type CreateOptions struct {
	ACL  []zk.ACL
	Data []byte
	Mode int32
}

/*
CreateOptionsBuilder is a builder for createOptions.
*/
type CreateOptionsBuilder struct {
	options CreateOptions
}

/*
NewCreateOptionsBuilder creates a new CreateOptionsBuilder.
*/
func NewCreateOptionsBuilder() *CreateOptionsBuilder {
	return &CreateOptionsBuilder{
		options: CreateOptions{
			ACL:  nil,
			Data: nil,
			Mode: 0,
		},
	}
}

/*
WithACL sets the ACL for the create operation.
*/
func (cob *CreateOptionsBuilder) WithACL(acl []zk.ACL) *CreateOptionsBuilder {
	cob.options.ACL = acl
	return cob
}

/*
WithData sets the data for the create operation.
*/
func (cob *CreateOptionsBuilder) WithData(data []byte) *CreateOptionsBuilder {
	cob.options.Data = data
	return cob
}

/*
WithMode sets the mode for the create operation.
*/
func (cob *CreateOptionsBuilder) WithMode(mode int32) *CreateOptionsBuilder {
	cob.options.Mode = mode
	return cob
}

/*
Build builds the CreateOptions.
*/
func (cob *CreateOptionsBuilder) Build() CreateOptions {
	return cob.options
}
